package pojo

import "time"

// TgChannelMember 监听到的 Telegram 指定频道成员。
type TgChannelMember struct {
	BaseModel
	ChannelID  string     `gorm:"column:channel_id;type:varchar(64);not null;uniqueIndex:uk_tg_channel_member_channel_tg,priority:1;index;comment:频道ID" json:"channelId"`
	TgID       int64      `gorm:"column:tg_id;type:bigint;not null;uniqueIndex:uk_tg_channel_member_channel_tg,priority:2;index;comment:Telegram用户ID" json:"tgId"`
	TgName     string     `gorm:"column:tg_name;size:64;not null;default:'';index;comment:Telegram @用户名" json:"tgName"`
	FirstName  string     `gorm:"column:first_name;size:128;not null;default:'';comment:Telegram展示名" json:"firstName"`
	Status     int8       `gorm:"column:status;type:tinyint;not null;default:1;index;comment:1=在频道 0=已离开" json:"status"`
	BindUserID *int64     `gorm:"column:bind_user_id;type:bigint;uniqueIndex;comment:绑定的tg_user.id" json:"bindUserId"`
	JoinedAt   *time.Time `gorm:"column:joined_at;comment:加入时间" json:"joinedAt"`
	LeftAt     *time.Time `gorm:"column:left_at;comment:离开时间" json:"leftAt"`
}

var TgChannelMemberTableName = "tg_channel_member"

func (TgChannelMember) TableName() string {
	return TgChannelMemberTableName
}

type TgBindChannelNameReq struct {
	TgName string `json:"tgName" binding:"required"`
}

type TgBindChannelNameBack struct {
	TgName           string `json:"tgName"`
	FreeLotteryCount int    `json:"freeLotteryCount"`
	AwardedCount     int    `json:"awardedCount"`
}
