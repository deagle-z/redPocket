package pojo

import "time"

type WithdrawOrderBr struct { // 巴西地区提现订单（Pix/TED/DOC）
	BaseModel
	TenantId             int64      `json:"tenantId" gorm:"column:tenant_id;type:bigint"`                              // 租户ID
	AppId                *int64     `json:"appId" gorm:"column:app_id;type:bigint"`                                    // 应用/项目ID（可选）
	UserId               int64      `json:"userId" gorm:"column:user_id;type:bigint"`                                  // 用户ID
	AccountId            *string    `json:"accountId" gorm:"column:account_id;type:varchar(64)"`                       // 账号ID/外部账号标识（可选）
	OrderNo              string     `json:"orderNo" gorm:"column:order_no;type:varchar(64);uniqueIndex"`               // 平台提现订单号（唯一）
	MerchantOrderNo      *string    `json:"merchantOrderNo" gorm:"column:merchant_order_no;type:varchar(64)"`          // 商户侧订单号/前端单号（可选）
	Currency             string     `json:"currency" gorm:"column:currency;type:varchar(8);default:BRL"`               // 币种
	Amount               float64    `json:"amount" gorm:"column:amount;type:numeric(18,6)"`                            // 提现金额
	Fee                  float64    `json:"fee" gorm:"column:fee;type:numeric(18,6);default:0"`                        // 手续费
	NetAmount            float64    `json:"netAmount" gorm:"column:net_amount;type:numeric(18,6);->"`                  // 实际打款金额（只读）
	Channel              string     `json:"channel" gorm:"column:channel;type:varchar(32);default:pix"`                // 渠道
	PayMethod            *string    `json:"payMethod" gorm:"column:pay_method;type:varchar(32)"`                       // 方式/子渠道
	Status               int        `json:"status" gorm:"column:status;type:tinyint;default:0"`                        // 状态：0待审核 1待打款 2打款中 3成功 4失败 5取消 6退回
	ReviewedBy           *int64     `json:"reviewedBy" gorm:"column:reviewed_by;type:bigint"`                          // 审核人ID（后台）
	ReviewedAt           *time.Time `json:"reviewedAt" gorm:"column:reviewed_at;type:datetime(3)"`                     // 审核时间
	PaidAt               *time.Time `json:"paidAt" gorm:"column:paid_at;type:datetime(3)"`                             // 打款完成时间
	FailCode             *string    `json:"failCode" gorm:"column:fail_code;type:varchar(64)"`                         // 失败码（可选）
	FailMsg              *string    `json:"failMsg" gorm:"column:fail_msg;type:varchar(255)"`                          // 失败原因（可选）
	ReceiverName         *string    `json:"receiverName" gorm:"column:receiver_name;type:varchar(128)"`                // 收款人姓名（可选）
	ReceiverDocument     *string    `json:"receiverDocument" gorm:"column:receiver_document;type:varchar(32)"`         // 收款人证件号（可选）
	ReceiverDocumentType *string    `json:"receiverDocumentType" gorm:"column:receiver_document_type;type:varchar(8)"` // 证件类型
	PixKeyType           *string    `json:"pixKeyType" gorm:"column:pix_key_type;type:varchar(16)"`                    // Pix Key类型
	PixKey               *string    `json:"pixKey" gorm:"column:pix_key;type:varchar(128)"`                            // Pix Key（可选）
	BankCode             *string    `json:"bankCode" gorm:"column:bank_code;type:varchar(16)"`                         // 银行代码（可选）
	BankName             *string    `json:"bankName" gorm:"column:bank_name;type:varchar(64)"`                         // 银行名称（可选）
	BranchNumber         *string    `json:"branchNumber" gorm:"column:branch_number;type:varchar(16)"`                 // 支行号/Agency（可选）
	AccountNumber        *string    `json:"accountNumber" gorm:"column:account_number;type:varchar(32)"`               // 账号（可选）
	AccountType          *string    `json:"accountType" gorm:"column:account_type;type:varchar(16)"`                   // 账户类型（可选）
	Provider             *string    `json:"provider" gorm:"column:provider;type:varchar(32)"`                          // 代付服务商（可选）
	ProviderPayoutNo     *string    `json:"providerPayoutNo" gorm:"column:provider_payout_no;type:varchar(128)"`       // 三方代付单号/流水号（可选）
	ProviderStatus       *string    `json:"providerStatus" gorm:"column:provider_status;type:varchar(64)"`             // 三方状态原文（可选）
	NotifyTime           *time.Time `json:"notifyTime" gorm:"column:notify_time;type:datetime(3)"`                     // 回调/通知时间（可选）
	NotifyCount          int        `json:"notifyCount" gorm:"column:notify_count;type:int;default:0"`                 // 通知次数
	IdempotencyKey       *string    `json:"idempotencyKey" gorm:"column:idempotency_key;type:varchar(128)"`            // 幂等键（可选）
	RiskLevel            int        `json:"riskLevel" gorm:"column:risk_level;type:tinyint;default:0"`                 // 风控等级
	Remark               *string    `json:"remark" gorm:"column:remark;type:varchar(255)"`                             // 备注（可选）
	Extra                *string    `json:"extra" gorm:"column:extra;type:json"`                                       // 扩展信息（可选）
	DeletedAt            *time.Time `json:"deletedAt" gorm:"column:deleted_at;type:datetime(3)"`                       // 软删除时间（可选）
}

type WithdrawOrderBrSearch struct {
	PageInfo
	TenantId             int64  `json:"tenantId"`             // 租户ID
	UserId               int64  `json:"userId"`               // 用户ID
	Status               *int   `json:"status"`               // 状态
	OrderNo              string `json:"orderNo"`              // 平台订单号
	MerchantOrderNo      string `json:"merchantOrderNo"`      // 商户订单号
	ProviderPayoutNo     string `json:"providerPayoutNo"`     // 三方代付单号
	Channel              string `json:"channel"`              // 渠道
	PayMethod            string `json:"payMethod"`            // 方式/子渠道
	ReceiverDocumentType string `json:"receiverDocumentType"` // 证件类型
	ReceiverDocument     string `json:"receiverDocument"`     // 证件号
}

type WithdrawOrderBrSet struct {
	ID                   int64      `json:"id"` // ID
	TenantId             int64      `json:"tenantId"`
	AppId                *int64     `json:"appId"`
	UserId               int64      `json:"userId"`
	AccountId            *string    `json:"accountId"`
	OrderNo              string     `json:"orderNo"`
	MerchantOrderNo      *string    `json:"merchantOrderNo"`
	Currency             string     `json:"currency"`
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

type WithdrawOrderBrBack struct {
	ID                   int64      `json:"id"`
	CreatedAt            time.Time  `json:"createdAt"`
	UpdatedAt            time.Time  `json:"updatedAt"`
	DeletedAt            *time.Time `json:"deletedAt"`
	TenantId             int64      `json:"tenantId"`
	AppId                *int64     `json:"appId"`
	UserId               int64      `json:"userId"`
	AccountId            *string    `json:"accountId"`
	OrderNo              string     `json:"orderNo"`
	MerchantOrderNo      *string    `json:"merchantOrderNo"`
	Currency             string     `json:"currency"`
	Amount               float64    `json:"amount"`
	Fee                  float64    `json:"fee"`
	NetAmount            float64    `json:"netAmount"`
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

type WithdrawOrderBrResp struct {
	BasePageResponse[WithdrawOrderBrBack]
}

var WithdrawOrderBrTableName = "withdraw_order_br"

func (WithdrawOrderBr) TableName() string {
	return WithdrawOrderBrTableName
}
