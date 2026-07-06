package pojo

import "time"

const (
	CryptoRechargeChannelUSDTTRC20 = "USDT_TRC20"
	CryptoRechargePayMethodTRC20   = "TRC20"
	CryptoRechargeCurrencyUSDT     = "USDT"
	CryptoRechargeNetworkTRC20     = "TRC20"
	CryptoRechargeTokenUSDT        = "USDT"

	UsdtRechargeStatusPending  int8 = 0
	UsdtRechargeStatusPaid     int8 = 1
	UsdtRechargeStatusExpired  int8 = 2
	UsdtRechargeStatusCanceled int8 = 3
)

type CryptoRechargeOrderReq struct {
	UserID                         int64   `json:"userId"`                         // TG用户ID
	Amount                         float64 `json:"amount"`                         // 平台货币充值金额
	Network                        string  `json:"network"`                        // 网络，当前仅支持TRC20
	Token                          string  `json:"token"`                          // 币种，当前仅支持USDT
	MerchantOrderNo                string  `json:"merchantOrderNo"`                // 商户订单号（可选）
	ActivityCode                   string  `json:"activityCode"`                   // 活动编码（预留，当前固定v2充值）
	ConfirmUnfinishedActivityCycle bool    `json:"confirmUnfinishedActivityCycle"` // 是否确认在未结束活动周期内继续充值
}

type CryptoPaymentBack struct {
	Network         string `json:"network"`
	Token           string `json:"token"`
	ContractAddress string `json:"contractAddress"`
	ReceiveAddress  string `json:"receiveAddress"`
	ExpectedAmount  string `json:"expectedAmount"`
	ExpireTime      string `json:"expireTime"`
	QRContent       string `json:"qrContent"`
	PlatformRate    string `json:"platformRate"`
}

type CryptoRechargeOptionBack struct {
	Network         string  `json:"network"`
	Token           string  `json:"token"`
	ContractAddress string  `json:"contractAddress"`
	ReceiveAddress  string  `json:"receiveAddress"`
	EstimatedAmount string  `json:"estimatedAmount"`
	PlatformRate    string  `json:"platformRate"`
	PlatformAmount  float64 `json:"platformAmount"`
}

type CryptoRechargeOptionsBack struct {
	PlatformAmount float64                    `json:"platformAmount"`
	Options        []CryptoRechargeOptionBack `json:"options"`
}

type CryptoRechargeOrderStatusBack struct {
	OrderNo          string             `json:"orderNo"`
	Status           int8               `json:"status"`
	RechargeStatus   int                `json:"rechargeStatus"`
	TxID             string             `json:"txId"`
	PayTime          *time.Time         `json:"payTime"`
	ExpireTime       string             `json:"expireTime"`
	RemainingSeconds int64              `json:"remainingSeconds"`
	CryptoPayment    *CryptoPaymentBack `json:"cryptoPayment,omitempty"`
}

type UsdtRechargeOrder struct {
	BaseModel
	TenantId            int64      `json:"tenantId" gorm:"column:tenant_id;type:bigint;index"`
	UserId              int64      `json:"userId" gorm:"column:user_id;type:bigint;index"`
	SourceChannelID     *int64     `json:"sourceChannelId" gorm:"column:source_channel_id;type:bigint;index"`
	OrderNo             string     `json:"orderNo" gorm:"column:order_no;type:varchar(64);uniqueIndex"`
	Network             string     `json:"network" gorm:"column:network;type:varchar(16);index"`
	Token               string     `json:"token" gorm:"column:token;type:varchar(16)"`
	ContractAddress     string     `json:"contractAddress" gorm:"column:contract_address;type:varchar(64);index"`
	ReceiveAddress      string     `json:"receiveAddress" gorm:"column:receive_address;type:varchar(64);index:idx_usdt_recharge_match,priority:1"`
	PlatformAmount      float64    `json:"platformAmount" gorm:"column:platform_amount;type:numeric(18,2)"`
	PlatformRate        string     `json:"platformRate" gorm:"column:platform_rate;type:varchar(32)"`
	BaseAmountMicro     int64      `json:"baseAmountMicro" gorm:"column:base_amount_micro;type:bigint"`
	ExpectedAmountMicro int64      `json:"expectedAmountMicro" gorm:"column:expected_amount_micro;type:bigint;index:idx_usdt_recharge_match,priority:2"`
	ExpectedAmount      string     `json:"expectedAmount" gorm:"column:expected_amount;type:varchar(32)"`
	Status              int8       `json:"status" gorm:"column:status;type:tinyint;default:0;index:idx_usdt_recharge_match,priority:3"`
	ExpireTime          time.Time  `json:"expireTime" gorm:"column:expire_time;type:datetime(3);index"`
	PaidAmountMicro     *int64     `json:"paidAmountMicro" gorm:"column:paid_amount_micro;type:bigint"`
	PaidAt              *time.Time `json:"paidAt" gorm:"column:paid_at;type:datetime(3)"`
	TxID                *string    `json:"txId" gorm:"column:tx_id;type:varchar(128);index"`
}

type UsdtRechargeTx struct {
	BaseModel
	TxID            string     `json:"txId" gorm:"column:tx_id;type:varchar(128);uniqueIndex"`
	Network         string     `json:"network" gorm:"column:network;type:varchar(16);index"`
	Token           string     `json:"token" gorm:"column:token;type:varchar(16)"`
	ContractAddress string     `json:"contractAddress" gorm:"column:contract_address;type:varchar(64);index"`
	FromAddress     string     `json:"fromAddress" gorm:"column:from_address;type:varchar(64);index"`
	ToAddress       string     `json:"toAddress" gorm:"column:to_address;type:varchar(64);index"`
	AmountMicro     int64      `json:"amountMicro" gorm:"column:amount_micro;type:bigint;index"`
	Amount          string     `json:"amount" gorm:"column:amount;type:varchar(32)"`
	BlockTimestamp  int64      `json:"blockTimestamp" gorm:"column:block_timestamp;type:bigint;index"`
	RawJSON         string     `json:"rawJson" gorm:"column:raw_json;type:longtext"`
	MatchedOrderNo  *string    `json:"matchedOrderNo" gorm:"column:matched_order_no;type:varchar(64);index"`
	MatchedAt       *time.Time `json:"matchedAt" gorm:"column:matched_at;type:datetime(3)"`
}

var UsdtRechargeOrderTableName = "usdt_recharge_order"
var UsdtRechargeTxTableName = "usdt_recharge_tx"

func (UsdtRechargeOrder) TableName() string {
	return UsdtRechargeOrderTableName
}

func (UsdtRechargeTx) TableName() string {
	return UsdtRechargeTxTableName
}
