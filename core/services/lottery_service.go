package services

import (
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/utils"
	"context"
	"math"
	"math/rand"
	"strconv"
	"time"
)

// PopOrRefillLotteryPool pops one award from the Redis batch pool.
// If the pool is empty, it rebuilds one batch from the active probability config.
func PopOrRefillLotteryPool(poolKey string, config pojo.SysTenantPrizePoolConfig, probMap map[float64]int) float64 {
	if config.TotalProbability <= 0 || config.Count <= 0 {
		return 0
	}

	ctx := context.Background()

	val, err := utils.RD.RPop(ctx, poolKey).Result()
	if err == nil {
		amount, _ := strconv.ParseFloat(val, 64)
		return amount
	}

	var pool []string
	for amount, prob := range probMap {
		slotCount := int(math.Round(float64(prob) / float64(config.TotalProbability) * float64(config.Count)))
		for i := 0; i < slotCount; i++ {
			pool = append(pool, strconv.FormatFloat(amount, 'f', -1, 64))
		}
	}
	if len(pool) == 0 {
		return 0
	}

	rand.Shuffle(len(pool), func(i, j int) { pool[i], pool[j] = pool[j], pool[i] })

	args := make([]any, len(pool))
	for i, v := range pool {
		args[i] = v
	}
	utils.RD.RPush(ctx, poolKey, args...)
	utils.RD.Expire(ctx, poolKey, 7*24*time.Hour)

	val, err = utils.RD.RPop(ctx, poolKey).Result()
	if err != nil {
		return 0
	}
	amount, _ := strconv.ParseFloat(val, 64)
	return amount
}
