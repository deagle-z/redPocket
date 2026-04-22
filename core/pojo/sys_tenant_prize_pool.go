package pojo

import "time"

// SysTenantPrizePool 奖池主表
type SysTenantPrizePool struct {
	ID            int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	TenantId      int64     `json:"tenantId" gorm:"type:bigint unsigned;not null;index"`          // 租户ID
	PoolCode      string    `json:"poolCode" gorm:"type:varchar(64);not null"`                   // 奖池编码（如 red_packet / slot / lucky）
	PoolName      string    `json:"poolName" gorm:"type:varchar(128);not null"`                  // 奖池名称
	Currency      string    `json:"currency" gorm:"type:varchar(16);not null;default:USD"`       // 币种
	Balance       float64   `json:"balance" gorm:"type:decimal(18,6);not null;default:0"`        // 当前奖池余额
	FrozenBalance float64   `json:"frozenBalance" gorm:"type:decimal(18,6);not null;default:0"` // 冻结金额（已开奖未结算）
	MinBalance    float64   `json:"minBalance" gorm:"type:decimal(18,6);default:0"`              // 最低安全值（低于触发补池）
	MaxBalance    *float64  `json:"maxBalance" gorm:"type:decimal(18,6);default:null"`           // 最高限制（超过可分红/抽水），NULL=不限
	RtpRate       float64   `json:"rtpRate" gorm:"type:decimal(5,2);default:90.00"`              // 目标RTP（返奖率 %）
	PumpRate      float64   `json:"pumpRate" gorm:"type:decimal(5,2);default:5.00"`              // 抽水比例（进入平台利润 %）
	InjectRate    float64   `json:"injectRate" gorm:"type:decimal(5,2);default:2.00"`            // 注入比例（进入奖池 %）
	Status        int8      `json:"status" gorm:"type:tinyint;not null;default:1"`               // 状态 1启用 0关闭
	Remark        *string   `json:"remark" gorm:"type:varchar(255)"`                             // 备注
	CreatedAt     time.Time `json:"createdAt" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt     time.Time `json:"updatedAt" gorm:"column:updated_at;autoUpdateTime"`
}

func (SysTenantPrizePool) TableName() string {
	return "sys_tenant_prize_pool"
}

// ---- 请求/响应 DTO ----

type SysTenantPrizePoolSearch struct {
	PageInfo
	TenantId int64  `json:"tenantId"`
	PoolCode string `json:"poolCode"`
	Status   *int8  `json:"status"`
}

type SysTenantPrizePoolResp struct {
	BasePageResponse[SysTenantPrizePool]
}

type SysTenantPrizePoolBalanceSet struct {
	PoolCode string   `json:"poolCode"`
	Balance  float64  `json:"balance"`
	Remark   *string  `json:"remark"`
}
