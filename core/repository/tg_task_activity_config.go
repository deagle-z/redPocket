package repository

import (
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/utils"
	"encoding/json"
	"errors"
	"strings"

	"gorm.io/gorm"
)

// GetTgTaskActivityConfigs 任务活动配置列表（分页）
func GetTgTaskActivityConfigs(db *gorm.DB, search pojo.TgTaskActivityConfigSearch) (result pojo.TgTaskActivityConfigResp) {
	var list []pojo.TgTaskActivityConfig
	query := db.Model(&pojo.TgTaskActivityConfig{})

	title := strings.TrimSpace(search.Title)
	if title != "" {
		query = query.Where("title LIKE ?", "%"+title+"%")
	}
	levelCode := strings.TrimSpace(search.LevelCode)
	if levelCode != "" {
		query = query.Where("level_code = ?", levelCode)
	}
	if search.Status != nil {
		query = query.Where("status = ?", *search.Status)
	}

	query.Count(&result.Total)
	query.Order("sort asc, id desc").
		Limit(search.PageSize).
		Offset(search.PageSize * search.CurrentPage).
		Find(&list)

	for _, item := range list {
		result.List = append(result.List, tgTaskActivityConfigBack(item))
	}
	result.PageSize = search.PageSize
	result.CurrentPage = search.CurrentPage
	return result
}

// GetTgTaskActivityConfigById 根据ID获取任务活动配置
func GetTgTaskActivityConfigById(db *gorm.DB, id int64) (result pojo.TgTaskActivityConfigBack, err error) {
	if id <= 0 {
		return result, errors.New("record_not_found")
	}
	var entity pojo.TgTaskActivityConfig
	db.Where("id = ?", id).First(&entity)
	if entity.ID == 0 {
		return result, errors.New("record_not_found")
	}
	return tgTaskActivityConfigBack(entity), nil
}

// SetTgTaskActivityConfig 创建或更新任务活动配置
func SetTgTaskActivityConfig(db *gorm.DB, req pojo.TgTaskActivityConfigSet) (result pojo.TgTaskActivityConfigBack, err error) {
	if err = validateTgTaskActivityConfigSet(&req); err != nil {
		return result, err
	}

	var entity pojo.TgTaskActivityConfig
	if req.ID > 0 {
		db.Where("id = ?", req.ID).First(&entity)
		if entity.ID == 0 {
			return result, errors.New("record_not_found_update")
		}
	}

	entity.Title = req.Title
	entity.TitleI18n = normalizeTaskActivityI18nJSON(req.TitleI18n)
	entity.SubTitle = req.SubTitle
	entity.SubTitleI18n = normalizeTaskActivityI18nJSON(req.SubTitleI18n)
	entity.LevelCode = req.LevelCode
	entity.LevelName = req.LevelName
	entity.Sort = req.Sort
	entity.Status = req.Status
	entity.TaskType = req.TaskType
	entity.RequiredInviteCount = req.RequiredInviteCount
	entity.RequiredRechargeAmount = req.RequiredRechargeAmount
	entity.DurationMinutes = req.DurationMinutes
	entity.RewardAmount = req.RewardAmount
	entity.RewardCurrency = req.RewardCurrency
	entity.RewardTarget = req.RewardTarget
	entity.StartAt = req.StartAt
	entity.EndAt = req.EndAt
	entity.Remark = req.Remark

	if entity.ID > 0 {
		err = db.Save(&entity).Error
	} else {
		err = db.Create(&entity).Error
	}
	if err != nil {
		return result, err
	}
	return tgTaskActivityConfigBack(entity), nil
}

// DelTgTaskActivityConfig 删除任务活动配置
func DelTgTaskActivityConfig(db *gorm.DB, id int64) (result string, err error) {
	var entity pojo.TgTaskActivityConfig
	db.Where("id = ?", id).First(&entity)
	if entity.ID == 0 {
		return result, errors.New("record_not_found_delete")
	}
	if err = db.Delete(&entity).Error; err != nil {
		return result, err
	}
	return "success", nil
}

