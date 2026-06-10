package pojo

import "time"

type ExchangeCode struct {
	BaseModel
	Code           string  `json:"code" gorm:"column:code;type:varchar(6);not null;uniqueIndex;comment:6位兑换码"`
	Amount         float64 `json:"amount" gorm:"column:amount;type:decimal(20,2);not null;default:0.00;comment:兑换金额"`
	MaxRedeemCount int     `json:"maxRedeemCount" gorm:"column:max_redeem_count;type:int;not null;default:0;comment:总可兑换次数"`
	RedeemCount    int     `json:"redeemCount" gorm:"column:redeem_count;type:int;not null;default:0;comment:已兑换次数"`
	Status         int8    `json:"status" gorm:"column:status;type:tinyint;not null;default:1;index;comment:状态 1启用 0禁用 -1删除"`
	Remark         string  `json:"remark" gorm:"column:remark;type:varchar(255);comment:备注"`
	CreatedBy      int64   `json:"createdBy" gorm:"column:created_by;type:bigint;not null;default:0;index;comment:创建人ID"`
}

type ExchangeCodeRedeem struct {
	BaseModel
	CodeID          int64   `json:"codeId" gorm:"column:code_id;type:bigint;not null;uniqueIndex:idx_exchange_code_redeem_once,priority:1;index;comment:兑换码ID"`
	Code            string  `json:"code" gorm:"column:code;type:varchar(6);not null;index;comment:兑换码"`
	UserID          int64   `json:"userId" gorm:"column:user_id;type:bigint;not null;uniqueIndex:idx_exchange_code_redeem_once,priority:2;index;comment:用户ID"`
	TenantID        int64   `json:"tenantId" gorm:"column:tenant_id;type:bigint;not null;default:0;index;comment:租户ID"`
	Amount          float64 `json:"amount" gorm:"column:amount;type:decimal(20,2);not null;default:0.00;comment:兑换金额"`
	BeforeBalance   float64 `json:"beforeBalance" gorm:"column:before_balance;type:decimal(20,2);not null;default:0.00;comment:兑换前余额"`
	AfterBalance    float64 `json:"afterBalance" gorm:"column:after_balance;type:decimal(20,2);not null;default:0.00;comment:兑换后余额"`
	CashHistoryID   int64   `json:"cashHistoryId" gorm:"column:cash_history_id;type:bigint;not null;default:0;index;comment:余额流水ID"`
	SourceChannelID *int64  `json:"sourceChannelId" gorm:"column:source_channel_id;type:bigint;index;comment:来源渠道ID"`
}

type ExchangeCodeBack struct {
	ID             int64     `json:"id"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
	Code           string    `json:"code"`
	Amount         float64   `json:"amount"`
	MaxRedeemCount int       `json:"maxRedeemCount"`
	RedeemCount    int       `json:"redeemCount"`
	Status         int8      `json:"status"`
	Remark         string    `json:"remark"`
	CreatedBy      int64     `json:"createdBy"`
}

type ExchangeCodeSearch struct {
	PageInfo
	Code   string `json:"code"`
	Status *int8  `json:"status"`
}

type ExchangeCodeSet struct {
	ID             int64   `json:"id"`
	Code           string  `json:"code"`
	Amount         float64 `json:"amount"`
	MaxRedeemCount int     `json:"maxRedeemCount"`
	GenerateCount  int     `json:"generateCount"` // 随机生成时的数量（仅新增且未指定code时生效，默认1）
	Status         *int8   `json:"status"`
	Remark         string  `json:"remark"`
}

type ExchangeCodePage struct {
	BasePageResponse[ExchangeCodeBack]
}

type ExchangeCodeRedeemReq struct {
	Code string `json:"code" binding:"required"`
}

type ExchangeCodeRedeemBack struct {
	Code       string    `json:"code"`
	Amount     float64   `json:"amount"`
	Balance    float64   `json:"balance"`
	RedeemedAt time.Time `json:"redeemedAt"`
}

func (ExchangeCode) TableName() string {
	return "exchange_code"
}

func (ExchangeCodeRedeem) TableName() string {
	return "exchange_code_redeem"
}
