package repository

import (
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/utils"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
	_ "time/tzdata"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	CheckInRewardsConfigKey = "signin_rewards"
	CheckInTimezone         = "America/New_York"
)

func ParseCheckInRewardsConfig(configValue string) ([]float64, error) {
	configValue = strings.TrimSpace(configValue)
	if configValue == "" {
		return nil, errors.New("checkin_rewards_config_invalid")
	}

	var rewards []float64
	if err := json.Unmarshal([]byte(configValue), &rewards); err != nil || len(rewards) == 0 {
		return nil, errors.New("checkin_rewards_config_invalid")
	}
	for i := range rewards {
		rewards[i] = utils.Truncate2(rewards[i])
		if rewards[i] <= 0 {
			return nil, errors.New("checkin_rewards_config_invalid")
		}
	}
	return rewards, nil
}

func CheckInDateInNewYork(now time.Time) string {
	return now.In(checkInLocation()).Format("2006-01-02")
}

func BuildCheckInAwardUni(userID int64, now time.Time) string {
	return fmt.Sprintf("checkin_%d_%s", userID, now.In(checkInLocation()).Format("20060102"))
}

func GetCurrentCheckInStatus(db *gorm.DB, userID int64, now time.Time) (pojo.TgUserCheckInStatusBack, error) {
	rewards, err := loadCheckInRewards(db)
	if err != nil {
		return pojo.TgUserCheckInStatusBack{}, err
	}

	today := CheckInDateInNewYork(now)
	totalDays, err := countUserCheckInDays(db, userID)
	if err != nil {
		return pojo.TgUserCheckInStatusBack{}, err
	}
	todayChecked, err := hasUserCheckedInDate(db, userID, today)
	if err != nil {
		return pojo.TgUserCheckInStatusBack{}, err
	}

	nextSeq := int(totalDays) + 1
	completed := nextSeq > len(rewards)
	nextRewardAmount := 0.0
	if !completed {
		nextRewardAmount = rewards[nextSeq-1]
	}

	return pojo.TgUserCheckInStatusBack{
		TodayChecked:     todayChecked,
		TotalCheckInDays: totalDays,
		NextSeq:          nextSeq,
		NextRewardAmount: nextRewardAmount,
		Rewards:          rewards,
		Completed:        completed,
		Timezone:         CheckInTimezone,
	}, nil
}

func DoCurrentUserCheckIn(db *gorm.DB, tenantID int64, userID int64, now time.Time) (pojo.TgUserCheckInBack, error) {
	rewards, err := loadCheckInRewards(db)
	if err != nil {
		return pojo.TgUserCheckInBack{}, err
	}

	lockKey := fmt.Sprintf("checkin:%d:%d", tenantID, userID)
	if utils.RD != nil {
		acquired, lockErr := utils.AcquireLock(lockKey, 5*time.Second)
		if lockErr != nil || !acquired {
			return pojo.TgUserCheckInBack{}, errors.New("operation_too_frequent")
		}
		defer func() {
			_ = utils.ReleaseLock(lockKey)
		}()
	}

	today := CheckInDateInNewYork(now)
	var result pojo.TgUserCheckInBack
	err = db.Transaction(func(tx *gorm.DB) error {
		var user pojo.TgUser
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ?", userID).First(&user).Error; err != nil {
			return err
		}
		if user.Status != 1 {
			return errors.New("user_disabled_contact_admin")
		}

		todayChecked, err := hasUserCheckedInDate(tx, userID, today)
		if err != nil {
			return err
		}
		if todayChecked {
			return errors.New("checkin_already_today")
		}

		totalDays, err := countUserCheckInDays(tx, userID)
		if err != nil {
			return err
		}
		checkInSeq := int(totalDays) + 1
		if checkInSeq > len(rewards) {
			return errors.New("checkin_activity_completed")
		}

		rewardAmount := rewards[checkInSeq-1]
		beforeBalance := utils.Truncate2(user.Balance)
		afterBalance := utils.Truncate2(beforeBalance + rewardAmount)

		if err := tx.Model(&pojo.TgUser{}).Where("id = ?", userID).Updates(map[string]any{
			"balance":     gorm.Expr("balance + ?", rewardAmount),
			"gift_amount": gorm.Expr("gift_amount + ?", rewardAmount),
			"gift_total":  gorm.Expr("gift_total + ?", rewardAmount),
		}).Error; err != nil {
			return err
		}
		if err := AddUserWithdrawRestrictedBalance(tx, user, rewardAmount, 0); err != nil {
			return err
		}

		record := pojo.TgUserCheckInRecord{
			TenantId:      tenantID,
			UserId:        userID,
			CheckInDate:   today,
			CheckInSeq:    checkInSeq,
			RewardAmount:  rewardAmount,
			BeforeBalance: beforeBalance,
			AfterBalance:  afterBalance,
		}
		if err := tx.Create(&record).Error; err != nil {
			return err
		}

		history := pojo.CashHistory{
			UserId:      userID,
			AwardUni:    BuildCheckInAwardUni(userID, now),
			Amount:      rewardAmount,
			StartAmount: beforeBalance,
			EndAmount:   afterBalance,
			CashMark:    "签到活动赠送",
			CashDesc:    fmt.Sprintf("累计签到第%d天赠送%.2f金币", checkInSeq, rewardAmount),
			Type:        pojo.CashHistoryTypeCheckInGift,
			IsGift:      1,
			FromUserId:  0,
		}
		if err := tx.Create(&history).Error; err != nil {
			return err
		}

		result = pojo.TgUserCheckInBack{
			RecordId:     record.ID,
			CheckInSeq:   checkInSeq,
			RewardAmount: rewardAmount,
			Balance:      afterBalance,
			TodayChecked: true,
			CheckInDate:  today,
		}
		return nil
	})
	if err != nil {
		return pojo.TgUserCheckInBack{}, err
	}
	return result, nil
}

