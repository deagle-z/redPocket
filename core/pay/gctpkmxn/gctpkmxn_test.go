package gctpk

import (
	"BaseGoUni/core/base"
	"BaseGoUni/core/pay"
	"encoding/json"
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
			"usernamemxn": " Alice ",
		},
	}

	if got := resolvePayinName(req); got != "Alice" {
		t.Fatalf("resolvePayinName = %q, want Alice", got)
	}
}

func TestResolveNotifyURLUsesConfigNotifyURL(t *testing.T) {
	cfg := base.GctpkPayConfig{NotifyURL: "https://config.example.com/mxn/notify"}

	if got := resolveNotifyURL(cfg); got != cfg.NotifyURL {
		t.Fatalf("resolveNotifyURL = %q, want config notify URL", got)
	}
}

func TestBuildPayoutQueryParamsIncludesDocumentFields(t *testing.T) {
	req := pay.PayoutQueryRequest{
		OrderNo:         "WD202606230001",
		ProviderOrderNo: "PAYOUT202606230001",
		RequestNo:       "REQ202606230001",
	}

	params := buildPayoutQueryParams(
		base.GctpkPayConfig{MerNo: "merchant"},
		req,
		"1760000000000",
		"REQ202606230001",
	)

	expected := map[string]string{
		"merNo":      "merchant",
		"requestNo":  "REQ202606230001",
		"merOrderNo": "WD202606230001",
		"orderNo":    "PAYOUT202606230001",
		"timestamp":  "1760000000000",
	}
	for key, want := range expected {
		if got := params[key]; got != want {
			t.Fatalf("params[%q] = %q, want %q", key, got, want)
		}
	}
}

func TestBuildPayoutQueryParamsOmitsEmptyProviderOrderNo(t *testing.T) {
	req := pay.PayoutQueryRequest{
		OrderNo:   "WD202606230001",
		RequestNo: "REQ202606230001",
	}

	params := buildPayoutQueryParams(
		base.GctpkPayConfig{MerNo: "merchant"},
		req,
		"1760000000000",
		"REQ202606230001",
	)

	if _, exists := params["orderNo"]; exists {
		t.Fatalf("params should omit empty orderNo, got %q", params["orderNo"])
	}
}

func TestPayoutQueryRespParsesAmountAndFlexibleSubCode(t *testing.T) {
	body := []byte(`{
		"code": 200,
		"msg": "success",
		"data": {
			"orderNo": "PAYOUT202606230001",
			"merOrderNo": "WD202606230001",
			"status": 7,
			"orderAmount": "123.45",
			"subMsg": "paid",
			"subCode": 1001,
			"payTime": "2026-06-23 12:30:00",
			"sign": "abc"
		}
	}`)

	var parsed payoutQueryResp
	if err := json.Unmarshal(body, &parsed); err != nil {
		t.Fatalf("json.Unmarshal payoutQueryResp error = %v", err)
	}
	if parsed.Data == nil {
		t.Fatal("parsed.Data is nil")
	}
	if got := parsed.Data.OrderAmount.Float64(); got != 123.45 {
		t.Fatalf("orderAmount = %v, want 123.45", got)
	}
	if got := parsed.Data.SubCode.String(); got != "1001" {
		t.Fatalf("subCode = %q, want 1001", got)
	}
}
