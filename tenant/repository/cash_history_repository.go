package repository

import (
	"BaseGoUni/core/pojo"

	"github.com/jinzhu/copier"
	"gorm.io/gorm"
)

func GetCashHistoryList(db *gorm.DB, tenantID int64, search pojo.CashHistorySearch) (result pojo.CashHistoryPage) {
	var cashHistoryList []pojo.CashHistory

	tenantUserIDs := db.Model(&pojo.TgUser{}).
		Select("id").
		Where("tenant_id = ?", tenantID)

	query := db.Table(pojo.AllCashHistoryShardingName).
		Where("user_id in (?)", tenantUserIDs)

	if search.UserId > 0 {
		query = query.Where("user_id = ?", search.UserId)
	}
	if search.CashMark != "" {
		query = query.Where("cash_mark LIKE ?", "%"+search.CashMark+"%")
	}

	query.Count(&result.Total)
	query = query.Order("id desc").Limit(search.PageSize).Offset(search.PageSize * search.CurrentPage)
	query.Find(&cashHistoryList)

	for _, history := range cashHistoryList {
		var tempResp pojo.CashHistoryResp
		_ = copier.Copy(&tempResp, &history)
		result.List = append(result.List, tempResp)
	}

	result.PageSize = search.PageSize
	result.CurrentPage = search.CurrentPage
	return result
}
