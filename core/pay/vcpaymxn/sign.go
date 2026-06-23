package vcpaymxn

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
)

var callbackSignExcludedKeys = map[string]struct{}{
	"code": {},
	"msg":  {},
}

// BuildSign builds a VcPay request/response signature.
// Empty values and sign itself are excluded. Remaining keys are sorted by ASCII
// order and joined as key=value pairs, then &key=appKey is appended before MD5.
func BuildSign(params map[string]string, appKey string) string {
	return buildSign(params, appKey, nil)
}

// BuildCallbackSign builds a VcPay callback signature. Per VcPay docs, callback
// signing excludes code and msg in addition to sign and empty values.
func BuildCallbackSign(params map[string]string, appKey string) string {
	return buildSign(params, appKey, callbackSignExcludedKeys)
}

// VerifySign verifies a normal VcPay request/response signature.
func VerifySign(params map[string]string, appKey string) bool {
	return verifySign(params, appKey, nil)
}

// VerifyCallbackSign verifies a VcPay callback signature.
func VerifyCallbackSign(params map[string]string, appKey string) bool {
	return verifySign(params, appKey, callbackSignExcludedKeys)
}

// BuildSignAny is the map[string]any variant for JSON-decoded payloads.
func BuildSignAny(params map[string]any, appKey string) string {
	return buildSignAny(params, appKey, nil)
}

// BuildCallbackSignAny is the callback map[string]any variant for JSON-decoded payloads.
func BuildCallbackSignAny(params map[string]any, appKey string) string {
	return buildSignAny(params, appKey, callbackSignExcludedKeys)
}

// VerifySignAny verifies a normal VcPay signature from a JSON-decoded payload.
func VerifySignAny(params map[string]any, appKey string) bool {
	return verifySignAny(params, appKey, nil)
}

// VerifyCallbackSignAny verifies a callback signature from a JSON-decoded payload.
func VerifyCallbackSignAny(params map[string]any, appKey string) bool {
	return verifySignAny(params, appKey, callbackSignExcludedKeys)
}

func verifySign(params map[string]string, appKey string, excluded map[string]struct{}) bool {
	sign := strings.TrimSpace(params["sign"])
	if sign == "" {
		return false
	}
	return strings.EqualFold(sign, buildSign(params, appKey, excluded))
}

func verifySignAny(params map[string]any, appKey string, excluded map[string]struct{}) bool {
	sign := strings.TrimSpace(valueToString(params["sign"]))
	if sign == "" {
		return false
	}
	return strings.EqualFold(sign, buildSignAny(params, appKey, excluded))
}

func buildSign(params map[string]string, appKey string, excluded map[string]struct{}) string {
	normalized := make(map[string]string, len(params))
	for key, value := range params {
		normalized[key] = value
	}
	return signString(buildSignString(normalized, appKey, excluded))
}

func buildSignAny(params map[string]any, appKey string, excluded map[string]struct{}) string {
	normalized := make(map[string]string, len(params))
	for key, value := range params {
		normalized[key] = valueToString(value)
	}
	return signString(buildSignString(normalized, appKey, excluded))
}

func buildSignString(params map[string]string, appKey string, excluded map[string]struct{}) string {
	keys := make([]string, 0, len(params))
	for key, value := range params {
		if key == "sign" || strings.TrimSpace(value) == "" {
			continue
		}
		if _, ok := excluded[key]; ok {
			continue
		}
		keys = append(keys, key)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys)+1)
	for _, key := range keys {
		parts = append(parts, key+"="+params[key])
	}
	parts = append(parts, "key="+appKey)
	return strings.Join(parts, "&")
}

func signString(raw string) string {
	sum := md5.Sum([]byte(raw))
	return strings.ToUpper(hex.EncodeToString(sum[:]))
}

func valueToString(value any) string {
	if value == nil {
		return ""
	}
	switch v := value.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	case fmt.Stringer:
		return v.String()
	default:
		return fmt.Sprint(v)
	}
}
