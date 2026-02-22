package repository

import (
	"BaseGoUni/core/pojo"
	"errors"
	"github.com/jinzhu/copier"
	"gorm.io/gorm"
)

func GetTgUserRebateRecords(db *gorm.DB, tenantID int64, search pojo.TgUserRebateRecordSearch) (result pojo.TgUserRebateRecordResp) {
	var records []pojo.TgUserRebateRecord
	query := db.Model(&pojo.TgUserRebateRecord{}).Where("tenant_id = ?", tenantID)
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

func GetTgUserRebateRecordByID(db *gorm.DB, tenantID int64, id int64) (result pojo.TgUserRebateRecordBack, err error) {
	var dbRecord pojo.TgUserRebateRecord
	db.Where("id = ? and tenant_id = ?", id, tenantID).First(&dbRecord)
	if dbRecord.ID == 0 {
		return result, errors.New("数据不存在")
	}
	_ = copier.Copy(&result, &dbRecord)
	return result, nil
}

func SetTgUserRebateRecord(db *gorm.DB, tenantID int64, req pojo.TgUserRebateRecordSet) (result pojo.TgUserRebateRecordBack, err error) {
	req.TenantId = &tenantID
	var dbRecord pojo.TgUserRebateRecord
	if req.ID > 0 {
		db.Where("id = ? and tenant_id = ?", req.ID, tenantID).First(&dbRecord)
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

func DelTgUserRebateRecord(db *gorm.DB, tenantID int64, id int64) (result string, err error) {
	var dbRecord pojo.TgUserRebateRecord
	db.Where("id = ? and tenant_id = ?", id, tenantID).First(&dbRecord)
	if dbRecord.ID == 0 {
		return result, errors.New("删除的数据不存在")
	}
	err = db.Delete(&dbRecord).Error
	if err != nil {
		return result, err
	}
	return "success", nil
}
