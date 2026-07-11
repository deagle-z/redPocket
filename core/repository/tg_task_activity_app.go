package repository

import (
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/utils"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type taskActivityProgressCandidate struct {
	Config          pojo.TgTaskActivityConfig
	QualifiedCount  int
	TotalRecharge   float64
	ProgressPercent float64
	Matched         bool
}

const (
	taskActivityCustomEnabledKey         = "task_activity_custom_enabled"
	taskActivityCustomRechargeAmountKey  = "task_activity_custom_recharge_amount"
	taskActivityCustomRewardPerUserKey   = "task_activity_custom_reward_per_user"
	taskActivityCustomDurationMinutesKey = "task_activity_custom_duration_minutes"
	taskActivityCustomRewardCurrencyKey  = "task_activity_custom_reward_currency"

	taskActivityCustomDefaultInviteCount     = 1
	taskActivityCustomMinInviteCount         = 1
	taskActivityCustomMaxInviteCount         = 1000
	taskActivityCustomDefaultRechargeAmount  = 50
	taskActivityCustomDefaultRewardPerUser   = 5
	taskActivityCustomDefaultDurationMinutes = 1440
	taskActivityCustomDefaultRewardCurrency  = "USD"
)

type taskActivityCustomSettings struct {
	Enabled         bool
	RechargeAmount  float64
	RewardPerUser   float64
	DurationMinutes int
	RewardCurrency  string
}

func GetAppTaskActivityList(db *gorm.DB, userID int64, lang string) ([]pojo.TgTaskActivityAppItem, error) {
	now := time.Now()
	if err := settleUserTaskActivityRecords(db, userID, now); err != nil {
		return nil, err
	}

	if record, ok, err := getProgressTaskActivityRecord(db, userID, false); err != nil {
		return nil, err
	} else if ok {
		return taskActivityAppItemsFromRecordConfigs(db, record, lang, now)
	}

	configs, err := taskActivityConfigsAt(db, now, taskActivityDisplayOrder())
	if err != nil {
		return nil, err
	}
	items := make([]pojo.TgTaskActivityAppItem, 0, len(configs)+1)
	for _, cfg := range configs {
		items = append(items, taskActivityAppItemFromConfig(cfg, lang))
	}
	if customCfg, ok, err := taskActivityCustomConfig(db, taskActivityCustomDefaultInviteCount); err != nil {
		return nil, err
	} else if ok {
		items = append(items, taskActivityAppItemFromConfig(customCfg, lang))
	}
	return items, nil
}

func ClaimAppTaskActivity(db *gorm.DB, userID int64, req pojo.TgTaskActivityClaimReq, lang string) (pojo.TgTaskActivityAppItem, error) {
	if userID <= 0 {
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
		if err := settleUserTaskActivityRecords(tx, userID, now); err != nil {
			return err
		}
		if existing, ok, err := getProgressTaskActivityRecord(tx, userID, true); err != nil {
			return err
		} else if ok {
			record = existing
			return nil
		}

		if strings.EqualFold(strings.TrimSpace(req.LevelCode), pojo.TaskActivityLevelCustom) {
			cfg, ok, err := taskActivityCustomConfig(tx, req.InviteCount)
			if err != nil {
				return err
			}
			if !ok {
				return errors.New("task_activity_custom_disabled")
			}
			record = newTaskActivityRecordFromConfig(userID, cfg, now, cfg.DurationMinutes)
			return tx.Create(&record).Error
		}

		configs, err := taskActivityConfigsAt(tx, now, taskActivityDisplayOrder())
		if err != nil {
			return err
		}
		if len(configs) == 0 {
			return errors.New("task_activity_not_found")
		}
		cfg := configs[0]
		durationMinutes := maxTaskActivityDurationMinutes(configs)
		if durationMinutes <= 0 {
			return errors.New("task_activity_duration_invalid")
		}

		record = newTaskActivityRecordFromConfig(userID, cfg, now, durationMinutes)
		return tx.Create(&record).Error
	})
	if err != nil {
		return pojo.TgTaskActivityAppItem{}, err
	}
	return taskActivityAppItemFromRecord(record, lang, time.Now()), nil
}

