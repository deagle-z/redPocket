package api

import "testing"

func TestCalculateAppWithdrawFeeFreeFirstThreeWithdrawals(t *testing.T) {
	tests := []struct {
		name       string
		todayCount int64
		amount     float64
		want       float64
	}{
		{name: "first withdrawal is free", todayCount: 0, amount: 100, want: 0},
		{name: "third withdrawal is free", todayCount: 2, amount: 100, want: 0},
		{name: "fourth withdrawal charges five percent", todayCount: 3, amount: 100, want: 5},
		{name: "fee is truncated to two decimals", todayCount: 3, amount: 10.19, want: 0.50},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := calculateAppWithdrawFee(tt.todayCount, tt.amount); got != tt.want {
				t.Fatalf("calculateAppWithdrawFee(%d, %.2f) = %.2f, want %.2f", tt.todayCount, tt.amount, got, tt.want)
			}
		})
	}
}

func TestResolveAppWithdrawV2ChannelUsesVcpayForMexico(t *testing.T) {
	if got := resolveAppWithdrawV2Channel(" mx "); got != "VCPAYMXN" {
		t.Fatalf("resolveAppWithdrawV2Channel() = %q, want VCPAYMXN", got)
	}
	if got := resolveAppWithdrawV2Provider("MX"); got != "VCPAYMXN" {
		t.Fatalf("resolveAppWithdrawV2Provider() = %q, want VCPAYMXN", got)
	}
}

func TestResolveAppWithdrawV2ChannelKeepsPixForOtherCountries(t *testing.T) {
	if got := resolveAppWithdrawV2Channel("BR"); got != "pix" {
		t.Fatalf("resolveAppWithdrawV2Channel() = %q, want pix", got)
	}
	if got := resolveAppWithdrawV2Provider("BR"); got != "" {
		t.Fatalf("resolveAppWithdrawV2Provider() = %q, want empty", got)
	}
}
