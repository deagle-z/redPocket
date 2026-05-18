package repository

import (
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/utils"
	"strings"
	"time"

	"gorm.io/gorm"
)

type rechargeOrderHistorySource struct {
	OrderNo     string
	Amount      float64
	NetAmount   float64
	Currency    string
	BonusAmount float64
	Time        time.Time
	Status      int
}

type withdrawOrderHistorySource struct {
	OrderNo      string
	Amount       float64
	NetAmount    float64
	Currency     string
	Fee          float64
	RejectReason *string
	Time         time.Time
	Status       int
}

func GetAppRechargeOrderHistory(db *gorm.DB, userID int64, search pojo.AppOrderHistorySearch) (result pojo.AppOrderHistoryResp) {
	if db == nil || userID <= 0 {
		result.PageSize = search.PageSize
		result.CurrentPage = search.CurrentPage
		return result
	}

	var orders []pojo.RechargeOrder
	query := db.Model(&pojo.RechargeOrder{}).Where("user_id = ?", userID)
	query.Count(&result.Total)
	query.Order("id desc").
		Limit(search.PageSize).
		Offset(search.PageSize * search.CurrentPage).
		Find(&orders)

	result.List = make([]pojo.AppOrderHistoryItem, 0, len(orders))
	for _, order := range orders {
		result.List = append(result.List, buildAppRechargeOrderHistoryItem(rechargeOrderHistorySource{
			OrderNo:     order.OrderNo,
			Amount:      order.Amount,
			NetAmount:   order.NetAmount,
			Currency:    order.Currency,
			BonusAmount: order.BonusAmount,
			Time:        order.CreatedAt,
			Status:      order.Status,
		}))
	}
	result.PageSize = search.PageSize
	result.CurrentPage = search.CurrentPage
	return result
}

func GetAppWithdrawOrderHistory(db *gorm.DB, userID int64, search pojo.AppOrderHistorySearch) (result pojo.AppOrderHistoryResp) {
	if db == nil || userID <= 0 {
		result.PageSize = search.PageSize
		result.CurrentPage = search.CurrentPage
		return result
	}

	var orders []pojo.WithdrawOrderBr
	query := db.Model(&pojo.WithdrawOrderBr{}).Where("user_id = ?", userID)
	query.Count(&result.Total)
	query.Order("id desc").
		Limit(search.PageSize).
		Offset(search.PageSize * search.CurrentPage).
		Find(&orders)

	result.List = make([]pojo.AppOrderHistoryItem, 0, len(orders))
	for _, order := range orders {
		result.List = append(result.List, buildAppWithdrawOrderHistoryItem(withdrawOrderHistorySource{
			OrderNo:      order.OrderNo,
			Amount:       order.Amount,
			NetAmount:    order.NetAmount,
			Currency:     order.Currency,
			Fee:          order.Fee,
			RejectReason: order.FailMsg,
			Time:         order.CreatedAt,
			Status:       order.Status,
		}))
	}
	result.PageSize = search.PageSize
	result.CurrentPage = search.CurrentPage
	return result
}

func buildAppRechargeOrderHistoryItem(order rechargeOrderHistorySource) pojo.AppOrderHistoryItem {
	bonusAmount := utils.Truncate2(order.BonusAmount)
	return pojo.AppOrderHistoryItem{
		OrderNo:        order.OrderNo,
		Amount:         utils.Truncate2(order.Amount),
		NetAmount:      utils.Truncate2(order.NetAmount),
		BonusAmount:    &bonusAmount,
		Currency:       strings.ToUpper(strings.TrimSpace(order.Currency)),
		CurrencySymbol: orderCurrencySymbol(order.Currency),
		Time:           order.Time,
		Status:         order.Status,
	}
}

func buildAppWithdrawOrderHistoryItem(order withdrawOrderHistorySource) pojo.AppOrderHistoryItem {
	fee := utils.Truncate2(order.Fee)
	rejectReason := optionalTrimmedString(order.RejectReason)
	return pojo.AppOrderHistoryItem{
		OrderNo:        order.OrderNo,
		Amount:         utils.Truncate2(order.Amount),
		NetAmount:      utils.Truncate2(order.NetAmount),
		Fee:            &fee,
		RejectReason:   rejectReason,
		Currency:       strings.ToUpper(strings.TrimSpace(order.Currency)),
		CurrencySymbol: orderCurrencySymbol(order.Currency),
		Time:           order.Time,
		Status:         order.Status,
	}
}

func optionalTrimmedString(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func orderCurrencySymbol(currency string) string {
	currency = strings.ToUpper(strings.TrimSpace(currency))
	switch currency {
	case "":
		return ""
	case "BRL":
		return "R$"
	case "USD", "MXN":
		return "$"
	case "CNY", "RMB":
		return "¥"
	case "EUR":
		return "€"
	default:
		return currency
	}
}