func RewardAppTaskActivity(db *gorm.DB, userID int64, lang string) (pojo.TgTaskActivityAppItem, error) {
	if userID <= 0 {
		return pojo.TgTaskActivityAppItem{}, errors.New("invalid_params")
	}
	now := time.Now()
	var record pojo.TgTaskActivityRecord

	err := db.Transaction(func(tx *gorm.DB) error {
		if err := settleUserTaskActivityRecords(tx, userID, now); err != nil {
			return err
		}
		current, ok, err := getProgressTaskActivityRecord(tx, userID, true)
		if err != nil {
			return err
		}
		if !ok {
			return errors.New("task_activity_no_active")
		}
		candidate, matched, err := selectTaskActivityCandidate(tx, current)
		if err != nil {
			return err
		}
		if !matched {
			return errors.New("task_activity_not_reached")
		}
		if err := rewardTaskActivityRecord(tx, current, candidate.Config, candidate.QualifiedCount, candidate.TotalRecharge, now); err != nil {
			return err
		}
		return tx.Where("id = ?", current.ID).First(&record).Error
	})
	if err != nil {
		return pojo.TgTaskActivityAppItem{}, err
	}
	return taskActivityAppItemFromRecord(record, lang, time.Now()), nil
}

func GetAppCurrentTaskActivity(db *gorm.DB, userID int64, lang string) (*pojo.TgTaskActivityAppItem, error) {
	now := time.Now()
	if err := settleUserTaskActivityRecords(db, userID, now); err != nil {
		return nil, err
	}
	record, ok, err := getProgressTaskActivityRecord(db, userID, false)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, nil
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
	if err := settleUserTaskActivityRecords(db, userID, now); err != nil {
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
		candidate, _, err := selectTaskActivityCandidate(tx, record)
		if err != nil {
			return err
		}
		if !usableTaskActivityCandidate(candidate) {
			continue
		}
		if err := tx.Model(&pojo.TgTaskActivityRecord{}).
			Where("id = ? AND status = ?", record.ID, pojo.TaskActivityRecordStatusProgress).
			Updates(taskActivityCandidateUpdates(candidate, pojo.TaskActivityRecordStatusProgress, nil, nil, 0)).Error; err != nil {
			return err
		}
	}
	return nil
}

func settleUserTaskActivityRecords(db *gorm.DB, userID int64, now time.Time) error {
	if userID <= 0 {
		return nil
	}
	var records []pojo.TgTaskActivityRecord
	if err := db.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("user_id = ? AND status = ? AND deadline_at <= ?", userID, pojo.TaskActivityRecordStatusProgress, now).
		Find(&records).Error; err != nil {
		return err
	}
	for _, record := range records {
		candidate, matched, err := selectTaskActivityCandidate(db, record)
		if err != nil {
			return err
		}
		if matched {
			if err := rewardTaskActivityRecord(db, record, candidate.Config, candidate.QualifiedCount, candidate.TotalRecharge, now); err != nil {
				return err
			}
			continue
		}
		updates := map[string]any{
			"status": pojo.TaskActivityRecordStatusExpired,
		}
		if usableTaskActivityCandidate(candidate) {
			updates = taskActivityCandidateUpdates(candidate, pojo.TaskActivityRecordStatusExpired, nil, nil, 0)
		}
		if err := db.Model(&pojo.TgTaskActivityRecord{}).
			Where("id = ? AND status = ?", record.ID, pojo.TaskActivityRecordStatusProgress).
			Updates(updates).Error; err != nil {
			return err
		}
	}
	return nil
}

