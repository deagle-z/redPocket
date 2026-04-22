package repository

import (
	"BaseGoUni/core/pojo"
	"errors"
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
	err := db.Delete(&entity).Error
	return err
}
