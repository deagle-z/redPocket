package pojo

import (
	"encoding/json"
	"strings"
	"time"
)

type WithdrawOrderBr struct { // 巴西地区提现订单（Pix/TED/DOC）
	BaseModel
	TenantId                 int64      `json:"tenantId" gorm:"column:tenant_id;type:bigint"`                                                   // 租户ID
	AppId                    *int64     `json:"appId" gorm:"column:app_id;type:bigint"`                                                         // 应用/项目ID（可选）
	UserId                   int64      `json:"userId" gorm:"column:user_id;type:bigint"`                                                       // 用户ID
	SourceChannelID          *int64     `json:"sourceChannelId" gorm:"column:source_channel_id;type:bigint;index"`                              // 来源渠道ID
	AccountId                *string    `json:"accountId" gorm:"column:account_id;type:varchar(64)"`                                            // 账号ID/外部账号标识（可选）
	OrderNo                  string     `json:"orderNo" gorm:"column:order_no;type:varchar(64);uniqueIndex"`                                    // 平台提现订单号（唯一）
	MerchantOrderNo          *string    `json:"merchantOrderNo" gorm:"column:merchant_order_no;type:varchar(64);uniqueIndex"`                   // 商户侧订单号/前端单号（可选）
	Currency                 string     `json:"currency" gorm:"column:currency;type:varchar(8);default:BRL"`                                    // 币种
	CountryCode              string     `json:"countryCode" gorm:"column:country_code;type:varchar(8);index"`                                   // 国家编码
	Amount                   float64    `json:"amount" gorm:"column:amount;type:numeric(18,2)"`                                                 // 提现金额
	Fee                      float64    `json:"fee" gorm:"column:fee;type:numeric(18,2);default:0"`                                             // 手续费
	NetAmount                float64    `json:"netAmount" gorm:"column:net_amount;type:numeric(18,2);->"`                                       // 实际打款金额（只读）
	Channel                  string     `json:"channel" gorm:"column:channel;type:varchar(32);default:pix"`                                     // 渠道
	PayMethod                *string    `json:"payMethod" gorm:"column:pay_method;type:varchar(32)"`                                            // 方式/子渠道
	Status                   int        `json:"status" gorm:"column:status;type:tinyint;default:0"`                                             // 状态：0待审核 1待打款 2打款中 3成功 4失败 5取消 6退回
	ReviewedBy               *int64     `json:"reviewedBy" gorm:"column:reviewed_by;type:bigint"`                                               // 审核人ID（后台）
	ReviewedAt               *time.Time `json:"reviewedAt" gorm:"column:reviewed_at;type:datetime(3)"`                                          // 审核时间
	PaidAt                   *time.Time `json:"paidAt" gorm:"column:paid_at;type:datetime(3)"`                                                  // 打款完成时间
	FailCode                 *string    `json:"failCode" gorm:"column:fail_code;type:varchar(64)"`                                              // 失败码（可选）
	FailMsg                  *string    `json:"failMsg" gorm:"column:fail_msg;type:varchar(255)"`                                               // 失败原因（可选）
	ReceiverName             *string    `json:"receiverName" gorm:"column:receiver_name;type:varchar(128)"`                                     // 收款人姓名（可选）
	ReceiverDocument         *string    `json:"receiverDocument" gorm:"column:receiver_document;type:varchar(32)"`                              // 收款人证件号（可选）
	ReceiverDocumentType     *string    `json:"receiverDocumentType" gorm:"column:receiver_document_type;type:varchar(8)"`                      // 证件类型
	PixKeyType               *string    `json:"pixKeyType" gorm:"column:pix_key_type;type:varchar(16)"`                                         // Pix Key类型
	PixKey                   *string    `json:"pixKey" gorm:"column:pix_key;type:varchar(128)"`                                                 // Pix Key（可选）
	BankCode                 *string    `json:"bankCode" gorm:"column:bank_code;type:varchar(16)"`                                              // 银行代码（可选）
	BankName                 *string    `json:"bankName" gorm:"column:bank_name;type:varchar(64)"`                                              // 银行名称（可选）
	BranchNumber             *string    `json:"branchNumber" gorm:"column:branch_number;type:varchar(16)"`                                      // 支行号/Agency（可选）
	AccountNumber            *string    `json:"accountNumber" gorm:"column:account_number;type:varchar(32)"`                                    // 账号（可选）
	AccountType              *string    `json:"accountType" gorm:"column:account_type;type:varchar(16)"`                                        // 账户类型（可选）
	Provider                 *string    `json:"provider" gorm:"column:provider;type:varchar(32)"`                                               // 代付服务商（可选）
	ProviderPayoutNo         *string    `json:"providerPayoutNo" gorm:"column:provider_payout_no;type:varchar(128)"`                            // 三方代付单号/流水号（可选）
	ProviderStatus           *string    `json:"providerStatus" gorm:"column:provider_status;type:varchar(64)"`                                  // 三方状态原文（可选）
	GiftRestrictedAmount     float64    `json:"giftRestrictedAmount" gorm:"column:gift_restricted_amount;type:numeric(18,2);default:0"`         // 本次提现占用的赠送受限金额
	RechargeRestrictedAmount float64    `json:"rechargeRestrictedAmount" gorm:"column:recharge_restricted_amount;type:numeric(18,2);default:0"` // 本次提现占用的充值受限金额
	UnrestrictedAmount       float64    `json:"unrestrictedAmount" gorm:"column:unrestricted_amount;type:numeric(18,2);default:0"`              // 本次提现占用的不受限金额
	GiftFlowRequired         float64    `json:"giftFlowRequired" gorm:"column:gift_flow_required;type:numeric(18,2);default:0"`                 // 本次提现占用的赠送流水
	RechargeFlowRequired     float64    `json:"rechargeFlowRequired" gorm:"column:recharge_flow_required;type:numeric(18,2);default:0"`         // 本次提现占用的充值流水
	NotifyTime               *time.Time `json:"notifyTime" gorm:"column:notify_time;type:datetime(3)"`                                          // 回调/通知时间（可选）
	NotifyCount              int        `json:"notifyCount" gorm:"column:notify_count;type:int;default:0"`                                      // 通知次数
	IdempotencyKey           *string    `json:"idempotencyKey" gorm:"column:idempotency_key;type:varchar(128)"`                                 // 幂等键（可选）
	RiskLevel                int        `json:"riskLevel" gorm:"column:risk_level;type:tinyint;default:0"`                                      // 风控等级
	Remark                   *string    `json:"remark" gorm:"column:remark;type:varchar(255)"`                                                  // 备注（可选）
	Extra                    *string    `json:"extra" gorm:"column:extra;type:json"`                                                            // 扩展信息（可选）
	DeletedAt                *time.Time `json:"deletedAt" gorm:"column:deleted_at;type:datetime(3)"`                                            // 软删除时间（可选）
}

