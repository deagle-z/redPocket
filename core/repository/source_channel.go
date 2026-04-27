package repository

import (
	"BaseGoUni/core/pojo"
	"errors"
	"net/url"
	"strings"

	"gorm.io/gorm"
)

func NormalizeSourceChannelCode(channelCode string) string {
	return strings.ToUpper(strings.TrimSpace(channelCode))
}

func FirstSourceChannelCode(values ...string) string {
	for _, value := range values {
		if code := NormalizeSourceChannelCode(value); code != "" {
			return code
		}
	}
	return ""
}

func BuildSourceChannelLinkURL(baseURL string, channelCode string) string {
	channelCode = NormalizeSourceChannelCode(channelCode)
	path := "/register?sc=" + url.QueryEscape(channelCode)
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL == "" {
		return path
	}
	return baseURL + path
}

// ResolveSourceChannelByCode 根据编码解析启用中的来源渠道。
func ResolveSourceChannelByCode(db *gorm.DB, tenantID int64, channelCode string) (*pojo.SysSourceChannel, error) {
	_ = tenantID
	channelCode = NormalizeSourceChannelCode(channelCode)
	if db == nil || channelCode == "" {
		return nil, nil
	}

	var channel pojo.SysSourceChannel
	query := db.Model(&pojo.SysSourceChannel{}).Where("channel_code = ? AND status = ?", channelCode, 1)
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
