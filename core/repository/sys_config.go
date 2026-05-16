package repository

import (
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/utils"
	"context"
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
		return result, errors.New("config_not_found")
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
		return result, errors.New("config_key_required")
	}

	var entity pojo.SysConfig
	oldConfigKey := ""
	if req.ID > 0 {
		db.Where("id = ?", req.ID).First(&entity)
		if entity.ID == 0 {
			return result, errors.New("record_not_found_update")
		}
		oldConfigKey = entity.ConfigKey
	} else {
		var exist pojo.SysConfig
		db.Where("config_key = ?", req.ConfigKey).First(&exist)
		if exist.ID > 0 {
			return result, errors.New("config_key_exists")
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
	clearSysConfigCache(oldConfigKey)
	clearSysConfigCache(entity.ConfigKey)
	_ = copier.Copy(&result, &entity)
	return result, nil
}

// DelSysConfig 删除系统配置
func DelSysConfig(db *gorm.DB, id int64) error {
	var entity pojo.SysConfig
	db.Where("id = ?", id).First(&entity)
	if entity.ID == 0 {
		return errors.New("record_not_found_delete")
	}
	if err := db.Delete(&entity).Error; err != nil {
		return err
	}
	clearSysConfigCache(entity.ConfigKey)
	return nil
}

func clearSysConfigCache(configKey string) {
	configKey = strings.TrimSpace(configKey)
	if configKey == "" || utils.RD == nil {
		return
	}

	if redisKey, ok := sysConfigRedisKeyMap[configKey]; ok {
		_ = utils.RD.Del(context.Background(), redisKey).Err()
	}
}

var sysConfigRedisKeyMap = map[string]string{
	"lucky_lose_rate":                "bgu_auth_group_lose_rate",
	"lucky_game2_lose_rate":          "bgu_auth_group_game2_lose_rate",
	"lucky_num_config":               "bgu_auth_group_num_config",
	"lucky_nums":                     "bgu_lucky_nums",
	"lucky_nums_amount":              "bgu_lucky_nums_amount",
	"lucky_send_commission":          "bgu_auth_group_send_commission",
	"lucky_grabbing_commission":      "bgu_auth_group_grabbing_commission",
	"lucky_send_pool_commission":     "bgu_auth_group_send_pool_commission",
	"lucky_grabbing_pool_commission": "bgu_auth_group_grabbing_pool_commission",
	"lucky_expire_time":              "bgu_lucky_expire_time",
	"trial_user_win_rate":            "bgu_trial_user_win_rate",
}
