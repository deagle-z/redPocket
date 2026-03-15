package pojo

import "time"

type SysPayChannelMethod struct {
	ID        int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	ChannelID int64     `json:"channelId" gorm:"column:channel_id;type:bigint unsigned;not null;uniqueIndex:uk_channel_method;comment:支付通道ID"`
	MethodID  int64     `json:"methodId" gorm:"column:method_id;type:bigint unsigned;not null;uniqueIndex:uk_channel_method;comment:支付方式ID"`
	CreatedAt time.Time `json:"createdAt" gorm:"column:created_at;autoCreateTime"`
}

type SysPayChannelMethodSet struct {
	ChannelID int64 `json:"channelId"`
	MethodID  int64 `json:"methodId"`
}

var SysPayChannelMethodTableName = "sys_pay_channel_method"

func (SysPayChannelMethod) TableName() string {
	return SysPayChannelMethodTableName
}
