package repository

import (
	"BaseGoUni/core/pojo"
	"errors"
	"fmt"
	"github.com/jinzhu/copier"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"log"
	"time"
)

// GetSysVipLevels VIP等级列表（分页）
func GetSysVipLevels(db *gorm.DB, search pojo.SysVipLevelSearch) (result pojo.SysVipLevelResp) {
	var list []pojo.SysVipLevel
	query := db.Model(&pojo.SysVipLevel{})

	if search.TenantID > 0 {
		query = query.Where("tenant_id = ?", search.TenantID)
	}
	if search.Status != nil {
		query = query.Where("status = ?", *search.Status)
	}

	query.Count(&result.Total)
	query = query.Order("sort asc, level asc").Limit(search.PageSize).Offset(search.PageSize * search.CurrentPage)
	query.Find(&list)

	for _, item := range list {
		var temp pojo.SysVipLevelBack
		_ = copier.Copy(&temp, &item)
		result.List = append(result.List, temp)
	}

	result.PageSize = search.PageSize
	result.CurrentPage = search.CurrentPage
	return result
}

// SetSysVipLevel 创建或更新VIP等级
func SetSysVipLevel(db *gorm.DB, req pojo.SysVipLevelSet) (result pojo.SysVipLevelBack, err error) {
	var entity pojo.SysVipLevel
	if req.ID > 0 {
		db.Where("id = ?", req.ID).First(&entity)
		if entity.ID == 0 {
			return result, errors.New("record_not_found_update")
		}
		_ = copier.Copy(&entity, &req)
		err = db.Save(&entity).Error
	} else {
		_ = copier.Copy(&entity, &req)
		err = db.Create(&entity).Error
	}
	if err != nil {
		return result, err
	}
	_ = copier.Copy(&result, &entity)
	return result, nil
}

// DelSysVipLevel 删除VIP等级
func DelSysVipLevel(db *gorm.DB, id int64) (result string, err error) {
	var entity pojo.SysVipLevel
	db.Where("id = ?", id).First(&entity)
	if entity.ID == 0 {
		return result, errors.New("record_not_found_delete")
	}
	err = db.Delete(&entity).Error
	if err != nil {
		return result, err
	}
	return "success", nil
}

// GetSysVipLevelById 根据ID获取VIP等级
func GetSysVipLevelById(db *gorm.DB, id int64) (result pojo.SysVipLevelBack, err error) {
	var entity pojo.SysVipLevel
	db.Where("id = ?", id).First(&entity)
	if entity.ID == 0 {
		return result, errors.New("record_not_found")
	}
	_ = copier.Copy(&result, &entity)
	return result, nil
}

// GetAppVipProgress 返回用户当前VIP进度信息（上一/当前/下一等级、进度百分比、下一等级奖励）
func GetAppVipProgress(db *gorm.DB, userID int64) (pojo.AppVipProgressBack, error) {
	var result pojo.AppVipProgressBack

	var user pojo.TgUser
	if err := db.Where("id = ?", userID).First(&user).Error; err != nil || user.ID == 0 {
		return result, errors.New("user_not_found")
	}

	var levels []pojo.SysVipLevel
	db.Where("tenant_id = ? AND status = 1", user.TenantId).Order("level asc").Find(&levels)

	// 当月充值金额
	now := time.Now()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	var monthRechargeAmount float64
	db.Model(&pojo.RechargeOrder{}).
		Select("COALESCE(SUM(amount), 0)").
		Where("user_id = ? AND status = 1 AND pay_time >= ?", userID, monthStart).
		Scan(&monthRechargeAmount)

	// 累计下注流水（lucky_history 中 amount + lose_money 之和）
	var totalBetAmount float64
	db.Model(&pojo.LuckyHistory{}).
		Select("COALESCE(SUM(amount + lose_money), 0)").
		Where("user_id = ?", userID).
		Scan(&totalBetAmount)

	currentLevel := 0
	if user.VipLevel != nil {
		currentLevel = *user.VipLevel
	}

	// 找当前/上一/下一等级行
	var prevRow, curRow, nextRow *pojo.SysVipLevel
	for i := range levels {
		lv := &levels[i]
		if lv.Level == currentLevel {
			curRow = lv
		} else if lv.Level < currentLevel {
			if prevRow == nil || lv.Level > prevRow.Level {
				prevRow = lv
			}
		} else { // lv.Level > currentLevel
			if nextRow == nil || lv.Level < nextRow.Level {
				nextRow = lv
			}
		}
	}

	toSimple := func(lv *pojo.SysVipLevel) *pojo.AppVipLevelSimple {
		if lv == nil {
			return nil
		}
		return &pojo.AppVipLevelSimple{
			Level:              lv.Level,
			LevelName:          lv.LevelName,
			UpgradeBonusAmount: lv.UpgradeBonusAmount,
		}
	}
	result.CurrentLevel = toSimple(curRow)
	result.PrevLevel = toSimple(prevRow)
	result.NextLevel = toSimple(nextRow)

	// 计算进度
	if nextRow == nil {
		result.Progress = 100
		result.CurrentValue = user.RechargeAmount
		result.TargetValue = 0
		result.NextBonusAmount = 0
		return result, nil
	}

	result.NextBonusAmount = nextRow.UpgradeBonusAmount

	upgradeType := 1
	if nextRow.UpgradeType != nil {
		upgradeType = int(*nextRow.UpgradeType)
	}

	if upgradeType == 2 {
		// 当月模式：仅看充值金额
		result.CurrentValue = monthRechargeAmount
		if nextRow.MonthRechargeAmount != nil {
			result.TargetValue = *nextRow.MonthRechargeAmount
		}
		if result.TargetValue > 0 {
			p := result.CurrentValue / result.TargetValue * 100
			if p > 100 {
				p = 100
			}
			result.Progress = p
		}
	} else {
		// 累计模式：充值金额和下注流水各占50%，两项进度的均值
		var rechargeTarget, betTarget float64
		if nextRow.TotalRechargeAmount != nil {
			rechargeTarget = *nextRow.TotalRechargeAmount
		}
		if nextRow.TotalValidBet != nil {
			betTarget = *nextRow.TotalValidBet
		}

		var rechargeProgress, betProgress float64
		if rechargeTarget > 0 {
			rechargeProgress = user.RechargeAmount / rechargeTarget * 100
			if rechargeProgress > 100 {
				rechargeProgress = 100
			}
		} else {
			rechargeProgress = 100
		}
		if betTarget > 0 {
			betProgress = totalBetAmount / betTarget * 100
			if betProgress > 100 {
				betProgress = 100
			}
		} else {
			betProgress = 100
		}

		result.Progress = (rechargeProgress + betProgress) / 2
		// CurrentValue / TargetValue 展示充值维度（主要指标）
		result.CurrentValue = user.RechargeAmount
		result.TargetValue = rechargeTarget
	}

	return result, nil
}

