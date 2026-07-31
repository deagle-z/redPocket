package pojo

import "time"

const (
	CryptoRechargeChannelUSDTTRC20 = "USDT_TRC20"
	CryptoRechargeChannelBTC       = "BTC"
	CryptoRechargeChannelETH       = "ETH"
	CryptoRechargePayMethodTRC20   = "TRC20"
	CryptoRechargePayMethodBTC     = "BTC"
	CryptoRechargePayMethodETH     = "ETH"
	CryptoRechargeCurrencyUSDT     = "USDT"
	CryptoRechargeCurrencyBTC      = "BTC"
	CryptoRechargeCurrencyETH      = "ETH"
	CryptoRechargeNetworkTRC20     = "TRC20"
	CryptoRechargeNetworkBitcoin   = "BITCOIN"
	CryptoRechargeNetworkEthereum  = "ETHEREUM"
	CryptoRechargeTokenUSDT        = "USDT"
	CryptoRechargeTokenBTC         = "BTC"
	CryptoRechargeTokenETH         = "ETH"

	CryptoRechargeDecimalsUSDT = 6
	CryptoRechargeDecimalsBTC  = 8
	CryptoRechargeDecimalsETH  = 18

	UsdtRechargeStatusPending  int8 = 0
	UsdtRechargeStatusPaid     int8 = 1
	UsdtRechargeStatusExpired  int8 = 2
	UsdtRechargeStatusCanceled int8 = 3

	CryptoRechargeTxReviewNew                int8 = 0
	CryptoRechargeTxReviewAutoMatched        int8 = 1
	CryptoRechargeTxReviewManualRequired     int8 = 2
	CryptoRechargeTxReviewManualSupplemented int8 = 3

	CryptoRechargeExceptionNoActiveOrder  = "NO_ACTIVE_ORDER"
	CryptoRechargeExceptionAmountMismatch = "AMOUNT_MISMATCH"
	CryptoRechargeExceptionLatePayment    = "LATE_PAYMENT"
	CryptoRechargeExceptionPreOrder       = "PRE_ORDER_PAYMENT"
	CryptoRechargeExceptionDuplicate      = "DUPLICATE_PAYMENT"
	CryptoRechargeExceptionLegacy         = "LEGACY_UNMATCHED"
	CryptoRechargeExceptionChainReorg     = "CHAIN_REORG"
	CryptoRechargeExceptionOrderState     = "ORDER_STATE_MISMATCH"

	CryptoRechargeManualStatusProcessing int8 = 0
	CryptoRechargeManualStatusCompleted  int8 = 1
)

