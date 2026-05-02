package repository

import (
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/utils"
	"time"

	"gorm.io/gorm"
)

func GetAdminDashboardStats(db *gorm.DB) pojo.TenantDashboardStatsBack {
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	yesterdayStart := todayStart.AddDate(0, 0, -1)
	tomorrowStart := todayStart.AddDate(0, 0, 1)
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	nextMonthStart := monthStart.AddDate(0, 1, 0)

	return pojo.TenantDashboardStatsBack{
		Today:                   getAdminDashboardPeriodStats(db, todayStart, tomorrowStart),
		Yesterday:               pojo.TenantDashboardPeriodStats{RegisterUsers: countAdminDashboardRegisterUsers(db, &yesterdayStart, &todayStart)},
		Month:                   getAdminDashboardPeriodStats(db, monthStart, nextMonthStart),
		TotalPlatformPumpAmount: getAdminDashboardPlatformPumpAmount(db, nil, nil),
		TotalRegisterUsers:      countAdminDashboardRegisterUsers(db, nil, nil),
		OnlineUsers:             utils.CountOnlineUsers(utils.KeyOnlineTgUsersAll),
	}
}

func GetAdminDashboardOnlineUsers(db *gorm.DB, search pojo.TenantDashboardDetailSearch) pojo.TenantDashboardUserDetailResp {
	var result pojo.TenantDashboardUserDetailResp
	offset := int64(search.PageSize * search.CurrentPage)
	items, total := utils.ListOnlineUsers(utils.KeyOnlineTgUsersAll, offset, int64(search.PageSize))
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
		Where("id IN ?", userIDs).
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
			Balance:      utils.Truncate2(user.Balance),
			Status:       user.Status,
			LastActiveAt: &activeAt,
		})
	}
	return result
}

