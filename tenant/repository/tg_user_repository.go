package repository

import (
	"BaseGoUni/core/pojo"
	coreRepo "BaseGoUni/core/repository"
	"BaseGoUni/core/utils"
	"errors"
	"github.com/jinzhu/copier"
	"gorm.io/gorm"
	"strings"
)

func GetTgUsers(db *gorm.DB, tenantID int64, search pojo.TgUserSearch) (result pojo.TgUserResp) {
	var users []pojo.TgUser
	query := db.Model(&pojo.TgUser{}).Where("tenant_id = ?", tenantID)
	if search.TgID > 0 {
		query = query.Where("tg_id = ?", search.TgID)
	}
	if uid := strings.TrimSpace(search.Uid); uid != "" {
		query = query.Where("uid = ?", uid)
	}
	if search.Username != "" {
		query = query.Where("username like ?", "%"+search.Username+"%")
	}
	if search.TgName != "" {
		query = query.Where("tg_name like ?", "%"+strings.TrimSpace(search.TgName)+"%")
	}
	if search.FirstName != "" {
		query = query.Where("first_name like ?", "%"+search.FirstName+"%")
	}
	if phone := strings.TrimSpace(search.Phone); phone != "" {
		query = query.Where("phone like ?", "%"+phone+"%")
	}
	if search.IsBot != nil {
		query = query.Where("is_bot = ?", *search.IsBot)
	}
	if search.Status != nil {
		query = query.Where("status = ?", *search.Status)
	}
	if search.ParentID != nil {
		query = query.Where("parent_id = ?", *search.ParentID)
	}
	if search.ParentUid != "" {
		query = query.Where("parent_id IN (?)", db.Model(&pojo.TgUser{}).
			Select("id").
			Where("tenant_id = ? AND uid = ?", tenantID, strings.TrimSpace(search.ParentUid)))
	}
	if search.InviteCode != "" {
		query = query.Where("invite_code = ?", search.InviteCode)
	}
	query.Count(&result.Total)
	query = query.Order("id desc").Limit(search.PageSize).Offset(search.PageSize * search.CurrentPage)
	query.Find(&users)
	for _, user := range users {
		var temp pojo.TgUserBack
		_ = copier.Copy(&temp, &user)
		result.List = append(result.List, temp)
	}
	fillTenantTgUserParentUIDs(db, tenantID, result.List)
	result.PageSize = search.PageSize
	result.CurrentPage = search.CurrentPage
	return result
}

func GetTgUserByID(db *gorm.DB, tenantID int64, id int64) (result pojo.TgUserBack, err error) {
	var dbUser pojo.TgUser
	db.Where("id = ? and tenant_id = ?", id, tenantID).First(&dbUser)
	if dbUser.ID == 0 {
		return result, errors.New("数据不存在")
	}
	_ = copier.Copy(&result, &dbUser)
	return result, nil
}

func GetTgUserWithdrawActivityFlow(db *gorm.DB, tenantID int64, id int64) (pojo.TgWithdrawActivityFlowBack, error) {
	var user pojo.TgUser
	if err := db.Select("id").Where("id = ? and tenant_id = ?", id, tenantID).First(&user).Error; err != nil {
		return pojo.TgWithdrawActivityFlowBack{}, err
	}
	if user.ID == 0 {
		return pojo.TgWithdrawActivityFlowBack{}, errors.New("数据不存在")
	}
	return coreRepo.GetUserWithdrawActivityFlow(db, user.ID)
}

func SetTgUser(db *gorm.DB, tenantID int64, req pojo.TgUserSet) (result pojo.TgUserBack, err error) {
	req.TenantId = tenantID
	var dbUser pojo.TgUser
	if req.ID > 0 {
		db.Where("id = ? and tenant_id = ?", req.ID, tenantID).First(&dbUser)
		if dbUser.ID == 0 {
			return result, errors.New("更新的数据不存在")
		}
		updates := buildTenantTgUserUpdateMap(req)
		if len(updates) > 0 {
			err = db.Model(&pojo.TgUser{}).Where("id = ? and tenant_id = ?", dbUser.ID, tenantID).Updates(updates).Error
		}
		if err == nil {
			db.Where("id = ? and tenant_id = ?", dbUser.ID, tenantID).First(&dbUser)
		}
	} else {
		_ = copier.Copy(&dbUser, &req)
		if req.FreeLotteryCount != nil {
			dbUser.FreeLotteryCount = *req.FreeLotteryCount
		}
		err = db.Create(&dbUser).Error
	}
	if err != nil {
		return result, err
	}
	_ = copier.Copy(&result, &dbUser)
	return result, nil
}