type CryptoRechargeOrderReq struct {
	UserID                         int64   `json:"userId"`                         // TG用户ID
	Amount                         float64 `json:"amount"`                         // 平台货币充值金额
	Network                        string  `json:"network"`                        // 网络：TRC20/BITCOIN/ETHEREUM
	Token                          string  `json:"token"`                          // 币种：USDT/BTC/ETH
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

type CryptoRechargeManualProofReq struct {
	SourceNetwork  string `json:"sourceNetwork"`
	SourceToken    string `json:"sourceToken"`
	TxID           string `json:"txId"`
	AmountDecimals int    `json:"amountDecimals"`
	AmountAtomic   string `json:"amountAtomic"`
}

type CryptoRechargeManualSupplementReq struct {
	OrderNo      string                         `json:"orderNo"`
	Reason       string                         `json:"reason"`
	Transactions []CryptoRechargeManualProofReq `json:"transactions"`
}

type CryptoRechargeExceptionSearch struct {
	PageInfo
	Network       string `json:"network" form:"network"`
	Token         string `json:"token" form:"token"`
	TxID          string `json:"txId" form:"txId"`
	ExceptionCode string `json:"exceptionCode" form:"exceptionCode"`
}

type CryptoRechargeManualSupplementBack struct {
	Supplement CryptoRechargeManualSupplement `json:"supplement"`
	Proofs     []CryptoRechargeManualProof    `json:"proofs"`
}

type UsdtRechargeOrder struct {
	BaseModel
	TenantId             int64      `json:"tenantId" gorm:"column:tenant_id;type:bigint;index"`
	UserId               int64      `json:"userId" gorm:"column:user_id;type:bigint;index"`
	SourceChannelID      *int64     `json:"sourceChannelId" gorm:"column:source_channel_id;type:bigint;index"`
	OrderNo              string     `json:"orderNo" gorm:"column:order_no;type:varchar(64);uniqueIndex"`
	Network              string     `json:"network" gorm:"column:network;type:varchar(16);index;index:idx_crypto_recharge_match,priority:1"`
	Token                string     `json:"token" gorm:"column:token;type:varchar(16);index:idx_crypto_recharge_match,priority:2"`
	ContractAddress      string     `json:"contractAddress" gorm:"column:contract_address;type:varchar(64);index"`
	ReceiveAddress       string     `json:"receiveAddress" gorm:"column:receive_address;type:varchar(128);index:idx_usdt_recharge_match,priority:1;index:idx_crypto_recharge_match,priority:3"`
	PlatformAmount       float64    `json:"platformAmount" gorm:"column:platform_amount;type:numeric(18,2)"`
	PlatformRate         string     `json:"platformRate" gorm:"column:platform_rate;type:varchar(32)"`
	BaseAmountMicro      int64      `json:"baseAmountMicro" gorm:"column:base_amount_micro;type:bigint"`
	ExpectedAmountMicro  int64      `json:"expectedAmountMicro" gorm:"column:expected_amount_micro;type:bigint;index:idx_usdt_recharge_match,priority:2"`
	AmountDecimals       int        `json:"amountDecimals" gorm:"column:amount_decimals;type:int;not null;default:6"`
	BaseAmountAtomic     string     `json:"baseAmountAtomic" gorm:"column:base_amount_atomic;type:varchar(96)"`
	ExpectedAmountAtomic string     `json:"expectedAmountAtomic" gorm:"column:expected_amount_atomic;type:varchar(96);index:idx_crypto_recharge_match,priority:4"`
	ExpectedAmount       string     `json:"expectedAmount" gorm:"column:expected_amount;type:varchar(96)"`
	Status               int8       `json:"status" gorm:"column:status;type:tinyint;default:0;index:idx_usdt_recharge_match,priority:3;index:idx_crypto_recharge_match,priority:5"`
	ExpireTime           time.Time  `json:"expireTime" gorm:"column:expire_time;type:datetime(3);index"`
	PaidAmountMicro      *int64     `json:"paidAmountMicro" gorm:"column:paid_amount_micro;type:bigint"`
	PaidAmountAtomic     *string    `json:"paidAmountAtomic" gorm:"column:paid_amount_atomic;type:varchar(96)"`
	PaidAt               *time.Time `json:"paidAt" gorm:"column:paid_at;type:datetime(3)"`
	TxID                 *string    `json:"txId" gorm:"column:tx_id;type:varchar(128);index"`
}

type UsdtRechargeTx struct {
	BaseModel
	TxID            string     `json:"txId" gorm:"column:tx_id;type:varchar(128);uniqueIndex"`
	Network         string     `json:"network" gorm:"column:network;type:varchar(16);index"`
	Token           string     `json:"token" gorm:"column:token;type:varchar(16)"`
	ContractAddress string     `json:"contractAddress" gorm:"column:contract_address;type:varchar(64);index"`
	FromAddress     string     `json:"fromAddress" gorm:"column:from_address;type:varchar(128);index"`
	ToAddress       string     `json:"toAddress" gorm:"column:to_address;type:varchar(128);index"`
	AmountMicro     int64      `json:"amountMicro" gorm:"column:amount_micro;type:bigint;index"`
	AmountDecimals  int        `json:"amountDecimals" gorm:"column:amount_decimals;type:int;not null;default:6"`
	AmountAtomic    string     `json:"amountAtomic" gorm:"column:amount_atomic;type:varchar(96);index"`
	Amount          string     `json:"amount" gorm:"column:amount;type:varchar(96)"`
	BlockHeight     int64      `json:"blockHeight" gorm:"column:block_height;type:bigint;index"`
	BlockHash       string     `json:"blockHash" gorm:"column:block_hash;type:varchar(128);index"`
	Confirmations   int        `json:"confirmations" gorm:"column:confirmations;type:int;not null;default:0"`
	BlockTimestamp  int64      `json:"blockTimestamp" gorm:"column:block_timestamp;type:bigint;index"`
	RawJSON         string     `json:"rawJson" gorm:"column:raw_json;type:longtext"`
	MatchedOrderNo  *string    `json:"matchedOrderNo" gorm:"column:matched_order_no;type:varchar(64);index"`
	MatchedAt       *time.Time `json:"matchedAt" gorm:"column:matched_at;type:datetime(3)"`
	ReviewStatus    int8       `json:"reviewStatus" gorm:"column:review_status;type:tinyint;not null;default:0;index"`
	ExceptionCode   string     `json:"exceptionCode" gorm:"column:exception_code;type:varchar(32);index"`
}

type CryptoRechargeScanCursor struct {
	BaseModel
	CursorKey            string `json:"cursorKey" gorm:"column:cursor_key;type:varchar(255);uniqueIndex"`
	Network              string `json:"network" gorm:"column:network;type:varchar(16);index"`
	Token                string `json:"token" gorm:"column:token;type:varchar(16)"`
	ReceiveAddress       string `json:"receiveAddress" gorm:"column:receive_address;type:varchar(128)"`
	ContractAddress      string `json:"contractAddress" gorm:"column:contract_address;type:varchar(128)"`
	LastBlockHeight      int64  `json:"lastBlockHeight" gorm:"column:last_block_height;type:bigint;not null;default:0"`
	LastBlockHash        string `json:"lastBlockHash" gorm:"column:last_block_hash;type:varchar(128)"`
	LastBlockTimestamp   int64  `json:"lastBlockTimestamp" gorm:"column:last_block_timestamp;type:bigint;not null;default:0"`
	PaginationCursor     string `json:"paginationCursor" gorm:"column:pagination_cursor;type:varchar(128)"`
	TargetBlockHeight    int64  `json:"targetBlockHeight" gorm:"column:target_block_height;type:bigint;not null;default:0"`
	TargetBlockTimestamp int64  `json:"targetBlockTimestamp" gorm:"column:target_block_timestamp;type:bigint;not null;default:0"`
}

type CryptoRechargeAmountReservation struct {
	BaseModel
	ReservationKey       string `json:"reservationKey" gorm:"column:reservation_key;type:char(64);uniqueIndex"`
	Network              string `json:"network" gorm:"column:network;type:varchar(16);index"`
	Token                string `json:"token" gorm:"column:token;type:varchar(16)"`
	ReceiveAddress       string `json:"receiveAddress" gorm:"column:receive_address;type:varchar(128)"`
	ExpectedAmountAtomic string `json:"expectedAmountAtomic" gorm:"column:expected_amount_atomic;type:varchar(96)"`
	OrderNo              string `json:"orderNo" gorm:"column:order_no;type:varchar(64);uniqueIndex"`
}

type CryptoRechargeManualSupplement struct {
	BaseModel
	OrderNo         string  `json:"orderNo" gorm:"column:order_no;type:varchar(64);uniqueIndex"`
	RechargeOrderID int64   `json:"rechargeOrderId" gorm:"column:recharge_order_id;type:bigint;index"`
	UserID          int64   `json:"userId" gorm:"column:user_id;type:bigint;index"`
	PlatformAmount  float64 `json:"platformAmount" gorm:"column:platform_amount;type:numeric(18,2)"`
	OperatorID      int64   `json:"operatorId" gorm:"column:operator_id;type:bigint;index"`
	Reason          string  `json:"reason" gorm:"column:reason;type:varchar(500)"`
	Status          int8    `json:"status" gorm:"column:status;type:tinyint;not null;default:0;index"`
}

type CryptoRechargeManualProof struct {
	BaseModel
	SupplementID   int64  `json:"supplementId" gorm:"column:supplement_id;type:bigint;index"`
	SourceNetwork  string `json:"sourceNetwork" gorm:"column:source_network;type:varchar(32);index"`
	SourceToken    string `json:"sourceToken" gorm:"column:source_token;type:varchar(32)"`
	TxID           string `json:"txId" gorm:"column:tx_id;type:varchar(128);uniqueIndex"`
	AmountDecimals int    `json:"amountDecimals" gorm:"column:amount_decimals;type:int;not null"`
	AmountAtomic   string `json:"amountAtomic" gorm:"column:amount_atomic;type:varchar(96)"`
}

type CryptoRechargeSchemaVersion struct {
	BaseModel
	MigrationKey string `json:"migrationKey" gorm:"column:migration_key;type:varchar(64);uniqueIndex"`
	Version      int    `json:"version" gorm:"column:version;type:int;not null"`
}

var UsdtRechargeOrderTableName = "usdt_recharge_order"
var UsdtRechargeTxTableName = "usdt_recharge_tx"
var CryptoRechargeScanCursorTableName = "crypto_recharge_scan_cursor"
var CryptoRechargeAmountReservationTableName = "crypto_recharge_amount_reservation"
var CryptoRechargeManualSupplementTableName = "crypto_recharge_manual_supplement"
var CryptoRechargeManualProofTableName = "crypto_recharge_manual_proof"
var CryptoRechargeSchemaVersionTableName = "crypto_recharge_schema_version"

func (UsdtRechargeOrder) TableName() string {
	return UsdtRechargeOrderTableName
}

func (UsdtRechargeTx) TableName() string {
	return UsdtRechargeTxTableName
}

func (CryptoRechargeScanCursor) TableName() string {
	return CryptoRechargeScanCursorTableName
}

func (CryptoRechargeAmountReservation) TableName() string {
	return CryptoRechargeAmountReservationTableName
}

func (CryptoRechargeManualSupplement) TableName() string {
	return CryptoRechargeManualSupplementTableName
}

func (CryptoRechargeManualProof) TableName() string {
	return CryptoRechargeManualProofTableName
}

func (CryptoRechargeSchemaVersion) TableName() string {
	return CryptoRechargeSchemaVersionTableName
}