func rewardTaskActivityRecord(tx *gorm.DB, record pojo.TgTaskActivityRecord, cfg pojo.TgTaskActivityConfig, qualifiedCount int, totalRecharge float64, rewardedAt time.Time) error {
	idempotencyKey := fmt.Sprintf("task_activity_reward:%d", record.ID)
	var existing pojo.TgUserRebateRecord
	tx.Where("idempotency_key = ?", idempotencyKey).First(&existing)
	if existing.ID > 0 {
		return tx.Model(&pojo.TgTaskActivityRecord{}).
			Where("id = ? AND status = ?", record.ID, pojo.TaskActivityRecordStatusProgress).
			Updates(taskActivityCandidateUpdates(taskActivityProgressCandidate{
				Config:         cfg,
				QualifiedCount: qualifiedCount,
				TotalRecharge:  totalRecharge,
			}, pojo.TaskActivityRecordStatusRewarded, &rewardedAt, &rewardedAt, existing.ID)).Error
	}

	var user pojo.TgUser
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", record.UserID).First(&user).Error; err != nil {
		return err
	}
	rewardAmount := utils.Truncate2(cfg.RewardAmount)
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
	remark := fmt.Sprintf("task_activity_reward record_id=%d config_id=%d", record.ID, cfg.ID)
	rebateRecord := pojo.TgUserRebateRecord{
		TenantId:       &tenantID,
		SubUserId:      record.UserID,
		ParentUserId:   record.UserID,
		SourceType:     pojo.TgUserRebateSourceTypeInviteTier,
		SourceOrderId:  fmt.Sprintf("task_activity:%d", record.ID),
		SourceAmount:   utils.Truncate2(totalRecharge),
		RebateRate:     0,
		RebateAmount:   rewardAmount,
		Currency:       normalizeTaskActivityRewardCurrency(cfg.RewardCurrency),
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
		Updates(taskActivityCandidateUpdates(taskActivityProgressCandidate{
			Config:         cfg,
			QualifiedCount: qualifiedCount,
			TotalRecharge:  totalRecharge,
		}, pojo.TaskActivityRecordStatusRewarded, &rewardedAt, &rewardedAt, rebateRecord.ID)).Error
}

func getProgressTaskActivityRecord(db *gorm.DB, userID int64, lock bool) (pojo.TgTaskActivityRecord, bool, error) {
	var record pojo.TgTaskActivityRecord
	query := db.Where("user_id = ? AND status IN ?", userID, taskActivityActiveRecordStatuses()).Order("id desc")
	if lock {
		query = query.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	if err := query.First(&record).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return record, false, nil
		}
		return record, false, err
	}
	return record, true, nil
}

func selectTaskActivityCandidate(tx *gorm.DB, record pojo.TgTaskActivityRecord) (taskActivityProgressCandidate, bool, error) {
	configs, err := taskActivityConfigsForRecord(tx, record)
	if err != nil {
		return taskActivityProgressCandidate{}, false, err
	}
	var best taskActivityProgressCandidate
	hasBest := false
	for _, cfg := range configs {
		qualifiedCount, totalRecharge, err := taskActivityProgressForThreshold(tx, record, cfg.RequiredRechargeAmount)
		if err != nil {
			return taskActivityProgressCandidate{}, false, err
		}
		progressPercent := float64(0)
		if cfg.RequiredInviteCount > 0 {
			progressPercent = utils.Truncate2(float64(qualifiedCount) / float64(cfg.RequiredInviteCount) * 100)
			if progressPercent > 100 {
				progressPercent = 100
			}
		}
		candidate := taskActivityProgressCandidate{
			Config:          cfg,
			QualifiedCount:  qualifiedCount,
			TotalRecharge:   totalRecharge,
			ProgressPercent: progressPercent,
			Matched:         cfg.RequiredInviteCount > 0 && qualifiedCount >= cfg.RequiredInviteCount,
		}
		if candidate.Matched {
			return candidate, true, nil
		}
		if !hasBest ||
			candidate.ProgressPercent > best.ProgressPercent ||
			(candidate.ProgressPercent == best.ProgressPercent && candidate.Config.RewardAmount > best.Config.RewardAmount) {
			best = candidate
			hasBest = true
		}
	}
	return best, false, nil
}