type WithdrawOrderBrSearch struct {
	PageInfo
	TenantId             int64  `json:"tenantId"`             // 租户ID
	UserId               int64  `json:"userId"`               // 用户ID
	UserUid              string `json:"userUid"`              // 用户UID
	Status               *int   `json:"status"`               // 状态
	OrderNo              string `json:"orderNo"`              // 平台订单号
	MerchantOrderNo      string `json:"merchantOrderNo"`      // 商户订单号
	ProviderPayoutNo     string `json:"providerPayoutNo"`     // 三方代付单号
	CountryCode          string `json:"countryCode"`          // 国家编码
	Channel              string `json:"channel"`              // 渠道
	PayMethod            string `json:"payMethod"`            // 方式/子渠道
	ReceiverDocumentType string `json:"receiverDocumentType"` // 证件类型
	ReceiverDocument     string `json:"receiverDocument"`     // 证件号
}

type WithdrawOrderBrSet struct {
	ID                   int64 `json:"id"` // ID
	jsonFields           map[string]struct{}
	TenantId             int64      `json:"tenantId"`
	AppId                *int64     `json:"appId"`
	UserId               int64      `json:"userId"`
	SourceChannelID      *int64     `json:"sourceChannelId"`
	AccountId            *string    `json:"accountId"`
	OrderNo              string     `json:"orderNo"`
	MerchantOrderNo      *string    `json:"merchantOrderNo"`
	Currency             string     `json:"currency"`
	CountryCode          string     `json:"countryCode"`
	Amount               float64    `json:"amount"`
	Fee                  float64    `json:"fee"`
	Channel              string     `json:"channel"`
	PayMethod            *string    `json:"payMethod"`
	Status               int        `json:"status"`
	ReviewedBy           *int64     `json:"reviewedBy"`
	ReviewedAt           *time.Time `json:"reviewedAt"`
	PaidAt               *time.Time `json:"paidAt"`
	FailCode             *string    `json:"failCode"`
	FailMsg              *string    `json:"failMsg"`
	ReceiverName         *string    `json:"receiverName"`
	ReceiverDocument     *string    `json:"receiverDocument"`
	ReceiverDocumentType *string    `json:"receiverDocumentType"`
	PixKeyType           *string    `json:"pixKeyType"`
	PixKey               *string    `json:"pixKey"`
	BankCode             *string    `json:"bankCode"`
	BankName             *string    `json:"bankName"`
	BranchNumber         *string    `json:"branchNumber"`
	AccountNumber        *string    `json:"accountNumber"`
	AccountType          *string    `json:"accountType"`
	Provider             *string    `json:"provider"`
	ProviderPayoutNo     *string    `json:"providerPayoutNo"`
	ProviderStatus       *string    `json:"providerStatus"`
	NotifyTime           *time.Time `json:"notifyTime"`
	NotifyCount          int        `json:"notifyCount"`
	IdempotencyKey       *string    `json:"idempotencyKey"`
	RiskLevel            int        `json:"riskLevel"`
	Remark               *string    `json:"remark"`
	Extra                *string    `json:"extra"`
}

