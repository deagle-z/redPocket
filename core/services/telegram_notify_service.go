package services

import (
	"BaseGoUni/core/utils"
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/go-telegram/bot"
)

var telegramNotifyBot *bot.Bot

func InitTelegramNotifier(botToken string) error {
	if strings.TrimSpace(botToken) == "" {
		log.Println("Telegram Bot Token 未配置，跳过通知初始化")
		return nil
	}
	if utils.GlobalConfig.Telegram.NotifyGroupID == 0 {
		log.Println("telegram.notifyGroupId 未配置，跳过通知初始化")
		return nil
	}

	b, err := bot.New(botToken)
	if err != nil {
		return fmt.Errorf("初始化 Telegram 通知失败: %v", err)
	}
	telegramNotifyBot = b
	utils.RegisterTelegramNotify(sendTelegramNotify)
	log.Printf("Telegram 通知初始化完成 notifyGroupId=%d", utils.GlobalConfig.Telegram.NotifyGroupID)
	return nil
}

func sendTelegramNotify(payload utils.TelegramNotifyPayload) {
	if telegramNotifyBot == nil || utils.GlobalConfig.Telegram.NotifyGroupID == 0 {
		return
	}
	text := buildTelegramNotifyText(payload)
	if strings.TrimSpace(text) == "" {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	if _, err := telegramNotifyBot.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: utils.GlobalConfig.Telegram.NotifyGroupID,
		Text:   text,
	}); err != nil {
		log.Printf("[telegram-notify] send failed event=%s userID=%d orderNo=%s err=%v", payload.Event, payload.UserID, payload.OrderNo, err)
	}
}

func buildTelegramNotifyText(payload utils.TelegramNotifyPayload) string {
	at := payload.CreatedAt
	if at.IsZero() {
		at = time.Now()
	}
	switch payload.Event {
	case utils.TelegramNotifyEventRegister:
		return fmt.Sprintf("注册通知\n用户ID: %d\nUID: %s\n账号: %s\n姓名: %s\n国家: %s\nIP: %s\n来源渠道: %s\n时间: %s",
			payload.UserID,
			payload.UID,
			firstNonEmpty(payload.Email, payload.Phone, payload.TgName, payload.Username),
			firstNonEmpty(payload.FirstName, "-"),
			firstNonEmpty(payload.Country, payload.Region, "-"),
			firstNonEmpty(payload.IP, "-"),
			firstNonEmpty(payload.SourceChannelCode, "-"),
			at.Format("2006-01-02 15:04:05"),
		)
	case utils.TelegramNotifyEventRecharge:
		return fmt.Sprintf("充值通知\n用户ID: %d\nUID: %s\n订单号: %s\n金额: %.2f %s\n入账: %.2f\n赠送: %.2f\n渠道: %s\n时间: %s",
			payload.UserID,
			payload.UID,
			payload.OrderNo,
			utils.Truncate2(payload.Amount),
			firstNonEmpty(payload.Currency, "-"),
			utils.Truncate2(payload.CreditAmount),
			utils.Truncate2(payload.BonusAmount),
			firstNonEmpty(payload.Channel, payload.PayMethod, "-"),
			at.Format("2006-01-02 15:04:05"),
		)
	case utils.TelegramNotifyEventWithdraw:
		return fmt.Sprintf("提现通知\n用户ID: %d\nUID: %s\n订单号: %s\n金额: %.2f %s\n手续费: %.2f\n渠道: %s\n时间: %s",
			payload.UserID,
			payload.UID,
			payload.OrderNo,
			utils.Truncate2(payload.Amount),
			firstNonEmpty(payload.Currency, "-"),
			utils.Truncate2(payload.Fee),
			firstNonEmpty(payload.Channel, "-"),
			at.Format("2006-01-02 15:04:05"),
		)
	default:
		return ""
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
