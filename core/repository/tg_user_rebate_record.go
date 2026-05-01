package repository

import (
	"BaseGoUni/core/pojo"
	"errors"

	"github.com/jinzhu/copier"
	"gorm.io/gorm"
)

// GetTgUserRebateRecords 反水/返佣记录列表（分页）
func GetTgUserRebateRecords(db *gorm.DB, search pojo.TgUserRebateRecordSearch) (result pojo.TgUserRebateRecordResp) {
	if search.PageSize <= 0 {
		search.PageSize = 10
	}
	if search.CurrentPage < 0 {
		search.CurrentPage = 0
	}

	query := db.Table(pojo.TgUserRebateRecordTableName + " AS r").
		Joins("LEFT JOIN " + pojo.TgUserTableName + " AS sub_user ON sub_user.id = r.sub_user_id").
		Joins("LEFT JOIN " + pojo.TgUserTableName + " AS parent_user ON parent_user.id = r.parent_user_id")

	if search.TenantId != nil {
		query = query.Where("r.tenant_id = ?", *search.TenantId)
	}
	if search.SubUserId > 0 {
		query = query.Where("r.sub_user_id = ?", search.SubUserId)
	}
	if search.SubUid != "" {
		query = query.Where("sub_user.uid = ?", search.SubUid)
	}
	if search.ParentUserId > 0 {
		query = query.Where("r.parent_user_id = ?", search.ParentUserId)
	}
	if search.ParentUid != "" {
		query = query.Where("parent_user.uid = ?", search.ParentUid)
	}
	if search.SourceType != nil {
		query = query.Where("r.source_type = ?", *search.SourceType)
	}
	if search.SourceOrderId != "" {
		query = query.Where("r.source_order_id = ?", search.SourceOrderId)
	}
	if search.Status != nil {
		query = query.Where("r.status = ?", *search.Status)
	}
	if search.IdempotencyKey != "" {
		query = query.Where("r.idempotency_key = ?", search.IdempotencyKey)
	}

	query.Count(&result.Total)
	query.Select(`
		r.id, r.created_at, r.updated_at, r.tenant_id, r.sub_user_id, sub_user.uid AS sub_uid,
		r.parent_user_id, parent_user.uid AS parent_uid, r.source_channel_id, r.source_type,
		r.source_order_id, r.source_amount, r.rebate_rate, r.rebate_amount, r.currency, r.status,
		r.settled_at, r.idempotency_key, r.remark`).
		Order("r.id desc").
		Limit(search.PageSize).
		Offset(search.PageSize * search.CurrentPage).
		Scan(&result.List)

	result.PageSize = search.PageSize
	result.CurrentPage = search.CurrentPage
	return result
}

// SetTgUserRebateRecord 创建或更新反水/返佣记录
func SetTgUserRebateRecord(db *gorm.DB, req pojo.TgUserRebateRecordSet) (result pojo.TgUserRebateRecordBack, err error) {
	var dbRecord pojo.TgUserRebateRecord
	if req.ID > 0 {
		db.Where("id = ?", req.ID).First(&dbRecord)
		if dbRecord.ID == 0 {
			return result, errors.New("record_not_found_update")
		}
		_ = copier.Copy(&dbRecord, &req)
		err = db.Save(&dbRecord).Error
	} else {
		_ = copier.Copy(&dbRecord, &req)
		err = db.Create(&dbRecord).Error
	}
	if err != nil {
		return result, err
	}
	_ = copier.Copy(&result, &dbRecord)
	return result, nil
}

// DelTgUserRebateRecord 删除反水/返佣记录
func DelTgUserRebateRecord(db *gorm.DB, id int64) (result string, err error) {
	var dbRecord pojo.TgUserRebateRecord
	db.Where("id = ?", id).First(&dbRecord)
	if dbRecord.ID == 0 {
		return result, errors.New("record_not_found_delete")
	}
	err = db.Delete(&dbRecord).Error
	if err != nil {
		return result, err
	}
	return "success", nil
}

// GetTgUserRebateRecordById 根据ID获取反水/返佣记录
func GetTgUserRebateRecordById(db *gorm.DB, id int64) (result pojo.TgUserRebateRecordBack, err error) {
	db.Table(pojo.TgUserRebateRecordTableName+" AS r").
		Select(`
			r.id, r.created_at, r.updated_at, r.tenant_id, r.sub_user_id, sub_user.uid AS sub_uid,
			r.parent_user_id, parent_user.uid AS parent_uid, r.source_channel_id, r.source_type,
			r.source_order_id, r.source_amount, r.rebate_rate, r.rebate_amount, r.currency, r.status,
			r.settled_at, r.idempotency_key, r.remark`).
		Joins("LEFT JOIN "+pojo.TgUserTableName+" AS sub_user ON sub_user.id = r.sub_user_id").
		Joins("LEFT JOIN "+pojo.TgUserTableName+" AS parent_user ON parent_user.id = r.parent_user_id").
		Where("r.id = ?", id).
		Scan(&result)
	if result.ID == 0 {
		return result, errors.New("record_not_found")
	}
	return result, nil
}
