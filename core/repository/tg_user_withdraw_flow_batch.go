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

func EnsureWithdrawFlowBatchForRechargeV2(tx *gorm.DB, user pojo.TgUser, order pojo.RechargeOrder, preRechargeBalance float64) error {
	if tx == nil || user.ID <= 0 || order.ID <= 0 || order.ActivityType == nil || *order.ActivityType != rechargeActivityTypeV2Gift {
		return nil
	}

	threshold := GetWithdrawActivityBalanceThreshold(tx)
	if clampNonNegative(preRechargeBalance) <= threshold {
		if err := CloseActiveWithdrawFlowBatches(tx, user.ID, pojo.WithdrawFlowBatchClosedReasonBalanceBelowThreshold); err != nil {
			return err
		}
	}

	var latest pojo.RechargeOrder
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", order.ID).First(&latest).Error; err != nil {
		return err
	}
	if latest.ID == 0 || latest.Status != 1 || latest.ActivityType == nil || *latest.ActivityType != rechargeActivityTypeV2Gift {
		return nil
	}

	creditAmount := 0.0
	if latest.CreditAmount != nil {
		creditAmount = utils.Truncate2(*latest.CreditAmount)
	}
	if creditAmount <= 0 {
		creditAmount = utils.Truncate2(latest.Amount - latest.Fee)
	}
	if creditAmount < 0 {
		creditAmount = 0
	}
	bonusAmount := utils.Truncate2(latest.BonusAmount)
	if bonusAmount < 0 {
		bonusAmount = 0
	}
	baseAmount := utils.Truncate2(creditAmount + bonusAmount)
	withdrawMultiplier := loadWithdrawFlowBatchMultiplier(tx, "withdraw_limit", 2, "v2充值到账金额提现所需流水倍数")
	giftMultiplier := loadWithdrawFlowBatchMultiplier(tx, "withdraw_gift_limit", 5, "v2充值赠送金额提现所需流水倍数")
	requiredFlow := utils.Truncate2(creditAmount*withdrawMultiplier + bonusAmount*giftMultiplier)
	if requiredFlow <= 0 {
		return nil
	}

	batch := pojo.TgUserWithdrawFlowBatch{
		TenantID:           latest.TenantId,
		UserID:             latest.UserId,
		SourceType:         pojo.WithdrawFlowBatchSourceRechargeV2,
		SourceOrderID:      latest.ID,
		SourceOrderNo:      strings.TrimSpace(latest.OrderNo),
		ActivityType:       rechargeActivityTypeV2Gift,
		ActivityCode:       RechargeActivityCodeV2Gift,
		CreditAmount:       creditAmount,
		BonusAmount:        bonusAmount,
		BaseAmount:         baseAmount,
		WithdrawMultiplier: withdrawMultiplier,
		GiftMultiplier:     giftMultiplier,
		RequiredFlow:       requiredFlow,
		CompletedFlow:      0,
		Status:             pojo.WithdrawFlowBatchStatusActive,
	}
	return tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&batch).Error
}

// EnsureWithdrawFlowBatchForGift 为纯赠送金额创建 v2 提现流水批次（所需流水 = 赠送额 × 赠送倍数）。
// 复用充值赠送的 withdraw_gift_limit 倍数配置。sourceOrderNo 需全局唯一以防重复建批次。
func EnsureWithdrawFlowBatchForGift(tx *gorm.DB, user pojo.TgUser, sourceType string, sourceID int64, sourceOrderNo string, activityCode string, giftAmount float64) error {
	if tx == nil || user.ID <= 0 {
		return nil
	}
	bonusAmount := utils.Truncate2(giftAmount)
	if bonusAmount <= 0 {
		return nil
	}
	giftMultiplier := loadWithdrawFlowBatchMultiplier(tx, "withdraw_gift_limit", 5, "v2充值赠送金额提现所需流水倍数")
	requiredFlow := utils.Truncate2(bonusAmount * giftMultiplier)
	if requiredFlow <= 0 {
		return nil
	}
	batch := pojo.TgUserWithdrawFlowBatch{
		TenantID:           user.TenantId,
		UserID:             user.ID,
		SourceType:         sourceType,
		SourceOrderID:      sourceID,
		SourceOrderNo:      strings.TrimSpace(sourceOrderNo),
		ActivityType:       0,
		ActivityCode:       activityCode,
		CreditAmount:       0,
		BonusAmount:        bonusAmount,
		BaseAmount:         bonusAmount,
		WithdrawMultiplier: 0,
		GiftMultiplier:     giftMultiplier,
		RequiredFlow:       requiredFlow,
		CompletedFlow:      0,
		Status:             pojo.WithdrawFlowBatchStatusActive,
	}
	return tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&batch).Error
}

