package pojo

import (
	"encoding/json"
	"testing"
)

func TestWithdrawOrderBrSetUnmarshalLocalDateTime(t *testing.T) {
	body := []byte(`{
		"id": 12,
		"reviewedAt": "2026-05-09 02:32:26",
		"paidAt": "2026-05-09 03:32:26",
		"notifyTime": "2026-05-09T04:32:26Z"
	}`)

	var req WithdrawOrderBrSet
	if err := json.Unmarshal(body, &req); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if req.ReviewedAt == nil || req.ReviewedAt.Format("2006-01-02 15:04:05") != "2026-05-09 02:32:26" {
		t.Fatalf("ReviewedAt = %v", req.ReviewedAt)
	}
	if req.PaidAt == nil || req.PaidAt.Format("2006-01-02 15:04:05") != "2026-05-09 03:32:26" {
		t.Fatalf("PaidAt = %v", req.PaidAt)
	}
	if req.NotifyTime == nil || req.NotifyTime.UTC().Format("2006-01-02 15:04:05") != "2026-05-09 04:32:26" {
		t.Fatalf("NotifyTime = %v", req.NotifyTime)
	}
}

func TestWithdrawOrderBrSetUnmarshalCountryCode(t *testing.T) {
	body := []byte(`{
		"id": 12,
		"countryCode": " br "
	}`)

	var req WithdrawOrderBrSet
	if err := json.Unmarshal(body, &req); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if req.CountryCode != "BR" {
		t.Fatalf("CountryCode = %q, want BR", req.CountryCode)
	}
	if !req.HasJSONField("countryCode") {
		t.Fatalf("HasJSONField(countryCode) = false, want true")
	}
	if req.HasJSONField("amount") {
		t.Fatalf("HasJSONField(amount) = true, want false")
	}
}

func TestWithdrawOrderBrSetNormalizeCountryCodeFromExtra(t *testing.T) {
	extra := `{"countryCode":"mx"}`
	req := WithdrawOrderBrSet{Extra: &extra}

	req.NormalizeCountryCodeFromExtra()

	if req.CountryCode != "MX" {
		t.Fatalf("CountryCode = %q, want MX", req.CountryCode)
	}
}
