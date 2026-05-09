package pojo

import "time"

const (
	WithdrawActivityCycleStatusActive = 1
	WithdrawActivityCycleStatusEnded  = 2

	WithdrawActivityCycleEndReasonWithdraw          = "withdraw"
	WithdrawActivityCycleEndReasonBalanceBelowLimit = "balance_below_threshold"
)

type TgUserWithdrawActivityCycle struct {
	BaseModel
	TenantId         int64      `json:"tenantId" gorm:"column:tenant_id;type:bigint;index"`
	UserID           int64      `json:"userId" gorm:"column:user_id;type:bigint;index:idx_withdraw_activity_cycle_user_status,priority:1"`
	Status           int        `json:"status" gorm:"column:status;type:tinyint;not null;default:1;index:idx_withdraw_activity_cycle_user_status,priority:2"`
	ActivityCode     string     `json:"activityCode" gorm:"column:activity_code;type:varchar(64);not null;default:''"`
	ActivityType     int8       `json:"activityType" gorm:"column:activity_type;type:tinyint;not null;default:0"`
	Multiplier       float64    `json:"multiplier" gorm:"column:multiplier;type:decimal(10,2);not null;default:2"`
	BaseAmount       float64    `json:"baseAmount" gorm:"column:base_amount;type:decimal(20,2);not null;default:0"`
	RequiredFlow     float64    `json:"requiredFlow" gorm:"column:required_flow;type:decimal(20,2);not null;default:0"`
	FlowStartValue   float64    `json:"flowStartValue" gorm:"column:flow_start_value;type:decimal(20,2);not null;default:0"`
	FlowConsumed     float64    `json:"flowConsumed" gorm:"column:flow_consumed;type:decimal(20,2);not null;default:0"`
	BalanceThreshold float64    `json:"balanceThreshold" gorm:"column:balance_threshold;type:decimal(20,2);not null;default:10"`
	LastRechargeNo   string     `json:"lastRechargeNo" gorm:"column:last_recharge_no;type:varchar(64);not null;default:''"`
	EndReason        string     `json:"endReason" gorm:"column:end_reason;type:varchar(64);not null;default:''"`
	StartedAt        *time.Time `json:"startedAt" gorm:"column:started_at;type:datetime(3)"`
	EndedAt          *time.Time `json:"endedAt" gorm:"column:ended_at;type:datetime(3)"`
}

var TgUserWithdrawActivityCycleTableName = "tg_user_withdraw_activity_cycle"

func (TgUserWithdrawActivityCycle) TableName() string {
	return TgUserWithdrawActivityCycleTableName
}
