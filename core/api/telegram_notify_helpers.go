package api

import (
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/utils"
	"log"
	"strings"
)

func notifyTelegramWithdraw(user pojo.TgUser, order pojo.WithdrawOrderBrBack) {
	dispatchAPITelegramNotify(utils.TelegramNotifyPayload{
		Event:             utils.TelegramNotifyEventWithdraw,
		UserID:            user.ID,
		TenantID:          user.TenantId,
		UID:               user.Uid,
		TgID:              user.TgID,
		Username:          apiStringValue(user.Username),
		TgName:            apiStringValue(user.TgName),
		FirstName:         apiStringValue(user.FirstName),
		Email:             strings.TrimSpace(user.Email),
		Phone:             apiStringValue(user.Phone),
		Country:           firstAPIString(order.CountryCode, apiStringValue(user.Country)),
		SourceChannelCode: apiStringValue(user.SourceChannelCode),
		OrderNo:           order.OrderNo,
		Amount:            order.Amount,
		Fee:               order.Fee,
		Currency:          order.Currency,
		Channel:           order.Channel,
		PayMethod:         apiStringValue(order.PayMethod),
		CreatedAt:         order.CreatedAt,
	})
}

func dispatchAPITelegramNotify(payload utils.TelegramNotifyPayload) {
	go func() {
		if err := utils.SendTelegramNotify(payload); err != nil {
			log.Printf("[telegram-notify] dispatch skipped event=%s userID=%d orderNo=%s err=%v", payload.Event, payload.UserID, payload.OrderNo, err)
		}
	}()
}

func apiStringValue(value *string) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}

func firstAPIString(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