func CloseActiveWithdrawFlowBatches(tx *gorm.DB, userID int64, reason string) error {
	if tx == nil || userID <= 0 {
		return nil
	}
	now := time.Now()
	return tx.Model(&pojo.TgUserWithdrawFlowBatch{}).
		Where("user_id = ? AND status = ?", userID, pojo.WithdrawFlowBatchStatusActive).
		Updates(map[string]any{
			"status":        pojo.WithdrawFlowBatchStatusClosed,
			"closed_reason": strings.TrimSpace(reason),
			"closed_at":     &now,
		}).Error
}

func RecordWithdrawFlowEvent(tx *gorm.DB, userID int64, tenantID int64, eventType string, eventKey string, sourceID int64, sourceOrderNo string, amount float64, occurredAt time.Time) error {
	if tx == nil || userID <= 0 || amount <= 0 {
		return nil
	}
	eventKey = strings.TrimSpace(eventKey)
	eventType = strings.TrimSpace(eventType)
	if eventKey == "" || eventType == "" {
		return nil
	}
	if occurredAt.IsZero() {
		occurredAt = time.Now()
	}

	outbox := pojo.TgUserWithdrawFlowOutbox{
		TenantID:      tenantID,
		UserID:        userID,
		EventType:     eventType,
		EventKey:      eventKey,
		SourceID:      sourceID,
		SourceOrderNo: strings.TrimSpace(sourceOrderNo),
		FlowAmount:    utils.Truncate2(amount),
		OccurredAt:    occurredAt,
		Status:        pojo.WithdrawFlowOutboxStatusPending,
	}
	if defaultPrefix := strings.TrimSpace(utils.CsConfig.DefaultHost.TablePrefix); defaultPrefix != "" {
		dbPrefix := strings.TrimSpace(utils.GetDbPrefix(tx))
		if dbPrefix != "" && dbPrefix != defaultPrefix {
			return allocateWithdrawFlowEvent(tx, outbox)
		}
	}
	return tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&outbox).Error
}

func ProcessWithdrawFlowOutboxBatch(db *gorm.DB, limit int) (int, error) {
	if db == nil {
		return 0, nil
	}
	if limit <= 0 || limit > 500 {
		limit = 100
	}

	now := time.Now()
	var items []pojo.TgUserWithdrawFlowOutbox
	if err := db.
		Where("status = ? AND (next_retry_at IS NULL OR next_retry_at <= ?)", pojo.WithdrawFlowOutboxStatusPending, now).
		Order("id ASC").
		Limit(limit).
		Find(&items).Error; err != nil {
		return 0, err
	}

	processed := 0
	for _, item := range items {
		if err := processWithdrawFlowOutboxItem(db, item.ID); err != nil {
			return processed, err
		}
		processed++
	}
	return processed, nil
}

func processWithdrawFlowOutboxItem(db *gorm.DB, outboxID int64) error {
	if outboxID <= 0 {
		return nil
	}
	return db.Transaction(func(tx *gorm.DB) error {
		var outbox pojo.TgUserWithdrawFlowOutbox
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ? AND status = ?", outboxID, pojo.WithdrawFlowOutboxStatusPending).
			First(&outbox).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil
			}
			return err
		}
		if outbox.NextRetryAt != nil && outbox.NextRetryAt.After(time.Now()) {
			return nil
		}

		if err := allocateWithdrawFlowEvent(tx, outbox); err != nil {
			attempts := outbox.Attempts + 1
			status := pojo.WithdrawFlowOutboxStatusPending
			nextRetryAt := time.Now().Add(withdrawFlowOutboxRetryDelay(attempts))
			if attempts >= withdrawFlowOutboxMaxAttempts {
				status = pojo.WithdrawFlowOutboxStatusFailed
			}
			return tx.Model(&pojo.TgUserWithdrawFlowOutbox{}).
				Where("id = ?", outbox.ID).
				Updates(map[string]any{
					"status":        status,
					"attempts":      attempts,
					"next_retry_at": &nextRetryAt,
					"last_error":    truncateWithdrawFlowOutboxError(err.Error()),
				}).Error
		}

		processedAt := time.Now()
		return tx.Model(&pojo.TgUserWithdrawFlowOutbox{}).
			Where("id = ?", outbox.ID).
			Updates(map[string]any{
				"status":        pojo.WithdrawFlowOutboxStatusDone,
				"processed_at":  &processedAt,
				"next_retry_at": nil,
				"last_error":    "",
			}).Error
	})
}