func validateTgTaskActivityConfigSet(req *pojo.TgTaskActivityConfigSet) error {
	req.Title = strings.TrimSpace(req.Title)
	req.SubTitle = strings.TrimSpace(req.SubTitle)
	req.LevelCode = strings.TrimSpace(req.LevelCode)
	req.LevelName = strings.TrimSpace(req.LevelName)
	req.TaskType = strings.TrimSpace(req.TaskType)
	req.RewardCurrency = strings.TrimSpace(req.RewardCurrency)
	req.RewardTarget = strings.TrimSpace(req.RewardTarget)
	req.Remark = strings.TrimSpace(req.Remark)

	if req.Title == "" {
		return errors.New("title_required")
	}
	levelName, ok := tgTaskActivityLevelNames[req.LevelCode]
	if !ok {
		return errors.New("invalid_level_code")
	}
	if req.LevelName == "" {
		req.LevelName = levelName
	}
	if req.TaskType == "" {
		req.TaskType = pojo.TaskActivityTypeInviteRecharge
	}
	if req.TaskType != pojo.TaskActivityTypeInviteRecharge {
		return errors.New("invalid_task_type")
	}
	if req.RequiredInviteCount <= 0 {
		return errors.New("required_invite_count_positive")
	}
	req.RequiredRechargeAmount = utils.Truncate2(req.RequiredRechargeAmount)
	if req.RequiredRechargeAmount <= 0 {
		return errors.New("required_recharge_amount_positive")
	}
	if req.DurationMinutes <= 0 {
		return errors.New("duration_minutes_positive")
	}
	req.RewardAmount = utils.Truncate2(req.RewardAmount)
	if req.RewardAmount <= 0 {
		return errors.New("reward_amount_positive")
	}
	if req.RewardCurrency == "" {
		req.RewardCurrency = "USD"
	}
	if req.RewardTarget == "" {
		req.RewardTarget = pojo.TaskActivityRewardTargetRebate
	}
	if req.RewardTarget != pojo.TaskActivityRewardTargetRebate {
		return errors.New("invalid_reward_target")
	}
	if req.Status != 0 && req.Status != 1 {
		return errors.New("invalid_status")
	}
	if req.StartAt != nil && req.EndAt != nil && !req.EndAt.After(*req.StartAt) {
		return errors.New("activity_end_at_must_after_start_at")
	}
	return nil
}

func tgTaskActivityConfigBack(entity pojo.TgTaskActivityConfig) pojo.TgTaskActivityConfigBack {
	return pojo.TgTaskActivityConfigBack{
		ID:                     entity.ID,
		CreatedAt:              entity.CreatedAt,
		UpdatedAt:              entity.UpdatedAt,
		Title:                  entity.Title,
		TitleI18n:              parseTaskActivityI18nJSON(entity.TitleI18n),
		SubTitle:               entity.SubTitle,
		SubTitleI18n:           parseTaskActivityI18nJSON(entity.SubTitleI18n),
		LevelCode:              entity.LevelCode,
		LevelName:              entity.LevelName,
		Sort:                   entity.Sort,
		Status:                 entity.Status,
		TaskType:               entity.TaskType,
		RequiredInviteCount:    entity.RequiredInviteCount,
		RequiredRechargeAmount: utils.Truncate2(entity.RequiredRechargeAmount),
		DurationMinutes:        entity.DurationMinutes,
		RewardAmount:           utils.Truncate2(entity.RewardAmount),
		RewardCurrency:         entity.RewardCurrency,
		RewardTarget:           entity.RewardTarget,
		StartAt:                entity.StartAt,
		EndAt:                  entity.EndAt,
		Remark:                 entity.Remark,
	}
}

func normalizeTaskActivityI18nJSON(value any) *string {
	if value == nil {
		return nil
	}
	switch typed := value.(type) {
	case string:
		typed = strings.TrimSpace(typed)
		if typed == "" {
			return nil
		}
		var raw map[string]any
		if err := json.Unmarshal([]byte(typed), &raw); err != nil {
			return nil
		}
		normalized := normalizeTaskActivityI18nMap(raw)
		if len(normalized) == 0 {
			return nil
		}
		bytes, _ := json.Marshal(normalized)
		result := string(bytes)
		return &result
	case map[string]string:
		raw := make(map[string]any, len(typed))
		for key, val := range typed {
			raw[key] = val
		}
		normalized := normalizeTaskActivityI18nMap(raw)
		if len(normalized) == 0 {
			return nil
		}
		bytes, _ := json.Marshal(normalized)
		result := string(bytes)
		return &result
	case map[string]any:
		normalized := normalizeTaskActivityI18nMap(typed)
		if len(normalized) == 0 {
			return nil
		}
		bytes, _ := json.Marshal(normalized)
		result := string(bytes)
		return &result
	default:
		bytes, err := json.Marshal(value)
		if err != nil {
			return nil
		}
		var raw map[string]any
		if err = json.Unmarshal(bytes, &raw); err != nil {
			return nil
		}
		normalized := normalizeTaskActivityI18nMap(raw)
		if len(normalized) == 0 {
			return nil
		}
		bytes, _ = json.Marshal(normalized)
		result := string(bytes)
		return &result
	}
}

func normalizeTaskActivityI18nMap(raw map[string]any) map[string]string {
	result := map[string]string{}
	for key, value := range raw {
		lang := strings.TrimSpace(key)
		text, ok := value.(string)
		if !ok {
			continue
		}
		text = strings.TrimSpace(text)
		if lang == "" || text == "" {
			continue
		}
		result[lang] = text
	}
	return result
}

func parseTaskActivityI18nJSON(value *string) any {
	if value == nil || strings.TrimSpace(*value) == "" {
		return nil
	}
	var result map[string]string
	if err := json.Unmarshal([]byte(*value), &result); err != nil {
		return nil
	}
	return result
}

var tgTaskActivityLevelNames = map[string]string{
	pojo.TaskActivityLevelPrimary:  "初级任务",
	pojo.TaskActivityLevelMiddle:   "中级任务",
	pojo.TaskActivityLevelAdvanced: "高级任务",
}
