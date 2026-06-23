package vcpaymxn

import (
	"BaseGoUni/core/base"
	"BaseGoUni/core/pay"
	"BaseGoUni/core/utils"
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	providerName           = "VCPAYMXN"
	defaultPayinTradeType  = "MX0001"
	defaultPayoutTradeType = "MX0001"
	defaultIdentityType    = "PHONE"
)

var mexicoMobileAreaPrefixes = []string{"55", "56", "33", "81", "22", "44", "66", "99"}

type Provider struct{}

func init() {
	pay.Register(&Provider{})
}

func (p *Provider) Name() string {
	return providerName
}

func (p *Provider) CreateOrder(req pay.PayRequest) (pay.PayResponse, error) {
	cfg := utils.GlobalConfig.Pay.Vcpaymxn
	if missing := missingConfigFields(cfg); len(missing) > 0 {
		return pay.PayResponse{}, fmt.Errorf("VCPAYMXN 配置不完整，core.yaml pay.vcpaymxn 缺少: %s", strings.Join(missing, ", "))
	}

	params := buildPayinParams(cfg, req, newNonce())
	params["sign"] = BuildSignAny(params, cfg.AppKey)

	respBody, err := postJSON(baseURL(cfg)+"/pay/save", params)
	if err != nil {
		return pay.PayResponse{}, fmt.Errorf("VCPAYMXN 代收请求失败: %w", err)
	}

	var apiResp tradeResp
	if err = json.Unmarshal(respBody, &apiResp); err != nil {
		return pay.PayResponse{}, fmt.Errorf("VCPAYMXN 代收响应解析失败: %w", err)
	}
	if apiResp.Code.String() != "200" {
		return pay.PayResponse{}, fmt.Errorf("VCPAYMXN 代收下单失败 code=%s msg=%s", apiResp.Code.String(), apiResp.Msg)
	}
	if strings.TrimSpace(apiResp.OutTradeNo) == "" && strings.TrimSpace(apiResp.TradeNo) == "" {
		return pay.PayResponse{}, fmt.Errorf("VCPAYMXN 代收下单失败: response order is empty")
	}

	return pay.PayResponse{
		PayURL:          apiResp.PayURL,
		ProviderTradeNo: apiResp.TradeNo,
	}, nil
}

func (p *Provider) CreatePayoutOrder(req pay.PayoutRequest) (pay.PayoutResponse, error) {
	cfg := utils.GlobalConfig.Pay.Vcpaymxn
	if missing := missingConfigFields(cfg); len(missing) > 0 {
		return pay.PayoutResponse{}, fmt.Errorf("VCPAYMXN 配置不完整，core.yaml pay.vcpaymxn 缺少: %s", strings.Join(missing, ", "))
	}
	if strings.TrimSpace(req.AccName) == "" {
		return pay.PayoutResponse{}, fmt.Errorf("VCPAYMXN 代付参数缺少 bank_owner")
	}
	if strings.TrimSpace(req.AccNo) == "" {
		return pay.PayoutResponse{}, fmt.Errorf("VCPAYMXN 代付参数缺少 bank_account")
	}
	if strings.TrimSpace(req.BankCode) == "" {
		return pay.PayoutResponse{}, fmt.Errorf("VCPAYMXN 代付参数缺少 bank_code")
	}

	params := buildPayoutParams(cfg, req, newNonce())
	params["sign"] = BuildSignAny(params, cfg.AppKey)

	respBody, err := postJSON(baseURL(cfg)+"/wd/save", params)
	if err != nil {
		return pay.PayoutResponse{}, fmt.Errorf("VCPAYMXN 代付请求失败: %w", err)
	}

	var apiResp tradeResp
	if err = json.Unmarshal(respBody, &apiResp); err != nil {
		return pay.PayoutResponse{}, fmt.Errorf("VCPAYMXN 代付响应解析失败: %w", err)
	}
	if apiResp.Code.String() != "200" {
		return pay.PayoutResponse{}, fmt.Errorf("VCPAYMXN 代付下单失败 code=%s msg=%s", apiResp.Code.String(), apiResp.Msg)
	}

	return pay.PayoutResponse{
		ProviderOrderNo: apiResp.TradeNo,
		Status:          normalizePayoutTradeState(apiResp.TradeState),
	}, nil
}

