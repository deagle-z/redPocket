package pojo

import (
	"time"
)

type LuckyHistory struct {
	BaseModel
	UserID    int64   `json:"userId" gorm:"type:bigint;index"`               // 用户ID
	FirstName string  `json:"firstName" gorm:"type:varchar(64)"`             // 用户名称
	LuckyID   int64   `json:"luckyId" gorm:"type:bigint;index"`              // 红包ID
	IsThunder int     `json:"isThunder" gorm:"type:int;default:0"`           // 是否中雷 0未中雷 1中雷
	Amount    float64 `json:"amount" gorm:"type:numeric(20,3)"`              // 领取金额
	LoseMoney float64 `json:"loseMoney" gorm:"type:numeric(20,3);default:0"` // 损失金额(中雷时)
	TenantId  int64   `json:"tenantId" gorm:"type:bigint;"`
}

type LuckyHistoryBack struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	UserID    int64     `json:"userId"`
	FirstName string    `json:"firstName"`
	LuckyID   int64     `json:"luckyId"`
	IsThunder int       `json:"isThunder"`
	Amount    float64   `json:"amount"`
	LoseMoney float64   `json:"loseMoney"`
	TenantId  int64     `json:"tenantId"`
}

type LuckyHistorySearch struct {
	PageInfo
	LuckyID  int64 `json:"luckyId"` // 红包ID
	UserID   int64 `json:"userId"`  // 用户ID
	TenantId int64 `json:"tenantId"`
}

type LuckyHistoryResp struct {
	BasePageResponse[LuckyHistoryBack]
}

type LuckyRecentWinnerSearch struct {
	Limit int `json:"limit"` // 返回条数，默认10，最大50
}

type LuckyRecentWinnerBack struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"userId"`
	FirstName string    `json:"firstName"`
	Avatar    *string   `json:"avatar"`
	Amount    float64   `json:"amount"`
	LuckyID   int64     `json:"luckyId"`
	CreatedAt time.Time `json:"createdAt"`
	TimeText  string    `json:"timeText"` // 刚刚/xx分钟前/xx小时前
}

type LuckyAppHistorySearch struct {
	PageInfo
	ActionType int   `json:"actionType"` // 0全部 1我发的 2我抢的
	ResultType int   `json:"resultType"` // 0全部 1盈利/中奖 2亏损/中雷
	StartTime  int64 `json:"startTime"`  // 开始时间（unix秒）
	EndTime    int64 `json:"endTime"`    // 结束时间（unix秒）
}

type LuckyAppHistoryBack struct {
	RecordType  string    `json:"recordType" gorm:"column:record_type"` // send/grab
	ActionType  int       `json:"actionType" gorm:"column:action_type"` // 1发包 2抢包
	RecordID    int64     `json:"recordId" gorm:"column:record_id"`
	LuckyID     int64     `json:"luckyId" gorm:"column:lucky_id"`
	LuckyAmount float64   `json:"luckyAmount" gorm:"column:lucky_amount"` // 发包总额
	GrabAmount  float64   `json:"grabAmount" gorm:"column:grab_amount"`   // 抢到金额
	LoseMoney   float64   `json:"loseMoney" gorm:"column:lose_money"`
	IsThunder   int       `json:"isThunder" gorm:"column:is_thunder"`
	Thunder     int       `json:"thunder" gorm:"column:thunder"`
	SenderID    int64     `json:"senderId" gorm:"column:sender_id"`
	SenderName  string    `json:"senderName" gorm:"column:sender_name"`
	Avatar      *string   `json:"avatar" gorm:"column:avatar"`
	Income      float64   `json:"income" gorm:"column:income"`
	Expense     float64   `json:"expense" gorm:"column:expense"`
	NetProfit   float64   `json:"netProfit" gorm:"column:net_amount"`
	CreatedAt   time.Time `json:"createdAt" gorm:"column:created_at"`
}

type LuckyAppHistoryResp struct {
	BasePageResponse[LuckyAppHistoryBack]
	TotalIncome   float64 `json:"totalIncome"`
	TotalExpense  float64 `json:"totalExpense"`
	NetProfitLoss float64 `json:"netProfitLoss"`
}

var LuckyHistoryTableName = "lucky_history"

func (LuckyHistory) TableName() string {
	return LuckyHistoryTableName
}
