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

// inviteRebateTier 邀请返佣阶梯：有效用户数达到 Threshold 时，给上级发放 Amount（每档每人仅一次，可累计）
type inviteRebateTier struct {
	Level     int
	Threshold int
	Amount    float64
}

// 按运营图写死：1人→15, 5人→20, 16人→25, 30人→30, 100人→35（阈值取每档区间起点，可累计）
var inviteRebateTiers = []inviteRebateTier{
	{Level: 1, Threshold: 1, Amount: 15},
	{Level: 2, Threshold: 5, Amount: 20},
	{Level: 3, Threshold: 16, Amount: 25},
	{Level: 4, Threshold: 30, Amount: 30},
	{Level: 5, Threshold: 100, Amount: 35},
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
	// 新增一名有效用户时，给上级结算邀请返佣阶梯奖励
	if res.RowsAffected > 0 {
		if err := GrantInviteRebateTiers(tx, *user.ParentID); err != nil {
			return false, err
		}
	}
	return true, nil
}

// GrantInviteRebateTiers 按当前有效用户数为上级发放尚未发放的阶梯奖励（自动发放，每档每人仅一次，可累计）。
func GrantInviteRebateTiers(tx *gorm.DB, parentID int64) error {
	if tx == nil || parentID <= 0 {
		return nil
	}

	var validUsers int64
	if err := tx.Model(&pojo.TgUser{}).
		Where("parent_id = ? AND status <> ? AND invite_valid_flag = ?", parentID, -1, 1).
		Count(&validUsers).Error; err != nil {
		return err
	}
	if validUsers <= 0 {
		return nil
	}

	var parent pojo.TgUser
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", parentID).First(&parent).Error; err != nil || parent.ID == 0 {
		return nil
	}
	if parent.Status != 1 {
		return nil
	}

	runningBalance := utils.Truncate2(parent.Balance)
	for _, tier := range inviteRebateTiers {
		if int64(tier.Threshold) > validUsers {
			continue
		}

		// 幂等：该上级该档已发过则跳过（唯一索引 user_id+tier_level）
		logRecord := pojo.TgUserInviteRewardLog{
			TenantID:   parent.TenantId,
			UserID:     parentID,
			TierLevel:  tier.Level,
			Threshold:  tier.Threshold,
			ValidUsers: int(validUsers),
			Amount:     tier.Amount,
		}
		created := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&logRecord)
		if created.Error != nil {
			return created.Error
		}
		if created.RowsAffected == 0 {
			continue
		}

		// 仅入账余额；提现限制走 v2 流水批次
		if err := tx.Model(&pojo.TgUser{}).Where("id = ?", parentID).Updates(map[string]any{
			"balance": gorm.Expr("balance + ?", tier.Amount),
		}).Error; err != nil {
			return err
		}

		history := pojo.CashHistory{
			UserId:          parentID,
			AwardUni:        fmt.Sprintf("invite_rebate_tier_%d_%d", parentID, tier.Level),
			Amount:          tier.Amount,
			StartAmount:     runningBalance,
			EndAmount:       utils.Truncate2(runningBalance + tier.Amount),
			CashMark:        "邀请返佣奖励",
			CashDesc:        fmt.Sprintf("有效用户满%d人奖励%.2f", tier.Threshold, tier.Amount),
			Type:            pojo.CashHistoryTypeInviteRebateTierReward,
			IsGift:          1,
			FromUserId:      0,
			SourceChannelID: parent.SourceChannelID,
		}
		if err := tx.Create(&history).Error; err != nil {
			return err
		}
		runningBalance = utils.Truncate2(runningBalance + tier.Amount)

		if err := EnsureWithdrawFlowBatchForGift(
			tx, parent,
			pojo.WithdrawFlowBatchSourceInviteRebate,
			logRecord.ID,
			fmt.Sprintf("invite_rebate_%d", logRecord.ID),
			pojo.WithdrawFlowBatchSourceInviteRebate,
			tier.Amount,
		); err != nil {
			return err
		}
	}
	return nil
}

// BackfillInviteRebateTiers 对历史存量有效用户补发邀请返佣阶梯奖励（一次性脚本，幂等可重复执行）。
// 传入带表前缀的 db（如 utils.NewPrefixDb(prefix)）。逐个上级单独事务处理，返回处理的上级数量。
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
			return GrantInviteRebateTiers(tx, pid)
		}); err != nil {
			return processed, fmt.Errorf("backfill invite rebate parent=%d: %w", pid, err)
		}
		processed++
	}
	return processed, nil
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
