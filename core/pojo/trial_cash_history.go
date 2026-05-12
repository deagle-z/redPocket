package pojo

import "time"

const (
	TrialCashTypeSendLucky     int8 = 1
	TrialCashTypeGrabLuckyWin  int8 = 2
	TrialCashTypeGrabLuckyLose int8 = 3
	TrialCashTypeThunderIncome int8 = 4
	TrialCashTypeExpireRefund  int8 = 5
)

type TrialCashHistory struct {
	BaseModel
	UserId      int64   `json:"userId" gorm:"type:bigint;index"`
	ActorType   string  `json:"actorType" gorm:"type:varchar(16);not null;default:'user';index"`
	AwardUni    string  `json:"awardUni" gorm:"type:varchar(128);uniqueIndex"`
	Amount      float64 `json:"amount" gorm:"type:decimal(20,2);not null;default:0.00"`
	StartAmount float64 `json:"startAmount" gorm:"type:decimal(20,2);not null;default:0.00"`
	EndAmount   float64 `json:"endAmount" gorm:"type:decimal(20,2);not null;default:0.00"`
	CashMark    string  `json:"cashMark" gorm:"type:varchar(32);index"`
	CashDesc    string  `json:"cashDesc" gorm:"type:varchar(255)"`
	Type        int8    `json:"type" gorm:"type:tinyint;default:0;index"`
	IsThunder   int8    `json:"isThunder" gorm:"type:tinyint;default:0"`
	LuckyID     int64   `json:"luckyId" gorm:"type:bigint;index"`
	TenantId    int64   `json:"tenantId" gorm:"type:bigint;index"`
}

type TrialCashHistoryResp struct {
	ID          int64     `json:"id"`
	CreatedAt   time.Time `json:"createdAt"`
	UserId      int64     `json:"userId"`
	ActorType   string    `json:"actorType"`
	Amount      float64   `json:"amount"`
	StartAmount float64   `json:"startAmount"`
	EndAmount   float64   `json:"endAmount"`
	CashMark    string    `json:"cashMark"`
	CashDesc    string    `json:"cashDesc"`
	Type        int8      `json:"type"`
	IsThunder   int8      `json:"isThunder"`
	LuckyID     int64     `json:"luckyId"`
	TenantId    int64     `json:"tenantId"`
}

type TrialCashHistorySearch struct {
	PageInfo
	UserId    int64  `json:"userId"`
	ActorType string `json:"actorType"`
	CashMark  string `json:"cashMark"`
	TenantId  int64  `json:"tenantId"`
}

type TrialCashHistoryPage struct {
	BasePageResponse[TrialCashHistoryResp]
}

var TrialCashHistoryTableName = "trial_cash_history"

func (TrialCashHistory) TableName() string {
	return TrialCashHistoryTableName
}
