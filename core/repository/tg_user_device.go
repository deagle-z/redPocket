package repository

import (
	"BaseGoUni/core/pojo"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
)

type normalizedTgDeviceInfo struct {
	fingerprint *string
	platform    *string
	model       *string
	os          *string
	osVersion   *string
	userAgent   *string
	reportedAt  *time.Time
}

func normalizeTgDeviceInfo(req pojo.TgDeviceInfoReq, userAgent string, requireFingerprint bool) (normalizedTgDeviceInfo, error) {
	var result normalizedTgDeviceInfo

	fingerprint := strings.TrimSpace(req.DeviceFingerprint)
	if fingerprint != "" || requireFingerprint {
		hash, err := normalizeRegisterDeviceFingerprint(fingerprint)
		if err != nil {
			return result, err
		}
		result.fingerprint = &hash
	}

	platform := strings.ToLower(strings.TrimSpace(req.DevicePlatform))
	if platform != "" {
		switch platform {
		case "android", "ios", "windows", "macos", "linux", "chromeos":
			result.platform = &platform
		default:
			return result, errors.New("device_platform_invalid")
		}
	}

	model := strings.TrimSpace(req.DeviceModel)
	if len([]rune(model)) > 128 {
		return result, errors.New("device_model_too_long")
	}
	if model != "" {
		result.model = &model
	}

	osName := strings.TrimSpace(req.DeviceOS)
	if len([]rune(osName)) > 32 {
		return result, errors.New("device_os_too_long")
	}
	if osName != "" {
		result.os = &osName
	}

	osVersion := strings.TrimSpace(req.DeviceOSVersion)
	if len([]rune(osVersion)) > 64 {
		return result, errors.New("device_os_version_too_long")
	}
	if osVersion != "" {
		result.osVersion = &osVersion
	}

	userAgent = strings.TrimSpace(userAgent)
	if userAgent != "" {
		userAgent = truncateRunes(userAgent, 1024)
		result.userAgent = &userAgent
	}

	if !result.empty() {
		now := time.Now()
		result.reportedAt = &now
	}
	return result, nil
}

func (info normalizedTgDeviceInfo) empty() bool {
	return info.fingerprint == nil &&
		info.platform == nil &&
		info.model == nil &&
		info.os == nil &&
		info.osVersion == nil &&
		info.userAgent == nil
}

func (info normalizedTgDeviceInfo) apply(user *pojo.TgUser) {
	if info.fingerprint != nil {
		user.DeviceFingerprint = info.fingerprint
	}
	if info.platform != nil {
		user.DevicePlatform = info.platform
	}
	if info.model != nil {
		user.DeviceModel = info.model
	}
	if info.os != nil {
		user.DeviceOS = info.os
	}
	if info.osVersion != nil {
		user.DeviceOSVersion = info.osVersion
	}
	if info.userAgent != nil {
		user.DeviceUserAgent = info.userAgent
	}
	if info.reportedAt != nil {
		user.DeviceReportedAt = info.reportedAt
	}
}

func (info normalizedTgDeviceInfo) updates() map[string]any {
	updates := map[string]any{}
	if info.fingerprint != nil {
		updates["device_fingerprint"] = *info.fingerprint
	}
	if info.platform != nil {
		updates["device_platform"] = *info.platform
	}
	if info.model != nil {
		updates["device_model"] = *info.model
	}
	if info.os != nil {
		updates["device_os"] = *info.os
	}
	if info.osVersion != nil {
		updates["device_os_version"] = *info.osVersion
	}
	if info.userAgent != nil {
		updates["device_user_agent"] = *info.userAgent
	}
	if info.reportedAt != nil {
		updates["device_reported_at"] = *info.reportedAt
	}
	return updates
}

func updateTgUserDeviceInfo(db *gorm.DB, userID int64, req pojo.TgDeviceInfoReq, userAgent string) error {
	info, err := normalizeTgDeviceInfo(req, userAgent, false)
	if err != nil {
		return err
	}
	if info.empty() {
		return nil
	}
	result := db.Model(&pojo.TgUser{}).
		Where("id = ? AND status <> ?", userID, -1).
		Updates(info.updates())
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("user_not_found")
	}
	return nil
}

// ReportTgUserDeviceInfo 保存已登录用户当天首次打开 H5 时上报的最近设备信息。
func ReportTgUserDeviceInfo(db *gorm.DB, userID int64, req pojo.TgDeviceInfoReq, userAgent string) error {
	return updateTgUserDeviceInfo(db, userID, req, userAgent)
}
