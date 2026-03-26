package repository

import (
	"BaseGoUni/core/pojo"
	"errors"
	"strings"

	"gorm.io/gorm"
)

// ResolveSourceChannelByCode 根据编码解析启用中的来源渠道。
func ResolveSourceChannelByCode(db *gorm.DB, tenantID int64, channelCode string) (*pojo.SysSourceChannel, error) {
	channelCode = strings.TrimSpace(channelCode)
	if db == nil || channelCode == "" {
		return nil, nil
	}

	var channel pojo.SysSourceChannel
	query := db.Model(&pojo.SysSourceChannel{}).
		Where("channel_code = ? AND status = ?", channelCode, 1)
	if tenantID > 0 {
		query = query.Where("tenant_id IN ?", []int64{tenantID, 0}).Order("tenant_id desc")
	} else {
		query = query.Where("tenant_id = ?", 0)
	}
	if err := query.First(&channel).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &channel, nil
}

// LoadUserSourceChannelSnapshot 读取用户来源渠道快照，便于复制到业务表。
func LoadUserSourceChannelSnapshot(db *gorm.DB, userID int64) (*int64, *string, error) {
	if db == nil || userID <= 0 {
		return nil, nil, nil
	}

	var user pojo.TgUser
	if err := db.Model(&pojo.TgUser{}).
		Select("source_channel_id", "source_channel_code").
		Where("id = ?", userID).
		First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, nil
		}
		return nil, nil, err
	}
	return user.SourceChannelID, user.SourceChannelCode, nil
}
