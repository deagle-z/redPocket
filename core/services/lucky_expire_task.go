package services

import (
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/utils"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/hibiken/asynq"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const TaskTypeLuckyExpire = "lucky:expire"

type LuckyExpirePayload struct {
	TablePrefix string `json:"tablePrefix"`
	LuckyID     int64  `json:"luckyId"`
}

var (
	asynqOnce   sync.Once
	asynqClient *asynq.Client
)

// InitLuckyExpireAsynq 初始化 asynq client + worker
func InitLuckyExpireAsynq() error {
	var initErr error
	asynqOnce.Do(func() {
		redisOpt := asynq.RedisClientOpt{
			Addr:     utils.GlobalConfig.Redis.Host,
			Password: utils.GlobalConfig.Redis.Pass,
			DB:       utils.GlobalConfig.Redis.Db,
		}
		asynqClient = asynq.NewClient(redisOpt)

		srv := asynq.NewServer(redisOpt, asynq.Config{
			Concurrency: 5,
			Queues: map[string]int{
				"default": 1,
			},
		})
		mux := asynq.NewServeMux()
		mux.HandleFunc(TaskTypeLuckyExpire, handleLuckyExpireTask)
		mux.HandleFunc(TaskTypeLuckyBotGrab, handleLuckyBotGrabTask)
		go func() {
			if err := srv.Run(mux); err != nil {
				log.Printf("asynq server stopped: %v", err)
			}
		}()
	})
	return initErr
}

// EnqueueLuckyExpireTask 入队红包过期任务
func EnqueueLuckyExpireTask(tablePrefix string, luckyID int64, expireAt time.Time) error {
	if asynqClient == nil || luckyID <= 0 {
		return nil
	}
	payload, _ := json.Marshal(LuckyExpirePayload{
		TablePrefix: tablePrefix,
		LuckyID:     luckyID,
	})
	task := asynq.NewTask(TaskTypeLuckyExpire, payload)
	_, err := asynqClient.Enqueue(task, asynq.ProcessAt(expireAt), asynq.MaxRetry(10))
	return err
}

func handleLuckyExpireTask(ctx context.Context, task *asynq.Task) error {
	var payload LuckyExpirePayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return err
	}
	if payload.LuckyID <= 0 || payload.TablePrefix == "" {
		return nil
	}
	db := utils.NewPrefixDb(payload.TablePrefix)
	if db == nil {
		return fmt.Errorf("db not ready for prefix=%s", payload.TablePrefix)
	}
	return refundExpiredLuckyMoney(db, payload.TablePrefix, payload.LuckyID)
}

func refundExpiredLuckyMoney(db *gorm.DB, tablePrefix string, luckyID int64) error {
	err := db.Transaction(func(tx *gorm.DB) error {
		var lucky pojo.LuckyMoney
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", luckyID).First(&lucky).Error; err != nil {
			return err
		}
		if lucky.Status != 1 {
			return nil
		}
		if !lucky.ExpireTime.IsZero() && time.Now().Before(lucky.ExpireTime) {
			return nil
		}

		refundAmount := lucky.Amount - lucky.Received
		if refundAmount < 0 {
			refundAmount = 0
		}

		updates := map[string]any{
			"status": 2,
		}
		if err := tx.Model(&pojo.LuckyMoney{}).Where("id = ?", luckyID).Updates(updates).Error; err != nil {
			return err
		}

		if refundAmount <= 0 {
			return nil
		}

		var sender pojo.TgUser
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", lucky.SenderID).First(&sender).Error; err != nil {
			return err
		}
		if sender.ID == 0 {
			return nil
		}

		if err := tx.Model(&pojo.TgUser{}).
			Where("id = ?", sender.ID).
			Update("balance", gorm.Expr("balance + ?", refundAmount)).Error; err != nil {
			return err
		}

		cashHistory := pojo.CashHistory{
			UserId:      sender.ID,
			AwardUni:    fmt.Sprintf("lucky_expire_refund_%d", lucky.ID),
			Amount:      refundAmount,
			StartAmount: sender.Balance,
			EndAmount:   sender.Balance + refundAmount,
			CashMark:    "红包过期退回",
			CashDesc:    fmt.Sprintf("红包#%d过期，退回未抢金额%.3f", lucky.ID, refundAmount),
			Type:        pojo.CashHistoryTypeLuckyExpireRefund,
			IsGift:      0,
			FromUserId:  0,
		}
		return tx.Create(&cashHistory).Error
	})
	if err != nil {
		return err
	}
	go BroadcastLuckyFinished(db, luckyID)
	return EnsureMinActiveLuckyPackets(db, tablePrefix)
}

func SweepExpiredLuckyPacketsAllHosts() {
	startAt := time.Now()
	var hostInfos []pojo.HostInfo
	if err := utils.Db.Model(&pojo.HostInfo{}).
		Where("enabled = ?", true).
		Where("table_prefix <> ''").
		Find(&hostInfos).Error; err != nil {
		log.Printf("[lucky] sweep host infos failed: %v", err)
		return
	}

	seenPrefixes := make(map[string]struct{}, len(hostInfos))
	totalMatched := 0
	totalProcessed := 0
	totalFailed := 0
	for _, hostInfo := range hostInfos {
		tablePrefix := hostInfo.TablePrefix
		if tablePrefix == "" {
			continue
		}
		if _, ok := seenPrefixes[tablePrefix]; ok {
			continue
		}
		seenPrefixes[tablePrefix] = struct{}{}
		matched, processed, failed := sweepExpiredLuckyPacketsByPrefix(tablePrefix)
		totalMatched += matched
		totalProcessed += processed
		totalFailed += failed
	}
	log.Printf("[lucky] sweep finished: prefixes=%d matched=%d processed=%d failed=%d cost=%.2fs",
		len(seenPrefixes), totalMatched, totalProcessed, totalFailed, time.Since(startAt).Seconds())
}

func sweepExpiredLuckyPacketsByPrefix(tablePrefix string) (matched int, processed int, failed int) {
	db := utils.NewPrefixDb(tablePrefix)
	if db == nil {
		log.Printf("[lucky] sweep skipped: db not ready for prefix=%s", tablePrefix)
		return 0, 0, 0
	}

	startAt := time.Now()
	now := time.Now()
	fallbackExpireAt := now.Add(-3 * time.Minute)
	var luckyIDs []int64
	if err := db.Model(&pojo.LuckyMoney{}).
		Where("status = ?", 1).
		Where("(expire_time > ? AND expire_time <= ?) OR (expire_time <= ? AND created_at <= ?)", time.Time{}, now, time.Time{}, fallbackExpireAt).
		Limit(200).
		Pluck("id", &luckyIDs).Error; err != nil {
		log.Printf("[lucky] sweep query failed: prefix=%s err=%v", tablePrefix, err)
		return 0, 0, 0
	}

	matched = len(luckyIDs)
	if matched == 0 {
		return matched, 0, 0
	}

	for _, luckyID := range luckyIDs {
		if err := refundExpiredLuckyMoney(db, tablePrefix, luckyID); err != nil {
			failed++
			log.Printf("[lucky] sweep refund failed: prefix=%s luckyID=%d err=%v", tablePrefix, luckyID, err)
			continue
		}
		processed++
	}
	log.Printf("[lucky] sweep prefix=%s matched=%d processed=%d failed=%d cost=%.2fs",
		tablePrefix, matched, processed, failed, time.Since(startAt).Seconds())
	return matched, processed, failed
}