// CheckAndUpgradeVipLevel 检查用户是否达到新的VIP等级，如达到则升级并发放奖励。
// 应在充值成功后异步（goroutine）调用，传入带表前缀的 db。
func CheckAndUpgradeVipLevel(db *gorm.DB, userID int64) {
	// 1. 读取用户
	var user pojo.TgUser
	if err := db.Where("id = ?", userID).First(&user).Error; err != nil || user.ID == 0 {
		return
	}

	// 2. 读取该租户全部启用的 VIP 等级（按 level 升序）
	var levels []pojo.SysVipLevel
	db.Where("tenant_id = ? AND status = 1", user.TenantId).Order("level asc").Find(&levels)
	if len(levels) == 0 {
		return
	}

	// 3. 统计累计充值次数（status=1 的成功订单）
	var totalRechargeCount int64
	db.Model(&pojo.RechargeOrder{}).Where("user_id = ? AND status = 1", userID).Count(&totalRechargeCount)

	// 4. 统计当月充值金额
	now := time.Now()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	var monthRechargeAmount float64
	db.Model(&pojo.RechargeOrder{}).
		Select("COALESCE(SUM(amount), 0)").
		Where("user_id = ? AND status = 1 AND pay_time >= ?", userID, monthStart).
		Scan(&monthRechargeAmount)

	// 5. 统计累计下注流水（lucky_history 中 amount + lose_money 之和）
	var totalBetAmount float64
	db.Model(&pojo.LuckyHistory{}).
		Select("COALESCE(SUM(amount + lose_money), 0)").
		Where("user_id = ?", userID).
		Scan(&totalBetAmount)

	// 6. 找出用户能达到的最高等级
	currentLevel := 0
	if user.VipLevel != nil {
		currentLevel = *user.VipLevel
	}
	targetLevel := currentLevel
	var targetRow *pojo.SysVipLevel
	for i := range levels {
		lv := &levels[i]
		if vipLevelQualifies(lv, user.RechargeAmount, int(totalRechargeCount), monthRechargeAmount, totalBetAmount) {
			if lv.Level > targetLevel {
				targetLevel = lv.Level
				targetRow = lv
			}
		}
	}
	if targetRow == nil {
		return // 未达到任何新等级
	}

	// 6. 升级（行锁保证并发安全）
	err := db.Transaction(func(tx *gorm.DB) error {
		var u pojo.TgUser
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", userID).First(&u).Error; err != nil {
			return err
		}
		cur := 0
		if u.VipLevel != nil {
			cur = *u.VipLevel
		}
		if targetLevel <= cur {
			return nil // 已被其他并发流程升过了
		}

		if err := tx.Model(&pojo.TgUser{}).Where("id = ?", userID).Updates(map[string]any{
			"vip_level":      targetLevel,
			"vip_level_name": targetRow.LevelName,
		}).Error; err != nil {
			return err
		}

		// 插入待领取奖励记录（唯一索引防重）
		if targetRow.UpgradeBonusAmount > 0 {
			rewardLog := pojo.SysVipRewardLog{
				TenantID:    user.TenantId,
				UserID:      userID,
				VipLevel:    targetLevel,
				LevelName:   targetRow.LevelName,
				RewardType:  pojo.VipRewardTypeUpgrade,
				BonusAmount: targetRow.UpgradeBonusAmount,
				Status:      pojo.VipRewardStatusPending,
			}
			if err := tx.Create(&rewardLog).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		log.Printf("[VIP] CheckAndUpgradeVipLevel user=%d err=%v", userID, err)
	}
}

// GetClaimableVipRewards 查询当前用户待领取的VIP奖励列表
func GetClaimableVipRewards(db *gorm.DB, userID int64) ([]pojo.SysVipRewardLogBack, error) {
	var list []pojo.SysVipRewardLog
	if err := db.Where("user_id = ? AND status = ?", userID, pojo.VipRewardStatusPending).
		Order("vip_level asc").Find(&list).Error; err != nil {
		return nil, err
	}
	var result []pojo.SysVipRewardLogBack
	for _, item := range list {
		var back pojo.SysVipRewardLogBack
		_ = copier.Copy(&back, &item)
		back.CreatedAt = item.CreatedAt.Format("2006-01-02 15:04:05")
		back.UpdatedAt = item.UpdatedAt.Format("2006-01-02 15:04:05")
		result = append(result, back)
	}
	return result, nil
}

// ClaimVipReward 领取指定VIP奖励（id=0 则领取全部待领取奖励）
func ClaimVipReward(db *gorm.DB, userID int64, rewardLogID int64, tablePrefix string) error {
	var logs []pojo.SysVipRewardLog
	query := db.Where("user_id = ? AND status = ?", userID, pojo.VipRewardStatusPending)
	if rewardLogID > 0 {
		query = query.Where("id = ?", rewardLogID)
	}
	if err := query.Find(&logs).Error; err != nil {
		return err
	}
	if len(logs) == 0 {
		return errors.New("vip_reward_unavailable")
	}

	return db.Transaction(func(tx *gorm.DB) error {
		var user pojo.TgUser
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", userID).First(&user).Error; err != nil {
			return err
		}

		totalBonus := 0.0
		for _, item := range logs {
			totalBonus += item.BonusAmount
		}

		// 更新用户余额
		if err := tx.Model(&pojo.TgUser{}).Where("id = ?", userID).Updates(map[string]any{
			"balance":     gorm.Expr("balance + ?", totalBonus),
			"gift_amount": gorm.Expr("gift_amount + ?", totalBonus),
			"gift_total":  gorm.Expr("gift_total + ?", totalBonus),
		}).Error; err != nil {
			return err
		}
		if err := AddUserWithdrawRestrictedBalance(tx, user, totalBonus, 0); err != nil {
			return err
		}

		runningBalance := user.Balance
		for _, item := range logs {
			// 标记已领取
			if err := tx.Model(&pojo.SysVipRewardLog{}).Where("id = ?", item.ID).
				Update("status", pojo.VipRewardStatusDone).Error; err != nil {
				return err
			}

			// 写流水
			history := pojo.CashHistory{
				UserId:      userID,
				AwardUni:    fmt.Sprintf("vip_reward_%d", item.ID),
				Amount:      item.BonusAmount,
				StartAmount: runningBalance,
				EndAmount:   runningBalance + item.BonusAmount,
				CashMark:    "VIP升级奖励",
				CashDesc:    fmt.Sprintf("领取%s升级奖励 %.2f", item.LevelName, item.BonusAmount),
				Type:        pojo.CashHistoryTypeAdminManualAward,
				IsGift:      1,
				FromUserId:  0,
			}
			if err := tx.Create(&history).Error; err != nil {
				return err
			}
			runningBalance += item.BonusAmount
		}
		return nil
	})
}

// vipLevelQualifies 判断用户当前数据是否满足某 VIP 等级的升级条件。
// nil 条件表示"无要求"，直接跳过。
// 累计模式：充值金额（占50%）与下注流水（占50%）必须同时满足。
func vipLevelQualifies(lv *pojo.SysVipLevel, totalRecharge float64, totalCount int, monthRecharge float64, totalBet float64) bool {
	upgradeType := 1
	if lv.UpgradeType != nil {
		upgradeType = int(*lv.UpgradeType)
	}
	if upgradeType == 2 {
		// 当月条件
		if lv.MonthRechargeAmount != nil && monthRecharge < *lv.MonthRechargeAmount {
			return false
		}
	} else {
		// 累计条件：充值次数、充值金额、下注流水必须同时满足
		if lv.TotalRechargeCount != nil && totalCount < *lv.TotalRechargeCount {
			return false
		}
		if lv.TotalRechargeAmount != nil && totalRecharge < *lv.TotalRechargeAmount {
			return false
		}
		if lv.TotalValidBet != nil && totalBet < *lv.TotalValidBet {
			return false
		}
	}
	return true
}
