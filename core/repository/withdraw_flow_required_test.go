package repository

import "testing"

func TestCalculateWithdrawAmountFlowRequiredUsesWholeWithdrawAmount(t *testing.T) {
	got := calculateWithdrawAmountFlowRequired(98.41, 5)
	if got != 492.05 {
		t.Fatalf("calculateWithdrawAmountFlowRequired() = %.2f, want 492.05", got)
	}
}

func TestCalculateWithdrawAmountFlowRequiredIgnoresInvalidValues(t *testing.T) {
	if got := calculateWithdrawAmountFlowRequired(98.41, 0); got != 0 {
		t.Fatalf("calculateWithdrawAmountFlowRequired() = %.2f, want 0.00", got)
	}
}
