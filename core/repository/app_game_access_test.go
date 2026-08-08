package repository

import "testing"

func TestParseRechargeGameUnlockEnabled(t *testing.T) {
	tests := []struct {
		raw     string
		enabled bool
		valid   bool
	}{
		{raw: "1", enabled: true, valid: true},
		{raw: " 0 ", enabled: false, valid: true},
		{raw: "true", enabled: false, valid: false},
		{raw: "", enabled: false, valid: false},
	}

	for _, test := range tests {
		enabled, valid := parseRechargeGameUnlockEnabled(test.raw)
		if enabled != test.enabled || valid != test.valid {
			t.Fatalf("parseRechargeGameUnlockEnabled(%q) = %t,%t; want %t,%t", test.raw, enabled, valid, test.enabled, test.valid)
		}
	}
}

func TestGetRechargeGameUnlockEnabledDefaultsToEnabled(t *testing.T) {
	if !GetRechargeGameUnlockEnabled(nil) {
		t.Fatalf("GetRechargeGameUnlockEnabled(nil) = false, want true")
	}
}
