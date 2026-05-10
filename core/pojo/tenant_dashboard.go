package pojo

import "time"

type TenantDashboardPeriodStats struct {
	RechargeAmount     float64 `json:"rechargeAmount"`
	BetAmount          float64 `json:"betAmount"`
	WithdrawAmount     float64 `json:"withdrawAmount"`
	RebateAmount       float64 `json:"rebateAmount"`
	PlatformPumpAmount float64 `json:"platformPumpAmount"`
	RechargeUsers      int64   `json:"rechargeUsers"`
	RegisterUsers      int64   `json:"registerUsers"`
}

type TenantDashboardStatsBack struct {
	Today                   TenantDashboardPeriodStats `json:"today"`
	Yesterday               TenantDashboardPeriodStats `json:"yesterday"`
	Month                   TenantDashboardPeriodStats `json:"month"`
	TotalPlatformPumpAmount float64                    `json:"totalPlatformPumpAmount"`
	TotalRegisterUsers      int64                      `json:"totalRegisterUsers"`
	OnlineUsers             int64                      `json:"onlineUsers"`
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
