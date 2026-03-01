package repository

import (
	"BaseGoUni/core/pojo"
	"github.com/jinzhu/copier"
	"gorm.io/gorm"
	"strconv"
	"time"
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

// GetRecentLuckyWinners 获取最近中奖列表（app端）
func GetRecentLuckyWinners(db *gorm.DB, search pojo.LuckyRecentWinnerSearch) (result []pojo.LuckyRecentWinnerBack) {
	limit := search.Limit
	if limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}

	type luckyRecentWinnerRow struct {
		ID        int64     `gorm:"column:id"`
		UserID    int64     `gorm:"column:user_id"`
		FirstName string    `gorm:"column:first_name"`
		Avatar    *string   `gorm:"column:avatar"`
		Amount    float64   `gorm:"column:amount"`
		LuckyID   int64     `gorm:"column:lucky_id"`
		CreatedAt time.Time `gorm:"column:created_at"`
	}

	var rows []luckyRecentWinnerRow
	_ = db.Table("lucky_history h").
		Select("h.id, h.user_id, h.first_name, h.amount, h.lucky_id, h.created_at, u.avatar").
		Joins("left join tg_user u on u.id = h.user_id").
		Where("h.is_thunder = ? and h.amount > 0", 0).
		Order("h.created_at desc").
		Limit(limit).
		Scan(&rows).Error

	now := time.Now()
	for _, row := range rows {
		result = append(result, pojo.LuckyRecentWinnerBack{
			ID:        row.ID,
			UserID:    row.UserID,
			FirstName: row.FirstName,
			Avatar:    row.Avatar,
			Amount:    row.Amount,
			LuckyID:   row.LuckyID,
			CreatedAt: row.CreatedAt,
			TimeText:  formatRecentTimeText(now.Sub(row.CreatedAt)),
		})
	}
	return result
}

func formatRecentTimeText(d time.Duration) string {
	if d < time.Minute {
		return "刚刚"
	}
	if d < time.Hour {
		return strconv.Itoa(int(d.Minutes())) + "分钟前"
	}
	if d < 24*time.Hour {
		return strconv.Itoa(int(d.Hours())) + "小时前"
	}
	return strconv.Itoa(int(d.Hours()/24)) + "天前"
}