func taskActivityProgressForThreshold(tx *gorm.DB, record pojo.TgTaskActivityRecord, requiredRechargeAmount float64) (int, float64, error) {
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
		Where("q.total_recharge >= ?", utils.Truncate2(requiredRechargeAmount)).
		Scan(&row).Error
	return row.QualifiedCount, utils.Truncate2(row.TotalRecharge), err
}

func taskActivityConfigsAt(db *gorm.DB, at time.Time, order string) ([]pojo.TgTaskActivityConfig, error) {
	var configs []pojo.TgTaskActivityConfig
	if order == "" {
		order = "sort asc, id desc"
	}
	err := db.Model(&pojo.TgTaskActivityConfig{}).
		Where("status = ?", 1).
		Where("task_type = ?", pojo.TaskActivityTypeInviteRecharge).
		Where("(start_at IS NULL OR start_at <= ?) AND (end_at IS NULL OR end_at > ?)", at, at).
		Order(order).
		Find(&configs).Error
	return configs, err
}

func taskActivityConfigsForRecord(db *gorm.DB, record pojo.TgTaskActivityRecord) ([]pojo.TgTaskActivityConfig, error) {
	if isCustomTaskActivityRecord(record) {
		return []pojo.TgTaskActivityConfig{taskActivityConfigFromRecord(record)}, nil
	}
	at := record.ClaimedAt
	if at.IsZero() {
		at = time.Now()
	}
	return taskActivityConfigsAt(db, at, taskActivityRewardDescOrder())
}

func taskActivityCandidateUpdates(candidate taskActivityProgressCandidate, status int8, completedAt *time.Time, rewardedAt *time.Time, rebateRecordID int64) map[string]any {
	updates := map[string]any{
		"progress_invite_count":    candidate.QualifiedCount,
		"progress_recharge_count":  candidate.QualifiedCount,
		"progress_recharge_amount": utils.Truncate2(candidate.TotalRecharge),
		"status":                   status,
	}
	if candidate.Config.ID > 0 {
		for key, value := range taskActivityConfigSnapshotUpdates(candidate.Config) {
			updates[key] = value
		}
	}
	if completedAt != nil {
		updates["completed_at"] = *completedAt
	}
	if rewardedAt != nil {
		updates["rewarded_at"] = *rewardedAt
	}
	if rebateRecordID > 0 {
		updates["rebate_record_id"] = rebateRecordID
	}
	return updates
}

func taskActivityConfigSnapshotUpdates(cfg pojo.TgTaskActivityConfig) map[string]any {
	return map[string]any{
		"config_id":                cfg.ID,
		"title":                    cfg.Title,
		"title_i18n":               cfg.TitleI18n,
		"sub_title":                cfg.SubTitle,
		"sub_title_i18n":           cfg.SubTitleI18n,
		"level_code":               cfg.LevelCode,
		"level_name":               cfg.LevelName,
		"task_type":                normalizeTaskActivityTaskType(cfg.TaskType),
		"required_invite_count":    cfg.RequiredInviteCount,
		"required_recharge_amount": utils.Truncate2(cfg.RequiredRechargeAmount),
		"duration_minutes":         cfg.DurationMinutes,
		"reward_amount":            utils.Truncate2(cfg.RewardAmount),
		"reward_currency":          normalizeTaskActivityRewardCurrency(cfg.RewardCurrency),
		"reward_target":            normalizeTaskActivityRewardTarget(cfg.RewardTarget),
	}
}

func newTaskActivityRecordFromConfig(userID int64, cfg pojo.TgTaskActivityConfig, claimedAt time.Time, durationMinutes int) pojo.TgTaskActivityRecord {
	return pojo.TgTaskActivityRecord{
		ConfigID:               0,
		UserID:                 userID,
		Title:                  cfg.Title,
		TitleI18n:              cfg.TitleI18n,
		SubTitle:               cfg.SubTitle,
		SubTitleI18n:           cfg.SubTitleI18n,
		LevelCode:              cfg.LevelCode,
		LevelName:              cfg.LevelName,
		TaskType:               normalizeTaskActivityTaskType(cfg.TaskType),
		RequiredInviteCount:    cfg.RequiredInviteCount,
		RequiredRechargeAmount: utils.Truncate2(cfg.RequiredRechargeAmount),
		DurationMinutes:        durationMinutes,
		RewardAmount:           utils.Truncate2(cfg.RewardAmount),
		RewardCurrency:         normalizeTaskActivityRewardCurrency(cfg.RewardCurrency),
		RewardTarget:           normalizeTaskActivityRewardTarget(cfg.RewardTarget),
		Status:                 pojo.TaskActivityRecordStatusProgress,
		ClaimedAt:              claimedAt,
		DeadlineAt:             claimedAt.Add(time.Duration(durationMinutes) * time.Minute),
	}
}

