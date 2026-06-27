package repository

import "testing"

func TestRechargeCallbackCreditBaseAmountPrefersProviderPayAmount(t *testing.T) {
	got := rechargeCallbackCreditBaseAmount(10000, 100)
	if got != 100 {
		t.Fatalf("rechargeCallbackCreditBaseAmount() = %.2f, want 100.00", got)
	}
}

func TestRechargeCallbackCreditBaseAmountFallsBackToOrderAmount(t *testing.T) {
	got := rechargeCallbackCreditBaseAmount(100, 0)
	if got != 100 {
		t.Fatalf("rechargeCallbackCreditBaseAmount() = %.2f, want 100.00", got)
	}
}
