package pojo

import "time"

const (
	WithdrawFlowBatchStatusActive    = 1
	WithdrawFlowBatchStatusCompleted = 2
	WithdrawFlowBatchStatusClosed    = 3

	WithdrawFlowBatchSourceRechargeV2      = "recharge_v2"
	WithdrawFlowBatchSourceVipWeeklySalary = "vip_weekly_salary"

	WithdrawFlowEventTypeGameBet            = "game_bet"
	WithdrawFlowEventTypeLuckySend          = "lucky_send"
	WithdrawFlowEventTypeLuckyGrab          = "lucky_grab"
	WithdrawFlowEventTypeLuckySenderGrabbed = "lucky_sender_grabbed"

	WithdrawFlowBatchClosedReasonBalanceBelowThreshold = "balance_below_threshold"
)

type TgUserWithdrawFlowBatch struct {
	BaseModel
	TenantID           int64      `json:"tenantId" gorm:"column:tenant_id;type:bigint;index:idx_withdraw_flow_batch_tenant_user_status,priority:1"`
	UserID             int64      `json:"userId" gorm:"column:user_id;type:bigint;index:idx_withdraw_flow_batch_user_status_id,priority:1;index:idx_withdraw_flow_batch_tenant_user_status,priority:2"`
	SourceType         string     `json:"sourceType" gorm:"column:source_type;type:varchar(32);not null;default:'';uniqueIndex:idx_withdraw_flow_batch_source,priority:1"`
	SourceOrderID      int64      `json:"sourceOrderId" gorm:"column:source_order_id;type:bigint;not null;default:0"`
	SourceOrderNo      string     `json:"sourceOrderNo" gorm:"column:source_order_no;type:varchar(64);not null;default:'';uniqueIndex:idx_withdraw_flow_batch_source,priority:2"`
	ActivityType       int8       `json:"activityType" gorm:"column:activity_type;type:tinyint;not null;default:0"`
	ActivityCode       string     `json:"activityCode" gorm:"column:activity_code;type:varchar(64);not null;default:''"`
	CreditAmount       float64    `json:"creditAmount" gorm:"column:credit_amount;type:decimal(20,2);not null;default:0"`
	BonusAmount        float64    `json:"bonusAmount" gorm:"column:bonus_amount;type:decimal(20,2);not null;default:0"`
	BaseAmount         float64    `json:"baseAmount" gorm:"column:base_amount;type:decimal(20,2);not null;default:0"`
	WithdrawMultiplier float64    `json:"withdrawMultiplier" gorm:"column:withdraw_multiplier;type:decimal(10,2);not null;default:2"`
	GiftMultiplier     float64    `json:"giftMultiplier" gorm:"column:gift_multiplier;type:decimal(10,2);not null;default:5"`
	RequiredFlow       float64    `json:"requiredFlow" gorm:"column:required_flow;type:decimal(20,2);not null;default:0"`
	CompletedFlow      float64    `json:"completedFlow" gorm:"column:completed_flow;type:decimal(20,2);not null;default:0"`
	Status             int        `json:"status" gorm:"column:status;type:tinyint;not null;default:1;index:idx_withdraw_flow_batch_user_status_id,priority:2;index:idx_withdraw_flow_batch_tenant_user_status,priority:3"`
	ClosedReason       string     `json:"closedReason" gorm:"column:closed_reason;type:varchar(64);not null;default:''"`
	CompletedAt        *time.Time `json:"completedAt" gorm:"column:completed_at;type:datetime(3)"`
	ClosedAt           *time.Time `json:"closedAt" gorm:"column:closed_at;type:datetime(3)"`
	LastFlowAt         *time.Time `json:"lastFlowAt" gorm:"column:last_flow_at;type:datetime(3)"`
}

var TgUserWithdrawFlowBatchTableName = "tg_user_withdraw_flow_batch"

func (TgUserWithdrawFlowBatch) TableName() string {
	return TgUserWithdrawFlowBatchTableName
}

type TgUserWithdrawFlowEvent struct {
	BaseModel
	TenantID        int64     `json:"tenantId" gorm:"column:tenant_id;type:bigint;index:idx_withdraw_flow_event_tenant_user,priority:1"`
	UserID          int64     `json:"userId" gorm:"column:user_id;type:bigint;index:idx_withdraw_flow_event_user_time,priority:1;index:idx_withdraw_flow_event_tenant_user,priority:2"`
	EventType       string    `json:"eventType" gorm:"column:event_type;type:varchar(32);not null;default:''"`
	EventKey        string    `json:"eventKey" gorm:"column:event_key;type:varchar(128);not null;default:'';uniqueIndex"`
	SourceID        int64     `json:"sourceId" gorm:"column:source_id;type:bigint;not null;default:0"`
	SourceOrderNo   string    `json:"sourceOrderNo" gorm:"column:source_order_no;type:varchar(64);not null;default:''"`
	FlowAmount      float64   `json:"flowAmount" gorm:"column:flow_amount;type:decimal(20,2);not null;default:0"`
	AllocatedAmount float64   `json:"allocatedAmount" gorm:"column:allocated_amount;type:decimal(20,2);not null;default:0"`
	OccurredAt      time.Time `json:"occurredAt" gorm:"column:occurred_at;type:datetime(3);index:idx_withdraw_flow_event_user_time,priority:2"`
}

