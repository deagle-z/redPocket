package services

import (
	"BaseGoUni/core/utils"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/go-telegram/bot"
	"github.com/hibiken/asynq"
)

const (
	TaskTypeTelegramWelcomeMessage = "telegram:welcome_message"
	TelegramWelcomeBatchDelay      = 3 * time.Second
	telegramWelcomeBatchTTL        = 2 * time.Minute
	telegramWelcomeScheduledTTL    = 30 * time.Second
)

type TelegramWelcomeMessagePayload struct {
	ChatID int64    `json:"chatId"`
	Names  []string `json:"names,omitempty"`
}

func (p TelegramWelcomeMessagePayload) valid() bool {
	if p.ChatID == 0 {
		return false
	}
	for _, name := range p.Names {
		if strings.TrimSpace(name) != "" {
			return true
		}
	}
	return false
}

func EnqueueTelegramWelcomeMessageTask(chatID int64, name string) error {
	name = strings.TrimSpace(name)
	if chatID == 0 || name == "" {
		log.Printf("[tg-welcome] welcome send enqueue skip reason=invalid_payload chat_id=%d name=%q", chatID, name)
		return nil
	}
	if asynqClient == nil {
		log.Printf("[tg-welcome] welcome send enqueue_failed reason=asynq_client_nil chat_id=%d name=%q", chatID, name)
		return errors.New("asynq client is nil")
	}
	if err := appendTelegramWelcomeBatchName(chatID, name); err != nil {
		log.Printf("[tg-welcome] welcome send enqueue_failed reason=redis_append_failed chat_id=%d name=%q err=%v", chatID, name, err)
		return err
	}
	return scheduleTelegramWelcomeBatch(chatID, TelegramWelcomeBatchDelay, fmt.Sprintf("name=%q", name))
}

func scheduleTelegramWelcomeBatch(chatID int64, delay time.Duration, reason string) error {
	if asynqClient == nil {
		log.Printf("[tg-welcome] welcome send enqueue_failed reason=asynq_client_nil chat_id=%d schedule_reason=%s", chatID, reason)
		return errors.New("asynq client is nil")
	}
	if utils.RD == nil {
		log.Printf("[tg-welcome] welcome send enqueue_failed reason=redis_client_nil chat_id=%d schedule_reason=%s", chatID, reason)
		return errors.New("redis client is nil")
	}
	scheduledKey := telegramWelcomeBatchScheduledKey(chatID)
	acquired, err := utils.RD.SetNX(context.Background(), scheduledKey, "1", telegramWelcomeScheduledTTL).Result()
	if err != nil {
		log.Printf("[tg-welcome] welcome send enqueue_failed reason=redis_setnx_failed chat_id=%d schedule_reason=%s err=%v", chatID, reason, err)
		return err
	}
	if !acquired {
		log.Printf("[tg-welcome] welcome send batch_appended chat_id=%d schedule_reason=%s note=batch_already_scheduled", chatID, reason)
		return nil
	}
	payload := TelegramWelcomeMessagePayload{ChatID: chatID}
	data, err := json.Marshal(payload)
	if err != nil {
		_ = utils.RD.Del(context.Background(), scheduledKey).Err()
		return err
	}
	task := asynq.NewTask(TaskTypeTelegramWelcomeMessage, data)
	info, err := asynqClient.Enqueue(task,
		asynq.ProcessIn(delay),
		asynq.MaxRetry(20),
		asynq.Timeout(30*time.Second),
	)
	if err != nil {
		_ = utils.RD.Del(context.Background(), scheduledKey).Err()
		log.Printf("[tg-welcome] welcome send enqueue_failed chat_id=%d delay=%s schedule_reason=%s err=%v", chatID, delay, reason, err)
		return err
	}
	taskID := ""
	if info != nil {
		taskID = info.ID
	}
	log.Printf("[tg-welcome] welcome send batch_enqueued chat_id=%d delay=%s task_id=%q schedule_reason=%s", chatID, delay, taskID, reason)
	return nil
}

func appendTelegramWelcomeBatchName(chatID int64, name string) error {
	if utils.RD == nil {
		return errors.New("redis client is nil")
	}
	key := telegramWelcomeBatchNamesKey(chatID)
	if err := utils.RD.RPush(context.Background(), key, name).Err(); err != nil {
		return err
	}
	return utils.RD.Expire(context.Background(), key, telegramWelcomeBatchTTL).Err()
}

func telegramWelcomeBatchNamesKey(chatID int64) string {
	return fmt.Sprintf("bgu_tg_welcome_batch_names_%d", chatID)
}

func telegramWelcomeBatchScheduledKey(chatID int64) string {
	return fmt.Sprintf("bgu_tg_welcome_batch_scheduled_%d", chatID)
}

