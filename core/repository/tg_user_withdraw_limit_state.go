package repository

import (
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/utils"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"strconv"
	"strings"
	"time"
)

const withdrawLimitEpsilon = 0.000001

type UserBalanceSourceSplit struct {
	GiftRestrictedAmount     float64
	RechargeRestrictedAmount float64
	UnrestrictedAmount       float64
}

func GetUserWithdrawSummary(db *gorm.DB, userID int64) (pojo.TgWithdrawSummaryBack, error) {
	var result pojo.TgWithdrawSummaryBack
	if db == nil || userID <= 0 {
		return result, errors.New("user_not_found")
	}

	err := db.Transaction(func(tx *gorm.DB) error {
		var user pojo.TgUser
		if err := tx.Where("id = ?", userID).First(&user).Error; err != nil {
			return err
		}
		if user.ID == 0 {
			return errors.New("user_not_found")
		}

		state, err := getOrInitUserWithdrawLimitState(tx, user)
		if err != nil {
			return err
		}

		totalFlow, err := GetUserTotalFlow(tx, user.ID)
		if err != nil {
			return err
		}

		giftMultiplier := loadWithdrawLimitMultiplier(tx, "withdraw_gift_limit")
		rechargeMultiplier := loadWithdrawLimitMultiplier(tx, "withdraw_limit")
		availableFlow := clampNonNegative(totalFlow - state.GiftFlowConsumed - state.RechargeFlowConsumed)
		giftRestricted := clampNonNegative(state.GiftRestrictedBalance)
		rechargeRestricted := clampNonNegative(state.RechargeRestrictedBalance)
		unrestricted := clampNonNegative(user.Balance - giftRestricted - rechargeRestricted)

		withdrawableGift := giftRestricted
		if giftMultiplier > 0 {
			withdrawableGift = minFloat(giftRestricted, availableFlow/giftMultiplier)
		}
		remainingFlow := clampNonNegative(availableFlow - utils.Truncate2(withdrawableGift*giftMultiplier))

		withdrawableRecharge := rechargeRestricted
		if rechargeMultiplier > 0 {
			withdrawableRecharge = minFloat(rechargeRestricted, remainingFlow/rechargeMultiplier)
		}

		totalWithdrawable := clampNonNegative(unrestricted + withdrawableGift + withdrawableRecharge)
		nonWithdrawable := clampNonNegative(user.Balance - totalWithdrawable)

		result = pojo.TgWithdrawSummaryBack{
			Balance:                    utils.Truncate2(user.Balance),
			WithdrawableAmount:         totalWithdrawable,
			NonWithdrawableAmount:      nonWithdrawable,
			UnrestrictedAmount:         unrestricted,
			GiftRestrictedAmount:       giftRestricted,
			RechargeRestrictedAmount:   rechargeRestricted,
			WithdrawableGiftAmount:     withdrawableGift,
			WithdrawableRechargeAmount: withdrawableRecharge,
			AvailableFlow:              availableFlow,
			GiftFlowConsumed:           state.GiftFlowConsumed,
			RechargeFlowConsumed:       state.RechargeFlowConsumed,
			GiftLimitMultiplier:        giftMultiplier,
			RechargeLimitMultiplier:    rechargeMultiplier,
		}
		return nil
	})

	return result, err
}

func EnsureUserWithdrawLimitState(tx *gorm.DB, user pojo.TgUser) error {
	if tx == nil || user.ID <= 0 {
		return nil
	}
	_, err := getOrInitUserWithdrawLimitState(tx, user)
	return err
}

func AddUserWithdrawRestrictedBalance(tx *gorm.DB, user pojo.TgUser, giftAmount float64, rechargeAmount float64) error {
	if tx == nil || user.ID <= 0 {
		return nil
	}
	if giftAmount <= 0 && rechargeAmount <= 0 {
		return nil
	}
	state, err := getOrInitUserWithdrawLimitState(tx, user)
	if err != nil {
		return err
	}
	state.GiftRestrictedBalance = clampNonNegative(state.GiftRestrictedBalance + giftAmount)
	state.RechargeRestrictedBalance = clampNonNegative(state.RechargeRestrictedBalance + rechargeAmount)
	return saveUserWithdrawLimitState(tx, &state)
}

func RestoreUserWithdrawRestrictedBalance(tx *gorm.DB, user pojo.TgUser, giftAmount float64, rechargeAmount float64) error {
	return AddUserWithdrawRestrictedBalance(tx, user, giftAmount, rechargeAmount)
}

