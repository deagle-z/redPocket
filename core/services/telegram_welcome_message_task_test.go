package services

import "testing"

func TestTelegramWelcomeMessagePayloadValid(t *testing.T) {
	payload := TelegramWelcomeMessagePayload{
		ChatID: -100123456,
		Names:  []string{"Alice"},
	}

	if !payload.valid() {
		t.Fatalf("payload.valid() = false, want true")
	}
}

func TestTelegramWelcomeMessagePayloadInvalid(t *testing.T) {
	tests := []TelegramWelcomeMessagePayload{
		{},
		{ChatID: -100123456},
		{Names: []string{"Alice"}},
	}

	for _, payload := range tests {
		if payload.valid() {
			t.Fatalf("payload.valid() = true for %+v, want false", payload)
		}
	}
}

func TestBuildTelegramWelcomeBatchText(t *testing.T) {
	got := buildTelegramWelcomeBatchText([]string{"Alice", "Bob", "@carol"})
	want := "¡Bienvenido Alice, Bob, @carol a unirte al canal oficial de LuckyCoins!"
	if got != want {
		t.Fatalf("buildTelegramWelcomeBatchText() = %q, want %q", got, want)
	}
}

func TestBuildTelegramWelcomeBatchTextSkipsBlankNames(t *testing.T) {
	got := buildTelegramWelcomeBatchText([]string{" Alice ", "", "Bob"})
	want := "¡Bienvenido Alice, Bob a unirte al canal oficial de LuckyCoins!"
	if got != want {
		t.Fatalf("buildTelegramWelcomeBatchText() = %q, want %q", got, want)
	}
}
