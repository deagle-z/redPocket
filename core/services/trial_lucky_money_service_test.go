package services

import (
	"BaseGoUni/core/pojo"
	"testing"
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
