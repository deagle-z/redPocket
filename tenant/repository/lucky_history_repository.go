package repository

import (
	"BaseGoUni/core/pojo"
	"github.com/jinzhu/copier"
	"gorm.io/gorm"
)

func GetLuckyHistoryList(db *gorm.DB, tenantID int64, search pojo.LuckyHistorySearch) (result pojo.LuckyHistoryResp) {
	var historyList []pojo.LuckyHistory
	query := db.Model(&pojo.LuckyHistory{}).Where("tenant_id = ?", tenantID)
	if search.LuckyID > 0 {
		query = query.Where("lucky_id = ?", search.LuckyID)
	}
	if search.UserID > 0 {
		query = query.Where("user_id = ?", search.UserID)
	}
	query.Count(&result.Total)
	query = query.Order("id desc").Limit(search.PageSize).Offset(search.PageSize * search.CurrentPage)
	query.Find(&historyList)
	for _, history := range historyList {
		var tempHistory pojo.LuckyHistoryBack
		_ = copier.Copy(&tempHistory, &history)
		result.List = append(result.List, tempHistory)
	}
	result.PageSize = search.PageSize
	result.CurrentPage = search.CurrentPage
	return result
}

func GetLuckyHistoryByLuckyID(db *gorm.DB, tenantID int64, luckyID int64) (result []pojo.LuckyHistory, err error) {
	err = db.Where("tenant_id = ? and lucky_id = ?", tenantID, luckyID).Order("id asc").Find(&result).Error
	return result, err
}
