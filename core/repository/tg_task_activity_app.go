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

func GetAppTaskActivityList(db *gorm.DB, userID int64, lang string) ([]pojo.TgTaskActivityAppItem, error) {
	now := time.Now()
	if err := expireUserTaskActivityRecords(db, userID, now); err != nil {
		return nil, err
	}

	var configs []pojo.TgTaskActivityConfig
	if err := db.Model(&pojo.TgTaskActivityConfig{}).
		Where("status = ?", 1).
		Where("(start_at IS NULL OR start_at <= ?) AND (end_at IS NULL OR end_at > ?)", now, now).
		Order("sort asc, id desc").
		Find(&configs).Error; err != nil {
		return nil, err
	}
	if len(configs) == 0 {
		return []pojo.TgTaskActivityAppItem{}, nil
	}

	configIDs := make([]int64, 0, len(configs))
	for _, cfg := range configs {
		configIDs = append(configIDs, cfg.ID)
	}
	var records []pojo.TgTaskActivityRecord
	if err := db.Where("user_id = ? AND config_id IN ? AND status IN ?", userID, configIDs, taskActivityActiveRecordStatuses()).
		Order("id desc").
		Find(&records).Error; err != nil {
		return nil, err
	}
	recordByConfig := map[int64]pojo.TgTaskActivityRecord{}
	for _, record := range records {
		if _, exists := recordByConfig[record.ConfigID]; exists {
			continue
		}
		recordByConfig[record.ConfigID] = record
	}

	result := make([]pojo.TgTaskActivityAppItem, 0, len(configs))
	for _, cfg := range configs {
		record, ok := recordByConfig[cfg.ID]
		if ok {
			result = append(result, taskActivityAppItemFromRecord(record, lang, now))
			continue
		}
		result = append(result, taskActivityAppItemFromConfig(cfg, lang))
	}
	return result, nil
}

func ClaimAppTaskActivity(db *gorm.DB, userID int64, configID int64, lang string) (pojo.TgTaskActivityAppItem, error) {
	if userID <= 0 || configID <= 0 {
		return pojo.TgTaskActivityAppItem{}, errors.New("invalid_params")
	}
	now := time.Now()
	var record pojo.TgTaskActivityRecord
	err := db.Transaction(func(tx *gorm.DB) error {
		var user pojo.TgUser
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Select("id").
			Where("id = ?", userID).
			First(&user).Error; err != nil {
			return err
		}
		if err := expireUserTaskActivityRecords(tx, userID, now); err != nil {
			return err
		}
		var existing pojo.TgTaskActivityRecord
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("user_id = ? AND status IN ?", userID, taskActivityActiveRecordStatuses()).
			Order("id desc").
			First(&existing).Error
		if err == nil && existing.ID > 0 {
			if existing.ConfigID != configID {
				return errors.New("task_activity_has_active")
			}
			record = existing
			return nil
		}
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		var cfg pojo.TgTaskActivityConfig
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", configID).First(&cfg).Error; err != nil {
			return err
		}
		if cfg.ID == 0 || cfg.Status != 1 {
			return errors.New("task_activity_not_found")
		}
		if cfg.StartAt != nil && cfg.StartAt.After(now) {
			return errors.New("task_activity_not_started")
		}
		if cfg.EndAt != nil && !cfg.EndAt.After(now) {
			return errors.New("task_activity_ended")
		}

		deadlineAt := now.Add(time.Duration(cfg.DurationMinutes) * time.Minute)
		record = pojo.TgTaskActivityRecord{
			ConfigID:               cfg.ID,
			UserID:                 userID,
			Title:                  cfg.Title,
			TitleI18n:              cfg.TitleI18n,
			SubTitle:               cfg.SubTitle,
			SubTitleI18n:           cfg.SubTitleI18n,
			LevelCode:              cfg.LevelCode,
			LevelName:              cfg.LevelName,
			TaskType:               cfg.TaskType,
			RequiredInviteCount:    cfg.RequiredInviteCount,
			RequiredRechargeAmount: utils.Truncate2(cfg.RequiredRechargeAmount),
			DurationMinutes:        cfg.DurationMinutes,
			RewardAmount:           utils.Truncate2(cfg.RewardAmount),
			RewardCurrency:         cfg.RewardCurrency,
			RewardTarget:           cfg.RewardTarget,
			Status:                 pojo.TaskActivityRecordStatusProgress,
			ClaimedAt:              now,
			DeadlineAt:             deadlineAt,
		}
		return tx.Create(&record).Error
	})
	if err != nil {
		return pojo.TgTaskActivityAppItem{}, err
	}
	return taskActivityAppItemFromRecord(record, lang, time.Now()), nil
}