func buildTenantTgUserUpdateMap(req pojo.TgUserSet) map[string]any {
	updates := map[string]any{
		"username":            req.Username,
		"tg_name":             req.TgName,
		"first_name":          req.FirstName,
		"avatar":              req.Avatar,
		"phone":               req.Phone,
		"country":             req.Country,
		"ip":                  req.Ip,
		"region":              req.Region,
		"remark":              req.Remark,
		"is_bot":              req.IsBot,
		"tg_id":               req.TgID,
		"sport_balance":       utils.Truncate2(req.SportBalance),
		"status":              req.Status,
		"parent_id":           req.ParentID,
		"invite_code":         req.InviteCode,
		"source_channel_id":   req.SourceChannelID,
		"source_channel_code": req.SourceChannelCode,
		"tenant_id":           req.TenantId,
	}
	if req.RebateRate != nil {
		updates["rebate_rate"] = utils.Truncate2(*req.RebateRate)
	}
	if req.FreeLotteryCount != nil {
		updates["free_lottery_count"] = *req.FreeLotteryCount
	}
	return updates
}

func SetTgUserStatus(db *gorm.DB, tenantID int64, id int64, status int8) (result pojo.TgUserBack, err error) {
	var dbUser pojo.TgUser
	db.Where("id = ? and tenant_id = ?", id, tenantID).First(&dbUser)
	if dbUser.ID == 0 {
		return result, errors.New("数据不存在")
	}
	err = db.Model(&dbUser).Update("status", status).Error
	if err != nil {
		return result, err
	}
	_ = copier.Copy(&result, &dbUser)
	result.Status = status
	return result, nil
}

func SetTgUserRebateRate(db *gorm.DB, tenantID int64, id int64, rebateRate float64) (result pojo.TgUserBack, err error) {
	var dbUser pojo.TgUser
	db.Where("id = ? and tenant_id = ?", id, tenantID).First(&dbUser)
	if dbUser.ID == 0 {
		return result, errors.New("数据不存在")
	}
	rebateRate = utils.Truncate2(rebateRate)
	err = db.Model(&dbUser).Update("rebate_rate", rebateRate).Error
	if err != nil {
		return result, err
	}
	_ = copier.Copy(&result, &dbUser)
	result.RebateRate = rebateRate
	return result, nil
}

func SetTgUserRebateType(db *gorm.DB, tenantID int64, id int64, rebateType int8) (result pojo.TgUserBack, err error) {
	var dbUser pojo.TgUser
	db.Where("id = ? and tenant_id = ?", id, tenantID).First(&dbUser)
	if dbUser.ID == 0 {
		return result, errors.New("数据不存在")
	}
	err = db.Model(&dbUser).Update("rebate_type", rebateType).Error
	if err != nil {
		return result, err
	}
	_ = copier.Copy(&result, &dbUser)
	result.RebateType = rebateType
	return result, nil
}

func SetTgUserRebateWithdrawDisabled(db *gorm.DB, tenantID int64, id int64, disabled int8) (result pojo.TgUserBack, err error) {
	var dbUser pojo.TgUser
	db.Where("id = ? and tenant_id = ?", id, tenantID).First(&dbUser)
	if dbUser.ID == 0 {
		return result, errors.New("数据不存在")
	}
	err = db.Model(&dbUser).Update("rebate_withdraw_disabled", disabled).Error
	if err != nil {
		return result, err
	}
	_ = copier.Copy(&result, &dbUser)
	result.RebateWithdrawDisabled = disabled
	return result, nil
}

func SetTgUserRechargeRebateRates(db *gorm.DB, tenantID int64, id int64, rates string) (result pojo.TgUserAdminBack, err error) {
	var dbUser pojo.TgUser
	db.Select("id").Where("id = ? and tenant_id = ?", id, tenantID).First(&dbUser)
	if dbUser.ID == 0 {
		return result, errors.New("数据不存在")
	}
	return coreRepo.SetTgUserRechargeRebateRates(db, id, rates)
}

