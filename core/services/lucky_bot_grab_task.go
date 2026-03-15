package services

import (
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/repository"
	"BaseGoUni/core/utils"
	"context"
	"encoding/json"
	"log"
	"math/rand/v2"
	"strconv"
	"strings"
	"time"

	"github.com/hibiken/asynq"
	"gorm.io/gorm"
)

const TaskTypeLuckyBotGrab = "lucky:bot_grab"

type LuckyBotGrabPayload struct {
	TablePrefix    string `json:"tablePrefix"`
	LuckyID        int64  `json:"luckyId"`
	GrabIndex      *int   `json:"grabIndex,omitempty"`
	RemainingCount int    `json:"remainingCount"`
}

func EnqueueLuckyBotGrabTask(db *gorm.DB, tablePrefix string, luckyID int64, grabIndex *int, remainingCount int) error {
	if asynqClient == nil || db == nil || luckyID <= 0 || tablePrefix == "" {
		return nil
	}
	if remainingCount <= 0 {
		remainingCount = 1
	}
	delay := getRandomGrabDelay(db)
	payload, _ := json.Marshal(LuckyBotGrabPayload{
		TablePrefix:    tablePrefix,
		LuckyID:        luckyID,
		GrabIndex:      grabIndex,
		RemainingCount: remainingCount,
	})
	task := asynq.NewTask(TaskTypeLuckyBotGrab, payload)
	_, err := asynqClient.Enqueue(task, asynq.ProcessIn(delay), asynq.MaxRetry(10))
	return err
}

func handleLuckyBotGrabTask(ctx context.Context, task *asynq.Task) error {
	var payload LuckyBotGrabPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return err
	}
	if payload.LuckyID <= 0 || payload.TablePrefix == "" {
		return nil
	}

	db := utils.NewPrefixDb(payload.TablePrefix)
	if db == nil {
		return nil
	}

	var lucky pojo.LuckyMoney
	if err := db.Where("id = ?", payload.LuckyID).First(&lucky).Error; err != nil || lucky.ID == 0 {
		return nil
	}
	if lucky.Status != 1 {
		return nil
	}

	grabbedCount, err := repository.GetLuckyHistoryCount(db, payload.LuckyID)
	if err != nil {
		return err
	}
	if int(grabbedCount) >= lucky.Number {
		return nil
	}
	if payload.RemainingCount <= 0 {
		payload.RemainingCount = pickRandomBotGrabCount(lucky.Number - int(grabbedCount))
	}

	minBalance := lucky.Amount * lucky.LoseRate
	botUser, err := pickRandomAvailableGrabBotUser(db, payload.LuckyID, minBalance)
	if err != nil {
		return err
	}
	if botUser.ID == 0 {
		botUser, err = pickRandomGrabBotUser(db, payload.LuckyID)
		if err != nil {
			return err
		}
	}
	if botUser.ID == 0 {
		return nil
	}
	if err = ensureBotBalance(db, &botUser, minBalance); err != nil {
		return err
	}

	grabIndex := 0
	if payload.GrabIndex != nil {
		grabIndex = *payload.GrabIndex
	} else {
		grabIndex, err = pickRandomAvailableGrabIndex(db, payload.LuckyID)
		if err != nil {
			return err
		}
	}

	result, err := GrabRedPacket(db, payload.LuckyID, botUser.ID, payload.TablePrefix, grabIndex)
	if err != nil {
		log.Printf("bot grab skipped. luckyId=%d botId=%d err=%v", payload.LuckyID, botUser.ID, err)
	} else {
		if err = BroadcastLuckyGrabResult(db, payload.LuckyID, result); err != nil {
			return err
		}
	}

	var latestLucky pojo.LuckyMoney
	if err = db.Where("id = ?", payload.LuckyID).First(&latestLucky).Error; err == nil && latestLucky.ID > 0 && latestLucky.Status == 1 {
		var remaining int64
		if countErr := db.Table("lucky_money_item").
			Where("red_packet_id = ? AND is_grabbed = 0", payload.LuckyID).
			Count(&remaining).Error; countErr != nil {
			return countErr
		}
		if remaining > 0 {
			if err := EnqueueLuckyBotGrabTask(db, payload.TablePrefix, payload.LuckyID, nil, int(remaining)); err != nil {
				log.Printf("[lucky] EnqueueLuckyBotGrabTask chain failed: luckyID=%d err=%v", payload.LuckyID, err)
			}
		}
	}

	return nil
}

func pickRandomAvailableGrabBotUser(db *gorm.DB, luckyID int64, minBalance float64) (pojo.TgUser, error) {
	var botUser pojo.TgUser
	err := db.Where("is_bot = ? AND status = ? AND balance >= ?", true, 1, minBalance).
		Where("NOT EXISTS (?)",
			db.Table("lucky_history").
				Select("1").
				Where("lucky_history.lucky_id = ? AND lucky_history.user_id = tg_user.id", luckyID)).
		Order("RAND()").
		First(&botUser).Error
	if err == gorm.ErrRecordNotFound {
		return pojo.TgUser{}, nil
	}
	return botUser, err
}

func pickRandomGrabBotUser(db *gorm.DB, luckyID int64) (pojo.TgUser, error) {
	var botUser pojo.TgUser
	err := db.Where("is_bot = ? AND status = ?", true, 1).
		Where("NOT EXISTS (?)",
			db.Table("lucky_history").
				Select("1").
				Where("lucky_history.lucky_id = ? AND lucky_history.user_id = tg_user.id", luckyID)).
		Order("RAND()").
		First(&botUser).Error
	if err == gorm.ErrRecordNotFound {
		return pojo.TgUser{}, nil
	}
	return botUser, err
}

func pickRandomAvailableGrabIndex(db *gorm.DB, luckyID int64) (int, error) {
	type seqRow struct {
		SeqNo int `gorm:"column:seq_no"`
	}
	var rows []seqRow
	err := db.Table("lucky_money_item").
		Select("seq_no").
		Where("red_packet_id = ? AND is_grabbed = 0", luckyID).
		Order("seq_no asc").
		Scan(&rows).Error
	if err != nil {
		return 0, err
	}
	if len(rows) == 0 {
		return 0, nil
	}
	return rows[rand.IntN(len(rows))].SeqNo, nil
}

func pickRandomBotGrabCount(remaining int) int {
	if remaining <= 0 {
		return 1
	}
	return remaining
}

func getRandomGrabDelay(db *gorm.DB) time.Duration {
	minSecond, maxSecond := getRandomGrabSecondRange(db)
	if maxSecond < minSecond {
		maxSecond = minSecond
	}
	seconds := minSecond
	if maxSecond > minSecond {
		seconds = minSecond + rand.IntN(maxSecond-minSecond+1)
	}
	return time.Duration(seconds) * time.Second
}

func getRandomGrabSecondRange(db *gorm.DB) (int, int) {
	const defaultMin = 10
	const defaultMax = 20

	configValue := getOrInitSysConfigValue(db, "random_grab_second", "10-20", "机器人随机抢包秒数区间")
	parts := strings.Split(strings.TrimSpace(configValue), "-")
	if len(parts) != 2 {
		return defaultMin, defaultMax
	}
	minSecond, errMin := strconv.Atoi(strings.TrimSpace(parts[0]))
	maxSecond, errMax := strconv.Atoi(strings.TrimSpace(parts[1]))
	if errMin != nil || errMax != nil || minSecond < 0 || maxSecond < minSecond {
		return defaultMin, defaultMax
	}
	return minSecond, maxSecond
}
