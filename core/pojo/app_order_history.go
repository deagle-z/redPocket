package pojo

import "time"

type AppOrderHistorySearch struct {
	PageInfo
}

type AppOrderHistoryItem struct {
	OrderNo        string    `json:"orderNo"`
	Amount         float64   `json:"amount"`
	NetAmount      float64   `json:"netAmount"`
	BonusAmount    *float64  `json:"bonusAmount,omitempty"`
	Fee            *float64  `json:"fee,omitempty"`
	RejectReason   *string   `json:"rejectReason,omitempty"`
	Currency       string    `json:"currency"`
	CurrencySymbol string    `json:"currencySymbol"`
	Time           time.Time `json:"time"`
	Status         int       `json:"status"`
}

type AppOrderHistoryResp struct {
	BasePageResponse[AppOrderHistoryItem]
}
