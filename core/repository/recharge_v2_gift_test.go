package repository

import (
	"BaseGoUni/core/pojo"
	"testing"
)

func TestRechargeV2GiftRate(t *testing.T) {
	tests := []struct {
		name    string
		amount  float64
		isFirst bool
		want    float64
	}{
		{name: "first recharge below 1000 uses 18 percent", amount: 999, isFirst: true, want: 18},
		{name: "first recharge at 1000 uses 20 percent", amount: 1000, isFirst: true, want: 20},
		{name: "first recharge below 5000 uses 20 percent", amount: 4999, isFirst: true, want: 20},
		{name: "first recharge at 5000 uses 25 percent", amount: 5000, isFirst: true, want: 25},
		{name: "non-first recharge uses 10 percent", amount: 5000, isFirst: false, want: 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := rechargeV2GiftRate(tt.amount, tt.isFirst); got != tt.want {
				t.Fatalf("rechargeV2GiftRate(%.2f, %t) = %.2f, want %.2f", tt.amount, tt.isFirst, got, tt.want)
			}
		})
	}
}

func TestCalculateRechargeV2GiftAmountByTier(t *testing.T) {
	tests := []struct {
		name    string
		amount  float64
		isFirst bool
		want    float64
	}{
		{name: "first recharge below 1000 gifts 18 percent", amount: 999, isFirst: true, want: 179.82},
		{name: "first recharge at 1000 gifts 20 percent", amount: 1000, isFirst: true, want: 200},
		{name: "first recharge below 5000 gifts 20 percent", amount: 4999, isFirst: true, want: 999.8},
		{name: "first recharge at 5000 gifts 25 percent", amount: 5000, isFirst: true, want: 1250},
		{name: "non-first recharge gifts 10 percent", amount: 1234.56, isFirst: false, want: 123.45},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := calculateRechargeV2GiftAmount(tt.amount, tt.isFirst); got != tt.want {
				t.Fatalf("calculateRechargeV2GiftAmount(%.2f, %t) = %.2f, want %.2f", tt.amount, tt.isFirst, got, tt.want)
			}
		})
	}
}

func TestCalculateRechargeV2GiftAmountTruncatesToCents(t *testing.T) {
	got := calculateRechargeV2GiftAmount(333.33, false)
	if got != 33.33 {
		t.Fatalf("calculateRechargeV2GiftAmount() = %.2f, want 33.33", got)
	}
}

func TestIsRechargeV2FirstRechargeAmount(t *testing.T) {
	tests := []struct {
		name           string
		rechargeAmount float64
		want           bool
	}{
		{name: "zero recharge amount is first recharge", rechargeAmount: 0, want: true},
		{name: "negative recharge amount is treated as first recharge", rechargeAmount: -1, want: true},
		{name: "positive recharge amount is not first recharge", rechargeAmount: 0.01, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isRechargeV2FirstRechargeAmount(tt.rechargeAmount); got != tt.want {
				t.Fatalf("isRechargeV2FirstRechargeAmount(%.2f) = %t, want %t", tt.rechargeAmount, got, tt.want)
			}
		})
	}
}

func TestResolveRechargeOrderCreateActivityTypeForcedV2(t *testing.T) {
	req := pojo.RechargeOrderAppReq{
		ActivityType: 1,
		ActivityCode: RechargeActivityCodeFirstRecharge3Day,
	}
	forced := rechargeActivityTypeV2Gift

	got, err := resolveRechargeOrderCreateActivityType(nil, 123, req, "", &forced)
	if err != nil {
		t.Fatalf("resolveRechargeOrderCreateActivityType() error = %v", err)
	}
	if got != rechargeActivityTypeV2Gift {
		t.Fatalf("resolveRechargeOrderCreateActivityType() = %d, want %d", got, rechargeActivityTypeV2Gift)
	}
}

func TestShouldConfirmUnfinishedActivityCycleForRechargeV2(t *testing.T) {
	if !shouldConfirmUnfinishedActivityCycleForRecharge(rechargeActivityTypeV2Gift) {
		t.Fatalf("shouldConfirmUnfinishedActivityCycleForRecharge(v2) = false, want true")
	}
	if !shouldConfirmUnfinishedActivityCycleForRecharge(0) {
		t.Fatalf("shouldConfirmUnfinishedActivityCycleForRecharge(0) = false, want true")
	}
	if shouldConfirmUnfinishedActivityCycleForRecharge(1) {
		t.Fatalf("shouldConfirmUnfinishedActivityCycleForRecharge(1) = true, want false")
	}
}

func TestRechargeActivityCodeByTypeSupportsV2Gift(t *testing.T) {
	if got := rechargeActivityCodeByType(rechargeActivityTypeV2Gift); got != RechargeActivityCodeV2Gift {
		t.Fatalf("rechargeActivityCodeByType(v2) = %q, want %q", got, RechargeActivityCodeV2Gift)
	}
}

func TestFormatRechargeActivityTypeSupportsV2Gift(t *testing.T) {
	activityType := rechargeActivityTypeV2Gift
	if got := formatRechargeActivityType(&activityType); got != "3(recharge_v2_gift)" {
		t.Fatalf("formatRechargeActivityType(v2) = %q", got)
	}
}
