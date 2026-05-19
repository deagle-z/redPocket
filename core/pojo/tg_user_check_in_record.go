package pojo

import "time"

// TgUserCheckInRecord 用户签到记录表。
type TgUserCheckInRecord struct {
	ID            int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	TenantId      int64     `json:"tenantId" gorm:"column:tenant_id;type:bigint;not null;index"`
	UserId        int64     `json:"userId" gorm:"column:user_id;type:bigint;not null;index;uniqueIndex:uk_check_in_user_date,priority:1"`
	CheckInDate   string    `json:"checkInDate" gorm:"column:check_in_date;type:date;not null;index;uniqueIndex:uk_check_in_user_date,priority:2"`
	CheckInSeq    int       `json:"checkInSeq" gorm:"column:check_in_seq;type:int;not null"`
	RewardAmount  float64   `json:"rewardAmount" gorm:"column:reward_amount;type:decimal(20,2);not null;default:0"`
	BeforeBalance float64   `json:"beforeBalance" gorm:"column:before_balance;type:decimal(20,2);not null;default:0"`
	AfterBalance  float64   `json:"afterBalance" gorm:"column:after_balance;type:decimal(20,2);not null;default:0"`
	CreatedAt     time.Time `json:"createdAt" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt     time.Time `json:"updatedAt" gorm:"column:updated_at;autoUpdateTime"`
}

func (TgUserCheckInRecord) TableName() string {
	return "tg_user_check_in_record"
}

type TgUserCheckInStatusBack struct {
	TodayChecked     bool      `json:"todayChecked"`
	TotalCheckInDays int64     `json:"totalCheckInDays"`
	NextSeq          int       `json:"nextSeq"`
	NextRewardAmount float64   `json:"nextRewardAmount"`
	Rewards          []float64 `json:"rewards"`
	Completed        bool      `json:"completed"`
	Timezone         string    `json:"timezone"`
}

type TgUserCheckInBack struct {
	RecordId     int64   `json:"recordId"`
	CheckInSeq   int     `json:"checkInSeq"`
	RewardAmount float64 `json:"rewardAmount"`
	Balance      float64 `json:"balance"`
	TodayChecked bool    `json:"todayChecked"`
	CheckInDate  string  `json:"checkInDate"`
}

type TgUserCheckInRecordBack struct {
	ID            int64     `json:"id"`
	TenantId      int64     `json:"tenantId"`
	UserId        int64     `json:"userId"`
	UserUid       string    `json:"userUid"`
	CheckInDate   string    `json:"checkInDate"`
	CheckInSeq    int       `json:"checkInSeq"`
	RewardAmount  float64   `json:"rewardAmount"`
	BeforeBalance float64   `json:"beforeBalance"`
	AfterBalance  float64   `json:"afterBalance"`
	CreatedAt     time.Time `json:"createdAt"`
}

type TgUserCheckInRecordSearch struct {
	PageInfo
	TenantId  int64  `json:"tenantId"`
	UserId    int64  `json:"userId"`
	UserUid   string `json:"userUid"`
	StartDate string `json:"startDate"`
	EndDate   string `json:"endDate"`
}

type TgUserCheckInRecordResp struct {
	BasePageResponse[TgUserCheckInRecordBack]
}
