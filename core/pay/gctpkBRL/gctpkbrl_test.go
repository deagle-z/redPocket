package gctpk

import (
	"BaseGoUni/core/base"
	"BaseGoUni/core/pay"
	"testing"
)

func TestBuildPayoutParamsMatchesBrlReadme(t *testing.T) {
	req := pay.PayoutRequest{
		OrderNo:      "WD202605090002",
		Amount:       1000,
		Currency:     "MXN",
		AccName:      "Joao",
		AccNo:        "user@example.com",
		BankCode:     "002",
		IdentityType: "EMAIL",
		IdentityNo:   "12345678901",
		Email:        "joao@example.com",
		Phone:        "5511999999999",
	}
	params := buildPayoutParams(
		base.GctpkPayConfig{MerNo: "merchant"},
		req,
		"https://example.com/brl/payoutNotify",
		"1760000000000",
		resolvePayoutIdentityType(req),
	)

	if got := params["bankCode"]; got != "PIX" {
		t.Fatalf("bankCode = %q, want PIX", got)
	}
	if got := params["identityType"]; got != "EMAIL" {
		t.Fatalf("identityType = %q, want EMAIL", got)
	}
	if got := params["identityNo"]; got != "12345678901" {
		t.Fatalf("identityNo = %q, want 12345678901", got)
	}
	if got := params["currency"]; got != "BRL" {
		t.Fatalf("currency = %q, want BRL", got)
	}
	if got := params["busiCode"]; got != "206001" {
		t.Fatalf("busiCode = %q, want 206001", got)
	}
}
