package vcpaypen

import (
	"BaseGoUni/core/base"
	"BaseGoUni/core/pay"
	"encoding/json"
	"strings"
	"testing"
)

func TestBuildPayinParams(t *testing.T) {
	cfg := base.VcpayPenPayConfig{
		AppID:     "app123",
		NotifyURL: "https://merchant.test/pay/notify",
		PageURL:   "https://merchant.test/return",
	}
	params, err := buildPayinParams(cfg, pay.PayRequest{
		OrderNo:   "ORDER123456",
		Amount:    10.25,
		NotifyURL: "https://request.test/pay/notify",
		ReturnURL: "https://request.test/return",
		ExtraFields: map[string]string{
			"document": "66666666",
			"name":     "TEST USER",
		},
	}, "nonce123")
	if err != nil {
		t.Fatalf("buildPayinParams() error = %v", err)
	}

	if params["app_id"] != "app123" {
		t.Fatalf("app_id = %v", params["app_id"])
	}
	if params["nonce_str"] != "nonce123" {
		t.Fatalf("nonce_str = %v", params["nonce_str"])
	}
	if params["trade_type"] != defaultPayinTradeType {
		t.Fatalf("trade_type = %v, want %v", params["trade_type"], defaultPayinTradeType)
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
	// 8 位纯数字推断为 DNI
	if params["identity_type"] != "DNI" {
		t.Fatalf("identity_type = %v, want DNI", params["identity_type"])
	}
	if params["identity"] != "66666666" {
		t.Fatalf("identity = %v", params["identity"])
	}
	if params["identity_name"] != "TEST USER" {
		t.Fatalf("identity_name = %v", params["identity_name"])
	}
}

func TestBuildPayinParamsTradeTypeWhitelist(t *testing.T) {
	cfg := base.VcpayPenPayConfig{AppID: "app123", PageURL: "https://merchant.test/return"}
	newReq := func(payMethod string) pay.PayRequest {
		return pay.PayRequest{
			OrderNo:     "ORDER123456",
			Amount:      1,
			PayMethod:   payMethod,
			ExtraFields: map[string]string{"document": "66666666"},
		}
	}

	params, err := buildPayinParams(cfg, newReq("pew002"), "nonce123")
	if err != nil {
		t.Fatalf("buildPayinParams() error = %v", err)
	}
	if params["trade_type"] != "PEW002" {
		t.Fatalf("trade_type = %v, want PEW002", params["trade_type"])
	}

	// 非白名单值（如通用的 pix）必须回落到默认值，否则三方返回 40104
	params, err = buildPayinParams(cfg, newReq("pix"), "nonce123")
	if err != nil {
		t.Fatalf("buildPayinParams() error = %v", err)
	}
	if params["trade_type"] != defaultPayinTradeType {
		t.Fatalf("trade_type = %v, want %v", params["trade_type"], defaultPayinTradeType)
	}
}

func TestBuildPayinParamsRequiresBackURLAndIdentity(t *testing.T) {
	req := pay.PayRequest{
		OrderNo:     "ORDER123456",
		Amount:      1,
		ExtraFields: map[string]string{"document": "66666666"},
	}
	if _, err := buildPayinParams(base.VcpayPenPayConfig{AppID: "app123"}, req, "nonce123"); err == nil {
		t.Fatal("buildPayinParams() error = nil, want back_url error")
	}

	cfg := base.VcpayPenPayConfig{AppID: "app123", PageURL: "https://merchant.test/return"}
	req.ExtraFields = nil
	if _, err := buildPayinParams(cfg, req, "nonce123"); err == nil {
		t.Fatal("buildPayinParams() error = nil, want identity error")
	}
}

func TestBuildPayoutParamsBankTransfer(t *testing.T) {
	cfg := base.VcpayPenPayConfig{
		AppID:           "app123",
		NotifyURL:       "https://merchant.test/pay/notify",
		PayoutNotifyURL: "https://merchant.test/wd/notify",
	}
	params, err := buildPayoutParams(cfg, pay.PayoutRequest{
		OrderNo:      "WD123456",
		Amount:       99.99,
		AccName:      "BENEFICIARY",
		AccNo:        "00212345678901234567", // 20 位 CCI
		BankCode:     "BCP",
		AccountType:  "CORRIENTE",
		IdentityType: "DNI",
		IdentityNo:   "66666666",
		NotifyURL:    "https://request.test/wd/notify",
	}, "nonce123")
	if err != nil {
		t.Fatalf("buildPayoutParams() error = %v", err)
	}

	if params["trade_type"] != defaultPayoutTradeType {
		t.Fatalf("trade_type = %v, want %v", params["trade_type"], defaultPayoutTradeType)
	}
	if params["order_amount"] != int64(9999) {
		t.Fatalf("order_amount = %v", params["order_amount"])
	}
	if params["notify_url"] != "https://merchant.test/wd/notify" {
		t.Fatalf("notify_url = %v", params["notify_url"])
	}
	if params["bank_code"] != "PEBB04" {
		t.Fatalf("bank_code = %v, want PEBB04", params["bank_code"])
	}
	if params["bank_owner"] != "BENEFICIARY" {
		t.Fatalf("bank_owner = %v", params["bank_owner"])
	}
	if params["bank_account"] != "00212345678901234567" {
		t.Fatalf("bank_account = %v", params["bank_account"])
	}
	if params["bank_account_type"] != "CA" {
		t.Fatalf("bank_account_type = %v, want CA", params["bank_account_type"])
	}
	if params["identity_type"] != "DNI" {
		t.Fatalf("identity_type = %v", params["identity_type"])
	}
	if params["identity"] != "66666666" {
		t.Fatalf("identity = %v", params["identity"])
	}
}

func TestBuildPayoutParamsWallet(t *testing.T) {
	cfg := base.VcpayPenPayConfig{AppID: "app123", PayoutNotifyURL: "https://merchant.test/wd/notify"}
	params, err := buildPayoutParams(cfg, pay.PayoutRequest{
		OrderNo:  "WD123456",
		Amount:   50,
		AccName:  "BENEFICIARY",
		AccNo:    "906666666", // 9 开头 9 位手机号
		BankCode: "YAPE",
		ExtraFields: map[string]string{
			"identity":      "10412345678",
			"identity_type": "TAX",
		},
	}, "nonce123")
	if err != nil {
		t.Fatalf("buildPayoutParams() error = %v", err)
	}

	if params["bank_code"] != "PEW001" {
		t.Fatalf("bank_code = %v, want PEW001", params["bank_code"])
	}
	// 电子钱包 trade_type 与 bank_code 同值
	if params["trade_type"] != "PEW001" {
		t.Fatalf("trade_type = %v, want PEW001", params["trade_type"])
	}
	if params["bank_account_type"] != defaultAccountType {
		t.Fatalf("bank_account_type = %v, want %v", params["bank_account_type"], defaultAccountType)
	}
	if params["identity_type"] != "RUC" {
		t.Fatalf("identity_type = %v, want RUC", params["identity_type"])
	}
	if params["identity"] != "10412345678" {
		t.Fatalf("identity = %v", params["identity"])
	}
}

func TestBuildPayoutParamsValidatesAccountAndIdentity(t *testing.T) {
	cfg := base.VcpayPenPayConfig{AppID: "app123", PayoutNotifyURL: "https://merchant.test/wd/notify"}
	baseReq := pay.PayoutRequest{
		OrderNo:    "WD123456",
		Amount:     50,
		AccName:    "BENEFICIARY",
		AccNo:      "00212345678901234567",
		BankCode:   "PEBB04",
		IdentityNo: "66666666",
	}

	// 钱包渠道用银行账号
	wallet := baseReq
	wallet.BankCode = "PEW001"
	if _, err := buildPayoutParams(cfg, wallet, "nonce123"); err == nil ||
		!strings.Contains(err.Error(), "电子钱包") {
		t.Fatalf("wallet account error = %v, want wallet format error", err)
	}

	// 银行渠道用手机号
	bank := baseReq
	bank.AccNo = "906666666"
	if _, err := buildPayoutParams(cfg, bank, "nonce123"); err == nil ||
		!strings.Contains(err.Error(), "CCI") {
		t.Fatalf("bank account error = %v, want CCI format error", err)
	}

	// 缺证件号
	noIdentity := baseReq
	noIdentity.IdentityNo = ""
	if _, err := buildPayoutParams(cfg, noIdentity, "nonce123"); err == nil ||
		!strings.Contains(err.Error(), "identity") {
		t.Fatalf("identity error = %v, want identity error", err)
	}
}

func TestResolvePayoutBankCodeMapsFrontendValues(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{in: "YAPE", want: "PEW001"},
		{in: "PLIN", want: "PEW002"},
		{in: "BCP", want: "PEBB04"},
		{in: "PENBCP", want: "PEBB04"},
		{in: "Banco de Credito del Peru", want: "PEBB04"},
		{in: "INTERBANK", want: "PEBI01"},
		{in: "BBVA_PERU", want: "PEBB03"},
		{in: "scotiabank peru", want: "PEBS02"},
		{in: "Caja Trujillo", want: "PEBC10"},
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
	if got := resolvePayoutBankCode(" PEBA01 "); got != "PEBA01" {
		t.Fatalf("provider code = %q, want PEBA01", got)
	}
	if got := resolvePayoutBankCode("BANXICO"); got != "BANXICO" {
		t.Fatalf("unknown code = %q, want BANXICO", got)
	}
}

func TestNormalizeIdentityType(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{in: "dni", want: "DNI"},
		{in: "Documento", want: "DNI"},
		{in: "passport", want: "PAS"},
		{in: "carnet de extranjeria", want: "CE"},
		{in: "tax_id", want: "RUC"},
		{in: "ruc", want: "RUC"},
		{in: "", want: ""},
		{in: "CLABE", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			if got := normalizeIdentityType(tt.in); got != tt.want {
				t.Fatalf("normalizeIdentityType(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
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
