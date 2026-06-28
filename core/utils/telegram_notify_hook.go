package utils

import (
	"errors"
	"time"
)

const (
	TelegramNotifyEventRegister = "register"
	TelegramNotifyEventRecharge = "recharge"
	TelegramNotifyEventWithdraw = "withdraw"
)

type TelegramNotifyPayload struct {
	Event             string
	UserID            int64
	TenantID          int64
	UID               string
	TgID              int64
	Username          string
	TgName            string
	FirstName         string
	Email             string
	Phone             string
	Country           string
	IP                string
	Region            string
	SourceChannelCode string
	OrderNo           string
	Amount            float64
	Fee               float64
	CreditAmount      float64
	BonusAmount       float64
	Currency          string
	Channel           string
	PayMethod         string
	ProviderTradeNo   string
	WithdrawSource    string
	CreatedAt         time.Time
}

var telegramNotify func(TelegramNotifyPayload)

func RegisterTelegramNotify(fn func(TelegramNotifyPayload)) {
	telegramNotify = fn
}

func SendTelegramNotify(payload TelegramNotifyPayload) error {
	if telegramNotify == nil {
		return errors.New("telegram notifier not ready")
	}
	telegramNotify(payload)
	return nil
}
