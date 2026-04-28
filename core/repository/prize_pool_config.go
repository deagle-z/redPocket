package repository

import (
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/utils"
	"context"
	"errors"
	"log"
	"strconv"

	"github.com/jinzhu/copier"
	"gorm.io/gorm"
)

// GetPrizePoolConfigByPoolId 根据 poolId 查询概率配置，不存在时返回空结构体（ID=0）
func GetPrizePoolConfigByPoolId(db *gorm.DB, poolId int64) (pojo.SysTenantPrizePoolConfig, error) {
	var entity pojo.SysTenantPrizePoolConfig
	err := db.Where("pool_id = ?", poolId).First(&entity).Error
	if err == gorm.ErrRecordNotFound {
		return entity, nil
	}
	return entity, err
}

// SetPrizePoolConfig 创建或更新奖池概率配置
func SetPrizePoolConfig(db *gorm.DB, req pojo.SysTenantPrizePoolConfigSet) (pojo.SysTenantPrizePoolConfig, error) {
	var entity pojo.SysTenantPrizePoolConfig

	// 按 pool_id 查找已有记录（upsert 语义）
	db.Where("pool_id = ?", req.PoolId).First(&entity)

	_ = copier.Copy(&entity, &req)

	var err error
	if entity.ID > 0 {
		err = db.Save(&entity).Error
	} else {
		err = db.Create(&entity).Error
	}
	if err != nil {
		return entity, err
	}
	if clearErr := ClearLotteryPoolBatchCache(entity.ID); clearErr != nil {
		log.Printf("[prize_pool_config] clear lottery pool batch cache failed configID=%d err=%v", entity.ID, clearErr)
	}
	return entity, nil
}

// GetFirstActivePrizePoolConfig 获取第一条启用状态的概率配置
func GetFirstActivePrizePoolConfig(db *gorm.DB) (pojo.SysTenantPrizePoolConfig, error) {
	var entity pojo.SysTenantPrizePoolConfig
	err := db.Where("status = 1").First(&entity).Error
	if err == gorm.ErrRecordNotFound {
		return entity, nil
	}
	return entity, err
}

// DelPrizePoolConfig 删除奖池概率配置
func DelPrizePoolConfig(db *gorm.DB, id int64) error {
	var entity pojo.SysTenantPrizePoolConfig
	db.Where("id = ?", id).First(&entity)
	if entity.ID == 0 {
		return errors.New("record_not_found")
	}
	if err := db.Delete(&entity).Error; err != nil {
		return err
	}
	if clearErr := ClearLotteryPoolBatchCache(entity.ID); clearErr != nil {
		log.Printf("[prize_pool_config] clear lottery pool batch cache failed configID=%d err=%v", entity.ID, clearErr)
	}
	return nil
}

// ClearLotteryPoolBatchCache clears generated lottery batches for a config across tenants.
func ClearLotteryPoolBatchCache(configID int64) error {
	if configID <= 0 || utils.RD == nil {
		return nil
	}

	ctx := context.Background()
	pattern := "lottery_pool:*:" + strconv.FormatInt(configID, 10)
	var cursor uint64
	for {
		keys, nextCursor, err := utils.RD.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return err
		}
		if len(keys) > 0 {
			if err := utils.RD.Del(ctx, keys...).Err(); err != nil {
				return err
			}
		}
		if nextCursor == 0 {
			return nil
		}
		cursor = nextCursor
	}
}
