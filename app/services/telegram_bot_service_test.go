package services

import (
	"testing"
	"time"

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
		"¡Bienvenido Bob a unirte al canal oficial de LuckyCoins!",
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

func TestShouldSendChatMemberWelcome(t *testing.T) {
	tests := []struct {
		name         string
		activeExists bool
		oldActive    bool
		newActive    bool
		want         bool
	}{
		{name: "new active member not active in database", activeExists: false, oldActive: false, newActive: true, want: true},
		{name: "previously left member rejoins", activeExists: false, oldActive: false, newActive: true, want: true},
		{name: "active member already in database", activeExists: true, oldActive: false, newActive: true, want: false},
		{name: "new inactive member", activeExists: false, oldActive: false, newActive: false, want: false},
		{name: "already active transition", activeExists: false, oldActive: true, newActive: true, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shouldSendChatMemberWelcome(tt.activeExists, tt.oldActive, tt.newActive)
			if got != tt.want {
				t.Fatalf("shouldSendChatMemberWelcome() = %t, want %t", got, tt.want)
			}
		})
	}
}

func TestWelcomeMessageDeleteDelay(t *testing.T) {
	if welcomeMessageDeleteDelay != 5*time.Minute {
		t.Fatalf("welcomeMessageDeleteDelay = %v, want %v", welcomeMessageDeleteDelay, 5*time.Minute)
	}
}
