package services

import (
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/utils"
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/go-telegram/bot"
)

var (
	tenantNameCache   = map[int64]string{}
	tenantNameCacheMu sync.RWMutex
)

// resolveTenantName 按租户ID查询租户名称（sys_tenant 在默认主库 utils.Db），命中后缓存。
func resolveTenantName(tenantID int64) string {
	if tenantID <= 0 || utils.Db == nil {
		return ""
	}
	tenantNameCacheMu.RLock()
	name, ok := tenantNameCache[tenantID]
	tenantNameCacheMu.RUnlock()
	if ok {
		return name
	}
	var tenant pojo.SysTenant
	if err := utils.Db.Select("tenant_name").Where("id = ?", tenantID).First(&tenant).Error; err == nil {
		name = strings.TrimSpace(tenant.TenantName)
	}
	if name != "" {
		tenantNameCacheMu.Lock()
		tenantNameCache[tenantID] = name
		tenantNameCacheMu.Unlock()
	}
	return name
}

// tenantLabel 返回用于通知展示的「所属商户」标签：有名称用名称，否则用 #ID。
func tenantLabel(tenantID int64) string {
	if name := resolveTenantName(tenantID); name != "" {
		return name
	}
	return fmt.Sprintf("#%d", tenantID)
}

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
	tenant := tenantLabel(payload.TenantID)
	switch payload.Event {
	case utils.TelegramNotifyEventRegister:
		return fmt.Sprintf("注册通知\n所属商户: %s\n用户ID: %d\nUID: %s\n账号: %s\n姓名: %s\n国家: %s\nIP: %s\n来源渠道: %s\n时间: %s",
			tenant,
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
		return fmt.Sprintf("充值通知\n所属商户: %s\n用户ID: %d\nUID: %s\n订单号: %s\n金额: %.2f %s\n入账: %.2f\n赠送: %.2f\n渠道: %s\n时间: %s",
			tenant,
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
		return fmt.Sprintf("提现通知\n所属商户: %s\n用户ID: %d\nUID: %s\n订单号: %s\n金额: %.2f %s\n手续费: %.2f\n渠道: %s\n时间: %s",
			tenant,
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
