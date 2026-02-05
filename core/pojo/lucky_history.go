package pojo

import (
	"time"
)

type LuckyHistory struct {
	BaseModel
	UserID     int64   `json:"userId" gorm:"type:bigint;index"`         // 用户ID
	FirstName  string  `json:"firstName" gorm:"type:varchar(64)"`      // 用户名称
	LuckyID    int64   `json:"luckyId" gorm:"type:bigint;index"`       // 红包ID
	IsThunder  int     `json:"isThunder" gorm:"type:int;default:0"`    // 是否中雷 0未中雷 1中雷
	Amount     float64 `json:"amount" gorm:"type:numeric(20,3)"`        // 领取金额
	LoseMoney  float64 `json:"loseMoney" gorm:"type:numeric(20,3);default:0"` // 损失金额(中雷时)
}

type LuckyHistoryBack struct {
	ID         int64     `json:"id"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
	UserID     int64     `json:"userId"`
	FirstName  string    `json:"firstName"`
	LuckyID    int64     `json:"luckyId"`
	IsThunder  int       `json:"isThunder"`
	Amount     float64   `json:"amount"`
	LoseMoney  float64   `json:"loseMoney"`
}

type LuckyHistorySearch struct {
	PageInfo
	LuckyID int64 `json:"luckyId"` // 红包ID
	UserID  int64 `json:"userId"`  // 用户ID
}

type LuckyHistoryResp struct {
	BasePageResponse[LuckyHistoryBack]
}

var LuckyHistoryTableName = "lucky_history"

func (LuckyHistory) TableName() string {
	return LuckyHistoryTableName
}