func AddTgUserRebateAmount(db *gorm.DB, tenantID int64, id int64, amount float64) (result pojo.TgUserAdminBack, err error) {
	var dbUser pojo.TgUser
	db.Select("id").Where("id = ? and tenant_id = ?", id, tenantID).First(&dbUser)
	if dbUser.ID == 0 {
		return result, errors.New("数据不存在")
	}
	return coreRepo.AddTgUserRebateAmount(db, id, amount)
}

func AdminCreateRechargeOrderV2(db *gorm.DB, tenantID int64, req pojo.AdminCreateRechargeOrderV2Req, tablePrefix string) (pojo.RechargeOrderAppBack, error) {
	var dbUser pojo.TgUser
	db.Select("id").Where("id = ? and tenant_id = ?", req.UserID, tenantID).First(&dbUser)
	if dbUser.ID == 0 {
		return pojo.RechargeOrderAppBack{}, errors.New("数据不存在")
	}
	return coreRepo.AdminCreateRechargeOrderV2(db, req.UserID, req.RechargeOrderAppReq, tablePrefix)
}

func AdminCreateCryptoRechargeOrder(db *gorm.DB, tenantID int64, req pojo.CryptoRechargeOrderReq, tablePrefix string) (pojo.RechargeOrderAppBack, error) {
	var dbUser pojo.TgUser
	db.Select("id").Where("id = ? and tenant_id = ?", req.UserID, tenantID).First(&dbUser)
	if dbUser.ID == 0 {
		return pojo.RechargeOrderAppBack{}, errors.New("数据不存在")
	}
	return coreRepo.AdminCreateCryptoRechargeOrder(db, req.UserID, req, tablePrefix)
}

func SetTgUserRemark(db *gorm.DB, tenantID int64, id int64, remark string) (result pojo.TgUserBack, err error) {
	var dbUser pojo.TgUser
	db.Where("id = ? and tenant_id = ?", id, tenantID).First(&dbUser)
	if dbUser.ID == 0 {
		return result, errors.New("数据不存在")
	}
	remark = strings.TrimSpace(remark)
	if len([]rune(remark)) > 255 {
		return result, errors.New("备注不能超过255个字符")
	}
	var remarkPtr *string
	if remark != "" {
		remarkPtr = &remark
	}
	err = db.Model(&dbUser).Update("remark", remarkPtr).Error
	if err != nil {
		return result, err
	}
	_ = copier.Copy(&result, &dbUser)
	result.Remark = remarkPtr
	return result, nil
}

func DelTgUser(db *gorm.DB, tenantID int64, id int64) (result string, err error) {
	var dbUser pojo.TgUser
	db.Where("id = ? and tenant_id = ?", id, tenantID).First(&dbUser)
	if dbUser.ID == 0 {
		return result, errors.New("删除的数据不存在")
	}
	err = db.Delete(&dbUser).Error
	if err != nil {
		return result, err
	}
	return "success", nil
}

type TgUserWithSubStats struct {
	pojo.TgUserBack
	SubUserCount      int64   `json:"subUserCount"`
	SubRechargeAmount float64 `json:"subRechargeAmount"`
	SubFlowAmount     float64 `json:"subFlowAmount"`
	SubProfitAmount   float64 `json:"subProfitAmount"`
	SubWithdrawAmount float64 `json:"subWithdrawAmount"`
}

type TgUserWithSubStatsResp struct {
	pojo.BasePageResponse[TgUserWithSubStats]
}

type TgUsersSubStatsSummary struct {
	SubRechargeAmount float64 `json:"subRechargeAmount"`
	SubFlowAmount     float64 `json:"subFlowAmount"`
	SubProfitAmount   float64 `json:"subProfitAmount"`
	SubWithdrawAmount float64 `json:"subWithdrawAmount"`
	RechargeUsers     int64   `json:"rechargeUsers"`
	ValidUsers        int64   `json:"validUsers"`
}

type tgUserMetricRow struct {
	UserId int64   `json:"userId"`
	Amount float64 `json:"amount"`
}

type tgUserTreeAmount struct {
	Recharge float64
	Flow     float64
	Profit   float64
	Withdraw float64
}

func luckyHistoryProfitSQL() string {
	return "sum(case when is_thunder = 0 then coalesce(nullif(actual_amount, 0), amount) else -lose_money end)"
}

type tgUserBetStat struct {
	Flow   float64
	Profit float64
}

