package services

import "testing"

func TestBuildLuckyNumsRewardMatchesLastThreeDigitsAfterRemovingDot(t *testing.T) {
	tests := []struct {
		name       string
		amount     float64
		config     string
		giftAmount float64
		wantHit    bool
		wantNum    string
	}{
		{name: "one digit before dot", amount: 1.11, config: "111|222|333", giftAmount: 1000, wantHit: true, wantNum: "111"},
		{name: "two digits before dot", amount: 12.22, config: "111|222|333", giftAmount: 1000, wantHit: true, wantNum: "222"},
		{name: "zero padded cents", amount: 30.03, config: "003", giftAmount: 1000, wantHit: true, wantNum: "003"},
		{name: "not matched", amount: 4.56, config: "111|222|333", giftAmount: 1000, wantHit: false},
		{name: "disabled by empty config", amount: 1.11, config: "", giftAmount: 1000, wantHit: false},
		{name: "disabled by zero gift", amount: 1.11, config: "111", giftAmount: 0, wantHit: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildLuckyNumsReward(tt.amount, tt.config, tt.giftAmount)
			if got.Triggered != tt.wantHit {
				t.Fatalf("Triggered = %v, want %v", got.Triggered, tt.wantHit)
			}
			if got.Num != tt.wantNum {
				t.Fatalf("Num = %q, want %q", got.Num, tt.wantNum)
			}
			if tt.wantHit && got.Amount != tt.giftAmount {
				t.Fatalf("Amount = %.2f, want %.2f", got.Amount, tt.giftAmount)
			}
		})
	}
}