func handleTelegramWelcomeMessageTask(ctx context.Context, task *asynq.Task) error {
	var payload TelegramWelcomeMessagePayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return err
	}
	if payload.ChatID == 0 {
		log.Printf("[tg-welcome] welcome send task skip reason=invalid_payload chat_id=%d", payload.ChatID)
		return nil
	}
	names := drainTelegramWelcomeBatchNames(payload.ChatID)
	if len(names) == 0 {
		names = payload.Names
	}
	text := buildTelegramWelcomeBatchText(names)
	if text == "" {
		log.Printf("[tg-welcome] welcome send task skip reason=empty_batch_text chat_id=%d names=%v", payload.ChatID, names)
		finishTelegramWelcomeBatchSchedule(payload.ChatID)
		return nil
	}
	token := utils.GlobalConfig.Telegram.BotToken
	if token == "" {
		return fmt.Errorf("telegram bot token is empty")
	}
	log.Printf("[tg-welcome] welcome send task_started chat_id=%d names_count=%d text=%q", payload.ChatID, len(names), text)
	b, err := bot.New(token, bot.WithSkipGetMe())
	if err != nil {
		log.Printf("[tg-welcome] welcome send task_bot_init_failed chat_id=%d err=%v", payload.ChatID, err)
		return err
	}
	sent, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: payload.ChatID,
		Text:   text,
	})
	if err != nil {
		requeueTelegramWelcomeBatchNames(payload.ChatID, names)
		log.Printf("[tg-welcome] welcome send task_failed chat_id=%d names_count=%d text=%q err=%v", payload.ChatID, len(names), text, err)
		return err
	}
	sentID := 0
	if sent != nil {
		sentID = sent.ID
	}
	log.Printf("[tg-welcome] welcome send task_sent chat_id=%d sent_message_id=%d names_count=%d text=%q", payload.ChatID, sentID, len(names), text)
	if sentID > 0 {
		if err := EnqueueTelegramDeleteMessageTask(payload.ChatID, sentID, TelegramDeleteMessageDefaultDelay); err != nil {
			log.Printf("[tg-welcome] welcome send delete_enqueue_failed chat_id=%d sent_message_id=%d err=%v", payload.ChatID, sentID, err)
		}
	}
	finishTelegramWelcomeBatchSchedule(payload.ChatID)
	return nil
}

func finishTelegramWelcomeBatchSchedule(chatID int64) {
	if chatID == 0 || utils.RD == nil {
		return
	}
	_ = utils.RD.Del(context.Background(), telegramWelcomeBatchScheduledKey(chatID)).Err()
	pending, err := utils.RD.LLen(context.Background(), telegramWelcomeBatchNamesKey(chatID)).Result()
	if err != nil {
		log.Printf("[tg-welcome] welcome send followup_check_failed chat_id=%d err=%v", chatID, err)
		return
	}
	if pending <= 0 {
		return
	}
	if err := scheduleTelegramWelcomeBatch(chatID, TelegramWelcomeBatchDelay, fmt.Sprintf("pending_names=%d", pending)); err != nil {
		log.Printf("[tg-welcome] welcome send followup_enqueue_failed chat_id=%d pending_names=%d err=%v", chatID, pending, err)
	}
}

func drainTelegramWelcomeBatchNames(chatID int64) []string {
	if utils.RD == nil {
		return nil
	}
	key := telegramWelcomeBatchNamesKey(chatID)
	names, err := utils.RD.LRange(context.Background(), key, 0, -1).Result()
	if err != nil {
		log.Printf("[tg-welcome] welcome send drain_failed chat_id=%d err=%v", chatID, err)
		return nil
	}
	_ = utils.RD.Del(context.Background(), key).Err()
	return normalizeTelegramWelcomeNames(names)
}

func requeueTelegramWelcomeBatchNames(chatID int64, names []string) {
	if utils.RD == nil || len(names) == 0 {
		return
	}
	key := telegramWelcomeBatchNamesKey(chatID)
	values := make([]interface{}, 0, len(names))
	for _, name := range normalizeTelegramWelcomeNames(names) {
		values = append(values, name)
	}
	if len(values) == 0 {
		return
	}
	if err := utils.RD.LPush(context.Background(), key, values...).Err(); err != nil {
		log.Printf("[tg-welcome] welcome send requeue_failed chat_id=%d names=%v err=%v", chatID, names, err)
		return
	}
	_ = utils.RD.Expire(context.Background(), key, telegramWelcomeBatchTTL).Err()
}

func buildTelegramWelcomeBatchText(names []string) string {
	names = normalizeTelegramWelcomeNames(names)
	if len(names) == 0 {
		return ""
	}
	return fmt.Sprintf("¡Bienvenido %s a unirte al canal oficial de LuckyCoins!", strings.Join(names, ", "))
}

func normalizeTelegramWelcomeNames(names []string) []string {
	result := make([]string, 0, len(names))
	for _, name := range names {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		result = append(result, name)
	}
	return result
}
