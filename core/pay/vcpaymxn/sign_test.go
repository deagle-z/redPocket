package vcpaymxn

import "testing"

func TestBuildSignSortsParamsAndSkipsEmptyAndSign(t *testing.T) {
	params := map[string]string{
		"out_trade_no": "22222222222222222",
		"sign":         "SHOULD_NOT_JOIN",
		"empty":        "",
		"nonce_str":    "abc",
		"app_id":       "1111111111111111",
	}

	raw := buildSignString(params, "secret", nil)
	if raw != "app_id=1111111111111111&nonce_str=abc&out_trade_no=22222222222222222&key=secret" {
		t.Fatalf("raw sign string = %q", raw)
	}

	if got := BuildSign(params, "secret"); got != "79CF600DDEF012F8BEC3C01E55657E0C" {
		t.Fatalf("BuildSign() = %s", got)
	}
}

func TestBuildCallbackSignExcludesCodeAndMsg(t *testing.T) {
	params := map[string]string{
		"code":         "200",
		"msg":          "SUCCESS",
		"app_id":       "1111111111111111",
		"nonce_str":    "abc",
		"out_trade_no": "22222222222222222",
		"trade_state":  "1",
	}

	raw := buildSignString(params, "secret", callbackSignExcludedKeys)
	if raw != "app_id=1111111111111111&nonce_str=abc&out_trade_no=22222222222222222&trade_state=1&key=secret" {
		t.Fatalf("raw callback sign string = %q", raw)
	}

	if got := BuildCallbackSign(params, "secret"); got != "271F89E19674E6F039C6AECCB002837C" {
		t.Fatalf("BuildCallbackSign() = %s", got)
	}
}

func TestVerifySignAny(t *testing.T) {
	params := map[string]any{
		"K2":   "V2",
		"K1":   "V1",
		"sign": "c45d057cc05ac317b0fa997ad6e2d21a",
	}

	if !VerifySignAny(params, "value") {
		t.Fatal("VerifySignAny() = false")
	}
}
