package hopopay

import (
	"BaseGoUni/core/base"
	"BaseGoUni/core/pay"
	"encoding/json"
	"testing"
)

func TestBuildPayinParamsUsesFixedAndroidPlatform(t *testing.T) {
	cfg := base.HopopayPayConfig{
		MerchantID: "merchant123",
		NotifyURL:  "https://merchant.test/pay/notify",
		PageURL:    "https://merchant.test/return",
	}
	params, err := buildPayinParams(cfg, pay.PayRequest{
		UserID:    12345,
		OrderNo:   "RC123456",
		Amount:    20,
		PayMethod: "paypal",
		ClientIP:  "172.65.2.4",
		ReturnURL: "https://request.test/return",
		ExtraFields: map[string]string{
			"platform": "ios",
		},
	}, 1764744945, "376299859b5fb4ed")
	if err != nil {
		t.Fatalf("buildPayinParams() error = %v", err)
	}
	if params["app_user_id"] != "12345" {
		t.Fatalf("app_user_id = %v", params["app_user_id"])
	}
	if params["amount"] != json.Number("20.00") {
		t.Fatalf("amount = %v", params["amount"])
	}
	if params["callback_url"] != "https://merchant.test/pay/notify" {
		t.Fatalf("callback_url = %v", params["callback_url"])
	}
	if params["return_url"] != "https://request.test/return" {
		t.Fatalf("return_url = %v", params["return_url"])
	}
	if params["extra_params"] != `{"pay_type":"paypal","platform":"android"}` {
		t.Fatalf("extra_params = %v", params["extra_params"])
	}
}

func TestBuildPayinParamsRequiresPayType(t *testing.T) {
	_, err := buildPayinParams(base.HopopayPayConfig{
		MerchantID: "merchant123",
		NotifyURL:  "https://merchant.test/pay/notify",
		PageURL:    "https://merchant.test/return",
	}, pay.PayRequest{
		UserID:   12345,
		OrderNo:  "RC123456",
		Amount:   20,
		ClientIP: "172.65.2.4",
	}, 1764744945, "376299859b5fb4ed")
	if err == nil || err.Error() != "HOPOPAY 代收参数缺少 pay_type" {
		t.Fatalf("error = %v", err)
	}
}

func TestBuildPayoutParamsUsesFixedAndroidPlatform(t *testing.T) {
	cfg := base.HopopayPayConfig{
		MerchantID:      "merchant123",
		PayoutNotifyURL: "https://merchant.test/wd/notify",
	}
	params, err := buildPayoutParams(cfg, pay.PayoutRequest{
		UserID:   12345,
		OrderNo:  "WD123456",
		Amount:   15,
		ClientIP: "172.65.2.4",
		AccName:  "TEST USER",
		AccNo:    "user@example.com",
		BankCode: "paypal",
		ExtraFields: map[string]string{
			"platform": "ios",
		},
	}, 1764744945, "376299859b5fb4ed")
	if err != nil {
		t.Fatalf("buildPayoutParams() error = %v", err)
	}
	if params["bank_code"] != "paypal" {
		t.Fatalf("bank_code = %v", params["bank_code"])
	}
	if params["extra_params"] != `{"platform":"android"}` {
		t.Fatalf("extra_params = %v", params["extra_params"])
	}
}

func TestRawJSONSignaturePreservesNumberFormatAndNull(t *testing.T) {
	rawBody := []byte(`{"status":3,"amount":10,"net_amount":9.62,"pay_at":null,"app_order_id":"RC123","signature":""}`)
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(rawBody, &raw); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	raw["signature"], _ = json.Marshal(BuildSignatureFromRawJSON(raw, "secret123"))
	if !VerifyRawJSONSignature(raw, "secret123") {
		t.Fatalf("VerifyRawJSONSignature() = false")
	}
}

func TestNormalizePayoutQueryStatus(t *testing.T) {
	tests := []struct {
		status int
		want   int
	}{
		{status: 3, want: 7},
		{status: 0, want: 0},
		{status: 1, want: 0},
		{status: 4, want: 0},
		{status: 2, want: 8},
		{status: 5, want: 8},
		{status: 6, want: 8},
		{status: 9, want: 8},
	}
	for _, tt := range tests {
		if got := normalizePayoutQueryStatus(tt.status); got != tt.want {
			t.Fatalf("normalizePayoutQueryStatus(%d) = %d, want %d", tt.status, got, tt.want)
		}
	}
}

func TestProviderRegistered(t *testing.T) {
	if pay.Get(providerName) == nil {
		t.Fatalf("provider %s is not registered", providerName)
	}
	if pay.GetPayout(providerName) == nil {
		t.Fatalf("payout provider %s is not registered", providerName)
	}
	if pay.GetPayoutQuery(providerName) == nil {
		t.Fatalf("payout query provider %s is not registered", providerName)
	}
}
