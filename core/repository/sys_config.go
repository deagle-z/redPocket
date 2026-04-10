package repository

import (
	"BaseGoUni/core/pojo"
	"errors"
	"strings"

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

// GetSysConfigs 系统配置列表（分页）
func GetSysConfigs(db *gorm.DB, search pojo.SysConfigSearch) (result pojo.SysConfigResp) {
	var list []pojo.SysConfig
	query := db.Model(&pojo.SysConfig{})

	if search.ConfigKey != "" {
		query = query.Where("config_key LIKE ?", "%"+strings.TrimSpace(search.ConfigKey)+"%")
	}
	if search.ConfigDesc != "" {
		query = query.Where("config_desc LIKE ?", "%"+strings.TrimSpace(search.ConfigDesc)+"%")
	}

	query.Count(&result.Total)
	query = query.Order("id asc").Limit(search.PageSize).Offset(search.PageSize * search.CurrentPage)
	query.Find(&list)

	for _, item := range list {
		var temp pojo.SysConfigBack
		_ = copier.Copy(&temp, &item)
		result.List = append(result.List, temp)
	}
	result.PageSize = search.PageSize
	result.CurrentPage = search.CurrentPage
	return result
}

// SetSysConfig 创建或更新系统配置
func SetSysConfig(db *gorm.DB, req pojo.SysConfigSet) (result pojo.SysConfigBack, err error) {
	req.ConfigKey = strings.TrimSpace(req.ConfigKey)
	req.ConfigValue = strings.TrimSpace(req.ConfigValue)
	req.ConfigDesc = strings.TrimSpace(req.ConfigDesc)
	if req.ConfigKey == "" {
		return result, errors.New("配置Key不能为空")
	}

	var entity pojo.SysConfig
	if req.ID > 0 {
		db.Where("id = ?", req.ID).First(&entity)
		if entity.ID == 0 {
			return result, errors.New("更新的数据不存在")
		}
	} else {
		var exist pojo.SysConfig
		db.Where("config_key = ?", req.ConfigKey).First(&exist)
		if exist.ID > 0 {
			return result, errors.New("配置Key已存在")
		}
	}

	_ = copier.Copy(&entity, &req)
	if req.ID > 0 {
		err = db.Save(&entity).Error
	} else {
		err = db.Create(&entity).Error
	}
	if err != nil {
		return result, err
	}
	_ = copier.Copy(&result, &entity)
	return result, nil
}

// DelSysConfig 删除系统配置
func DelSysConfig(db *gorm.DB, id int64) error {
	var entity pojo.SysConfig
	db.Where("id = ?", id).First(&entity)
	if entity.ID == 0 {
		return errors.New("删除的数据不存在")
	}
	return db.Delete(&entity).Error
}
