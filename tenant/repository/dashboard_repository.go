package repository

import (
	"BaseGoUni/core/pojo"
	coreRepo "BaseGoUni/core/repository"
	"BaseGoUni/core/utils"
	"time"

	"gorm.io/gorm"
)

func GetDashboardStats(db *gorm.DB, tenantID int64) pojo.TenantDashboardStatsBack {
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	yesterdayStart := todayStart.AddDate(0, 0, -1)
	tomorrowStart := todayStart.AddDate(0, 0, 1)
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	nextMonthStart := monthStart.AddDate(0, 1, 0)

	return pojo.TenantDashboardStatsBack{
		Today:                   getDashboardPeriodStats(db, tenantID, todayStart, tomorrowStart),
		Yesterday:               pojo.TenantDashboardPeriodStats{RegisterUsers: countDashboardRegisterUsers(db, tenantID, &yesterdayStart, &todayStart)},
		Month:                   getDashboardPeriodStats(db, tenantID, monthStart, nextMonthStart),
		TotalPlatformPumpAmount: getDashboardPlatformPumpAmount(db, tenantID, nil, nil),
		TotalRegisterUsers:      countDashboardRegisterUsers(db, tenantID, nil, nil),
		OnlineUsers:             utils.CountOnlineUsers(utils.OnlineTgUsersKey(tenantID)),
	}
}

func GetDashboardMonthlyBalances(db *gorm.DB, tenantID int64, year int) pojo.TenantDashboardMonthlyBalanceResp {
	return coreRepo.GetAdminDashboardMonthlyBalances(db, tenantID, year)
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
		Where("tenant_id = ? AND status = ? AND coalesce(is_dev, 0) = 0 AND pay_time >= ? AND pay_time < ?", tenantID, 1, start, end)
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
		ParentID       *int64     `gorm:"column:parent_id"`
		ParentUid      *string    `gorm:"column:parent_uid"`
		Balance        float64    `gorm:"column:balance"`
		Status         int8       `gorm:"column:status"`
	}

	var rows []rechargeUserRow
	_ = db.Table(pojo.RechargeOrderTableName+" ro").
		Select(`ro.user_id,
			COALESCE(SUM(`+coreRepo.DashboardRechargePaidAmountExpr("ro.")+`), 0) AS recharge_amount,
			COUNT(*) AS recharge_count,
			MAX(ro.pay_time) AS last_recharge_at,
			tu.id, tu.uid, tu.tg_id, tu.username, tu.first_name, tu.phone, tu.parent_id, parent.uid AS parent_uid, tu.balance, tu.status`).
		Joins("LEFT JOIN "+pojo.TgUserTableName+" tu ON tu.id = ro.user_id AND tu.tenant_id = ?", tenantID).
		Joins("LEFT JOIN "+pojo.TgUserTableName+" parent ON parent.id = tu.parent_id AND parent.tenant_id = tu.tenant_id").
		Where("ro.tenant_id = ? AND ro.status = ? AND coalesce(ro.is_dev, 0) = 0 AND ro.pay_time >= ? AND ro.pay_time < ?", tenantID, 1, start, end).
		Group("ro.user_id, tu.id, tu.uid, tu.tg_id, tu.username, tu.first_name, tu.phone, tu.parent_id, parent.uid, tu.balance, tu.status").
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
			ParentID:       row.ParentID,
			ParentUid:      row.ParentUid,
			Balance:        row.Balance,
			Status:         row.Status,
			RechargeAmount: row.RechargeAmount,
			RechargeCount:  row.RechargeCount,
			LastRechargeAt: row.LastRechargeAt,
		})
	}
	return result
}

func GetDashboardAgentRanks(db *gorm.DB, tenantID int64, search pojo.TenantDashboardDetailSearch) pojo.TenantDashboardAgentRankResp {
	search.TenantId = tenantID
	return coreRepo.GetAdminDashboardAgentRanks(db, search)
}

func GetDashboardRegisterUsers(db *gorm.DB, tenantID int64, search pojo.TenantDashboardDetailSearch) pojo.TenantDashboardUserDetailResp {
	start, end, hasRange := dashboardDetailPeriodRange(search.Period)
	var result pojo.TenantDashboardUserDetailResp

	query := db.Model(&pojo.TgUser{}).Where("tenant_id = ? AND is_bot = ?", tenantID, false)
	if hasRange {
		query = query.Where("created_at >= ? AND created_at < ?", start, end)
	}
	_ = query.Count(&result.Total).Error

	var users []pojo.TgUser
	listQuery := db.Model(&pojo.TgUser{}).Where("tenant_id = ? AND is_bot = ?", tenantID, false)
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
			Balance:      user.Balance,
			Status:       user.Status,
			RegisteredAt: &registeredAt,
		})
	}
	return result
}

