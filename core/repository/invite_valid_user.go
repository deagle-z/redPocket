package repository

import (
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/utils"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	InviteValidMinRecharge float64 = 0
	InviteValidMinBet      float64 = 850
	InviteBetRebateRate    float64 = 0.1
)

// BetRebateDisabledConfigKey 下注返佣开关：0=正常返佣 1=关闭(所有下注返佣都不返)
const BetRebateDisabledConfigKey = "bet_rebate_disabled"

// isBetRebateDisabled 读取下注返佣开关；记录缺失时自动初始化为 0（默认正常返佣）。
func isBetRebateDisabled(db *gorm.DB) bool {
	if db == nil {
		return false
	}
	var cfg pojo.SysConfig
	err := db.Where("config_key = ?", BetRebateDisabledConfigKey).First(&cfg).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			_ = db.Create(&pojo.SysConfig{
				ConfigKey:   BetRebateDisabledConfigKey,
				ConfigValue: "0",
				ConfigDesc:  "下注返佣开关 0=正常返佣 1=关闭(所有下注返佣都不返)",
			}).Error
		}
		return false
	}
	return strings.TrimSpace(cfg.ConfigValue) == "1"
}

type inviteRewardTier struct {
	RechargeAmount float64
	BetAmount      float64
}

// inviteRewardTierForCount 按对应阶段人数所在档位返回充值阶段/投注达标阶段奖励金额。
// 档位：1-5→5+10，6-15→8+12，16-30→10+15，31-100→12+18，100以上→15+20。
func inviteRewardTierForCount(users int64) inviteRewardTier {
	switch {
	case users <= 0:
		return inviteRewardTier{}
	case users <= 5:
		return inviteRewardTier{RechargeAmount: 5, BetAmount: 10}
	case users <= 15:
		return inviteRewardTier{RechargeAmount: 8, BetAmount: 12}
	case users <= 30:
		return inviteRewardTier{RechargeAmount: 10, BetAmount: 15}
	case users <= 100:
		return inviteRewardTier{RechargeAmount: 12, BetAmount: 18}
	default:
		return inviteRewardTier{RechargeAmount: 15, BetAmount: 20}
	}
}

func hasInviteQualifyingRecharge(totalRecharge float64) bool {
	totalRecharge = utils.Truncate2(totalRecharge)
	if InviteValidMinRecharge <= 0 {
		return totalRecharge > 0
	}
	return totalRecharge >= InviteValidMinRecharge
}

// EnsureInviteRechargeReward 在直属下级首次真实充值后，按充值用户阶段档位给上级发放第一段邀请奖励。
func EnsureInviteRechargeReward(tx *gorm.DB, userID int64, qualifiedAt time.Time) error {
	if tx == nil || userID <= 0 {
		return nil
	}

	var user pojo.TgUser
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", userID).First(&user).Error; err != nil {
		return err
	}
	if user.ID == 0 || user.Status == -1 || user.ParentID == nil || *user.ParentID <= 0 {
		return nil
	}

	totalRecharge, err := GetInviteUserRechargeAmount(tx, userID)
	if err != nil {
		return err
	}
	if !hasInviteQualifyingRecharge(totalRecharge) {
		return nil
	}

	return grantInviteRechargeRewardForUser(tx, user, totalRecharge, qualifiedAt)
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

	totalRecharge, err := GetInviteUserRechargeAmount(tx, userID)
	if err != nil {
		return false, err
	}
	totalBet, err := GetInviteValidUserBetAmount(tx, userID)
	if err != nil {
		return false, err
	}
	if !hasInviteQualifyingRecharge(totalRecharge) || totalBet < InviteValidMinBet {
		return false, nil
	}
	if qualifiedAt.IsZero() {
		qualifiedAt = time.Now()
	}
	if err := grantInviteRechargeRewardForUser(tx, user, totalRecharge, qualifiedAt); err != nil {
		return false, err
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
	if res.RowsAffected > 0 {
		// 第二笔（投注达标）档位锁定为该下级充值时的档位
		position, err := inviteRechargeStagePosition(tx, *user.ParentID, userID)
		if err != nil {
			return false, err
		}
		if position <= 0 {
			// 理论上充值段已先发放；兜底用当前充值用户数
			position, err = countInviteRechargedUsers(tx, *user.ParentID)
			if err != nil {
				return false, err
			}
		}
		tier := inviteRewardTierForCount(position)
		if err := grantInviteRewardStage(tx, *user.ParentID, userID, pojo.InviteRewardStageBet, position, tier.BetAmount); err != nil {
			return false, err
		}
	}
	return true, nil
}

