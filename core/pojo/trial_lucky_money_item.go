package pojo

import "time"

type TrialLuckyMoneyItem struct {
	ID        uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	CreatedAt time.Time `json:"createdAt" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time `json:"updatedAt" gorm:"column:updated_at;autoUpdateTime"`

	RedPacketID int64      `json:"redPacketId" gorm:"type:bigint;not null;index"`
	SeqNo       uint       `json:"seqNo" gorm:"type:int unsigned;not null;index"`
	Amount      float64    `json:"amount" gorm:"type:decimal(18,2);not null;default:0.00"`
	IsGrabbed   int8       `json:"isGrabbed" gorm:"type:tinyint;not null;default:0;index"`
	Thunder     int8       `json:"thunder" gorm:"type:tinyint;not null;default:0"`
	GrabbedUid  *int64     `json:"grabbedUid" gorm:"type:bigint;index"`
	GrabbedType string     `json:"grabbedType" gorm:"type:varchar(16);not null;default:'user';index"`
	GrabbedAt   *time.Time `json:"grabbedAt" gorm:"column:grabbed_at"`
}

var TrialLuckyMoneyItemTableName = "trial_lucky_money_item"

func (TrialLuckyMoneyItem) TableName() string {
	return TrialLuckyMoneyItemTableName
}
