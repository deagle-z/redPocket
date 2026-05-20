package repository

import (
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/utils"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"strings"
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
	return utils.Truncate2(pool.Balance), nil
}

// GetPrizePoolByCode 查询指定奖池，不存在则自动创建默认记录。
func GetPrizePoolByCode(db *gorm.DB, poolCode string) (pojo.SysTenantPrizePool, error) {
	poolCode = strings.TrimSpace(poolCode)
	if poolCode == "" {
		return pojo.SysTenantPrizePool{}, errors.New("pool_code_required")
	}

	var pool pojo.SysTenantPrizePool
	err := db.Where("pool_code = ?", poolCode).First(&pool).Error
	if err == nil && pool.ID > 0 {
		return pool, nil
	}
	if err != nil && err != gorm.ErrRecordNotFound {
		return pojo.SysTenantPrizePool{}, err
	}

	pool = pojo.SysTenantPrizePool{
		PoolCode: poolCode,
		PoolName: poolCode,
		Currency: "USD",
		Status:   1,
	}
	if err = db.Create(&pool).Error; err != nil {
		return pojo.SysTenantPrizePool{}, err
	}
	return pool, nil
}

// SetPrizePoolBalance 直接设置奖池余额，并记录一条流水。
func SetPrizePoolBalance(db *gorm.DB, req pojo.SysTenantPrizePoolBalanceSet) (pojo.SysTenantPrizePool, error) {
	poolCode := strings.TrimSpace(req.PoolCode)
	if poolCode == "" {
		return pojo.SysTenantPrizePool{}, errors.New("pool_code_required")
	}
	if req.Balance < 0 {
		return pojo.SysTenantPrizePool{}, errors.New("pool_balance_non_negative")
	}
	req.Balance = utils.Truncate2(req.Balance)

	var result pojo.SysTenantPrizePool
	err := db.Transaction(func(tx *gorm.DB) error {
		pool, err := GetPrizePoolByCode(tx, poolCode)
		if err != nil {
			return err
		}

		beforeBalance := pool.Balance
		afterBalance := req.Balance
		changeAmount := utils.Truncate2(afterBalance - beforeBalance)
		changeType := pojo.PrizePoolChangeTypeIn
		if changeAmount < 0 {
			changeType = pojo.PrizePoolChangeTypeOut
		}

		if err = tx.Model(&pojo.SysTenantPrizePool{}).
			Where("id = ?", pool.ID).
			Update("balance", afterBalance).Error; err != nil {
			return err
		}

		record := pojo.SysTenantPrizePoolRecord{
			TenantId:      pool.TenantId,
			PoolId:        pool.ID,
			ChangeType:    changeType,
			Amount:        changeAmount,
			BeforeBalance: utils.Truncate2(beforeBalance),
			AfterBalance:  afterBalance,
			Remark:        req.Remark,
		}
		if err = tx.Create(&record).Error; err != nil {
			return err
		}

		if err = tx.Where("id = ?", pool.ID).First(&result).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return pojo.SysTenantPrizePool{}, err
	}

	go broadcastPrizePoolThrottled(result.TenantId, result.PoolCode, result.Balance)
	return result, nil
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
	amount = utils.Truncate2(amount)

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
	afterBalance := utils.Truncate2(beforeBalance + amount)

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
		BeforeBalance: utils.Truncate2(beforeBalance),
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
// 抢包流水：SUM(amount + lose_money) WHERE user_id=? AND grab_type!=2
// 发包流水：SUM(amount) WHERE sender_id=?
func GetUserTotalFlow(db *gorm.DB, userID int64) (float64, error) {
	var row struct {
		Flow float64 `gorm:"column:flow"`
	}
	if err := db.Raw(`
		SELECT
			COALESCE((SELECT SUM(amount + lose_money) FROM lucky_history WHERE user_id = ? AND grab_type != 2), 0) +
			COALESCE((SELECT SUM(amount) FROM lucky_money WHERE sender_id = ?), 0) AS flow
	`, userID, userID).Scan(&row).Error; err != nil {
		return 0, err
	}

	return utils.Truncate2(row.Flow), nil
}

// GetUsedLotteryCount 查询用户已消耗抽奖次数（user_lottery_record 记录数）
func GetUsedLotteryCount(db *gorm.DB, userID int64) (int64, error) {
	var count int64
	err := db.Model(&pojo.UserLotteryRecord{}).
		Where("user_id = ? AND peer_amount > 0", userID).
		Count(&count).Error
	return count, err
}

// GetPrizePoolOutRecordsApp returns latest lottery consumption records for app display.
func GetPrizePoolOutRecordsApp(db *gorm.DB, page pojo.PageInfo) (result pojo.SysTenantPrizePoolOutRecordResp) {
	if page.PageSize <= 0 || page.PageSize > 10 {
		page.PageSize = 10
	}
	if page.CurrentPage < 0 {
		page.CurrentPage = 0
	}

	query := db.Table("sys_tenant_prize_pool_record AS r").
		Where("r.change_type = ?", pojo.PrizePoolChangeTypeOut)

	query.Count(&result.Total)
	query.Select(`r.id, r.tenant_id, r.pool_id, r.user_id,
				COALESCE(NULLIF(tg_user.first_name, ''), NULLIF(tg_user.username, ''), '') AS user_name,
				r.change_type, r.amount, r.before_balance, r.after_balance, r.consumed_amount, r.created_at`).
		Joins("LEFT JOIN " + pojo.TgUserTableName + " ON " + pojo.TgUserTableName + ".id = r.user_id").
		Order("r.id desc").
		Limit(page.PageSize).
		Offset(page.PageSize * page.CurrentPage).
		Scan(&result.List)

	result.PageSize = page.PageSize
	result.CurrentPage = page.CurrentPage
	return result
}

// CreateLotteryDrawRecord writes one lottery consumption record.
// It records consumed_amount only and does not change prize pool balance.
func CreateLotteryDrawRecord(db *gorm.DB, tenantID int64, poolID int64, userID int64, consumedAmount float64, remark *string) error {
	if consumedAmount <= 0 {
		return nil
	}
	consumedAmount = utils.Truncate2(consumedAmount)

	var pool pojo.SysTenantPrizePool
	err := gorm.ErrRecordNotFound
	if poolID > 0 {
		err = db.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ? AND tenant_id = ?", poolID, tenantID).
			First(&pool).Error
	}
	if err == gorm.ErrRecordNotFound {
		err = db.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("tenant_id = ? AND pool_code = ?", tenantID, "lucky").
			First(&pool).Error
	}
	if err == gorm.ErrRecordNotFound {
		pool = pojo.SysTenantPrizePool{
			TenantId: tenantID,
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
	record := pojo.SysTenantPrizePoolRecord{
		TenantId:       tenantID,
		PoolId:         pool.ID,
		UserId:         userIdPtr,
		ChangeType:     pojo.PrizePoolChangeTypeOut,
		Amount:         0,
		BeforeBalance:  utils.Truncate2(pool.Balance),
		AfterBalance:   utils.Truncate2(pool.Balance),
		ConsumedAmount: &consumedAmount,
		Remark:         remark,
	}
	return db.Create(&record).Error
}

// CreateLotteryBotDrawRecord writes one bot lottery pool out record without counted consumption.
func CreateLotteryBotDrawRecord(db *gorm.DB, tenantID int64, poolID int64, userID int64, awardAmount float64, remark *string) error {
	pool, err := lockLotteryPrizePool(db, tenantID, poolID)
	if err != nil {
		return err
	}
	awardAmount = utils.Truncate2(awardAmount)
	if awardAmount < 0 {
		awardAmount = 0
	}
	afterBalance := utils.Truncate2(pool.Balance - awardAmount)
	if awardAmount > 0 {
		if err := db.Model(&pojo.SysTenantPrizePool{}).
			Where("id = ?", pool.ID).
			Update("balance", gorm.Expr("balance - ?", awardAmount)).Error; err != nil {
			return fmt.Errorf("更新奖池余额失败: %v", err)
		}
	}

	userIdPtr := &userID
	record := pojo.SysTenantPrizePoolRecord{
		TenantId:      tenantID,
		PoolId:        pool.ID,
		UserId:        userIdPtr,
		ChangeType:    pojo.PrizePoolChangeTypeOut,
		Amount:        awardAmount,
		BeforeBalance: utils.Truncate2(pool.Balance),
		AfterBalance:  afterBalance,
		Remark:        remark,
	}
	return db.Create(&record).Error
}

func lockLotteryPrizePool(db *gorm.DB, tenantID int64, poolID int64) (pojo.SysTenantPrizePool, error) {
	var pool pojo.SysTenantPrizePool
	err := gorm.ErrRecordNotFound
	if poolID > 0 {
		err = db.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ? AND tenant_id = ?", poolID, tenantID).
			First(&pool).Error
	}
	if err == gorm.ErrRecordNotFound {
		err = db.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("tenant_id = ? AND pool_code = ?", tenantID, "lucky").
			First(&pool).Error
	}
	if err == gorm.ErrRecordNotFound {
		pool = pojo.SysTenantPrizePool{
			TenantId: tenantID,
			PoolCode: "lucky",
			PoolName: "lucky",
			Currency: "USD",
			Status:   1,
		}
		if err = db.Create(&pool).Error; err != nil {
			return pojo.SysTenantPrizePool{}, fmt.Errorf("创建奖池失败: %v", err)
		}
		return pool, nil
	}
	if err != nil {
		return pojo.SysTenantPrizePool{}, fmt.Errorf("查询奖池失败: %v", err)
	}
	return pool, nil
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
		"balance":  utils.Truncate2(balance),
	})
}
