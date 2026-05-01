package pojo

import "time"

// SysTenantPrizePoolRecord 奖池流水表
type SysTenantPrizePoolRecord struct {
	ID             int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	TenantId       int64     `json:"tenantId" gorm:"type:bigint unsigned;not null;index"` // 租户ID
	PoolId         int64     `json:"poolId" gorm:"type:bigint unsigned;not null;index"`   // 奖池ID
	UserId         *int64    `json:"userId" gorm:"type:bigint unsigned;index"`            // 用户ID
	ChangeType     string    `json:"changeType" gorm:"type:varchar(16);not null"`         // 变动类型：in/out
	Amount         float64   `json:"amount" gorm:"type:decimal(18,2);not null"`           // 变动金额
	BeforeBalance  float64   `json:"beforeBalance" gorm:"type:decimal(18,2);not null"`    // 变动前余额
	AfterBalance   float64   `json:"afterBalance" gorm:"type:decimal(18,2);not null"`     // 变动后余额
	ConsumedAmount *float64  `json:"consumedAmount" gorm:"type:decimal(10,2)"`            // 本次消耗金额
	Remark         *string   `json:"remark" gorm:"type:varchar(255)"`                     // 备注
	CreatedAt      time.Time `json:"createdAt" gorm:"column:created_at;autoCreateTime"`
}

func (SysTenantPrizePoolRecord) TableName() string {
	return "sys_tenant_prize_pool_record"
}

// ---- 变动类型常量 ----
const (
	PrizePoolChangeTypeIn  = "in"  // 流入
	PrizePoolChangeTypeOut = "out" // 流出
)

// ---- 请求/响应 DTO ----

type SysTenantPrizePoolRecordSearch struct {
	PageInfo
	PoolId     int64  `json:"poolId"`
	TenantId   int64  `json:"tenantId"`
	ChangeType string `json:"changeType"`
	UserId     int64  `json:"userId"`
	StartTime  int64  `json:"startTime"` // unix 秒
	EndTime    int64  `json:"endTime"`   // unix 秒
}

type SysTenantPrizePoolRecordResp struct {
	BasePageResponse[SysTenantPrizePoolRecord]
}

type SysTenantPrizePoolOutRecordBack struct {
	ID             int64     `json:"id"`
	TenantId       int64     `json:"tenantId"`
	PoolId         int64     `json:"poolId"`
	UserId         *int64    `json:"userId"`
	UserName       string    `json:"userName"`
	ChangeType     string    `json:"changeType"`
	Amount         float64   `json:"amount"`
	BeforeBalance  float64   `json:"beforeBalance"`
	AfterBalance   float64   `json:"afterBalance"`
	ConsumedAmount *float64  `json:"consumedAmount"`
	CreatedAt      time.Time `json:"createdAt"`
}

type SysTenantPrizePoolOutRecordResp struct {
	BasePageResponse[SysTenantPrizePoolOutRecordBack]
}

// LotteryChancesResp 抽奖次数查询响应
type LotteryChancesResp struct {
	TotalFlow      float64   `json:"totalFlow"`      // 累计总流水
	EarnedCount    int64     `json:"earnedCount"`    // 已获得次数 = floor(totalFlow/1000)
	UsedCount      int64     `json:"usedCount"`      // 已使用次数
	AvailableCount int64     `json:"availableCount"` // 可用次数
	Amounts        []float64 `json:"amounts"`        // 奖池金额列表
}
