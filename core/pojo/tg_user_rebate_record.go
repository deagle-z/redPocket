package pojo

import "time"

// TgUserRebateRecord 下级给上级反水/返佣记录表
type TgUserRebateRecord struct {
	BaseModel
	TenantId       *int64     `json:"tenantId" gorm:"column:tenant_id;type:bigint"`                   // 租户ID（多租户可用；单租户可删除）
	SubUserId      int64      `json:"subUserId" gorm:"column:sub_user_id;type:bigint;not null"`       // 下级用户ID（tg_user.id）
	ParentUserId   int64      `json:"parentUserId" gorm:"column:parent_user_id;type:bigint;not null"` // 上级用户ID（tg_user.id）
	SourceType     int        `json:"sourceType" gorm:"column:source_type;type:tinyint;not null"`     // 来源类型：1=下注流水 2=充值 3=游戏输赢 4=手动补单 5=其他
	SourceOrderId  string     `json:"sourceOrderId" gorm:"column:source_order_id;type:varchar(64);not null"`
	SourceAmount   float64    `json:"sourceAmount" gorm:"column:source_amount;type:decimal(20,3);not null;default:0.000"`
	RebateRate     float64    `json:"rebateRate" gorm:"column:rebate_rate;type:decimal(10,6);not null;default:0.000000"`
	RebateAmount   float64    `json:"rebateAmount" gorm:"column:rebate_amount;type:decimal(20,3);not null;default:0.000"`
	Currency       string     `json:"currency" gorm:"column:currency;type:varchar(16);not null;default:USDT"`
	Status         int        `json:"status" gorm:"column:status;type:tinyint;not null;default:0"` // 状态：0=待结算 1=已结算 2=作废/冲正
	SettledAt      *time.Time `json:"settledAt" gorm:"column:settled_at;type:datetime(3)"`
	IdempotencyKey string     `json:"idempotencyKey" gorm:"column:idempotency_key;type:varchar(80);uniqueIndex;not null"`
	Remark         *string    `json:"remark" gorm:"column:remark;type:varchar(255)"`
}

type TgUserRebateRecordSearch struct {
	PageInfo
	TenantId      *int64 `json:"tenantId"`
	SubUserId     int64  `json:"subUserId"`
	ParentUserId  int64  `json:"parentUserId"`
	SourceType    *int   `json:"sourceType"`
	SourceOrderId string `json:"sourceOrderId"`
	Status        *int   `json:"status"`
	IdempotencyKey string `json:"idempotencyKey"`
}

type TgUserRebateRecordSet struct {
	ID             int64      `json:"id"`
	TenantId       *int64     `json:"tenantId"`
	SubUserId      int64      `json:"subUserId"`
	ParentUserId   int64      `json:"parentUserId"`
	SourceType     int        `json:"sourceType"`
	SourceOrderId  string     `json:"sourceOrderId"`
	SourceAmount   float64    `json:"sourceAmount"`
	RebateRate     float64    `json:"rebateRate"`
	RebateAmount   float64    `json:"rebateAmount"`
	Currency       string     `json:"currency"`
	Status         int        `json:"status"`
	SettledAt      *time.Time `json:"settledAt"`
	IdempotencyKey string     `json:"idempotencyKey"`
	Remark         *string    `json:"remark"`
}

type TgUserRebateRecordBack struct {
	ID             int64      `json:"id"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`
	TenantId       *int64     `json:"tenantId"`
	SubUserId      int64      `json:"subUserId"`
	ParentUserId   int64      `json:"parentUserId"`
	SourceType     int        `json:"sourceType"`
	SourceOrderId  string     `json:"sourceOrderId"`
	SourceAmount   float64    `json:"sourceAmount"`
	RebateRate     float64    `json:"rebateRate"`
	RebateAmount   float64    `json:"rebateAmount"`
	Currency       string     `json:"currency"`
	Status         int        `json:"status"`
	SettledAt      *time.Time `json:"settledAt"`
	IdempotencyKey string     `json:"idempotencyKey"`
	Remark         *string    `json:"remark"`
}

type TgUserRebateRecordResp struct {
	BasePageResponse[TgUserRebateRecordBack]
}

var TgUserRebateRecordTableName = "tg_user_rebate_record"

func (TgUserRebateRecord) TableName() string {
	return TgUserRebateRecordTableName
}
