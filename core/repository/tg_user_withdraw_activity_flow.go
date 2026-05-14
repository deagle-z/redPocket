package repository

import (
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/utils"
	"errors"
	"gorm.io/gorm"
)

func GetUserWithdrawActivityFlow(db *gorm.DB, userID int64) (pojo.TgWithdrawActivityFlowBack, error) {
	var result pojo.TgWithdrawActivityFlowBack
	if db == nil || userID <= 0 {
		return result, errors.New("user_not_found")
	}

	var user pojo.TgUser
	if err := db.Select("id, balance").Where("id = ?", userID).First(&user).Error; err != nil {
		return result, err
	}
	if user.ID == 0 {
		return result, errors.New("user_not_found")
	}

	totalFlow, err := GetUserTotalFlow(db, user.ID)
	if err != nil {
		return result, err
	}

	var cycles []pojo.TgUserWithdrawActivityCycle
	if err := db.Where("user_id = ?", user.ID).
		Order("status ASC, id DESC").
		Limit(10).
		Find(&cycles).Error; err != nil {
		return result, err
	}

	result.UserID = user.ID
	result.Balance = utils.Truncate2(user.Balance)
	result.TotalFlow = utils.Truncate2(totalFlow)
	result.Activities = make([]pojo.TgWithdrawActivityFlowCycleBack, 0, len(cycles))
	for _, cycle := range cycles {
		item := buildWithdrawActivityFlowCycle(cycle, totalFlow)
		result.Activities = append(result.Activities, item)
		if result.ActiveActivity == nil && cycle.Status == pojo.WithdrawActivityCycleStatusActive {
			active := item
			result.ActiveActivity = &active
			result.HasActivity = true
		}
	}

	return result, nil
}

func buildWithdrawActivityFlowCycle(cycle pojo.TgUserWithdrawActivityCycle, totalFlow float64) pojo.TgWithdrawActivityFlowCycleBack {
	currentFlow := clampNonNegative(totalFlow - cycle.FlowStartValue)
	availableFlow := clampNonNegative(currentFlow - cycle.FlowConsumed)
	remainingFlow := clampNonNegative(cycle.RequiredFlow - availableFlow)
	progressPercent := 0.0
	if cycle.RequiredFlow > 0 {
		progressPercent = utils.Truncate2(minFloat(100, availableFlow/cycle.RequiredFlow*100))
	}

	return pojo.TgWithdrawActivityFlowCycleBack{
		ID:               cycle.ID,
		CreatedAt:        cycle.CreatedAt,
		UpdatedAt:        cycle.UpdatedAt,
		ActivityCode:     cycle.ActivityCode,
		ActivityType:     cycle.ActivityType,
		Status:           cycle.Status,
		Multiplier:       utils.Truncate2(cycle.Multiplier),
		BaseAmount:       utils.Truncate2(cycle.BaseAmount),
		RequiredFlow:     utils.Truncate2(cycle.RequiredFlow),
		FlowStartValue:   utils.Truncate2(cycle.FlowStartValue),
		FlowConsumed:     utils.Truncate2(cycle.FlowConsumed),
		CurrentFlow:      currentFlow,
		AvailableFlow:    availableFlow,
		RemainingFlow:    remainingFlow,
		ProgressPercent:  progressPercent,
		BalanceThreshold: utils.Truncate2(cycle.BalanceThreshold),
		LastRechargeNo:   cycle.LastRechargeNo,
		EndReason:        cycle.EndReason,
		StartedAt:        cycle.StartedAt,
		EndedAt:          cycle.EndedAt,
	}
}
