package api

import "testing"

func TestCanLaunchAppGameAfterMinimumRecharge(t *testing.T) {
	if !canLaunchAppGame(50, false) {
		t.Fatalf("canLaunchAppGame() = false, want true")
	}
}

func TestCanLaunchAppGameRejectsRechargeBelowMinimumWithoutRebateTransfer(t *testing.T) {
	if canLaunchAppGame(49.99, false) {
		t.Fatalf("canLaunchAppGame() = true, want false")
	}
}

func TestCanLaunchAppGameAfterRebateTransfer(t *testing.T) {
	if !canLaunchAppGame(0, true) {
		t.Fatalf("canLaunchAppGame() = false, want true")
	}
}

func TestCanLaunchAppGameAfterRebateTransferBelowMinimumRecharge(t *testing.T) {
	if !canLaunchAppGame(10, true) {
		t.Fatalf("canLaunchAppGame() = false, want true")
	}
}

func TestCanLaunchAppGameRejectsWithoutRechargeOrRebateTransfer(t *testing.T) {
	if canLaunchAppGame(0, false) {
		t.Fatalf("canLaunchAppGame() = true, want false")
	}
}
