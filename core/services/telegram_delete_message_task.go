package services

import (
	"BaseGoUni/core/utils"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/go-telegram/bot"
	"github.com/hibiken/asynq"
)

const (
	TaskTypeTelegramDeleteMessage     = "telegram:delete_message"
	TelegramDeleteMessageDefaultDelay = 5 * time.Minute
)

type TelegramDeleteMessagePayload struct {
	ChatID    int64 `json:"chatId"`
	MessageID int   `json:"messageId"`
}

func (p TelegramDeleteMessagePayload) valid() bool {
	return p.ChatID != 0 && p.MessageID != 0
}

func EnqueueTelegramDeleteMessageTask(chatID int64, messageID int, delay time.Duration) error {
	if delay <= 0 {
		delay = TelegramDeleteMessageDefaultDelay
	}
	payload := TelegramDeleteMessagePayload{
		ChatID:    chatID,
		MessageID: messageID,
	}
	if !payload.valid() {
		log.Printf("[tg-welcome] welcome delete enqueue skip reason=invalid_payload chat_id=%d message_id=%d delay=%s", chatID, messageID, delay)
		return nil
	}
	if asynqClient == nil {
		log.Printf("[tg-welcome] welcome delete enqueue_failed reason=asynq_client_nil chat_id=%d message_id=%d delay=%s", chatID, messageID, delay)
		return errors.New("asynq client is nil")
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	task := asynq.NewTask(TaskTypeTelegramDeleteMessage, data)
	info, err := asynqClient.Enqueue(task, asynq.ProcessIn(delay), asynq.MaxRetry(10))
	if err != nil {
		log.Printf("[tg-welcome] welcome delete enqueue_failed chat_id=%d message_id=%d delay=%s err=%v", chatID, messageID, delay, err)
		return err
	}
	taskID := ""
	if info != nil {
		taskID = info.ID
	}
	log.Printf("[tg-welcome] welcome delete enqueued chat_id=%d message_id=%d delay=%s task_id=%q", chatID, messageID, delay, taskID)
	return nil
}

func handleTelegramDeleteMessageTask(ctx context.Context, task *asynq.Task) error {
	var payload TelegramDeleteMessagePayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return err
	}
	if !payload.valid() {
		log.Printf("[tg-welcome] welcome delete task skip reason=invalid_payload chat_id=%d message_id=%d", payload.ChatID, payload.MessageID)
		return nil
	}
	token := utils.GlobalConfig.Telegram.BotToken
	if token == "" {
		return fmt.Errorf("telegram bot token is empty")
	}
	log.Printf("[tg-welcome] welcome delete task_started chat_id=%d message_id=%d", payload.ChatID, payload.MessageID)
	b, err := bot.New(token, bot.WithSkipGetMe())
	if err != nil {
		log.Printf("[tg-welcome] welcome delete task_bot_init_failed chat_id=%d message_id=%d err=%v", payload.ChatID, payload.MessageID, err)
		return err
	}
	_, err = b.DeleteMessage(ctx, &bot.DeleteMessageParams{
		ChatID:    payload.ChatID,
		MessageID: payload.MessageID,
	})
	if err != nil {
		log.Printf("[tg-welcome] welcome delete task_failed chat_id=%d message_id=%d err=%v", payload.ChatID, payload.MessageID, err)
		return err
	}
	log.Printf("[tg-welcome] welcome delete task_deleted chat_id=%d message_id=%d", payload.ChatID, payload.MessageID)
	return nil
}
