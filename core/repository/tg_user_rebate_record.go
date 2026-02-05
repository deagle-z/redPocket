package repository

import (
	"BaseGoUni/core/pojo"
	"errors"
	"github.com/jinzhu/copier"
	"gorm.io/gorm"
)

// GetTgUserRebateRecords 反水/返佣记录列表（分页）
func GetTgUserRebateRecords(db *gorm.DB, search pojo.TgUserRebateRecordSearch) (result pojo.TgUserRebateRecordResp) {
	var records []pojo.TgUserRebateRecord
	query := db.Model(&pojo.TgUserRebateRecord{})

	if search.TenantId != nil {
		query = query.Where("tenant_id = ?", *search.TenantId)
	}
	if search.SubUserId > 0 {
		query = query.Where("sub_user_id = ?", search.SubUserId)
	}
	if search.ParentUserId > 0 {
		query = query.Where("parent_user_id = ?", search.ParentUserId)
	}
	if search.SourceType != nil {
		query = query.Where("source_type = ?", *search.SourceType)
	}
	if search.SourceOrderId != "" {
		query = query.Where("source_order_id = ?", search.SourceOrderId)
	}
	if search.Status != nil {
		query = query.Where("status = ?", *search.Status)
	}
	if search.IdempotencyKey != "" {
		query = query.Where("idempotency_key = ?", search.IdempotencyKey)
	}

	query.Count(&result.Total)
	query = query.Order("id desc").Limit(search.PageSize).Offset(search.PageSize * search.CurrentPage)
	query.Find(&records)

	for _, record := range records {
		var temp pojo.TgUserRebateRecordBack
		_ = copier.Copy(&temp, &record)
		result.List = append(result.List, temp)
	}

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
			return result, errors.New("更新的数据不存在")
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
		return result, errors.New("删除的数据不存在")
	}
	err = db.Delete(&dbRecord).Error
	if err != nil {
		return result, err
	}
	return "success", nil
}

// GetTgUserRebateRecordById 根据ID获取反水/返佣记录
func GetTgUserRebateRecordById(db *gorm.DB, id int64) (result pojo.TgUserRebateRecordBack, err error) {
	var dbRecord pojo.TgUserRebateRecord
	db.Where("id = ?", id).First(&dbRecord)
	if dbRecord.ID == 0 {
		return result, errors.New("数据不存在")
	}
	_ = copier.Copy(&result, &dbRecord)
	return result, nil
}
