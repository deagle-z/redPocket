package api

import (
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/repository"
	"BaseGoUni/core/services"
	"BaseGoUni/core/utils"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type lotteryConsumption struct {
	UseFree    bool
	PeerAmount float64
}

type flowLotterySyncResult struct {
	TotalFlow                 float64
	RawTotalFlow              float64
	EarnedCount               int64
	UsedCount                 int64
	FlowLotteryTotalCount     int
	FlowLotteryAvailableCount int
}

func calculateLotteryAvailableCount(flowAvailableCount int, freeCount int) int64 {
	if flowAvailableCount < 0 {
		flowAvailableCount = 0
	}
	if freeCount < 0 {
		freeCount = 0
	}
	return int64(flowAvailableCount + freeCount)
}

func selectLotteryConsumption(freeCount int, flowAvailableCount int, peerAmount float64) (lotteryConsumption, bool) {
	if freeCount > 0 {
		return lotteryConsumption{UseFree: true, PeerAmount: 0}, true
	}
	if flowAvailableCount <= 0 {
		return lotteryConsumption{}, false
	}
	return lotteryConsumption{PeerAmount: peerAmount}, true
}

func syncFlowLotteryCount(tx *gorm.DB, userID int64, peerAmount float64) (pojo.TgUser, flowLotterySyncResult, error) {
	var user pojo.TgUser
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ?", userID).First(&user).Error; err != nil {
		return user, flowLotterySyncResult{}, err
	}

	rawTotalFlow, err := repository.GetUserTotalFlow(tx, userID)
	if err != nil {
		return user, flowLotterySyncResult{}, err
	}

	if peerAmount <= 0 {
		peerAmount = 1000
	}
	totalFlow := utils.Truncate2(rawTotalFlow - user.FlowLotteryBaseFlow)
	if totalFlow < 0 {
		totalFlow = 0
	}
	earnedCount := int64(math.Floor(totalFlow / peerAmount))
	usedCount, err := repository.GetUsedLotteryCountAfter(tx, userID, user.FlowLotteryBaseRecordID)
	if err != nil {
		return user, flowLotterySyncResult{}, err
	}

	oldTotal := user.FlowLotteryTotalCount
	oldAvailable := user.FlowLotteryAvailableCount
	newTotal := int(earnedCount)
	newAvailable := int(earnedCount - usedCount)
	if newAvailable < 0 {
		newAvailable = 0
	}

	if newTotal != oldTotal || newAvailable != oldAvailable {
		if err := tx.Model(&pojo.TgUser{}).Where("id = ?", userID).Updates(map[string]any{
			"flow_lottery_total_count":     newTotal,
			"flow_lottery_available_count": newAvailable,
		}).Error; err != nil {
			return user, flowLotterySyncResult{}, err
		}
		user.FlowLotteryTotalCount = newTotal
		user.FlowLotteryAvailableCount = newAvailable
	}

	return user, flowLotterySyncResult{
		TotalFlow:                 totalFlow,
		RawTotalFlow:              rawTotalFlow,
		EarnedCount:               earnedCount,
		UsedCount:                 usedCount,
		FlowLotteryTotalCount:     user.FlowLotteryTotalCount,
		FlowLotteryAvailableCount: user.FlowLotteryAvailableCount,
	}, nil
}

func calculateFlowLotteryConsumedCount(totalCount int, availableCount int) int {
	if totalCount <= availableCount {
		return 0
	}
	return totalCount - availableCount
}

