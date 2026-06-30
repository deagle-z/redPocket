package repository

import (
	"BaseGoUni/core/pojo"
	"testing"
)

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

func TestWithdrawFlowBatchBonusMultiplierConfigUsesDedicatedRebateTransferKey(t *testing.T) {
	got := withdrawFlowBatchBonusMultiplierConfigForSource(pojo.WithdrawFlowBatchSourceRebateTransfer)
	if got.key != withdrawFlowBatchRebateTransferLimitConfigKey {
		t.Fatalf("key = %q, want %q", got.key, withdrawFlowBatchRebateTransferLimitConfigKey)
	}
	if got.defaultValue != defaultWithdrawFlowBatchRebateTransferMultiplier {
		t.Fatalf("defaultValue = %.2f, want %.2f", got.defaultValue, defaultWithdrawFlowBatchRebateTransferMultiplier)
	}
}

func TestWithdrawFlowBatchBonusMultiplierConfigKeepsGiftKeyForOtherSources(t *testing.T) {
	got := withdrawFlowBatchBonusMultiplierConfigForSource("register_gift")
	if got.key != withdrawFlowBatchGiftLimitConfigKey {
		t.Fatalf("key = %q, want %q", got.key, withdrawFlowBatchGiftLimitConfigKey)
	}
	if got.defaultValue != defaultWithdrawFlowBatchGiftMultiplier {
		t.Fatalf("defaultValue = %.2f, want %.2f", got.defaultValue, defaultWithdrawFlowBatchGiftMultiplier)
	}
}
