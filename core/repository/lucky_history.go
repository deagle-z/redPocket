package repository

import (
	"BaseGoUni/core/pojo"
	"github.com/jinzhu/copier"
	"gorm.io/gorm"
)

// CreateLuckyHistory 创建领取记录
func CreateLuckyHistory(db *gorm.DB, history *pojo.LuckyHistory) error {
	return db.Create(history).Error
}

// GetLuckyHistoryByLuckyId 获取红包的所有领取记录
func GetLuckyHistoryByLuckyId(db *gorm.DB, luckyID int64) ([]pojo.LuckyHistory, error) {
	var historyList []pojo.LuckyHistory
	err := db.Where("lucky_id = ?", luckyID).Order("id asc").Find(&historyList).Error
	return historyList, err
}

// CheckUserGrabbed 检查用户是否已领取
func CheckUserGrabbed(db *gorm.DB, luckyID int64, userID int64) (bool, error) {
	var count int64
	err := db.Model(&pojo.LuckyHistory{}).
		Where("lucky_id = ? AND user_id = ?", luckyID, userID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetLuckyHistoryList 领取历史列表（分页）
func GetLuckyHistoryList(db *gorm.DB, search pojo.LuckyHistorySearch) (result pojo.LuckyHistoryResp) {
	var historyList []pojo.LuckyHistory
	query := db.Model(&pojo.LuckyHistory{})

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

// GetLuckyHistoryCount 获取红包已领取数量
func GetLuckyHistoryCount(db *gorm.DB, luckyID int64) (int64, error) {
	var count int64
	err := db.Model(&pojo.LuckyHistory{}).
		Where("lucky_id = ?", luckyID).
		Count(&count).Error
	return count, err
}
