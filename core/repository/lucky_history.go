package repository

import (
	"BaseGoUni/core/pojo"
	"fmt"
	"github.com/jinzhu/copier"
	"gorm.io/gorm"
	"strconv"
	"time"
)

type hiddenLuckyGrabRecord struct {
	HistoryID  int64
	LuckyID    int64
	UserID     int64
	Amount     float64
	LoseMoney  float64
	IsThunder  int
	CreatedAt  time.Time
}

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

	luckyIDs := make([]int64, 0, len(rows))
	for _, row := range rows {
		if row.LuckyID > 0 {
			luckyIDs = append(luckyIDs, row.LuckyID)
		}
	}
	hiddenMap := getPendingHiddenGrabRecordMap(db, 0, luckyIDs)

	now := time.Now()
	for _, row := range rows {
		amount := row.Amount
		if hidden, ok := hiddenMap[row.ID]; ok && hidden.HistoryID == row.ID {
			amount = 0
		}
		result = append(result, pojo.LuckyRecentWinnerBack{
			ID:        row.ID,
			UserID:    row.UserID,
			FirstName: row.FirstName,
			Avatar:    row.Avatar,
			Amount:    amount,
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

// GetLuckyAppHistoryUnion app端发包+抢包历史（union）
func GetLuckyAppHistoryUnion(db *gorm.DB, userID int64, search pojo.LuckyAppHistorySearch) (result pojo.LuckyAppHistoryResp) {
	unionSQL := `
		SELECT
			'send' AS record_type,
			1 AS action_type,
			m.id AS record_id,
			m.id AS lucky_id,
			m.amount AS lucky_amount,
			0 AS grab_amount,
			0 AS lose_money,
			0 AS is_thunder,
			m.thunder AS thunder,
			m.sender_id AS sender_id,
			m.sender_name AS sender_name,
			u.avatar AS avatar,
			0 AS income,
			m.amount AS expense,
			-m.amount AS net_amount,
			m.created_at AS created_at
		FROM lucky_money m
		LEFT JOIN tg_user u ON u.id = m.sender_id
		WHERE m.sender_id = ?

		UNION ALL

		SELECT
			'grab' AS record_type,
			2 AS action_type,
			h.id AS record_id,
			h.lucky_id AS lucky_id,
			m.amount AS lucky_amount,
			h.amount AS grab_amount,
			h.lose_money AS lose_money,
			h.is_thunder AS is_thunder,
			m.thunder AS thunder,
			m.sender_id AS sender_id,
			m.sender_name AS sender_name,
			u.avatar AS avatar,
			CASE WHEN h.is_thunder = 0 THEN h.amount ELSE 0 END AS income,
			CASE WHEN h.is_thunder = 1 THEN h.lose_money ELSE 0 END AS expense,
			CASE WHEN h.is_thunder = 0 THEN h.amount ELSE -h.lose_money END AS net_amount,
			h.created_at AS created_at
		FROM lucky_history h
		LEFT JOIN lucky_money m ON m.id = h.lucky_id
		LEFT JOIN tg_user u ON u.id = m.sender_id
		WHERE h.user_id = ?
	`

	args := []interface{}{userID, userID}
	whereSQL := " WHERE 1=1"

	if search.ActionType == 1 || search.ActionType == 2 {
		whereSQL += " AND t.action_type = ?"
		args = append(args, search.ActionType)
	}

	if search.ResultType == 1 {
		whereSQL += " AND (t.net_amount > 0 OR (t.record_type = 'grab' AND t.is_thunder = 0 AND t.grab_amount > 0))"
	} else if search.ResultType == 2 {
		whereSQL += " AND (t.net_amount < 0 OR t.is_thunder = 1)"
	}

	if search.StartTime > 0 {
		whereSQL += " AND t.created_at >= ?"
		args = append(args, time.Unix(search.StartTime, 0))
	}

	if search.EndTime > 0 {
		whereSQL += " AND t.created_at <= ?"
		args = append(args, time.Unix(search.EndTime, 0))
	}

	countSQL := fmt.Sprintf("SELECT COUNT(1) AS total FROM (%s) t %s", unionSQL, whereSQL)
	_ = db.Raw(countSQL, args...).Scan(&result.Total).Error

	type summaryRow struct {
		TotalIncome   float64 `gorm:"column:total_income"`
		TotalExpense  float64 `gorm:"column:total_expense"`
		NetProfitLoss float64 `gorm:"column:net_profit_loss"`
	}
	var summary summaryRow
	summarySQL := fmt.Sprintf(`
		SELECT
			COALESCE(SUM(t.income), 0) AS total_income,
			COALESCE(SUM(t.expense), 0) AS total_expense,
			COALESCE(SUM(t.net_amount), 0) AS net_profit_loss
		FROM (%s) t %s
	`, unionSQL, whereSQL)
	_ = db.Raw(summarySQL, args...).Scan(&summary).Error
	result.TotalIncome = summary.TotalIncome
	result.TotalExpense = summary.TotalExpense
	result.NetProfitLoss = summary.NetProfitLoss

	hiddenMapAll := getPendingHiddenGrabRecordMap(db, userID, nil)
	for _, hidden := range hiddenMapAll {
		if !hiddenRecordMatchesSearch(hidden, search) {
			continue
		}
		income := 0.0
		expense := 0.0
		net := 0.0
		if hidden.IsThunder == 0 {
			income = hidden.Amount
			net = hidden.Amount
		} else {
			expense = hidden.LoseMoney
			net = -hidden.LoseMoney
		}
		result.TotalIncome -= income
		result.TotalExpense -= expense
		result.NetProfitLoss -= net
	}

	listSQL := fmt.Sprintf(`
		SELECT
			t.record_type,
			t.action_type,
			t.record_id,
			t.lucky_id,
			t.lucky_amount,
			t.grab_amount,
			t.lose_money,
			t.is_thunder,
			t.thunder,
			t.sender_id,
			t.sender_name,
			t.avatar,
			t.income,
			t.expense,
			t.net_amount,
			t.created_at
		FROM (%s) t %s
		ORDER BY t.created_at DESC, t.record_id DESC
		LIMIT ? OFFSET ?
	`, unionSQL, whereSQL)
	listArgs := append(args, search.PageSize, search.PageSize*search.CurrentPage)
	_ = db.Raw(listSQL, listArgs...).Scan(&result.List).Error

	for i := range result.List {
		row := &result.List[i]
		if row.RecordType != "grab" {
			continue
		}
		hidden, ok := hiddenMapAll[row.RecordID]
		if !ok {
			continue
		}
		row.GrabAmount = 0
		row.LoseMoney = 0
		row.Income = 0
		row.Expense = 0
		row.NetProfit = 0
		row.IsThunder = hidden.IsThunder
	}

	result.PageSize = search.PageSize
	result.CurrentPage = search.CurrentPage
	return result
}

func hiddenRecordMatchesSearch(record hiddenLuckyGrabRecord, search pojo.LuckyAppHistorySearch) bool {
	if search.ActionType == 1 {
		return false
	}
	if search.ActionType != 0 && search.ActionType != 2 {
		return false
	}

	if search.StartTime > 0 && record.CreatedAt.Before(time.Unix(search.StartTime, 0)) {
		return false
	}
	if search.EndTime > 0 && record.CreatedAt.After(time.Unix(search.EndTime, 0)) {
		return false
	}

	if search.ResultType == 1 {
		return record.IsThunder == 0 && record.Amount > 0
	}
	if search.ResultType == 2 {
		return record.IsThunder == 1
	}
	return true
}

func getPendingHiddenGrabRecordMap(db *gorm.DB, targetUserID int64, luckyIDs []int64) map[int64]hiddenLuckyGrabRecord {
	result := make(map[int64]hiddenLuckyGrabRecord)
	if db == nil {
		return result
	}

	var luckyList []pojo.LuckyMoney
	query := db.Model(&pojo.LuckyMoney{}).Where("status = ?", 1)
	if len(luckyIDs) > 0 {
		query = query.Where("id IN ?", luckyIDs)
	}
	if err := query.Find(&luckyList).Error; err != nil || len(luckyList) == 0 {
		return result
	}

	activeLuckyIDs := make([]int64, 0, len(luckyList))
	for _, lucky := range luckyList {
		activeLuckyIDs = append(activeLuckyIDs, lucky.ID)
	}

	type luckyCountRow struct {
		LuckyID      int64 `gorm:"column:lucky_id"`
		GrabbedCount int64 `gorm:"column:grabbed_count"`
	}
	grabbedCountMap := map[int64]int64{}
	var countRows []luckyCountRow
	_ = db.Model(&pojo.LuckyHistory{}).
		Select("lucky_id, COUNT(1) AS grabbed_count").
		Where("lucky_id IN ?", activeLuckyIDs).
		Group("lucky_id").
		Scan(&countRows).Error
	for _, row := range countRows {
		grabbedCountMap[row.LuckyID] = row.GrabbedCount
	}

	itemMap := map[int64][]pojo.LuckyMoneyItem{}
	var allItems []pojo.LuckyMoneyItem
	_ = db.Where("red_packet_id IN ?", activeLuckyIDs).Order("red_packet_id asc, seq_no asc").Find(&allItems).Error
	for _, item := range allItems {
		id := int64(item.RedPacketID)
		itemMap[id] = append(itemMap[id], item)
	}

	now := time.Now()
	for _, lucky := range luckyList {
		grabbedCount := int(grabbedCountMap[lucky.ID])
		expireAt := lucky.ExpireTime
		if expireAt.IsZero() {
			expireAt = lucky.CreatedAt.Add(3 * time.Minute)
		}
		if !shouldHidePendingSecondLastAmount(lucky.Status, expireAt, lucky.Number, grabbedCount, now) {
			continue
		}

		latestItem, found := findLatestGrabbedItem(itemMap[lucky.ID])
		if !found || latestItem.GrabbedUid == nil {
			continue
		}
		grabbedUserID := int64(*latestItem.GrabbedUid)
		if targetUserID > 0 && grabbedUserID != targetUserID {
			continue
		}

		var history pojo.LuckyHistory
		historyQuery := db.Model(&pojo.LuckyHistory{}).
			Where("lucky_id = ? AND user_id = ?", lucky.ID, grabbedUserID).
			Order("id desc")
		if latestItem.GrabbedAt != nil {
			windowStart := latestItem.GrabbedAt.Add(-2 * time.Second)
			windowEnd := latestItem.GrabbedAt.Add(2 * time.Second)
			historyQuery = historyQuery.Where("created_at >= ? AND created_at <= ?", windowStart, windowEnd)
		}
		if err := historyQuery.First(&history).Error; err != nil || history.ID == 0 {
			continue
		}

		result[history.ID] = hiddenLuckyGrabRecord{
			HistoryID: history.ID,
			LuckyID:   lucky.ID,
			UserID:    history.UserID,
			Amount:    history.Amount,
			LoseMoney: history.LoseMoney,
			IsThunder: history.IsThunder,
			CreatedAt: history.CreatedAt,
		}
	}

	return result
}