// tenantTgUserBetStatsByUser 按用户聚合投注流水与盈利（投注记录分表，按分片分组查询）。
// 流水 = sum(bet_amount)；盈利 = sum(bet_amount - win_amount)（平台对玩家盈利）。
func tenantTgUserBetStatsByUser(db *gorm.DB, userIDs []int64) map[int64]tgUserBetStat {
	statByUser := make(map[int64]tgUserBetStat, len(userIDs))
	if len(userIDs) == 0 {
		return statByUser
	}

	shardGroups := make(map[int][]int64)
	for _, uid := range userIDs {
		idx := pojo.AppUserBetRecordShardIndex(uid)
		shardGroups[idx] = append(shardGroups[idx], uid)
	}

	type betRow struct {
		UserID int64   `gorm:"column:user_id"`
		Flow   float64 `gorm:"column:flow"`
		Profit float64 `gorm:"column:profit"`
	}
	for idx, ids := range shardGroups {
		table := pojo.AppUserBetRecordShardTableName(idx)
		if !db.Migrator().HasTable(table) {
			continue
		}
		var rows []betRow
		_ = db.Table(table).
			Select("user_id as user_id, "+
				"coalesce(sum(coalesce(bet_amount, 0)), 0) as flow, "+
				"coalesce(sum(coalesce(bet_amount, 0) - coalesce(win_amount, 0)), 0) as profit").
			Where("user_id in (?) and coalesce(deleted_flag, 0) = 0", ids).
			Group("user_id").
			Scan(&rows).Error
		for _, r := range rows {
			statByUser[r.UserID] = tgUserBetStat{
				Flow:   utils.Truncate2(r.Flow),
				Profit: utils.Truncate2(r.Profit),
			}
		}
	}
	return statByUser
}

// tenantTgUserBetStatsTotal 汇总一组用户的投注流水之和与盈利之和。
func tenantTgUserBetStatsTotal(db *gorm.DB, userIDs []int64) (flow float64, profit float64) {
	statByUser := tenantTgUserBetStatsByUser(db, userIDs)
	for _, s := range statByUser {
		flow += s.Flow
		profit += s.Profit
	}
	return utils.Truncate2(flow), utils.Truncate2(profit)
}

