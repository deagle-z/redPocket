package repository

import (
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/utils"
	"fmt"
	"strconv"
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

// InviteRechargeRebateRatesConfigKey 邀请充值返佣比例配置键（sys_config，逗号分隔：第1次,第2次,第3次及以上）。
const InviteRechargeRebateRatesConfigKey = "invite_recharge_rebate_rates"

// defaultInviteRechargeRebateRates 默认充值返佣比例(%)：第1次40 第2次45 第3次及以上60。
var defaultInviteRechargeRebateRates = [3]float64{40, 45, 60}

// GetInviteRechargeRebateRates 读取充值返佣比例 [第1次, 第2次, 第3次及以上]；
// 配置缺失时自动初始化默认值，解析失败回退默认值。前后端统一从此配置取值。
func GetInviteRechargeRebateRates(db *gorm.DB) [3]float64 {
	rates := defaultInviteRechargeRebateRates
	if db == nil {
		return rates
	}
	var cfg pojo.SysConfig
	db.Where("config_key = ?", InviteRechargeRebateRatesConfigKey).First(&cfg)
	if cfg.ID == 0 {
		_ = db.Create(&pojo.SysConfig{
			ConfigKey:   InviteRechargeRebateRatesConfigKey,
			ConfigValue: fmt.Sprintf("%g,%g,%g", rates[0], rates[1], rates[2]),
			ConfigDesc:  "邀请充值返佣比例(%)，逗号分隔：第1次,第2次,第3次及以上",
		}).Error
		return rates
	}
	if parsed, ok := parseInviteRechargeRebateRates(cfg.ConfigValue); ok {
		return parsed
	}
	return rates
}

// parseInviteRechargeRebateRates 解析 "40,50,60" 形式的配置值。
func parseInviteRechargeRebateRates(raw string) ([3]float64, bool) {
	parts := strings.Split(strings.TrimSpace(raw), ",")
	if len(parts) < 3 {
		return [3]float64{}, false
	}
	var rates [3]float64
	for i := 0; i < 3; i++ {
		v, err := strconv.ParseFloat(strings.TrimSpace(parts[i]), 64)
		if err != nil || v < 0 {
			return [3]float64{}, false
		}
		rates[i] = v
	}
	return rates, true
}

func hasInviteQualifyingRecharge(totalRecharge float64) bool {
	totalRecharge = utils.Truncate2(totalRecharge)
	if InviteValidMinRecharge <= 0 {
		return totalRecharge > 0
	}
	return totalRecharge >= InviteValidMinRecharge
}

// inviteRechargeRebateRate 按"该下级自己的成功充值次数"返回充值返佣比例(%)。
// 档位比例从 sys_config（InviteRechargeRebateRatesConfigKey）读取：第1次 / 第2次 / 第3次及以上。
func inviteRechargeRebateRate(db *gorm.DB, rechargeCount int64) float64 {
	if rechargeCount <= 0 {
		return 0
	}
	rates := GetInviteRechargeRebateRates(db)
	switch {
	case rechargeCount == 1:
		return rates[0]
	case rechargeCount == 2:
		return rates[1]
	default:
		return rates[2]
	}
}

// countUserSuccessfulRecharges 统计该用户的成功充值次数（含手动回调 is_dev=1，用于确定返佣档位）。
func countUserSuccessfulRecharges(tx *gorm.DB, userID int64) (int64, error) {
	if tx == nil || userID <= 0 {
		return 0, nil
	}
	var cnt int64
	err := tx.Model(&pojo.RechargeOrder{}).
		Where("user_id = ? AND status = ?", userID, 1).
		Count(&cnt).Error
	return cnt, err
}

// ApplyInviteRechargeRebate 下级每次成功充值后，按"该下级自己的充值次数"档位
// （默认第1次40% / 第2次45% / 第3次及以上60%，比例可在 sys_config 配置）把该次充值额对应比例返给直属上级。
// 充值即返，无投注门槛；到账位置遵循上级 rebate_type。
func ApplyInviteRechargeRebate(tx *gorm.DB, order pojo.RechargeOrder, occurredAt time.Time) error {
	if tx == nil || order.UserId <= 0 {
		return nil
	}
	amount := utils.Truncate2(order.Amount)
	if amount <= 0 {
		return nil
	}

	var user pojo.TgUser
	if err := tx.Where("id = ?", order.UserId).First(&user).Error; err != nil {
		return err
	}
	if user.ID == 0 || user.Status == -1 || user.ParentID == nil || *user.ParentID <= 0 {
		return nil
	}

	// 该下级累计成功充值次数（本次在回调中已置 status=1，故已包含本次）
	rechargeCount, err := countUserSuccessfulRecharges(tx, user.ID)
	if err != nil {
		return err
	}
	rate := inviteRechargeRebateRate(tx, rechargeCount)
	if rate <= 0 {
		return nil
	}
	rebateAmount := utils.Truncate2(utils.ToMoney(amount).Multiply(rate / 100).ToDollars())
	if rebateAmount < 0.01 {
		return nil
	}
	if occurredAt.IsZero() {
		occurredAt = time.Now()
	}

	return grantInviteRechargeRebate(tx, *user.ParentID, user, order, rate, rebateAmount, rechargeCount, occurredAt)
}

// grantInviteRechargeRebate 把一笔充值返佣发放给上级，按订单幂等。
// rebate_type=2：进可用余额(balance)并挂 v2 提现流水批次；否则(=1)进返水余额并写返水记录。
func grantInviteRechargeRebate(tx *gorm.DB, parentID int64, subUser pojo.TgUser, order pojo.RechargeOrder, rate float64, rebateAmount float64, rechargeCount int64, occurredAt time.Time) error {
	if tx == nil || parentID <= 0 || rebateAmount <= 0 {
		return nil
	}

	var parent pojo.TgUser
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", parentID).First(&parent).Error; err != nil || parent.ID == 0 {
		return nil
	}
	if parent.Status != 1 {
		return nil
	}

	if parent.RebateType == 2 {
		awardUni := fmt.Sprintf("invite_recharge_rebate_%d_%s", parentID, order.OrderNo)
		var exist int64
		if err := tx.Model(&pojo.CashHistory{}).Where("award_uni = ?", awardUni).Count(&exist).Error; err != nil {
			return err
		}
		if exist > 0 {
			return nil
		}
		if err := tx.Model(&pojo.TgUser{}).Where("id = ?", parentID).Updates(map[string]any{
			"balance": gorm.Expr("balance + ?", rebateAmount),
		}).Error; err != nil {
			return err
		}
		history := pojo.CashHistory{
			UserId:          parentID,
			AwardUni:        awardUni,
			Amount:          rebateAmount,
			StartAmount:     utils.Truncate2(parent.Balance),
			EndAmount:       utils.Truncate2(parent.Balance + rebateAmount),
			CashMark:        "邀请充值返佣",
			CashDesc:        fmt.Sprintf("邀请充值返佣%.2f（下级第%d次充值，比例%.0f%%）", rebateAmount, rechargeCount, rate),
			Type:            pojo.CashHistoryTypeInviteRebateTierReward,
			IsGift:          1,
			FromUserId:      subUser.ID,
			SourceChannelID: parent.SourceChannelID,
		}
		if err := tx.Create(&history).Error; err != nil {
			return err
		}
		return EnsureWithdrawFlowBatchForGift(
			tx, parent,
			pojo.WithdrawFlowBatchSourceInviteRebate,
			history.ID,
			fmt.Sprintf("invite_recharge_rebate_%d", history.ID),
			pojo.WithdrawFlowBatchSourceInviteRebate,
			rebateAmount,
		)
	}

	tenantID := parent.TenantId
	remark := "invite_recharge_rebate"
	record := pojo.TgUserRebateRecord{
		TenantId:        &tenantID,
		SubUserId:       subUser.ID,
		ParentUserId:    parentID,
		SourceChannelID: parent.SourceChannelID,
		SourceType:      pojo.TgUserRebateSourceTypeRecharge,
		SourceOrderId:   order.OrderNo,
		SourceAmount:    utils.Truncate2(order.Amount),
		RebateRate:      rate,
		RebateAmount:    rebateAmount,
		Currency:        "USDT",
		Status:          1,
		SettledAt:       &occurredAt,
		IdempotencyKey:  fmt.Sprintf("invite_recharge_rebate:%s:%d", order.OrderNo, parentID),
		Remark:          &remark,
	}
	result := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&record)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return nil
	}
	return tx.Model(&pojo.TgUser{}).Where("id = ?", parentID).Updates(map[string]any{
		"rebate_amount":       gorm.Expr("rebate_amount + ?", rebateAmount),
		"rebate_total_amount": gorm.Expr("rebate_total_amount + ?", rebateAmount),
	}).Error
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

	// 仅翻转「有效用户」标记用于统计；邀请返佣已改为按下级充值次数即时返佣，
	// 投注达标段奖励与下注流水返佣均已停用。
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
	return true, nil
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

	// 下注流水返佣（0.1% 投注返水）已整体停用，仅保留「有效用户」标记翻转用于统计。
	if _, err := EnsureInviteValidUser(tx, subUser.ID, occurredAt); err != nil {
		return err
	}
	return nil
}
