package pojo

import "time"

// SysTenantPrizePoolRecord 奖池流水表
type SysTenantPrizePoolRecord struct {
	ID            int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	TenantId      int64     `json:"tenantId" gorm:"type:bigint unsigned;not null;index"` // 租户ID
	PoolId        int64     `json:"poolId" gorm:"type:bigint unsigned;not null;index"`   // 奖池ID
	BizType       string    `json:"bizType" gorm:"type:varchar(32);not null"`            // 业务类型：bet/win/lose/inject/withdraw/system
	BizId         *string   `json:"bizId" gorm:"type:varchar(64);index"`                 // 业务ID（订单ID/红包ID等）
	UserId        *int64    `json:"userId" gorm:"type:bigint unsigned;index"`            // 用户ID
	ChangeType    string    `json:"changeType" gorm:"type:varchar(16);not null"`         // 变动类型：in/out
	Amount        float64   `json:"amount" gorm:"type:decimal(18,6);not null"`           // 变动金额
	BeforeBalance float64   `json:"beforeBalance" gorm:"type:decimal(18,6);not null"`    // 变动前余额
	AfterBalance  float64   `json:"afterBalance" gorm:"type:decimal(18,6);not null"`     // 变动后余额
	Remark        *string   `json:"remark" gorm:"type:varchar(255)"`                     // 备注
	CreatedAt     time.Time `json:"createdAt" gorm:"column:created_at;autoCreateTime"`
}

func (SysTenantPrizePoolRecord) TableName() string {
	return "sys_tenant_prize_pool_record"
}

// ---- 业务类型常量 ----
const (
	PrizePoolBizTypeBet      = "bet"      // 投注
	PrizePoolBizTypeWin      = "win"      // 中奖
	PrizePoolBizTypeLose     = "lose"     // 亏损
	PrizePoolBizTypeInject   = "inject"   // 注入
	PrizePoolBizTypeWithdraw = "withdraw" // 提取
	PrizePoolBizTypeSystem   = "system"   // 系统调整
)

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
	BizType    string `json:"bizType"`
	ChangeType string `json:"changeType"`
	UserId     int64  `json:"userId"`
	StartTime  int64  `json:"startTime"` // unix 秒
	EndTime    int64  `json:"endTime"`   // unix 秒
}

type SysTenantPrizePoolRecordResp struct {
	BasePageResponse[SysTenantPrizePoolRecord]
}
