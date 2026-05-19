package services

import (
	"BaseGoUni/core/pojo"
	"testing"
	"time"
)

func TestCanTrialUserGrabSenderAllowsOwnUserPacket(t *testing.T) {
	if !canTrialUserGrabSender(pojo.TrialActorUser, 10, 10) {
		t.Fatalf("canTrialUserGrabSender() = false, want true")
	}
}

func TestCanTrialBotGrabSenderRejectsOwnBotPacket(t *testing.T) {
	if canTrialBotGrabSender(pojo.TrialActorBot, 10, 10) {
		t.Fatalf("canTrialBotGrabSender() = true, want false")
	}
}

func TestTrialUserDefaultBalanceIsOneThousand(t *testing.T) {
	if pojo.TrialUserDefaultBalance != 1000 {
		t.Fatalf("TrialUserDefaultBalance = %.2f, want 1000", pojo.TrialUserDefaultBalance)
	}
}

func TestResolveTrialDailyBalanceRefreshTopsUpLowBalanceOncePerDay(t *testing.T) {
	now := time.Date(2026, 5, 19, 10, 0, 0, 0, time.Local)

	balance, refreshed, shouldMark := resolveTrialDailyBalanceRefresh(800, nil, now)

	if balance != pojo.TrialUserDefaultBalance || !refreshed || !shouldMark {
		t.Fatalf("resolveTrialDailyBalanceRefresh() = (%.2f,%v,%v), want (%.2f,true,true)", balance, refreshed, shouldMark, pojo.TrialUserDefaultBalance)
	}
}

func TestResolveTrialDailyBalanceRefreshKeepsHighBalanceButMarksChecked(t *testing.T) {
	now := time.Date(2026, 5, 19, 10, 0, 0, 0, time.Local)

	balance, refreshed, shouldMark := resolveTrialDailyBalanceRefresh(1200, nil, now)

	if balance != 1200 || refreshed || !shouldMark {
		t.Fatalf("resolveTrialDailyBalanceRefresh() = (%.2f,%v,%v), want (1200,false,true)", balance, refreshed, shouldMark)
	}
}

func TestResolveTrialDailyBalanceRefreshSkipsWhenAlreadyCheckedToday(t *testing.T) {
	now := time.Date(2026, 5, 19, 10, 0, 0, 0, time.Local)
	checkedAt := time.Date(2026, 5, 19, 1, 0, 0, 0, time.Local)

	balance, refreshed, shouldMark := resolveTrialDailyBalanceRefresh(200, &checkedAt, now)

	if balance != 200 || refreshed || shouldMark {
		t.Fatalf("resolveTrialDailyBalanceRefresh() = (%.2f,%v,%v), want (200,false,false)", balance, refreshed, shouldMark)
	}
}

func TestValidateTrialSendAmountRejectsAboveMax(t *testing.T) {
	if err := validateTrialSendAmount(501); err == nil || err.Error() != "amount_max_500" {
		t.Fatalf("validateTrialSendAmount(501) error = %v, want amount_max_500", err)
	}
}

func TestValidateTrialSendAmountAllowsMax(t *testing.T) {
	if err := validateTrialSendAmount(500); err != nil {
		t.Fatalf("validateTrialSendAmount(500) error = %v, want nil", err)
	}
}

func TestCapTrialLuckySendAmountLimitsRobotAmount(t *testing.T) {
	if got := capTrialLuckySendAmount(1000); got != pojo.TrialLuckySendMaxAmount {
		t.Fatalf("capTrialLuckySendAmount(1000) = %.2f, want %.2f", got, pojo.TrialLuckySendMaxAmount)
	}
}

