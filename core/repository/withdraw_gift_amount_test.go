package repository

import "testing"

func TestWithdrawGiftAmountAfterDeductClampsToZero(t *testing.T) {
	got := withdrawGiftAmountAfterDeduct(26.34, 50.82, 0.41)
	if got != 0 {
		t.Fatalf("withdrawGiftAmountAfterDeduct() = %.2f, want 0.00", got)
	}
}

func TestWithdrawGiftAmountAfterDeductCapsAtRemainingBalance(t *testing.T) {
	got := withdrawGiftAmountAfterDeduct(26.34, 0, 0.41)
	if got != 0.41 {
		t.Fatalf("withdrawGiftAmountAfterDeduct() = %.2f, want 0.41", got)
	}
}

func TestWithdrawGiftAmountAfterRefundRestoresGiftPortion(t *testing.T) {
	got := withdrawGiftAmountAfterRefund(0.41, 26.34, 51.23)
	if got != 26.75 {
		t.Fatalf("withdrawGiftAmountAfterRefund() = %.2f, want 26.75", got)
	}
}
