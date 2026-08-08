package repository

import (
	"BaseGoUni/core/pojo"
	"testing"
)

func TestParseRechargeV2GiftRates(t *testing.T) {
	if got, ok := parseRechargeV2GiftRates("8,9,10,11,12,13,14,15,16,17,18"); !ok || got != [rechargeV2GiftRateTierCount]float64{8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18} {
		t.Fatalf("parseRechargeV2GiftRates valid = %v,%t", got, ok)
	}
	if _, ok := parseRechargeV2GiftRates("8,9,10,11"); ok {
		t.Fatalf("parseRechargeV2GiftRates with fewer than 11 parts should fail")
	}
	if _, ok := parseRechargeV2GiftRates("8,9,10,x,12,13,14,15,16,17,18"); ok {
		t.Fatalf("parseRechargeV2GiftRates with non-number should fail")
	}
}

func TestRechargeV2GiftRateByNumber(t *testing.T) {
	// nil db -> 默认 [8,9,10,11,12,13,14,15,16,17,18]
	tests := []struct {
		number int
		want   float64
	}{
		{number: 1, want: 8},
		{number: 2, want: 9},
		{number: 3, want: 10},
		{number: 4, want: 11},
		{number: 9, want: 16},
		{number: 11, want: 18},
		{number: 12, want: 18},
		{number: 0, want: 8},
	}
	for _, tt := range tests {
		if got := rechargeV2GiftRateByNumber(nil, tt.number); got != tt.want {
			t.Fatalf("rechargeV2GiftRateByNumber(nil, %d) = %.2f, want %.2f", tt.number, got, tt.want)
		}
	}
}

func TestCalculateRechargeV2GiftAmountByNumber(t *testing.T) {
	// nil db -> 默认 [8,9,10,11,12,13,14,15,16,17,18]，门槛 50
	tests := []struct {
		name   string
		amount float64
		number int
		want   float64
	}{
		{name: "below threshold no gift", amount: 49.99, number: 1, want: 0},
		{name: "at threshold 50 first recharge 8%", amount: 50, number: 1, want: 4},
		{name: "first recharge 8%", amount: 1000, number: 1, want: 80},
		{name: "second recharge 9%", amount: 1000, number: 2, want: 90},
		{name: "third recharge 10%", amount: 1000, number: 3, want: 100},
		{name: "eleventh+ recharge 18%", amount: 1234.56, number: 12, want: 222.22},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := calculateRechargeV2GiftAmount(nil, tt.amount, tt.number); got != tt.want {
				t.Fatalf("calculateRechargeV2GiftAmount(nil, %.2f, %d) = %.2f, want %.2f", tt.amount, tt.number, got, tt.want)
			}
		})
	}
}

func TestRechargeGiftBaseAmountPrefersActualCreditAmount(t *testing.T) {
	creditAmount := 20.0
	order := pojo.RechargeOrder{
		Amount:       500,
		CreditAmount: &creditAmount,
	}
	if got := rechargeGiftBaseAmount(order); got != 20 {
		t.Fatalf("rechargeGiftBaseAmount() = %.2f, want 20.00", got)
	}
	if got := calculateRechargeV2GiftAmount(nil, rechargeGiftBaseAmount(order), 1); got != 0 {
		t.Fatalf("calculateRechargeV2GiftAmount(actual 20) = %.2f, want 0.00", got)
	}
}

func TestCalculateRechargeFreeLotteryCount(t *testing.T) {
	tests := []struct {
		amount float64
		want   int
	}{
		{amount: 0, want: 0},
		{amount: 199.99, want: 0},
		{amount: 200, want: 1},
		{amount: 399.99, want: 1},
		{amount: 400, want: 2},
		{amount: 1000, want: 5},
	}

	for _, tt := range tests {
		if got := calculateRechargeFreeLotteryCount(tt.amount); got != tt.want {
			t.Fatalf("calculateRechargeFreeLotteryCount(%.2f) = %d, want %d", tt.amount, got, tt.want)
		}
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