// GetTgUsersWithSubStats 分页列出用户；传 parentID 时列出直属下级，每行返回该用户所有下级（不限层级）的充值/流水/盈利/提现聚合金额
func GetTgUsersWithSubStats(db *gorm.DB, tenantID int64, search pojo.TgUserSearch) (result TgUserWithSubStatsResp) {
	if search.ParentID == nil && search.ParentUid != "" {
		parentUid := strings.TrimSpace(search.ParentUid)
		parentQuery := db.Model(&pojo.TgUser{}).Select("id").Where("uid = ?", parentUid)
		if tenantID > 0 {
			parentQuery = parentQuery.Where("tenant_id = ?", tenantID)
		}
		var parent pojo.TgUser
		_ = parentQuery.First(&parent).Error
		if parent.ID == 0 {
			result.PageSize = search.PageSize
			result.CurrentPage = search.CurrentPage
			return result
		}
		search.ParentID = &parent.ID
	}

	var users []pojo.TgUser
	query := db.Model(&pojo.TgUser{})
	if tenantID > 0 {
		query = query.Where("tenant_id = ?", tenantID)
	}
	if search.TgID > 0 {
		query = query.Where("tg_id = ?", search.TgID)
	}
	if uid := strings.TrimSpace(search.Uid); uid != "" {
		query = query.Where("uid = ?", uid)
	}
	if search.Username != "" {
		query = query.Where("username like ?", "%"+search.Username+"%")
	}
	if search.TgName != "" {
		query = query.Where("tg_name like ?", "%"+strings.TrimSpace(search.TgName)+"%")
	}
	if search.FirstName != "" {
		query = query.Where("first_name like ?", "%"+search.FirstName+"%")
	}
	if phone := strings.TrimSpace(search.Phone); phone != "" {
		query = query.Where("phone like ?", "%"+phone+"%")
	}
	if search.IsBot != nil {
		query = query.Where("is_bot = ?", *search.IsBot)
	}
	if search.Status != nil {
		query = query.Where("status = ?", *search.Status)
	}
	if search.ParentID != nil {
		query = query.Where("parent_id = ?", *search.ParentID)
	}
	if search.InviteCode != "" {
		query = query.Where("invite_code = ?", search.InviteCode)
	}

	// 按充值金额倒序：LEFT JOIN 充值汇总子查询（status=1 成功充值）
	rechargeSub := db.Model(&pojo.RechargeOrder{}).
		Select("user_id, sum(amount) as recharge_amt").
		Where("status = ? and coalesce(is_dev, 0) = 0", 1)
	if tenantID > 0 {
		rechargeSub = rechargeSub.Where("tenant_id = ?", tenantID)
	}
	rechargeSub = rechargeSub.Group("user_id")
	query = query.Joins("LEFT JOIN (?) AS rsum ON rsum.user_id = tg_user.id", rechargeSub)

	query.Count(&result.Total)
	query = query.Order("coalesce(rsum.recharge_amt, 0) desc").
		Order("tg_user.id desc").
		Limit(search.PageSize).Offset(search.PageSize * search.CurrentPage)
	query.Find(&users)

	// 口径：每行只统计用户本人
	metricUserIDs := make([]int64, 0, len(users))
	for _, user := range users {
		metricUserIDs = append(metricUserIDs, user.ID)
	}

	rechargeSumsByUser := make(map[int64]float64)
	withdrawSumsByUser := make(map[int64]float64)
	subUserCountsByUser := make(map[int64]int64)

	if len(metricUserIDs) > 0 {
		var subUserCounts []struct {
			UserID       int64 `gorm:"column:user_id"`
			SubUserCount int64 `gorm:"column:sub_user_count"`
		}
		subUserQuery := db.Model(&pojo.TgUser{})
		if tenantID > 0 {
			subUserQuery = subUserQuery.Where("tenant_id = ?", tenantID)
		}
		_ = subUserQuery.
			Select("parent_id as user_id, count(*) as sub_user_count").
			Where("parent_id in (?) and status <> ?", metricUserIDs, -1).
			Group("parent_id").
			Scan(&subUserCounts).Error
		for _, item := range subUserCounts {
			subUserCountsByUser[item.UserID] = item.SubUserCount
		}

		var rechargeSums []tgUserMetricRow
		rechargeQuery := db.Model(&pojo.RechargeOrder{})
		if tenantID > 0 {
			rechargeQuery = rechargeQuery.Where("tenant_id = ?", tenantID)
		}
		_ = rechargeQuery.
			Select("user_id as user_id, sum(amount) as amount").
			Where("status = ? and coalesce(is_dev, 0) = 0 and user_id in (?)", 1, metricUserIDs).
			Group("user_id").
			Scan(&rechargeSums).Error
		for _, item := range rechargeSums {
			rechargeSumsByUser[item.UserId] = item.Amount
		}

		var withdrawSums []tgUserMetricRow
		withdrawQuery := db.Model(&pojo.WithdrawOrderBr{})
		if tenantID > 0 {
			withdrawQuery = withdrawQuery.Where("tenant_id = ?", tenantID)
		}
		_ = withdrawQuery.
			Select("user_id as user_id, sum(amount) as amount").
			Where("status = ? and user_id in (?)", 3, metricUserIDs).
			Group("user_id").
			Scan(&withdrawSums).Error
		for _, item := range withdrawSums {
			withdrawSumsByUser[item.UserId] = item.Amount
		}
	}

	// 流水/盈利口径：遍历 user_id 查询投注记录（分表）
	betStatByUser := tenantTgUserBetStatsByUser(db, metricUserIDs)

	for _, user := range users {
		var temp TgUserWithSubStats
		_ = copier.Copy(&temp, &user)
		temp.SubUserCount = subUserCountsByUser[user.ID]
		temp.SubRechargeAmount = utils.Truncate2(rechargeSumsByUser[user.ID])
		temp.SubWithdrawAmount = utils.Truncate2(withdrawSumsByUser[user.ID])
		temp.SubFlowAmount = betStatByUser[user.ID].Flow
		temp.SubProfitAmount = betStatByUser[user.ID].Profit
		result.List = append(result.List, temp)
	}
	fillTenantTgUserWithSubStatsParentUIDs(db, tenantID, result.List)
	result.PageSize = search.PageSize
	result.CurrentPage = search.CurrentPage
	return result
}

