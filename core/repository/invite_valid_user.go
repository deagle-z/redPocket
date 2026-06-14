package repository

import (
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/utils"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	InviteValidMinRecharge float64 = 50
	InviteValidMinBet      float64 = 1600
	InviteBetRebateRate    float64 = 0.1
)

// inviteRebateRateForValidCount 按上级当前有效用户数所在档位返回单个用户的返佣金额。
// 档位（按运营图）：1-5→15，6-15→20，16-30→25，31-100→30，100以上→35
func inviteRebateRateForValidCount(validUsers int64) float64 {
	switch {
	case validUsers <= 0:
		return 0
	case validUsers <= 5:
		return 15
	case validUsers <= 15:
		return 20
	case validUsers <= 30:
		return 25
	case validUsers <= 100:
		return 30
	default:
		return 35
	}
}

func EnsureInviteValidUser(tx *gorm.DB, userID int64, qualifiedAt time.Time) (bool, error) {
	if tx == nil || userID <= 0 {
		return false, nil
	}

	var user pojo.TgUser
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", userID).First(&user).Error; err != nil {
		return false, err
	}
	if user.ID == 0 || user.Status == -1 {
		return false, nil
	}
	if user.InviteValidFlag == 1 {
		return true, nil
	}
	if user.ParentID == nil || *user.ParentID <= 0 {
		return false, nil
	}

	totalRecharge := utils.Truncate2(user.RechargeAmount)
	totalBet, err := GetInviteValidUserBetAmount(tx, userID)
	if err != nil {
		return false, err
	}
	if totalRecharge < InviteValidMinRecharge || totalBet < InviteValidMinBet {
		return false, nil
	}
	if qualifiedAt.IsZero() {
		qualifiedAt = time.Now()
	}

	res := tx.Model(&pojo.TgUser{}).
		Where("id = ? AND invite_valid_flag = ?", userID, 0).
		Updates(map[string]any{
			"invite_valid_flag":            int8(1),
			"invite_valid_at":              qualifiedAt,
			"invite_valid_recharge_amount": totalRecharge,
			"invite_valid_bet_amount":      totalBet,
		})
	if res.Error != nil {
		return false, res.Error
	}
	// 新增一名有效用户时，给上级按当前档位发放该用户的返佣
	if res.RowsAffected > 0 {
		if err := GrantInviteRebateForValidUser(tx, *user.ParentID, userID); err != nil {
			return false, err
		}
	}
	return true, nil
}

// GrantInviteRebateForValidUser 当 subUserID 成为 parentID 的有效下级时，
// 按上级当前有效用户数所在档位的单价，给上级发放该下级对应的返佣（每个下级仅一次）。
func GrantInviteRebateForValidUser(tx *gorm.DB, parentID int64, subUserID int64) error {
	if tx == nil || parentID <= 0 || subUserID <= 0 {
		return nil
	}
	var validUsers int64
	if err := tx.Model(&pojo.TgUser{}).
		Where("parent_id = ? AND status <> ? AND invite_valid_flag = ?", parentID, -1, 1).
		Count(&validUsers).Error; err != nil {
		return err
	}
	return grantInviteRebate(tx, parentID, subUserID, validUsers)
}

// grantInviteRebate 按给定的有效用户数(档位)给上级发放某下级的返佣。
// 实时路径：validUsers=当前有效总数（=该下级的达标位次）；补发路径：validUsers=按达标时间升序的位次。
func grantInviteRebate(tx *gorm.DB, parentID int64, subUserID int64, validUsers int64) error {
	if tx == nil || parentID <= 0 || subUserID <= 0 {
		return nil
	}
	amount := inviteRebateRateForValidCount(validUsers)
	if amount <= 0 {
		return nil
	}

	var parent pojo.TgUser
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", parentID).First(&parent).Error; err != nil || parent.ID == 0 {
		return nil
	}
	if parent.Status != 1 {
		return nil
	}

	// 幂等：同一上级对同一下级仅发一次（唯一索引 user_id+sub_user_id）
	logRecord := pojo.TgUserInviteRewardLog{
		TenantID:   parent.TenantId,
		UserID:     parentID,
		SubUserID:  subUserID,
		ValidUsers: int(validUsers),
		Amount:     amount,
	}
	created := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&logRecord)
	if created.Error != nil {
		return created.Error
	}
	if created.RowsAffected == 0 {
		return nil
	}

	// 仅入账余额；提现限制走 v2 流水批次
	if err := tx.Model(&pojo.TgUser{}).Where("id = ?", parentID).Updates(map[string]any{
		"balance": gorm.Expr("balance + ?", amount),
	}).Error; err != nil {
		return err
	}

	history := pojo.CashHistory{
		UserId:          parentID,
		AwardUni:        fmt.Sprintf("invite_rebate_%d_%d", parentID, subUserID),
		Amount:          amount,
		StartAmount:     utils.Truncate2(parent.Balance),
		EndAmount:       utils.Truncate2(parent.Balance + amount),
		CashMark:        "邀请返佣奖励",
		CashDesc:        fmt.Sprintf("有效用户返佣%.2f（当前有效用户%d人）", amount, validUsers),
		Type:            pojo.CashHistoryTypeInviteRebateTierReward,
		IsGift:          1,
		FromUserId:      subUserID,
		SourceChannelID: parent.SourceChannelID,
	}
	if err := tx.Create(&history).Error; err != nil {
		return err
	}

	return EnsureWithdrawFlowBatchForGift(
		tx, parent,
		pojo.WithdrawFlowBatchSourceInviteRebate,
		logRecord.ID,
		fmt.Sprintf("invite_rebate_%d", logRecord.ID),
		pojo.WithdrawFlowBatchSourceInviteRebate,
		amount,
	)
}

