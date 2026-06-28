package pojo

import "time"

type TenantDashboardPeriodStats struct {
	RechargeAmount      float64 `json:"rechargeAmount"`
	BetAmount           float64 `json:"betAmount"`
	WithdrawAmount      float64 `json:"withdrawAmount"`
	RebateAmount        float64 `json:"rebateAmount"`
	PlatformPumpAmount  float64 `json:"platformPumpAmount"`
	RechargeUsers       int64   `json:"rechargeUsers"`
	RepeatRechargeUsers int64   `json:"repeatRechargeUsers"`
	RegisterUsers       int64   `json:"registerUsers"`
}

type TenantDashboardStatsBack struct {
	Today                   TenantDashboardPeriodStats `json:"today"`
	Yesterday               TenantDashboardPeriodStats `json:"yesterday"`
	Month                   TenantDashboardPeriodStats `json:"month"`
	TotalPlatformPumpAmount float64                    `json:"totalPlatformPumpAmount"`
	TotalRegisterUsers      int64                      `json:"totalRegisterUsers"`
	OnlineUsers             int64                      `json:"onlineUsers"`
}

type TenantDashboardMonthlyBalanceBack struct {
	Month          string  `json:"month"`
	RechargeAmount float64 `json:"rechargeAmount"`
	WithdrawAmount float64 `json:"withdrawAmount"`
	BalanceAmount  float64 `json:"balanceAmount"`
}

type TenantDashboardMonthlyBalanceResp struct {
	Year                int                                 `json:"year"`
	List                []TenantDashboardMonthlyBalanceBack `json:"list"`
	TotalRechargeAmount float64                             `json:"totalRechargeAmount"`
	TotalWithdrawAmount float64                             `json:"totalWithdrawAmount"`
	TotalBalanceAmount  float64                             `json:"totalBalanceAmount"`
}

type TenantDashboardDetailSearch struct {
	PageInfo
	Period   string `json:"period"`
	TenantId int64  `json:"tenantId"`
}

type TenantDashboardUserDetailBack struct {
	ID             int64      `json:"id"`
	TenantId       int64      `json:"tenantId"`
	Uid            string     `json:"uid"`
	TgID           int64      `json:"tgId"`
	Username       *string    `json:"username"`
	FirstName      *string    `json:"firstName"`
	Phone          *string    `json:"phone"`
	ParentID       *int64     `json:"parentId"`
	ParentUid      *string    `json:"parentUid"`
	Balance        float64    `json:"balance"`
	Status         int8       `json:"status"`
	RechargeAmount float64    `json:"rechargeAmount"`
	RechargeCount  int64      `json:"rechargeCount"`
	LastRechargeAt *time.Time `json:"lastRechargeAt"`
	LastActiveAt   *time.Time `json:"lastActiveAt"`
	RegisteredAt   *time.Time `json:"registeredAt"`
}

type TenantDashboardUserDetailResp struct {
	BasePageResponse[TenantDashboardUserDetailBack]
}

// TenantDashboardAgentRankBack 代理排行：按下级用户成功充值金额聚合到上级用户。
type TenantDashboardAgentRankBack struct {
	ID                int64      `json:"id"`
	TenantId          int64      `json:"tenantId"`
	TenantName        *string    `json:"tenantName"`
	Uid               string     `json:"uid"`
	TgID              int64      `json:"tgId"`
	Username          *string    `json:"username"`
	FirstName         *string    `json:"firstName"`
	Phone             *string    `json:"phone"`
	Balance           float64    `json:"balance"`
	Status            int8       `json:"status"`
	SubRechargeAmount float64    `json:"subRechargeAmount"`
	SubRechargeUsers  int64      `json:"subRechargeUsers"`
	SubRechargeCount  int64      `json:"subRechargeCount"`
	LastRechargeAt    *time.Time `json:"lastRechargeAt"`
}

type TenantDashboardAgentRankResp struct {
	BasePageResponse[TenantDashboardAgentRankBack]
}

// TenantDashboardOrderDetailBack 充值/提现订单明细行
type TenantDashboardOrderDetailBack struct {
	ID        int64      `json:"id"`
	OrderNo   string     `json:"orderNo"`
	TenantId  int64      `json:"tenantId"`
	UserID    int64      `json:"userId"`
	Uid       string     `json:"uid"`
	Username  *string    `json:"username"`
	FirstName *string    `json:"firstName"`
	Phone     *string    `json:"phone"`
	Amount    float64    `json:"amount"`
	Fee       float64    `json:"fee"`
	NetAmount float64    `json:"netAmount"`
	Channel   string     `json:"channel"`
	Status    int        `json:"status"`
	Time      *time.Time `json:"time"`
}

// TenantDashboardOrderDetailResp 订单明细分页响应，附带区间汇总金额
type TenantDashboardOrderDetailResp struct {
	BasePageResponse[TenantDashboardOrderDetailBack]
	TotalAmount float64 `json:"totalAmount"`
}