func (p *Provider) QueryPayoutOrder(req pay.PayoutQueryRequest) (pay.PayoutQueryResponse, error) {
	cfg := utils.GlobalConfig.Pay.Vcpaymxn
	if missing := missingConfigFields(cfg); len(missing) > 0 {
		return pay.PayoutQueryResponse{}, fmt.Errorf("VCPAYMXN 配置不完整，core.yaml pay.vcpaymxn 缺少: %s", strings.Join(missing, ", "))
	}
	if strings.TrimSpace(req.OrderNo) == "" {
		return pay.PayoutQueryResponse{}, fmt.Errorf("VCPAYMXN 代付查询参数缺少 out_trade_no")
	}

	params := buildPayoutQueryParams(cfg, req, newNonce())
	params["sign"] = BuildSignAny(params, cfg.AppKey)

	respBody, err := postJSON(baseURL(cfg)+"/wd/query", params)
	if err != nil {
		return pay.PayoutQueryResponse{}, fmt.Errorf("VCPAYMXN 代付查询请求失败: %w", err)
	}

	var apiResp tradeResp
	if err = json.Unmarshal(respBody, &apiResp); err != nil {
		return pay.PayoutQueryResponse{}, fmt.Errorf("VCPAYMXN 代付查询响应解析失败: %w", err)
	}
	if apiResp.Code.String() != "200" {
		return pay.PayoutQueryResponse{}, fmt.Errorf("VCPAYMXN 代付查询失败 code=%s msg=%s", apiResp.Code.String(), apiResp.Msg)
	}

	return pay.PayoutQueryResponse{
		ProviderOrderNo: apiResp.TradeNo,
		OrderNo:         apiResp.OutTradeNo,
		Status:          normalizePayoutTradeState(apiResp.TradeState),
		Amount:          centsToAmount(apiResp.OrderAmount.Int64()),
		SubMsg:          apiResp.Msg,
		SubCode:         apiResp.Code.String(),
	}, nil
}

func QueryBalance() (BalanceResponse, error) {
	cfg := utils.GlobalConfig.Pay.Vcpaymxn
	if missing := missingConfigFields(cfg); len(missing) > 0 {
		return BalanceResponse{}, fmt.Errorf("VCPAYMXN 配置不完整，core.yaml pay.vcpaymxn 缺少: %s", strings.Join(missing, ", "))
	}

	params := map[string]any{
		"app_id":    cfg.AppID,
		"nonce_str": newNonce(),
	}
	params["sign"] = BuildSignAny(params, cfg.AppKey)

	respBody, err := postJSON(baseURL(cfg)+"/pay/balance", params)
	if err != nil {
		return BalanceResponse{}, fmt.Errorf("VCPAYMXN 余额查询请求失败: %w", err)
	}

	var apiResp balanceResp
	if err = json.Unmarshal(respBody, &apiResp); err != nil {
		return BalanceResponse{}, fmt.Errorf("VCPAYMXN 余额查询响应解析失败: %w", err)
	}
	if apiResp.Code.String() != "200" {
		return BalanceResponse{}, fmt.Errorf("VCPAYMXN 余额查询失败 code=%s msg=%s", apiResp.Code.String(), apiResp.Msg)
	}

	return BalanceResponse{
		Balance:       centsToAmount(apiResp.Balance.Int64()),
		TotalBalance:  centsToAmount(apiResp.TotalBalance.Int64()),
		FreezeBalance: centsToAmount(apiResp.FreezeBalance.Int64()),
	}, nil
}

func buildPayinParams(cfg base.VcpayMxnPayConfig, req pay.PayRequest, nonce string) map[string]any {
	params := map[string]any{
		"app_id":       cfg.AppID,
		"nonce_str":    nonce,
		"trade_type":   "MX0001",
		"order_amount": amountToCents(pay.ResolveOrderAmount(req)),
		"out_trade_no": strings.TrimSpace(req.OrderNo),
		"notify_url":   resolvePayinNotifyURL(cfg, req),
		"back_url":     resolvePayinBackURL(cfg, req),
	}
	if identityName := resolveExtraField(req.ExtraFields, "identity_name", "identityName", "namemxn", "name", "payerName", "username"); identityName != "" {
		params["identity_name"] = identityName
	}
	if identityType := resolveExtraField(req.ExtraFields, "identity_type", "identityType"); identityType != "" {
		params["identity_type"] = identityType
	} else if identity := resolvePayinIdentity(req); identity != "" {
		params["identity_type"] = defaultIdentityType
		params["identity"] = identity
	}
	if _, ok := params["identity"]; !ok {
		if identity := resolvePayinIdentity(req); identity != "" {
			params["identity"] = identity
		}
	}
	return params
}

