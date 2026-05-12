package pojo

import "time"

const TrialActorUser = "user"
const TrialActorBot = "bot"

type TrialLuckyMoney struct {
	BaseModel
	SenderID     int64     `json:"senderId" gorm:"type:bigint;index"`
	SenderType   string    `json:"senderType" gorm:"type:varchar(16);not null;default:'user';index"`
	SenderName   string    `json:"senderName" gorm:"type:varchar(64)"`
	SenderAvatar *string   `json:"senderAvatar" gorm:"type:varchar(512)"`
	Amount       float64   `json:"amount" gorm:"type:decimal(20,2);not null;default:0.00"`
	Received     float64   `json:"received" gorm:"type:decimal(20,2);not null;default:0.00"`
	Number       int       `json:"number" gorm:"type:int;not null;default:0"`
	Thunder      int       `json:"thunder" gorm:"type:int;not null;default:0"`
	GameMode     int       `json:"gameMode" gorm:"type:tinyint;not null;default:0"`
	ChatID       int64     `json:"chatId" gorm:"type:bigint;index"`
	RedList      string    `json:"redList" gorm:"type:text"`
	LoseRate     float64   `json:"loseRate" gorm:"type:decimal(10,2);not null;default:1.80"`
	Status       int       `json:"status" gorm:"type:int;not null;default:1;index"`
	TenantId     int64     `json:"tenantId" gorm:"type:bigint;index"`
	ExpireTime   time.Time `json:"expireTime" gorm:"column:expire_time;index"`
}

type TrialLuckyMoneySearch struct {
	PageInfo
	SenderID int64 `json:"senderId"`
	ChatID   int64 `json:"chatId"`
	Status   *int  `json:"status"`
	TenantId int64 `json:"tenantId"`
}

type TrialLuckyMoneySend struct {
	Amount   float64 `json:"amount" binding:"required,gt=0"`
	Thunder  int     `json:"thunder"`
	GameMode int     `json:"gameMode"`
	Number   *int    `json:"number"`
	ChatID   int64   `json:"chatId"`
	TenantId int64   `json:"tenantId"`
}

type TrialLuckyMoneyGrab struct {
	LuckyID      int64 `json:"luckyId" binding:"required,gt=0"`
	GrabIndex    *int  `json:"grabIndex"`
	OddEvenGuess *int  `json:"oddEvenGuess"`
}

type TrialLuckyMoneyBack struct {
	ID           int64     `json:"id"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
	SenderID     int64     `json:"senderId"`
	SenderType   string    `json:"senderType"`
	SenderName   string    `json:"senderName"`
	SenderAvatar *string   `json:"senderAvatar"`
	Amount       float64   `json:"amount"`
	Received     float64   `json:"received"`
	Number       int       `json:"number"`
	Thunder      int       `json:"thunder"`
	GameMode     int       `json:"gameMode"`
	ChatID       int64     `json:"chatId"`
	RedList      string    `json:"redList"`
	LoseRate     float64   `json:"loseRate"`
	Status       int       `json:"status"`
	TenantId     int64     `json:"tenantId"`
	ExpireTime   time.Time `json:"expireTime"`
}

type TrialLuckyMoneyResp struct {
	BasePageResponse[TrialLuckyMoneyBack]
}

type TrialLuckyMoneyAppListSearch struct {
	PageInfo
	ChatID   int64 `json:"chatId"`
	Status   *int  `json:"status"`
	LuckyID  int64 `json:"luckyId"`
	GameMode *int  `json:"gameMode"`
}

type TrialLuckyMoneyAppItemBack struct {
	SeqNo      uint    `json:"seqNo"`
	Amount     float64 `json:"amount"`
	IsGrabbed  int8    `json:"isGrabbed"`
	Thunder    int8    `json:"thunder"`
	IsGrabMine int8    `json:"isGrabMine"`
}

type TrialLuckyMoneyAppBack struct {
	ID               int64                        `json:"id"`
	SenderID         int64                        `json:"senderId"`
	SenderType       string                       `json:"senderType"`
	SenderName       string                       `json:"senderName"`
	SenderAvatar     *string                      `json:"senderAvatar"`
	Amount           float64                      `json:"amount"`
	Received         float64                      `json:"received"`
	Number           int                          `json:"number"`
	GrabbedCount     int64                        `json:"grabbedCount"`
	Thunder          int                          `json:"thunder"`
	GameMode         int                          `json:"gameMode"`
	HitCount         int64                        `json:"hitCount"`
	LoseRate         float64                      `json:"loseRate"`
	Status           int                          `json:"status"`
	ExpireTime       time.Time                    `json:"expireTime"`
	RemainingSeconds int64                        `json:"remainingSeconds"`
	RemainingText    string                       `json:"remainingText"`
	Items            []TrialLuckyMoneyAppItemBack `json:"items"`
	CreatedAt        time.Time                    `json:"createdAt"`
}

type TrialLuckyMoneyAppResp struct {
	BasePageResponse[TrialLuckyMoneyAppBack]
}

type TrialMeResp struct {
	TrialBalance float64 `json:"trialBalance"`
}

var TrialLuckyMoneyTableName = "trial_lucky_money"

func (TrialLuckyMoney) TableName() string {
	return TrialLuckyMoneyTableName
}