// inviteRechargeStagePosition 读取该下级充值段(stage=recharge)记录里的档位位次（无记录返回 0）。
func inviteRechargeStagePosition(tx *gorm.DB, parentID int64, subUserID int64) (int64, error) {
	if tx == nil || parentID <= 0 || subUserID <= 0 {
		return 0, nil
	}
	var rec pojo.TgUserInviteRewardLog
	if err := tx.Where("user_id = ? AND sub_user_id = ? AND stage = ?", parentID, subUserID, pojo.InviteRewardStageRecharge).
		First(&rec).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, nil
		}
		return 0, err
	}
	return int64(rec.ValidUsers), nil
}

func grantInviteRechargeRewardForUser(tx *gorm.DB, user pojo.TgUser, totalRecharge float64, qualifiedAt time.Time) error {
	if tx == nil || user.ID <= 0 || user.ParentID == nil || *user.ParentID <= 0 {
		return nil
	}
	if !hasInviteQualifyingRecharge(totalRecharge) {
		return nil
	}
	rechargeUsers, err := countInviteRechargedUsers(tx, *user.ParentID)
	if err != nil {
		return err
	}
	tier := inviteRewardTierForCount(rechargeUsers)
	return grantInviteRewardStage(tx, *user.ParentID, user.ID, pojo.InviteRewardStageRecharge, rechargeUsers, tier.RechargeAmount)
}

func countInviteValidUsers(tx *gorm.DB, parentID int64) (int64, error) {
	if tx == nil || parentID <= 0 {
		return 0, nil
	}
	var validUsers int64
	err := tx.Model(&pojo.TgUser{}).
		Where("parent_id = ? AND status <> ? AND invite_valid_flag = ?", parentID, -1, 1).
		Count(&validUsers).Error
	return validUsers, err
}

func countInviteRechargedUsers(tx *gorm.DB, parentID int64) (int64, error) {
	if tx == nil || parentID <= 0 {
		return 0, nil
	}
	var rechargeUsers int64
	err := tx.Table("recharge_order ro").
		Joins("inner join tg_user tu on tu.id = ro.user_id").
		Where("tu.parent_id = ? AND tu.status <> ? AND ro.status = ? AND COALESCE(ro.is_dev, 0) = 0", parentID, -1, 1).
		Select("COUNT(DISTINCT ro.user_id)").
		Scan(&rechargeUsers).Error
	return rechargeUsers, err
}

func hasLegacyInviteReward(tx *gorm.DB, parentID int64, subUserID int64) (bool, error) {
	if tx == nil || parentID <= 0 || subUserID <= 0 {
		return false, nil
	}
	var count int64
	err := tx.Model(&pojo.TgUserInviteRewardLog{}).
		Where("user_id = ? AND sub_user_id = ? AND (stage IS NULL OR stage = '' OR stage = ?)", parentID, subUserID, pojo.InviteRewardStageLegacy).
		Count(&count).Error
	return count > 0, err
}

