package repository

import (
	"testing"
	"time"
)

func TestParseFirstRechargeGiftConfigV2(t *testing.T) {
	cfg, err := parseFirstRechargeGiftConfigV2("20|30|50")
	if err != nil {
		t.Fatalf("parseFirstRechargeGiftConfigV2() error = %v", err)
	}
	if len(cfg.Rates) != 3 {
		t.Fatalf("len(Rates) = %d, want 3", len(cfg.Rates))
	}
	want := []float64{20, 30, 50}
	for i := range want {
		if cfg.Rates[i] != want[i] {
			t.Fatalf("Rates[%d] = %.2f, want %.2f", i, cfg.Rates[i], want[i])
		}
	}
}

func TestParseFirstRechargeGiftConfigV2RejectsInvalidConfig(t *testing.T) {
	tests := []string{"", "20", "20(20|30|50)", "20|0|50", "20|-30|50", "20|30|abc"}
	for _, tt := range tests {
		t.Run(tt, func(t *testing.T) {
			if _, err := parseFirstRechargeGiftConfigV2(tt); err == nil {
				t.Fatalf("parseFirstRechargeGiftConfigV2(%q) expected error", tt)
			}
		})
	}
}

func TestFirstRechargeGiftV2DayIndex(t *testing.T) {
	loc := time.FixedZone("test", 8*60*60)
	start := time.Date(2026, 5, 1, 23, 30, 0, 0, loc)
	tests := []struct {
		name string
		pay  time.Time
		want int
		ok   bool
	}{
		{name: "same natural day", pay: time.Date(2026, 5, 1, 23, 59, 0, 0, loc), want: 1, ok: true},
		{name: "next natural day", pay: time.Date(2026, 5, 2, 0, 1, 0, 0, loc), want: 2, ok: true},
		{name: "third natural day", pay: time.Date(2026, 5, 3, 12, 0, 0, 0, loc), want: 3, ok: true},
		{name: "fourth natural day", pay: time.Date(2026, 5, 4, 0, 0, 0, 0, loc), ok: false},
		{name: "before start", pay: time.Date(2026, 4, 30, 23, 59, 0, 0, loc), ok: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := firstRechargeGiftV2DayIndex(start, tt.pay)
			if ok != tt.ok {
				t.Fatalf("ok = %v, want %v", ok, tt.ok)
			}
			if got != tt.want {
				t.Fatalf("day = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestFirstRechargeGiftV2DayIndexUsesPayTimeLocation(t *testing.T) {
	loc := time.FixedZone("Asia/Shanghai", 8*60*60)
	start := time.Date(2026, 5, 13, 15, 55, 15, 0, time.UTC)
	pay := time.Date(2026, 5, 13, 23, 55, 15, 0, loc)

	got, ok := firstRechargeGiftV2DayIndex(start, pay)
	if !ok {
		t.Fatalf("ok = false, want true")
	}
	if got != 1 {
		t.Fatalf("day = %d, want 1", got)
	}
}

func TestCalculateFirstRechargeGiftV2Amount(t *testing.T) {
	got := calculateFirstRechargeGiftV2Amount(50000, 30)
	if got != 15000 {
		t.Fatalf("calculateFirstRechargeGiftV2Amount() = %.2f, want 15000.00", got)
	}
}

func TestBuildFirstRecharge3DayRatesMarksMissedPreviousDayExpired(t *testing.T) {
	cfg := firstRechargeGiftConfigV2{Rates: []float64{20, 30, 50}}
	statuses := buildFirstRecharge3DayRates(cfg, 3, map[int]bool{1: true})

	if statuses[0].Status != "done" {
		t.Fatalf("day1 status = %q, want done", statuses[0].Status)
	}
	if statuses[1].Status != "expired" {
		t.Fatalf("day2 status = %q, want expired", statuses[1].Status)
	}
	if statuses[2].Status != "expired" {
		t.Fatalf("day3 status = %q, want expired", statuses[2].Status)
	}
}