func TestPickTrialLuckyItemForUserTargetWinThunderMode(t *testing.T) {
	lucky := pojo.TrialLuckyMoney{GameMode: 0, Thunder: 8}
	items := []pojo.TrialLuckyMoneyItem{
		{SeqNo: 1, Amount: 1.18},
		{SeqNo: 2, Amount: 1.17},
	}

	item, swap, ok := pickTrialLuckyItemForUserTarget(items, lucky, 0, nil, true)

	if !ok {
		t.Fatalf("pickTrialLuckyItemForUserTarget() ok = false, want true")
	}
	if item.SeqNo != 2 {
		t.Fatalf("picked seq = %d, want 2", item.SeqNo)
	}
	if swap != nil {
		t.Fatalf("swap = %#v, want nil", swap)
	}
	if trialIsThunder(lucky, item.Amount, nil) {
		t.Fatalf("picked item is thunder, want win")
	}
}

func TestPickTrialLuckyItemForUserTargetLoseThunderMode(t *testing.T) {
	lucky := pojo.TrialLuckyMoney{GameMode: 0, Thunder: 8}
	items := []pojo.TrialLuckyMoneyItem{
		{SeqNo: 1, Amount: 1.17},
		{SeqNo: 2, Amount: 1.18},
	}

	item, _, ok := pickTrialLuckyItemForUserTarget(items, lucky, 0, nil, false)

	if !ok {
		t.Fatalf("pickTrialLuckyItemForUserTarget() ok = false, want true")
	}
	if item.SeqNo != 2 {
		t.Fatalf("picked seq = %d, want 2", item.SeqNo)
	}
	if !trialIsThunder(lucky, item.Amount, nil) {
		t.Fatalf("picked item is not thunder, want lose")
	}
}

func TestPickTrialLuckyItemForUserTargetWinOddEvenMode(t *testing.T) {
	guessEven := 0
	lucky := pojo.TrialLuckyMoney{GameMode: 1}
	items := []pojo.TrialLuckyMoneyItem{
		{SeqNo: 1, Amount: 1.11},
		{SeqNo: 2, Amount: 1.12},
	}

	item, _, ok := pickTrialLuckyItemForUserTarget(items, lucky, 0, &guessEven, true)

	if !ok {
		t.Fatalf("pickTrialLuckyItemForUserTarget() ok = false, want true")
	}
	if item.SeqNo != 2 {
		t.Fatalf("picked seq = %d, want 2", item.SeqNo)
	}
	if trialIsThunder(lucky, item.Amount, &guessEven) {
		t.Fatalf("picked item is thunder, want odd/even guess win")
	}
}

func TestPickTrialLuckyItemForUserTargetSwapsFixedGrabIndexAmount(t *testing.T) {
	lucky := pojo.TrialLuckyMoney{GameMode: 0, Thunder: 8}
	items := []pojo.TrialLuckyMoneyItem{
		{ID: 10, SeqNo: 1, Amount: 1.18},
		{ID: 20, SeqNo: 2, Amount: 1.17},
	}

	item, swap, ok := pickTrialLuckyItemForUserTarget(items, lucky, 1, nil, true)

	if !ok {
		t.Fatalf("pickTrialLuckyItemForUserTarget() ok = false, want true")
	}
	if item.SeqNo != 1 {
		t.Fatalf("picked seq = %d, want fixed seq 1", item.SeqNo)
	}
	if item.Amount != 1.17 {
		t.Fatalf("picked amount = %.2f, want swapped win amount 1.17", item.Amount)
	}
	if swap == nil {
		t.Fatalf("swap = nil, want amount swap")
	}
	if swap.PrimaryID != 10 || swap.MatchID != 20 || swap.PrimaryAmount != 1.17 || swap.MatchAmount != 1.18 {
		t.Fatalf("swap = %#v, want primary 10->1.17 and match 20->1.18", swap)
	}
	if trialIsThunder(lucky, item.Amount, nil) {
		t.Fatalf("picked swapped item is thunder, want win")
	}
}

