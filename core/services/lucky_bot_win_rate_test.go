package services

import (
	"BaseGoUni/core/pojo"
	"testing"
)

func TestPickLuckyItemForTargetWinThunderMode(t *testing.T) {
	lucky := pojo.LuckyMoney{GameMode: 0, Thunder: 8}
	items := []pojo.LuckyMoneyItem{
		{ID: 10, SeqNo: 1, Amount: 1.18},
		{ID: 20, SeqNo: 2, Amount: 1.17},
	}

	item, swap, ok := pickLuckyItemForTarget(items, lucky, 0, nil, true)

	if !ok {
		t.Fatalf("pickLuckyItemForTarget() ok = false, want true")
	}
	if item.SeqNo != 2 {
		t.Fatalf("picked seq = %d, want 2", item.SeqNo)
	}
	if swap != nil {
		t.Fatalf("swap = %#v, want nil", swap)
	}
	if luckyAmountIsThunder(lucky, item.Amount, nil) {
		t.Fatalf("picked item is thunder, want win")
	}
}

func TestPickLuckyItemForTargetLoseThunderMode(t *testing.T) {
	lucky := pojo.LuckyMoney{GameMode: 0, Thunder: 8}
	items := []pojo.LuckyMoneyItem{
		{ID: 10, SeqNo: 1, Amount: 1.17},
		{ID: 20, SeqNo: 2, Amount: 1.18},
	}

	item, swap, ok := pickLuckyItemForTarget(items, lucky, 0, nil, false)

	if !ok {
		t.Fatalf("pickLuckyItemForTarget() ok = false, want true")
	}
	if item.SeqNo != 2 {
		t.Fatalf("picked seq = %d, want 2", item.SeqNo)
	}
	if swap != nil {
		t.Fatalf("swap = %#v, want nil", swap)
	}
	if !luckyAmountIsThunder(lucky, item.Amount, nil) {
		t.Fatalf("picked item is not thunder, want lose")
	}
}

func TestPickLuckyItemForTargetWinOddEvenMode(t *testing.T) {
	guessEven := 0
	lucky := pojo.LuckyMoney{GameMode: 1}
	items := []pojo.LuckyMoneyItem{
		{ID: 10, SeqNo: 1, Amount: 1.11},
		{ID: 20, SeqNo: 2, Amount: 1.12},
	}

	item, _, ok := pickLuckyItemForTarget(items, lucky, 0, &guessEven, true)

	if !ok {
		t.Fatalf("pickLuckyItemForTarget() ok = false, want true")
	}
	if item.SeqNo != 2 {
		t.Fatalf("picked seq = %d, want 2", item.SeqNo)
	}
	if luckyAmountIsThunder(lucky, item.Amount, &guessEven) {
		t.Fatalf("picked item is thunder, want odd/even guess win")
	}
}

func TestPickLuckyItemForTargetLoseOddEvenMode(t *testing.T) {
	guessEven := 0
	lucky := pojo.LuckyMoney{GameMode: 1}
	items := []pojo.LuckyMoneyItem{
		{ID: 10, SeqNo: 1, Amount: 1.12},
		{ID: 20, SeqNo: 2, Amount: 1.11},
	}

	item, _, ok := pickLuckyItemForTarget(items, lucky, 0, &guessEven, false)

	if !ok {
		t.Fatalf("pickLuckyItemForTarget() ok = false, want true")
	}
	if item.SeqNo != 2 {
		t.Fatalf("picked seq = %d, want 2", item.SeqNo)
	}
	if !luckyAmountIsThunder(lucky, item.Amount, &guessEven) {
		t.Fatalf("picked item is not thunder, want odd/even guess lose")
	}
}

