package repository

import (
	"BaseGoUni/core/pojo"
	"testing"
)

func TestBuildWithdrawActivityFlowActiveCycle(t *testing.T) {
	cycle := pojo.TgUserWithdrawActivityCycle{
		BaseModel:      pojo.BaseModel{ID: 12},
		Status:         pojo.WithdrawActivityCycleStatusActive,
		ActivityCode:   "first_recharge_3day",
		ActivityType:   1,
		Multiplier:     5,
		BaseAmount:     100,
		RequiredFlow:   500,
		FlowStartValue: 40,
		FlowConsumed:   60,
	}

	got := buildWithdrawActivityFlowCycle(cycle, 275)

	if got.ID != 12 {
		t.Fatalf("ID = %d, want 12", got.ID)
	}
	if got.ActivityCode != "first_recharge_3day" {
		t.Fatalf("ActivityCode = %q, want first_recharge_3day", got.ActivityCode)
	}
	if got.CurrentFlow != 235 {
		t.Fatalf("CurrentFlow = %.2f, want 235.00", got.CurrentFlow)
	}
	if got.AvailableFlow != 175 {
		t.Fatalf("AvailableFlow = %.2f, want 175.00", got.AvailableFlow)
	}
	if got.RemainingFlow != 325 {
		t.Fatalf("RemainingFlow = %.2f, want 325.00", got.RemainingFlow)
	}
	if got.ProgressPercent != 35 {
		t.Fatalf("ProgressPercent = %.2f, want 35.00", got.ProgressPercent)
	}
}

func TestBuildWithdrawActivityFlowClampsCompletedCycle(t *testing.T) {
	cycle := pojo.TgUserWithdrawActivityCycle{
		BaseModel:      pojo.BaseModel{ID: 13},
		Status:         pojo.WithdrawActivityCycleStatusEnded,
		RequiredFlow:   200,
		FlowStartValue: 10,
		FlowConsumed:   0,
	}

	got := buildWithdrawActivityFlowCycle(cycle, 260)

	if got.CurrentFlow != 250 {
		t.Fatalf("CurrentFlow = %.2f, want 250.00", got.CurrentFlow)
	}
	if got.AvailableFlow != 250 {
		t.Fatalf("AvailableFlow = %.2f, want 250.00", got.AvailableFlow)
	}
	if got.RemainingFlow != 0 {
		t.Fatalf("RemainingFlow = %.2f, want 0.00", got.RemainingFlow)
	}
	if got.ProgressPercent != 100 {
		t.Fatalf("ProgressPercent = %.2f, want 100.00", got.ProgressPercent)
	}
}
