package pojo

import (
	"time"
)

type AuthGroup struct {
	BaseModel
	GroupID            int64   `json:"groupId" gorm:"type:bigint;uniqueIndex"`         // Telegram群组ID
	GroupName          string  `json:"groupName" gorm:"type:varchar(128)"`             // 群组名称
	Status             int     `json:"status" gorm:"type:int;default:1"`               // 状态 1启用 0禁用
	ServiceURL         string  `json:"serviceUrl" gorm:"type:varchar(255)"`            // 客服链接
	RechargeURL        string  `json:"rechargeUrl" gorm:"type:varchar(255)"`           // 充值链接
	ChannelURL         string  `json:"channelUrl" gorm:"type:varchar(255)"`            // 玩法链接
	SendPacketImage    string  `json:"sendPacketImage" gorm:"type:varchar(255)"`       // 发红包图片url
	GrabPacketImage    string  `json:"grabPacketImage" gorm:"type:varchar(255)"`       // 抢红包图片url
	LoseRate           float64 `json:"loseRate" gorm:"type:decimal(10,2);default:1.8"` // 中雷倍数
	NumConfig          string  `json:"numConfig" gorm:"type:varchar(50);default:'3'"`  // 红包数量配置
	SendCommission     int     `json:"sendCommission" gorm:"type:int;default:2"`       // 发包中雷抽成百分比
	GrabbingCommission int     `json:"grabbingCommission" gorm:"type:int;default:3"`   // 抢红包抽成百分比
	DeleteMsg          int     `json:"deleteMsg" gorm:"type:int;default:0"`            //0 不开启消息删除 1 开启消息删除
	WhiteIds           string  `json:"whiteIds" gorm:"type:varchar(255)"`              //不删除消息白名单
}

type AuthGroupBack struct {
	ID                 int64     `json:"id"`
	CreatedAt          time.Time `json:"createdAt"`
	UpdatedAt          time.Time `json:"updatedAt"`
	GroupID            int64     `json:"groupId"`
	GroupName          string    `json:"groupName"`
	Status             int       `json:"status"`
	ServiceURL         string    `json:"serviceUrl"`
	RechargeURL        string    `json:"rechargeUrl"`
	ChannelURL         string    `json:"channelUrl"`
	SendPacketImage    string    `json:"sendPacketImage"`
	LoseRate           float64   `json:"loseRate"`
	NumConfig          string    `json:"numConfig"`
	SendCommission     int       `json:"sendCommission"`
	GrabbingCommission int       `json:"grabbingCommission"`
	DeleteMsg          int       `json:"deleteMsg"`
	WhiteIds           string    `json:"whiteIds"`
}

type AuthGroupSearch struct {
	PageInfo
	GroupID int64 `json:"groupId"` // 群组ID
	Status  *int  `json:"status"`  // 状态
}

type AuthGroupResp struct {
	BasePageResponse[AuthGroupBack]
}

var AuthGroupTableName = "auth_group"

func (AuthGroup) TableName() string {
	return AuthGroupTableName
}
