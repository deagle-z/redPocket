package repository

import (
	"BaseGoUni/core/pojo"
	"gorm.io/gorm"
	"time"
)

// GetUserLotteryRecords 抽奖记录分页列表
func GetUserLotteryRecords(db *gorm.DB, search pojo.UserLotteryRecordSearch) (result pojo.UserLotteryRecordResp) {
	query := db.Model(&pojo.UserLotteryRecord{})

	if search.TenantId > 0 {
		query = query.Where("tenant_id = ?", search.TenantId)
	}
	if search.UserId > 0 {
		query = query.Where("user_id = ?", search.UserId)
	}
	if search.PoolId > 0 {
		query = query.Where("pool_id = ?", search.PoolId)
	}
	if search.Status != nil {
		query = query.Where("status = ?", *search.Status)
	}
	if search.StartTime > 0 {
		query = query.Where("created_at >= ?", time.Unix(search.StartTime, 0))
	}
	if search.EndTime > 0 {
		query = query.Where("created_at <= ?", time.Unix(search.EndTime, 0))
	}

	query.Count(&result.Total)
	query.Order("id desc").Limit(search.PageSize).Offset(search.PageSize * search.CurrentPage).Find(&result.List)

	result.PageSize = search.PageSize
	result.CurrentPage = search.CurrentPage
	return result
}

// GetUserLotteryRecordById 根据ID查询
func GetUserLotteryRecordById(db *gorm.DB, id int64) (pojo.UserLotteryRecord, error) {
	var entity pojo.UserLotteryRecord
	err := db.Where("id = ?", id).First(&entity).Error
	return entity, err
}

// CreateLotteryRecord 创建抽奖记录
func CreateLotteryRecord(db *gorm.DB, record pojo.UserLotteryRecord) (pojo.UserLotteryRecord, error) {
	err := db.Create(&record).Error
	return record, err
}

// UpdateLotteryRecordStatus 更新抽奖记录状态
func UpdateLotteryRecordStatus(db *gorm.DB, id int64, status int8) error {
	return db.Model(&pojo.UserLotteryRecord{}).
		Where("id = ?", id).
		Update("status", status).Error
}

// LotteryHistoryItem 抽奖历史条目（名称+金额）
type LotteryHistoryItem struct {
	Name        string  `json:"name"`
	AwardAmount float64 `json:"awardAmount"`
}

// GetLotteryHistory 查询抽奖历史（JOIN tg_user 取名称），只返回中奖记录
func GetLotteryHistory(db *gorm.DB, userId int64, limit int) ([]LotteryHistoryItem, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	var items []LotteryHistoryItem
	err := db.Table(pojo.UserLotteryRecord{}.TableName()+" r").
		Select("COALESCE(u.first_name, u.username, CONCAT('user_', r.user_id)) AS name, r.award_amount").
		Joins("LEFT JOIN "+pojo.TgUserTableName+" u ON u.id = r.user_id").
		Where("r.user_id = ? AND r.award_amount > 0", userId).
		Order("r.id DESC").
		Limit(limit).
		Scan(&items).Error
	return items, err
}

// CountUserPendingLottery 查询用户待结算抽奖记录数
func CountUserPendingLottery(db *gorm.DB, userId int64) (int64, error) {
	var count int64
	err := db.Model(&pojo.UserLotteryRecord{}).
		Where("user_id = ? AND status = ?", userId, pojo.LotteryStatusPending).
		Count(&count).Error
	return count, err
}
