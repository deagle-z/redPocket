package repository

import (
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/utils"
	"errors"
	"strings"
	"time"

	"github.com/jinzhu/copier"
	"gorm.io/gorm"
)

// GetSysSourceChannels 投流来源渠道列表（分页）
func GetSysSourceChannels(db *gorm.DB, search pojo.SysSourceChannelSearch, linkBaseURL string) (result pojo.SysSourceChannelResp) {
	var list []pojo.SysSourceChannel
	query := db.Model(&pojo.SysSourceChannel{})

	if search.TenantId > 0 {
		query = query.Where("tenant_id = ?", search.TenantId)
	}
	if search.ChannelCode != "" {
		query = query.Where("channel_code LIKE ?", "%"+NormalizeSourceChannelCode(search.ChannelCode)+"%")
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
		result.List = append(result.List, buildSysSourceChannelBack(db, item, linkBaseURL))
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
			return result, errors.New("record_not_found_update")
		}
		req.ChannelCode = entity.ChannelCode
		_ = copier.Copy(&entity, &req)
		entity.ChannelCode = NormalizeSourceChannelCode(entity.ChannelCode)
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
	result.LinkURL = BuildSourceChannelLinkURL("", entity.ChannelCode)
	stats, _ := GetSysSourceChannelStats(db, entity.ID)
	result.Stats = stats
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
	return "", errors.New("source_channel_code_generate_failed")
}

// DelSysSourceChannel 删除投流来源渠道
func DelSysSourceChannel(db *gorm.DB, id int64) (result string, err error) {
	var entity pojo.SysSourceChannel
	db.Where("id = ?", id).First(&entity)
	if entity.ID == 0 {
		return result, errors.New("record_not_found_delete")
	}
	err = db.Delete(&entity).Error
	if err != nil {
		return result, err
	}
	return "success", nil
}

// GetSysSourceChannelById 根据ID获取投流来源渠道
func GetSysSourceChannelById(db *gorm.DB, id int64, linkBaseURL string) (result pojo.SysSourceChannelBack, err error) {
	var entity pojo.SysSourceChannel
	db.Where("id = ?", id).First(&entity)
	if entity.ID == 0 {
		return result, errors.New("record_not_found")
	}
	return buildSysSourceChannelBack(db, entity, linkBaseURL), nil
}

func buildSysSourceChannelBack(db *gorm.DB, entity pojo.SysSourceChannel, linkBaseURL string) pojo.SysSourceChannelBack {
	var result pojo.SysSourceChannelBack
	_ = copier.Copy(&result, &entity)
	result.LinkURL = BuildSourceChannelLinkURL(linkBaseURL, entity.ChannelCode)
	stats, _ := GetSysSourceChannelStats(db, entity.ID)
	result.Stats = stats
	return result
}

func GetSysSourceChannelStats(db *gorm.DB, channelID int64) (pojo.SysSourceChannelStatsBack, error) {
	var result pojo.SysSourceChannelStatsBack
	if db == nil || channelID <= 0 {
		return result, errors.New("invalid_params")
	}

	var channel pojo.SysSourceChannel
	if err := db.Select("id").Where("id = ?", channelID).First(&channel).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return result, errors.New("record_not_found")
		}
		return result, err
	}

	startOfDay, endOfDay := todayRange()

	if err := db.Model(&pojo.TgUser{}).
		Where("source_channel_id = ?", channelID).
		Count(&result.RegisterUsers).Error; err != nil {
		return result, err
	}
	if err := db.Model(&pojo.TgUser{}).
		Where("source_channel_id = ? AND created_at >= ? AND created_at < ?", channelID, startOfDay, endOfDay).
		Count(&result.TodayRegisterUsers).Error; err != nil {
		return result, err
	}
	if err := db.Model(&pojo.RechargeOrder{}).
		Where("source_channel_id = ? AND status = ?", channelID, 1).
		Select("COUNT(DISTINCT user_id)").
		Scan(&result.RechargeUsers).Error; err != nil {
		return result, err
	}
	if err := db.Model(&pojo.RechargeOrder{}).
		Where("source_channel_id = ? AND status = ? AND pay_time >= ? AND pay_time < ?", channelID, 1, startOfDay, endOfDay).
		Select("COUNT(DISTINCT user_id)").
		Scan(&result.TodayRechargeUsers).Error; err != nil {
		return result, err
	}
	if err := db.Model(&pojo.RechargeOrder{}).
		Where("source_channel_id = ? AND status = ?", channelID, 1).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&result.TotalRechargeAmount).Error; err != nil {
		return result, err
	}
	if err := db.Model(&pojo.RechargeOrder{}).
		Where("source_channel_id = ? AND status = ? AND pay_time >= ? AND pay_time < ?", channelID, 1, startOfDay, endOfDay).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&result.TodayRechargeAmount).Error; err != nil {
		return result, err
	}
	if err := db.Model(&pojo.WithdrawOrderBr{}).
		Where("source_channel_id = ? AND status = ?", channelID, 3).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&result.TotalWithdrawAmount).Error; err != nil {
		return result, err
	}
	if err := db.Model(&pojo.WithdrawOrderBr{}).
		Where("source_channel_id = ? AND status = ? AND paid_at >= ? AND paid_at < ?", channelID, 3, startOfDay, endOfDay).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&result.TodayWithdrawAmount).Error; err != nil {
		return result, err
	}
	if err := db.Model(&pojo.AttributionEvent{}).
		Where("source_channel_id = ?", channelID).
		Count(&result.EventCount).Error; err != nil {
		return result, err
	}
	if err := db.Model(&pojo.AttributionEvent{}).
		Where("source_channel_id = ?", channelID).
		Select("COUNT(DISTINCT NULLIF(visitor_id, ''))").
		Scan(&result.UniqueVisitors).Error; err != nil {
		return result, err
	}
	if err := db.Model(&pojo.AttributionEvent{}).
		Where("source_channel_id = ?", channelID).
		Select("COUNT(DISTINCT user_id)").
		Scan(&result.UniqueUsers).Error; err != nil {
		return result, err
	}
	return result, nil
}

func todayRange() (time.Time, time.Time) {
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	return startOfDay, startOfDay.Add(24 * time.Hour)
}
