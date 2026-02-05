package pojo

import (
	"time"
)

// TgUser Telegram 用户表
type TgUser struct {
	ID        int64          `gorm:"primaryKey;autoIncrement;comment:自增主键" json:"id"`
	CreatedAt time.Time      `gorm:"column:created_at;comment:创建时间" json:"created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at;comment:更新时间" json:"updated_at"`

	Username  *string `gorm:"size:64;comment:Telegram 用户名（不保证唯一，用户可修改）" json:"username"`
	FirstName *string `gorm:"size:128;comment:Telegram 名（first_name / 展示名）" json:"first_name"`

	Avatar *string `gorm:"size:1024;comment:头像URL" json:"avatar"`

	TgID int64 `gorm:"column:tg_id;uniqueIndex;not null;comment:Telegram 用户ID（唯一且稳定）" json:"tg_id"`

	Balance    float64 `gorm:"type:decimal(20,3);not null;default:0.000;comment:账户可用余额" json:"balance"`
	GiftAmount float64 `gorm:"type:decimal(20,3);not null;default:0.000;comment:赠送余额（可用）" json:"gift_amount"`
	GiftTotal  float64 `gorm:"type:decimal(20,3);not null;default:0.000;comment:累计赠送金额" json:"gift_total"`
	RebateAmount      float64 `gorm:"type:decimal(20,3);not null;default:0.000;comment:可用返水余额" json:"rebate_amount"`
	RebateTotalAmount float64 `gorm:"type:decimal(20,3);not null;default:0.000;comment:累计返水金额" json:"rebate_total_amount"`

	Status int8 `gorm:"not null;default:1;index;comment:状态 1=正常 0=禁用 -1=删除" json:"status"`

	ParentID   *int64  `gorm:"index;comment:上级/邀请人用户ID" json:"parent_id"`
	InviteCode *string `gorm:"size:32;index;comment:邀请码（用户自身的邀请码）" json:"invite_code"`
}

type TgUserSearch struct {
	PageInfo
	TgID       int64  `json:"tgId"`       // Telegram用户ID
	Username   string `json:"username"`   // Telegram用户名
	FirstName  string `json:"firstName"`  // 展示名
	Status     *int8  `json:"status"`     // 状态
	ParentID   *int64 `json:"parentId"`   // 上级/邀请人用户ID
	InviteCode string `json:"inviteCode"` // 邀请码
}

type TgUserSet struct {
	ID         int64   `json:"id"` // ID
	Username   *string `json:"username"`
	FirstName  *string `json:"firstName"`
	Avatar     *string `json:"avatar"`
	TgID       int64   `json:"tgId"`
	Balance    float64 `json:"balance"`
	GiftAmount float64 `json:"giftAmount"`
	GiftTotal  float64 `json:"giftTotal"`
	Status     int8    `json:"status"`
	ParentID   *int64  `json:"parentId"`
	InviteCode *string `json:"inviteCode"`
}

type TgUserStatusSet struct {
	ID     int64 `json:"id"`
	Status int8  `json:"status"` // 1=正常 0=禁用 -1=删除
}

type TgUserBack struct {
	ID         int64     `json:"id"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
	Username   *string   `json:"username"`
	FirstName  *string   `json:"firstName"`
	Avatar     *string   `json:"avatar"`
	TgID       int64     `json:"tgId"`
	Balance    float64   `json:"balance"`
	GiftAmount float64   `json:"giftAmount"`
	GiftTotal  float64   `json:"giftTotal"`
	Status     int8      `json:"status"`
	ParentID   *int64    `json:"parentId"`
	InviteCode *string   `json:"inviteCode"`
}

type TgUserResp struct {
	BasePageResponse[TgUserBack]
}

var TgUserTableName = "tg_user"

// TableName 指定表名
func (TgUser) TableName() string {
	return TgUserTableName
}
