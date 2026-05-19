package repository

import (
	"math"
	"testing"
	"time"
)

func TestParseCheckInRewardsConfigRejectsEmptyConfig(t *testing.T) {
	_, err := ParseCheckInRewardsConfig("")
	if err == nil || err.Error() != "checkin_rewards_config_invalid" {
		t.Fatalf("ParseCheckInRewardsConfig() err = %v, want checkin_rewards_config_invalid", err)
	}
}

func TestParseCheckInRewardsConfigRejectsInvalidJSON(t *testing.T) {
	_, err := ParseCheckInRewardsConfig("1,2,3")
	if err == nil || err.Error() != "checkin_rewards_config_invalid" {
		t.Fatalf("ParseCheckInRewardsConfig() err = %v, want checkin_rewards_config_invalid", err)
	}
}

func TestParseCheckInRewardsConfigRejectsNonPositiveReward(t *testing.T) {
	_, err := ParseCheckInRewardsConfig("[1,0,3]")
	if err == nil || err.Error() != "checkin_rewards_config_invalid" {
		t.Fatalf("ParseCheckInRewardsConfig() err = %v, want checkin_rewards_config_invalid", err)
	}
}

func TestParseCheckInRewardsConfigTruncatesRewards(t *testing.T) {
	rewards, err := ParseCheckInRewardsConfig("[1,2.567,3]")
	if err != nil {
		t.Fatalf("ParseCheckInRewardsConfig() err = %v, want nil", err)
	}
	if len(rewards) != 3 {
		t.Fatalf("len(rewards) = %d, want 3", len(rewards))
	}
	if math.Abs(rewards[1]-2.56) > 0.000001 {
		t.Fatalf("rewards[1] = %.2f, want 2.56", rewards[1])
	}
}

func TestCheckInDateInNewYorkUsesNewYorkDay(t *testing.T) {
	utcTime := time.Date(2026, 5, 20, 3, 30, 0, 0, time.UTC)

	got := CheckInDateInNewYork(utcTime)

	if got != "2026-05-19" {
		t.Fatalf("CheckInDateInNewYork() = %q, want 2026-05-19", got)
	}
}

func TestCheckInAwardUniUsesNewYorkDate(t *testing.T) {
	utcTime := time.Date(2026, 5, 20, 3, 30, 0, 0, time.UTC)

	got := BuildCheckInAwardUni(1001, utcTime)

	if got != "checkin_1001_20260519" {
		t.Fatalf("BuildCheckInAwardUni() = %q, want checkin_1001_20260519", got)
	}
}
