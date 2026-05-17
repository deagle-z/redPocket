package services

import (
	"testing"
	"time"
)

func TestTelegramDeleteMessagePayloadValid(t *testing.T) {
	payload := TelegramDeleteMessagePayload{
		ChatID:    -100123456,
		MessageID: 42,
	}

	if !payload.valid() {
		t.Fatalf("payload.valid() = false, want true")
	}
}

func TestTelegramDeleteMessagePayloadInvalid(t *testing.T) {
	tests := []TelegramDeleteMessagePayload{
		{},
		{ChatID: -100123456},
		{MessageID: 42},
	}

	for _, payload := range tests {
		if payload.valid() {
			t.Fatalf("payload.valid() = true for %+v, want false", payload)
		}
	}
}

func TestTelegramDeleteMessageDefaultDelay(t *testing.T) {
	if TelegramDeleteMessageDefaultDelay != 5*time.Minute {
		t.Fatalf("TelegramDeleteMessageDefaultDelay = %v, want %v", TelegramDeleteMessageDefaultDelay, 5*time.Minute)
	}
}