// BackfillInviteRebateTiers 对历史存量有效用户补发邀请返佣（一次性脚本，幂等可重复执行）。
// 传入带表前缀的 db（如 utils.NewPrefixDb(prefix)）。逐个上级单独事务处理，返回处理的上级数量。
// 口径：每个上级的有效下级按达标时间(invite_valid_at)升序，第 k 个按当前有效数 k 所在档位单价发放。
func BackfillInviteRebateTiers(db *gorm.DB) (int, error) {
	if db == nil {
		return 0, nil
	}
	var parentIDs []int64
	if err := db.Model(&pojo.TgUser{}).
		Where("parent_id IS NOT NULL AND parent_id > 0 AND status <> ? AND invite_valid_flag = ?", -1, 1).
		Distinct().
		Pluck("parent_id", &parentIDs).Error; err != nil {
		return 0, err
	}
	processed := 0
	for _, pid := range parentIDs {
		if pid <= 0 {
			continue
		}
		if err := db.Transaction(func(tx *gorm.DB) error {
			return backfillInviteRebateForParent(tx, pid)
		}); err != nil {
			return processed, fmt.Errorf("backfill invite rebate parent=%d: %w", pid, err)
		}
		processed++
	}
	return processed, nil
}

// backfillInviteRebateForParent 按有效下级达标时间升序，逐个补发（幂等，已发过的下级跳过）。
func backfillInviteRebateForParent(tx *gorm.DB, parentID int64) error {
	var subUserIDs []int64
	if err := tx.Model(&pojo.TgUser{}).
		Where("parent_id = ? AND status <> ? AND invite_valid_flag = ?", parentID, -1, 1).
		Order("invite_valid_at asc, id asc").
		Pluck("id", &subUserIDs).Error; err != nil {
		return err
	}
	for idx, subUserID := range subUserIDs {
		if err := grantInviteRebate(tx, parentID, subUserID, int64(idx+1)); err != nil {
			return err
		}
	}
	return nil
}

func GetInviteValidUserBetAmount(db *gorm.DB, userID int64) (float64, error) {
	if db == nil || userID <= 0 {
		return 0, nil
	}

	var totalBet float64
	table := pojo.AppUserBetRecordTableNameByUserID(userID)
	if !db.Migrator().HasTable(table) {
		return 0, nil
	}
	err := db.Table(table).
		Where("user_id = ? AND COALESCE(deleted_flag, 0) = 0 AND COALESCE(bet_amount, 0) > 0", userID).
		Select("COALESCE(SUM(COALESCE(bet_amount, 0)), 0)").
		Scan(&totalBet).Error
	if err != nil {
		return 0, err
	}
	return utils.Truncate2(totalBet), nil
}

func ApplyInviteBetRebate(tx *gorm.DB, subUser pojo.TgUser, betAmount float64, traceID string, occurredAt time.Time) error {
	if tx == nil || subUser.ID <= 0 {
		return nil
	}
	betAmount = utils.Truncate2(betAmount)
	traceID = strings.TrimSpace(traceID)
	if betAmount <= 0 || traceID == "" {
		return nil
	}
	if subUser.ParentID == nil || *subUser.ParentID <= 0 {
		return nil
	}

	valid, err := EnsureInviteValidUser(tx, subUser.ID, occurredAt)
	if err != nil || !valid {
		return err
	}

	parentID := *subUser.ParentID
	var parent pojo.TgUser
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", parentID).First(&parent).Error; err != nil || parent.ID == 0 {
		return nil
	}
	if parent.Status != 1 {
		return nil
	}

	rebateAmount := utils.Truncate2(utils.ToMoney(betAmount).Multiply(InviteBetRebateRate / 100).ToDollars())
	if rebateAmount < 0.01 {
		return nil
	}
	if occurredAt.IsZero() {
		occurredAt = time.Now()
	}

	tenantID := subUser.TenantId
	remark := "invite_bet_rebate"
	record := pojo.TgUserRebateRecord{
		TenantId:        &tenantID,
		SubUserId:       subUser.ID,
		ParentUserId:    parentID,
		SourceChannelID: subUser.SourceChannelID,
		SourceType:      pojo.TgUserRebateSourceTypeBetFlow,
		SourceOrderId:   traceID,
		SourceAmount:    betAmount,
		RebateRate:      InviteBetRebateRate,
		RebateAmount:    rebateAmount,
		Currency:        "USDT",
		Status:          1,
		SettledAt:       &occurredAt,
		IdempotencyKey:  fmt.Sprintf("invite_bet_rebate:%s:%d", traceID, parentID),
		Remark:          &remark,
	}
	result := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&record)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return nil
	}

	return tx.Model(&pojo.TgUser{}).
		Where("id = ?", parentID).
		Updates(map[string]any{
			"rebate_amount":       gorm.Expr("rebate_amount + ?", rebateAmount),
			"rebate_total_amount": gorm.Expr("rebate_total_amount + ?", rebateAmount),
		}).Error
}
