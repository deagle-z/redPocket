package repository

import (
	"BaseGoUni/core/pojo"
	"errors"
	"github.com/jinzhu/copier"
	"gorm.io/gorm"
)

// CreatePlatformProfitLedgerIfAbsent 幂等写入平台盈利流水
func CreatePlatformProfitLedgerIfAbsent(tx *gorm.DB, entity pojo.PlatformProfitLedger) error {
	var existing pojo.PlatformProfitLedger
	err := tx.Where("source_type = ? AND source_id = ?", entity.SourceType, entity.SourceId).First(&existing).Error
	if err == nil && existing.ID > 0 {
		return nil
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	return tx.Create(&entity).Error
}

// GetPlatformProfitLedgers 平台盈利流水列表（分页）
func GetPlatformProfitLedgers(db *gorm.DB, search pojo.PlatformProfitLedgerSearch) (result pojo.PlatformProfitLedgerResp) {
	var list []pojo.PlatformProfitLedger
	query := db.Model(&pojo.PlatformProfitLedger{})

	if search.TenantId != nil {
		query = query.Where("tenant_id = ?", *search.TenantId)
	}
	if search.UserId > 0 {
		query = query.Where("user_id = ?", search.UserId)
	}
	if search.SourceType != "" {
		query = query.Where("source_type = ?", search.SourceType)
	}
	if search.SourceId != "" {
		query = query.Where("source_id LIKE ?", "%"+search.SourceId+"%")
	}
	if search.MinNet != nil {
		query = query.Where("(income_amount - expense_amount) >= ?", *search.MinNet)
	}
	if search.MaxNet != nil {
		query = query.Where("(income_amount - expense_amount) <= ?", *search.MaxNet)
	}

	query.Count(&result.Total)
	query = query.Order("id desc").Limit(search.PageSize).Offset(search.PageSize * search.CurrentPage)
	query.Find(&list)

	for _, item := range list {
		var temp pojo.PlatformProfitLedgerBack
		_ = copier.Copy(&temp, &item)
		result.List = append(result.List, temp)
	}

	result.PageSize = search.PageSize
	result.CurrentPage = search.CurrentPage
	return result
}

// SetPlatformProfitLedger 创建或更新平台盈利流水
func SetPlatformProfitLedger(db *gorm.DB, req pojo.PlatformProfitLedgerSet) (result pojo.PlatformProfitLedgerBack, err error) {
	var entity pojo.PlatformProfitLedger
	if req.ID > 0 {
		db.Where("id = ?", req.ID).First(&entity)
		if entity.ID == 0 {
			return result, errors.New("更新的数据不存在")
		}
		_ = copier.Copy(&entity, &req)
		err = db.Save(&entity).Error
	} else {
		_ = copier.Copy(&entity, &req)
		err = db.Create(&entity).Error
	}
	if err != nil {
		return result, err
	}

	db.Where("id = ?", entity.ID).First(&entity)
	_ = copier.Copy(&result, &entity)
	return result, nil
}

// DelPlatformProfitLedger 删除平台盈利流水
func DelPlatformProfitLedger(db *gorm.DB, id int64) (result string, err error) {
	var entity pojo.PlatformProfitLedger
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

// GetPlatformProfitLedgerById 根据ID获取平台盈利流水
func GetPlatformProfitLedgerById(db *gorm.DB, id int64) (result pojo.PlatformProfitLedgerBack, err error) {
	var entity pojo.PlatformProfitLedger
	db.Where("id = ?", id).First(&entity)
	if entity.ID == 0 {
		return result, errors.New("数据不存在")
	}
	_ = copier.Copy(&result, &entity)
	return result, nil
}
