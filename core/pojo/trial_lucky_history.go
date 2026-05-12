package pojo

import "time"

type TrialLuckyHistory struct {
	BaseModel
	UserID       int64   `json:"userId" gorm:"type:bigint;index"`
	ActorType    string  `json:"actorType" gorm:"type:varchar(16);not null;default:'user';index"`
	FirstName    string  `json:"firstName" gorm:"type:varchar(64)"`
	LuckyID      int64   `json:"luckyId" gorm:"type:bigint;index"`
	IsThunder    int     `json:"isThunder" gorm:"type:tinyint;default:0"`
	GrabType     int     `json:"grabType" gorm:"type:tinyint;default:1"`
	Amount       float64 `json:"amount" gorm:"type:decimal(20,2);default:0.00"`
	ActualAmount float64 `json:"actualAmount" gorm:"type:decimal(20,2);default:0.00"`
	LoseMoney    float64 `json:"loseMoney" gorm:"type:decimal(20,2);default:0.00"`
	Guess        *int    `json:"guess" gorm:"type:tinyint"`
	TenantId     int64   `json:"tenantId" gorm:"type:bigint;index"`
}

type TrialLuckyHistoryBack struct {
	ID           int64     `json:"id"`
	CreatedAt    time.Time `json:"createdAt"`
	UserID       int64     `json:"userId"`
	ActorType    string    `json:"actorType"`
	FirstName    string    `json:"firstName"`
	LuckyID      int64     `json:"luckyId"`
	IsThunder    int       `json:"isThunder"`
	GrabType     int       `json:"grabType"`
	Amount       float64   `json:"amount"`
	ActualAmount float64   `json:"actualAmount"`
	LoseMoney    float64   `json:"loseMoney"`
	Guess        *int      `json:"guess"`
	TenantId     int64     `json:"tenantId"`
}

type TrialLuckyHistorySearch struct {
	PageInfo
	LuckyID  int64 `json:"luckyId"`
	UserID   int64 `json:"userId"`
	TenantId int64 `json:"tenantId"`
}

type TrialLuckyHistoryResp struct {
	BasePageResponse[TrialLuckyHistoryBack]
}

var TrialLuckyHistoryTableName = "trial_lucky_history"

func (TrialLuckyHistory) TableName() string {
	return TrialLuckyHistoryTableName
}
