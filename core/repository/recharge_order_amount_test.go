package repository

import (
	"BaseGoUni/core/pojo"
	"testing"
)

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

func TestNormalizeRechargeOrderAmount(t *testing.T) {
	tests := []struct {
		name   string
		amount float64
		want   float64
	}{
		{name: "integer amount is unchanged", amount: 50, want: 50},
		{name: "decimal amount keeps two decimals", amount: 29.99, want: 29.99},
		{name: "extra precision is truncated", amount: 29.999, want: 29.99},
		{name: "amount below one keeps cents", amount: 0.99, want: 0.99},
		{name: "negative amount becomes zero", amount: -1, want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizeRechargeOrderAmount(tt.amount); got != tt.want {
				t.Fatalf("normalizeRechargeOrderAmount(%v) = %v, want %v", tt.amount, got, tt.want)
			}
		})
	}
}

func TestFormatRechargeMinAmount(t *testing.T) {
	tests := []struct {
		amount float64
		want   string
	}{
		{amount: 29.99, want: "29.99"},
		{amount: 50, want: "50"},
	}

	for _, tt := range tests {
		if got := formatRechargeMinAmount(tt.amount); got != tt.want {
			t.Fatalf("formatRechargeMinAmount(%v) = %q, want %q", tt.amount, got, tt.want)
		}
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
		{name: "keeps cents when rate is one", amount: 29.99, rate: 1, want: 29.99},
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

func TestApplyRechargeV2DefaultProviderUsesVcpayForMexico(t *testing.T) {
	req := applyRechargeV2DefaultProvider(pojo.RechargeOrderAppReq{
		CountryCode: " mx ",
	})

	if req.Channel != "VCPAYMXN" {
		t.Fatalf("Channel = %q, want VCPAYMXN", req.Channel)
	}
	if req.Currency != "MXN" {
		t.Fatalf("Currency = %q, want MXN", req.Currency)
	}
}

func TestApplyRechargeV2DefaultProviderKeepsExplicitChannel(t *testing.T) {
	req := applyRechargeV2DefaultProvider(pojo.RechargeOrderAppReq{
		CountryCode: "MX",
		Channel:     "GCTPKMXN",
	})

	if req.Channel != "GCTPKMXN" {
		t.Fatalf("Channel = %q, want GCTPKMXN", req.Channel)
	}
}
