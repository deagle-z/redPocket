package repository

import "testing"

func TestNormalizeTgChannelName(t *testing.T) {
	got, err := NormalizeTgChannelName(" Osanvnei ")
	if err != nil {
		t.Fatalf("NormalizeTgChannelName returned error: %v", err)
	}
	if got != "@Osanvnei" {
		t.Fatalf("NormalizeTgChannelName = %q, want %q", got, "@Osanvnei")
	}

	got, err = NormalizeTgChannelName("@Osanvnei")
	if err != nil {
		t.Fatalf("NormalizeTgChannelName returned error: %v", err)
	}
	if got != "@Osanvnei" {
		t.Fatalf("NormalizeTgChannelName = %q, want %q", got, "@Osanvnei")
	}
}

func TestNormalizeTgChannelNameRequiresValue(t *testing.T) {
	if _, err := NormalizeTgChannelName(" @ "); err == nil {
		t.Fatal("NormalizeTgChannelName should reject empty values")
	}
}
