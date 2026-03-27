package repository

import (
	"BaseGoUni/core/pojo"
	"fmt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// DepositPrizePool 向奖池注入金额（事务内调用）
// 若奖池记录不存在则自动创建；使用 SELECT FOR UPDATE 保证并发安全
func DepositPrizePool(tx *gorm.DB, tenantId int64, poolCode string, bizType string, bizId string, userId int64, amount float64) error {
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
	bizIdPtr := &bizId
	userIdPtr := &userId
	record := pojo.SysTenantPrizePoolRecord{
		TenantId:      tenantId,
		PoolId:        pool.ID,
		BizType:       bizType,
		BizId:         bizIdPtr,
		UserId:        userIdPtr,
		ChangeType:    pojo.PrizePoolChangeTypeIn,
		Amount:        amount,
		BeforeBalance: beforeBalance,
		AfterBalance:  afterBalance,
	}
	if err := tx.Create(&record).Error; err != nil {
		return fmt.Errorf("写入奖池流水失败: %v", err)
	}

	return nil
}