func isCustomTaskActivityRecord(record pojo.TgTaskActivityRecord) bool {
	return strings.EqualFold(strings.TrimSpace(record.LevelCode), pojo.TaskActivityLevelCustom)
}

func usableTaskActivityCandidate(candidate taskActivityProgressCandidate) bool {
	return candidate.Config.ID > 0 || strings.EqualFold(strings.TrimSpace(candidate.Config.LevelCode), pojo.TaskActivityLevelCustom)
}

func taskActivityConfigFromRecord(record pojo.TgTaskActivityRecord) pojo.TgTaskActivityConfig {
	return pojo.TgTaskActivityConfig{
		BaseModel:              pojo.BaseModel{ID: record.ConfigID},
		Title:                  record.Title,
		TitleI18n:              record.TitleI18n,
		SubTitle:               record.SubTitle,
		SubTitleI18n:           record.SubTitleI18n,
		LevelCode:              record.LevelCode,
		LevelName:              record.LevelName,
		Status:                 1,
		TaskType:               normalizeTaskActivityTaskType(record.TaskType),
		RequiredInviteCount:    record.RequiredInviteCount,
		RequiredRechargeAmount: utils.Truncate2(record.RequiredRechargeAmount),
		DurationMinutes:        record.DurationMinutes,
		RewardAmount:           utils.Truncate2(record.RewardAmount),
		RewardCurrency:         normalizeTaskActivityRewardCurrency(record.RewardCurrency),
		RewardTarget:           normalizeTaskActivityRewardTarget(record.RewardTarget),
	}
}

func taskActivityCustomConfig(db *gorm.DB, inviteCount int) (pojo.TgTaskActivityConfig, bool, error) {
	settings, ok, err := taskActivityCustomSettingsFromSysConfig(db)
	if err != nil || !ok {
		return pojo.TgTaskActivityConfig{}, ok, err
	}
	inviteCount = normalizeTaskActivityCustomInviteCount(inviteCount)
	rewardAmount := utils.Truncate2(settings.RewardPerUser * float64(inviteCount))
	rechargeAmount := utils.Truncate2(settings.RechargeAmount)
	currency := normalizeTaskActivityRewardCurrency(settings.RewardCurrency)
	return pojo.TgTaskActivityConfig{
		Title:                  "自定义任务",
		SubTitle:               fmt.Sprintf("邀请%d位好友注册并且充值≥%s%s", inviteCount, formatTaskActivityAmountText(rechargeAmount), currency),
		LevelCode:              pojo.TaskActivityLevelCustom,
		LevelName:              "自定义任务",
		Status:                 1,
		TaskType:               pojo.TaskActivityTypeInviteRecharge,
		RequiredInviteCount:    inviteCount,
		RequiredRechargeAmount: rechargeAmount,
		DurationMinutes:        settings.DurationMinutes,
		RewardAmount:           rewardAmount,
		RewardCurrency:         currency,
		RewardTarget:           pojo.TaskActivityRewardTargetRebate,
	}, true, nil
}

