package repository

import (
	"BaseGoUni/core/pojo"
	"errors"
	"github.com/jinzhu/copier"
	"gorm.io/gorm"
	"time"
)

// GetSysPayChannels 支付通道列表（分页）
func GetSysPayChannels(db *gorm.DB, search pojo.SysPayChannelSearch) (result pojo.SysPayChannelResp) {
	var list []pojo.SysPayChannel
	query := db.Model(&pojo.SysPayChannel{}).Where("deleted_at = 0")

	if search.ChannelCode != "" {
		query = query.Where("channel_code LIKE ?", "%"+search.ChannelCode+"%")
	}
	if search.ChannelName != "" {
		query = query.Where("channel_name LIKE ?", "%"+search.ChannelName+"%")
	}
	if search.ChannelType != "" {
		query = query.Where("channel_type = ?", search.ChannelType)
	}
	if search.ProviderType != "" {
		query = query.Where("provider_type = ?", search.ProviderType)
	}
	if search.CountryCode != "" {
		query = query.Where("country_code = ?", search.CountryCode)
	}
	if search.Status != nil {
		query = query.Where("status = ?", *search.Status)
	}

	query.Count(&result.Total)
	query = query.Order("sort asc, id desc").Limit(search.PageSize).Offset(search.PageSize * search.CurrentPage)
	query.Find(&list)

	for _, item := range list {
		var temp pojo.SysPayChannelBack
		_ = copier.Copy(&temp, &item)
		result.List = append(result.List, temp)
	}

	result.PageSize = search.PageSize
	result.CurrentPage = search.CurrentPage
	return result
}

// SetSysPayChannel 创建或更新支付通道
func SetSysPayChannel(db *gorm.DB, req pojo.SysPayChannelSet) (result pojo.SysPayChannelBack, err error) {
	var entity pojo.SysPayChannel
	if req.ID > 0 {
		db.Where("id = ? AND deleted_at = 0", req.ID).First(&entity)
		if entity.ID == 0 {
			return result, errors.New("更新的数据不存在")
		}
		_ = copier.Copy(&entity, &req)
		err = db.Save(&entity).Error
	} else {
		_ = copier.Copy(&entity, &req)
		err = db.Create(&entity).Error
	}
	if err != nil {
		return result, err
	}
	_ = copier.Copy(&result, &entity)
	return result, nil
}

// DelSysPayChannel 软删除支付通道
func DelSysPayChannel(db *gorm.DB, id int64) (result string, err error) {
	var entity pojo.SysPayChannel
	db.Where("id = ? AND deleted_at = 0", id).First(&entity)
	if entity.ID == 0 {
		return result, errors.New("删除的数据不存在")
	}
	err = db.Model(&entity).Update("deleted_at", time.Now().Unix()).Error
	if err != nil {
		return result, err
	}
	return "success", nil
}

// GetSysPayChannelById 根据ID获取支付通道
func GetSysPayChannelById(db *gorm.DB, id int64) (result pojo.SysPayChannelBack, err error) {
	var entity pojo.SysPayChannel
	db.Where("id = ? AND deleted_at = 0", id).First(&entity)
	if entity.ID == 0 {
		return result, errors.New("数据不存在")
	}
	_ = copier.Copy(&result, &entity)
	return result, nil
}