// GetDashboardRechargeOrders 充值总额明细（本租户，成功且非手动回调），按支付时间倒序。
func GetDashboardRechargeOrders(db *gorm.DB, tenantID int64, search pojo.TenantDashboardDetailSearch) pojo.TenantDashboardOrderDetailResp {
	start, end := dashboardPeriodRange(search.Period)
	var result pojo.TenantDashboardOrderDetailResp

	_ = db.Model(&pojo.RechargeOrder{}).
		Where("tenant_id = ? AND status = ? AND coalesce(is_dev, 0) = 0 AND pay_time >= ? AND pay_time < ?", tenantID, 1, start, end).
		Count(&result.Total).Error
	result.TotalAmount = sumDashboardAmount(db.Model(&pojo.RechargeOrder{}).
		Where("tenant_id = ? AND status = ? AND coalesce(is_dev, 0) = 0 AND pay_time >= ? AND pay_time < ?", tenantID, 1, start, end),
		coreRepo.DashboardRechargePaidAmountExpr(""))

	type orderRow struct {
		ID        int64      `gorm:"column:id"`
		OrderNo   string     `gorm:"column:order_no"`
		UserID    int64      `gorm:"column:user_id"`
		Amount    float64    `gorm:"column:amount"`
		Fee       float64    `gorm:"column:fee"`
		Channel   string     `gorm:"column:channel"`
		Status    int        `gorm:"column:status"`
		PayTime   *time.Time `gorm:"column:pay_time"`
		Uid       string     `gorm:"column:uid"`
		Username  *string    `gorm:"column:username"`
		FirstName *string    `gorm:"column:first_name"`
		Phone     *string    `gorm:"column:phone"`
	}
	var rows []orderRow
	_ = db.Table(pojo.RechargeOrderTableName+" ro").
		Select(`ro.id, ro.order_no, ro.user_id, `+coreRepo.DashboardRechargePaidAmountExpr("ro.")+` AS amount, ro.fee, ro.channel, ro.status, ro.pay_time,
			tu.uid, tu.username, tu.first_name, tu.phone`).
		Joins("LEFT JOIN "+pojo.TgUserTableName+" tu ON tu.id = ro.user_id AND tu.tenant_id = ?", tenantID).
		Where("ro.tenant_id = ? AND ro.status = ? AND coalesce(ro.is_dev, 0) = 0 AND ro.pay_time >= ? AND ro.pay_time < ?", tenantID, 1, start, end).
		Order("ro.pay_time DESC, ro.id DESC").
		Limit(search.PageSize).
		Offset(search.PageSize * search.CurrentPage).
		Scan(&rows).Error

	result.PageSize = search.PageSize
	result.CurrentPage = search.CurrentPage
	result.List = make([]pojo.TenantDashboardOrderDetailBack, 0, len(rows))
	for _, row := range rows {
		result.List = append(result.List, pojo.TenantDashboardOrderDetailBack{
			ID:        row.ID,
			OrderNo:   row.OrderNo,
			TenantId:  tenantID,
			UserID:    row.UserID,
			Uid:       row.Uid,
			Username:  row.Username,
			FirstName: row.FirstName,
			Phone:     row.Phone,
			Amount:    utils.Truncate2(row.Amount),
			Fee:       utils.Truncate2(row.Fee),
			Channel:   row.Channel,
			Status:    row.Status,
			Time:      row.PayTime,
		})
	}
	return result
}