func ConsumeUserBalanceSources(tx *gorm.DB, user pojo.TgUser, giftDebit float64, regularDebit float64) (UserBalanceSourceSplit, error) {
	var result UserBalanceSourceSplit
	if tx == nil || user.ID <= 0 {
		return result, nil
	}
	if giftDebit <= 0 && regularDebit <= 0 {
		return result, nil
	}

	state, err := getOrInitUserWithdrawLimitState(tx, user)
	if err != nil {
		return result, err
	}

	origGift := clampNonNegative(state.GiftRestrictedBalance)
	origRecharge := clampNonNegative(state.RechargeRestrictedBalance)
	unrestrictedAvailable := clampNonNegative(user.Balance - origGift - origRecharge)

	result.GiftRestrictedAmount = minFloat(clampNonNegative(giftDebit), origGift)
	result.UnrestrictedAmount = minFloat(clampNonNegative(regularDebit), unrestrictedAvailable)

	remainingRegular := clampNonNegative(regularDebit - result.UnrestrictedAmount)
	result.RechargeRestrictedAmount = minFloat(remainingRegular, origRecharge)

	state.GiftRestrictedBalance = clampNonNegative(origGift - result.GiftRestrictedAmount)
	state.RechargeRestrictedBalance = clampNonNegative(origRecharge - result.RechargeRestrictedAmount)

	return result, saveUserWithdrawLimitState(tx, &state)
}

func ReserveWithdrawLimitForOrder(tx *gorm.DB, user pojo.TgUser, order *pojo.WithdrawOrderBr) error {
	if tx == nil || order == nil || user.ID <= 0 || order.Amount <= 0 {
		return nil
	}

	state, err := getOrInitUserWithdrawLimitState(tx, user)
	if err != nil {
		return err
	}

	giftMultiplier := loadWithdrawLimitMultiplier(tx, "withdraw_gift_limit")
	rechargeMultiplier := loadWithdrawLimitMultiplier(tx, "withdraw_limit")

	giftRestrictedAmount := minFloat(order.Amount, clampNonNegative(state.GiftRestrictedBalance))
	remainingAmount := clampNonNegative(order.Amount - giftRestrictedAmount)
	rechargeRestrictedAmount := minFloat(remainingAmount, clampNonNegative(state.RechargeRestrictedBalance))
	unrestrictedAmount := clampNonNegative(order.Amount - giftRestrictedAmount - rechargeRestrictedAmount)

	giftFlowRequired := utils.Truncate2(giftRestrictedAmount * giftMultiplier)
	rechargeFlowRequired := utils.Truncate2(rechargeRestrictedAmount * rechargeMultiplier)

	totalFlow, err := GetUserTotalFlow(tx, user.ID)
	if err != nil {
		return err
	}
	availableFlow := clampNonNegative(totalFlow - state.GiftFlowConsumed - state.RechargeFlowConsumed)
	requiredFlow := utils.Truncate2(giftFlowRequired + rechargeFlowRequired)
	if availableFlow+withdrawLimitEpsilon < requiredFlow {
		return errors.New(utils.I18nMessage("withdraw_flow_insufficient", map[string]interface{}{
			"available": fmt.Sprintf("%.2f", availableFlow),
			"required":  fmt.Sprintf("%.2f", requiredFlow),
		}))
	}

	state.GiftRestrictedBalance = clampNonNegative(state.GiftRestrictedBalance - giftRestrictedAmount)
	state.RechargeRestrictedBalance = clampNonNegative(state.RechargeRestrictedBalance - rechargeRestrictedAmount)
	state.GiftFlowConsumed = clampNonNegative(state.GiftFlowConsumed + giftFlowRequired)
	state.RechargeFlowConsumed = clampNonNegative(state.RechargeFlowConsumed + rechargeFlowRequired)
	if err := saveUserWithdrawLimitState(tx, &state); err != nil {
		return err
	}

	order.GiftRestrictedAmount = giftRestrictedAmount
	order.RechargeRestrictedAmount = rechargeRestrictedAmount
	order.UnrestrictedAmount = unrestrictedAmount
	order.GiftFlowRequired = giftFlowRequired
	order.RechargeFlowRequired = rechargeFlowRequired

	return tx.Model(&pojo.WithdrawOrderBr{}).Where("id = ?", order.ID).Updates(map[string]any{
		"gift_restricted_amount":     giftRestrictedAmount,
		"recharge_restricted_amount": rechargeRestrictedAmount,
		"unrestricted_amount":        unrestrictedAmount,
		"gift_flow_required":         giftFlowRequired,
		"recharge_flow_required":     rechargeFlowRequired,
	}).Error
}