type withdrawOrderBrSetJSON struct {
	ID                   int64             `json:"id"`
	TenantId             int64             `json:"tenantId"`
	AppId                *int64            `json:"appId"`
	UserId               int64             `json:"userId"`
	SourceChannelID      *int64            `json:"sourceChannelId"`
	AccountId            *string           `json:"accountId"`
	OrderNo              string            `json:"orderNo"`
	MerchantOrderNo      *string           `json:"merchantOrderNo"`
	Currency             string            `json:"currency"`
	CountryCode          string            `json:"countryCode"`
	Amount               float64           `json:"amount"`
	Fee                  float64           `json:"fee"`
	Channel              string            `json:"channel"`
	PayMethod            *string           `json:"payMethod"`
	Status               int               `json:"status"`
	ReviewedBy           *int64            `json:"reviewedBy"`
	ReviewedAt           *withdrawJSONTime `json:"reviewedAt"`
	PaidAt               *withdrawJSONTime `json:"paidAt"`
	FailCode             *string           `json:"failCode"`
	FailMsg              *string           `json:"failMsg"`
	ReceiverName         *string           `json:"receiverName"`
	ReceiverDocument     *string           `json:"receiverDocument"`
	ReceiverDocumentType *string           `json:"receiverDocumentType"`
	PixKeyType           *string           `json:"pixKeyType"`
	PixKey               *string           `json:"pixKey"`
	BankCode             *string           `json:"bankCode"`
	BankName             *string           `json:"bankName"`
	BranchNumber         *string           `json:"branchNumber"`
	AccountNumber        *string           `json:"accountNumber"`
	AccountType          *string           `json:"accountType"`
	Provider             *string           `json:"provider"`
	ProviderPayoutNo     *string           `json:"providerPayoutNo"`
	ProviderStatus       *string           `json:"providerStatus"`
	NotifyTime           *withdrawJSONTime `json:"notifyTime"`
	NotifyCount          int               `json:"notifyCount"`
	IdempotencyKey       *string           `json:"idempotencyKey"`
	RiskLevel            int               `json:"riskLevel"`
	Remark               *string           `json:"remark"`
	Extra                *string           `json:"extra"`
}

