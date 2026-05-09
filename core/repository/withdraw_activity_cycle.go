package repository

import (
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/utils"
	"errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"strconv"
	"strings"
	"time"
)

const (
	withdrawActivityLimitConfigKey            = "withdraw_activity_limit"
	withdrawActivityBalanceThresholdConfigKey = "withdraw_activity_balance_threshold"
)

func GetActiveWithdrawActivityCycle(db *gorm.DB, userID int64) (pojo.TgUserWithdrawActivityCycle, error) {
	var cycle pojo.TgUserWithdrawActivityCycle
	if db == nil || userID <= 0 {
		return cycle, nil
	}
	err := db.Where("user_id = ? AND status = ?", userID, pojo.WithdrawActivityCycleStatusActive).
		Order("id DESC").
		First(&cycle).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return pojo.TgUserWithdrawActivityCycle{}, nil
	}
	return cycle, err
}

func getLockedActiveWithdrawActivityCycle(tx *gorm.DB, userID int64) (pojo.TgUserWithdrawActivityCycle, error) {
	var cycle pojo.TgUserWithdrawActivityCycle
	if tx == nil || userID <= 0 {
		return cycle, nil
	}
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("user_id = ? AND status = ?", userID, pojo.WithdrawActivityCycleStatusActive).
		Order("id DESC").
		First(&cycle).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return pojo.TgUserWithdrawActivityCycle{}, nil
	}
	return cycle, err
}

func EnsureWithdrawActivityCycleForRecharge(tx *gorm.DB, user pojo.TgUser, order pojo.RechargeOrder, multiplier float64, threshold float64, activityCode string) error {
	if tx == nil || user.ID <= 0 || order.ID <= 0 {
		return nil
	}
	multiplier = normalizePositiveConfig(multiplier, 5)
	threshold = normalizePositiveConfig(threshold, 10)
	baseAmount := clampNonNegative(user.Balance)
	now := time.Now()
	totalFlow, err := GetUserTotalFlow(tx, user.ID)
	if err != nil {
		return err
	}

	cycle, err := getLockedActiveWithdrawActivityCycle(tx, user.ID)
	if err != nil {
		return err
	}
	if cycle.ID > 0 {
		if multiplier < cycle.Multiplier {
			multiplier = cycle.Multiplier
		}
		updates := map[string]any{
			"tenant_id":         user.TenantId,
			"activity_code":     strings.TrimSpace(activityCode),
			"activity_type":     activityTypeValue(order.ActivityType),
			"multiplier":        multiplier,
			"base_amount":       baseAmount,
			"required_flow":     utils.Truncate2(baseAmount * multiplier),
			"balance_threshold": threshold,
			"last_recharge_no":  strings.TrimSpace(order.OrderNo),
		}
		return tx.Model(&pojo.TgUserWithdrawActivityCycle{}).Where("id = ?", cycle.ID).Updates(updates).Error
	}

	cycle = pojo.TgUserWithdrawActivityCycle{
		TenantId:         user.TenantId,
		UserID:           user.ID,
		Status:           pojo.WithdrawActivityCycleStatusActive,
		ActivityCode:     strings.TrimSpace(activityCode),
		ActivityType:     activityTypeValue(order.ActivityType),
		Multiplier:       multiplier,
		BaseAmount:       baseAmount,
		RequiredFlow:     utils.Truncate2(baseAmount * multiplier),
		FlowStartValue:   totalFlow,
		FlowConsumed:     0,
		BalanceThreshold: threshold,
		LastRechargeNo:   strings.TrimSpace(order.OrderNo),
		StartedAt:        &now,
	}
	return tx.Create(&cycle).Error
}

func RefreshActiveWithdrawActivityCycleForRecharge(tx *gorm.DB, user pojo.TgUser, order pojo.RechargeOrder) error {
	if tx == nil || user.ID <= 0 || order.ID <= 0 {
		return nil
	}
	cycle, err := getLockedActiveWithdrawActivityCycle(tx, user.ID)
	if err != nil || cycle.ID == 0 {
		return err
	}
	baseAmount := clampNonNegative(user.Balance)
	multiplier := normalizePositiveConfig(cycle.Multiplier, 5)
	threshold := normalizePositiveConfig(cycle.BalanceThreshold, 10)
	return tx.Model(&pojo.TgUserWithdrawActivityCycle{}).Where("id = ?", cycle.ID).Updates(map[string]any{
		"tenant_id":         user.TenantId,
		"base_amount":       baseAmount,
		"required_flow":     utils.Truncate2(baseAmount * multiplier),
		"balance_threshold": threshold,
		"last_recharge_no":  strings.TrimSpace(order.OrderNo),
	}).Error
}

func EndWithdrawActivityCycle(tx *gorm.DB, userID int64, reason string) error {
	if tx == nil || userID <= 0 {
		return nil
	}
	cycle, err := getLockedActiveWithdrawActivityCycle(tx, userID)
	if err != nil || cycle.ID == 0 {
		return err
	}
	now := time.Now()
	return tx.Model(&pojo.TgUserWithdrawActivityCycle{}).Where("id = ?", cycle.ID).Updates(map[string]any{
		"status":     pojo.WithdrawActivityCycleStatusEnded,
		"end_reason": strings.TrimSpace(reason),
		"ended_at":   &now,
	}).Error
}

func CanBypassWithdrawActivityCycleByBalance(userBalance float64, cycle pojo.TgUserWithdrawActivityCycle) bool {
	if cycle.ID == 0 || cycle.Status != pojo.WithdrawActivityCycleStatusActive {
		return false
	}
	threshold := normalizePositiveConfig(cycle.BalanceThreshold, 10)
	return clampNonNegative(userBalance) <= threshold
}

func GetWithdrawActivityCycleMultiplier(db *gorm.DB) float64 {
	return loadWithdrawActivityFloatConfig(db, withdrawActivityLimitConfigKey, 5, "活动充值提现所需流水倍数")
}

func GetWithdrawActivityBalanceThreshold(db *gorm.DB) float64 {
	return loadWithdrawActivityFloatConfig(db, withdrawActivityBalanceThresholdConfigKey, 10, "活动周期余额低于等于该值时结束流水要求")
}

func loadWithdrawActivityFloatConfig(db *gorm.DB, key string, defaultValue float64, desc string) float64 {
	if db == nil {
		return defaultValue
	}
	var cfg pojo.SysConfig
	if err := db.Where("config_key = ?", key).First(&cfg).Error; err != nil || cfg.ID == 0 {
		_ = db.Clauses(clause.OnConflict{DoNothing: true}).Create(&pojo.SysConfig{
			ConfigKey:   key,
			ConfigValue: strconv.FormatFloat(defaultValue, 'f', -1, 64),
			ConfigDesc:  desc,
		}).Error
		return defaultValue
	}
	value, err := strconv.ParseFloat(strings.TrimSpace(cfg.ConfigValue), 64)
	if err != nil || value <= 0 {
		return defaultValue
	}
	return value
}

func normalizePositiveConfig(value float64, defaultValue float64) float64 {
	if value <= 0 {
		return defaultValue
	}
	return value
}

func activityTypeValue(value *int8) int8 {
	if value == nil {
		return 0
	}
	return *value
}
