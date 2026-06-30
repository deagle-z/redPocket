package api

import "testing"

func TestCanLaunchAppGameAfterRecharge(t *testing.T) {
	if !canLaunchAppGame(10, false) {
		t.Fatalf("canLaunchAppGame() = false, want true")
	}
}

func TestCanLaunchAppGameAfterRebateTransfer(t *testing.T) {
	if !canLaunchAppGame(0, true) {
		t.Fatalf("canLaunchAppGame() = false, want true")
	}
}

func TestCanLaunchAppGameRejectsWithoutRechargeOrRebateTransfer(t *testing.T) {
	if canLaunchAppGame(0, false) {
		t.Fatalf("canLaunchAppGame() = true, want false")
	}
}
