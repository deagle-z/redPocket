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
