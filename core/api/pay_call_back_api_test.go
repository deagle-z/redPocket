package api

import (
	"BaseGoUni/core/base"
	"BaseGoUni/core/pay/gctpk"
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
