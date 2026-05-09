package gctpk

import (
	"BaseGoUni/core/base"
	"BaseGoUni/core/pay"
	"testing"
)

func TestBuildPayoutParamsMatchesMxnReadme(t *testing.T) {
	req := pay.PayoutRequest{
		OrderNo:      "WD202605090001",
		Amount:       200000,
		Currency:     "BRL",
		AccName:      "Tom",
		AccNo:        "123456789012345678",
		BankCode:     "002",
		IdentityType: "clabe",
		BusiCode:     "",
		Email:        "tom@example.com",
		Phone:        "5511999999999",
	}
	params := buildPayoutParams(
		base.GctpkPayConfig{MerNo: "merchant"},
		req,
		"https://example.com/payoutNotify",
		"1760000000000",
		resolvePayoutIdentityType(req),
	)

	if got := params["identityType"]; got != "clabe" {
		t.Fatalf("identityType = %q, want clabe", got)
	}
	if got := params["currency"]; got != "MXN" {
		t.Fatalf("currency = %q, want MXN", got)
	}
	if got := params["busiCode"]; got != "207001" {
		t.Fatalf("busiCode = %q, want 207001", got)
	}
	if got := params["notifyUrl"]; got != "https://example.com/payoutNotify" {
		t.Fatalf("notifyUrl = %q, want config payout notify URL", got)
	}
}

func TestResolvePayinNameUsesMxnExtraField(t *testing.T) {
	req := pay.PayRequest{
		ExtraFields: map[string]string{
			"nameMXN": " Alice ",
		},
	}

	if got := resolvePayinName(req); got != "Alice" {
		t.Fatalf("resolvePayinName = %q, want Alice", got)
	}
}
