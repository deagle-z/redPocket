package pojo

import "time"

type RechargeOrder struct { // 充值记录/充值订单表
	BaseModel
	TenantId        int64      `json:"tenantId" gorm:"column:tenant_id;type:bigint"`                        // 租户ID
	AppId           *int64     `json:"appId" gorm:"column:app_id;type:bigint"`                              // 应用/项目ID（可选）
	UserId          int64      `json:"userId" gorm:"column:user_id;type:bigint"`                            // 用户ID
	AccountId       *string    `json:"accountId" gorm:"column:account_id;type:varchar(64)"`                 // 账号ID/外部账号标识（可选）
	OrderNo         string     `json:"orderNo" gorm:"column:order_no;type:varchar(64);uniqueIndex"`         // 平台充值订单号（唯一）
	MerchantOrderNo *string    `json:"merchantOrderNo" gorm:"column:merchant_order_no;type:varchar(64)"`    // 商户侧订单号/前端单号（可选）
	Channel         string     `json:"channel" gorm:"column:channel;type:varchar(32)"`                      // 充值渠道
	PayMethod       *string    `json:"payMethod" gorm:"column:pay_method;type:varchar(32)"`                 // 支付方式/子渠道
	Currency        string     `json:"currency" gorm:"column:currency;type:varchar(8);default:USD"`         // 币种
	Amount          float64    `json:"amount" gorm:"column:amount;type:numeric(18,6)"`                      // 充值金额
	Fee             float64    `json:"fee" gorm:"column:fee;type:numeric(18,6);default:0"`                  // 手续费
	NetAmount       float64    `json:"netAmount" gorm:"column:net_amount;type:numeric(18,6);->"`            // 净入账金额（只读）
	CreditAmount    *float64   `json:"creditAmount" gorm:"column:credit_amount;type:numeric(18,6)"`         // 实际入账金额（可选）
	BonusAmount     float64    `json:"bonusAmount" gorm:"column:bonus_amount;type:numeric(18,6);default:0"` // 赠送金额
	Status          int        `json:"status" gorm:"column:status;type:tinyint;default:0"`                  // 状态
	PayTime         *time.Time `json:"payTime" gorm:"column:pay_time;type:datetime(3)"`                     // 支付成功时间
	ExpireTime      *time.Time `json:"expireTime" gorm:"column:expire_time;type:datetime(3)"`               // 过期时间（可选）
	NotifyTime      *time.Time `json:"notifyTime" gorm:"column:notify_time;type:datetime(3)"`               // 回调/通知时间（可选）
	Provider        *string    `json:"provider" gorm:"column:provider;type:varchar(32)"`                    // 支付服务商
	ProviderTradeNo *string    `json:"providerTradeNo" gorm:"column:provider_trade_no;type:varchar(128)"`   // 三方交易号/流水号
	ProviderStatus  *string    `json:"providerStatus" gorm:"column:provider_status;type:varchar(64)"`       // 三方状态原文（可选）
	NotifyCount     int        `json:"notifyCount" gorm:"column:notify_count;type:int;default:0"`           // 通知次数
	NotifyLastAt    *time.Time `json:"notifyLastAt" gorm:"column:notify_last_at;type:datetime(3)"`          // 最后一次通知时间
	IdempotencyKey  *string    `json:"idempotencyKey" gorm:"column:idempotency_key;type:varchar(128)"`      // 幂等键（可选）
	Title           *string    `json:"title" gorm:"column:title;type:varchar(128)"`                         // 充值标题/套餐名（可选）
	Remark          *string    `json:"remark" gorm:"column:remark;type:varchar(255)"`                       // 备注（可选）
	Extra           *string    `json:"extra" gorm:"column:extra;type:json"`                                 // 扩展信息（可选）
	DeletedAt       *time.Time `json:"deletedAt" gorm:"column:deleted_at;type:datetime(3)"`                 // 软删除时间（可选）
}

type RechargeOrderSearch struct {
	PageInfo
	TenantId        int64  `json:"tenantId"`        // 租户ID
	UserId          int64  `json:"userId"`          // 用户ID
	Status          *int   `json:"status"`          // 状态
	OrderNo         string `json:"orderNo"`         // 平台订单号
	MerchantOrderNo string `json:"merchantOrderNo"` // 商户订单号
	ProviderTradeNo string `json:"providerTradeNo"` // 三方交易号
	Channel         string `json:"channel"`         // 充值渠道
	PayMethod       string `json:"payMethod"`       // 支付方式
}

type RechargeOrderSet struct {
	ID              int64      `json:"id"` // ID
	TenantId        int64      `json:"tenantId"`
	AppId           *int64     `json:"appId"`
	UserId          int64      `json:"userId"`
	AccountId       *string    `json:"accountId"`
	OrderNo         string     `json:"orderNo"`
	MerchantOrderNo *string    `json:"merchantOrderNo"`
	Channel         string     `json:"channel"`
	PayMethod       *string    `json:"payMethod"`
	Currency        string     `json:"currency"`
	Amount          float64    `json:"amount"`
	Fee             float64    `json:"fee"`
	CreditAmount    *float64   `json:"creditAmount"`
	BonusAmount     float64    `json:"bonusAmount"`
	Status          int        `json:"status"`
	PayTime         *time.Time `json:"payTime"`
	ExpireTime      *time.Time `json:"expireTime"`
	NotifyTime      *time.Time `json:"notifyTime"`
	Provider        *string    `json:"provider"`
	ProviderTradeNo *string    `json:"providerTradeNo"`
	ProviderStatus  *string    `json:"providerStatus"`
	NotifyCount     int        `json:"notifyCount"`
	NotifyLastAt    *time.Time `json:"notifyLastAt"`
	IdempotencyKey  *string    `json:"idempotencyKey"`
	Title           *string    `json:"title"`
	Remark          *string    `json:"remark"`
	Extra           *string    `json:"extra"`
}

type RechargeOrderBack struct {
	ID              int64      `json:"id"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
	DeletedAt       *time.Time `json:"deletedAt"`
	TenantId        int64      `json:"tenantId"`
	AppId           *int64     `json:"appId"`
	UserId          int64      `json:"userId"`
	AccountId       *string    `json:"accountId"`
	OrderNo         string     `json:"orderNo"`
	MerchantOrderNo *string    `json:"merchantOrderNo"`
	Channel         string     `json:"channel"`
	PayMethod       *string    `json:"payMethod"`
	Currency        string     `json:"currency"`
	Amount          float64    `json:"amount"`
	Fee             float64    `json:"fee"`
	NetAmount       float64    `json:"netAmount"`
	CreditAmount    *float64   `json:"creditAmount"`
	BonusAmount     float64    `json:"bonusAmount"`
	Status          int        `json:"status"`
	PayTime         *time.Time `json:"payTime"`
	ExpireTime      *time.Time `json:"expireTime"`
	NotifyTime      *time.Time `json:"notifyTime"`
	Provider        *string    `json:"provider"`
	ProviderTradeNo *string    `json:"providerTradeNo"`
	ProviderStatus  *string    `json:"providerStatus"`
	NotifyCount     int        `json:"notifyCount"`
	NotifyLastAt    *time.Time `json:"notifyLastAt"`
	IdempotencyKey  *string    `json:"idempotencyKey"`
	Title           *string    `json:"title"`
	Remark          *string    `json:"remark"`
	Extra           *string    `json:"extra"`
}

type RechargeOrderResp struct {
	BasePageResponse[RechargeOrderBack]
}

var RechargeOrderTableName = "recharge_order"

func (RechargeOrder) TableName() string {
	return RechargeOrderTableName
}