func GetAdminDashboardRechargeUsers(db *gorm.DB, search pojo.TenantDashboardDetailSearch) pojo.TenantDashboardUserDetailResp {
	start, end := adminDashboardPeriodRange(search.Period)
	var result pojo.TenantDashboardUserDetailResp

	baseQuery := db.Model(&pojo.RechargeOrder{}).
		Where("status = ? AND pay_time >= ? AND pay_time < ?", 1, start, end)
	_ = baseQuery.Distinct("user_id").Count(&result.Total).Error

	type rechargeUserRow struct {
		UserID         int64      `gorm:"column:user_id"`
		TenantId       int64      `gorm:"column:tenant_id"`
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
			ro.tenant_id,
			COALESCE(SUM(ro.amount), 0) AS recharge_amount,
			COUNT(*) AS recharge_count,
			MAX(ro.pay_time) AS last_recharge_at,
			tu.id, tu.uid, tu.tg_id, tu.username, tu.first_name, tu.phone, tu.balance, tu.status`).
		Joins("LEFT JOIN "+pojo.TgUserTableName+" tu ON tu.id = ro.user_id").
		Where("ro.status = ? AND ro.pay_time >= ? AND ro.pay_time < ?", 1, start, end).
		Group("ro.user_id, ro.tenant_id, tu.id, tu.uid, tu.tg_id, tu.username, tu.first_name, tu.phone, tu.balance, tu.status").
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
			TenantId:       row.TenantId,
			Uid:            row.Uid,
			TgID:           row.TgID,
			Username:       row.Username,
			FirstName:      row.FirstName,
			Phone:          row.Phone,
			Balance:        utils.Truncate2(row.Balance),
			Status:         row.Status,
			RechargeAmount: utils.Truncate2(row.RechargeAmount),
			RechargeCount:  row.RechargeCount,
			LastRechargeAt: row.LastRechargeAt,
		})
	}
	return result
}

func GetAdminDashboardRegisterUsers(db *gorm.DB, search pojo.TenantDashboardDetailSearch) pojo.TenantDashboardUserDetailResp {
	start, end, hasRange := adminDashboardDetailPeriodRange(search.Period)
	var result pojo.TenantDashboardUserDetailResp

	query := db.Model(&pojo.TgUser{}).Where("is_bot = ?", false)
	if hasRange {
		query = query.Where("created_at >= ? AND created_at < ?", start, end)
	}
	_ = query.Count(&result.Total).Error

	var users []pojo.TgUser
	listQuery := db.Model(&pojo.TgUser{}).Where("is_bot = ?", false)
	if hasRange {
		listQuery = listQuery.Where("created_at >= ? AND created_at < ?", start, end)
	}
	_ = listQuery.Order("created_at desc, id desc").
		Limit(search.PageSize).
		Offset(search.PageSize * search.CurrentPage).
		Find(&users).Error

	result.PageSize = search.PageSize
	result.CurrentPage = search.CurrentPage
	result.List = make([]pojo.TenantDashboardUserDetailBack, 0, len(users))
	for _, user := range users {
		registeredAt := user.CreatedAt
		result.List = append(result.List, pojo.TenantDashboardUserDetailBack{
			ID:           user.ID,
			TenantId:     user.TenantId,
			Uid:          user.Uid,
			TgID:         user.TgID,
			Username:     user.Username,
			FirstName:    user.FirstName,
			Phone:        user.Phone,
			Balance:      utils.Truncate2(user.Balance),
			Status:       user.Status,
			RegisteredAt: &registeredAt,
		})
	}
	return result
}

func adminDashboardPeriodRange(period string) (time.Time, time.Time) {
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	if period == "yesterday" {
		yesterdayStart := todayStart.AddDate(0, 0, -1)
		return yesterdayStart, todayStart
	}
	if period == "month" {
		monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		return monthStart, monthStart.AddDate(0, 1, 0)
	}
	return todayStart, todayStart.AddDate(0, 0, 1)
}

func adminDashboardDetailPeriodRange(period string) (time.Time, time.Time, bool) {
	if period == "total" {
		return time.Time{}, time.Time{}, false
	}
	start, end := adminDashboardPeriodRange(period)
	return start, end, true
}

func getAdminDashboardPeriodStats(db *gorm.DB, start time.Time, end time.Time) pojo.TenantDashboardPeriodStats {
	var result pojo.TenantDashboardPeriodStats

	result.RechargeAmount = sumAdminDashboardAmount(db.Model(&pojo.RechargeOrder{}).
		Where("status = ? AND pay_time >= ? AND pay_time < ?", 1, start, end),
		"amount")

	_ = db.Model(&pojo.RechargeOrder{}).
		Where("status = ? AND pay_time >= ? AND pay_time < ?", 1, start, end).
		Distinct("user_id").
		Count(&result.RechargeUsers).Error

	result.BetAmount = sumAdminDashboardAmount(db.Table(pojo.LuckyHistoryTableName+" lh").
		Joins("JOIN "+pojo.TgUserTableName+" tu ON tu.id = lh.user_id AND tu.is_bot = ?", false).
		Where("lh.created_at >= ? AND lh.created_at < ?", start, end),
		"lh.amount + lh.lose_money")

	result.WithdrawAmount = sumAdminDashboardAmount(db.Model(&pojo.WithdrawOrderBr{}).
		Where("status = ? AND paid_at >= ? AND paid_at < ?", 3, start, end),
		"amount")

	result.RebateAmount = sumAdminDashboardAmount(db.Model(&pojo.TgUserRebateRecord{}).
		Where("status = ? AND created_at >= ? AND created_at < ?", 1, start, end),
		"rebate_amount")

	result.PlatformPumpAmount = getAdminDashboardPlatformPumpAmount(db, &start, &end)
	result.RegisterUsers = countAdminDashboardRegisterUsers(db, &start, &end)

	return result
}

func countAdminDashboardRegisterUsers(db *gorm.DB, start *time.Time, end *time.Time) int64 {
	var total int64
	query := db.Model(&pojo.TgUser{}).Where("is_bot = ?", false)
	if start != nil && end != nil {
		query = query.Where("created_at >= ? AND created_at < ?", *start, *end)
	}
	_ = query.Count(&total).Error
	return total
}

func sumAdminDashboardAmount(query *gorm.DB, expr string) float64 {
	var row struct {
		Value float64 `gorm:"column:value"`
	}
	_ = query.Select("COALESCE(SUM(" + expr + "), 0) AS value").Scan(&row).Error
	return utils.Truncate2(row.Value)
}

func getAdminDashboardPlatformPumpAmount(db *gorm.DB, start *time.Time, end *time.Time) float64 {
	query := db.Model(&pojo.PlatformProfitLedger{}).
		Where("source_type IN ?", []string{
			pojo.PlatformProfitSourceLuckyGrabCommission,
			pojo.PlatformProfitSourceLuckyThunderCommission,
		})
	if start != nil && end != nil {
		query = query.Where("created_at >= ? AND created_at < ?", *start, *end)
	}
	return sumAdminDashboardAmount(query, "COALESCE(actual_income_amount, income_amount - COALESCE(rebate_amount, 0))")
}
