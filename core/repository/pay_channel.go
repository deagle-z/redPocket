package repository

import (
	"BaseGoUni/core/pojo"
	"errors"
	"github.com/jinzhu/copier"
	"gorm.io/gorm"
)

// GetPayChannels 支付通道列表（分页）
func GetPayChannels(db *gorm.DB, search pojo.PayChannelSearch) (result pojo.PayChannelResp) {
	var list []pojo.PayChannel
	query := db.Model(&pojo.PayChannel{})

	if search.TenantId > 0 {
		query = query.Where("tenant_id = ?", search.TenantId)
	}
	if search.ChannelCode != "" {
		query = query.Where("channel_code = ?", search.ChannelCode)
	}
	if search.ChannelName != "" {
		query = query.Where("channel_name LIKE ?", "%"+search.ChannelName+"%")
	}
	if search.ChannelType != "" {
		query = query.Where("channel_type = ?", search.ChannelType)
	}
	if search.ProviderName != "" {
		query = query.Where("provider_name LIKE ?", "%"+search.ProviderName+"%")
	}
	if search.Status != nil {
		query = query.Where("status = ?", *search.Status)
	}

	query.Count(&result.Total)
	query = query.Order("sort_no asc, id desc").Limit(search.PageSize).Offset(search.PageSize * search.CurrentPage)
	query.Find(&list)

	for _, item := range list {
		var temp pojo.PayChannelBack
		_ = copier.Copy(&temp, &item)
		result.List = append(result.List, temp)
	}

	result.PageSize = search.PageSize
	result.CurrentPage = search.CurrentPage
	return result
}

// SetPayChannel 创建或更新支付通道
func SetPayChannel(db *gorm.DB, req pojo.PayChannelSet) (result pojo.PayChannelBack, err error) {
	var entity pojo.PayChannel
	if req.ID > 0 {
		db.Where("id = ?", req.ID).First(&entity)
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

// DelPayChannel 删除支付通道
func DelPayChannel(db *gorm.DB, id int64) (result string, err error) {
	var entity pojo.PayChannel
	db.Where("id = ?", id).First(&entity)
	if entity.ID == 0 {
		return result, errors.New("删除的数据不存在")
	}
	err = db.Delete(&entity).Error
	if err != nil {
		return result, err
	}
	return "success", nil
}

// GetPayChannelById 根据ID获取支付通道
func GetPayChannelById(db *gorm.DB, id int64) (result pojo.PayChannelBack, err error) {
	var entity pojo.PayChannel
	db.Where("id = ?", id).First(&entity)
	if entity.ID == 0 {
		return result, errors.New("数据不存在")
	}
	_ = copier.Copy(&result, &entity)
	return result, nil
}
