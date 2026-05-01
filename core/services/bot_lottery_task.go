package services

import (
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/repository"
	"BaseGoUni/core/utils"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	botLotteryConfigKey         = "bot_grap_second"
	botLotteryTablePrefix       = "redv2"
	botLotteryDefaultSecond     = 30
	botLotteryMinPositiveSecond = 3
)

var botLotteryRunning sync.Map

func StartBotLotteryTask() {
	startBotLotteryTask(botLotteryTablePrefix)
}

func startBotLotteryTask(tablePrefix string) {
	if _, loaded := botLotteryRunning.LoadOrStore(tablePrefix, struct{}{}); loaded {
		return
	}
	go runBotLotteryLoop(tablePrefix)
	log.Printf("[lottery_bot] task started: prefix=%s", tablePrefix)
}

func runBotLotteryLoop(tablePrefix string) {
	for {
		db := utils.NewPrefixDb(tablePrefix)
		if db == nil {
			time.Sleep(time.Duration(botLotteryDefaultSecond) * time.Second)
			continue
		}

		second := GetBotGrapSecond(db)
		if second <= 0 {
			time.Sleep(time.Duration(botLotteryDefaultSecond) * time.Second)
			continue
		}

		time.Sleep(time.Duration(second) * time.Second)
		if err := RunOneBotLottery(db, tablePrefix, second); err != nil {
			log.Printf("[lottery_bot] draw skipped: prefix=%s err=%v", tablePrefix, err)
		}
	}
}

func GetBotGrapSecond(db *gorm.DB) int {
	configValue := getOrInitSysConfigValue(
		db,
		botLotteryConfigKey,
		strconv.Itoa(botLotteryDefaultSecond),
		"机器人抽奖间隔秒数，<=0表示关闭",
	)
	second, err := strconv.Atoi(strings.TrimSpace(configValue))
	if err != nil {
		return botLotteryDefaultSecond
	}
	if second <= 0 {
		return 0
	}
	if second < botLotteryMinPositiveSecond {
		return botLotteryMinPositiveSecond
	}
	return second
}

func RunOneBotLottery(db *gorm.DB, tablePrefix string, intervalSecond int) error {
	if db == nil || strings.TrimSpace(tablePrefix) == "" {
		return nil
	}

	lockTTL := time.Duration(intervalSecond) * time.Second
	if lockTTL < 5*time.Second {
		lockTTL = 5 * time.Second
	}
	lockKey := fmt.Sprintf("bot_lottery_draw:%s", tablePrefix)
	acquired, err := utils.AcquireLock(lockKey, lockTTL)
	if err != nil || !acquired {
		return nil
	}

	config, err := repository.GetFirstActivePrizePoolConfig(db)
	if err != nil {
		return err
	}
	if config.ID == 0 {
		return nil
	}

	botUser, err := pickRandomLuckyBotUser(db)
	if err != nil {
		return err
	}
	if botUser.ID == 0 {
		return nil
	}

	tenantID := botUser.TenantId
	if tenantID == 0 {
		tenantID = config.TenantId
	}
	poolKey := fmt.Sprintf("lottery_pool:%d:%d", tenantID, config.ID)
	awardAmount := utils.Truncate2(PopOrRefillLotteryPool(poolKey, config, config.GetAmountProbMap()))

	return db.Transaction(func(tx *gorm.DB) error {
		var lockedBot pojo.TgUser
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ? AND is_bot = ? AND status = ?", botUser.ID, true, 1).
			First(&lockedBot).Error; err != nil {
			return err
		}

		status := pojo.LotteryStatusNoAward
		if awardAmount > 0 {
			status = pojo.LotteryStatusAwarded
		}
		remark := "bot_lottery"
		record := pojo.UserLotteryRecord{
			TenantId:      tenantID,
			UserId:        lockedBot.ID,
			PoolId:        config.PoolId,
			ConfigId:      config.ID,
			PeerAmount:    0,
			AwardAmount:   awardAmount,
			BeforeBalance: utils.Truncate2(lockedBot.Balance),
			AfterBalance:  utils.Truncate2(lockedBot.Balance),
			Status:        status,
			Remark:        &remark,
		}
		if err := tx.Create(&record).Error; err != nil {
			return err
		}

		poolRecordRemark := fmt.Sprintf("bot_lottery_draw_%d", record.ID)
		return repository.CreateLotteryBotDrawRecord(tx, tenantID, config.PoolId, lockedBot.ID, awardAmount, &poolRecordRemark)
	})
}
