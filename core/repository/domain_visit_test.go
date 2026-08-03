package repository

import (
	"BaseGoUni/core/pojo"
	"testing"
	"time"
)

func TestDomainVisitScopeForBinding(t *testing.T) {
	tests := []struct {
		name          string
		binding       string
		wantCondition string
		wantArg       string
		wantErr       bool
	}{
		{name: "exact", binding: "promo.example.com", wantCondition: "domain = ?", wantArg: "promo.example.com"},
		{name: "wildcard", binding: "*.example.com", wantCondition: "domain LIKE ?", wantArg: "%.example.com"},
		{name: "normalized exact", binding: " Promo.Example.com. ", wantCondition: "domain = ?", wantArg: "promo.example.com"},
		{name: "invalid wildcard", binding: "*.", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scope, err := domainVisitScopeForBinding(tt.binding)
			if tt.wantErr {
				if err == nil {
					t.Fatal("domainVisitScopeForBinding() expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("domainVisitScopeForBinding() error = %v", err)
			}
			if scope.condition != tt.wantCondition {
				t.Fatalf("condition = %q, want %q", scope.condition, tt.wantCondition)
			}
			if len(scope.args) != 1 || scope.args[0] != tt.wantArg {
				t.Fatalf("args = %#v, want %q", scope.args, tt.wantArg)
			}
		})
	}
}

func TestNewDomainVisitDetailResponseHasEveryHour(t *testing.T) {
	start := time.Date(2026, 8, 1, 0, 0, 0, 0, time.Local)
	result := newDomainVisitDetailResponse(pojo.PageInfo{CurrentPage: 2, PageSize: 20}, start)
	if result.Date != "2026-08-01" || len(result.Hourly) != 24 {
		t.Fatalf("unexpected response date=%q hours=%d", result.Date, len(result.Hourly))
	}
	for hour, row := range result.Hourly {
		if row.Hour != hour || row.PageViews != 0 || row.UniqueVisitors != 0 {
			t.Fatalf("hour %d = %#v", hour, row)
		}
	}
	if result.List == nil || result.CurrentPage != 2 || result.PageSize != 20 {
		t.Fatalf("unexpected pagination response: %#v", result.BasePageResponse)
	}
}
