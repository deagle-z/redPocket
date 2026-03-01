package pojo

import "time"

// LuckyMoneyItem 红包明细（子红包）
type LuckyMoneyItem struct {
	ID uint64 `json:"id" gorm:"primaryKey;autoIncrement;comment:主键ID"`

	RedPacketID uint64  `json:"redPacketId" gorm:"column:red_packet_id;not null;index:uk_red_packet_seq,unique,priority:1;index:idx_red_packet_grabbed,priority:1;comment:红包ID（主表ID）"`
	SeqNo       uint    `json:"seqNo" gorm:"column:seq_no;not null;index:uk_red_packet_seq,unique,priority:2;index:idx_red_packet_grabbed,priority:3;comment:序号（从1开始，红包内唯一）"`
	Amount      float64 `json:"amount" gorm:"type:decimal(18,2);not null;comment:金额"`

	IsGrabbed  int8       `json:"isGrabbed" gorm:"column:is_grabbed;type:tinyint(1);not null;default:0;index:idx_red_packet_grabbed,priority:2;comment:是否被抢：0否 1是"`
	Thunder    int8       `json:"thunder" gorm:"column:thunder;type:tinyint(1);not null;default:0;comment:是否中雷"`
	GrabbedUid *uint64    `json:"grabbedUid" gorm:"column:grabbed_uid;default:null;index:idx_grabbed_uid,priority:1;comment:抢到用户ID"`
	GrabbedAt  *time.Time `json:"grabbedAt" gorm:"column:grabbed_at;default:null;index:idx_grabbed_uid,priority:2;comment:抢到时间"`

	CreatedAt time.Time `json:"createdAt" gorm:"column:created_at;not null;autoCreateTime;comment:创建时间"`
	UpdatedAt time.Time `json:"updatedAt" gorm:"column:updated_at;not null;autoUpdateTime;comment:更新时间"`
}

type LuckyMoneyItemSearch struct {
	PageInfo
	RedPacketID uint64  `json:"redPacketId"`
	SeqNo       uint    `json:"seqNo"`
	IsGrabbed   *int8   `json:"isGrabbed"`
	GrabbedUid  *uint64 `json:"grabbedUid"`
}

type LuckyMoneyItemSet struct {
	ID          uint64     `json:"id"`
	RedPacketID uint64     `json:"redPacketId"`
	SeqNo       uint       `json:"seqNo"`
	Amount      float64    `json:"amount"`
	IsGrabbed   int8       `json:"isGrabbed"`
	Thunder     int8       `json:"thunder"`
	GrabbedUid  *uint64    `json:"grabbedUid"`
	GrabbedAt   *time.Time `json:"grabbedAt"`
}

type LuckyMoneyItemBack struct {
	ID          uint64     `json:"id"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
	RedPacketID uint64     `json:"redPacketId"`
	SeqNo       uint       `json:"seqNo"`
	Amount      float64    `json:"amount"`
	IsGrabbed   int8       `json:"isGrabbed"`
	Thunder     int8       `json:"thunder"`
	GrabbedUid  *uint64    `json:"grabbedUid"`
	GrabbedAt   *time.Time `json:"grabbedAt"`
}

type LuckyMoneyItemResp struct {
	BasePageResponse[LuckyMoneyItemBack]
}

var LuckyMoneyItemTableName = "lucky_money_item"

func (LuckyMoneyItem) TableName() string {
	return LuckyMoneyItemTableName
}
