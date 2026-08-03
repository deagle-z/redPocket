package utils

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestNormalizeDomain(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "origin URL", in: "https://WWW.Example.com:8443/path", want: "www.example.com"},
		{name: "plain host and port", in: "Promo.Example.com:8080", want: "promo.example.com"},
		{name: "trailing dot", in: "EXAMPLE.COM.", want: "example.com"},
		{name: "null origin", in: "null", want: ""},
		{name: "empty", in: " ", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NormalizeDomain(tt.in); got != tt.want {
				t.Fatalf("NormalizeDomain() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestAttributionRequestDomainPrecedence(t *testing.T) {
	tests := []struct {
		name    string
		origin  string
		referer string
		host    string
		want    string
	}{
		{name: "origin first", origin: "https://promo.example.com", referer: "https://ref.example.com/page", host: "api.example.com", want: "promo.example.com"},
		{name: "referer after null origin", origin: "null", referer: "https://ref.example.com/page", host: "api.example.com", want: "ref.example.com"},
		{name: "request host last", host: "API.Example.com:8080", want: "api.example.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(recorder)
			ctx.Request = httptest.NewRequest("POST", "http://"+tt.host+"/api/v1/app/attribution/event", nil)
			ctx.Request.Host = tt.host
			if tt.origin != "" {
				ctx.Request.Header.Set("Origin", tt.origin)
			}
			if tt.referer != "" {
				ctx.Request.Header.Set("Referer", tt.referer)
			}

			if got := AttributionRequestDomain(ctx); got != tt.want {
				t.Fatalf("AttributionRequestDomain() = %q, want %q", got, tt.want)
			}
		})
	}
}
