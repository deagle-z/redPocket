package repository

import (
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/utils"
	"time"

	"gorm.io/gorm"
)

func GetDashboardStats(db *gorm.DB, tenantID int64) pojo.TenantDashboardStatsBack {
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	tomorrowStart := todayStart.AddDate(0, 0, 1)
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	nextMonthStart := monthStart.AddDate(0, 1, 0)

	return pojo.TenantDashboardStatsBack{
		Today:                   getDashboardPeriodStats(db, tenantID, todayStart, tomorrowStart),
		Month:                   getDashboardPeriodStats(db, tenantID, monthStart, nextMonthStart),
		TotalPlatformPumpAmount: getDashboardPlatformPumpAmount(db, tenantID, nil, nil),
		OnlineUsers:             utils.CountOnlineUsers(utils.OnlineTgUsersKey(tenantID)),
	}
}

func GetDashboardOnlineUsers(db *gorm.DB, tenantID int64, search pojo.TenantDashboardDetailSearch) pojo.TenantDashboardUserDetailResp {
	var result pojo.TenantDashboardUserDetailResp
	offset := int64(search.PageSize * search.CurrentPage)
	items, total := utils.ListOnlineUsers(utils.OnlineTgUsersKey(tenantID), offset, int64(search.PageSize))
	result.Total = total
	result.PageSize = search.PageSize
	result.CurrentPage = search.CurrentPage
	if len(items) == 0 {
		return result
	}

	userIDs := make([]int64, 0, len(items))
	activeMap := make(map[int64]time.Time, len(items))
	for _, item := range items {
		userIDs = append(userIDs, item.UserID)
		activeMap[item.UserID] = time.Unix(item.LastActive, 0)
	}

	var users []pojo.TgUser
	_ = db.Model(&pojo.TgUser{}).
		Where("tenant_id = ? AND id IN ?", tenantID, userIDs).
		Find(&users).Error

	userMap := make(map[int64]pojo.TgUser, len(users))
	for _, user := range users {
		userMap[user.ID] = user
	}

	for _, userID := range userIDs {
		user, ok := userMap[userID]
		if !ok {
			continue
		}
		activeAt := activeMap[userID]
		result.List = append(result.List, pojo.TenantDashboardUserDetailBack{
			ID:           user.ID,
			TenantId:     user.TenantId,
			Uid:          user.Uid,
			TgID:         user.TgID,
			Username:     user.Username,
			FirstName:    user.FirstName,
			Phone:        user.Phone,
			Balance:      user.Balance,
			Status:       user.Status,
			LastActiveAt: &activeAt,
		})
	}
	return result
}

func GetDashboardRechargeUsers(db *gorm.DB, tenantID int64, search pojo.TenantDashboardDetailSearch) pojo.TenantDashboardUserDetailResp {
	start, end := dashboardPeriodRange(search.Period)
	var result pojo.TenantDashboardUserDetailResp

	baseQuery := db.Model(&pojo.RechargeOrder{}).
		Where("tenant_id = ? AND status = ? AND pay_time >= ? AND pay_time < ?", tenantID, 1, start, end)
	_ = baseQuery.Distinct("user_id").Count(&result.Total).Error

	type rechargeUserRow struct {
		UserID         int64      `gorm:"column:user_id"`
		RechargeAmount float64    `gorm:"column:recharge_amount"`
		RechargeCount  int64      `gorm:"column:recharge_count"`
		LastRechargeAt *time.Time `gorm:"column:last_recharge_at"`
		ID             int64      `gorm:"column:id"`
		Uid            string     `gorm:"column:uid"`
		TgID           int64      `gorm:"column:tg_id"`
		Username       *string    `gorm:"column:username"`
		FirstName      *string    `gorm:"column:first_name"`
		Phone          *string    `gorm:"column:phone"`
		Balance        float64    `gorm:"column:balance"`
		Status         int8       `gorm:"column:status"`
	}

	var rows []rechargeUserRow
	_ = db.Table(pojo.RechargeOrderTableName+" ro").
		Select(`ro.user_id,
			COALESCE(SUM(ro.amount), 0) AS recharge_amount,
			COUNT(*) AS recharge_count,
			MAX(ro.pay_time) AS last_recharge_at,
			tu.id, tu.uid, tu.tg_id, tu.username, tu.first_name, tu.phone, tu.balance, tu.status`).
		Joins("LEFT JOIN "+pojo.TgUserTableName+" tu ON tu.id = ro.user_id AND tu.tenant_id = ?", tenantID).
		Where("ro.tenant_id = ? AND ro.status = ? AND ro.pay_time >= ? AND ro.pay_time < ?", tenantID, 1, start, end).
		Group("ro.user_id, tu.id, tu.uid, tu.tg_id, tu.username, tu.first_name, tu.phone, tu.balance, tu.status").
		Order("recharge_amount DESC, recharge_count DESC, ro.user_id DESC").
		Limit(search.PageSize).
		Offset(search.PageSize * search.CurrentPage).
		Scan(&rows).Error

	result.PageSize = search.PageSize
	result.CurrentPage = search.CurrentPage
	result.List = make([]pojo.TenantDashboardUserDetailBack, 0, len(rows))
	for _, row := range rows {
		result.List = append(result.List, pojo.TenantDashboardUserDetailBack{
			ID:             row.ID,
			TenantId:       tenantID,
			Uid:            row.Uid,
			TgID:           row.TgID,
			Username:       row.Username,
			FirstName:      row.FirstName,
			Phone:          row.Phone,
			Balance:        row.Balance,
			Status:         row.Status,
			RechargeAmount: row.RechargeAmount,
			RechargeCount:  row.RechargeCount,
			LastRechargeAt: row.LastRechargeAt,
		})
	}
	return result
}

