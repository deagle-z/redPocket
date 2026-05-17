package repository

import "testing"

func TestFloorRechargeAmount(t *testing.T) {
	tests := []struct {
		name   string
		amount float64
		want   float64
	}{
		{name: "integer amount is unchanged", amount: 50, want: 50},
		{name: "decimal amount is floored", amount: 50.6, want: 50},
		{name: "amount below one becomes zero", amount: 0.9, want: 0},
		{name: "negative amount becomes zero", amount: -1, want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := floorRechargeAmount(tt.amount); got != tt.want {
				t.Fatalf("floorRechargeAmount(%v) = %v, want %v", tt.amount, got, tt.want)
			}
		})
	}
}

func TestCalculateRechargeProviderAmount(t *testing.T) {
	tests := []struct {
		name   string
		amount float64
		rate   float64
		want   float64
	}{
		{name: "floors converted provider amount", amount: 10, rate: 5.06, want: 50},
		{name: "keeps integer converted provider amount", amount: 10, rate: 5, want: 50},
		{name: "uses original amount when rate is missing", amount: 50.6, rate: 0, want: 50.6},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := calculateRechargeProviderAmount(tt.amount, tt.rate); got != tt.want {
				t.Fatalf("calculateRechargeProviderAmount(%v, %v) = %v, want %v", tt.amount, tt.rate, got, tt.want)
			}
		})
	}
}