func buildPayoutParams(cfg base.VcpayMxnPayConfig, req pay.PayoutRequest, nonce string) map[string]any {
	params := map[string]any{
		"app_id":        cfg.AppID,
		"nonce_str":     nonce,
		"trade_type":    resolvePayoutTradeType(req),
		"order_amount":  amountToCents(pay.ResolvePayoutAmount(req)),
		"out_trade_no":  strings.TrimSpace(req.OrderNo),
		"notify_url":    resolvePayoutNotifyURL(cfg, req),
		"bank_code":     resolvePayoutBankCode(req.BankCode),
		"bank_owner":    strings.TrimSpace(req.AccName),
		"bank_account":  strings.TrimSpace(req.AccNo),
		"identity_type": defaultIdentityType,
		"identity":      randomMexicanMobilePhone(),
	}
	if bankAttach := resolveExtraField(req.ExtraFields, "bank_attach", "bankAttach"); bankAttach != "" {
		params["bank_attach"] = bankAttach
	}
	return params
}

func buildPayoutQueryParams(cfg base.VcpayMxnPayConfig, req pay.PayoutQueryRequest, nonce string) map[string]any {
	return map[string]any{
		"app_id":       cfg.AppID,
		"nonce_str":    nonce,
		"out_trade_no": strings.TrimSpace(req.OrderNo),
	}
}

func missingConfigFields(cfg base.VcpayMxnPayConfig) []string {
	missing := make([]string, 0, 3)
	if strings.TrimSpace(cfg.AppID) == "" {
		missing = append(missing, "appId")
	}
	if strings.TrimSpace(cfg.AppKey) == "" {
		missing = append(missing, "appKey")
	}
	if strings.TrimSpace(cfg.BaseURL) == "" {
		missing = append(missing, "baseUrl")
	}
	return missing
}

func baseURL(cfg base.VcpayMxnPayConfig) string {
	return strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/")
}

func resolvePayinTradeType(req pay.PayRequest) string {
	return firstNonEmpty(
		resolveExtraField(req.ExtraFields, "trade_type", "tradeType", "busiCode"),
		strings.TrimSpace(req.PayMethod),
		defaultPayinTradeType,
	)
}

func resolvePayoutTradeType(req pay.PayoutRequest) string {
	return firstNonEmpty(
		resolveExtraField(req.ExtraFields, "trade_type", "tradeType"),
		strings.TrimSpace(req.BusiCode),
		resolvePayoutTradeTypeFromIdentityType(req),
		defaultPayoutTradeType,
	)
}

func resolvePayoutTradeTypeFromIdentityType(req pay.PayoutRequest) string {
	identityType := firstNonEmpty(
		strings.TrimSpace(req.IdentityType),
		resolveExtraField(req.ExtraFields, "identityTypeMXNW", "identity_type_mxnw", "identityType", "identity_type"),
	)
	switch strings.ToUpper(strings.TrimSpace(identityType)) {
	case "CLABE", "SPEI":
		return "MX0001"
	case "BANK", "BANK_ACCOUNT", "BANKACCOUNT":
		return "MX0000"
	default:
		return ""
	}
}

func resolvePayinNotifyURL(cfg base.VcpayMxnPayConfig, _ pay.PayRequest) string {
	return strings.TrimSpace(cfg.NotifyURL)
}

func resolvePayinBackURL(cfg base.VcpayMxnPayConfig, _ pay.PayRequest) string {
	return strings.TrimSpace(cfg.PageURL)
}

func resolvePayoutNotifyURL(cfg base.VcpayMxnPayConfig, _ pay.PayoutRequest) string {
	return firstNonEmpty(strings.TrimSpace(cfg.PayoutNotifyURL), strings.TrimSpace(cfg.NotifyURL))
}

func resolvePayinIdentity(req pay.PayRequest) string {
	return resolveExtraField(req.ExtraFields, "identity", "phonemxn", "phoneMxn", "phoneMXN", "phone", "mobile", "mobilePhone")
}

func resolveExtraField(fields map[string]string, aliases ...string) string {
	for _, alias := range aliases {
		if value := strings.TrimSpace(fields[alias]); value != "" {
			return value
		}
	}
	normalized := make(map[string]string, len(fields))
	for key, value := range fields {
		normalizedKey := normalizeExtraFieldKey(key)
		if normalizedKey == "" || strings.TrimSpace(value) == "" {
			continue
		}
		if _, exists := normalized[normalizedKey]; !exists {
			normalized[normalizedKey] = strings.TrimSpace(value)
		}
	}
	for _, alias := range aliases {
		if value := normalized[normalizeExtraFieldKey(alias)]; value != "" {
			return value
		}
	}
	return ""
}

func normalizeExtraFieldKey(key string) string {
	var builder strings.Builder
	for _, r := range strings.ToLower(strings.TrimSpace(key)) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			builder.WriteRune(r)
		}
	}
	return builder.String()
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func randomMexicanMobilePhone() string {
	prefix := mexicoMobileAreaPrefixes[randomIndex(len(mexicoMobileAreaPrefixes))]
	return prefix + randomDigits(8)
}

