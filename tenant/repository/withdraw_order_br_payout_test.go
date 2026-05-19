package repository

import (
	"BaseGoUni/core/pojo"
	"testing"

	"gorm.io/gorm"
)

func TestApplyTenantWithdrawStatusSideEffectsSubmitsPayoutOnApproval(t *testing.T) {
	called := false
	original := submitTenantWithdrawPayout
	submitTenantWithdrawPayout = func(_ *gorm.DB, order *pojo.WithdrawOrderBr) error {
		called = true
		if order.ID != 99 {
			t.Fatalf("order.ID = %d, want 99", order.ID)
		}
		return nil
	}
	defer func() { submitTenantWithdrawPayout = original }()

	order := pojo.WithdrawOrderBr{BaseModel: pojo.BaseModel{ID: 99}, Status: 1}
	if err := applyTenantWithdrawStatusSideEffects(nil, 0, &order); err != nil {
		t.Fatalf("applyTenantWithdrawStatusSideEffects() error = %v", err)
	}
	if !called {
		t.Fatal("expected approval to submit payout")
	}
}

func TestApplyTenantWithdrawStatusSideEffectsSkipsPayoutForNonApproval(t *testing.T) {
	called := false
	original := submitTenantWithdrawPayout
	submitTenantWithdrawPayout = func(_ *gorm.DB, _ *pojo.WithdrawOrderBr) error {
		called = true
		return nil
	}
	defer func() { submitTenantWithdrawPayout = original }()

	order := pojo.WithdrawOrderBr{BaseModel: pojo.BaseModel{ID: 99}, Status: 6}
	if err := applyTenantWithdrawStatusSideEffects(nil, 0, &order); err != nil {
		t.Fatalf("applyTenantWithdrawStatusSideEffects() error = %v", err)
	}
	if called {
		t.Fatal("expected rejected order not to submit payout")
	}
}