func GetAppCurrentTaskActivity(db *gorm.DB, userID int64, lang string) (*pojo.TgTaskActivityAppItem, error) {
	now := time.Now()
	if err := expireUserTaskActivityRecords(db, userID, now); err != nil {
		return nil, err
	}
	var record pojo.TgTaskActivityRecord
	if err := db.Where("user_id = ? AND status IN ?", userID, taskActivityActiveRecordStatuses()).
		Order("id desc").
		First(&record).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	item := taskActivityAppItemFromRecord(record, lang, now)
	return &item, nil
}

func GetAppTaskActivityRecords(db *gorm.DB, userID int64, page pojo.PageInfo, lang string) (pojo.TgTaskActivityAppRecordResp, error) {
	var result pojo.TgTaskActivityAppRecordResp
	if page.PageSize <= 0 {
		page.PageSize = 10
	}
	if page.CurrentPage < 0 {
		page.CurrentPage = 0
	}
	now := time.Now()
	if err := expireUserTaskActivityRecords(db, userID, now); err != nil {
		return result, err
	}

	var records []pojo.TgTaskActivityRecord
	query := db.Model(&pojo.TgTaskActivityRecord{}).Where("user_id = ?", userID)
	if err := query.Count(&result.Total).Error; err != nil {
		return result, err
	}
	if err := query.Order("id desc").
		Limit(page.PageSize).
		Offset(page.PageSize * page.CurrentPage).
		Find(&records).Error; err != nil {
		return result, err
	}
	result.List = make([]pojo.TgTaskActivityAppItem, 0, len(records))
	for _, record := range records {
		result.List = append(result.List, taskActivityAppItemFromRecord(record, lang, now))
	}
	result.PageSize = page.PageSize
	result.CurrentPage = page.CurrentPage
	return result, nil
}

func ApplyTaskActivityRechargeProgress(tx *gorm.DB, order pojo.RechargeOrder, occurredAt time.Time) error {
	if tx == nil || order.UserId <= 0 {
		return nil
	}
	if order.IsDev != nil && *order.IsDev == 1 {
		return nil
	}
	if occurredAt.IsZero() {
		occurredAt = time.Now()
	}

	var subUser pojo.TgUser
	if err := tx.Where("id = ?", order.UserId).First(&subUser).Error; err != nil || subUser.ID == 0 {
		return nil
	}
	if subUser.ParentID == nil || *subUser.ParentID <= 0 {
		return nil
	}
	parentID := *subUser.ParentID

	var records []pojo.TgTaskActivityRecord
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("user_id = ? AND status = ? AND claimed_at <= ? AND deadline_at >= ?", parentID, pojo.TaskActivityRecordStatusProgress, occurredAt, occurredAt).
		Find(&records).Error; err != nil {
		return err
	}
	for _, record := range records {
		qualifiedCount, totalRecharge, err := taskActivityProgressForRecord(tx, record)
		if err != nil {
			return err
		}
		updates := map[string]any{
			"progress_invite_count":    qualifiedCount,
			"progress_recharge_count":  qualifiedCount,
			"progress_recharge_amount": utils.Truncate2(totalRecharge),
		}
		if qualifiedCount >= record.RequiredInviteCount {
			if err := rewardTaskActivityRecord(tx, record, qualifiedCount, totalRecharge, occurredAt); err != nil {
				return err
			}
			continue
		}
		if err := tx.Model(&pojo.TgTaskActivityRecord{}).Where("id = ? AND status = ?", record.ID, pojo.TaskActivityRecordStatusProgress).Updates(updates).Error; err != nil {
			return err
		}
	}
	return nil
}

