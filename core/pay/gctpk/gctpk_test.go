package gctpk

import (
	"BaseGoUni/core/base"
	"testing"
)

func TestResolveNotifyURLUsesConfigNotifyURL(t *testing.T) {
	cfg := base.GctpkPayConfig{NotifyURL: "https://config.example.com/gctpk/notify"}

	if got := resolveNotifyURL(cfg); got != cfg.NotifyURL {
		t.Fatalf("resolveNotifyURL = %q, want config notify URL", got)
	}
}
