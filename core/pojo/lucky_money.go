package pojo

import (
	"time"
)

type LuckyMoney struct {
	BaseModel
	SenderID   int64   `json:"senderId" gorm:"type:bigint;index"`                    // 发送者ID
	SenderName string  `json:"senderName" gorm:"type:varchar(64)"`                  // 发送者名称
	Amount     float64 `json:"amount" gorm:"type:numeric(20,3)"`                     // 红包总金额
	Received   float64 `json:"received" gorm:"type:numeric(20,3);default:0"`        // 已领取金额
	Number     int     `json:"number" gorm:"type:int"`                               // 红包总个数
	Lucky      int     `json:"lucky" gorm:"type:int;default:1"`                     // 是否有效 1有效
	Thunder    int     `json:"thunder" gorm:"type:int"`                              // 雷号(1-9)
	ChatID     int64   `json:"chatId" gorm:"type:bigint;index"`                     // Telegram群组ID
	RedList    string  `json:"redList" gorm:"type:text"`                             // 红包金额列表(JSON)
	LoseRate   float64 `json:"loseRate" gorm:"type:numeric(10,2);default:1.8"`     // 中雷倍数
	Status     int     `json:"status" gorm:"type:int;default:1"`                    // 状态 1进行中 2已完成
}

type LuckyMoneyBack struct {
	ID         int64     `json:"id"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
	SenderID   int64     `json:"senderId"`
	SenderName string    `json:"senderName"`
	Amount     float64   `json:"amount"`
	Received   float64   `json:"received"`
	Number     int       `json:"number"`
	Lucky      int       `json:"lucky"`
	Thunder    int       `json:"thunder"`
	ChatID     int64     `json:"chatId"`
	RedList    string    `json:"redList"`
	LoseRate   float64   `json:"loseRate"`
	Status     int       `json:"status"`
}

type LuckyMoneySearch struct {
	PageInfo
	SenderID int64 `json:"senderId"` // 发送者ID
	ChatID   int64 `json:"chatId"`   // 群组ID
	Status   *int  `json:"status"`   // 状态
}

type LuckyMoneySend struct {
	Amount  float64 `json:"amount" binding:"required,gt=0"`  // 红包金额，必填，大于0
	Thunder int     `json:"thunder" binding:"required,gte=1,lte=9"` // 雷号，必填，1-9
	Number  *int    `json:"number"`                          // 红包数量，可选
	ChatID  int64   `json:"chatId"`                          // 群组ID
}

type LuckyMoneyGrab struct {
	LuckyID int64 `json:"luckyId" binding:"required,gt=0"` // 红包ID，必填
}

type LuckyMoneyResp struct {
	BasePageResponse[LuckyMoneyBack]
}

var LuckyMoneyTableName = "lucky_money"

func (LuckyMoney) TableName() string {
	return LuckyMoneyTableName
}
