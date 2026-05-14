package api

import "testing"

func TestCalculateLotteryAvailableCountIncludesFreeCount(t *testing.T) {
	available := calculateLotteryAvailableCount(3, 2, 4)
	if available != 5 {
		t.Fatalf("calculateLotteryAvailableCount() = %d, want %d", available, 5)
	}
}

func TestSelectLotteryConsumptionPrefersFreeCount(t *testing.T) {
	consumption, ok := selectLotteryConsumption(1, 0, 0, 1000)
	if !ok {
		t.Fatalf("selectLotteryConsumption() ok = false, want true")
	}
	if !consumption.UseFree {
		t.Fatalf("selectLotteryConsumption() UseFree = false, want true")
	}
	if consumption.PeerAmount != 0 {
		t.Fatalf("selectLotteryConsumption() PeerAmount = %.2f, want 0", consumption.PeerAmount)
	}
}

func TestSelectLotteryConsumptionUsesFlowAfterFreeCount(t *testing.T) {
	consumption, ok := selectLotteryConsumption(0, 3, 2, 1000)
	if !ok {
		t.Fatalf("selectLotteryConsumption() ok = false, want true")
	}
	if consumption.UseFree {
		t.Fatalf("selectLotteryConsumption() UseFree = true, want false")
	}
	if consumption.PeerAmount != 1000 {
		t.Fatalf("selectLotteryConsumption() PeerAmount = %.2f, want 1000", consumption.PeerAmount)
	}
}

func TestSelectLotteryConsumptionRejectsWhenNoChance(t *testing.T) {
	_, ok := selectLotteryConsumption(0, 2, 2, 1000)
	if ok {
		t.Fatalf("selectLotteryConsumption() ok = true, want false")
	}
}
