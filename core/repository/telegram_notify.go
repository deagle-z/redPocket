package repository

import (
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/utils"
	"log"
	"strings"

	"gorm.io/gorm"
)

func notifyTelegramRegister(user pojo.TgUser) {
	dispatchTelegramNotify(utils.TelegramNotifyPayload{
		Event:             utils.TelegramNotifyEventRegister,
		UserID:            user.ID,
		TenantID:          user.TenantId,
		UID:               user.Uid,
		TgID:              user.TgID,
		Username:          valueString(user.Username),
		TgName:            valueString(user.TgName),
		FirstName:         valueString(user.FirstName),
		Email:             strings.TrimSpace(user.Email),
		Phone:             valueString(user.Phone),
		Country:           valueString(user.Country),
		IP:                valueString(user.Ip),
		Region:            valueString(user.Region),
		SourceChannelCode: valueString(user.SourceChannelCode),
		CreatedAt:         user.CreatedAt,
	})
}

func notifyTelegramRechargeByOrderNo(db *gorm.DB, orderNo string) {
	orderNo = strings.TrimSpace(orderNo)
	if orderNo == "" {
		return
	}
	var order pojo.RechargeOrder
	if err := db.Where("order_no = ?", orderNo).First(&order).Error; err != nil || order.ID == 0 || order.Status != 1 {
		log.Printf("[telegram-notify] recharge order not found or not success orderNo=%s err=%v", orderNo, err)
		return
	}
	var user pojo.TgUser
	if err := db.Where("id = ?", order.UserId).First(&user).Error; err != nil || user.ID == 0 {
		log.Printf("[telegram-notify] recharge user not found orderNo=%s userID=%d err=%v", orderNo, order.UserId, err)
		return
	}
	createdAt := order.CreatedAt
	if order.PayTime != nil {
		createdAt = *order.PayTime
	}
	dispatchTelegramNotify(utils.TelegramNotifyPayload{
		Event:             utils.TelegramNotifyEventRecharge,
		UserID:            user.ID,
		TenantID:          user.TenantId,
		UID:               user.Uid,
		TgID:              user.TgID,
		Username:          valueString(user.Username),
		TgName:            valueString(user.TgName),
		FirstName:         valueString(user.FirstName),
		Email:             strings.TrimSpace(user.Email),
		Phone:             valueString(user.Phone),
		Country:           valueString(user.Country),
		SourceChannelCode: valueString(user.SourceChannelCode),
		OrderNo:           order.OrderNo,
		Amount:            order.Amount,
		Fee:               order.Fee,
		CreditAmount:      floatPtrValue(order.CreditAmount),
		BonusAmount:       order.BonusAmount,
		Currency:          order.Currency,
		Channel:           order.Channel,
		PayMethod:         valueString(order.PayMethod),
		ProviderTradeNo:   valueString(order.ProviderTradeNo),
		CreatedAt:         createdAt,
	})
}

func dispatchTelegramNotify(payload utils.TelegramNotifyPayload) {
	go func() {
		if err := utils.SendTelegramNotify(payload); err != nil {
			log.Printf("[telegram-notify] dispatch skipped event=%s userID=%d orderNo=%s err=%v", payload.Event, payload.UserID, payload.OrderNo, err)
		}
	}()
}

func floatPtrValue(value *float64) float64 {
	if value == nil {
		return 0
	}
	return *value
}
