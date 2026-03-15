package pojo

import "time"

type SysPayChannel struct {
	BaseModel
	ChannelCode  string  `json:"channelCode" gorm:"column:channel_code;type:varchar(64);uniqueIndex;not null;comment:通道编码，如 cashfree / paytm / razorpay"`
	ChannelName  string  `json:"channelName" gorm:"column:channel_name;type:varchar(128);not null;comment:通道名称"`
	ChannelType  string  `json:"channelType" gorm:"column:channel_type;type:varchar(32);not null;index;comment:通道类型：deposit=充值 withdraw=提现 both=充值+提现"`
	ProviderType string  `json:"providerType" gorm:"column:provider_type;type:varchar(32);not null;default:'third_party';comment:提供方类型：third_party=三方 native=自有"`
	CountryCode  *string `json:"countryCode" gorm:"column:country_code;type:varchar(16);comment:适用国家码，如 BR / IN / PH"`
	Icon         *string `json:"icon" gorm:"column:icon;type:varchar(255);comment:图标"`
	Remark       *string `json:"remark" gorm:"column:remark;type:varchar(255);comment:备注"`
	Sort         int     `json:"sort" gorm:"column:sort;type:int;not null;default:0;comment:排序值，越小越靠前"`
	Status       int8    `json:"status" gorm:"column:status;type:tinyint;not null;default:1;comment:状态 1启用 0禁用"`
	DeletedAt    int64   `json:"deletedAt" gorm:"column:deleted_at;type:bigint;not null;default:0;comment:软删除时间戳，0=未删除"`
}

type SysPayChannelSearch struct {
	PageInfo
	ChannelCode  string `json:"channelCode"`
	ChannelName  string `json:"channelName"`
	ChannelType  string `json:"channelType"`
	ProviderType string `json:"providerType"`
	CountryCode  string `json:"countryCode"`
	Status       *int8  `json:"status"`
}

type SysPayChannelSet struct {
	ID           int64   `json:"id"`
	ChannelCode  string  `json:"channelCode"`
	ChannelName  string  `json:"channelName"`
	ChannelType  string  `json:"channelType"`
	ProviderType string  `json:"providerType"`
	CountryCode  *string `json:"countryCode"`
	Icon         *string `json:"icon"`
	Remark       *string `json:"remark"`
	Sort         int     `json:"sort"`
	Status       int8    `json:"status"`
}

type SysPayChannelBack struct {
	ID           int64     `json:"id"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
	ChannelCode  string    `json:"channelCode"`
	ChannelName  string    `json:"channelName"`
	ChannelType  string    `json:"channelType"`
	ProviderType string    `json:"providerType"`
	CountryCode  *string   `json:"countryCode"`
	Icon         *string   `json:"icon"`
	Remark       *string   `json:"remark"`
	Sort         int       `json:"sort"`
	Status       int8      `json:"status"`
}

type SysPayChannelResp struct {
	BasePageResponse[SysPayChannelBack]
}

var SysPayChannelTableName = "sys_pay_channel"

func (SysPayChannel) TableName() string {
	return SysPayChannelTableName
}
