package repository

import "testing"

func TestOrderCurrencySymbol(t *testing.T) {
	tests := []struct {
		name     string
		currency string
		want     string
	}{
		{name: "brl", currency: "BRL", want: "R$"},
		{name: "usd", currency: "USD", want: "$"},
		{name: "mxn lowercase", currency: "mxn", want: "$"},
		{name: "cny", currency: "CNY", want: "¥"},
		{name: "unknown fallback", currency: "cop", want: "COP"},
		{name: "blank", currency: " ", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := orderCurrencySymbol(tt.currency); got != tt.want {
				t.Fatalf("orderCurrencySymbol(%q) = %q, want %q", tt.currency, got, tt.want)
			}
		})
	}
}

func TestBuildAppRechargeOrderHistoryItem(t *testing.T) {
	item := buildAppRechargeOrderHistoryItem(rechargeOrderHistorySource{
		OrderNo:     "RC123",
		Amount:      50.6,
		NetAmount:   50,
		Currency:    "brl",
		BonusAmount: 5.2,
		Status:      1,
	})

	if item.OrderNo != "RC123" {
		t.Fatalf("OrderNo = %q, want RC123", item.OrderNo)
	}
	if item.BonusAmount != 5.2 {
		t.Fatalf("BonusAmount = %.2f, want 5.20", item.BonusAmount)
	}
	if item.Fee != nil {
		t.Fatalf("Fee = %v, want nil for recharge", *item.Fee)
	}
	if item.RejectReason != nil {
		t.Fatalf("RejectReason = %q, want nil for recharge", *item.RejectReason)
	}
}

func TestBuildAppWithdrawOrderHistoryItem(t *testing.T) {
	reason := "receiver info invalid"
	item := buildAppWithdrawOrderHistoryItem(withdrawOrderHistorySource{
		OrderNo:      "WD123",
		Amount:       100,
		NetAmount:    95,
		Currency:     "mxn",
		Fee:          5,
		Status:       4,
		RejectReason: &reason,
	})

	if item.OrderNo != "WD123" {
		t.Fatalf("OrderNo = %q, want WD123", item.OrderNo)
	}
	if item.BonusAmount != nil {
		t.Fatalf("BonusAmount = %v, want nil for withdraw", *item.BonusAmount)
	}
	if item.Fee == nil || *item.Fee != 5 {
		t.Fatalf("Fee = %v, want 5", item.Fee)
	}
	if item.RejectReason == nil || *item.RejectReason != reason {
		t.Fatalf("RejectReason = %v, want %q", item.RejectReason, reason)
	}
}