func taskActivityCustomSettingsFromSysConfig(db *gorm.DB) (taskActivityCustomSettings, bool, error) {
	enabledRaw, err := taskActivitySysConfigValue(db, taskActivityCustomEnabledKey, "1", "自定义邀请任务开关")
	if err != nil {
		return taskActivityCustomSettings{}, false, err
	}
	if !taskActivityConfigEnabled(enabledRaw) {
		return taskActivityCustomSettings{}, false, nil
	}

	rechargeRaw, err := taskActivitySysConfigValue(db, taskActivityCustomRechargeAmountKey, strconv.FormatFloat(taskActivityCustomDefaultRechargeAmount, 'f', -1, 64), "自定义邀请任务单人充值门槛")
	if err != nil {
		return taskActivityCustomSettings{}, false, err
	}
	rewardRaw, err := taskActivitySysConfigValue(db, taskActivityCustomRewardPerUserKey, strconv.FormatFloat(taskActivityCustomDefaultRewardPerUser, 'f', -1, 64), "自定义邀请任务单人奖励金额")
	if err != nil {
		return taskActivityCustomSettings{}, false, err
	}
	durationRaw, err := taskActivitySysConfigValue(db, taskActivityCustomDurationMinutesKey, strconv.Itoa(taskActivityCustomDefaultDurationMinutes), "自定义邀请任务限制分钟")
	if err != nil {
		return taskActivityCustomSettings{}, false, err
	}
	currencyRaw, err := taskActivitySysConfigValue(db, taskActivityCustomRewardCurrencyKey, taskActivityCustomDefaultRewardCurrency, "自定义邀请任务奖励币种")
	if err != nil {
		return taskActivityCustomSettings{}, false, err
	}

	settings := taskActivityCustomSettings{
		Enabled:         true,
		RechargeAmount:  positiveFloatOrDefault(rechargeRaw, taskActivityCustomDefaultRechargeAmount),
		RewardPerUser:   positiveFloatOrDefault(rewardRaw, taskActivityCustomDefaultRewardPerUser),
		DurationMinutes: positiveIntOrDefault(durationRaw, taskActivityCustomDefaultDurationMinutes),
		RewardCurrency:  normalizeTaskActivityRewardCurrency(currencyRaw),
	}
	return settings, true, nil
}

func taskActivitySysConfigValue(db *gorm.DB, configKey string, defaultValue string, configDesc string) (string, error) {
	var cfg pojo.SysConfig
	err := db.Where("config_key = ?", configKey).First(&cfg).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		cfg = pojo.SysConfig{
			ConfigKey:   configKey,
			ConfigValue: defaultValue,
			ConfigDesc:  configDesc,
		}
		if createErr := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&cfg).Error; createErr != nil {
			return "", createErr
		}
		return defaultValue, nil
	}
	if err != nil {
		return "", err
	}
	value := strings.TrimSpace(cfg.ConfigValue)
	if value == "" {
		return defaultValue, nil
	}
	return value, nil
}

func taskActivityConfigEnabled(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "1", "true", "on", "yes", "enabled":
		return true
	default:
		return false
	}
}

func normalizeTaskActivityCustomInviteCount(value int) int {
	if value <= 0 {
		return taskActivityCustomDefaultInviteCount
	}
	if value < taskActivityCustomMinInviteCount {
		return taskActivityCustomMinInviteCount
	}
	if value > taskActivityCustomMaxInviteCount {
		return taskActivityCustomMaxInviteCount
	}
	return value
}

func positiveFloatOrDefault(value string, defaultValue float64) float64 {
	parsed, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
	if err != nil || parsed <= 0 {
		return defaultValue
	}
	return utils.Truncate2(parsed)
}

func positiveIntOrDefault(value string, defaultValue int) int {
	parsed, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil || parsed <= 0 {
		return defaultValue
	}
	return parsed
}

func formatTaskActivityAmountText(value float64) string {
	text := strconv.FormatFloat(utils.Truncate2(value), 'f', 2, 64)
	text = strings.TrimRight(strings.TrimRight(text, "0"), ".")
	if text == "" {
		return "0"
	}
	return text
}

func maxTaskActivityDurationMinutes(configs []pojo.TgTaskActivityConfig) int {
	maxValue := 0
	for _, cfg := range configs {
		if cfg.DurationMinutes > maxValue {
			maxValue = cfg.DurationMinutes
		}
	}
	return maxValue
}

