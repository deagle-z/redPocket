package api

import (
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/repository"
	"BaseGoUni/core/utils"
	"context"
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

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

	totalFlow, err := repository.GetUserTotalFlow(db, userID)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}

	earnedCount := int64(math.Floor(totalFlow / peerAmount))

	usedCount, err := repository.GetUsedLotteryCount(db, userID)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}

	availableCount := earnedCount - usedCount
	if availableCount < 0 {
		availableCount = 0
	}

	// 获取奖池金额列表
	var amounts []float64
	if configErr == nil && config.ID > 0 {
		amounts = config.GetAmountsSlice()
	}

	utils.SuccessObjBack(ctx, pojo.LotteryChancesResp{
		TotalFlow:      totalFlow,
		EarnedCount:    earnedCount,
		UsedCount:      usedCount,
		AvailableCount: availableCount,
		Amounts:        amounts,
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
	peerAmount := config.PeerAmount
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

	// 3. 锁内二次校验可用次数
	totalFlow, err := repository.GetUserTotalFlow(db, userID)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	earnedCount := int64(math.Floor(totalFlow / peerAmount))

	usedCount, err := repository.GetUsedLotteryCount(db, userID)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	if earnedCount-usedCount <= 0 {
		utils.ErrorBack(ctx, "暂无可用抽奖次数")
		return
	}

	// 4. 从 Redis 批次池弹出本次奖励
	probMap := config.GetAmountProbMap()
	poolKey := fmt.Sprintf("lottery_pool:%d:%d", tenantId, config.ID)
	awardAmount := popOrRefillLotteryPool(poolKey, config, probMap)

	// 5. 事务：写记录 + 发奖
	var recordID int64
	txErr := db.Transaction(func(tx *gorm.DB) error {
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
			PeerAmount:  peerAmount,
			AwardAmount: awardAmount,
			Status:      status,
		}

		if awardAmount > 0 {
			// 锁定用户行
			var user pojo.TgUser
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
				Where("id = ?", userID).First(&user).Error; err != nil {
				return err
			}
			record.BeforeBalance = user.Balance
			record.AfterBalance = user.Balance + awardAmount

			// 更新余额 + gift 字段
			if err := tx.Model(&pojo.TgUser{}).Where("id = ?", userID).Updates(map[string]any{
				"balance":     gorm.Expr("balance + ?", awardAmount),
				"gift_amount": gorm.Expr("gift_amount + ?", awardAmount),
				"gift_total":  gorm.Expr("gift_total + ?", awardAmount),
			}).Error; err != nil {
				return err
			}
		}

		if err := tx.Create(&record).Error; err != nil {
			return err
		}
		recordID = record.ID

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

// popOrRefillLotteryPool 从 Redis 批次池弹出一个奖励；池空时按配置重新生成并打乱
func popOrRefillLotteryPool(poolKey string, config pojo.SysTenantPrizePoolConfig, probMap map[float64]int) float64 {
	ctx := context.Background()

	val, err := utils.RD.RPop(ctx, poolKey).Result()
	if err == nil {
		amount, _ := strconv.ParseFloat(val, 64)
		return amount
	}

	// 池已耗尽，生成新一轮
	var pool []string
	for amount, prob := range probMap {
		slotCount := int(math.Round(float64(prob) / float64(config.TotalProbability) * float64(config.Count)))
		for i := 0; i < slotCount; i++ {
			pool = append(pool, strconv.FormatFloat(amount, 'f', -1, 64))
		}
	}

	// Fisher-Yates 打乱
	rand.Shuffle(len(pool), func(i, j int) { pool[i], pool[j] = pool[j], pool[i] })

	// 写入 Redis，保留 7 天
	args := make([]any, len(pool))
	for i, v := range pool {
		args[i] = v
	}
	utils.RD.RPush(ctx, poolKey, args...)
	utils.RD.Expire(ctx, poolKey, 7*24*time.Hour)

	// 弹出第一个
	val, err = utils.RD.RPop(ctx, poolKey).Result()
	if err != nil {
		return 0
	}
	amount, _ := strconv.ParseFloat(val, 64)
	return amount
}
