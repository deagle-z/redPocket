package repository

import "testing"

func TestNormalizeSourceChannelCode(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "empty", in: "", want: ""},
		{name: "trim", in: "  abc123  ", want: "ABC123"},
		{name: "already uppercase", in: "ABC123", want: "ABC123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NormalizeSourceChannelCode(tt.in); got != tt.want {
				t.Fatalf("NormalizeSourceChannelCode() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFirstSourceChannelCode(t *testing.T) {
	got := FirstSourceChannelCode(" ", " abc123 ", "def456")
	if got != "ABC123" {
		t.Fatalf("FirstSourceChannelCode() = %q, want %q", got, "ABC123")
	}
}

func TestBuildSourceChannelLinkURL(t *testing.T) {
	tests := []struct {
		name    string
		baseURL string
		code    string
		want    string
	}{
		{name: "base url", baseURL: "https://example.com", code: "abc123", want: "https://example.com/register?sc=ABC123"},
		{name: "trim trailing slash", baseURL: "https://example.com/", code: " abc123 ", want: "https://example.com/register?sc=ABC123"},
		{name: "no base url", baseURL: "", code: "abc123", want: "/register?sc=ABC123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := BuildSourceChannelLinkURL(tt.baseURL, tt.code); got != tt.want {
				t.Fatalf("BuildSourceChannelLinkURL() = %q, want %q", got, tt.want)
			}
		})
	}
}
