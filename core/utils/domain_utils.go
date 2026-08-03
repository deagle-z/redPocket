package utils

import (
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
)

const maxDomainLength = 253

// NormalizeDomain extracts and normalizes a hostname from a URL or host value.
func NormalizeDomain(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" || strings.EqualFold(raw, "null") {
		return ""
	}

	parsed, err := url.Parse(raw)
	if err != nil || parsed.Host == "" {
		parsed, err = url.Parse("//" + raw)
	}
	if err != nil {
		return ""
	}

	domain := strings.TrimSuffix(strings.ToLower(strings.TrimSpace(parsed.Hostname())), ".")
	if domain == "" || len(domain) > maxDomainLength {
		return ""
	}
	return domain
}

// AttributionRequestDomain returns the H5 domain represented by the request.
// Browser Origin and Referer take precedence because the API may use a separate host.
func AttributionRequestDomain(ctx *gin.Context) string {
	for _, raw := range []string{
		ctx.GetHeader("Origin"),
		ctx.GetHeader("Referer"),
		ctx.Request.Host,
	} {
		if domain := NormalizeDomain(raw); domain != "" {
			return domain
		}
	}
	return ""
}
