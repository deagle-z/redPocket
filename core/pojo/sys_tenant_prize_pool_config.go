package pojo

import (
	"strconv"
	"strings"
	"time"
)

// SysTenantPrizePoolConfig 奖池概率配置表（简化版）
type SysTenantPrizePoolConfig struct {
	ID               int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	TenantId         int64     `json:"tenantId" gorm:"type:bigint unsigned;not null"`
	PoolId           int64     `json:"poolId" gorm:"type:bigint unsigned;not null;index:idx_pool_id"`
	Probabilities    string    `json:"probabilities" gorm:"type:varchar(255);not null"` // 概率列表，如 2|10|30|40
	Amounts          string    `json:"amounts" gorm:"type:varchar(255);not null"`       // 金额列表，如 2|10|30|40
	TotalProbability int       `json:"totalProbability" gorm:"type:int;not null;default:100"`
	Count            int       `json:"count" gorm:"type:int;not null;default:1"`                // 每批人数
	PeerAmount       float64   `json:"peerAmount" gorm:"type:decimal(10,2);not null;default:0"` // 单次抽奖需要金额
	Status           int8      `json:"status" gorm:"type:tinyint;not null;default:1"`
	Remark           *string   `json:"remark" gorm:"type:varchar(255)"`
	CreatedAt        time.Time `json:"createdAt" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt        time.Time `json:"updatedAt" gorm:"column:updated_at;autoUpdateTime"`
}

func (SysTenantPrizePoolConfig) TableName() string {
	return "sys_tenant_prize_pool_config"
}

// GetAmountProbMap 返回 {金额 -> 概率} 映射。
func (c *SysTenantPrizePoolConfig) GetAmountProbMap() map[float64]int {
	amounts := strings.Split(c.Amounts, "|")
	probs := strings.Split(c.Probabilities, "|")

	m := make(map[float64]int, len(amounts))
	for i := 0; i < len(amounts) && i < len(probs); i++ {
		amount, err1 := strconv.ParseFloat(strings.TrimSpace(amounts[i]), 64)
		prob, err2 := strconv.Atoi(strings.TrimSpace(probs[i]))
		if err1 != nil || err2 != nil {
			continue
		}
		m[amount] = prob
	}
	return m
}

// GetAmountsSlice 按 Amounts 原始顺序返回金额列表。
func (c *SysTenantPrizePoolConfig) GetAmountsSlice() []float64 {
	parts := strings.Split(c.Amounts, "|")
	amounts := make([]float64, 0, len(parts))
	for _, part := range parts {
		amount, err := strconv.ParseFloat(strings.TrimSpace(part), 64)
		if err != nil {
			continue
		}
		amounts = append(amounts, amount)
	}
	return amounts
}

type SysTenantPrizePoolConfigSearch struct {
	PageInfo
	TenantId int64 `json:"tenantId"`
	PoolId   int64 `json:"poolId"`
	Status   *int8 `json:"status"`
}

type SysTenantPrizePoolConfigResp struct {
	BasePageResponse[SysTenantPrizePoolConfig]
}

type SysTenantPrizePoolConfigSet struct {
	ID               int64   `json:"id"`
	PoolId           int64   `json:"poolId"`
	Probabilities    string  `json:"probabilities"`
	Amounts          string  `json:"amounts"`
	TotalProbability int     `json:"totalProbability"`
	Count            int     `json:"count"`
	PeerAmount       float64 `json:"peerAmount"`
	Status           int8    `json:"status"`
	Remark           *string `json:"remark"`
}