// GetDashboardWithdrawOrders 提现总额明细（本租户，已打款 status=3），按打款时间倒序。
func GetDashboardWithdrawOrders(db *gorm.DB, tenantID int64, search pojo.TenantDashboardDetailSearch) pojo.TenantDashboardOrderDetailResp {
	start, end := dashboardPeriodRange(search.Period)
	var result pojo.TenantDashboardOrderDetailResp

	_ = db.Model(&pojo.WithdrawOrderBr{}).
		Where("tenant_id = ? AND status = ? AND paid_at >= ? AND paid_at < ?", tenantID, 3, start, end).
		Count(&result.Total).Error
	result.TotalAmount = sumDashboardAmount(db.Model(&pojo.WithdrawOrderBr{}).
		Where("tenant_id = ? AND status = ? AND paid_at >= ? AND paid_at < ?", tenantID, 3, start, end),
		"amount")

	type orderRow struct {
		ID        int64      `gorm:"column:id"`
		OrderNo   string     `gorm:"column:order_no"`
		UserID    int64      `gorm:"column:user_id"`
		Amount    float64    `gorm:"column:amount"`
		Fee       float64    `gorm:"column:fee"`
		NetAmount float64    `gorm:"column:net_amount"`
		Channel   string     `gorm:"column:channel"`
		Status    int        `gorm:"column:status"`
		PaidAt    *time.Time `gorm:"column:paid_at"`
		Uid       string     `gorm:"column:uid"`
		Username  *string    `gorm:"column:username"`
		FirstName *string    `gorm:"column:first_name"`
		Phone     *string    `gorm:"column:phone"`
	}
	var rows []orderRow
	_ = db.Table(pojo.WithdrawOrderBrTableName+" wo").
		Select(`wo.id, wo.order_no, wo.user_id, wo.amount, wo.fee, wo.net_amount, wo.channel, wo.status, wo.paid_at,
			tu.uid, tu.username, tu.first_name, tu.phone`).
		Joins("LEFT JOIN "+pojo.TgUserTableName+" tu ON tu.id = wo.user_id AND tu.tenant_id = ?", tenantID).
		Where("wo.tenant_id = ? AND wo.status = ? AND wo.paid_at >= ? AND wo.paid_at < ?", tenantID, 3, start, end).
		Order("wo.paid_at DESC, wo.id DESC").
		Limit(search.PageSize).
		Offset(search.PageSize * search.CurrentPage).
		Scan(&rows).Error

	result.PageSize = search.PageSize
	result.CurrentPage = search.CurrentPage
	result.List = make([]pojo.TenantDashboardOrderDetailBack, 0, len(rows))
	for _, row := range rows {
		result.List = append(result.List, pojo.TenantDashboardOrderDetailBack{
			ID:        row.ID,
			OrderNo:   row.OrderNo,
			TenantId:  tenantID,
			UserID:    row.UserID,
			Uid:       row.Uid,
			Username:  row.Username,
			FirstName: row.FirstName,
			Phone:     row.Phone,
			Amount:    utils.Truncate2(row.Amount),
			Fee:       utils.Truncate2(row.Fee),
			NetAmount: utils.Truncate2(row.NetAmount),
			Channel:   row.Channel,
			Status:    row.Status,
			Time:      row.PaidAt,
		})
	}
	return result
}

func dashboardPeriodRange(period string) (time.Time, time.Time) {
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

func dashboardDetailPeriodRange(period string) (time.Time, time.Time, bool) {
	if period == "total" {
		return time.Time{}, time.Time{}, false
	}
	start, end := dashboardPeriodRange(period)
	return start, end, true
}

func getDashboardPeriodStats(db *gorm.DB, tenantID int64, start time.Time, end time.Time) pojo.TenantDashboardPeriodStats {
	var result pojo.TenantDashboardPeriodStats

	result.RechargeAmount = sumDashboardAmount(db.Model(&pojo.RechargeOrder{}).
		Where("tenant_id = ? AND status = ? AND coalesce(is_dev, 0) = 0 AND pay_time >= ? AND pay_time < ?", tenantID, 1, start, end),
		coreRepo.DashboardRechargePaidAmountExpr(""))

	_ = db.Model(&pojo.RechargeOrder{}).
		Where("tenant_id = ? AND status = ? AND coalesce(is_dev, 0) = 0 AND pay_time >= ? AND pay_time < ?", tenantID, 1, start, end).
		Distinct("user_id").
		Count(&result.RechargeUsers).Error

	// 复充人数：成功充值且非首充的去重用户（不含手动回调）
	_ = db.Model(&pojo.RechargeOrder{}).
		Where("tenant_id = ? AND status = ? AND coalesce(is_dev, 0) = 0 AND is_first_recharge = ? AND pay_time >= ? AND pay_time < ?", tenantID, 1, false, start, end).
		Distinct("user_id").
		Count(&result.RepeatRechargeUsers).Error

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
	result.RegisterUsers = countDashboardRegisterUsers(db, tenantID, &start, &end)

	return result
}

func countDashboardRegisterUsers(db *gorm.DB, tenantID int64, start *time.Time, end *time.Time) int64 {
	var total int64
	query := db.Model(&pojo.TgUser{}).Where("tenant_id = ? AND is_bot = ?", tenantID, false)
	if start != nil && end != nil {
		query = query.Where("created_at >= ? AND created_at < ?", *start, *end)
	}
	_ = query.Count(&total).Error
	return total
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
