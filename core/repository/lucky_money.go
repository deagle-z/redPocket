package repository

import (
	"BaseGoUni/core/pojo"
	"encoding/json"
	"errors"
	"github.com/jinzhu/copier"
	"gorm.io/gorm"
)

// CreateLuckyMoney 创建红包
func CreateLuckyMoney(db *gorm.DB, luckyMoney *pojo.LuckyMoney) error {
	return db.Create(luckyMoney).Error
}

// GetLuckyMoney 获取红包详情
func GetLuckyMoney(db *gorm.DB, luckyID int64) (pojo.LuckyMoney, error) {
	var luckyMoney pojo.LuckyMoney
	err := db.Where("id = ?", luckyID).First(&luckyMoney).Error
	return luckyMoney, err
}

// UpdateLuckyMoney 更新红包状态
func UpdateLuckyMoney(db *gorm.DB, luckyID int64, updates map[string]interface{}) error {
	return db.Model(&pojo.LuckyMoney{}).Where("id = ?", luckyID).Updates(updates).Error
}

// GetLuckyMoneyList 红包列表查询（分页）
func GetLuckyMoneyList(db *gorm.DB, search pojo.LuckyMoneySearch) (result pojo.LuckyMoneyResp) {
	var luckyMoneyList []pojo.LuckyMoney
	query := db.Model(&pojo.LuckyMoney{})

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

// GetLuckyMoneyRedList 获取红包金额列表
func GetLuckyMoneyRedList(db *gorm.DB, luckyID int64) ([]float64, error) {
	var luckyMoney pojo.LuckyMoney
	err := db.Where("id = ?", luckyID).Select("red_list").First(&luckyMoney).Error
	if err != nil {
		return nil, err
	}

	var redList []float64
	err = json.Unmarshal([]byte(luckyMoney.RedList), &redList)
	if err != nil {
		return nil, err
	}

	return redList, nil
}

// CheckLuckyMoneyStatus 检查红包状态
func CheckLuckyMoneyStatus(db *gorm.DB, luckyID int64) (bool, error) {
	var luckyMoney pojo.LuckyMoney
	err := db.Where("id = ?", luckyID).First(&luckyMoney).Error
	if err != nil {
		return false, err
	}
	if luckyMoney.Status != 1 {
		return false, errors.New("红包已结束")
	}
	return true, nil
}