const withdrawFlowOutboxMaxAttempts = 10

func withdrawFlowOutboxRetryDelay(attempts int) time.Duration {
	if attempts < 1 {
		attempts = 1
	}
	delay := time.Duration(attempts*attempts) * time.Second
	if delay > 5*time.Minute {
		return 5 * time.Minute
	}
	return delay
}

func truncateWithdrawFlowOutboxError(message string) string {
	message = strings.TrimSpace(message)
	if len(message) <= 512 {
		return message
	}
	return message[:512]
}

func allocateWithdrawFlowEvent(tx *gorm.DB, outbox pojo.TgUserWithdrawFlowOutbox) error {
	if tx == nil || outbox.UserID <= 0 || outbox.FlowAmount <= 0 {
		return nil
	}
	var batches []pojo.TgUserWithdrawFlowBatch
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Select("id", "tenant_id", "user_id", "required_flow", "completed_flow").
		Where("user_id = ? AND status = ? AND completed_flow < required_flow", outbox.UserID, pojo.WithdrawFlowBatchStatusActive).
		Order("id ASC").
		Find(&batches).Error; err != nil {
		return err
	}
	if len(batches) == 0 {
		return nil
	}

	event := pojo.TgUserWithdrawFlowEvent{
		TenantID:      outbox.TenantID,
		UserID:        outbox.UserID,
		EventType:     outbox.EventType,
		EventKey:      outbox.EventKey,
		SourceID:      outbox.SourceID,
		SourceOrderNo: strings.TrimSpace(outbox.SourceOrderNo),
		FlowAmount:    utils.Truncate2(outbox.FlowAmount),
		OccurredAt:    outbox.OccurredAt,
	}
	insert := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&event)
	if insert.Error != nil {
		return insert.Error
	}
	if insert.RowsAffected == 0 {
		return nil
	}

	remainingFlow := event.FlowAmount
	allocatedTotal := 0.0
	for _, batch := range batches {
		if remainingFlow <= 0 {
			break
		}
		batchRemaining := clampNonNegative(batch.RequiredFlow - batch.CompletedFlow)
		if batchRemaining <= 0 {
			if err := completeWithdrawFlowBatch(tx, batch.ID, outbox.OccurredAt, batch.RequiredFlow); err != nil {
				return err
			}
			continue
		}
		allocated := minFloat(remainingFlow, batchRemaining)
		if allocated <= 0 {
			continue
		}
		if err := tx.Create(&pojo.TgUserWithdrawFlowAllocation{
			TenantID:        batch.TenantID,
			UserID:          outbox.UserID,
			EventID:         event.ID,
			BatchID:         batch.ID,
			AllocatedAmount: allocated,
		}).Error; err != nil {
			return err
		}

		newCompleted := utils.Truncate2(batch.CompletedFlow + allocated)
		updates := map[string]any{
			"completed_flow": newCompleted,
			"last_flow_at":   outbox.OccurredAt,
		}
		if newCompleted+withdrawLimitEpsilon >= batch.RequiredFlow {
			updates["completed_flow"] = utils.Truncate2(batch.RequiredFlow)
			updates["status"] = pojo.WithdrawFlowBatchStatusCompleted
			updates["completed_at"] = outbox.OccurredAt
		}
		if err := tx.Model(&pojo.TgUserWithdrawFlowBatch{}).Where("id = ?", batch.ID).Updates(updates).Error; err != nil {
			return err
		}
		remainingFlow = clampNonNegative(remainingFlow - allocated)
		allocatedTotal = utils.Truncate2(allocatedTotal + allocated)
	}

	if allocatedTotal > 0 {
		return tx.Model(&pojo.TgUserWithdrawFlowEvent{}).Where("id = ?", event.ID).Update("allocated_amount", allocatedTotal).Error
	}
	return nil
}