func GetCurrentUserCheckInRecords(db *gorm.DB, userID int64, limit int) ([]pojo.TgUserCheckInRecordBack, error) {
	if limit <= 0 {
		limit = 30
	}
	if limit > 100 {
		limit = 100
	}

	var records []pojo.TgUserCheckInRecord
	if err := db.Where("user_id = ?", userID).
		Order("check_in_date desc, id desc").
		Limit(limit).
		Find(&records).Error; err != nil {
		return nil, err
	}

	result := make([]pojo.TgUserCheckInRecordBack, 0, len(records))
	for _, record := range records {
		result = append(result, pojo.TgUserCheckInRecordBack{
			ID:            record.ID,
			TenantId:      record.TenantId,
			UserId:        record.UserId,
			CheckInDate:   record.CheckInDate,
			CheckInSeq:    record.CheckInSeq,
			RewardAmount:  utils.Truncate2(record.RewardAmount),
			BeforeBalance: utils.Truncate2(record.BeforeBalance),
			AfterBalance:  utils.Truncate2(record.AfterBalance),
			CreatedAt:     record.CreatedAt,
		})
	}
	return result, nil
}

func GetCheckInRecordsAdmin(db *gorm.DB, search pojo.TgUserCheckInRecordSearch) (pojo.TgUserCheckInRecordResp, error) {
	var result pojo.TgUserCheckInRecordResp
	query := db.Model(&pojo.TgUserCheckInRecord{})

	if search.TenantId > 0 {
		query = query.Where("tenant_id = ?", search.TenantId)
	}
	if search.UserId > 0 {
		query = query.Where("user_id = ?", search.UserId)
	}
	if userUid := strings.TrimSpace(search.UserUid); userUid != "" {
		query = query.Where("user_id IN (?)", db.Model(&pojo.TgUser{}).
			Select("id").
			Where("uid = ?", userUid))
	}
	if strings.TrimSpace(search.StartDate) != "" {
		query = query.Where("check_in_date >= ?", strings.TrimSpace(search.StartDate))
	}
	if strings.TrimSpace(search.EndDate) != "" {
		query = query.Where("check_in_date <= ?", strings.TrimSpace(search.EndDate))
	}

	if err := query.Count(&result.Total).Error; err != nil {
		return result, err
	}

	var records []pojo.TgUserCheckInRecordBack
	if err := query.Select(`
			id,
			tenant_id,
			user_id,
			DATE_FORMAT(check_in_date, '%Y-%m-%d') AS check_in_date,
			check_in_seq,
			reward_amount,
			before_balance,
			after_balance,
			created_at`).
		Order("id desc").
		Limit(search.PageSize).
		Offset(search.PageSize * search.CurrentPage).
		Scan(&records).Error; err != nil {
		return result, err
	}

	userUIDs, err := getCheckInRecordUserUIDs(db, records)
	if err != nil {
		return result, err
	}

	result.List = make([]pojo.TgUserCheckInRecordBack, 0, len(records))
	for _, record := range records {
		result.List = append(result.List, pojo.TgUserCheckInRecordBack{
			ID:            record.ID,
			TenantId:      record.TenantId,
			UserId:        record.UserId,
			UserUid:       userUIDs[record.UserId],
			CheckInDate:   record.CheckInDate,
			CheckInSeq:    record.CheckInSeq,
			RewardAmount:  utils.Truncate2(record.RewardAmount),
			BeforeBalance: utils.Truncate2(record.BeforeBalance),
			AfterBalance:  utils.Truncate2(record.AfterBalance),
			CreatedAt:     record.CreatedAt,
		})
	}
	result.PageSize = search.PageSize
	result.CurrentPage = search.CurrentPage
	return result, nil
}

func getCheckInRecordUserUIDs(db *gorm.DB, records []pojo.TgUserCheckInRecordBack) (map[int64]string, error) {
	userIDs := make([]int64, 0, len(records))
	seen := make(map[int64]struct{}, len(records))
	for _, record := range records {
		if _, ok := seen[record.UserId]; ok {
			continue
		}
		seen[record.UserId] = struct{}{}
		userIDs = append(userIDs, record.UserId)
	}
	if len(userIDs) == 0 {
		return map[int64]string{}, nil
	}

	var users []pojo.TgUser
	if err := db.Select("id, uid").Where("id IN ?", userIDs).Find(&users).Error; err != nil {
		return nil, err
	}

	result := make(map[int64]string, len(users))
	for _, user := range users {
		result[user.ID] = user.Uid
	}
	return result, nil
}

func loadCheckInRewards(db *gorm.DB) ([]float64, error) {
	var config pojo.SysConfig
	if err := db.Where("config_key = ?", CheckInRewardsConfigKey).First(&config).Error; err != nil {
		return nil, errors.New("checkin_rewards_config_invalid")
	}
	return ParseCheckInRewardsConfig(config.ConfigValue)
}

func countUserCheckInDays(db *gorm.DB, userID int64) (int64, error) {
	var total int64
	err := db.Model(&pojo.TgUserCheckInRecord{}).Where("user_id = ?", userID).Count(&total).Error
	return total, err
}

func hasUserCheckedInDate(db *gorm.DB, userID int64, checkInDate string) (bool, error) {
	var count int64
	err := db.Model(&pojo.TgUserCheckInRecord{}).
		Where("user_id = ? AND check_in_date = ?", userID, checkInDate).
		Count(&count).Error
	return count > 0, err
}

func checkInLocation() *time.Location {
	loc, err := time.LoadLocation(CheckInTimezone)
	if err != nil {
		return time.UTC
	}
	return loc
}
