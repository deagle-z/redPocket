package repository

import (
	"BaseGoUni/core/pojo"
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
	if search.FirstName != "" {
		query = query.Where("first_name like ?", "%"+search.FirstName+"%")
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

func SetTgUser(db *gorm.DB, tenantID int64, req pojo.TgUserSet) (result pojo.TgUserBack, err error) {
	req.TenantId = tenantID
	var dbUser pojo.TgUser
	if req.ID > 0 {
		db.Where("id = ? and tenant_id = ?", req.ID, tenantID).First(&dbUser)
		if dbUser.ID == 0 {
			return result, errors.New("更新的数据不存在")
		}
		_ = copier.Copy(&dbUser, &req)
		err = db.Save(&dbUser).Error
	} else {
		_ = copier.Copy(&dbUser, &req)
		err = db.Create(&dbUser).Error
	}
	if err != nil {
		return result, err
	}
	_ = copier.Copy(&result, &dbUser)
	return result, nil
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
	SubRechargeAmount float64 `json:"subRechargeAmount"`
	SubFlowAmount     float64 `json:"subFlowAmount"`
	SubWithdrawAmount float64 `json:"subWithdrawAmount"`
}

type TgUserWithSubStatsResp struct {
	pojo.BasePageResponse[TgUserWithSubStats]
}

type TgUsersSubStatsSummary struct {
	SubRechargeAmount float64 `json:"subRechargeAmount"`
	SubFlowAmount     float64 `json:"subFlowAmount"`
	SubWithdrawAmount float64 `json:"subWithdrawAmount"`
}

type tgUserMetricRow struct {
	UserId int64   `json:"userId"`
	Amount float64 `json:"amount"`
}

type tgUserTreeAmount struct {
	Recharge float64
	Flow     float64
	Withdraw float64
}

// GetTgUsersWithSubStats 列表并返回所有下级（不限层级）的充值/流水/提现聚合金额
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

	// 构建整棵用户树：parent -> []children
	var allUsers []pojo.TgUser
	allUsersQuery := db.Model(&pojo.TgUser{})
	if tenantID > 0 {
		allUsersQuery = allUsersQuery.Where("tenant_id = ?", tenantID)
	}
	if search.IsBot != nil {
		allUsersQuery = allUsersQuery.Where("is_bot = ?", *search.IsBot)
	}
	_ = allUsersQuery.Find(&allUsers).Error
	childrenMap := make(map[int64][]int64)
	for _, user := range allUsers {
		if user.ParentID == nil {
			continue
		}
		childrenMap[*user.ParentID] = append(childrenMap[*user.ParentID], user.ID)
	}

	var descendantIDs []int64
	if search.ParentID != nil {
		queue := make([]int64, 0, len(childrenMap[*search.ParentID]))
		queue = append(queue, childrenMap[*search.ParentID]...)
		for len(queue) > 0 {
			cur := queue[0]
			queue = queue[1:]
			descendantIDs = append(descendantIDs, cur)
			queue = append(queue, childrenMap[cur]...)
		}
		if len(descendantIDs) == 0 {
			result.PageSize = search.PageSize
			result.CurrentPage = search.CurrentPage
			return result
		}
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
	if search.FirstName != "" {
		query = query.Where("first_name like ?", "%"+search.FirstName+"%")
	}
	if search.IsBot != nil {
		query = query.Where("is_bot = ?", *search.IsBot)
	}
	if search.Status != nil {
		query = query.Where("status = ?", *search.Status)
	}
	if search.ParentID != nil {
		query = query.Where("id in (?)", descendantIDs)
	}
	if search.InviteCode != "" {
		query = query.Where("invite_code = ?", search.InviteCode)
	}
	query.Count(&result.Total)
	query = query.Order("id desc").Limit(search.PageSize).Offset(search.PageSize * search.CurrentPage)
	query.Find(&users)

	// 个人维度金额（非下级）
	rechargeOwn := make(map[int64]float64)
	withdrawOwn := make(map[int64]float64)
	flowOwn := make(map[int64]float64)

	var rechargeSums []tgUserMetricRow
	rechargeQuery := db.Model(&pojo.RechargeOrder{})
	if tenantID > 0 {
		rechargeQuery = rechargeQuery.Where("tenant_id = ?", tenantID)
	}
	_ = rechargeQuery.
		Select("user_id as user_id, sum(amount) as amount").
		Where("status = ?", 2).
		Group("user_id").
		Scan(&rechargeSums).Error
	for _, item := range rechargeSums {
		rechargeOwn[item.UserId] = item.Amount
	}

	var withdrawSums []tgUserMetricRow
	withdrawQuery := db.Model(&pojo.WithdrawOrderBr{})
	if tenantID > 0 {
		withdrawQuery = withdrawQuery.Where("tenant_id = ?", tenantID)
	}
	_ = withdrawQuery.
		Select("user_id as user_id, sum(amount) as amount").
		Where("status = ?", 3).
		Group("user_id").
		Scan(&withdrawSums).Error
	for _, item := range withdrawSums {
		withdrawOwn[item.UserId] = item.Amount
	}

	// 流水口径：lucky_history.amount 聚合
	var flowSums []tgUserMetricRow
	flowQuery := db.Model(&pojo.LuckyHistory{})
	if tenantID > 0 {
		flowQuery = flowQuery.Where("tenant_id = ?", tenantID)
	}
	_ = flowQuery.
		Select("user_id as user_id, sum(amount) as amount").
		Group("user_id").
		Scan(&flowSums).Error
	for _, item := range flowSums {
		flowOwn[item.UserId] = item.Amount
	}

	// DFS 计算每个节点“包含自身+全部后代”的金额
	memo := make(map[int64]tgUserTreeAmount)
	visiting := make(map[int64]bool)
	var calcTotal func(userID int64) tgUserTreeAmount
	calcTotal = func(userID int64) tgUserTreeAmount {
		if v, ok := memo[userID]; ok {
			return v
		}
		if visiting[userID] {
			return tgUserTreeAmount{}
		}
		visiting[userID] = true
		total := tgUserTreeAmount{
			Recharge: rechargeOwn[userID],
			Flow:     flowOwn[userID],
			Withdraw: withdrawOwn[userID],
		}
		for _, childID := range childrenMap[userID] {
			childTotal := calcTotal(childID)
			total.Recharge += childTotal.Recharge
			total.Flow += childTotal.Flow
			total.Withdraw += childTotal.Withdraw
		}
		visiting[userID] = false
		memo[userID] = total
		return total
	}

	for _, user := range users {
		var temp TgUserWithSubStats
		_ = copier.Copy(&temp, &user)
		total := calcTotal(user.ID)
		temp.SubRechargeAmount = total.Recharge - rechargeOwn[user.ID]
		temp.SubFlowAmount = total.Flow - flowOwn[user.ID]
		temp.SubWithdrawAmount = total.Withdraw - withdrawOwn[user.ID]
		result.List = append(result.List, temp)
	}
	fillTenantTgUserWithSubStatsParentUIDs(db, tenantID, result.List)
	result.PageSize = search.PageSize
	result.CurrentPage = search.CurrentPage
	return result
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

// GetTgUsersWithSubStatsSummary 返回下级（不限层级）的充值金额之和、流水之和、提现金额之和
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
			Where("status = ? and user_id in (?)", 2, subUsersQuery).
			Scan(&result.SubRechargeAmount).Error

		flowQuery := db.Model(&pojo.LuckyHistory{})
		if tenantID > 0 {
			flowQuery = flowQuery.Where("tenant_id = ?", tenantID)
		}
		_ = flowQuery.
			Select("coalesce(sum(amount), 0)").
			Where("user_id in (?)", subUsersQuery).
			Scan(&result.SubFlowAmount).Error

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

	var allUsers []pojo.TgUser
	allUsersQuery := db.Model(&pojo.TgUser{})
	if tenantID > 0 {
		allUsersQuery = allUsersQuery.Where("tenant_id = ?", tenantID)
	}
	if search.IsBot != nil {
		allUsersQuery = allUsersQuery.Where("is_bot = ?", *search.IsBot)
	}
	_ = allUsersQuery.Find(&allUsers).Error

	childrenMap := make(map[int64][]int64)
	for _, user := range allUsers {
		if user.ParentID == nil {
			continue
		}
		childrenMap[*user.ParentID] = append(childrenMap[*user.ParentID], user.ID)
	}

	var descendantIDs []int64
	queue := make([]int64, 0, len(childrenMap[*parentID]))
	queue = append(queue, childrenMap[*parentID]...)
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		descendantIDs = append(descendantIDs, cur)
		queue = append(queue, childrenMap[cur]...)
	}

	if len(descendantIDs) == 0 {
		return result
	}

	rechargeQuery := db.Model(&pojo.RechargeOrder{})
	if tenantID > 0 {
		rechargeQuery = rechargeQuery.Where("tenant_id = ?", tenantID)
	}
	_ = rechargeQuery.
		Select("coalesce(sum(amount), 0)").
		Where("status = ? and user_id in (?)", 2, descendantIDs).
		Scan(&result.SubRechargeAmount).Error

	flowQuery := db.Model(&pojo.LuckyHistory{})
	if tenantID > 0 {
		flowQuery = flowQuery.Where("tenant_id = ?", tenantID)
	}
	_ = flowQuery.
		Select("coalesce(sum(amount), 0)").
		Where("user_id in (?)", descendantIDs).
		Scan(&result.SubFlowAmount).Error

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
