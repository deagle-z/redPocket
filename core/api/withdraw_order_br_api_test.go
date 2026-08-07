package api

import (
	"BaseGoUni/core/pojo"
	"testing"
)

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

func TestResolveAppWithdrawV2ChannelUsesExistingVcpayProviderForPeru(t *testing.T) {
	if got := resolveAppWithdrawV2Channel(" pe "); got != "VCPAYPEN" {
		t.Fatalf("resolveAppWithdrawV2Channel() = %q, want VCPAYPEN", got)
	}
	if got := resolveAppWithdrawV2Provider("PE"); got != "VCPAYPEN" {
		t.Fatalf("resolveAppWithdrawV2Provider() = %q, want VCPAYPEN", got)
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

func TestApplyWithdrawReceiverSnapshotDoesNotStoreProviderIdentityTypeAsDocumentType(t *testing.T) {
	var req pojo.WithdrawOrderBrSet

	applyWithdrawReceiverSnapshot(&req, "PE", nil, map[string]string{
		"accNameMxnW":      "Ernesto Morales",
		"accNoMXNW":        "137835103778065956",
		"bankCodeMXNW":     "MXNBANCOPPEL",
		"identityTypeMXNW": "BANK_ACCOUNT",
	})

	if req.ReceiverDocumentType != nil {
		t.Fatalf("ReceiverDocumentType = %q, want nil", *req.ReceiverDocumentType)
	}
	if req.PixKeyType == nil || *req.PixKeyType != "BANK_ACCOUNT" {
		t.Fatalf("PixKeyType = %v, want BANK_ACCOUNT", req.PixKeyType)
	}
}