func TestPickTrialLuckyItemForUserTargetFallsBackWhenNoMatch(t *testing.T) {
	lucky := pojo.TrialLuckyMoney{GameMode: 0, Thunder: 8}
	items := []pojo.TrialLuckyMoneyItem{
		{SeqNo: 1, Amount: 1.18},
		{SeqNo: 2, Amount: 2.28},
	}

	item, swap, ok := pickTrialLuckyItemForUserTarget(items, lucky, 0, nil, true)

	if !ok {
		t.Fatalf("pickTrialLuckyItemForUserTarget() ok = false, want true")
	}
	if item.SeqNo != 1 {
		t.Fatalf("picked fallback seq = %d, want 1", item.SeqNo)
	}
	if swap != nil {
		t.Fatalf("swap = %#v, want nil", swap)
	}
}

func TestNormalizeTrialUserWinRate(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want float64
	}{
		{name: "empty uses default", raw: "", want: 0.80},
		{name: "invalid uses default", raw: "abc", want: 0.80},
		{name: "negative uses default", raw: "-0.1", want: 0.80},
		{name: "above one uses default", raw: "1.1", want: 0.80},
		{name: "zero is valid", raw: "0", want: 0},
		{name: "one is valid", raw: "1", want: 1},
		{name: "custom valid", raw: "0.65", want: 0.65},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizeTrialUserWinRate(tt.raw); got != tt.want {
				t.Fatalf("normalizeTrialUserWinRate(%q) = %.2f, want %.2f", tt.raw, got, tt.want)
			}
		})
	}
}

func TestParseTrialLuckyFlowLotteryRewardConfig(t *testing.T) {
	tests := []struct {
		name          string
		raw           string
		wantThreshold float64
		wantCount     int
		wantEnabled   bool
	}{
		{name: "empty disabled", raw: "", wantEnabled: false},
		{name: "invalid disabled", raw: "abc", wantEnabled: false},
		{name: "missing reward disabled", raw: "1000", wantEnabled: false},
		{name: "negative threshold disabled", raw: "-1000:1", wantEnabled: false},
		{name: "zero threshold disabled", raw: "0:1", wantEnabled: false},
		{name: "zero reward disabled", raw: "1000:0", wantEnabled: false},
		{name: "negative reward disabled", raw: "1000:-1", wantEnabled: false},
		{name: "valid", raw: "1000:1", wantThreshold: 1000, wantCount: 1, wantEnabled: true},
		{name: "trims spaces", raw: " 500.5 : 2 ", wantThreshold: 500.5, wantCount: 2, wantEnabled: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			threshold, count, enabled := parseTrialLuckyFlowLotteryRewardConfig(tt.raw)
			if threshold != tt.wantThreshold || count != tt.wantCount || enabled != tt.wantEnabled {
				t.Fatalf("parseTrialLuckyFlowLotteryRewardConfig(%q) = (%.2f,%d,%v), want (%.2f,%d,%v)",
					tt.raw, threshold, count, enabled, tt.wantThreshold, tt.wantCount, tt.wantEnabled)
			}
		})
	}
}

func TestResolveTrialLuckyFlowLotteryRewardConfigUsesDBValue(t *testing.T) {
	threshold, count, enabled := resolveTrialLuckyFlowLotteryRewardConfig("", "1000:1")

	if threshold != 1000 || count != 1 || !enabled {
		t.Fatalf("resolveTrialLuckyFlowLotteryRewardConfig() = (%.2f,%d,%v), want (1000,1,true)", threshold, count, enabled)
	}
}

func TestResolveTrialLuckyFlowLotteryRewardConfigPrefersDBValueOverCacheValue(t *testing.T) {
	threshold, count, enabled := resolveTrialLuckyFlowLotteryRewardConfig("500:2", "1000:1")

	if threshold != 1000 || count != 1 || !enabled {
		t.Fatalf("resolveTrialLuckyFlowLotteryRewardConfig() = (%.2f,%d,%v), want (1000,1,true)", threshold, count, enabled)
	}
}

