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
	Type        int8    `yaml:"type" json:"type" gorm:"type:tinyint;default:1"`                   // 变动类型，见 CashHistoryType 常量
	IsGift      int8    `yaml:"isGift" json:"isGift" gorm:"type:tinyint(1);default:0"`            // 0否 1是
	FromUserId  int64   `yaml:"fromUserId" json:"fromUserId" gorm:"type:bigint;index:ccf"`        // 来源用户id
}

const (
	CashHistoryTypeUnknown                int8 = 0
	CashHistoryTypeSendRedPacket          int8 = 1  // 发送红包
	CashHistoryTypeGrabRedPacketWin       int8 = 2  // 抢红包未中雷收益
	CashHistoryTypeGrabRedPacketThunder   int8 = 3  // 抢红包中雷损失
	CashHistoryTypeRedPacketThunderIncome int8 = 4  // 发包者中雷收益
	CashHistoryTypeRedPacketCommission    int8 = 5  // 红包相关抽成
	CashHistoryTypeRechargeCredit         int8 = 6  // 充值到账
	CashHistoryTypeAdminManualAward       int8 = 7  // 后台手工加款
	CashHistoryTypeAdminManualDeduct      int8 = 8  // 后台手工扣款
	CashHistoryTypeWithdrawApply          int8 = 9  // 提现申请扣款
	CashHistoryTypeWithdrawRefund         int8 = 10 // 提现失败/取消/退回返还
	CashHistoryTypeRebateTransfer         int8 = 11 // 佣金(返水)转余额
	CashHistoryTypeLuckyExpireRefund      int8 = 12 // 红包过期退回
)

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