// GetLotteryChances app端查询抽奖次数及奖池金额列表
func GetLotteryChances(ctx *gin.Context) {
	userID := ctx.MustGet("userId").(int64)
	db := ctx.MustGet("db").(*gorm.DB)

	// 获取配置（用于 peerAmount 阈值）
	config, configErr := repository.GetFirstActivePrizePoolConfig(db)
	peerAmount := 1000.0
	if configErr == nil && config.ID > 0 && config.PeerAmount > 0 {
		peerAmount = config.PeerAmount
	}

	var user pojo.TgUser
	var syncResult flowLotterySyncResult
	if err := db.Transaction(func(tx *gorm.DB) error {
		var syncErr error
		user, syncResult, syncErr = syncFlowLotteryCount(tx, userID, peerAmount)
		return syncErr
	}); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}

	freeCount := user.FreeLotteryCount
	flowAvailableCount := user.FlowLotteryAvailableCount
	availableCount := calculateLotteryAvailableCount(flowAvailableCount, freeCount)
	flowConsumedCount := calculateFlowLotteryConsumedCount(syncResult.FlowLotteryTotalCount, syncResult.FlowLotteryAvailableCount)
	currentFlow := utils.Truncate2(syncResult.TotalFlow - float64(flowConsumedCount)*peerAmount)
	if currentFlow < 0 {
		currentFlow = 0
	}
	remainingFlow := utils.Truncate2(peerAmount - currentFlow)
	if flowAvailableCount > 0 {
		currentFlow = peerAmount
		remainingFlow = 0
	} else if remainingFlow < 0 {
		remainingFlow = 0
	}

	// 获取奖池金额列表
	var amounts []float64
	if configErr == nil && config.ID > 0 {
		amounts = config.GetAmountsSlice()
	}

	utils.SuccessObjBack(ctx, pojo.LotteryChancesResp{
		TotalFlow:                 syncResult.TotalFlow,
		CurrentFlow:               currentFlow,
		PeerAmount:                peerAmount,
		RemainingFlow:             remainingFlow,
		EarnedCount:               syncResult.EarnedCount,
		UsedCount:                 syncResult.UsedCount,
		FreeCount:                 freeCount,
		FlowLotteryTotalCount:     syncResult.FlowLotteryTotalCount,
		FlowLotteryAvailableCount: syncResult.FlowLotteryAvailableCount,
		AvailableCount:            availableCount,
		Amounts:                   amounts,
	})
}

// GetLotteryHistory app端查询抽奖历史（名称+金额）
func GetLotteryHistory(ctx *gin.Context) {
	userID := ctx.MustGet("userId").(int64)
	db := ctx.MustGet("db").(*gorm.DB)

	limitStr := ctx.DefaultQuery("limit", "20")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 20
	}

	items, err := repository.GetLotteryHistory(db, userID, limit)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	if items == nil {
		items = []repository.LotteryHistoryItem{}
	}
	utils.SuccessObjBack(ctx, items)
}

