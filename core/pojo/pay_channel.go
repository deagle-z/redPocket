package pojo

import "time"

type PayChannel struct { // 三方支付通道配置表
	BaseModel
	TenantId            int64   `json:"tenantId" gorm:"column:tenant_id;type:bigint;default:0"`                              // 租户ID（0=平台公共）
	ChannelCode         string  `json:"channelCode" gorm:"column:channel_code;type:varchar(32)"`                             // 通道编码（唯一）
	ChannelName         string  `json:"channelName" gorm:"column:channel_name;type:varchar(64)"`                             // 通道名称
	ChannelType         string  `json:"channelType" gorm:"column:channel_type;type:varchar(32)"`                             // 通道类型：CARD/BANK/USDT/EWALLET/AGGREGATOR/OTHER
	ProviderName        string  `json:"providerName" gorm:"column:provider_name;type:varchar(64)"`                           // 三方提供商名称（展示用）
	Status              int     `json:"status" gorm:"column:status;type:tinyint;default:1"`                                  // 状态：0禁用 1启用
	SortNo              int     `json:"sortNo" gorm:"column:sort_no;type:int;default:0"`                                     // 排序号（越小越靠前）
	SupportedCurrencies string  `json:"supportedCurrencies" gorm:"column:supported_currencies;type:varchar(255);default:''"` // 支持币种（逗号分隔）
	SupportedCountries  string  `json:"supportedCountries" gorm:"column:supported_countries;type:varchar(255);default:''"`   // 支持国家/地区（逗号分隔）
	SupportedType       string  `json:"supportedType" gorm:"column:supported_type;type:varchar(255);default:''"`             // 支付方式 bank usdt
	MinAmount           float64 `json:"minAmount" gorm:"column:min_amount;type:numeric(18,2);default:0"`                     // 单笔最小金额
	MaxAmount           float64 `json:"maxAmount" gorm:"column:max_amount;type:numeric(18,2);default:0"`                     // 单笔最大金额（0表示不限制）
	FeeMode             int     `json:"feeMode" gorm:"column:fee_mode;type:tinyint;default:1"`                               // 手续费模式：1百分比 2固定 3百分比+固定
	FeeRate             float64 `json:"feeRate" gorm:"column:fee_rate;type:numeric(10,6);default:0"`                         // 手续费百分比
	FeeFixed            float64 `json:"feeFixed" gorm:"column:fee_fixed;type:numeric(18,2);default:0"`                       // 固定手续费
	WithdrawFeeMode     int     `json:"withdrawFeeMode" gorm:"column:withdraw_fee_mode;type:tinyint;default:1"`              // 提现手续费模式
	WithdrawFeeRate     float64 `json:"withdrawFeeRate" gorm:"column:withdraw_fee_rate;type:numeric(10,6);default:0"`        // 提现手续费百分比
	WithdrawFeeFixed    float64 `json:"withdrawFeeFixed" gorm:"column:withdraw_fee_fixed;type:numeric(18,2);default:0"`      // 提现固定手续费
	GatewayBaseURL      string  `json:"gatewayBaseUrl" gorm:"column:gateway_base_url;type:varchar(255);default:''"`          // 网关地址/接口前缀
	SignType            string  `json:"signType" gorm:"column:sign_type;type:varchar(32);default:''"`                        // 签名算法
	ApiKey              string  `json:"apiKey" gorm:"column:api_key;type:varchar(255);default:''"`                           // API Key
	ApiSecret           string  `json:"apiSecret" gorm:"column:api_secret;type:varchar(255);default:''"`                     // API Secret
	RsaPrivateKey       *string `json:"rsaPrivateKey" gorm:"column:rsa_private_key;type:text"`                               // RSA私钥
	RsaPublicKey        *string `json:"rsaPublicKey" gorm:"column:rsa_public_key;type:text"`                                 // RSA公钥/平台公钥
	NotifyURL           string  `json:"notifyUrl" gorm:"column:notify_url;type:varchar(255);default:''"`                     // 支付回调通知URL（服务端）
	ReturnURL           string  `json:"returnUrl" gorm:"column:return_url;type:varchar(255);default:''"`                     // 前端同步跳转URL（可选）
	ExtraConfig         *string `json:"extraConfig" gorm:"column:extra_config;type:json"`                                    // 扩展配置JSON
	Remark              string  `json:"remark" gorm:"column:remark;type:varchar(255);default:''"`                            // 备注
}

