package repository

import "testing"

func TestCalculateWithdrawNetAmount(t *testing.T) {
	tests := []struct {
		name   string
		amount float64
		fee    float64
		rate   float64
		want   float64
	}{
		{name: "converts fee deducted amount and rounds up for provider", amount: 44, fee: 0, rate: 3.27, want: 144},
		{name: "deducts fee before conversion", amount: 100, fee: 5, rate: 2.5, want: 238},
		{name: "falls back to rate one when rate is missing", amount: 100, fee: 10, rate: 0, want: 90},
		{name: "returns zero when fee consumes amount", amount: 50, fee: 50, rate: 3, want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := calculateWithdrawNetAmount(tt.amount, tt.fee, tt.rate); got != tt.want {
				t.Fatalf("calculateWithdrawNetAmount(%v, %v, %v) = %v, want %v", tt.amount, tt.fee, tt.rate, got, tt.want)
			}
		})
	}
}