func TestResolveTrialLuckyFlowLotteryRewardConfigFallsBackToCacheValue(t *testing.T) {
	threshold, count, enabled := resolveTrialLuckyFlowLotteryRewardConfig("500:2", "")

	if threshold != 500 || count != 2 || !enabled {
		t.Fatalf("resolveTrialLuckyFlowLotteryRewardConfig() = (%.2f,%d,%v), want (500,2,true)", threshold, count, enabled)
	}
}

func TestCalculateTrialLuckyFlowLotteryReward(t *testing.T) {
	tests := []struct {
		name        string
		totalFlow   float64
		threshold   float64
		rewardCount int
		alreadySent bool
		wantCount   int
		wantAward   bool
	}{
		{name: "below threshold", totalFlow: 999.99, threshold: 1000, rewardCount: 1, wantCount: 0, wantAward: false},
		{name: "reaches threshold", totalFlow: 1000, threshold: 1000, rewardCount: 1, wantCount: 1, wantAward: true},
		{name: "above threshold still one time", totalFlow: 3000, threshold: 1000, rewardCount: 2, wantCount: 2, wantAward: true},
		{name: "already sent skips", totalFlow: 3000, threshold: 1000, rewardCount: 2, alreadySent: true, wantCount: 0, wantAward: false},
		{name: "disabled threshold", totalFlow: 3000, threshold: 0, rewardCount: 2, wantCount: 0, wantAward: false},
		{name: "disabled reward", totalFlow: 3000, threshold: 1000, rewardCount: 0, wantCount: 0, wantAward: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count, awarded := calculateTrialLuckyFlowLotteryReward(tt.totalFlow, tt.threshold, tt.rewardCount, tt.alreadySent)
			if count != tt.wantCount || awarded != tt.wantAward {
				t.Fatalf("calculateTrialLuckyFlowLotteryReward() = (%d,%v), want (%d,%v)",
					count, awarded, tt.wantCount, tt.wantAward)
			}
		})
	}
}

func TestBuildTrialLuckyFlowLotteryRewardProgress(t *testing.T) {
	tests := []struct {
		name            string
		totalFlow       float64
		threshold       float64
		rewardCount     int
		rewarded        bool
		wantEnabled     bool
		wantRemaining   float64
		wantProgress    float64
		wantCanReward   bool
		wantRewardCount int
	}{
		{name: "disabled", totalFlow: 500, threshold: 0, rewardCount: 1, wantEnabled: false, wantRemaining: 0, wantProgress: 0, wantCanReward: false},
		{name: "below threshold", totalFlow: 250, threshold: 1000, rewardCount: 1, wantEnabled: true, wantRemaining: 750, wantProgress: 25, wantCanReward: false},
		{name: "reaches threshold", totalFlow: 1000, threshold: 1000, rewardCount: 1, wantEnabled: true, wantRemaining: 0, wantProgress: 100, wantCanReward: true, wantRewardCount: 1},
		{name: "above threshold caps progress", totalFlow: 1500, threshold: 1000, rewardCount: 2, wantEnabled: true, wantRemaining: 0, wantProgress: 100, wantCanReward: true, wantRewardCount: 2},
		{name: "rewarded cannot reward again", totalFlow: 1500, threshold: 1000, rewardCount: 2, rewarded: true, wantEnabled: true, wantRemaining: 0, wantProgress: 100, wantCanReward: false, wantRewardCount: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			progress := buildTrialLuckyFlowLotteryRewardProgress(tt.totalFlow, tt.threshold, tt.rewardCount, tt.rewarded)
			if progress.Enabled != tt.wantEnabled || progress.RemainingFlow != tt.wantRemaining || progress.ProgressPercent != tt.wantProgress || progress.CanReward != tt.wantCanReward || progress.AvailableRewardCount != tt.wantRewardCount {
				t.Fatalf("buildTrialLuckyFlowLotteryRewardProgress() = %#v, want enabled=%v remaining=%.2f progress=%.2f canReward=%v rewardCount=%d",
					progress, tt.wantEnabled, tt.wantRemaining, tt.wantProgress, tt.wantCanReward, tt.wantRewardCount)
			}
		})
	}
}
