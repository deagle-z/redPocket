package services

import (
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/utils"
	"context"
	"errors"
	"log"
	"strconv"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	trialLuckyFlowLotteryRewardConfig = "trial_lucky_flow_lottery_reward"
)

func parseTrialLuckyFlowLotteryRewardConfig(raw string) (float64, int, bool) {
	parts := strings.Split(strings.TrimSpace(raw), ":")
	if len(parts) != 2 {
		return 0, 0, false
	}
	threshold, err := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
	if err != nil || threshold <= 0 {
		return 0, 0, false
	}
	rewardCount, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil || rewardCount <= 0 {
		return 0, 0, false
	}
	return threshold, rewardCount, true
}

func resolveTrialLuckyFlowLotteryRewardConfig(cacheValue string, dbValue string) (float64, int, bool) {
	if threshold, rewardCount, enabled := parseTrialLuckyFlowLotteryRewardConfig(dbValue); enabled {
		return threshold, rewardCount, true
	}
	return parseTrialLuckyFlowLotteryRewardConfig(cacheValue)
}

func calculateTrialLuckyFlowLotteryReward(totalFlow float64, threshold float64, rewardCount int, alreadySent bool) (int, bool) {
	if alreadySent || threshold <= 0 || rewardCount <= 0 || totalFlow < threshold {
		return 0, false
	}
	return rewardCount, true
}

func awardTrialLuckyFlowLotteryIfNeeded(tx *gorm.DB, userID int64, tenantID int64, luckyID int64, historyID int64, tablePrefix string) (int, bool, error) {
	threshold, rewardCount, enabled := getTrialLuckyFlowLotteryRewardConfig(tx, tablePrefix)
	if !enabled {
		return 0, false, nil
	}

	var existingCount int64
	if err := tx.Model(&pojo.TrialLuckyFlowLotteryReward{}).
		Where("tenant_id = ? AND user_id = ?", tenantID, userID).
		Count(&existingCount).Error; err != nil {
		return 0, false, err
	}

	var totalFlow float64
	if err := tx.Model(&pojo.TrialLuckyHistory{}).
		Select("COALESCE(SUM(amount + lose_money), 0)").
		Where("tenant_id = ? AND user_id = ? AND actor_type = ? AND grab_type = ?", tenantID, userID, pojo.TrialActorUser, 1).
		Scan(&totalFlow).Error; err != nil {
		return 0, false, err
	}

	awardCount, awarded := calculateTrialLuckyFlowLotteryReward(totalFlow, threshold, rewardCount, existingCount > 0)
	if !awarded {
		log.Printf("[trial-lottery-reward] skipped user_id=%d tenant_id=%d lucky_id=%d history_id=%d total_flow=%.2f threshold=%.2f existing_count=%d",
			userID, tenantID, luckyID, historyID, totalFlow, threshold, existingCount)
		return 0, false, nil
	}

	record := pojo.TrialLuckyFlowLotteryReward{
		TenantId:          tenantID,
		UserID:            userID,
		ThresholdAmount:   threshold,
		RewardCount:       awardCount,
		TotalFlowSnapshot: utils.Truncate2(totalFlow),
		SourceLuckyID:     luckyID,
		SourceHistoryID:   historyID,
	}
	result := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&record)
	if result.Error != nil {
		return 0, false, result.Error
	}
	if result.RowsAffected == 0 {
		log.Printf("[trial-lottery-reward] duplicate skipped user_id=%d tenant_id=%d lucky_id=%d history_id=%d total_flow=%.2f threshold=%.2f reward_count=%d",
			userID, tenantID, luckyID, historyID, totalFlow, threshold, awardCount)
		return 0, false, nil
	}

	if err := tx.Model(&pojo.TgUser{}).
		Where("id = ?", userID).
		Update("free_lottery_count", gorm.Expr("free_lottery_count + ?", awardCount)).Error; err != nil {
		return 0, false, err
	}

	log.Printf("[trial-lottery-reward] awarded user_id=%d tenant_id=%d lucky_id=%d history_id=%d total_flow=%.2f threshold=%.2f reward_count=%d",
		userID, tenantID, luckyID, historyID, totalFlow, threshold, awardCount)
	return awardCount, true, nil
}