func randomDigits(length int) string {
	if length <= 0 {
		return ""
	}
	buf := make([]byte, length)
	if _, err := rand.Read(buf); err != nil {
		seed := time.Now().UnixNano()
		for i := range buf {
			buf[i] = byte((seed + int64(i)) % 10)
		}
	}
	for i := range buf {
		buf[i] = '0' + (buf[i] % 10)
	}
	return string(buf)
}

func randomIndex(length int) int {
	if length <= 1 {
		return 0
	}
	var b [1]byte
	if _, err := rand.Read(b[:]); err != nil {
		return int(time.Now().UnixNano() % int64(length))
	}
	return int(b[0]) % length
}

func amountToCents(amount float64) int64 {
	if amount <= 0 || math.IsNaN(amount) || math.IsInf(amount, 0) {
		return 0
	}
	return int64(math.Round(amount * 100))
}

func normalizePayoutTradeState(state int) int {
	switch state {
	case 1:
		return 7
	default:
		return state
	}
}

func centsToAmount(cents int64) float64 {
	return float64(cents) / 100
}

func newNonce() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err == nil {
		return hex.EncodeToString(b[:])
	}
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func postJSON(url string, payload map[string]any) ([]byte, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		log.Printf("[%s] third-party request marshal failed url=%s err=%v", providerName, url, err)
		return nil, err
	}
	log.Printf("[%s] third-party request url=%s params=%s", providerName, url, string(body))

	resp, err := http.Post(url, "application/json", bytes.NewReader(body)) //nolint:noctx
	if err != nil {
		log.Printf("[%s] third-party request failed url=%s err=%v", providerName, url, err)
		return nil, err
	}
	defer resp.Body.Close()

	respBody, readErr := io.ReadAll(resp.Body)
	log.Printf("[%s] third-party response url=%s status=%d body=%s", providerName, url, resp.StatusCode, string(respBody))
	if readErr != nil {
		log.Printf("[%s] third-party response read failed url=%s status=%d err=%v", providerName, url, resp.StatusCode, readErr)
		return respBody, readErr
	}
	return respBody, nil
}

type tradeResp struct {
	Code        flexString `json:"code"`
	Msg         string     `json:"msg"`
	AppID       string     `json:"app_id"`
	NonceStr    string     `json:"nonce_str"`
	NotifyURL   string     `json:"notify_url"`
	OrderAmount flexInt    `json:"order_amount"`
	OrderFee    flexInt    `json:"order_fee"`
	OutTradeNo  string     `json:"out_trade_no"`
	PayURL      string     `json:"pay_url"`
	TradeNo     string     `json:"trade_no"`
	TradeState  int        `json:"trade_state"`
	TradeType   string     `json:"trade_type"`
	Sign        string     `json:"sign"`
}

type BalanceResponse struct {
	Balance       float64
	TotalBalance  float64
	FreezeBalance float64
}

type balanceResp struct {
	Code          flexString `json:"code"`
	Msg           string     `json:"msg"`
	AppID         string     `json:"app_id"`
	NonceStr      string     `json:"nonce_str"`
	Balance       flexInt    `json:"balance"`
	TotalBalance  flexInt    `json:"total_balance"`
	FreezeBalance flexInt    `json:"freeze_balance"`
	Sign          string     `json:"sign"`
}

type flexInt int64

func (v *flexInt) UnmarshalJSON(data []byte) error {
	data = bytes.TrimSpace(data)
	if bytes.Equal(data, []byte("null")) || len(data) == 0 {
		*v = 0
		return nil
	}
	var n int64
	if err := json.Unmarshal(data, &n); err == nil {
		*v = flexInt(n)
		return nil
	}
	var text string
	if err := json.Unmarshal(data, &text); err == nil {
		parsed, _ := strconv.ParseInt(strings.TrimSpace(text), 10, 64)
		*v = flexInt(parsed)
		return nil
	}
	return nil
}

func (v flexInt) Int64() int64 {
	return int64(v)
}

type flexString string

func (v *flexString) UnmarshalJSON(data []byte) error {
	data = bytes.TrimSpace(data)
	if bytes.Equal(data, []byte("null")) {
		*v = ""
		return nil
	}
	var text string
	if err := json.Unmarshal(data, &text); err == nil {
		*v = flexString(text)
		return nil
	}
	*v = flexString(string(data))
	return nil
}

func (v flexString) String() string {
	return strings.TrimSpace(string(v))
}

func (v flexString) Int() int {
	n, _ := strconv.Atoi(v.String())
	return n
}
