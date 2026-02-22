package repository

import (
	"BaseGoUni/core/pojo"
	"errors"
	"github.com/jinzhu/copier"
	"gorm.io/gorm"
)

func GetLuckyMoneyList(db *gorm.DB, tenantID int64, search pojo.LuckyMoneySearch) (result pojo.LuckyMoneyResp) {
	var luckyMoneyList []pojo.LuckyMoney
	query := db.Model(&pojo.LuckyMoney{}).Where("tenant_id = ?", tenantID)
	if search.SenderID > 0 {
		query = query.Where("sender_id = ?", search.SenderID)
	}
	if search.ChatID > 0 {
		query = query.Where("chat_id = ?", search.ChatID)
	}
	if search.Status != nil {
		query = query.Where("status = ?", *search.Status)
	}
	query.Count(&result.Total)
	query = query.Order("id desc").Limit(search.PageSize).Offset(search.PageSize * search.CurrentPage)
	query.Find(&luckyMoneyList)
	for _, lucky := range luckyMoneyList {
		var tempLucky pojo.LuckyMoneyBack
		_ = copier.Copy(&tempLucky, &lucky)
		result.List = append(result.List, tempLucky)
	}
	result.PageSize = search.PageSize
	result.CurrentPage = search.CurrentPage
	return result
}

func GetLuckyMoneyByID(db *gorm.DB, tenantID int64, id int64) (result pojo.LuckyMoney, err error) {
	db.Where("id = ? and tenant_id = ?", id, tenantID).First(&result)
	if result.ID == 0 {
		return result, errors.New("数据不存在")
	}
	return result, nil
}