type withdrawJSONTime struct {
	time.Time
}

func (t *withdrawJSONTime) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}
	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	layouts := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02 15:04:05.999",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		"2006-01-02",
	}
	var lastErr error
	for _, layout := range layouts {
		parsed, err := time.ParseInLocation(layout, value, time.Local)
		if err == nil {
			t.Time = parsed
			return nil
		}
		lastErr = err
	}
	return lastErr
}

func (req *WithdrawOrderBrSet) UnmarshalJSON(data []byte) error {
	var raw withdrawOrderBrSetJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}
	req.jsonFields = make(map[string]struct{}, len(fields))
	for field := range fields {
		req.jsonFields[field] = struct{}{}
	}
	req.ID = raw.ID
	req.TenantId = raw.TenantId
	req.AppId = raw.AppId
	req.UserId = raw.UserId
	req.SourceChannelID = raw.SourceChannelID
	req.AccountId = raw.AccountId
	req.OrderNo = raw.OrderNo
	req.MerchantOrderNo = raw.MerchantOrderNo
	req.Currency = raw.Currency
	req.CountryCode = NormalizeWithdrawCountryCode(raw.CountryCode)
	req.Amount = raw.Amount
	req.Fee = raw.Fee
	req.Channel = raw.Channel
	req.PayMethod = raw.PayMethod
	req.Status = raw.Status
	req.ReviewedBy = raw.ReviewedBy
	req.ReviewedAt = withdrawJSONTimePtr(raw.ReviewedAt)
	req.PaidAt = withdrawJSONTimePtr(raw.PaidAt)
	req.FailCode = raw.FailCode
	req.FailMsg = raw.FailMsg
	req.ReceiverName = raw.ReceiverName
	req.ReceiverDocument = raw.ReceiverDocument
	req.ReceiverDocumentType = raw.ReceiverDocumentType
	req.PixKeyType = raw.PixKeyType
	req.PixKey = raw.PixKey
	req.BankCode = raw.BankCode
	req.BankName = raw.BankName
	req.BranchNumber = raw.BranchNumber
	req.AccountNumber = raw.AccountNumber
	req.AccountType = raw.AccountType
	req.Provider = raw.Provider
	req.ProviderPayoutNo = raw.ProviderPayoutNo
	req.ProviderStatus = raw.ProviderStatus
	req.NotifyTime = withdrawJSONTimePtr(raw.NotifyTime)
	req.NotifyCount = raw.NotifyCount
	req.IdempotencyKey = raw.IdempotencyKey
	req.RiskLevel = raw.RiskLevel
	req.Remark = raw.Remark
	req.Extra = raw.Extra
	return nil
}

func (req WithdrawOrderBrSet) HasJSONFields() bool {
	return len(req.jsonFields) > 0
}

func (req WithdrawOrderBrSet) HasJSONField(field string) bool {
	_, ok := req.jsonFields[field]
	return ok
}

func NormalizeWithdrawCountryCode(value string) string {
	return strings.ToUpper(strings.TrimSpace(value))
}

func (req *WithdrawOrderBrSet) NormalizeCountryCodeFromExtra() {
	req.CountryCode = NormalizeWithdrawCountryCode(req.CountryCode)
	if req.CountryCode != "" || req.Extra == nil || strings.TrimSpace(*req.Extra) == "" {
		return
	}
	var extra map[string]any
	if err := json.Unmarshal([]byte(*req.Extra), &extra); err != nil {
		return
	}
	if countryCode, ok := extra["countryCode"].(string); ok {
		req.CountryCode = NormalizeWithdrawCountryCode(countryCode)
	}
}

