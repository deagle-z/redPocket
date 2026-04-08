package repository

import (
	"BaseGoUni/core/pojo"
	"errors"

	"github.com/jinzhu/copier"
	"gorm.io/gorm"
)

// GetSysConfigByKey 根据key获取系统配置
func GetSysConfigByKey(db *gorm.DB, key string) (result pojo.SysConfigBack, err error) {
	var entity pojo.SysConfig
	db.Where("config_key = ?", key).First(&entity)
	if entity.ID == 0 {
		return result, errors.New("配置不存在")
	}
	_ = copier.Copy(&result, &entity)
	return result, nil
}
