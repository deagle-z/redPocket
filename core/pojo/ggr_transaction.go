package pojo

import "time"

const (
	GGRTransactionResultFailed  = 0
	GGRTransactionResultSuccess = 1
)

type GGRTransaction struct {
	BaseModel
	TxnID              string    `json:"txnId" gorm:"column:txn_id;type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin;not null;uniqueIndex:uk_ggr_transaction_event,priority:1"`
	UserID             int64     `json:"userId" gorm:"column:user_id;type:bigint;not null;index:idx_ggr_transaction_user_time,priority:1"`
	UID                string    `json:"uid" gorm:"column:uid;type:varchar(255);not null;index"`
	AgentCode          string    `json:"agentCode" gorm:"column:agent_code;type:varchar(128);not null"`
	UserCode           string    `json:"userCode" gorm:"column:user_code;type:varchar(255);not null"`
	UserToken          string    `json:"userToken" gorm:"column:user_token;type:varchar(255);not null"`
	GameType           string    `json:"gameType" gorm:"column:game_type;type:varchar(16);not null"`
	ProviderCode       string    `json:"providerCode" gorm:"column:provider_code;type:varchar(64);not null;index"`
	GameCode           string    `json:"gameCode" gorm:"column:game_code;type:varchar(255);not null;index"`
	BetType            string    `json:"betType" gorm:"column:bet_type;type:varchar(128);not null"`
	BetMoney           float64   `json:"betMoney" gorm:"column:bet_money;type:decimal(20,2);not null"`
	WinMoney           float64   `json:"winMoney" gorm:"column:win_money;type:decimal(20,2);not null"`
	RoundID            string    `json:"roundId" gorm:"column:round_id;type:varchar(255);not null;default:'';index"`
	TxnType            string    `json:"txnType" gorm:"column:txn_type;type:varchar(32);not null;uniqueIndex:uk_ggr_transaction_event,priority:2"`
	Info               string    `json:"info" gorm:"column:info;type:longtext"`
	AgentBalance       *float64  `json:"agentBalance" gorm:"column:agent_balance;type:decimal(20,2)"`
	RequestUserBalance *float64  `json:"requestUserBalance" gorm:"column:request_user_balance;type:decimal(20,2)"`
	StartBalance       float64   `json:"startBalance" gorm:"column:start_balance;type:decimal(20,2);not null"`
	EndBalance         float64   `json:"endBalance" gorm:"column:end_balance;type:decimal(20,2);not null"`
	RequestFingerprint string    `json:"requestFingerprint" gorm:"column:request_fingerprint;type:char(64);not null"`
	ResultStatus       int       `json:"resultStatus" gorm:"column:result_status;type:tinyint;not null;index"`
	ResultMessage      string    `json:"resultMessage" gorm:"column:result_message;type:varchar(64);not null;default:''"`
	ReceivedAt         time.Time `json:"receivedAt" gorm:"column:received_at;type:datetime(3);not null;index:idx_ggr_transaction_user_time,priority:2"`
}

var GGRTransactionTableName = "ggr_transaction"

func (GGRTransaction) TableName() string {
	return GGRTransactionTableName
}