func grantInviteRewardStage(tx *gorm.DB, parentID int64, subUserID int64, stage string, stageUsers int64, amount float64) error {
	if tx == nil || parentID <= 0 || subUserID <= 0 {
		return nil
	}
	stage = strings.TrimSpace(stage)
	amount = utils.Truncate2(amount)
	if stage == "" || amount <= 0 {
		return nil
	}
	// 下注返佣开关：开启(=1)时投注达标段的邀请奖励也一并停发（充值段不受影响）。
	if stage == pojo.InviteRewardStageBet && isBetRebateDisabled(tx) {
		return nil
	}
	legacyExists, err := hasLegacyInviteReward(tx, parentID, subUserID)
	if err != nil || legacyExists {
		return err
	}

	var parent pojo.TgUser
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", parentID).First(&parent).Error; err != nil || parent.ID == 0 {
		return nil
	}
	if parent.Status != 1 {
		return nil
	}

	logRecord := pojo.TgUserInviteRewardLog{
		TenantID:   parent.TenantId,
		UserID:     parentID,
		SubUserID:  subUserID,
		Stage:      stage,
		ValidUsers: int(stageUsers),
		Amount:     amount,
	}
	created := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&logRecord)
	if created.Error != nil {
		return created.Error
	}
	if created.RowsAffected == 0 {
		return nil
	}

	// 充值返水(stage=recharge)且上级 rebate_type=2 时：返到可用余额(balance)并挂 v2 提现流水批次限制。
	// 其余情况（含投注达标段、rebate_type=1）：计入上级「返水余额」(rebate_amount/rebate_total_amount)，走佣金转余额提现。
	if stage == pojo.InviteRewardStageRecharge && parent.RebateType == 2 {
		if err := tx.Model(&pojo.TgUser{}).Where("id = ?", parentID).Updates(map[string]any{
			"balance": gorm.Expr("balance + ?", amount),
		}).Error; err != nil {
			return err
		}

		history := pojo.CashHistory{
			UserId:          parentID,
			AwardUni:        fmt.Sprintf("invite_rebate_%s_%d_%d", stage, parentID, subUserID),
			Amount:          amount,
			StartAmount:     utils.Truncate2(parent.Balance),
			EndAmount:       utils.Truncate2(parent.Balance + amount),
			CashMark:        "邀请充值返水",
			CashDesc:        fmt.Sprintf("邀请充值奖励%.2f（当前充值用户%d人）", amount, stageUsers),
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
			fmt.Sprintf("invite_rebate_%s_%d", stage, logRecord.ID),
			pojo.WithdrawFlowBatchSourceInviteRebate,
			amount,
		)
	}

	now := time.Now()
	remark := fmt.Sprintf("invite_tier_%s_reward", stage)
	rebateRecord := pojo.TgUserRebateRecord{
		TenantId:        &parent.TenantId,
		SubUserId:       subUserID,
		ParentUserId:    parentID,
		SourceChannelID: parent.SourceChannelID,
		SourceType:      pojo.TgUserRebateSourceTypeInviteTier,
		SourceOrderId:   fmt.Sprintf("invite_tier_%s_%d_%d", stage, parentID, subUserID),
		SourceAmount:    0,
		RebateRate:      0,
		RebateAmount:    amount,
		Currency:        "USDT",
		Status:          1,
		SettledAt:       &now,
		IdempotencyKey:  fmt.Sprintf("invite_tier:%s:%d:%d", stage, parentID, subUserID),
		Remark:          &remark,
	}
	if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&rebateRecord).Error; err != nil {
		return err
	}

	return tx.Model(&pojo.TgUser{}).Where("id = ?", parentID).Updates(map[string]any{
		"rebate_amount":       gorm.Expr("rebate_amount + ?", amount),
		"rebate_total_amount": gorm.Expr("rebate_total_amount + ?", amount),
	}).Error
}

// BackfillInviteRebateTiers 对历史存量有效用户补发邀请返佣（一次性脚本，幂等可重复执行）。
// 传入带表前缀的 db（如 utils.NewPrefixDb(prefix)）。逐个上级单独事务处理，返回处理的上级数量。
// 兼容旧规则：已有 legacy 发放记录的下级不会重复发放拆分后的两段奖励。
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

func backfillInviteRebateForParent(tx *gorm.DB, parentID int64) error {
	var subUserIDs []int64
	if err := tx.Model(&pojo.TgUser{}).
		Where("parent_id = ? AND status <> ? AND invite_valid_flag = ?", parentID, -1, 1).
		Order("invite_valid_at asc, id asc").
		Pluck("id", &subUserIDs).Error; err != nil {
		return err
	}
	for idx, subUserID := range subUserIDs {
		stageUsers := int64(idx + 1)
		tier := inviteRewardTierForCount(stageUsers)
		if err := grantInviteRewardStage(tx, parentID, subUserID, pojo.InviteRewardStageRecharge, stageUsers, tier.RechargeAmount); err != nil {
			return err
		}
		if err := grantInviteRewardStage(tx, parentID, subUserID, pojo.InviteRewardStageBet, stageUsers, tier.BetAmount); err != nil {
			return err
		}
	}
	return nil
}

func GetInviteUserRechargeAmount(db *gorm.DB, userID int64) (float64, error) {
	if db == nil || userID <= 0 {
		return 0, nil
	}
	var totalRecharge float64
	err := db.Model(&pojo.RechargeOrder{}).
		Where("user_id = ? AND status = ? AND COALESCE(is_dev, 0) = 0", userID, 1).
		Select("COALESCE(SUM(COALESCE(amount, 0)), 0)").
		Scan(&totalRecharge).Error
	if err != nil {
		return 0, err
	}
	return utils.Truncate2(totalRecharge), nil
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

	// 下注返佣开关：开启(=1)时所有下注返佣都不返（不影响邀请有效用户/阶梯奖励）。
	if isBetRebateDisabled(tx) {
		return nil
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
