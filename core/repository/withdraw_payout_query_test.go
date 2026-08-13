package repository

import (
	"BaseGoUni/core/pay"
	"BaseGoUni/core/pojo"
	"testing"
	"time"
)

func TestWithdrawPayoutQueryTimeoutIsFiveHours(t *testing.T) {
	if WithdrawPayoutQueryTimeout != 5*time.Hour {
		t.Fatalf("WithdrawPayoutQueryTimeout = %s, want 5h", WithdrawPayoutQueryTimeout)
	}
}

func TestWithdrawPayoutQueryDecisionMapsProviderStatuses(t *testing.T) {
	tests := []struct {
		name       string
		status     int
		wantAction withdrawPayoutQueryAction
	}{
		{name: "success", status: 7, wantAction: withdrawPayoutQueryActionSuccess},
		{name: "failed status 2", status: 2, wantAction: withdrawPayoutQueryActionFailed},
		{name: "failed status 6", status: 6, wantAction: withdrawPayoutQueryActionFailed},
		{name: "failed status 8", status: 8, wantAction: withdrawPayoutQueryActionFailed},
		{name: "processing status 0", status: 0, wantAction: withdrawPayoutQueryActionProcessing},
		{name: "processing status 9", status: 9, wantAction: withdrawPayoutQueryActionProcessing},
		{name: "unknown", status: 99, wantAction: withdrawPayoutQueryActionProcessing},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := withdrawPayoutQueryDecision(pay.PayoutQueryResponse{Status: tt.status}); got != tt.wantAction {
				t.Fatalf("withdrawPayoutQueryDecision(status=%d) = %v, want %v", tt.status, got, tt.wantAction)
			}
		})
	}
}

func TestWithdrawPayoutProviderCodeUsesProviderOnly(t *testing.T) {
	provider := " GCTPKMXN "
	order := pojo.WithdrawOrderBr{
		Channel:  "GCTPKBRL",
		Provider: &provider,
	}

	if got := withdrawPayoutProviderCode(order); got != "GCTPKMXN" {
		t.Fatalf("withdrawPayoutProviderCode() = %q, want GCTPKMXN", got)
	}
}

func TestWithdrawPayoutProviderCodeDoesNotFallbackToChannel(t *testing.T) {
	order := pojo.WithdrawOrderBr{Channel: "GCTPKMXN"}

	if got := withdrawPayoutProviderCode(order); got != "" {
		t.Fatalf("withdrawPayoutProviderCode() = %q, want empty", got)
	}
}

func TestWithdrawPayoutQueryRequestUsesMerchantOrderAndProviderOrder(t *testing.T) {
	merchantOrderNo := " WD202606230001 "
	providerPayoutNo := " PAYOUT202606230001 "
	order := pojo.WithdrawOrderBr{
		MerchantOrderNo:  &merchantOrderNo,
		ProviderPayoutNo: &providerPayoutNo,
	}

	req := withdrawPayoutQueryRequest(order)

	if req.OrderNo != "WD202606230001" {
		t.Fatalf("OrderNo = %q, want WD202606230001", req.OrderNo)
	}
	if req.ProviderOrderNo != "PAYOUT202606230001" {
		t.Fatalf("ProviderOrderNo = %q, want PAYOUT202606230001", req.ProviderOrderNo)
	}
}
