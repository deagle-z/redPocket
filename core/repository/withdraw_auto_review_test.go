package repository

import (
	"BaseGoUni/core/pojo"
	"testing"
)

func TestWithdrawAutoReviewEligible(t *testing.T) {
	tests := []struct {
		name   string
		tenant pojo.SysTenant
		amount float64
		want   bool
	}{
		{
			name: "enabled and amount under limit",
			tenant: pojo.SysTenant{
				WithdrawAutoReviewEnabled:    1,
				WithdrawAutoReviewMaxAmount:  100,
				WithdrawAutoReviewDailyLimit: 1000,
			},
			amount: 99.99,
			want:   true,
		},
		{
			name: "enabled and amount equals limit",
			tenant: pojo.SysTenant{
				WithdrawAutoReviewEnabled:    1,
				WithdrawAutoReviewMaxAmount:  100,
				WithdrawAutoReviewDailyLimit: 1000,
			},
			amount: 100,
			want:   true,
		},
		{
			name: "disabled",
			tenant: pojo.SysTenant{
				WithdrawAutoReviewEnabled:    0,
				WithdrawAutoReviewMaxAmount:  100,
				WithdrawAutoReviewDailyLimit: 1000,
			},
			amount: 50,
			want:   false,
		},
		{
			name: "zero max amount disables",
			tenant: pojo.SysTenant{
				WithdrawAutoReviewEnabled:    1,
				WithdrawAutoReviewMaxAmount:  0,
				WithdrawAutoReviewDailyLimit: 1000,
			},
			amount: 50,
			want:   false,
		},
		{
			name: "zero daily limit disables",
			tenant: pojo.SysTenant{
				WithdrawAutoReviewEnabled:    1,
				WithdrawAutoReviewMaxAmount:  100,
				WithdrawAutoReviewDailyLimit: 0,
			},
			amount: 50,
			want:   false,
		},
		{
			name: "amount over limit",
			tenant: pojo.SysTenant{
				WithdrawAutoReviewEnabled:    1,
				WithdrawAutoReviewMaxAmount:  100,
				WithdrawAutoReviewDailyLimit: 1000,
			},
			amount: 100.01,
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := withdrawAutoReviewEligible(tt.tenant, tt.amount); got != tt.want {
				t.Fatalf("withdrawAutoReviewEligible() = %v, want %v", got, tt.want)
			}
		})
	}
}
