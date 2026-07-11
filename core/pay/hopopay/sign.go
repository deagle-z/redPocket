package hopopay

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"
)

// BuildSignature builds a Hopopay request signature.
func BuildSignature(params map[string]any, secret string) string {
	keys := sortedSignKeys(params)
	var builder strings.Builder
	for _, key := range keys {
		builder.WriteString(base64.StdEncoding.EncodeToString([]byte(key)))
		builder.WriteString(base64.StdEncoding.EncodeToString([]byte(signValueToString(params[key]))))
		builder.WriteByte('&')
	}
	builder.WriteString(secret)
	sum := md5.Sum([]byte(builder.String()))
	return hex.EncodeToString(sum[:])
}

// VerifySignature verifies a Hopopay request/response signature from decoded values.
func VerifySignature(params map[string]any, secret string) bool {
	sign := strings.TrimSpace(signValueToString(params["signature"]))
	if sign == "" {
		return false
	}
	return strings.EqualFold(sign, BuildSignature(params, secret))
}

// BuildSignatureFromRawJSON builds a Hopopay callback signature while preserving
// the numeric JSON token format from the original callback body.
func BuildSignatureFromRawJSON(params map[string]json.RawMessage, secret string) string {
	keys := make([]string, 0, len(params))
	for key := range params {
		if key == "signature" {
			continue
		}
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var builder strings.Builder
	for _, key := range keys {
		builder.WriteString(base64.StdEncoding.EncodeToString([]byte(key)))
		builder.WriteString(base64.StdEncoding.EncodeToString([]byte(rawJSONSignValue(params[key]))))
		builder.WriteByte('&')
	}
	builder.WriteString(secret)
	sum := md5.Sum([]byte(builder.String()))
	return hex.EncodeToString(sum[:])
}

// VerifyRawJSONSignature verifies a Hopopay callback signature from raw JSON values.
func VerifyRawJSONSignature(params map[string]json.RawMessage, secret string) bool {
	signatureRaw, ok := params["signature"]
	if !ok {
		return false
	}
	sign := strings.TrimSpace(rawJSONSignValue(signatureRaw))
	if sign == "" {
		return false
	}
	return strings.EqualFold(sign, BuildSignatureFromRawJSON(params, secret))
}

func sortedSignKeys(params map[string]any) []string {
	keys := make([]string, 0, len(params))
	for key := range params {
		if key == "signature" {
			continue
		}
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func signValueToString(value any) string {
	if value == nil {
		return ""
	}
	switch v := value.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	case json.Number:
		return v.String()
	case fmt.Stringer:
		return v.String()
	}

	rv := reflect.ValueOf(value)
	if rv.IsValid() && (rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array) {
		encoded, _ := json.Marshal(value)
		return string(encoded)
	}
	return fmt.Sprint(value)
}

func rawJSONSignValue(raw json.RawMessage) string {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 {
		return ""
	}
	if bytes.Equal(trimmed, []byte("null")) {
		return "null"
	}
	if trimmed[0] == '"' {
		var text string
		if err := json.Unmarshal(trimmed, &text); err == nil {
			return text
		}
	}
	if trimmed[0] == '[' || trimmed[0] == '{' {
		var compact bytes.Buffer
		if err := json.Compact(&compact, trimmed); err == nil {
			return compact.String()
		}
	}
	return string(trimmed)
}