func rewardTaskActivityRecord(tx *gorm.DB, record pojo.TgTaskActivityRecord, qualifiedCount int, totalRecharge float64, rewardedAt time.Time) error {
	idempotencyKey := fmt.Sprintf("task_activity_reward:%d", record.ID)
	var existing pojo.TgUserRebateRecord
	tx.Where("idempotency_key = ?", idempotencyKey).First(&existing)
	if existing.ID > 0 {
		return tx.Model(&pojo.TgTaskActivityRecord{}).
			Where("id = ? AND status = ?", record.ID, pojo.TaskActivityRecordStatusProgress).
			Updates(map[string]any{
				"progress_invite_count":    qualifiedCount,
				"progress_recharge_count":  qualifiedCount,
				"progress_recharge_amount": utils.Truncate2(totalRecharge),
				"status":                   pojo.TaskActivityRecordStatusRewarded,
				"completed_at":             rewardedAt,
				"rewarded_at":              rewardedAt,
				"rebate_record_id":         existing.ID,
			}).Error
	}

	var user pojo.TgUser
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", record.UserID).First(&user).Error; err != nil {
		return err
	}
	rewardAmount := utils.Truncate2(record.RewardAmount)
	if rewardAmount <= 0 {
		return errors.New("task_activity_reward_amount_invalid")
	}
	if err := tx.Model(&pojo.TgUser{}).Where("id = ?", record.UserID).Updates(map[string]any{
		"rebate_amount":       gorm.Expr("rebate_amount + ?", rewardAmount),
		"rebate_total_amount": gorm.Expr("rebate_total_amount + ?", rewardAmount),
	}).Error; err != nil {
		return err
	}
	tenantID := user.TenantId
	remark := fmt.Sprintf("task_activity_reward record_id=%d", record.ID)
	rebateRecord := pojo.TgUserRebateRecord{
		TenantId:       &tenantID,
		SubUserId:      record.UserID,
		ParentUserId:   record.UserID,
		SourceType:     pojo.TgUserRebateSourceTypeInviteTier,
		SourceOrderId:  fmt.Sprintf("task_activity:%d", record.ID),
		SourceAmount:   utils.Truncate2(totalRecharge),
		RebateRate:     0,
		RebateAmount:   rewardAmount,
		Currency:       normalizeTaskActivityRewardCurrency(record.RewardCurrency),
		Status:         1,
		SettledAt:      &rewardedAt,
		IdempotencyKey: idempotencyKey,
		Remark:         &remark,
	}
	if err := tx.Create(&rebateRecord).Error; err != nil {
		return err
	}
	return tx.Model(&pojo.TgTaskActivityRecord{}).
		Where("id = ? AND status = ?", record.ID, pojo.TaskActivityRecordStatusProgress).
		Updates(map[string]any{
			"progress_invite_count":    qualifiedCount,
			"progress_recharge_count":  qualifiedCount,
			"progress_recharge_amount": utils.Truncate2(totalRecharge),
			"status":                   pojo.TaskActivityRecordStatusRewarded,
			"completed_at":             rewardedAt,
			"rewarded_at":              rewardedAt,
			"rebate_record_id":         rebateRecord.ID,
		}).Error
}

func taskActivityProgressForRecord(tx *gorm.DB, record pojo.TgTaskActivityRecord) (int, float64, error) {
	type progressRow struct {
		QualifiedCount int
		TotalRecharge  float64
	}
	subQuery := tx.Table(pojo.TgUserTableName+" AS u").
		Select("u.id, COALESCE(SUM(ro.amount), 0) AS total_recharge").
		Joins("JOIN "+pojo.RechargeOrderTableName+" AS ro ON ro.user_id = u.id").
		Where("u.parent_id = ?", record.UserID).
		Where("u.created_at >= ? AND u.created_at <= ?", record.ClaimedAt, record.DeadlineAt).
		Where("ro.status = ? AND COALESCE(ro.is_dev, 0) = 0", 1).
		Where("ro.pay_time >= ? AND ro.pay_time <= ?", record.ClaimedAt, record.DeadlineAt).
		Group("u.id")

	var row progressRow
	err := tx.Table("(?) AS q", subQuery).
		Select("COUNT(1) AS qualified_count, COALESCE(SUM(q.total_recharge), 0) AS total_recharge").
		Where("q.total_recharge >= ?", record.RequiredRechargeAmount).
		Scan(&row).Error
	return row.QualifiedCount, utils.Truncate2(row.TotalRecharge), err
}

func expireUserTaskActivityRecords(db *gorm.DB, userID int64, now time.Time) error {
	if userID <= 0 {
		return nil
	}
	return db.Model(&pojo.TgTaskActivityRecord{}).
		Where("user_id = ? AND status = ? AND deadline_at < ?", userID, pojo.TaskActivityRecordStatusProgress, now).
		Update("status", pojo.TaskActivityRecordStatusExpired).Error
}

func taskActivityActiveRecordStatuses() []int8 {
	return []int8{
		pojo.TaskActivityRecordStatusProgress,
	}
}

