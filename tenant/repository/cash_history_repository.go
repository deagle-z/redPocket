package repository

import (
	"BaseGoUni/core/pojo"

	"github.com/jinzhu/copier"
	"gorm.io/gorm"
	"strings"
)

func GetCashHistoryList(db *gorm.DB, tenantID int64, search pojo.CashHistorySearch) (result pojo.CashHistoryPage) {
	var cashHistoryList []pojo.CashHistory

	tenantUserIDs := db.Model(&pojo.TgUser{}).
		Select("id").
		Where("tenant_id = ?", tenantID)
	if uid := strings.TrimSpace(search.Uid); uid != "" {
		tenantUserIDs = tenantUserIDs.Where("uid = ?", uid)
	}

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
	fillCashHistoryUIDs(db, result.List)

	result.PageSize = search.PageSize
	result.CurrentPage = search.CurrentPage
	return result
}

func fillCashHistoryUIDs(db *gorm.DB, list []pojo.CashHistoryResp) {
	userIDs := make([]int64, 0, len(list))
	seen := make(map[int64]bool, len(list))
	for _, item := range list {
		if item.UserId <= 0 || seen[item.UserId] {
			continue
		}
		seen[item.UserId] = true
		userIDs = append(userIDs, item.UserId)
	}
	if len(userIDs) == 0 {
		return
	}

	var users []pojo.TgUser
	_ = db.Model(&pojo.TgUser{}).Select("id, uid").Where("id IN ?", userIDs).Find(&users).Error
	uidMap := make(map[int64]string, len(users))
	for _, user := range users {
		uidMap[user.ID] = user.Uid
	}
	for i := range list {
		list[i].Uid = uidMap[list[i].UserId]
	}
}