func GetTrialLuckyFlowLotteryRewardProgress(db *gorm.DB, userID int64, tablePrefix string) (pojo.TrialLuckyFlowLotteryRewardProgress, error) {
	threshold, rewardCount, enabled := getTrialLuckyFlowLotteryRewardConfig(db, tablePrefix)

	var user pojo.TgUser
	if err := db.Select("id", "tenant_id", "free_lottery_count").Where("id = ?", userID).First(&user).Error; err != nil {
		return pojo.TrialLuckyFlowLotteryRewardProgress{}, err
	}
	tenantID := user.TenantId

	var totalFlow float64
	if err := db.Model(&pojo.TrialLuckyHistory{}).
		Select("COALESCE(SUM(amount + lose_money), 0)").
		Where("tenant_id = ? AND user_id = ? AND actor_type = ? AND grab_type = ?", tenantID, userID, pojo.TrialActorUser, 1).
		Scan(&totalFlow).Error; err != nil {
		return pojo.TrialLuckyFlowLotteryRewardProgress{}, err
	}

	var rewardCountRows int64
	if err := db.Model(&pojo.TrialLuckyFlowLotteryReward{}).
		Where("tenant_id = ? AND user_id = ?", tenantID, userID).
		Count(&rewardCountRows).Error; err != nil {
		return pojo.TrialLuckyFlowLotteryRewardProgress{}, err
	}

	var drawCount int64
	if err := db.Model(&pojo.UserLotteryRecord{}).
		Where("tenant_id = ? AND user_id = ?", tenantID, userID).
		Count(&drawCount).Error; err != nil {
		return pojo.TrialLuckyFlowLotteryRewardProgress{}, err
	}

	progress := buildTrialLuckyFlowLotteryRewardProgress(totalFlow, threshold, rewardCount, rewardCountRows > 0)
	progress.Drawn = drawCount > 0
	progress.FreeLotteryCount = user.FreeLotteryCount
	progress.Enabled = enabled
	return progress, nil
}

func buildTrialLuckyFlowLotteryRewardProgress(totalFlow float64, threshold float64, rewardCount int, rewarded bool) pojo.TrialLuckyFlowLotteryRewardProgress {
	progress := pojo.TrialLuckyFlowLotteryRewardProgress{
		Enabled:         threshold > 0 && rewardCount > 0,
		ThresholdAmount: utils.Truncate2(threshold),
		RewardCount:     rewardCount,
		TotalFlow:       utils.Truncate2(totalFlow),
		Rewarded:        rewarded,
	}
	if !progress.Enabled {
		return progress
	}

	remaining := threshold - totalFlow
	if remaining < 0 {
		remaining = 0
	}
	percent := totalFlow / threshold * 100
	if percent > 100 {
		percent = 100
	}
	progress.RemainingFlow = utils.Truncate2(remaining)
	progress.ProgressPercent = utils.Truncate2(percent)
	progress.AvailableRewardCount, progress.CanReward = calculateTrialLuckyFlowLotteryReward(totalFlow, threshold, rewardCount, rewarded)
	return progress
}

func getTrialLuckyFlowLotteryRewardConfig(tx *gorm.DB, tablePrefix string) (float64, int, bool) {
	cacheValue := ""
	if tablePrefix != "" && utils.RD != nil {
		if value, err := getTrialLuckyFlowLotteryRewardCacheValue(tablePrefix); err == nil {
			cacheValue = value
		} else {
			log.Printf("[trial-lottery-reward] config cache unavailable prefix=%q key=%s err=%v", tablePrefix, trialLuckyFlowLotteryRewardConfig, err)
		}
	}

	dbValue := ""
	var config pojo.SysConfig
	if err := tx.Where("config_key = ?", trialLuckyFlowLotteryRewardConfig).First(&config).Error; err == nil {
		dbValue = config.ConfigValue
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Printf("[trial-lottery-reward] config db query failed key=%s err=%v", trialLuckyFlowLotteryRewardConfig, err)
	}

	threshold, rewardCount, enabled := resolveTrialLuckyFlowLotteryRewardConfig(cacheValue, dbValue)
	if enabled && strings.TrimSpace(cacheValue) != strings.TrimSpace(dbValue) && tablePrefix != "" && utils.RD != nil {
		setTrialLuckyFlowLotteryRewardCacheValue(tablePrefix, dbValue)
	}
	return threshold, rewardCount, enabled
}

func getTrialLuckyFlowLotteryRewardCacheValue(tablePrefix string) (string, error) {
	redisKey := "bgu_ct_" + tablePrefix + "_" + trialLuckyFlowLotteryRewardConfig
	data := utils.RD.Get(context.Background(), redisKey)
	if data == nil {
		return "", errors.New("redis_nil_command")
	}
	if err := data.Err(); err != nil {
		return "", err
	}
	return data.Val(), nil
}

func setTrialLuckyFlowLotteryRewardCacheValue(tablePrefix string, value string) {
	if strings.TrimSpace(value) == "" {
		return
	}
	redisKey := "bgu_ct_" + tablePrefix + "_" + trialLuckyFlowLotteryRewardConfig
	if err := utils.RD.SetEX(context.Background(), redisKey, value, utils.GetRandomRangeSecond(20*60, 40*60)).Err(); err != nil {
		log.Printf("[trial-lottery-reward] config cache set failed prefix=%q key=%s err=%v", tablePrefix, trialLuckyFlowLotteryRewardConfig, err)
	}
}
