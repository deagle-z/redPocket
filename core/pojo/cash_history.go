package pojo

import "time"

type CashHistory struct { // 积分变动记录
	BaseModel
	UserId      int64   `yaml:"userId" json:"userId" gorm:"type:bigint;uniqueIndex:ut"`           // 用户id
	AwardUni    string  `yaml:"awardUni" json:"awardUni" gorm:"type:varchar(128);uniqueIndex:ut"` // 奖励唯一key
	Amount      float64 `yaml:"amount" json:"amount" gorm:"type:numeric(20,3)"`                   // 变动积分
	StartAmount float64 `yaml:"startAmount" json:"startAmount" gorm:"type:numeric(20,3)"`         // 变动前积分
	EndAmount   float64 `yaml:"endAmount" json:"endAmount" gorm:"type:numeric(20,3)"`             // 变动后积分
	CashMark    string  `yaml:"cashMark" json:"cashMark" gorm:"type:varchar(32);index:ccf"`       // 积分备注
	CashDesc    string  `yaml:"cashDesc" json:"cashDesc" gorm:"type:varchar(64);index:ccf"`       // 积分描述
	Type        int8    `yaml:"type" json:"type" gorm:"type:tinyint;default:1"`                   // 1发红包金额 2抢红包收益金额 3抢红包中雷金额 4抽成金额
	IsGift      int8    `yaml:"isGift" json:"isGift" gorm:"type:tinyint(1);default:0"`            // 0否 1是
	FromUserId  int64   `yaml:"fromUserId" json:"fromUserId" gorm:"type:bigint;index:ccf"`        // 来源用户id
}

type CashHistoryResp struct {
	CreatedAt   time.Time `json:"createdAt"`
	UserId      int64     `json:"userId"`
	Amount      float64   `json:"amount"`
	StartAmount float64   `json:"startAmount"`
	EndAmount   float64   `json:"endAmount"`
	CashMark    string    `json:"cashMark"`
	CashDesc    string    `json:"cashDesc"`
	Type        int8      `json:"type"`
	IsGift      int8      `json:"isGift"`
	FromUserId  int64     `json:"fromUserId"`
}

type CashHistorySearch struct {
	PageInfo
	UserId   int64  `json:"userId"`   // 用户ID，0表示查询所有用户（管理端）
	CashMark string `json:"cashMark"` // 余额备注，可选
}

type CashHistoryPage struct {
	BasePageResponse[CashHistoryResp]
}

var CashHistoryTableName = "cash_history"
var CashHistoryShardingName = "user_id"
var CashHistoryShards = 8
var AllCashHistoryShardingName = "all_cash_history"

func (CashHistory) TableName() string {
	return CashHistoryTableName
}
