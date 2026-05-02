package repository

import (
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/utils"
	"errors"
	"github.com/jinzhu/copier"
	"gorm.io/gorm"
)

// CreatePlatformProfitLedgerIfAbsent 幂等写入平台盈利流水
func CreatePlatformProfitLedgerIfAbsent(tx *gorm.DB, entity pojo.PlatformProfitLedger) error {
	entity.IncomeAmount = utils.Truncate2(entity.IncomeAmount)
	entity.ExpenseAmount = utils.Truncate2(entity.ExpenseAmount)
	entity.RebateAmount = utils.Truncate2(entity.RebateAmount)
	if entity.ActualIncomeAmount == nil {
		actual := utils.Truncate2(entity.IncomeAmount - entity.RebateAmount)
		if actual < 0 {
			actual = 0
		}
		entity.ActualIncomeAmount = &actual
	} else {
		actual := utils.Truncate2(*entity.ActualIncomeAmount)
		if actual < 0 {
			actual = 0
		}
		entity.ActualIncomeAmount = &actual
	}
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
	query := db.Table(pojo.PlatformProfitLedgerTableName + " AS p").
		Joins("LEFT JOIN " + pojo.TgUserTableName + " AS u ON u.id = p.user_id")

	if search.TenantId != nil {
		query = query.Where("p.tenant_id = ?", *search.TenantId)
	}
	if search.UserId > 0 {
		query = query.Where("p.user_id = ?", search.UserId)
	}
	if search.SourceType != "" {
		query = query.Where("p.source_type = ?", search.SourceType)
	}
	if search.SourceId != "" {
		query = query.Where("p.source_id LIKE ?", "%"+search.SourceId+"%")
	}
	if search.MinNet != nil {
		minNet := utils.Truncate2(*search.MinNet)
		query = query.Where("(p.income_amount - p.expense_amount) >= ?", minNet)
	}
	if search.MaxNet != nil {
		maxNet := utils.Truncate2(*search.MaxNet)
		query = query.Where("(p.income_amount - p.expense_amount) <= ?", maxNet)
	}

	query.Count(&result.Total)
	query.Select(`
		p.id, p.created_at, p.updated_at, p.tenant_id, p.user_id, u.uid AS user_uid,
		COALESCE(NULLIF(u.first_name, ''), NULLIF(u.username, ''), '') AS user_name,
		p.source_channel_id, p.source_type, p.source_id, p.income_amount, p.expense_amount,
		p.rebate_amount, COALESCE(p.actual_income_amount, p.income_amount - COALESCE(p.rebate_amount, 0)) AS actual_income_amount,
		(p.income_amount - p.expense_amount) AS net_amount, p.remark`).
		Order("p.id desc").
		Limit(search.PageSize).
		Offset(search.PageSize * search.CurrentPage).
		Scan(&result.List)

	for i := range result.List {
		result.List[i].IncomeAmount = utils.Truncate2(result.List[i].IncomeAmount)
		result.List[i].ExpenseAmount = utils.Truncate2(result.List[i].ExpenseAmount)
		result.List[i].RebateAmount = utils.Truncate2(result.List[i].RebateAmount)
		result.List[i].ActualIncomeAmount = utils.Truncate2(result.List[i].ActualIncomeAmount)
		result.List[i].NetAmount = utils.Truncate2(result.List[i].NetAmount)
	}

	result.PageSize = search.PageSize
	result.CurrentPage = search.CurrentPage
	return result
}

// SetPlatformProfitLedger 创建或更新平台盈利流水
func SetPlatformProfitLedger(db *gorm.DB, req pojo.PlatformProfitLedgerSet) (result pojo.PlatformProfitLedgerBack, err error) {
	req.IncomeAmount = utils.Truncate2(req.IncomeAmount)
	req.ExpenseAmount = utils.Truncate2(req.ExpenseAmount)
	req.RebateAmount = utils.Truncate2(req.RebateAmount)
	req.ActualIncomeAmount = utils.Truncate2(req.ActualIncomeAmount)
	if req.ActualIncomeAmount <= 0 && req.IncomeAmount > 0 {
		req.ActualIncomeAmount = utils.Truncate2(req.IncomeAmount - req.RebateAmount)
		if req.ActualIncomeAmount < 0 {
			req.ActualIncomeAmount = 0
		}
	}
	var entity pojo.PlatformProfitLedger
	if req.ID > 0 {
		db.Where("id = ?", req.ID).First(&entity)
		if entity.ID == 0 {
			return result, errors.New("record_not_found_update")
		}
		_ = copier.Copy(&entity, &req)
		entity.ActualIncomeAmount = &req.ActualIncomeAmount
		err = db.Save(&entity).Error
	} else {
		_ = copier.Copy(&entity, &req)
		entity.ActualIncomeAmount = &req.ActualIncomeAmount
		err = db.Create(&entity).Error
	}
	if err != nil {
		return result, err
	}

	db.Where("id = ?", entity.ID).First(&entity)
	_ = copier.Copy(&result, &entity)
	result.IncomeAmount = utils.Truncate2(result.IncomeAmount)
	result.ExpenseAmount = utils.Truncate2(result.ExpenseAmount)
	result.RebateAmount = utils.Truncate2(entity.RebateAmount)
	result.ActualIncomeAmount = resolvePlatformProfitActualIncome(entity)
	result.NetAmount = utils.Truncate2(result.NetAmount)
	return result, nil
}

// DelPlatformProfitLedger 删除平台盈利流水
func DelPlatformProfitLedger(db *gorm.DB, id int64) (result string, err error) {
	var entity pojo.PlatformProfitLedger
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

// GetPlatformProfitLedgerById 根据ID获取平台盈利流水
func GetPlatformProfitLedgerById(db *gorm.DB, id int64) (result pojo.PlatformProfitLedgerBack, err error) {
	var entity pojo.PlatformProfitLedger
	db.Where("id = ?", id).First(&entity)
	if entity.ID == 0 {
		return result, errors.New("record_not_found")
	}
	_ = copier.Copy(&result, &entity)
	result.IncomeAmount = utils.Truncate2(result.IncomeAmount)
	result.ExpenseAmount = utils.Truncate2(result.ExpenseAmount)
	result.RebateAmount = utils.Truncate2(entity.RebateAmount)
	result.ActualIncomeAmount = resolvePlatformProfitActualIncome(entity)
	result.NetAmount = utils.Truncate2(result.NetAmount)
	return result, nil
}

func resolvePlatformProfitActualIncome(item pojo.PlatformProfitLedger) float64 {
	if item.ActualIncomeAmount != nil {
		return utils.Truncate2(*item.ActualIncomeAmount)
	}
	actual := utils.Truncate2(item.IncomeAmount - item.RebateAmount)
	if actual < 0 {
		return 0
	}
	return actual
}
