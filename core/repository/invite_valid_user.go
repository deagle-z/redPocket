package repository

import (
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/utils"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	// InviteValidMinRecharge 有效用户门槛：累计充值 > 此金额即算有效用户（无投注门槛）
	InviteValidMinRecharge float64 = 50
	InviteValidMinBet      float64 = 850
	InviteBetRebateRate    float64 = 0.1
)

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
	// 累计充值「大于」门槛即算达标（如 >50）
	return totalRecharge > InviteValidMinRecharge
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
	// 有效用户门槛：累计充值 > InviteValidMinRecharge（默认>50）即可，不再要求投注流水。
	if !hasInviteQualifyingRecharge(totalRecharge) {
		return false, nil
	}
	if qualifiedAt.IsZero() {
		qualifiedAt = time.Now()
	}

	// 仅翻转「有效用户」标记用于统计；邀请返佣已改为按下级充值次数即时返佣，
	// 投注达标段奖励与下注流水返佣均已停用。totalBet 仅作快照记录。
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