func withdrawJSONTimePtr(value *withdrawJSONTime) *time.Time {
	if value == nil || value.IsZero() {
		return nil
	}
	parsed := value.Time
	return &parsed
}

type WithdrawOrderBrBack struct {
	ID                       int64      `json:"id"`
	CreatedAt                time.Time  `json:"createdAt"`
	UpdatedAt                time.Time  `json:"updatedAt"`
	DeletedAt                *time.Time `json:"deletedAt"`
	TenantId                 int64      `json:"tenantId"`
	AppId                    *int64     `json:"appId"`
	UserId                   int64      `json:"userId"`
	UserUid                  string     `json:"userUid"`
	SourceChannelID          *int64     `json:"sourceChannelId"`
	AccountId                *string    `json:"accountId"`
	OrderNo                  string     `json:"orderNo"`
	MerchantOrderNo          *string    `json:"merchantOrderNo"`
	Currency                 string     `json:"currency"`
	CountryCode              string     `json:"countryCode"`
	Amount                   float64    `json:"amount"`
	Fee                      float64    `json:"fee"`
	NetAmount                float64    `json:"netAmount"`
	Channel                  string     `json:"channel"`
	PayMethod                *string    `json:"payMethod"`
	Status                   int        `json:"status"`
	ReviewedBy               *int64     `json:"reviewedBy"`
	ReviewedAt               *time.Time `json:"reviewedAt"`
	PaidAt                   *time.Time `json:"paidAt"`
	FailCode                 *string    `json:"failCode"`
	FailMsg                  *string    `json:"failMsg"`
	ReceiverName             *string    `json:"receiverName"`
	ReceiverDocument         *string    `json:"receiverDocument"`
	ReceiverDocumentType     *string    `json:"receiverDocumentType"`
	PixKeyType               *string    `json:"pixKeyType"`
	PixKey                   *string    `json:"pixKey"`
	BankCode                 *string    `json:"bankCode"`
	BankName                 *string    `json:"bankName"`
	BranchNumber             *string    `json:"branchNumber"`
	AccountNumber            *string    `json:"accountNumber"`
	AccountType              *string    `json:"accountType"`
	Provider                 *string    `json:"provider"`
	ProviderPayoutNo         *string    `json:"providerPayoutNo"`
	ProviderStatus           *string    `json:"providerStatus"`
	GiftRestrictedAmount     float64    `json:"giftRestrictedAmount"`
	RechargeRestrictedAmount float64    `json:"rechargeRestrictedAmount"`
	UnrestrictedAmount       float64    `json:"unrestrictedAmount"`
	GiftFlowRequired         float64    `json:"giftFlowRequired"`
	RechargeFlowRequired     float64    `json:"rechargeFlowRequired"`
	NotifyTime               *time.Time `json:"notifyTime"`
	NotifyCount              int        `json:"notifyCount"`
	IdempotencyKey           *string    `json:"idempotencyKey"`
	RiskLevel                int        `json:"riskLevel"`
	Remark                   *string    `json:"remark"`
	Extra                    *string    `json:"extra"`
}

type WithdrawOrderBrResp struct {
	BasePageResponse[WithdrawOrderBrBack]
}

type AppCreateWithdrawOrderReq struct {
	Amount      float64           `json:"amount"`
	CountryCode string            `json:"countryCode"`
	AccountID   *int64            `json:"accountId"`
	FieldValues map[string]string `json:"fieldValues"`
}

type AppCreateWithdrawOrderResp struct {
	OrderNo      string  `json:"orderNo"`
	Fee          float64 `json:"fee,omitempty"`
	RebateAmount float64 `json:"rebateAmount,omitempty"`
}

var WithdrawOrderBrTableName = "withdraw_order_br"

func (WithdrawOrderBr) TableName() string {
	return WithdrawOrderBrTableName
}
