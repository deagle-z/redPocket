package pojo

import "time"

type PlatformProfitLedger struct {
	BaseModel
	TenantId      int64   `json:"tenantId" gorm:"column:tenant_id;type:bigint;not null;default:0;comment:租户ID"`
	UserId        int64   `json:"userId" gorm:"column:user_id;type:bigint;not null;default:0;comment:关联tg_user.id"`
	SourceType    string  `json:"sourceType" gorm:"column:source_type;type:varchar(64);not null;comment:来源类型"`
	SourceId      string  `json:"sourceId" gorm:"column:source_id;type:varchar(128);not null;uniqueIndex:uk_source;comment:来源唯一号"`
	IncomeAmount  float64 `json:"incomeAmount" gorm:"column:income_amount;type:decimal(20,3);not null;default:0.000;comment:盈利金额"`
	ExpenseAmount float64 `json:"expenseAmount" gorm:"column:expense_amount;type:decimal(20,3);not null;default:0.000;comment:亏损金额"`
	NetAmount     float64 `json:"netAmount" gorm:"column:net_amount;type:decimal(20,3);->;comment:净额"`
	Remark        string  `json:"remark" gorm:"column:remark;type:varchar(255);not null;default:'';comment:备注"`
}

const (
	PlatformProfitSourceRegisterGift           = "register_gift"
	PlatformProfitSourceRechargeGift           = "recharge_gift"
	PlatformProfitSourceLuckyGrabCommission    = "lucky_grab_commission"
	PlatformProfitSourceLuckyThunderCommission = "lucky_thunder_commission"
)

type PlatformProfitLedgerSearch struct {
	PageInfo
	TenantId   *int64   `json:"tenantId"`
	UserId     int64    `json:"userId"`
	SourceType string   `json:"sourceType"`
	SourceId   string   `json:"sourceId"`
	MinNet     *float64 `json:"minNet"`
	MaxNet     *float64 `json:"maxNet"`
}

type PlatformProfitLedgerSet struct {
	ID            int64   `json:"id"`
	TenantId      int64   `json:"tenantId"`
	UserId        int64   `json:"userId"`
	SourceType    string  `json:"sourceType"`
	SourceId      string  `json:"sourceId"`
	IncomeAmount  float64 `json:"incomeAmount"`
	ExpenseAmount float64 `json:"expenseAmount"`
	Remark        string  `json:"remark"`
}

type PlatformProfitLedgerBack struct {
	ID            int64     `json:"id"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
	TenantId      int64     `json:"tenantId"`
	UserId        int64     `json:"userId"`
	SourceType    string    `json:"sourceType"`
	SourceId      string    `json:"sourceId"`
	IncomeAmount  float64   `json:"incomeAmount"`
	ExpenseAmount float64   `json:"expenseAmount"`
	NetAmount     float64   `json:"netAmount"`
	Remark        string    `json:"remark"`
}

type PlatformProfitLedgerResp struct {
	BasePageResponse[PlatformProfitLedgerBack]
}

var PlatformProfitLedgerTableName = "platform_profit_ledger"

func (PlatformProfitLedger) TableName() string {
	return PlatformProfitLedgerTableName
}
