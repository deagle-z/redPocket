package pojo

import "time"

const (
	TaskActivityTypeInviteRecharge = "invite_recharge"

	TaskActivityLevelPrimary  = "primary"
	TaskActivityLevelMiddle   = "middle"
	TaskActivityLevelAdvanced = "advanced"

	TaskActivityRewardTargetRebate = "rebate"
)

type TgTaskActivityConfig struct {
	BaseModel
	Title                  string     `json:"title" gorm:"column:title;type:varchar(128);not null;default:''"`
	TitleI18n              *string    `json:"titleI18n" gorm:"column:title_i18n;type:json"`
	SubTitle               string     `json:"subTitle" gorm:"column:sub_title;type:varchar(255);not null;default:''"`
	SubTitleI18n           *string    `json:"subTitleI18n" gorm:"column:sub_title_i18n;type:json"`
	LevelCode              string     `json:"levelCode" gorm:"column:level_code;type:varchar(32);not null;default:'';index"`
	LevelName              string     `json:"levelName" gorm:"column:level_name;type:varchar(64);not null;default:''"`
	Sort                   int        `json:"sort" gorm:"column:sort;type:int;not null;default:0;index"`
	Status                 int8       `json:"status" gorm:"column:status;type:tinyint;not null;default:1;index"`
	TaskType               string     `json:"taskType" gorm:"column:task_type;type:varchar(32);not null;default:invite_recharge;index"`
	RequiredInviteCount    int        `json:"requiredInviteCount" gorm:"column:required_invite_count;type:int;not null;default:0"`
	RequiredRechargeAmount float64    `json:"requiredRechargeAmount" gorm:"column:required_recharge_amount;type:decimal(20,2);not null;default:0"`
	DurationMinutes        int        `json:"durationMinutes" gorm:"column:duration_minutes;type:int;not null;default:0"`
	RewardAmount           float64    `json:"rewardAmount" gorm:"column:reward_amount;type:decimal(20,2);not null;default:0"`
	RewardCurrency         string     `json:"rewardCurrency" gorm:"column:reward_currency;type:varchar(16);not null;default:USD"`
	RewardTarget           string     `json:"rewardTarget" gorm:"column:reward_target;type:varchar(32);not null;default:rebate"`
	StartAt                *time.Time `json:"startAt" gorm:"column:start_at;type:datetime(3)"`
	EndAt                  *time.Time `json:"endAt" gorm:"column:end_at;type:datetime(3)"`
	Remark                 string     `json:"remark" gorm:"column:remark;type:varchar(255);not null;default:''"`
}

type TgTaskActivityConfigSearch struct {
	PageInfo
	Title     string `json:"title"`
	LevelCode string `json:"levelCode"`
	Status    *int8  `json:"status"`
}

type TgTaskActivityConfigSet struct {
	ID                     int64      `json:"id"`
	Title                  string     `json:"title"`
	TitleI18n              any        `json:"titleI18n"`
	SubTitle               string     `json:"subTitle"`
	SubTitleI18n           any        `json:"subTitleI18n"`
	LevelCode              string     `json:"levelCode"`
	LevelName              string     `json:"levelName"`
	Sort                   int        `json:"sort"`
	Status                 int8       `json:"status"`
	TaskType               string     `json:"taskType"`
	RequiredInviteCount    int        `json:"requiredInviteCount"`
	RequiredRechargeAmount float64    `json:"requiredRechargeAmount"`
	DurationMinutes        int        `json:"durationMinutes"`
	RewardAmount           float64    `json:"rewardAmount"`
	RewardCurrency         string     `json:"rewardCurrency"`
	RewardTarget           string     `json:"rewardTarget"`
	StartAt                *time.Time `json:"startAt"`
	EndAt                  *time.Time `json:"endAt"`
	Remark                 string     `json:"remark"`
}

type TgTaskActivityConfigBack struct {
	ID                     int64      `json:"id"`
	CreatedAt              time.Time  `json:"createdAt"`
	UpdatedAt              time.Time  `json:"updatedAt"`
	Title                  string     `json:"title"`
	TitleI18n              any        `json:"titleI18n"`
	SubTitle               string     `json:"subTitle"`
	SubTitleI18n           any        `json:"subTitleI18n"`
	LevelCode              string     `json:"levelCode"`
	LevelName              string     `json:"levelName"`
	Sort                   int        `json:"sort"`
	Status                 int8       `json:"status"`
	TaskType               string     `json:"taskType"`
	RequiredInviteCount    int        `json:"requiredInviteCount"`
	RequiredRechargeAmount float64    `json:"requiredRechargeAmount"`
	DurationMinutes        int        `json:"durationMinutes"`
	RewardAmount           float64    `json:"rewardAmount"`
	RewardCurrency         string     `json:"rewardCurrency"`
	RewardTarget           string     `json:"rewardTarget"`
	StartAt                *time.Time `json:"startAt"`
	EndAt                  *time.Time `json:"endAt"`
	Remark                 string     `json:"remark"`
}

type TgTaskActivityConfigResp struct {
	BasePageResponse[TgTaskActivityConfigBack]
}

var TgTaskActivityConfigTableName = "tg_task_activity_config"

func (TgTaskActivityConfig) TableName() string {
	return TgTaskActivityConfigTableName
}