func TestPickLuckyItemForTargetSwapsFixedGrabIndexAmount(t *testing.T) {
	lucky := pojo.LuckyMoney{GameMode: 0, Thunder: 8}
	items := []pojo.LuckyMoneyItem{
		{ID: 10, SeqNo: 1, Amount: 1.18},
		{ID: 20, SeqNo: 2, Amount: 1.17},
	}

	item, swap, ok := pickLuckyItemForTarget(items, lucky, 1, nil, true)

	if !ok {
		t.Fatalf("pickLuckyItemForTarget() ok = false, want true")
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
	if luckyAmountIsThunder(lucky, item.Amount, nil) {
		t.Fatalf("picked swapped item is thunder, want win")
	}
}

func TestPickLuckyBotThunderModeTargetRandomizesWithinDailyThunderQuota(t *testing.T) {
	lucky := pojo.LuckyMoney{GameMode: 0, Thunder: 8}
	items := []pojo.LuckyMoneyItem{
		{ID: 10, SeqNo: 1, Amount: 1.18},
		{ID: 20, SeqNo: 2, Amount: 1.17},
	}

	targetWin, canGrab := pickLuckyBotThunderModeTarget(items, lucky, 4, 4, 0, 0.49)
	if !canGrab {
		t.Fatalf("canGrab = false, want true")
	}
	if targetWin == nil || *targetWin {
		t.Fatalf("targetWin = %v, want false when random falls inside thunder quota", targetWin)
	}

	targetWin, canGrab = pickLuckyBotThunderModeTarget(items, lucky, 4, 4, 0, 0.75)
	if !canGrab {
		t.Fatalf("canGrab = false, want true")
	}
	if targetWin == nil || !*targetWin {
		t.Fatalf("targetWin = %v, want true when random falls outside thunder quota", targetWin)
	}
}

func TestPickLuckyBotThunderModeTargetUsesFloorHalfDailyThunderQuota(t *testing.T) {
	lucky := pojo.LuckyMoney{GameMode: 0, Thunder: 8}
	items := []pojo.LuckyMoneyItem{
		{ID: 10, SeqNo: 1, Amount: 1.18},
		{ID: 20, SeqNo: 2, Amount: 1.17},
	}

	targetWin, canGrab := pickLuckyBotThunderModeTarget(items, lucky, 3, 2, 1, 0)
	if !canGrab {
		t.Fatalf("canGrab = false, want true")
	}
	if targetWin == nil || !*targetWin {
		t.Fatalf("targetWin = %v, want true after robot already grabbed floor(3/2) thunder packet", targetWin)
	}
}

func TestPickLuckyBotThunderModeTargetStopsWhenQuotaReachedAndOnlyThunderLeft(t *testing.T) {
	lucky := pojo.LuckyMoney{GameMode: 0, Thunder: 8}
	items := []pojo.LuckyMoneyItem{
		{ID: 20, SeqNo: 2, Amount: 1.28},
	}

	targetWin, canGrab := pickLuckyBotThunderModeTarget(items, lucky, 2, 1, 1, 0)
	if canGrab {
		t.Fatalf("canGrab = true, want false when robot thunder quota is reached and only thunder remains")
	}
	if targetWin != nil {
		t.Fatalf("targetWin = %v, want nil when bot should stop", targetWin)
	}
}

func TestPickLuckyBotThunderModeTargetTakesThunderWhenOnlyThunderLeftAndQuotaRemains(t *testing.T) {
	lucky := pojo.LuckyMoney{GameMode: 0, Thunder: 8}
	items := []pojo.LuckyMoneyItem{
		{ID: 10, SeqNo: 1, Amount: 1.18},
	}

	targetWin, canGrab := pickLuckyBotThunderModeTarget(items, lucky, 2, 2, 0, 0.99)
	if !canGrab {
		t.Fatalf("canGrab = false, want true while robot thunder quota remains")
	}
	if targetWin == nil || *targetWin {
		t.Fatalf("targetWin = %v, want false when only thunder remains and quota remains", targetWin)
	}
}

func TestPickLuckyBotThunderModeTargetWinsWhenCurrentPacketHasNoThunder(t *testing.T) {
	lucky := pojo.LuckyMoney{GameMode: 0, Thunder: 8}
	items := []pojo.LuckyMoneyItem{
		{ID: 10, SeqNo: 1, Amount: 1.17},
		{ID: 20, SeqNo: 2, Amount: 1.19},
	}

	targetWin, canGrab := pickLuckyBotThunderModeTarget(items, lucky, 6, 4, 0, 0)
	if !canGrab {
		t.Fatalf("canGrab = false, want true when current packet has non-thunder amounts")
	}
	if targetWin == nil || !*targetWin {
		t.Fatalf("targetWin = %v, want true when current packet has no thunder amount", targetWin)
	}
}

func TestLuckyBotAutoGrabTargetModeForBotSenderUsesNoTarget(t *testing.T) {
	tests := []pojo.LuckyMoney{
		{GameMode: 0, Thunder: 8},
		{GameMode: 1, Thunder: 8},
	}

	for _, lucky := range tests {
		if got := luckyBotAutoGrabTargetModeFor(lucky, true); got != luckyBotAutoGrabNoTarget {
			t.Fatalf("luckyBotAutoGrabTargetModeFor(%+v, true) = %d, want no target", lucky, got)
		}
	}
}

func TestLuckyBotAutoGrabTargetModeForRealUserSender(t *testing.T) {
	if got := luckyBotAutoGrabTargetModeFor(pojo.LuckyMoney{GameMode: 0, Thunder: 8}, false); got != luckyBotAutoGrabDailyThunderTarget {
		t.Fatalf("thunder mode target = %d, want daily thunder target", got)
	}
	if got := luckyBotAutoGrabTargetModeFor(pojo.LuckyMoney{GameMode: 1, Thunder: 8}, false); got != luckyBotAutoGrabWinRateTarget {
		t.Fatalf("odd/even mode target = %d, want win-rate target", got)
	}
}

func TestNormalizeLuckyBotWinRate(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want float64
	}{
		{name: "empty uses default", raw: "", want: luckyBotWinRateDefault},
		{name: "invalid uses default", raw: "abc", want: luckyBotWinRateDefault},
		{name: "negative uses default", raw: "-0.1", want: luckyBotWinRateDefault},
		{name: "above one uses default", raw: "1.1", want: luckyBotWinRateDefault},
		{name: "zero is valid", raw: "0", want: 0},
		{name: "one is valid", raw: "1", want: 1},
		{name: "custom valid", raw: "0.65", want: 0.65},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizeLuckyBotWinRate(tt.raw); got != tt.want {
				t.Fatalf("normalizeLuckyBotWinRate(%q) = %.2f, want %.2f", tt.raw, got, tt.want)
			}
		})
	}
}