func taskActivityAppItemFromConfig(cfg pojo.TgTaskActivityConfig, lang string) pojo.TgTaskActivityAppItem {
	return pojo.TgTaskActivityAppItem{
		ConfigID:               cfg.ID,
		Title:                  resolveTaskActivityI18nText(cfg.TitleI18n, cfg.Title, lang),
		SubTitle:               resolveTaskActivityI18nText(cfg.SubTitleI18n, cfg.SubTitle, lang),
		LevelCode:              cfg.LevelCode,
		LevelName:              cfg.LevelName,
		RequiredInviteCount:    cfg.RequiredInviteCount,
		RequiredRechargeAmount: utils.Truncate2(cfg.RequiredRechargeAmount),
		DurationMinutes:        cfg.DurationMinutes,
		RewardAmount:           utils.Truncate2(cfg.RewardAmount),
		RewardCurrency:         normalizeTaskActivityRewardCurrency(cfg.RewardCurrency),
		RewardTarget:           cfg.RewardTarget,
		Claimed:                false,
		ProgressPercent:        0,
	}
}

func taskActivityAppItemFromRecord(record pojo.TgTaskActivityRecord, lang string, now time.Time) pojo.TgTaskActivityAppItem {
	status := record.Status
	claimedAt := record.ClaimedAt
	deadlineAt := record.DeadlineAt
	remainingSeconds := int64(0)
	if record.Status == pojo.TaskActivityRecordStatusProgress && record.DeadlineAt.After(now) {
		remainingSeconds = int64(record.DeadlineAt.Sub(now).Seconds())
	}
	progressPercent := float64(0)
	if record.RequiredInviteCount > 0 {
		progressPercent = utils.Truncate2(float64(record.ProgressInviteCount) / float64(record.RequiredInviteCount) * 100)
		if progressPercent > 100 {
			progressPercent = 100
		}
	}
	return pojo.TgTaskActivityAppItem{
		ConfigID:               record.ConfigID,
		RecordID:               record.ID,
		Title:                  resolveTaskActivityI18nText(record.TitleI18n, record.Title, lang),
		SubTitle:               resolveTaskActivityI18nText(record.SubTitleI18n, record.SubTitle, lang),
		LevelCode:              record.LevelCode,
		LevelName:              record.LevelName,
		RequiredInviteCount:    record.RequiredInviteCount,
		RequiredRechargeAmount: utils.Truncate2(record.RequiredRechargeAmount),
		DurationMinutes:        record.DurationMinutes,
		RewardAmount:           utils.Truncate2(record.RewardAmount),
		RewardCurrency:         normalizeTaskActivityRewardCurrency(record.RewardCurrency),
		RewardTarget:           record.RewardTarget,
		Claimed:                true,
		Status:                 &status,
		ClaimedAt:              &claimedAt,
		DeadlineAt:             &deadlineAt,
		RemainingSeconds:       remainingSeconds,
		ProgressInviteCount:    record.ProgressInviteCount,
		ProgressRechargeCount:  record.ProgressRechargeCount,
		ProgressRechargeAmount: utils.Truncate2(record.ProgressRechargeAmount),
		ProgressPercent:        progressPercent,
	}
}

func resolveTaskActivityI18nText(raw *string, fallback string, lang string) string {
	valuesAny := parseTaskActivityI18nJSON(raw)
	values, ok := valuesAny.(map[string]string)
	if !ok || len(values) == 0 {
		return fallback
	}
	lang = normalizeTaskActivityLang(lang)
	if lang != "" {
		if value := strings.TrimSpace(values[lang]); value != "" {
			return value
		}
		if idx := strings.Index(lang, "-"); idx > 0 {
			if value := strings.TrimSpace(values[lang[:idx]]); value != "" {
				return value
			}
		}
	}
	return fallback
}

func normalizeTaskActivityLang(lang string) string {
	lang = strings.TrimSpace(lang)
	if lang == "" {
		return ""
	}
	if idx := strings.Index(lang, ","); idx >= 0 {
		lang = strings.TrimSpace(lang[:idx])
	}
	if idx := strings.Index(lang, ";"); idx >= 0 {
		lang = strings.TrimSpace(lang[:idx])
	}
	return lang
}

func normalizeTaskActivityRewardCurrency(currency string) string {
	currency = strings.TrimSpace(currency)
	if currency == "" {
		return "USD"
	}
	return currency
}