func GetUserWithdrawFlowBatchSummary(db *gorm.DB, userID int64) (pojo.TgWithdrawFlowBatchSummaryBack, error) {
	var result pojo.TgWithdrawFlowBatchSummaryBack
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

	summary, err := getUnfinishedWithdrawFlowBatchTotals(db, user.ID)
	if err != nil {
		return result, err
	}
	result.Balance = utils.Truncate2(user.Balance)
	result.HasUnfinishedBatch = summary.Count > 0
	result.RequiredFlow = summary.RequiredFlow
	result.CompletedFlow = summary.CompletedFlow
	result.RemainingFlow = summary.RemainingFlow
	result.UnfinishedBatchCount = summary.Count
	result.WithdrawMultiplier = loadWithdrawFlowBatchMultiplier(db, "withdraw_limit", 2, "v2充值到账金额提现所需流水倍数")
	result.GiftMultiplier = loadWithdrawFlowBatchMultiplier(db, "withdraw_gift_limit", 5, "v2充值赠送金额提现所需流水倍数")
	return result, nil
}

func GetUserWithdrawFlowBatchOverview(db *gorm.DB, userID int64) (pojo.TgWithdrawFlowBatchOverviewBack, error) {
	var result pojo.TgWithdrawFlowBatchOverviewBack
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

	var totalRow struct {
		Count         int64   `gorm:"column:count"`
		RequiredFlow  float64 `gorm:"column:required_flow"`
		CompletedFlow float64 `gorm:"column:completed_flow"`
	}
	if err := db.Model(&pojo.TgUserWithdrawFlowBatch{}).
		Select("COUNT(*) AS count, COALESCE(SUM(required_flow), 0) AS required_flow, COALESCE(SUM(completed_flow), 0) AS completed_flow").
		Where("user_id = ?", user.ID).
		Scan(&totalRow).Error; err != nil {
		return result, err
	}

	unfinished, err := getUnfinishedWithdrawFlowBatchTotals(db, user.ID)
	if err != nil {
		return result, err
	}

	var batches []pojo.TgUserWithdrawFlowBatch
	if err := db.Model(&pojo.TgUserWithdrawFlowBatch{}).
		Where("user_id = ?", user.ID).
		Order("id desc").
		Limit(100).
		Find(&batches).Error; err != nil {
		return result, err
	}

	result.UserID = user.ID
	result.Balance = utils.Truncate2(user.Balance)
	result.TotalFlow = totalFlow
	result.TotalRequiredFlow = utils.Truncate2(totalRow.RequiredFlow)
	result.TotalCompletedFlow = utils.Truncate2(totalRow.CompletedFlow)
	result.HasUnfinishedBatch = unfinished.Count > 0
	result.UnfinishedRequiredFlow = unfinished.RequiredFlow
	result.UnfinishedCompletedFlow = unfinished.CompletedFlow
	result.UnfinishedRemainingFlow = unfinished.RemainingFlow
	result.BatchCount = totalRow.Count
	result.UnfinishedBatchCount = unfinished.Count
	result.Batches = make([]pojo.TgWithdrawFlowBatchBack, 0, len(batches))
	for _, batch := range batches {
		result.Batches = append(result.Batches, buildWithdrawFlowBatchBack(batch))
	}
	return result, nil
}