type PayChannelSearch struct {
	PageInfo
	TenantId     int64  `json:"tenantId"`     // 租户ID
	ChannelCode  string `json:"channelCode"`  // 通道编码
	ChannelName  string `json:"channelName"`  // 通道名称
	ChannelType  string `json:"channelType"`  // 通道类型
	ProviderName string `json:"providerName"` // 提供商名称
	Status       *int   `json:"status"`       // 状态
}

type PayChannelSet struct {
	ID                  int64   `json:"id"` // ID
	TenantId            int64   `json:"tenantId"`
	ChannelCode         string  `json:"channelCode"`
	ChannelName         string  `json:"channelName"`
	ChannelType         string  `json:"channelType"`
	ProviderName        string  `json:"providerName"`
	Status              int     `json:"status"`
	SortNo              int     `json:"sortNo"`
	SupportedCurrencies string  `json:"supportedCurrencies"`
	SupportedCountries  string  `json:"supportedCountries"`
	SupportedType       string  `json:"supportedType"`
	MinAmount           float64 `json:"minAmount"`
	MaxAmount           float64 `json:"maxAmount"`
	FeeMode             int     `json:"feeMode"`
	FeeRate             float64 `json:"feeRate"`
	FeeFixed            float64 `json:"feeFixed"`
	WithdrawFeeMode     int     `json:"withdrawFeeMode"`
	WithdrawFeeRate     float64 `json:"withdrawFeeRate"`
	WithdrawFeeFixed    float64 `json:"withdrawFeeFixed"`
	GatewayBaseURL      string  `json:"gatewayBaseUrl"`
	SignType            string  `json:"signType"`
	ApiKey              string  `json:"apiKey"`
	ApiSecret           string  `json:"apiSecret"`
	RsaPrivateKey       *string `json:"rsaPrivateKey"`
	RsaPublicKey        *string `json:"rsaPublicKey"`
	NotifyURL           string  `json:"notifyUrl"`
	ReturnURL           string  `json:"returnUrl"`
	ExtraConfig         *string `json:"extraConfig"`
	Remark              string  `json:"remark"`
}

type PayChannelBack struct {
	CreatedAt           time.Time `json:"createdAt"`
	UpdatedAt           time.Time `json:"updatedAt"`
	ID                  int64     `json:"id"`
	TenantId            int64     `json:"tenantId"`
	ChannelCode         string    `json:"channelCode"`
	ChannelName         string    `json:"channelName"`
	ChannelType         string    `json:"channelType"`
	ProviderName        string    `json:"providerName"`
	Status              int       `json:"status"`
	SortNo              int       `json:"sortNo"`
	SupportedCurrencies string    `json:"supportedCurrencies"`
	SupportedCountries  string    `json:"supportedCountries"`
	SupportedType       string    `json:"supportedType"`
	MinAmount           float64   `json:"minAmount"`
	MaxAmount           float64   `json:"maxAmount"`
	FeeMode             int       `json:"feeMode"`
	FeeRate             float64   `json:"feeRate"`
	FeeFixed            float64   `json:"feeFixed"`
	WithdrawFeeMode     int       `json:"withdrawFeeMode"`
	WithdrawFeeRate     float64   `json:"withdrawFeeRate"`
	WithdrawFeeFixed    float64   `json:"withdrawFeeFixed"`
	GatewayBaseURL      string    `json:"gatewayBaseUrl"`
	SignType            string    `json:"signType"`
	ApiKey              string    `json:"apiKey"`
	ApiSecret           string    `json:"apiSecret"`
	RsaPrivateKey       *string   `json:"rsaPrivateKey"`
	RsaPublicKey        *string   `json:"rsaPublicKey"`
	NotifyURL           string    `json:"notifyUrl"`
	ReturnURL           string    `json:"returnUrl"`
	ExtraConfig         *string   `json:"extraConfig"`
	Remark              string    `json:"remark"`
}

type PayChannelResp struct {
	BasePageResponse[PayChannelBack]
}

var PayChannelTableName = "pay_channel"

func (PayChannel) TableName() string {
	return PayChannelTableName
}