func collectTenantTgUserDescendantIDs(childrenMap map[int64][]int64, rootID int64) []int64 {
	descendantIDs := make([]int64, 0)
	seen := map[int64]struct{}{rootID: {}}
	queue := append([]int64(nil), childrenMap[rootID]...)
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		if _, ok := seen[cur]; ok {
			continue
		}
		seen[cur] = struct{}{}
		descendantIDs = append(descendantIDs, cur)
		queue = append(queue, childrenMap[cur]...)
	}
	return descendantIDs
}

func tenantTgUserIDSetToSlice(idSet map[int64]struct{}) []int64 {
	if len(idSet) == 0 {
		return nil
	}
	ids := make([]int64, 0, len(idSet))
	for id := range idSet {
		ids = append(ids, id)
	}
	return ids
}

func tenantTgUserSubStatsTotal(userIDs []int64, rechargeSumsByUser, flowSumsByUser, profitSumsByUser, withdrawSumsByUser map[int64]float64) tgUserTreeAmount {
	var total tgUserTreeAmount
	for _, userID := range userIDs {
		total.Recharge += rechargeSumsByUser[userID]
		total.Flow += flowSumsByUser[userID]
		total.Profit += profitSumsByUser[userID]
		total.Withdraw += withdrawSumsByUser[userID]
	}
	return total
}

func fillTenantTgUserParentUIDs(db *gorm.DB, tenantID int64, users []pojo.TgUserBack) {
	parentUIDMap := getTenantParentUIDMap(db, tenantID, users)
	if len(parentUIDMap) == 0 {
		return
	}

	for i := range users {
		if users[i].ParentID == nil {
			continue
		}
		if uid, ok := parentUIDMap[*users[i].ParentID]; ok && uid != "" {
			users[i].ParentUid = &uid
		}
	}
}

func fillTenantTgUserWithSubStatsParentUIDs(db *gorm.DB, tenantID int64, users []TgUserWithSubStats) {
	baseUsers := make([]pojo.TgUserBack, 0, len(users))
	for _, user := range users {
		baseUsers = append(baseUsers, user.TgUserBack)
	}
	parentUIDMap := getTenantParentUIDMap(db, tenantID, baseUsers)
	if len(parentUIDMap) == 0 {
		return
	}

	for i := range users {
		if users[i].ParentID == nil {
			continue
		}
		if uid, ok := parentUIDMap[*users[i].ParentID]; ok && uid != "" {
			users[i].ParentUid = &uid
		}
	}
}

func getTenantParentUIDMap(db *gorm.DB, tenantID int64, users []pojo.TgUserBack) map[int64]string {
	parentIDs := make([]int64, 0, len(users))
	seen := make(map[int64]struct{}, len(users))
	for _, user := range users {
		if user.ParentID == nil {
			continue
		}
		parentID := *user.ParentID
		if _, ok := seen[parentID]; ok {
			continue
		}
		seen[parentID] = struct{}{}
		parentIDs = append(parentIDs, parentID)
	}
	if len(parentIDs) == 0 {
		return nil
	}

	var parents []pojo.TgUser
	query := db.Model(&pojo.TgUser{}).Select("id, uid").Where("id IN ?", parentIDs)
	if tenantID > 0 {
		query = query.Where("tenant_id = ?", tenantID)
	}
	_ = query.Find(&parents).Error

	parentUIDMap := make(map[int64]string, len(parents))
	for _, parent := range parents {
		parentUIDMap[parent.ID] = parent.Uid
	}
	return parentUIDMap
}

