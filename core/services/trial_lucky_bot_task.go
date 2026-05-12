package services

import (
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/utils"
	"context"
	"encoding/json"
	"log"
	"math/rand/v2"

	"github.com/hibiken/asynq"
	"gorm.io/gorm"
)

const TaskTypeTrialLuckyBotGrab = "trial_lucky:bot_grab"

type TrialLuckyBotGrabPayload struct {
	TablePrefix    string `json:"tablePrefix"`
	LuckyID        int64  `json:"luckyId"`
	GrabIndex      *int   `json:"grabIndex,omitempty"`
	RemainingCount int    `json:"remainingCount"`
}

func EnqueueTrialLuckyBotGrabTask(db *gorm.DB, tablePrefix string, luckyID int64, grabIndex *int, remainingCount int) error {
	if db == nil || luckyID <= 0 || tablePrefix == "" {
		return nil
	}
	if asynqClient == nil {
		return AutoTrialBotGrab(db, luckyID)
	}
	if remainingCount <= 0 {
		remainingCount = 1
	}
	payload, _ := json.Marshal(TrialLuckyBotGrabPayload{
		TablePrefix:    tablePrefix,
		LuckyID:        luckyID,
		GrabIndex:      grabIndex,
		RemainingCount: remainingCount,
	})
	task := asynq.NewTask(TaskTypeTrialLuckyBotGrab, payload)
	_, err := asynqClient.Enqueue(task, asynq.ProcessIn(getRandomGrabDelay(db)), asynq.MaxRetry(10))
	return err
}

func handleTrialLuckyBotGrabTask(ctx context.Context, task *asynq.Task) error {
	var payload TrialLuckyBotGrabPayload
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

	var lucky pojo.TrialLuckyMoney
	if err := db.Where("id = ? AND status = ?", payload.LuckyID, 1).First(&lucky).Error; err != nil || lucky.ID == 0 {
		return nil
	}

	var remaining int64
	if err := db.Model(&pojo.TrialLuckyMoneyItem{}).
		Where("red_packet_id = ? AND is_grabbed = ?", lucky.ID, 0).
		Count(&remaining).Error; err != nil {
		return err
	}
	if remaining <= 0 {
		return nil
	}

	grabIndex := 0
	if payload.GrabIndex != nil {
		grabIndex = *payload.GrabIndex
	} else {
		var seqRows []struct {
			SeqNo int `gorm:"column:seq_no"`
		}
		if err := db.Model(&pojo.TrialLuckyMoneyItem{}).
			Select("seq_no").
			Where("red_packet_id = ? AND is_grabbed = ?", lucky.ID, 0).
			Order("seq_no asc").
			Scan(&seqRows).Error; err != nil {
			return err
		}
		if len(seqRows) == 0 {
			return nil
		}
		grabIndex = seqRows[rand.IntN(len(seqRows))].SeqNo
	}

	minBalance := utils.Truncate2(lucky.Amount * lucky.LoseRate)
	if lucky.GameMode == 1 {
		minBalance = utils.Truncate2((lucky.Amount / float64(maxInt(1, lucky.Number))) * lucky.LoseRate)
	}
	bot, err := pickRandomAvailableTrialGrabBotUser(db, lucky.ID, minBalance)
	if err != nil {
		return err
	}
	if bot.ID == 0 {
		bot, err = pickRandomTrialGrabBotUser(db, lucky.ID)
		if err != nil {
			return err
		}
	}
	if bot.ID == 0 {
		return nil
	}
	if err = ensureTrialBotBalance(db, &bot, minBalance); err != nil {
		return err
	}
	if err = grabTrialRedPacketByBot(db, lucky.ID, bot.ID, grabIndex); err != nil {
		log.Printf("[trial_lucky] bot grab skipped luckyID=%d botID=%d err=%v", lucky.ID, bot.ID, err)
	}

	var latest pojo.TrialLuckyMoney
	if err = db.Where("id = ?", lucky.ID).First(&latest).Error; err == nil && latest.ID > 0 && latest.Status == 1 {
		if err = db.Model(&pojo.TrialLuckyMoneyItem{}).
			Where("red_packet_id = ? AND is_grabbed = ?", lucky.ID, 0).
			Count(&remaining).Error; err != nil {
			return err
		}
		if remaining > 0 {
			if err := EnqueueTrialLuckyBotGrabTask(db, payload.TablePrefix, lucky.ID, nil, int(remaining)); err != nil {
				log.Printf("[trial_lucky] enqueue chain grab failed luckyID=%d err=%v", lucky.ID, err)
			}
		}
	}
	if latest.ID > 0 && latest.Status != 1 {
		_ = EnsureMinActiveTrialLuckyPackets(db, payload.TablePrefix)
	}
	return nil
}

func pickRandomAvailableTrialGrabBotUser(db *gorm.DB, luckyID int64, minBalance float64) (pojo.TrialBotUser, error) {
	var bot pojo.TrialBotUser
	err := db.Where("status = ? AND balance >= ?", 1, minBalance).
		Where("NOT EXISTS (?)",
			db.Table("trial_lucky_history").
				Select("1").
				Where("trial_lucky_history.lucky_id = ? AND trial_lucky_history.actor_type = ? AND trial_lucky_history.user_id = trial_bot_user.id", luckyID, pojo.TrialActorBot)).
		Order("RAND()").
		First(&bot).Error
	if err == gorm.ErrRecordNotFound {
		return pojo.TrialBotUser{}, nil
	}
	return bot, err
}

func pickRandomTrialGrabBotUser(db *gorm.DB, luckyID int64) (pojo.TrialBotUser, error) {
	var bot pojo.TrialBotUser
	err := db.Where("status = ?", 1).
		Where("NOT EXISTS (?)",
			db.Table("trial_lucky_history").
				Select("1").
				Where("trial_lucky_history.lucky_id = ? AND trial_lucky_history.actor_type = ? AND trial_lucky_history.user_id = trial_bot_user.id", luckyID, pojo.TrialActorBot)).
		Order("RAND()").
		First(&bot).Error
	if err == gorm.ErrRecordNotFound {
		return pojo.TrialBotUser{}, nil
	}
	return bot, err
}

func pickRandomTrialBotGrabCount(remaining int) int {
	if remaining <= 0 {
		return 1
	}
	return remaining
}