func dashboardPeriodRange(period string) (time.Time, time.Time) {
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	if period == "month" {
		monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		return monthStart, monthStart.AddDate(0, 1, 0)
	}
	return todayStart, todayStart.AddDate(0, 0, 1)
}

func getDashboardPeriodStats(db *gorm.DB, tenantID int64, start time.Time, end time.Time) pojo.TenantDashboardPeriodStats {
	var result pojo.TenantDashboardPeriodStats

	result.RechargeAmount = sumDashboardAmount(db.Model(&pojo.RechargeOrder{}).
		Where("tenant_id = ? AND status = ? AND pay_time >= ? AND pay_time < ?", tenantID, 1, start, end),
		"amount")

	_ = db.Model(&pojo.RechargeOrder{}).
		Where("tenant_id = ? AND status = ? AND pay_time >= ? AND pay_time < ?", tenantID, 1, start, end).
		Distinct("user_id").
		Count(&result.RechargeUsers).Error

	result.BetAmount = sumDashboardAmount(db.Table(pojo.LuckyHistoryTableName+" lh").
		Joins("JOIN "+pojo.TgUserTableName+" tu ON tu.id = lh.user_id AND tu.tenant_id = ? AND tu.is_bot = ?", tenantID, false).
		Where("lh.tenant_id = ? AND lh.created_at >= ? AND lh.created_at < ?", tenantID, start, end),
		"lh.amount + lh.lose_money")

	result.WithdrawAmount = sumDashboardAmount(db.Model(&pojo.WithdrawOrderBr{}).
		Where("tenant_id = ? AND status = ? AND paid_at >= ? AND paid_at < ?", tenantID, 3, start, end),
		"amount")

	result.RebateAmount = sumDashboardAmount(db.Model(&pojo.TgUserRebateRecord{}).
		Where("tenant_id = ? AND status = ? AND created_at >= ? AND created_at < ?", tenantID, 1, start, end),
		"rebate_amount")

	result.PlatformPumpAmount = getDashboardPlatformPumpAmount(db, tenantID, &start, &end)

	return result
}

func sumDashboardAmount(query *gorm.DB, expr string) float64 {
	var row struct {
		Value float64 `gorm:"column:value"`
	}
	_ = query.Select("COALESCE(SUM(" + expr + "), 0) AS value").Scan(&row).Error
	return utils.Truncate2(row.Value)
}

func getDashboardPlatformPumpAmount(db *gorm.DB, tenantID int64, start *time.Time, end *time.Time) float64 {
	query := db.Model(&pojo.PlatformProfitLedger{}).
		Where("tenant_id = ? AND source_type IN ?", tenantID, []string{
			pojo.PlatformProfitSourceLuckyGrabCommission,
			pojo.PlatformProfitSourceLuckyThunderCommission,
		})
	if start != nil && end != nil {
		query = query.Where("created_at >= ? AND created_at < ?", *start, *end)
	}
	return sumDashboardAmount(query, "income_amount")
}
