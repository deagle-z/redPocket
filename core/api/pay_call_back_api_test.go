package api

import (
	"BaseGoUni/core/base"
	"BaseGoUni/core/pay/gctpk"
	"BaseGoUni/core/pay/vcpaymxn"
	"BaseGoUni/core/utils"
	"encoding/json"
	"testing"
)

func TestBuildGctpkPayoutNotifyParams(t *testing.T) {
	req := gctpkPayoutNotifyReq{
		MerNo:       "861129000037519981",
		MerOrderNo:  "ST23122601326948192743502",
		OrderNo:     "231226000000000481959263",
		OrderAmount: "200000.00",
		PayTime:     "2023-12-26 16:28:47",
		Status:      7,
		ResultCode:  "",
		ResultMsg:   "",
	}

	params := buildGctpkPayoutNotifyParams(req)
	expected := map[string]string{
		"merNo":       req.MerNo,
		"merOrderNo":  req.MerOrderNo,
		"orderNo":     req.OrderNo,
		"orderAmount": req.OrderAmount,
		"payTime":     req.PayTime,
		"status":      "7",
		"resultCode":  "",
		"resultMsg":   "",
	}
	for key, want := range expected {
		if got := params[key]; got != want {
			t.Fatalf("%s = %q, want %q", key, got, want)
		}
	}
	if _, ok := params["payAmount"]; ok {
		t.Fatal("payout notify params must not include payAmount")
	}
}

func TestValidateGctpkPayoutConfigAcceptsSignedNotify(t *testing.T) {
	req := gctpkPayoutNotifyReq{
		MerNo:       "861129000037519981",
		MerOrderNo:  "ST23122601326948192743502",
		OrderNo:     "231226000000000481959263",
		OrderAmount: "200000.00",
		PayTime:     "2023-12-26 16:28:47",
		Status:      7,
		ResultCode:  "",
		ResultMsg:   "",
	}
	cfg := base.GctpkPayConfig{
		MerNo:  req.MerNo,
		Secret: "test-secret",
	}
	req.Sign = gctpk.BuildSign(buildGctpkPayoutNotifyParams(req), cfg.Secret)

	if _, _, err := validateGctpkPayoutConfig(req, cfg, "GCTPKBRL"); err != nil {
		t.Fatalf("validateGctpkPayoutConfig() error = %v", err)
	}
}

func TestGctpkPayoutCallbackConfigDoesNotRequireMerchantOrSignMatch(t *testing.T) {
	req := gctpkPayoutNotifyReq{
		MerNo:       "callback-merchant",
		MerOrderNo:  "RC20260511143309rrELN2",
		OrderNo:     "2605113000000394280561965",
		OrderAmount: "100.00",
		PayTime:     "2026-05-12 12:32:18",
		Status:      7,
		Sign:        "",
	}
	cfg := base.GctpkPayConfig{
		MerNo:  "configured-merchant",
		Secret: "configured-secret",
	}

	if _, _, err := validateGctpkPayoutConfig(req, cfg, "GCTPKMXN"); err != nil {
		t.Fatalf("validateGctpkPayoutConfig() error = %v", err)
	}
}

func TestGctpkPayinCallbackConfigDoesNotRequireMerchantOrSignMatch(t *testing.T) {
	req := gctpkNotifyReq{
		PayAmount:  "100.00",
		OrderNo:    "2605113000000394280561965",
		MerNo:      "callback-merchant",
		MerOrderNo: "RC20260511143309rrELN2",
		PayTime:    "2026-05-12 12:32:18",
		BusiCode:   "107001",
		Status:     5,
		Sign:       "",
	}
	cfg := base.GctpkPayConfig{
		MerNo:  "configured-merchant",
		Secret: "configured-secret",
	}

	if _, _, err := validateGctpkPayConfig(req, cfg, "GCTPKMXN"); err != nil {
		t.Fatalf("validateGctpkPayConfig() error = %v", err)
	}
}

func TestValidateVcpayMxnNotifyConfigAcceptsCallbackSignWithoutCodeMsg(t *testing.T) {
	oldConfig := utils.GlobalConfig
	defer func() {
		utils.GlobalConfig = oldConfig
	}()

	utils.GlobalConfig.Pay.Vcpaymxn = base.VcpayMxnPayConfig{
		AppID:  "app123",
		AppKey: "secret",
	}
	req := vcpayMxnNotifyReq{
		Code:        "200",
		Msg:         "SUCCESS",
		AppID:       "app123",
		NonceStr:    "nonce123",
		NotifyURL:   "https://merchant.test/wd/notify",
		OrderAmount: 10000,
		OrderFee:    0,
		OutTradeNo:  "WD123456",
		TradeNo:     "8888888888888888",
		TradeState:  1,
		TradeType:   "MX0001",
	}
	req.Sign = vcpaymxn.BuildCallbackSignAny(buildVcpayMxnNotifyParams(req), "secret")

	if _, err := validateVcpayMxnNotifyConfig(req); err != nil {
		t.Fatalf("validateVcpayMxnNotifyConfig() error = %v", err)
	}
}

func TestVcpayMxnNotifyAmountConvertsCentsToPayAmount(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want float64
	}{
		{name: "numeric cents", raw: `{"order_amount":10000}`, want: 100},
		{name: "string cents", raw: `{"order_amount":"12345"}`, want: 123.45},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req vcpayMxnNotifyReq
			if err := json.Unmarshal([]byte(tt.raw), &req); err != nil {
				t.Fatalf("json.Unmarshal() error = %v", err)
			}
			if got := centsToPayAmount(req.OrderAmount.Int64()); got != tt.want {
				t.Fatalf("centsToPayAmount(%d) = %.2f, want %.2f", req.OrderAmount.Int64(), got, tt.want)
			}
		})
	}
}
