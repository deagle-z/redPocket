package services

import (
	"testing"
	"time"
)

func TestGGRTransactionCleanupCutoffUsesTwentyFourHourRetention(t *testing.T) {
	now := time.Date(2026, time.August, 16, 12, 30, 0, 0, time.UTC)
	want := time.Date(2026, time.August, 15, 12, 30, 0, 0, time.UTC)
	if got := GGRTransactionCleanupCutoff(now); !got.Equal(want) {
		t.Fatalf("GGRTransactionCleanupCutoff() = %s, want %s", got, want)
	}
}
