package services

import (
	"testing"

	"github.com/go-telegram/bot/models"
)

func TestBuildRequiredChannelWelcomeMessages(t *testing.T) {
	message := &models.Message{
		Chat: models.Chat{
			ID:       -100123456,
			Username: "LuckyCoinsOfficial",
		},
		NewChatMembers: []models.User{
			{ID: 1, Username: "alice"},
			{ID: 2, FirstName: "Bob", LastName: "Smith"},
		},
	}

	messages := buildRequiredChannelWelcomeMessages("@LuckyCoinsOfficial", message)

	want := []string{
		"¡Bienvenido @alice a unirte al canal oficial de LuckyCoins!",
		"¡Bienvenido Bob Smith a unirte al canal oficial de LuckyCoins!",
	}
	if len(messages) != len(want) {
		t.Fatalf("welcome messages len = %d, want %d (%v)", len(messages), len(want), messages)
	}
	for i := range want {
		if messages[i] != want[i] {
			t.Fatalf("welcome message[%d] = %q, want %q", i, messages[i], want[i])
		}
	}
}

func TestBuildRequiredChannelWelcomeMessagesSkipsOtherChats(t *testing.T) {
	message := &models.Message{
		Chat: models.Chat{
			ID:       -100123456,
			Username: "OtherChannel",
		},
		NewChatMembers: []models.User{
			{ID: 1, Username: "alice"},
		},
	}

	messages := buildRequiredChannelWelcomeMessages("@LuckyCoinsOfficial", message)

	if len(messages) != 0 {
		t.Fatalf("welcome messages len = %d, want 0 (%v)", len(messages), messages)
	}
}