// DrawLottery app端抽奖
// 不扣余额；每积累 peer_amount 流水获得一次机会
func DrawLottery(ctx *gin.Context) {
	userID := ctx.MustGet("userId").(int64)
	tenantId := ctx.MustGet("tenantId").(int64)
	db := ctx.MustGet("db").(*gorm.DB)

	// 1. 获取配置
	config, err := repository.GetFirstActivePrizePoolConfig(db)
	if err != nil || config.ID == 0 {
		utils.ErrorBack(ctx, "抽奖配置未找到")
		return
	}
	peerAmount := utils.Truncate2(config.PeerAmount)
	if peerAmount <= 0 {
		peerAmount = 1000
	}

	// 2. 分布式锁
	lockKey := fmt.Sprintf("lottery_draw:%d", userID)
	acquired, lockErr := utils.AcquireLock(lockKey, 5*time.Second)
	if lockErr != nil || !acquired {
		utils.ErrorBack(ctx, "操作频繁，请稍后再试")
		return
	}
	defer func(lockKey string) {
		err := utils.ReleaseLock(lockKey)
		if err != nil {

		}
	}(lockKey)

	// 3. 准备 Redis 批次池，本次奖励在事务内确认有次数后再弹出
	probMap := config.GetAmountProbMap()
	poolKey := fmt.Sprintf("lottery_pool:%d:%d", tenantId, config.ID)

	// 4. 事务：同步流水次数 + 扣次数 + 写记录 + 发奖
	var recordID int64
	var awardAmount float64
	txErr := db.Transaction(func(tx *gorm.DB) error {
		user, _, err := syncFlowLotteryCount(tx, userID, peerAmount)
		if err != nil {
			return err
		}
		consumption, ok := selectLotteryConsumption(user.FreeLotteryCount, user.FlowLotteryAvailableCount, peerAmount)
		if !ok {
			return fmt.Errorf("暂无可用抽奖次数")
		}

		if consumption.UseFree {
			if user.FreeLotteryCount <= 0 {
				return fmt.Errorf("暂无可用抽奖次数")
			}
			result := tx.Model(&pojo.TgUser{}).
				Where("id = ? AND free_lottery_count > 0", userID).
				Update("free_lottery_count", gorm.Expr("free_lottery_count - 1"))
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected == 0 {
				return fmt.Errorf("暂无可用抽奖次数")
			}
		} else {
			result := tx.Model(&pojo.TgUser{}).
				Where("id = ? AND flow_lottery_available_count > 0", userID).
				Update("flow_lottery_available_count", gorm.Expr("flow_lottery_available_count - 1"))
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected == 0 {
				return fmt.Errorf("暂无可用抽奖次数")
			}
		}

		awardAmount = utils.Truncate2(services.PopOrRefillLotteryPool(poolKey, config, probMap))

		// 写抽奖记录
		status := pojo.LotteryStatusNoAward
		if awardAmount > 0 {
			status = pojo.LotteryStatusAwarded
		}

		record := pojo.UserLotteryRecord{
			TenantId:    tenantId,
			UserId:      userID,
			PoolId:      config.PoolId,
			ConfigId:    config.ID,
			PeerAmount:  consumption.PeerAmount,
			AwardAmount: awardAmount,
			Status:      status,
		}

		if awardAmount > 0 {
			record.BeforeBalance = user.Balance
			record.AfterBalance = utils.Truncate2(user.Balance + awardAmount)

			// 更新余额 + gift 字段
			if err := tx.Model(&pojo.TgUser{}).Where("id = ?", userID).Updates(map[string]any{
				"balance":     gorm.Expr("balance + ?", awardAmount),
				"gift_amount": gorm.Expr("gift_amount + ?", awardAmount),
				"gift_total":  gorm.Expr("gift_total + ?", awardAmount),
			}).Error; err != nil {
				return err
			}
			if err := repository.AddUserWithdrawRestrictedBalance(tx, user, awardAmount, 0); err != nil {
				return err
			}
		}

		if err := tx.Create(&record).Error; err != nil {
			return err
		}
		recordID = record.ID

		poolRecordRemark := fmt.Sprintf("lottery_draw_%d", record.ID)
		if err := repository.CreateLotteryDrawRecord(tx, tenantId, config.PoolId, userID, consumption.PeerAmount, &poolRecordRemark); err != nil {
			return err
		}

		// 写中奖流水
		if awardAmount > 0 {
			cashHistory := pojo.CashHistory{
				UserId:      userID,
				AwardUni:    fmt.Sprintf("lottery_award_%d", record.ID),
				Amount:      awardAmount,
				StartAmount: record.BeforeBalance,
				EndAmount:   record.AfterBalance,
				CashMark:    "抽奖中奖",
				CashDesc:    fmt.Sprintf("抽奖中奖 %.2f", awardAmount),
				Type:        pojo.CashHistoryTypeLotteryAward,
				IsGift:      1,
			}
			if err := tx.Create(&cashHistory).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if txErr != nil {
		utils.ErrorBack(ctx, txErr.Error())
		return
	}

	utils.SuccessObjBack(ctx, gin.H{
		"recordId":    recordID,
		"awardAmount": awardAmount,
	})
}
