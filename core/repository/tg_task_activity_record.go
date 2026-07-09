package repository

import (
	"BaseGoUni/core/pojo"
	"errors"
	"strings"

	"gorm.io/gorm"
)

// GetTgTaskActivityRecords 任务活动领取记录列表（分页）
func GetTgTaskActivityRecords(db *gorm.DB, search pojo.TgTaskActivityRecordSearch) (result pojo.TgTaskActivityRecordResp) {
	if search.PageSize <= 0 {
		search.PageSize = 10
	}
	if search.CurrentPage < 0 {
		search.CurrentPage = 0
	}

	query := db.Table(pojo.TgTaskActivityRecordTableName + " AS r").
		Joins("LEFT JOIN " + pojo.TgUserTableName + " AS u ON u.id = r.user_id")

	if search.UserID > 0 {
		query = query.Where("r.user_id = ?", search.UserID)
	}
	if uid := strings.TrimSpace(search.Uid); uid != "" {
		query = query.Where("u.uid = ?", uid)
	}
	if search.ConfigID > 0 {
		query = query.Where("r.config_id = ?", search.ConfigID)
	}
	if title := strings.TrimSpace(search.Title); title != "" {
		query = query.Where("r.title LIKE ?", "%"+title+"%")
	}
	if levelCode := strings.TrimSpace(search.LevelCode); levelCode != "" {
		query = query.Where("r.level_code = ?", levelCode)
	}
	if search.Status != nil {
		query = query.Where("r.status = ?", *search.Status)
	}

	query.Count(&result.Total)
	query.Select(`
		r.id, r.created_at, r.updated_at, r.config_id, r.user_id, u.uid,
		r.title, r.sub_title, r.level_code, r.level_name, r.task_type,
		r.required_invite_count, r.required_recharge_amount, r.duration_minutes,
		r.reward_amount, r.reward_currency, r.reward_target,
		r.progress_invite_count, r.progress_recharge_count, r.progress_recharge_amount,
		r.status, r.claimed_at, r.deadline_at, r.completed_at, r.rewarded_at,
		r.rebate_record_id, r.remark`).
		Order("r.id desc").
		Limit(search.PageSize).
		Offset(search.PageSize * search.CurrentPage).
		Scan(&result.List)

	result.PageSize = search.PageSize
	result.CurrentPage = search.CurrentPage
	return result
}

// GetTgTaskActivityRecordById 根据ID获取任务活动领取记录
func GetTgTaskActivityRecordById(db *gorm.DB, id int64) (result pojo.TgTaskActivityRecordBack, err error) {
	if id <= 0 {
		return result, errors.New("record_not_found")
	}
	db.Table(pojo.TgTaskActivityRecordTableName+" AS r").
		Select(`
			r.id, r.created_at, r.updated_at, r.config_id, r.user_id, u.uid,
			r.title, r.sub_title, r.level_code, r.level_name, r.task_type,
			r.required_invite_count, r.required_recharge_amount, r.duration_minutes,
			r.reward_amount, r.reward_currency, r.reward_target,
			r.progress_invite_count, r.progress_recharge_count, r.progress_recharge_amount,
			r.status, r.claimed_at, r.deadline_at, r.completed_at, r.rewarded_at,
			r.rebate_record_id, r.remark`).
		Joins("LEFT JOIN "+pojo.TgUserTableName+" AS u ON u.id = r.user_id").
		Where("r.id = ?", id).
		Scan(&result)
	if result.ID == 0 {
		return result, errors.New("record_not_found")
	}
	return result, nil
}
