package repository

import (
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/utils"
	"fmt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"sync"
	"time"
)

// GetPrizePoolBalance 查询指定奖池当前余额，不存在则返回 0
func GetPrizePoolBalance(db *gorm.DB, poolCode string) (float64, error) {
	var pool pojo.SysTenantPrizePool
	err := db.Where("pool_code = ?", poolCode).First(&pool).Error
	if err == gorm.ErrRecordNotFound {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return pool.Balance, nil
}

// prizePoolThrottle 节流控制：key = "tenantId:poolCode"，value = 上次推送时间
var prizePoolThrottle sync.Map

// DepositPrizePool 向奖池注入金额（事务内调用）
// 若奖池记录不存在则自动创建；使用 SELECT FOR UPDATE 保证并发安全
// 事务提交后异步触发节流推送
func DepositPrizePool(tx *gorm.DB, tenantId int64, poolCode string, userId int64, amount float64) error {
	if amount <= 0 {
		return nil
	}

	// 锁定奖池行，不存在则创建
	var pool pojo.SysTenantPrizePool
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("tenant_id = ? AND pool_code = ?", tenantId, poolCode).
		First(&pool).Error
	if err == gorm.ErrRecordNotFound {
		pool = pojo.SysTenantPrizePool{
			TenantId: tenantId,
			PoolCode: poolCode,
			PoolName: poolCode,
			Currency: "USD",
			Status:   1,
		}
		if createErr := tx.Create(&pool).Error; createErr != nil {
			return fmt.Errorf("创建奖池失败: %v", createErr)
		}
		// 重新锁定
		if err2 := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ?", pool.ID).First(&pool).Error; err2 != nil {
			return fmt.Errorf("锁定奖池失败: %v", err2)
		}
	} else if err != nil {
		return fmt.Errorf("查询奖池失败: %v", err)
	}

	beforeBalance := pool.Balance
	afterBalance := beforeBalance + amount

	// 更新奖池余额
	if err := tx.Model(&pojo.SysTenantPrizePool{}).
		Where("id = ?", pool.ID).
		Update("balance", gorm.Expr("balance + ?", amount)).Error; err != nil {
		return fmt.Errorf("更新奖池余额失败: %v", err)
	}

	// 写入流水
	userIdPtr := &userId
	record := pojo.SysTenantPrizePoolRecord{
		TenantId:      tenantId,
		PoolId:        pool.ID,
		UserId:        userIdPtr,
		ChangeType:    pojo.PrizePoolChangeTypeIn,
		Amount:        amount,
		BeforeBalance: beforeBalance,
		AfterBalance:  afterBalance,
	}
	if err := tx.Create(&record).Error; err != nil {
		return fmt.Errorf("写入奖池流水失败: %v", err)
	}

	// 事务提交后异步触发节流推送（用 afterBalance 作为乐观估算值，避免再次查库）
	go broadcastPrizePoolThrottled(tenantId, poolCode, afterBalance)

	return nil
}

// GetUserTotalFlow 计算用户累计流水
// 抢包流水：SUM(amount + lose_money) WHERE user_id=? AND grab_type=1
// 发包流水：SUM(amount) WHERE sender_id=?
func GetUserTotalFlow(db *gorm.DB, userID int64) (float64, error) {
	var grabFlow float64
	if err := db.Model(&pojo.LuckyHistory{}).
		Select("COALESCE(SUM(amount + lose_money), 0)").
		Where("user_id = ? AND grab_type = 1", userID).
		Scan(&grabFlow).Error; err != nil {
		return 0, err
	}

	var sendFlow float64
	if err := db.Model(&pojo.LuckyMoney{}).
		Select("COALESCE(SUM(amount), 0)").
		Where("sender_id = ?", userID).
		Scan(&sendFlow).Error; err != nil {
		return 0, err
	}

	return grabFlow + sendFlow, nil
}

// GetUsedLotteryCount 查询用户已消耗抽奖次数（user_lottery_record 记录数）
func GetUsedLotteryCount(db *gorm.DB, userID int64) (int64, error) {
	var count int64
	err := db.Model(&pojo.UserLotteryRecord{}).
		Where("user_id = ?", userID).
		Count(&count).Error
	return count, err
}

// CreateLotteryDrawRecord 写入一条抽奖消耗流水（change_type=out，consumed_amount=1000）
// 不改变奖池余额（amount=0），等奖品逻辑上线后再补充
func CreateLotteryDrawRecord(db *gorm.DB, userID int64) error {
	var pool pojo.SysTenantPrizePool
	err := db.Where("tenant_id = ? AND pool_code = ?", "lucky").First(&pool).Error
	if err == gorm.ErrRecordNotFound {
		pool = pojo.SysTenantPrizePool{
			PoolCode: "lucky",
			PoolName: "lucky",
			Currency: "USD",
			Status:   1,
		}
		if err = db.Create(&pool).Error; err != nil {
			return fmt.Errorf("创建奖池失败: %v", err)
		}
	} else if err != nil {
		return fmt.Errorf("查询奖池失败: %v", err)
	}

	userIdPtr := &userID
	consumed := 1000.0
	record := pojo.SysTenantPrizePoolRecord{
		PoolId:         pool.ID,
		UserId:         userIdPtr,
		ChangeType:     pojo.PrizePoolChangeTypeOut,
		Amount:         0,
		BeforeBalance:  pool.Balance,
		AfterBalance:   pool.Balance,
		ConsumedAmount: &consumed,
	}
	return db.Create(&record).Error
}

// broadcastPrizePoolThrottled 节流推送奖池余额，同一奖池 1s 内最多推送 1 次
func broadcastPrizePoolThrottled(tenantId int64, poolCode string, balance float64) {
	key := fmt.Sprintf("%d:%s", tenantId, poolCode)
	now := time.Now()

	actual, loaded := prizePoolThrottle.LoadOrStore(key, now)
	if loaded {
		last := actual.(time.Time)
		if now.Sub(last) < time.Second {
			return // 节流，跳过本次
		}
		prizePoolThrottle.Store(key, now)
	}

	_ = utils.BroadcastWsWithType("prize_pool_balance", map[string]interface{}{
		"tenantId": tenantId,
		"poolCode": poolCode,
		"balance":  balance,
	})
}