func buildWithdrawFlowBatchBack(batch pojo.TgUserWithdrawFlowBatch) pojo.TgWithdrawFlowBatchBack {
	requiredFlow := utils.Truncate2(batch.RequiredFlow)
	completedFlow := utils.Truncate2(batch.CompletedFlow)
	progressPercent := 0.0
	if requiredFlow > 0 {
		progressPercent = utils.Truncate2(completedFlow / requiredFlow * 100)
		if progressPercent > 100 {
			progressPercent = 100
		}
	}
	return pojo.TgWithdrawFlowBatchBack{
		ID:                 batch.ID,
		CreatedAt:          batch.CreatedAt,
		UpdatedAt:          batch.UpdatedAt,
		TenantID:           batch.TenantID,
		UserID:             batch.UserID,
		SourceType:         batch.SourceType,
		SourceOrderID:      batch.SourceOrderID,
		SourceOrderNo:      batch.SourceOrderNo,
		ActivityType:       batch.ActivityType,
		ActivityCode:       batch.ActivityCode,
		CreditAmount:       utils.Truncate2(batch.CreditAmount),
		BonusAmount:        utils.Truncate2(batch.BonusAmount),
		BaseAmount:         utils.Truncate2(batch.BaseAmount),
		WithdrawMultiplier: utils.Truncate2(batch.WithdrawMultiplier),
		GiftMultiplier:     utils.Truncate2(batch.GiftMultiplier),
		RequiredFlow:       requiredFlow,
		CompletedFlow:      completedFlow,
		RemainingFlow:      clampNonNegative(requiredFlow - completedFlow),
		ProgressPercent:    progressPercent,
		Status:             batch.Status,
		ClosedReason:       batch.ClosedReason,
		CompletedAt:        batch.CompletedAt,
		ClosedAt:           batch.ClosedAt,
		LastFlowAt:         batch.LastFlowAt,
	}
}

func EnsureNoUnfinishedWithdrawFlowBatches(tx *gorm.DB, userID int64) error {
	summary, err := getUnfinishedWithdrawFlowBatchTotals(tx, userID)
	if err != nil {
		return err
	}
	if summary.Count > 0 {
		return errors.New(utils.I18nMessage("withdraw_flow_insufficient", map[string]interface{}{
			"available": fmt.Sprintf("%.2f", summary.CompletedFlow),
			"required":  fmt.Sprintf("%.2f", summary.RequiredFlow),
		}))
	}
	return nil
}

type withdrawFlowBatchTotals struct {
	Count         int64
	RequiredFlow  float64
	CompletedFlow float64
	RemainingFlow float64
}

func getUnfinishedWithdrawFlowBatchTotals(db *gorm.DB, userID int64) (withdrawFlowBatchTotals, error) {
	var result withdrawFlowBatchTotals
	if db == nil || userID <= 0 {
		return result, nil
	}
	var row struct {
		Count         int64   `gorm:"column:count"`
		RequiredFlow  float64 `gorm:"column:required_flow"`
		CompletedFlow float64 `gorm:"column:completed_flow"`
	}
	err := db.Model(&pojo.TgUserWithdrawFlowBatch{}).
		Select("COUNT(*) AS count, COALESCE(SUM(required_flow), 0) AS required_flow, COALESCE(SUM(completed_flow), 0) AS completed_flow").
		Where("user_id = ? AND status = ? AND completed_flow < required_flow", userID, pojo.WithdrawFlowBatchStatusActive).
		Scan(&row).Error
	if err != nil {
		return result, err
	}
	result.Count = row.Count
	result.RequiredFlow = utils.Truncate2(row.RequiredFlow)
	result.CompletedFlow = utils.Truncate2(row.CompletedFlow)
	result.RemainingFlow = clampNonNegative(result.RequiredFlow - result.CompletedFlow)
	return result, nil
}

func completeWithdrawFlowBatch(tx *gorm.DB, batchID int64, completedAt time.Time, requiredFlow float64) error {
	if tx == nil || batchID <= 0 {
		return nil
	}
	if completedAt.IsZero() {
		completedAt = time.Now()
	}
	return tx.Model(&pojo.TgUserWithdrawFlowBatch{}).Where("id = ?", batchID).Updates(map[string]any{
		"completed_flow": utils.Truncate2(requiredFlow),
		"status":         pojo.WithdrawFlowBatchStatusCompleted,
		"completed_at":   completedAt,
		"last_flow_at":   completedAt,
	}).Error
}

func loadWithdrawFlowBatchMultiplier(tx *gorm.DB, key string, defaultValue float64, desc string) float64 {
	if tx == nil || strings.TrimSpace(key) == "" {
		return defaultValue
	}
	var cfg pojo.SysConfig
	if err := tx.Where("config_key = ?", key).First(&cfg).Error; err != nil || cfg.ID == 0 {
		_ = tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&pojo.SysConfig{
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
