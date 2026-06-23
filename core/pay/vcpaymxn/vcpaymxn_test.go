package vcpaymxn

import (
	"BaseGoUni/core/base"
	"BaseGoUni/core/pay"
	"encoding/json"
	"regexp"
	"testing"
)

func TestBuildPayinParams(t *testing.T) {
	cfg := base.VcpayMxnPayConfig{
		AppID:     "app123",
		NotifyURL: "https://merchant.test/pay/notify",
		PageURL:   "https://merchant.test/return",
	}
	params := buildPayinParams(cfg, pay.PayRequest{
		OrderNo:   "ORDER123456",
		Amount:    10.25,
		NotifyURL: "https://request.test/pay/notify",
		ReturnURL: "https://request.test/return",
		ExtraFields: map[string]string{
			"phoneMXN": "5512345678",
			"name":     "TEST USER",
		},
	}, "nonce123")

	if params["app_id"] != "app123" {
		t.Fatalf("app_id = %v", params["app_id"])
	}
	if params["nonce_str"] != "nonce123" {
		t.Fatalf("nonce_str = %v", params["nonce_str"])
	}
	if params["trade_type"] != defaultPayinTradeType {
		t.Fatalf("trade_type = %v", params["trade_type"])
	}
	if params["order_amount"] != int64(1025) {
		t.Fatalf("order_amount = %v", params["order_amount"])
	}
	if params["notify_url"] != "https://merchant.test/pay/notify" {
		t.Fatalf("notify_url = %v", params["notify_url"])
	}
	if params["back_url"] != "https://request.test/return" {
		t.Fatalf("back_url = %v", params["back_url"])
	}
	if params["identity_type"] != defaultIdentityType {
		t.Fatalf("identity_type = %v", params["identity_type"])
	}
	if params["identity"] != "5512345678" {
		t.Fatalf("identity = %v", params["identity"])
	}
	if params["identity_name"] != "TEST USER" {
		t.Fatalf("identity_name = %v", params["identity_name"])
	}
}

func TestBuildPayoutParams(t *testing.T) {
	cfg := base.VcpayMxnPayConfig{
		AppID:           "app123",
		NotifyURL:       "https://merchant.test/pay/notify",
		PayoutNotifyURL: "https://merchant.test/wd/notify",
	}
	params := buildPayoutParams(cfg, pay.PayoutRequest{
		OrderNo:      "WD123456",
		Amount:       99.99,
		AccName:      "BENEFICIARY",
		AccNo:        "012180001234567891",
		BankCode:     "MXNHSBC",
		IdentityType: "CLABE",
		IdentityNo:   "NOT_A_PHONE",
		Phone:        "5512345678",
		NotifyURL:    "https://request.test/wd/notify",
		ExtraFields: map[string]string{
			"identity_type": "DOCUMENT",
			"identity":      "EXTRA_IDENTITY",
			"phone":         "5599999999",
		},
	}, "nonce123")

	if params["trade_type"] != defaultPayoutTradeType {
		t.Fatalf("trade_type = %v", params["trade_type"])
	}
	if params["order_amount"] != int64(9999) {
		t.Fatalf("order_amount = %v", params["order_amount"])
	}
	if params["notify_url"] != "https://merchant.test/wd/notify" {
		t.Fatalf("notify_url = %v", params["notify_url"])
	}
	if params["bank_code"] != "MXH002" {
		t.Fatalf("bank_code = %v, want MXH002", params["bank_code"])
	}
	if params["bank_owner"] != "BENEFICIARY" {
		t.Fatalf("bank_owner = %v", params["bank_owner"])
	}
	if params["identity_type"] != defaultIdentityType {
		t.Fatalf("identity_type = %v", params["identity_type"])
	}
	identity, ok := params["identity"].(string)
	if !ok {
		t.Fatalf("identity type = %T, want string", params["identity"])
	}
	if identity == "5512345678" || identity == "5599999999" {
		t.Fatalf("identity = %q, should be randomly generated", identity)
	}
	if !regexp.MustCompile(`^(55|56|33|81|22|44|66|99)\d{8}$`).MatchString(identity) {
		t.Fatalf("identity = %q, want Mexico mobile phone format", identity)
	}
}

func TestResolvePayoutTradeTypeUsesIdentityTypeMxnW(t *testing.T) {
	tests := []struct {
		name string
		req  pay.PayoutRequest
		want string
	}{
		{
			name: "clabe extra field",
			req: pay.PayoutRequest{
				ExtraFields: map[string]string{"identityTypeMXNW": "CLABE"},
			},
			want: "MX0001",
		},
		{
			name: "bank extra field",
			req: pay.PayoutRequest{
				ExtraFields: map[string]string{"identityTypeMXNW": "BANK"},
			},
			want: "MX0000",
		},
		{
			name: "request identity type",
			req: pay.PayoutRequest{
				IdentityType: "CLABE",
			},
			want: "MX0001",
		},
		{
			name: "explicit trade type wins",
			req: pay.PayoutRequest{
				BusiCode:    "MX0000",
				ExtraFields: map[string]string{"identityTypeMXNW": "CLABE"},
			},
			want: "MX0000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := resolvePayoutTradeType(tt.req); got != tt.want {
				t.Fatalf("resolvePayoutTradeType() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestResolvePayoutBankCodeMapsFrontendValues(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{in: "MXNSTP", want: "MXS006"},
		{in: "MXNBANAMEX", want: "MXB003"},
		{in: "MXNBBVABANCOMER", want: "MXB019"},
		{in: "MXNBCT", want: "MXB006"},
		{in: "MXNIBA", want: "MXI002"},
		{in: "MXNBDB", want: "MXB002"},
		{in: "ABC_CAPITAL", want: "MXA001"},
		{in: "CREDIT_SUISSE", want: "MXC014"},
		{in: "SPIN_BY_OXXO", want: "MXS005"},
	}

	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			if got := resolvePayoutBankCode(tt.in); got != tt.want {
				t.Fatalf("resolvePayoutBankCode(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestResolvePayoutBankCodeKeepsProviderCodeAndUnknown(t *testing.T) {
	if got := resolvePayoutBankCode(" MXA001 "); got != "MXA001" {
		t.Fatalf("provider code = %q, want MXA001", got)
	}
	if got := resolvePayoutBankCode("BANXICO"); got != "BANXICO" {
		t.Fatalf("unknown code = %q, want BANXICO", got)
	}
}

func TestTradeRespParsesStringAmounts(t *testing.T) {
	var resp tradeResp
	if err := json.Unmarshal([]byte(`{
		"code":"200",
		"msg":"SUCCESS",
		"order_amount":"10000",
		"order_fee":"5",
		"out_trade_no":"WD123456",
		"trade_no":"8888888888888888",
		"trade_state":1
	}`), &resp); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if resp.Code.String() != "200" {
		t.Fatalf("code = %q", resp.Code.String())
	}
	if resp.OrderAmount.Int64() != 10000 {
		t.Fatalf("order_amount = %d", resp.OrderAmount.Int64())
	}
	if centsToAmount(resp.OrderAmount.Int64()) != 100 {
		t.Fatalf("amount = %v", centsToAmount(resp.OrderAmount.Int64()))
	}
	if got := normalizePayoutTradeState(resp.TradeState); got != 7 {
		t.Fatalf("normalized status = %d", got)
	}
}
