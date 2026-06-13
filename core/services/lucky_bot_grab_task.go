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

type luckyBotAutoGrabTargetMode int

const (
	luckyBotAutoGrabNoTarget luckyBotAutoGrabTargetMode = iota
	luckyBotAutoGrabWinRateTarget
	luckyBotAutoGrabDailyThunderTarget
)

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
	// 红包玩法已停止：机器人不再抢包
	return nil

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

	minBalance := utils.Truncate2(lucky.Amount * lucky.LoseRate)
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

	var oddEvenGuess *int
	if lucky.GameMode == 1 {
		guess := rand.IntN(2) // 0=偶, 1=奇
		oddEvenGuess = &guess
	}
	targetWin, canGrab, err := resolveLuckyBotAutoGrabTargetWin(db, lucky)
	if err != nil {
		return err
	}
	if !canGrab {
		log.Printf("bot grab stopped. luckyId=%d botId=%d reason=thunder_quota_reached", payload.LuckyID, botUser.ID)
		return nil
	}
	result, err := grabRedPacketWithTarget(db, payload.LuckyID, botUser.ID, payload.TablePrefix, grabIndex, oddEvenGuess, targetWin)
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

func resolveLuckyBotAutoGrabTargetWin(db *gorm.DB, lucky pojo.LuckyMoney) (*bool, bool, error) {
	senderIsBot, err := isLuckyMoneyUserBot(db, lucky.SenderID)
	if err != nil {
		return nil, false, err
	}
	switch luckyBotAutoGrabTargetModeFor(lucky, senderIsBot) {
	case luckyBotAutoGrabNoTarget:
		return nil, true, nil
	case luckyBotAutoGrabWinRateTarget:
		targetWin := rand.Float64() < GetLuckyBotGrabWinRate(db)
		return &targetWin, true, nil
	}

	var items []pojo.LuckyMoneyItem
	if err := db.Where("red_packet_id = ?", lucky.ID).Order("seq_no asc").Find(&items).Error; err != nil {
		return nil, false, err
	}
	// 兼容旧数据：没有明细的历史红包不做雷包总数统计，继续走原抢包流程。
	if len(items) == 0 {
		return nil, true, nil
	}

	dayStart, dayEnd := luckyBotThunderDayRange(time.Now())
	dayTotalThunder, err := countDailyRealUserLuckyThunderItems(db, dayStart, dayEnd, false)
	if err != nil {
		return nil, false, err
	}
	dayRemainingThunder, err := countDailyRealUserLuckyThunderItems(db, dayStart, dayEnd, true)
	if err != nil {
		return nil, false, err
	}
	botThunderGrabbed, err := countDailyLuckyBotGrabbedThunder(db, dayStart, dayEnd)
	if err != nil {
		return nil, false, err
	}
	targetWin, canGrab := pickLuckyBotThunderModeTarget(items, lucky, int(dayTotalThunder), int(dayRemainingThunder), botThunderGrabbed, rand.Float64())
	return targetWin, canGrab, nil
}

func luckyBotAutoGrabTargetModeFor(lucky pojo.LuckyMoney, senderIsBot bool) luckyBotAutoGrabTargetMode {
	if senderIsBot {
		return luckyBotAutoGrabNoTarget
	}
	if lucky.GameMode == 1 {
		return luckyBotAutoGrabWinRateTarget
	}
	return luckyBotAutoGrabDailyThunderTarget
}

func luckyBotThunderDayRange(now time.Time) (time.Time, time.Time) {
	localNow := now.In(time.Local)
	dayStart := time.Date(localNow.Year(), localNow.Month(), localNow.Day(), 0, 0, 0, 0, localNow.Location())
	return dayStart, dayStart.AddDate(0, 0, 1)
}

func countDailyRealUserLuckyThunderItems(db *gorm.DB, dayStart time.Time, dayEnd time.Time, onlyRemaining bool) (int64, error) {
	var count int64
	query := db.Table(pojo.LuckyMoneyItemTableName+" AS i").
		Joins("JOIN "+pojo.LuckyMoneyTableName+" AS l ON l.id = i.red_packet_id").
		Joins("JOIN "+pojo.TgUserTableName+" AS sender ON sender.id = l.sender_id").
		Where("l.created_at >= ? AND l.created_at < ?", dayStart, dayEnd).
		Where("l.game_mode = ? AND sender.is_bot = ?", 0, false).
		Where("MOD(CAST(ROUND(i.amount * 100) AS UNSIGNED), 10) = l.thunder")
	if onlyRemaining {
		query = query.Where("l.status = ? AND i.is_grabbed = ?", 1, 0)
	}
	err := query.Count(&count).Error
	return count, err
}

func countDailyLuckyBotGrabbedThunder(db *gorm.DB, dayStart time.Time, dayEnd time.Time) (int64, error) {
	var count int64
	err := db.Table(pojo.LuckyMoneyItemTableName+" AS i").
		Joins("JOIN "+pojo.LuckyMoneyTableName+" AS l ON l.id = i.red_packet_id").
		Joins("JOIN "+pojo.TgUserTableName+" AS sender ON sender.id = l.sender_id").
		Joins("JOIN "+pojo.TgUserTableName+" AS grabber ON grabber.id = i.grabbed_uid").
		Where("l.created_at >= ? AND l.created_at < ?", dayStart, dayEnd).
		Where("l.game_mode = ? AND sender.is_bot = ?", 0, false).
		Where("i.is_grabbed = ? AND i.thunder = ? AND grabber.is_bot = ?", 1, 1, true).
		Count(&count).Error
	return count, err
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
