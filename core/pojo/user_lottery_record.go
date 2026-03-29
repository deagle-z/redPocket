package pojo

import "time"

// UserLotteryRecord 用户抽奖记录表
type UserLotteryRecord struct {
	ID            int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	TenantId      int64     `json:"tenantId" gorm:"type:bigint unsigned;not null;index"`
	UserId        int64     `json:"userId" gorm:"type:bigint unsigned;not null;index"`
	PoolId        int64     `json:"poolId" gorm:"type:bigint unsigned;not null;index"`
	ConfigId      int64     `json:"configId" gorm:"type:bigint unsigned;not null"`          // 概率配置ID（冗余快照）
	PeerAmount    float64   `json:"peerAmount" gorm:"type:decimal(10,2);not null;default:0"` // 本次消耗金额
	AwardAmount   float64   `json:"awardAmount" gorm:"type:decimal(10,2);not null;default:0"` // 中奖金额（0=未中奖）
	BeforeBalance float64   `json:"beforeBalance" gorm:"type:decimal(18,6);not null;default:0"` // 抽奖前用户余额
	AfterBalance  float64   `json:"afterBalance" gorm:"type:decimal(18,6);not null;default:0"`  // 抽奖后用户余额
	Status        int8      `json:"status" gorm:"type:tinyint;not null;default:0"`              // 0待结算 1已发放 2未中奖
	Remark        *string   `json:"remark" gorm:"type:varchar(255)"`
	CreatedAt     time.Time `json:"createdAt" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt     time.Time `json:"updatedAt" gorm:"column:updated_at;autoUpdateTime"`
}

func (UserLotteryRecord) TableName() string {
	return "user_lottery_record"
}

const (
	LotteryStatusPending  int8 = 0 // 待结算
	LotteryStatusAwarded  int8 = 1 // 已发放
	LotteryStatusNoAward  int8 = 2 // 未中奖
)

// ---- 请求/响应 DTO ----

type UserLotteryRecordSearch struct {
	PageInfo
	TenantId int64  `json:"tenantId"`
	UserId   int64  `json:"userId"`
	PoolId   int64  `json:"poolId"`
	Status   *int8  `json:"status"`
	StartTime int64 `json:"startTime"` // unix 秒
	EndTime   int64 `json:"endTime"`   // unix 秒
}

type UserLotteryRecordResp struct {
	BasePageResponse[UserLotteryRecord]
}
