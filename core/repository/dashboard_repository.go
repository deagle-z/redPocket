package repository

import (
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/utils"
	"time"

	"gorm.io/gorm"
)

func GetAdminDashboardStats(db *gorm.DB, tenantID int64) pojo.TenantDashboardStatsBack {
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	yesterdayStart := todayStart.AddDate(0, 0, -1)
	tomorrowStart := todayStart.AddDate(0, 0, 1)
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	nextMonthStart := monthStart.AddDate(0, 1, 0)
	onlineKey := utils.KeyOnlineTgUsersAll
	if tenantID > 0 {
		onlineKey = utils.OnlineTgUsersKey(tenantID)
	}

	return pojo.TenantDashboardStatsBack{
		Today:                   getAdminDashboardPeriodStats(db, tenantID, todayStart, tomorrowStart),
		Yesterday:               pojo.TenantDashboardPeriodStats{RegisterUsers: countAdminDashboardRegisterUsers(db, tenantID, &yesterdayStart, &todayStart)},
		Month:                   getAdminDashboardPeriodStats(db, tenantID, monthStart, nextMonthStart),
		TotalPlatformPumpAmount: getAdminDashboardPlatformPumpAmount(db, tenantID, nil, nil),
		TotalRegisterUsers:      countAdminDashboardRegisterUsers(db, tenantID, nil, nil),
		OnlineUsers:             utils.CountOnlineUsers(onlineKey),
	}
}

func GetAdminDashboardOnlineUsers(db *gorm.DB, search pojo.TenantDashboardDetailSearch) pojo.TenantDashboardUserDetailResp {
	var result pojo.TenantDashboardUserDetailResp
	offset := int64(search.PageSize * search.CurrentPage)
	onlineKey := utils.KeyOnlineTgUsersAll
	if search.TenantId > 0 {
		onlineKey = utils.OnlineTgUsersKey(search.TenantId)
	}
	items, total := utils.ListOnlineUsers(onlineKey, offset, int64(search.PageSize))
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
	userQuery := db.Model(&pojo.TgUser{}).Where("id IN ?", userIDs)
	userQuery = filterAdminDashboardTenant(userQuery, "tenant_id", search.TenantId)
	_ = userQuery.Find(&users).Error

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
	baseQuery = filterAdminDashboardTenant(baseQuery, "tenant_id", search.TenantId)
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
	rowQuery := db.Table(pojo.RechargeOrderTableName+" ro").
		Select(`ro.user_id,
			ro.tenant_id,
			COALESCE(SUM(ro.amount), 0) AS recharge_amount,
			COUNT(*) AS recharge_count,
			MAX(ro.pay_time) AS last_recharge_at,
			tu.id, tu.uid, tu.tg_id, tu.username, tu.first_name, tu.phone, tu.balance, tu.status`).
		Joins("LEFT JOIN "+pojo.TgUserTableName+" tu ON tu.id = ro.user_id").
		Where("ro.status = ? AND ro.pay_time >= ? AND ro.pay_time < ?", 1, start, end)
	rowQuery = filterAdminDashboardTenant(rowQuery, "ro.tenant_id", search.TenantId)
	_ = rowQuery.
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
	query = filterAdminDashboardTenant(query, "tenant_id", search.TenantId)
	if hasRange {
		query = query.Where("created_at >= ? AND created_at < ?", start, end)
	}
	_ = query.Count(&result.Total).Error

	var users []pojo.TgUser
	listQuery := db.Model(&pojo.TgUser{}).Where("is_bot = ?", false)
	listQuery = filterAdminDashboardTenant(listQuery, "tenant_id", search.TenantId)
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

func getAdminDashboardPeriodStats(db *gorm.DB, tenantID int64, start time.Time, end time.Time) pojo.TenantDashboardPeriodStats {
	var result pojo.TenantDashboardPeriodStats

	rechargeAmountQuery := db.Model(&pojo.RechargeOrder{}).
		Where("status = ? AND pay_time >= ? AND pay_time < ?", 1, start, end)
	rechargeAmountQuery = filterAdminDashboardTenant(rechargeAmountQuery, "tenant_id", tenantID)
	result.RechargeAmount = sumAdminDashboardAmount(rechargeAmountQuery,
		"amount")

	rechargeUsersQuery := db.Model(&pojo.RechargeOrder{}).
		Where("status = ? AND pay_time >= ? AND pay_time < ?", 1, start, end)
	rechargeUsersQuery = filterAdminDashboardTenant(rechargeUsersQuery, "tenant_id", tenantID)
	_ = rechargeUsersQuery.Distinct("user_id").Count(&result.RechargeUsers).Error

	betQuery := db.Table(pojo.LuckyHistoryTableName+" lh").
		Joins("JOIN "+pojo.TgUserTableName+" tu ON tu.id = lh.user_id AND tu.is_bot = ?", false).
		Where("lh.created_at >= ? AND lh.created_at < ?", start, end)
	betQuery = filterAdminDashboardTenant(betQuery, "lh.tenant_id", tenantID)
	result.BetAmount = sumAdminDashboardAmount(betQuery,
		"lh.amount + lh.lose_money")

	withdrawQuery := db.Model(&pojo.WithdrawOrderBr{}).
		Where("status = ? AND paid_at >= ? AND paid_at < ?", 3, start, end)
	withdrawQuery = filterAdminDashboardTenant(withdrawQuery, "tenant_id", tenantID)
	result.WithdrawAmount = sumAdminDashboardAmount(withdrawQuery,
		"amount")

	rebateQuery := db.Model(&pojo.TgUserRebateRecord{}).
		Where("status = ? AND created_at >= ? AND created_at < ?", 1, start, end)
	rebateQuery = filterAdminDashboardTenant(rebateQuery, "tenant_id", tenantID)
	result.RebateAmount = sumAdminDashboardAmount(rebateQuery,
		"rebate_amount")

	result.PlatformPumpAmount = getAdminDashboardPlatformPumpAmount(db, tenantID, &start, &end)
	result.RegisterUsers = countAdminDashboardRegisterUsers(db, tenantID, &start, &end)

	return result
}

func countAdminDashboardRegisterUsers(db *gorm.DB, tenantID int64, start *time.Time, end *time.Time) int64 {
	var total int64
	query := db.Model(&pojo.TgUser{}).Where("is_bot = ?", false)
	query = filterAdminDashboardTenant(query, "tenant_id", tenantID)
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

func getAdminDashboardPlatformPumpAmount(db *gorm.DB, tenantID int64, start *time.Time, end *time.Time) float64 {
	query := db.Model(&pojo.PlatformProfitLedger{}).
		Where("source_type IN ?", []string{
			pojo.PlatformProfitSourceLuckyGrabCommission,
			pojo.PlatformProfitSourceLuckyThunderCommission,
		})
	query = filterAdminDashboardTenant(query, "tenant_id", tenantID)
	if start != nil && end != nil {
		query = query.Where("created_at >= ? AND created_at < ?", *start, *end)
	}
	return sumAdminDashboardAmount(query, "COALESCE(actual_income_amount, income_amount - COALESCE(rebate_amount, 0))")
}

func filterAdminDashboardTenant(query *gorm.DB, column string, tenantID int64) *gorm.DB {
	if tenantID <= 0 {
		return query
	}
	return query.Where(column+" = ?", tenantID)
}
