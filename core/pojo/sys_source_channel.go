package pojo

import "time"

// SysSourceChannel 投流来源渠道
type SysSourceChannel struct {
	BaseModel
	TenantId    int64   `json:"tenantId" gorm:"column:tenant_id;type:bigint;not null;default:0;index;comment:租户ID，0=平台公共"`
	ChannelCode string  `json:"channelCode" gorm:"column:channel_code;type:varchar(64);uniqueIndex;not null;index:idx_source_channel_code_status;comment:渠道编码"`
	ChannelName string  `json:"channelName" gorm:"column:channel_name;type:varchar(128);not null;comment:渠道名称"`
	ParentID    *int64  `json:"parentId" gorm:"column:parent_id;type:bigint;index;comment:父渠道ID"`
	Level       int8    `json:"level" gorm:"column:level;type:tinyint;not null;default:1;comment:渠道层级 1主渠道 2子渠道"`
	Status      int8    `json:"status" gorm:"column:status;type:tinyint;not null;default:1;index:idx_source_channel_code_status;comment:状态 1启用 0禁用"`
	Sort        int     `json:"sort" gorm:"column:sort;type:int;not null;default:0;comment:排序值，越小越靠前"`
	Remark      *string `json:"remark" gorm:"column:remark;type:varchar(255);comment:备注"`
}

type SysSourceChannelSearch struct {
	PageInfo
	TenantId    int64  `json:"tenantId"`
	ChannelCode string `json:"channelCode"`
	ChannelName string `json:"channelName"`
	ParentID    *int64 `json:"parentId"`
	Level       *int8  `json:"level"`
	Status      *int8  `json:"status"`
}

type SysSourceChannelSet struct {
	ID          int64   `json:"id"`
	TenantId    int64   `json:"tenantId"`
	ChannelCode string  `json:"channelCode"`
	ChannelName string  `json:"channelName"`
	ParentID    *int64  `json:"parentId"`
	Level       int8    `json:"level"`
	Status      int8    `json:"status"`
	Sort        int     `json:"sort"`
	Remark      *string `json:"remark"`
}

type SysSourceChannelBack struct {
	ID          int64                     `json:"id"`
	CreatedAt   time.Time                 `json:"createdAt"`
	UpdatedAt   time.Time                 `json:"updatedAt"`
	TenantId    int64                     `json:"tenantId"`
	ChannelCode string                    `json:"channelCode"`
	ChannelName string                    `json:"channelName"`
	ParentID    *int64                    `json:"parentId"`
	Level       int8                      `json:"level"`
	Status      int8                      `json:"status"`
	Sort        int                       `json:"sort"`
	Remark      *string                   `json:"remark"`
	LinkURL     string                    `json:"linkUrl"`
	Stats       SysSourceChannelStatsBack `json:"stats"`
}

type SysSourceChannelStatsBack struct {
	RegisterUsers       int64   `json:"registerUsers"`
	TodayRegisterUsers  int64   `json:"todayRegisterUsers"`
	RechargeUsers       int64   `json:"rechargeUsers"`
	TodayRechargeUsers  int64   `json:"todayRechargeUsers"`
	TotalRechargeAmount float64 `json:"totalRechargeAmount"`
	TodayRechargeAmount float64 `json:"todayRechargeAmount"`
	TotalWithdrawAmount float64 `json:"totalWithdrawAmount"`
	TodayWithdrawAmount float64 `json:"todayWithdrawAmount"`
	EventCount          int64   `json:"eventCount"`
	UniqueVisitors      int64   `json:"uniqueVisitors"`
	UniqueUsers         int64   `json:"uniqueUsers"`
}

type SysSourceChannelResp struct {
	BasePageResponse[SysSourceChannelBack]
}

var SysSourceChannelTableName = "sys_source_channel"

func (SysSourceChannel) TableName() string {
	return SysSourceChannelTableName
}