func taskActivityRewardDescOrder() string {
	return "reward_amount desc, required_invite_count desc, required_recharge_amount desc, sort asc, id desc"
}

func taskActivityDisplayOrder() string {
	return "sort asc, id asc"
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
		RewardTarget:           normalizeTaskActivityRewardTarget(cfg.RewardTarget),
		Claimed:                false,
		ProgressPercent:        0,
	}
}

func taskActivityAppItemsFromRecordConfigs(db *gorm.DB, record pojo.TgTaskActivityRecord, lang string, now time.Time) ([]pojo.TgTaskActivityAppItem, error) {
	if isCustomTaskActivityRecord(record) {
		return []pojo.TgTaskActivityAppItem{taskActivityAppItemFromRecord(record, lang, now)}, nil
	}
	configs, err := taskActivityConfigsAt(db, record.ClaimedAt, taskActivityDisplayOrder())
	if err != nil {
		return nil, err
	}
	if len(configs) == 0 {
		return []pojo.TgTaskActivityAppItem{taskActivityAppItemFromRecord(record, lang, now)}, nil
	}

	items := make([]pojo.TgTaskActivityAppItem, 0, len(configs)+1)
	for _, cfg := range configs {
		qualifiedCount, totalRecharge, err := taskActivityProgressForThreshold(db, record, cfg.RequiredRechargeAmount)
		if err != nil {
			return nil, err
		}
		items = append(items, taskActivityAppItemFromConfigAndRecord(cfg, record, qualifiedCount, totalRecharge, lang, now))
	}
	if customCfg, ok, err := taskActivityCustomConfig(db, taskActivityCustomDefaultInviteCount); err != nil {
		return nil, err
	} else if ok {
		qualifiedCount, totalRecharge, err := taskActivityProgressForThreshold(db, record, customCfg.RequiredRechargeAmount)
		if err != nil {
			return nil, err
		}
		items = append(items, taskActivityAppItemFromConfigAndRecord(customCfg, record, qualifiedCount, totalRecharge, lang, now))
	}
	return items, nil
}

func taskActivityAppItemFromConfigAndRecord(cfg pojo.TgTaskActivityConfig, record pojo.TgTaskActivityRecord, qualifiedCount int, totalRecharge float64, lang string, now time.Time) pojo.TgTaskActivityAppItem {
	item := taskActivityAppItemFromConfig(cfg, lang)
	status := record.Status
	claimedAt := record.ClaimedAt
	deadlineAt := record.DeadlineAt
	remainingSeconds := int64(0)
	if record.Status == pojo.TaskActivityRecordStatusProgress && record.DeadlineAt.After(now) {
		remainingSeconds = int64(record.DeadlineAt.Sub(now).Seconds())
	}
	progressPercent := float64(0)
	if cfg.RequiredInviteCount > 0 {
		progressPercent = utils.Truncate2(float64(qualifiedCount) / float64(cfg.RequiredInviteCount) * 100)
		if progressPercent > 100 {
			progressPercent = 100
		}
	}
	item.RecordID = record.ID
	item.Claimed = true
	item.Status = &status
	item.ClaimedAt = &claimedAt
	item.DeadlineAt = &deadlineAt
	item.RemainingSeconds = remainingSeconds
	item.ProgressInviteCount = qualifiedCount
	item.ProgressRechargeCount = qualifiedCount
	item.ProgressRechargeAmount = utils.Truncate2(totalRecharge)
	item.ProgressPercent = progressPercent
	return item
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
		RewardTarget:           normalizeTaskActivityRewardTarget(record.RewardTarget),
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

func normalizeTaskActivityTaskType(taskType string) string {
	taskType = strings.TrimSpace(taskType)
	if taskType == "" {
		return pojo.TaskActivityTypeInviteRecharge
	}
	return taskType
}

func normalizeTaskActivityRewardTarget(target string) string {
	target = strings.TrimSpace(target)
	if target == "" {
		return pojo.TaskActivityRewardTargetRebate
	}
	return target
}