// GetTgUsersWithSubStatsSummary 返回下级（不限层级）的充值金额之和、流水之和、盈利之和、提现金额之和、有效用户数
// 有效用户口径：有成功充值记录的下级用户数（status=1 且 is_dev=0），按 user_id 去重。
// parentID 为空：口径为全量 parent_id 非空的用户集合
// parentID 非空：口径为该 parentID 的所有后代（不含自身）
func GetTgUsersWithSubStatsSummary(db *gorm.DB, tenantID int64, search pojo.TgUserSearch) (result TgUsersSubStatsSummary) {
	parentID := search.ParentID
	if parentID == nil {
		subUsersQuery := db.Model(&pojo.TgUser{}).
			Select("id").
			Where("parent_id is not null")
		if tenantID > 0 {
			subUsersQuery = subUsersQuery.Where("tenant_id = ?", tenantID)
		}
		if search.IsBot != nil {
			subUsersQuery = subUsersQuery.Where("is_bot = ?", *search.IsBot)
		}

		rechargeQuery := db.Model(&pojo.RechargeOrder{})
		if tenantID > 0 {
			rechargeQuery = rechargeQuery.Where("tenant_id = ?", tenantID)
		}
		_ = rechargeQuery.
			Select("coalesce(sum(amount), 0)").
			Where("status = ? and coalesce(is_dev, 0) = 0 and user_id in (?)", 1, subUsersQuery).
			Scan(&result.SubRechargeAmount).Error

		rechargeUsersQuery := db.Model(&pojo.RechargeOrder{})
		if tenantID > 0 {
			rechargeUsersQuery = rechargeUsersQuery.Where("tenant_id = ?", tenantID)
		}
		_ = rechargeUsersQuery.
			Where("status = ? and coalesce(is_dev, 0) = 0 and user_id in (?)", 1, subUsersQuery).
			Distinct("user_id").
			Count(&result.RechargeUsers).Error
		result.ValidUsers = result.RechargeUsers

		// 流水/盈利口径：遍历 user_id 查询投注记录（分表）
		var subUserIDs []int64
		subIDQuery := db.Model(&pojo.TgUser{}).Where("parent_id is not null")
		if tenantID > 0 {
			subIDQuery = subIDQuery.Where("tenant_id = ?", tenantID)
		}
		if search.IsBot != nil {
			subIDQuery = subIDQuery.Where("is_bot = ?", *search.IsBot)
		}
		_ = subIDQuery.Pluck("id", &subUserIDs).Error
		result.SubFlowAmount, result.SubProfitAmount = tenantTgUserBetStatsTotal(db, subUserIDs)

		withdrawQuery := db.Model(&pojo.WithdrawOrderBr{})
		if tenantID > 0 {
			withdrawQuery = withdrawQuery.Where("tenant_id = ?", tenantID)
		}
		_ = withdrawQuery.
			Select("coalesce(sum(amount), 0)").
			Where("status = ? and user_id in (?)", 3, subUsersQuery).
			Scan(&result.SubWithdrawAmount).Error
		return result
	}

	// 口径：只统计直接下级（parent_id = parentID），不递归整棵子树
	var descendantIDs []int64
	directChildrenQuery := db.Model(&pojo.TgUser{}).Where("parent_id = ?", *parentID)
	if tenantID > 0 {
		directChildrenQuery = directChildrenQuery.Where("tenant_id = ?", tenantID)
	}
	if search.IsBot != nil {
		directChildrenQuery = directChildrenQuery.Where("is_bot = ?", *search.IsBot)
	}
	_ = directChildrenQuery.Pluck("id", &descendantIDs).Error

	if len(descendantIDs) == 0 {
		return result
	}

	rechargeQuery := db.Model(&pojo.RechargeOrder{})
	if tenantID > 0 {
		rechargeQuery = rechargeQuery.Where("tenant_id = ?", tenantID)
	}
	_ = rechargeQuery.
		Select("coalesce(sum(amount), 0)").
		Where("status = ? and coalesce(is_dev, 0) = 0 and user_id in (?)", 1, descendantIDs).
		Scan(&result.SubRechargeAmount).Error

	rechargeUsersQuery := db.Model(&pojo.RechargeOrder{})
	if tenantID > 0 {
		rechargeUsersQuery = rechargeUsersQuery.Where("tenant_id = ?", tenantID)
	}
	_ = rechargeUsersQuery.
		Where("status = ? and coalesce(is_dev, 0) = 0 and user_id in (?)", 1, descendantIDs).
		Distinct("user_id").
		Count(&result.RechargeUsers).Error
	result.ValidUsers = result.RechargeUsers

	// 流水/盈利口径：遍历 user_id 查询投注记录（分表）
	result.SubFlowAmount, result.SubProfitAmount = tenantTgUserBetStatsTotal(db, descendantIDs)

	withdrawQuery := db.Model(&pojo.WithdrawOrderBr{})
	if tenantID > 0 {
		withdrawQuery = withdrawQuery.Where("tenant_id = ?", tenantID)
	}
	_ = withdrawQuery.
		Select("coalesce(sum(amount), 0)").
		Where("status = ? and user_id in (?)", 3, descendantIDs).
		Scan(&result.SubWithdrawAmount).Error

	return result
}
