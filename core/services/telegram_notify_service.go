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

const (
	telegramNotifyInitAttempts = 3
	telegramNotifySendAttempts = 3
	telegramNotifyRetryDelay   = 500 * time.Millisecond
	telegramNotifySendTimeout  = 8 * time.Second
)

func InitTelegramNotifier(botToken string) error {
	if strings.TrimSpace(botToken) == "" {
		log.Println("Telegram Bot Token 未配置，跳过通知初始化")
		return nil
	}
	if utils.GlobalConfig.Telegram.NotifyGroupID == 0 {
		log.Println("telegram.notifyGroupId 未配置，跳过通知初始化")
		return nil
	}

	var lastErr error
	for attempt := 1; attempt <= telegramNotifyInitAttempts; attempt++ {
		b, err := bot.New(botToken, bot.WithSkipGetMe())
		if err == nil {
			telegramNotifyBot = b
			utils.RegisterTelegramNotify(sendTelegramNotify)
			log.Printf("Telegram 通知初始化完成 notifyGroupId=%d", utils.GlobalConfig.Telegram.NotifyGroupID)
			return nil
		}

		lastErr = err
		if attempt < telegramNotifyInitAttempts {
			log.Printf("[telegram-notify] init retry attempt=%d/%d err=%v", attempt, telegramNotifyInitAttempts, err)
			time.Sleep(time.Duration(attempt) * telegramNotifyRetryDelay)
		}
	}

	return fmt.Errorf("初始化 Telegram 通知失败: %v", lastErr)
}

func sendTelegramNotify(payload utils.TelegramNotifyPayload) {
	if telegramNotifyBot == nil || utils.GlobalConfig.Telegram.NotifyGroupID == 0 {
		return
	}
	text := buildTelegramNotifyText(payload)
	if strings.TrimSpace(text) == "" {
		return
	}

	var lastErr error
	for attempt := 1; attempt <= telegramNotifySendAttempts; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), telegramNotifySendTimeout)
		_, err := telegramNotifyBot.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: utils.GlobalConfig.Telegram.NotifyGroupID,
			Text:   text,
		})
		cancel()
		if err == nil {
			return
		}

		lastErr = err
		if attempt < telegramNotifySendAttempts {
			log.Printf("[telegram-notify] send retry event=%s userID=%d orderNo=%s attempt=%d/%d err=%v", payload.Event, payload.UserID, payload.OrderNo, attempt, telegramNotifySendAttempts, err)
			time.Sleep(time.Duration(attempt) * telegramNotifyRetryDelay)
		}
	}

	if lastErr != nil {
		log.Printf("[telegram-notify] send failed event=%s userID=%d orderNo=%s attempts=%d err=%v", payload.Event, payload.UserID, payload.OrderNo, telegramNotifySendAttempts, lastErr)
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
		actualAmount := rechargeNotifyActualAmount(payload)
		return fmt.Sprintf("充值通知\n所属商户: %s\n用户ID: %d\nUID: %s\n订单号: %s\n实际入款: %.2f %s\n赠送: %.2f\n渠道: %s\n时间: %s",
			tenant,
			payload.UserID,
			payload.UID,
			payload.OrderNo,
			actualAmount,
			firstNonEmpty(payload.Currency, "-"),
			utils.Truncate2(payload.BonusAmount),
			firstNonEmpty(payload.Channel, payload.PayMethod, "-"),
			at.Format("2006-01-02 15:04:05"),
		)
	case utils.TelegramNotifyEventWithdraw:
		return fmt.Sprintf("提现通知\n所属商户: %s\n提现类型: %s\n用户ID: %d\nUID: %s\n订单号: %s\n金额: %.2f %s\n手续费: %.2f\n渠道: %s\n时间: %s",
			tenant,
			withdrawSourceLabel(payload.WithdrawSource),
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

func rechargeNotifyActualAmount(payload utils.TelegramNotifyPayload) float64 {
	if payload.CreditAmount > 0 {
		return utils.Truncate2(payload.CreditAmount)
	}
	return utils.Truncate2(payload.Amount)
}

func withdrawSourceLabel(source string) string {
	switch strings.ToLower(strings.TrimSpace(source)) {
	case "rebate":
		return "佣金提现"
	default:
		return "普通提现"
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
