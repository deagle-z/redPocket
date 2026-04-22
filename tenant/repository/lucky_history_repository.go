package repository

import (
	"BaseGoUni/core/pojo"
	"github.com/jinzhu/copier"
	"gorm.io/gorm"
	"strings"
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

func GetLuckyHistoryUserFlowList(db *gorm.DB, tenantID int64, search pojo.LuckyHistoryUserFlowSearch) (result pojo.LuckyHistoryUserFlowResp) {
	type luckyHistoryUserFlowRow struct {
		UserID     int64   `gorm:"column:user_id"`
		Avatar     *string `gorm:"column:avatar"`
		FirstName  string  `gorm:"column:first_name"`
		FlowAmount float64 `gorm:"column:flow_amount"`
	}

	query := db.Model(&pojo.LuckyHistory{}).Where("tenant_id = ?", tenantID)
	if search.UserID > 0 {
		query = query.Where("user_id = ?", search.UserID)
	}

	_ = query.Distinct("user_id").Count(&result.Total).Error

	var rows []luckyHistoryUserFlowRow
	_ = query.
		Joins("left join tg_user on tg_user.id = lucky_history.user_id").
		Select("lucky_history.user_id, MAX(tg_user.avatar) as avatar, MAX(lucky_history.first_name) AS first_name, COALESCE(SUM(lucky_history.amount + lucky_history.lose_money), 0) AS flow_amount").
		Group("lucky_history.user_id").
		Order("flow_amount desc, lucky_history.user_id desc").
		Limit(20).
		Scan(&rows).Error

	for _, row := range rows {
		result.List = append(result.List, pojo.LuckyHistoryUserFlowBack{
			UserID:     row.UserID,
			Avatar:     row.Avatar,
			FirstName:  maskLuckyHistoryUserName(row.FirstName),
			FlowAmount: row.FlowAmount,
		})
	}

	result.PageSize = 20
	result.CurrentPage = 0
	return result
}

func maskLuckyHistoryUserName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return ""
	}
	runes := []rune(name)
	if len(runes) == 1 {
		return "*"
	}
	if len(runes) == 2 {
		return string(runes[0]) + "*"
	}
	if len(runes) <= 4 {
		return string(runes[0]) + "**" + string(runes[len(runes)-1])
	}
	return string(runes[0]) + "***" + string(runes[len(runes)-2:])
}

func GetLuckyHistoryByLuckyID(db *gorm.DB, tenantID int64, luckyID int64) (result []pojo.LuckyHistory, err error) {
	err = db.Where("tenant_id = ? and lucky_id = ?", tenantID, luckyID).Order("id asc").Find(&result).Error
	return result, err
}
