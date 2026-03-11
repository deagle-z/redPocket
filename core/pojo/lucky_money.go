package pojo

import (
	"time"
)

type LuckyMoney struct {
	BaseModel
	SenderID   int64     `json:"senderId" gorm:"type:bigint;index"`              // 发送者ID
	SenderName string    `json:"senderName" gorm:"type:varchar(64)"`             // 发送者名称
	Amount     float64   `json:"amount" gorm:"type:numeric(20,3)"`               // 红包总金额
	Received   float64   `json:"received" gorm:"type:numeric(20,3);default:0"`   // 已领取金额
	Number     int       `json:"number" gorm:"type:int"`                         // 红包总个数
	Lucky      int       `json:"lucky" gorm:"type:int;default:1"`                // 是否有效 1有效
	Thunder    int       `json:"thunder" gorm:"type:int"`                        // 雷号(1-9)
	ChatID     int64     `json:"chatId" gorm:"type:bigint;index"`                // Telegram群组ID
	RedList    string    `json:"redList" gorm:"type:text"`                       // 红包金额列表(JSON)
	LoseRate   float64   `json:"loseRate" gorm:"type:numeric(10,2);default:1.8"` // 中雷倍数
	Status     int       `json:"status" gorm:"type:int;default:1"`               // 状态 1进行中 2已完成
	TenantId   int64     `json:"tenantId" gorm:"type:bigint;"`
	ExpireTime time.Time `json:"expireTime" gorm:"column:expire_time;"`
}

type LuckyMoneyBack struct {
	ID           int64     `json:"id"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
	SenderID     int64     `json:"senderId"`
	SenderName   string    `json:"senderName"`
	SenderAvatar *string   `json:"senderAvatar"`
	Amount       float64   `json:"amount"`
	Received     float64   `json:"received"`
	Number       int       `json:"number"`
	Lucky        int       `json:"lucky"`
	Thunder      int       `json:"thunder"`
	ChatID       int64     `json:"chatId"`
	RedList      string    `json:"redList"`
	LoseRate     float64   `json:"loseRate"`
	Status       int       `json:"status"`
	TenantId     int64     `json:"tenantId"`
	ExpireTime   time.Time `json:"expireTime"`
}

type LuckyMoneySearch struct {
	PageInfo
	SenderID int64 `json:"senderId"` // 发送者ID
	ChatID   int64 `json:"chatId"`   // 群组ID
	Status   *int  `json:"status"`   // 状态
	TenantId int64 `json:"tenantId"`
}

type LuckyMoneySend struct {
	Amount   float64 `json:"amount" binding:"required,gt=0"` // 红包金额，必填，大于0
	Thunder  int     `json:"thunder"`                        // 雷号，必填，0-9
	Number   *int    `json:"number"`                         // 红包数量，可选
	ChatID   int64   `json:"chatId"`                         // 群组ID
	TenantId int64   `json:"tenantId"`
}

type LuckyMoneyGrab struct {
	LuckyID   int64 `json:"luckyId" binding:"required,gt=0"` // 红包ID，必填
	GrabIndex *int  `json:"grabIndex"`                       // 抢第几个包（1-based，可选）
}

type LuckyMoneyResp struct {
	BasePageResponse[LuckyMoneyBack]
}

type LuckyMoneyAppListSearch struct {
	PageInfo
	ChatID  int64 `json:"chatId"`  // 群组ID，可选
	Status  *int  `json:"status"`  // 状态，可选（默认进行中）
	LuckyID int64 `json:"luckyId"` // 红包ID，可选
}

type LuckyMoneyAppItemBack struct {
	SeqNo         uint    `json:"seqNo"`
	Amount        float64 `json:"amount"`
	ThunderAmount float64 `json:"thunderAmount"`
	IsGrabbed     int8    `json:"isGrabbed"`
	Thunder       int8    `json:"thunder"`
	IsGrabMine    int8    `json:"isGrabMine"` // 是否本人抢到
}

type LuckyMoneyAppBack struct {
	ID               int64                   `json:"id"`
	SenderID         int64                   `json:"senderId"`
	SenderName       string                  `json:"senderName"`
	SenderAvatar     *string                 `json:"senderAvatar"`
	Amount           float64                 `json:"amount"`
	Received         float64                 `json:"received"`
	Number           int                     `json:"number"`
	GrabbedCount     int64                   `json:"grabbedCount"`
	Thunder          int                     `json:"thunder"`
	HitCount         int64                   `json:"hitCount"` // 中雷次数
	LoseRate         float64                 `json:"loseRate"`
	Status           int                     `json:"status"`
	ExpireTime       time.Time               `json:"expireTime"`
	RemainingSeconds int64                   `json:"remainingSeconds"`
	RemainingText    string                  `json:"remainingText"`
	Items            []LuckyMoneyAppItemBack `json:"items"`
	CreatedAt        time.Time               `json:"createdAt"`
}

type LuckyMoneyAppResp struct {
	BasePageResponse[LuckyMoneyAppBack]
}

type LuckyMoneyAppDetailReq struct {
	LuckyID int64 `json:"luckyId" binding:"required,gt=0"`
}

type LuckyMoneyAppDetailSummary struct {
	ID           int64     `json:"id"`
	Status       int       `json:"status"`
	StatusText   string    `json:"statusText"`
	Amount       float64   `json:"amount"`
	Thunder      int       `json:"thunder"`
	LoseRate     float64   `json:"loseRate"`
	ExpireTime   time.Time `json:"expireTime"`
	GrabbedCount int64     `json:"grabbedCount"`
	Number       int       `json:"number"`
	GameText     string    `json:"gameText"`
	RoomText     string    `json:"roomText"`
	UnitAmount   string    `json:"unitAmount"`
}

type LuckyMoneyAppDetailSender struct {
	SenderID     int64     `json:"senderId"`
	SenderName   string    `json:"senderName"`
	SenderAvatar *string   `json:"senderAvatar"`
	SendTime     time.Time `json:"sendTime"`
}

type LuckyMoneyAppDetailFinance struct {
	SendAmount     float64 `json:"sendAmount"`
	ReceivedAmount float64 `json:"receivedAmount"`
	RefundAmount   float64 `json:"refundAmount"`
	ThunderIncome  float64 `json:"thunderIncome"`
	HitCount       int64   `json:"hitCount"`
	FinalProfit    float64 `json:"finalProfit"`
}

type LuckyMoneyAppDetailParticipant struct {
	SeqNo         uint      `json:"seqNo"`
	UserID        int64     `json:"userId"`
	FirstName     string    `json:"firstName"`
	Avatar        *string   `json:"avatar"`
	Amount        float64   `json:"amount"`
	ThunderAmount float64   `json:"thunderAmount"`
	IsThunder     int       `json:"isThunder"`
	CreatedAt     time.Time `json:"createdAt"`
}

type LuckyMoneyAppDetailResp struct {
	Summary          LuckyMoneyAppDetailSummary       `json:"summary"`
	Sender           LuckyMoneyAppDetailSender        `json:"sender"`
	Finance          LuckyMoneyAppDetailFinance       `json:"finance"`
	ParticipantCount int64                            `json:"participantCount"`
	Participants     []LuckyMoneyAppDetailParticipant `json:"participants"`
}

var LuckyMoneyTableName = "lucky_money"

func (LuckyMoney) TableName() string {
	return LuckyMoneyTableName
}