func RefundWithdrawLimitForOrder(tx *gorm.DB, user pojo.TgUser, order pojo.WithdrawOrderBr) error {
	if tx == nil || user.ID <= 0 {
		return nil
	}
	if order.GiftRestrictedAmount <= 0 &&
		order.RechargeRestrictedAmount <= 0 &&
		order.GiftFlowRequired <= 0 &&
		order.RechargeFlowRequired <= 0 {
		return nil
	}

	state, err := getOrInitUserWithdrawLimitState(tx, user)
	if err != nil {
		return err
	}

	state.GiftRestrictedBalance = clampNonNegative(state.GiftRestrictedBalance + order.GiftRestrictedAmount)
	state.RechargeRestrictedBalance = clampNonNegative(state.RechargeRestrictedBalance + order.RechargeRestrictedAmount)
	state.GiftFlowConsumed = clampNonNegative(state.GiftFlowConsumed - order.GiftFlowRequired)
	state.RechargeFlowConsumed = clampNonNegative(state.RechargeFlowConsumed - order.RechargeFlowRequired)

	return saveUserWithdrawLimitState(tx, &state)
}

func RestoreLuckyRefundRestrictedBalance(tx *gorm.DB, user pojo.TgUser, lucky pojo.LuckyMoney, refundAmount float64) error {
	if tx == nil || user.ID <= 0 || refundAmount <= 0 || lucky.Amount <= 0 {
		return nil
	}

	ratio := utils.Truncate2(refundAmount / lucky.Amount)
	if ratio > 1 {
		ratio = 1
	}
	giftAmount := utils.Truncate2(lucky.GiftRestrictedAmount * ratio)
	rechargeAmount := utils.Truncate2(lucky.RechargeRestrictedAmount * ratio)
	return RestoreUserWithdrawRestrictedBalance(tx, user, giftAmount, rechargeAmount)
}

func getOrInitUserWithdrawLimitState(tx *gorm.DB, user pojo.TgUser) (pojo.TgUserWithdrawLimitState, error) {
	var state pojo.TgUserWithdrawLimitState
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("user_id = ?", user.ID).First(&state).Error
	if err == nil && state.ID > 0 {
		return state, nil
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return state, err
	}

	now := time.Now()
	giftRestricted := clampNonNegative(user.GiftAmount)
	if giftRestricted > user.Balance {
		giftRestricted = clampNonNegative(user.Balance)
	}
	rechargeRestricted := clampNonNegative(user.Balance - giftRestricted)
	newState := pojo.TgUserWithdrawLimitState{
		UserID:                    user.ID,
		GiftRestrictedBalance:     giftRestricted,
		RechargeRestrictedBalance: rechargeRestricted,
		GiftFlowConsumed:          0,
		RechargeFlowConsumed:      0,
		InitializedAt:             &now,
	}
	if createErr := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&newState).Error; createErr != nil {
		return state, createErr
	}

	err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("user_id = ?", user.ID).First(&state).Error
	return state, err
}

func saveUserWithdrawLimitState(tx *gorm.DB, state *pojo.TgUserWithdrawLimitState) error {
	if tx == nil || state == nil || state.ID <= 0 {
		return nil
	}
	return tx.Model(&pojo.TgUserWithdrawLimitState{}).Where("id = ?", state.ID).Updates(map[string]any{
		"gift_restricted_balance":     state.GiftRestrictedBalance,
		"recharge_restricted_balance": state.RechargeRestrictedBalance,
		"gift_flow_consumed":          state.GiftFlowConsumed,
		"recharge_flow_consumed":      state.RechargeFlowConsumed,
	}).Error
}

func loadWithdrawLimitMultiplier(tx *gorm.DB, key string) float64 {
	defaultValue := 1.0
	if tx == nil || strings.TrimSpace(key) == "" {
		return defaultValue
	}
	var cfg pojo.SysConfig
	if err := tx.Where("config_key = ?", key).First(&cfg).Error; err != nil || cfg.ID == 0 {
		return defaultValue
	}
	value, err := strconv.ParseFloat(strings.TrimSpace(cfg.ConfigValue), 64)
	if err != nil || value <= 0 {
		return defaultValue
	}
	return value
}

func clampNonNegative(value float64) float64 {
	if value < withdrawLimitEpsilon {
		return 0
	}
	return utils.Truncate2(value)
}

func minFloat(a float64, b float64) float64 {
	if a < b {
		return utils.Truncate2(a)
	}
	return utils.Truncate2(b)
}
