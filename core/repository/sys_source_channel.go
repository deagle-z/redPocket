package repository

import (
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/utils"
	"errors"
	"strings"

	"github.com/jinzhu/copier"
	"gorm.io/gorm"
)

// GetSysSourceChannels 投流来源渠道列表（分页）
func GetSysSourceChannels(db *gorm.DB, search pojo.SysSourceChannelSearch) (result pojo.SysSourceChannelResp) {
	var list []pojo.SysSourceChannel
	query := db.Model(&pojo.SysSourceChannel{})

	if search.TenantId > 0 {
		query = query.Where("tenant_id = ?", search.TenantId)
	}
	if search.ChannelCode != "" {
		query = query.Where("channel_code LIKE ?", "%"+search.ChannelCode+"%")
	}
	if search.ChannelName != "" {
		query = query.Where("channel_name LIKE ?", "%"+search.ChannelName+"%")
	}
	if search.ParentID != nil {
		query = query.Where("parent_id = ?", *search.ParentID)
	}
	if search.Level != nil {
		query = query.Where("level = ?", *search.Level)
	}
	if search.Status != nil {
		query = query.Where("status = ?", *search.Status)
	}

	query.Count(&result.Total)
	query = query.Order("sort asc, id desc").Limit(search.PageSize).Offset(search.PageSize * search.CurrentPage)
	query.Find(&list)

	for _, item := range list {
		var temp pojo.SysSourceChannelBack
		_ = copier.Copy(&temp, &item)
		result.List = append(result.List, temp)
	}

	result.PageSize = search.PageSize
	result.CurrentPage = search.CurrentPage
	return result
}

// SetSysSourceChannel 创建或更新投流来源渠道
func SetSysSourceChannel(db *gorm.DB, req pojo.SysSourceChannelSet) (result pojo.SysSourceChannelBack, err error) {
	var entity pojo.SysSourceChannel
	if req.ID > 0 {
		db.Where("id = ?", req.ID).First(&entity)
		if entity.ID == 0 {
			return result, errors.New("更新的数据不存在")
		}
		req.ChannelCode = entity.ChannelCode
		_ = copier.Copy(&entity, &req)
		entity.ChannelCode = strings.TrimSpace(entity.ChannelCode)
		err = db.Save(&entity).Error
	} else {
		_ = copier.Copy(&entity, &req)
		entity.ChannelCode, err = generateUniqueSourceChannelCode(db)
		if err != nil {
			return result, err
		}
		err = db.Create(&entity).Error
	}
	if err != nil {
		return result, err
	}
	_ = copier.Copy(&result, &entity)
	return result, nil
}

func generateUniqueSourceChannelCode(db *gorm.DB) (string, error) {
	for i := 0; i < 20; i++ {
		code := strings.ToUpper(utils.RandomString(8))
		var count int64
		if err := db.Model(&pojo.SysSourceChannel{}).Where("channel_code = ?", code).Count(&count).Error; err != nil {
			return "", err
		}
		if count == 0 {
			return code, nil
		}
	}
	return "", errors.New("生成渠道编码失败，请重试")
}

// DelSysSourceChannel 删除投流来源渠道
func DelSysSourceChannel(db *gorm.DB, id int64) (result string, err error) {
	var entity pojo.SysSourceChannel
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

// GetSysSourceChannelById 根据ID获取投流来源渠道
func GetSysSourceChannelById(db *gorm.DB, id int64) (result pojo.SysSourceChannelBack, err error) {
	var entity pojo.SysSourceChannel
	db.Where("id = ?", id).First(&entity)
	if entity.ID == 0 {
		return result, errors.New("数据不存在")
	}
	_ = copier.Copy(&result, &entity)
	return result, nil
}
