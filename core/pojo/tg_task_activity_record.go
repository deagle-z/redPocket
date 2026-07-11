package pojo

import "time"

const (
	TaskActivityRecordStatusProgress  = 0 // 进行中
	TaskActivityRecordStatusCompleted = 1 // 已达标待发放
	TaskActivityRecordStatusRewarded  = 2 // 已发放
	TaskActivityRecordStatusExpired   = 3 // 已过期
	TaskActivityRecordStatusCanceled  = 4 // 已取消
)

// TgTaskActivityRecord 用户领取任务活动后的进度记录
type TgTaskActivityRecord struct {
	BaseModel
	ConfigID               int64      `json:"configId" gorm:"column:config_id;type:bigint;not null;index;index:idx_task_activity_user_config,priority:2"`
	UserID                 int64      `json:"userId" gorm:"column:user_id;type:bigint;not null;index;index:idx_task_activity_user_config,priority:1"`
	Title                  string     `json:"title" gorm:"column:title;type:varchar(128);not null;default:''"`
	TitleI18n              *string    `json:"titleI18n" gorm:"column:title_i18n;type:json"`
	SubTitle               string     `json:"subTitle" gorm:"column:sub_title;type:varchar(255);not null;default:''"`
	SubTitleI18n           *string    `json:"subTitleI18n" gorm:"column:sub_title_i18n;type:json"`
	LevelCode              string     `json:"levelCode" gorm:"column:level_code;type:varchar(32);not null;default:'';index"`
	LevelName              string     `json:"levelName" gorm:"column:level_name;type:varchar(64);not null;default:''"`
	TaskType               string     `json:"taskType" gorm:"column:task_type;type:varchar(32);not null;default:invite_recharge;index"`
	RequiredInviteCount    int        `json:"requiredInviteCount" gorm:"column:required_invite_count;type:int;not null;default:0"`
	RequiredRechargeAmount float64    `json:"requiredRechargeAmount" gorm:"column:required_recharge_amount;type:decimal(20,2);not null;default:0"`
	DurationMinutes        int        `json:"durationMinutes" gorm:"column:duration_minutes;type:int;not null;default:0"`
	RewardAmount           float64    `json:"rewardAmount" gorm:"column:reward_amount;type:decimal(20,2);not null;default:0"`
	RewardCurrency         string     `json:"rewardCurrency" gorm:"column:reward_currency;type:varchar(16);not null;default:USD"`
	RewardTarget           string     `json:"rewardTarget" gorm:"column:reward_target;type:varchar(32);not null;default:rebate"`
	ProgressInviteCount    int        `json:"progressInviteCount" gorm:"column:progress_invite_count;type:int;not null;default:0"`
	ProgressRechargeCount  int        `json:"progressRechargeCount" gorm:"column:progress_recharge_count;type:int;not null;default:0"`
	ProgressRechargeAmount float64    `json:"progressRechargeAmount" gorm:"column:progress_recharge_amount;type:decimal(20,2);not null;default:0"`
	Status                 int8       `json:"status" gorm:"column:status;type:tinyint;not null;default:0;index"`
	ClaimedAt              time.Time  `json:"claimedAt" gorm:"column:claimed_at;type:datetime(3);not null;index"`
	DeadlineAt             time.Time  `json:"deadlineAt" gorm:"column:deadline_at;type:datetime(3);not null;index"`
	CompletedAt            *time.Time `json:"completedAt" gorm:"column:completed_at;type:datetime(3)"`
	RewardedAt             *time.Time `json:"rewardedAt" gorm:"column:rewarded_at;type:datetime(3)"`
	RebateRecordID         *int64     `json:"rebateRecordId" gorm:"column:rebate_record_id;type:bigint;index"`
	Remark                 string     `json:"remark" gorm:"column:remark;type:varchar(255);not null;default:''"`
}

type TgTaskActivityRecordSearch struct {
	PageInfo
	UserID    int64  `json:"userId"`
	Uid       string `json:"uid"`
	ConfigID  int64  `json:"configId"`
	Title     string `json:"title"`
	LevelCode string `json:"levelCode"`
	Status    *int8  `json:"status"`
}

type TgTaskActivityRecordBack struct {
	ID                     int64      `json:"id"`
	CreatedAt              time.Time  `json:"createdAt"`
	UpdatedAt              time.Time  `json:"updatedAt"`
	ConfigID               int64      `json:"configId"`
	UserID                 int64      `json:"userId"`
	Uid                    string     `json:"uid"`
	Title                  string     `json:"title"`
	TitleI18n              any        `json:"titleI18n"`
	SubTitle               string     `json:"subTitle"`
	SubTitleI18n           any        `json:"subTitleI18n"`
	LevelCode              string     `json:"levelCode"`
	LevelName              string     `json:"levelName"`
	TaskType               string     `json:"taskType"`
	RequiredInviteCount    int        `json:"requiredInviteCount"`
	RequiredRechargeAmount float64    `json:"requiredRechargeAmount"`
	DurationMinutes        int        `json:"durationMinutes"`
	RewardAmount           float64    `json:"rewardAmount"`
	RewardCurrency         string     `json:"rewardCurrency"`
	RewardTarget           string     `json:"rewardTarget"`
	ProgressInviteCount    int        `json:"progressInviteCount"`
	ProgressRechargeCount  int        `json:"progressRechargeCount"`
	ProgressRechargeAmount float64    `json:"progressRechargeAmount"`
	Status                 int8       `json:"status"`
	ClaimedAt              time.Time  `json:"claimedAt"`
	DeadlineAt             time.Time  `json:"deadlineAt"`
	CompletedAt            *time.Time `json:"completedAt"`
	RewardedAt             *time.Time `json:"rewardedAt"`
	RebateRecordID         *int64     `json:"rebateRecordId"`
	Remark                 string     `json:"remark"`
}

type TgTaskActivityRecordResp struct {
	BasePageResponse[TgTaskActivityRecordBack]
}

type TgTaskActivityAppItem struct {
	ConfigID               int64      `json:"configId"`
	RecordID               int64      `json:"recordId"`
	Title                  string     `json:"title"`
	SubTitle               string     `json:"subTitle"`
	LevelCode              string     `json:"levelCode"`
	LevelName              string     `json:"levelName"`
	RequiredInviteCount    int        `json:"requiredInviteCount"`
	RequiredRechargeAmount float64    `json:"requiredRechargeAmount"`
	DurationMinutes        int        `json:"durationMinutes"`
	RewardAmount           float64    `json:"rewardAmount"`
	RewardCurrency         string     `json:"rewardCurrency"`
	RewardTarget           string     `json:"rewardTarget"`
	Claimed                bool       `json:"claimed"`
	Status                 *int8      `json:"status"`
	ClaimedAt              *time.Time `json:"claimedAt"`
	DeadlineAt             *time.Time `json:"deadlineAt"`
	RemainingSeconds       int64      `json:"remainingSeconds"`
	ProgressInviteCount    int        `json:"progressInviteCount"`
	ProgressRechargeCount  int        `json:"progressRechargeCount"`
	ProgressRechargeAmount float64    `json:"progressRechargeAmount"`
	ProgressPercent        float64    `json:"progressPercent"`
}

type TgTaskActivityAppRecordResp struct {
	BasePageResponse[TgTaskActivityAppItem]
}

type TgTaskActivityClaimReq struct {
	LevelCode   string `json:"levelCode"`
	InviteCount int    `json:"inviteCount"`
}

var TgTaskActivityRecordTableName = "tg_task_activity_record"

func (TgTaskActivityRecord) TableName() string {
	return TgTaskActivityRecordTableName
}