var TgUserWithdrawFlowEventTableName = "tg_user_withdraw_flow_event"

func (TgUserWithdrawFlowEvent) TableName() string {
	return TgUserWithdrawFlowEventTableName
}

type TgUserWithdrawFlowAllocation struct {
	BaseModel
	TenantID        int64   `json:"tenantId" gorm:"column:tenant_id;type:bigint"`
	UserID          int64   `json:"userId" gorm:"column:user_id;type:bigint;index:idx_withdraw_flow_alloc_user_batch,priority:1"`
	EventID         int64   `json:"eventId" gorm:"column:event_id;type:bigint;uniqueIndex:idx_withdraw_flow_alloc_event_batch,priority:1"`
	BatchID         int64   `json:"batchId" gorm:"column:batch_id;type:bigint;index;uniqueIndex:idx_withdraw_flow_alloc_event_batch,priority:2;index:idx_withdraw_flow_alloc_user_batch,priority:2"`
	AllocatedAmount float64 `json:"allocatedAmount" gorm:"column:allocated_amount;type:decimal(20,2);not null;default:0"`
}

var TgUserWithdrawFlowAllocationTableName = "tg_user_withdraw_flow_allocation"

func (TgUserWithdrawFlowAllocation) TableName() string {
	return TgUserWithdrawFlowAllocationTableName
}

type TgWithdrawFlowBatchSummaryBack struct {
	Balance              float64 `json:"balance"`
	HasUnfinishedBatch   bool    `json:"hasUnfinishedBatch"`
	RequiredFlow         float64 `json:"requiredFlow"`
	CompletedFlow        float64 `json:"completedFlow"`
	RemainingFlow        float64 `json:"remainingFlow"`
	UnfinishedBatchCount int64   `json:"unfinishedBatchCount"`
}

type TgWithdrawFlowBatchOverviewBack struct {
	UserID                  int64                     `json:"userId"`
	Balance                 float64                   `json:"balance"`
	TotalFlow               float64                   `json:"totalFlow"`
	TotalRequiredFlow       float64                   `json:"totalRequiredFlow"`
	TotalCompletedFlow      float64                   `json:"totalCompletedFlow"`
	HasUnfinishedBatch      bool                      `json:"hasUnfinishedBatch"`
	UnfinishedRequiredFlow  float64                   `json:"unfinishedRequiredFlow"`
	UnfinishedCompletedFlow float64                   `json:"unfinishedCompletedFlow"`
	UnfinishedRemainingFlow float64                   `json:"unfinishedRemainingFlow"`
	BatchCount              int64                     `json:"batchCount"`
	UnfinishedBatchCount    int64                     `json:"unfinishedBatchCount"`
	Batches                 []TgWithdrawFlowBatchBack `json:"batches"`
}

type TgWithdrawFlowBatchBack struct {
	ID                 int64      `json:"id"`
	CreatedAt          time.Time  `json:"createdAt"`
	UpdatedAt          time.Time  `json:"updatedAt"`
	TenantID           int64      `json:"tenantId"`
	UserID             int64      `json:"userId"`
	SourceType         string     `json:"sourceType"`
	SourceOrderID      int64      `json:"sourceOrderId"`
	SourceOrderNo      string     `json:"sourceOrderNo"`
	ActivityType       int8       `json:"activityType"`
	ActivityCode       string     `json:"activityCode"`
	CreditAmount       float64    `json:"creditAmount"`
	BonusAmount        float64    `json:"bonusAmount"`
	BaseAmount         float64    `json:"baseAmount"`
	WithdrawMultiplier float64    `json:"withdrawMultiplier"`
	GiftMultiplier     float64    `json:"giftMultiplier"`
	RequiredFlow       float64    `json:"requiredFlow"`
	CompletedFlow      float64    `json:"completedFlow"`
	RemainingFlow      float64    `json:"remainingFlow"`
	ProgressPercent    float64    `json:"progressPercent"`
	Status             int        `json:"status"`
	ClosedReason       string     `json:"closedReason"`
	CompletedAt        *time.Time `json:"completedAt"`
	ClosedAt           *time.Time `json:"closedAt"`
	LastFlowAt         *time.Time `json:"lastFlowAt"`
}
