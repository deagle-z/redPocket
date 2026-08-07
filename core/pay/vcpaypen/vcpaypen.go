package vcpaypen

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
	providerName = "VCPAYPEN"
	// defaultPayinTradeType 秘鲁代收默认走 Yape（使用率最高）
	defaultPayinTradeType = "PEW001"
	// defaultPayoutTradeType 秘鲁银行转账代付固定值
	defaultPayoutTradeType = "PEN000"
	// defaultIdentityType 秘鲁证件类型缺省按身份证处理
	defaultIdentityType = "DNI"
	// defaultAccountType 缺省储蓄账户
	defaultAccountType = "SA"
)

// payinTradeTypes 代收 TradeType 白名单，见 readme.md「代收 TradeType」
var payinTradeTypes = map[string]struct{}{
	"PEW001": {}, // Yape
	"PEW002": {}, // Plin
	"PET000": {}, // Bank Transfer
	"PEP001": {}, // PagoEfectivo
}

// identityTypes 秘鲁证件类型白名单，见 readme.md 代收/代付请求体
var identityTypes = map[string]struct{}{
	"DNI": {}, // 身份证
	"PAS": {}, // 护照
	"CE":  {}, // 外国人居留许可证
	"RUC": {}, // 税号
}

type Provider struct{}

func init() {
	pay.Register(&Provider{})
}

func (p *Provider) Name() string {
	return providerName
}

func (p *Provider) CreateOrder(req pay.PayRequest) (pay.PayResponse, error) {
	cfg := utils.GlobalConfig.Pay.Vcpaypen
	if missing := missingConfigFields(cfg); len(missing) > 0 {
		return pay.PayResponse{}, fmt.Errorf("VCPAYPEN 配置不完整，core.yaml pay.vcpaypen 缺少: %s", strings.Join(missing, ", "))
	}

	params, err := buildPayinParams(cfg, req, newNonce())
	if err != nil {
		return pay.PayResponse{}, err
	}
	params["sign"] = BuildSignAny(params, cfg.AppKey)

	respBody, err := postJSON(baseURL(cfg)+"/pay/save", params)
	if err != nil {
		return pay.PayResponse{}, fmt.Errorf("VCPAYPEN 代收请求失败: %w", err)
	}

	var apiResp tradeResp
	if err = json.Unmarshal(respBody, &apiResp); err != nil {
		return pay.PayResponse{}, fmt.Errorf("VCPAYPEN 代收响应解析失败: %w", err)
	}
	if apiResp.Code.String() != "200" {
		return pay.PayResponse{}, fmt.Errorf("VCPAYPEN 代收下单失败 code=%s msg=%s", apiResp.Code.String(), apiResp.Msg)
	}
	if strings.TrimSpace(apiResp.OutTradeNo) == "" && strings.TrimSpace(apiResp.TradeNo) == "" {
		return pay.PayResponse{}, fmt.Errorf("VCPAYPEN 代收下单失败: response order is empty")
	}

	return pay.PayResponse{
		PayURL:          apiResp.PayURL,
		ProviderTradeNo: apiResp.TradeNo,
	}, nil
}

func (p *Provider) CreatePayoutOrder(req pay.PayoutRequest) (pay.PayoutResponse, error) {
	cfg := utils.GlobalConfig.Pay.Vcpaypen
	if missing := missingConfigFields(cfg); len(missing) > 0 {
		return pay.PayoutResponse{}, fmt.Errorf("VCPAYPEN 配置不完整，core.yaml pay.vcpaypen 缺少: %s", strings.Join(missing, ", "))
	}
	if strings.TrimSpace(req.AccName) == "" {
		return pay.PayoutResponse{}, fmt.Errorf("VCPAYPEN 代付参数缺少 bank_owner")
	}
	if strings.TrimSpace(req.AccNo) == "" {
		return pay.PayoutResponse{}, fmt.Errorf("VCPAYPEN 代付参数缺少 bank_account")
	}
	if strings.TrimSpace(req.BankCode) == "" {
		return pay.PayoutResponse{}, fmt.Errorf("VCPAYPEN 代付参数缺少 bank_code")
	}

	params, err := buildPayoutParams(cfg, req, newNonce())
	if err != nil {
		return pay.PayoutResponse{}, err
	}
	params["sign"] = BuildSignAny(params, cfg.AppKey)

	respBody, err := postJSON(baseURL(cfg)+"/wd/save", params)
	if err != nil {
		return pay.PayoutResponse{}, fmt.Errorf("VCPAYPEN 代付请求失败: %w", err)
	}

	var apiResp tradeResp
	if err = json.Unmarshal(respBody, &apiResp); err != nil {
		return pay.PayoutResponse{}, fmt.Errorf("VCPAYPEN 代付响应解析失败: %w", err)
	}
	if apiResp.Code.String() != "200" {
		return pay.PayoutResponse{}, fmt.Errorf("VCPAYPEN 代付下单失败 code=%s msg=%s", apiResp.Code.String(), apiResp.Msg)
	}

	return pay.PayoutResponse{
		ProviderOrderNo: apiResp.TradeNo,
		Status:          normalizePayoutTradeState(apiResp.TradeState),
	}, nil
}

func (p *Provider) QueryPayoutOrder(req pay.PayoutQueryRequest) (pay.PayoutQueryResponse, error) {
	cfg := utils.GlobalConfig.Pay.Vcpaypen
	if missing := missingConfigFields(cfg); len(missing) > 0 {
		return pay.PayoutQueryResponse{}, fmt.Errorf("VCPAYPEN 配置不完整，core.yaml pay.vcpaypen 缺少: %s", strings.Join(missing, ", "))
	}
	if strings.TrimSpace(req.OrderNo) == "" {
		return pay.PayoutQueryResponse{}, fmt.Errorf("VCPAYPEN 代付查询参数缺少 out_trade_no")
	}

	params := buildPayoutQueryParams(cfg, req, newNonce())
	params["sign"] = BuildSignAny(params, cfg.AppKey)

	respBody, err := postJSON(baseURL(cfg)+"/wd/query", params)
	if err != nil {
		return pay.PayoutQueryResponse{}, fmt.Errorf("VCPAYPEN 代付查询请求失败: %w", err)
	}

	var apiResp tradeResp
	if err = json.Unmarshal(respBody, &apiResp); err != nil {
		return pay.PayoutQueryResponse{}, fmt.Errorf("VCPAYPEN 代付查询响应解析失败: %w", err)
	}
	if apiResp.Code.String() != "200" {
		return pay.PayoutQueryResponse{}, fmt.Errorf("VCPAYPEN 代付查询失败 code=%s msg=%s", apiResp.Code.String(), apiResp.Msg)
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

// PayinQueryResult 代收查单结果
type PayinQueryResult struct {
	ProviderOrderNo string  // 平台单号 trade_no
	OrderNo         string  // 商户单号 out_trade_no
	TradeState      int     // 0 等待付款 / 1 支付成功 / 2 支付失败 / 3 支付中
	TradeType       string  // 代收 TradeType
	Amount          float64 // 订单金额（元）
	Fee             float64 // 手续费（元）
	SubCode         string
	SubMsg          string
}

// QueryPayinOrder 查询代收订单（POST /pay/query）。
// pay 包没有代收查单接口，故与 QueryBalance 一样做成包级导出函数。
// 注意：三方限频 10 秒一次。
func QueryPayinOrder(orderNo string) (PayinQueryResult, error) {
	cfg := utils.GlobalConfig.Pay.Vcpaypen
	if missing := missingConfigFields(cfg); len(missing) > 0 {
		return PayinQueryResult{}, fmt.Errorf("VCPAYPEN 配置不完整，core.yaml pay.vcpaypen 缺少: %s", strings.Join(missing, ", "))
	}
	if strings.TrimSpace(orderNo) == "" {
		return PayinQueryResult{}, fmt.Errorf("VCPAYPEN 代收查询参数缺少 out_trade_no")
	}

	params := map[string]any{
		"app_id":       cfg.AppID,
		"nonce_str":    newNonce(),
		"out_trade_no": strings.TrimSpace(orderNo),
	}
	params["sign"] = BuildSignAny(params, cfg.AppKey)

	respBody, err := postJSON(baseURL(cfg)+"/pay/query", params)
	if err != nil {
		return PayinQueryResult{}, fmt.Errorf("VCPAYPEN 代收查询请求失败: %w", err)
	}

	var apiResp tradeResp
	if err = json.Unmarshal(respBody, &apiResp); err != nil {
		return PayinQueryResult{}, fmt.Errorf("VCPAYPEN 代收查询响应解析失败: %w", err)
	}
	if apiResp.Code.String() != "200" {
		return PayinQueryResult{}, fmt.Errorf("VCPAYPEN 代收查询失败 code=%s msg=%s", apiResp.Code.String(), apiResp.Msg)
	}

	return PayinQueryResult{
		ProviderOrderNo: apiResp.TradeNo,
		OrderNo:         apiResp.OutTradeNo,
		TradeState:      apiResp.TradeState,
		TradeType:       apiResp.TradeType,
		Amount:          centsToAmount(apiResp.OrderAmount.Int64()),
		Fee:             centsToAmount(apiResp.OrderFee.Int64()),
		SubCode:         apiResp.Code.String(),
		SubMsg:          apiResp.Msg,
	}, nil
}

// QueryBalance 查询商户余额（POST /pay/balance）。注意：三方限频 10 秒一次。
func QueryBalance() (BalanceResponse, error) {
	cfg := utils.GlobalConfig.Pay.Vcpaypen
	if missing := missingConfigFields(cfg); len(missing) > 0 {
		return BalanceResponse{}, fmt.Errorf("VCPAYPEN 配置不完整，core.yaml pay.vcpaypen 缺少: %s", strings.Join(missing, ", "))
	}

	params := map[string]any{
		"app_id":    cfg.AppID,
		"nonce_str": newNonce(),
	}
	params["sign"] = BuildSignAny(params, cfg.AppKey)

	respBody, err := postJSON(baseURL(cfg)+"/pay/balance", params)
	if err != nil {
		return BalanceResponse{}, fmt.Errorf("VCPAYPEN 余额查询请求失败: %w", err)
	}

	var apiResp balanceResp
	if err = json.Unmarshal(respBody, &apiResp); err != nil {
		return BalanceResponse{}, fmt.Errorf("VCPAYPEN 余额查询响应解析失败: %w", err)
	}
	if apiResp.Code.String() != "200" {
		return BalanceResponse{}, fmt.Errorf("VCPAYPEN 余额查询失败 code=%s msg=%s", apiResp.Code.String(), apiResp.Msg)
	}

	return BalanceResponse{
		Balance:       centsToAmount(apiResp.Balance.Int64()),
		TotalBalance:  centsToAmount(apiResp.TotalBalance.Int64()),
		FreezeBalance: centsToAmount(apiResp.FreezeBalance.Int64()),
	}, nil
}

func buildPayinParams(cfg base.VcpayPenPayConfig, req pay.PayRequest, nonce string) (map[string]any, error) {
	// 秘鲁代收 back_url 必填（string[12,128]）
	backURL := resolvePayinBackURL(cfg, req)
	if backURL == "" {
		return nil, fmt.Errorf("VCPAYPEN 代收参数缺少 back_url，请配置 core.yaml pay.vcpaypen.pageUrl")
	}
	// 秘鲁代收 identity / identity_type 必填，是付款人真实证件号
	identity := resolvePayinIdentity(req)
	if identity == "" {
		return nil, fmt.Errorf("VCPAYPEN 代收参数缺少 identity（付款人证件号）")
	}
	identityType := normalizeIdentityType(resolveExtraField(req.ExtraFields,
		"identity_type", "identityType", "documentType", "docType", "tipoDocumento"))
	if identityType == "" {
		identityType = guessIdentityType(identity)
	}

	params := map[string]any{
		"app_id":        cfg.AppID,
		"nonce_str":     nonce,
		"trade_type":    resolvePayinTradeType(req),
		"order_amount":  amountToCents(pay.ResolveOrderAmount(req)),
		"out_trade_no":  strings.TrimSpace(req.OrderNo),
		"notify_url":    resolvePayinNotifyURL(cfg, req),
		"back_url":      backURL,
		"identity_type": identityType,
		"identity":      identity,
	}
	if identityName := resolveExtraField(req.ExtraFields,
		"identity_name", "identityName", "namemxn", "name", "payerName", "username"); identityName != "" {
		params["identity_name"] = identityName
	}
	return params, nil
}

func buildPayoutParams(cfg base.VcpayPenPayConfig, req pay.PayoutRequest, nonce string) (map[string]any, error) {
	bankCode := resolvePayoutBankCode(req.BankCode)
	bankAccount := strings.TrimSpace(req.AccNo)
	if err := validatePayoutBankAccount(bankCode, bankAccount); err != nil {
		return nil, err
	}

	identity := firstNonEmpty(
		strings.TrimSpace(req.IdentityNo),
		resolveExtraField(req.ExtraFields, "identity", "documentmxn", "document", "dni", "idNumber", "identityNo"),
	)
	if identity == "" {
		return nil, fmt.Errorf("VCPAYPEN 代付参数缺少 identity（收款人证件号）")
	}
	identityType := normalizeIdentityType(firstNonEmpty(
		req.IdentityType,
		resolveExtraField(req.ExtraFields, "identity_type", "identityType", "documentType", "docType", "tipoDocumento"),
	))
	if identityType == "" {
		identityType = guessIdentityType(identity)
	}

	params := map[string]any{
		"app_id":            cfg.AppID,
		"nonce_str":         nonce,
		"trade_type":        resolvePayoutTradeType(req, bankCode),
		"order_amount":      amountToCents(pay.ResolvePayoutAmount(req)),
		"out_trade_no":      strings.TrimSpace(req.OrderNo),
		"notify_url":        resolvePayoutNotifyURL(cfg, req),
		"bank_code":         bankCode,
		"bank_owner":        strings.TrimSpace(req.AccName),
		"bank_account":      bankAccount,
		"bank_account_type": resolvePayoutAccountType(req),
		"identity_type":     identityType,
		"identity":          identity,
	}
	if bankAttach := resolveExtraField(req.ExtraFields, "bank_attach", "bankAttach"); bankAttach != "" {
		params["bank_attach"] = bankAttach
	}
	return params, nil
}

func buildPayoutQueryParams(cfg base.VcpayPenPayConfig, req pay.PayoutQueryRequest, nonce string) map[string]any {
	return map[string]any{
		"app_id":       cfg.AppID,
		"nonce_str":    nonce,
		"out_trade_no": strings.TrimSpace(req.OrderNo),
	}
}

// validatePayoutBankAccount 电子钱包填 9 开头 9 位手机号，银行填 20 位 CCI 账号。
func validatePayoutBankAccount(bankCode, bankAccount string) error {
	if isWalletBankCode(bankCode) {
		if len(bankAccount) != 9 || bankAccount[0] != '9' || !isAllDigits(bankAccount) {
			return fmt.Errorf("VCPAYPEN 代付电子钱包账号需为 9 开头的 9 位手机号: %s", bankAccount)
		}
		return nil
	}
	if len(bankAccount) != 20 || !isAllDigits(bankAccount) {
		return fmt.Errorf("VCPAYPEN 代付银行账号需为 20 位 CCI 数字账号: %s", bankAccount)
	}
	return nil
}

func isAllDigits(s string) bool {
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

func missingConfigFields(cfg base.VcpayPenPayConfig) []string {
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

func baseURL(cfg base.VcpayPenPayConfig) string {
	return strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/")
}

// resolvePayinTradeType 只接受秘鲁代收 TradeType 白名单内的值，
// 避免把 pix / 通用支付方式名原样发给三方（三方会返回 40104）。
func resolvePayinTradeType(req pay.PayRequest) string {
	candidates := []string{
		resolveExtraField(req.ExtraFields, "trade_type", "tradeType", "busiCode"),
		strings.TrimSpace(req.PayMethod),
	}
	for _, candidate := range candidates {
		code := strings.ToUpper(strings.TrimSpace(candidate))
		if _, ok := payinTradeTypes[code]; ok {
			return code
		}
	}
	return defaultPayinTradeType
}

// resolvePayoutTradeType 电子钱包 trade_type 与 bank_code 同值，银行转账固定 PEN000。
func resolvePayoutTradeType(req pay.PayoutRequest, bankCode string) string {
	candidates := []string{
		resolveExtraField(req.ExtraFields, "trade_type", "tradeType"),
		strings.TrimSpace(req.BusiCode),
	}
	for _, candidate := range candidates {
		switch code := strings.ToUpper(strings.TrimSpace(candidate)); code {
		case "PEW001", "PEW002", "PEN000":
			return code
		}
	}
	if isWalletBankCode(bankCode) {
		return strings.ToUpper(strings.TrimSpace(bankCode))
	}
	return defaultPayoutTradeType
}

// resolvePayoutAccountType 归一 bank_account_type：储蓄 SA / 活期 CA。
func resolvePayoutAccountType(req pay.PayoutRequest) string {
	raw := firstNonEmpty(
		req.AccountType,
		resolveExtraField(req.ExtraFields,
			"bank_account_type", "bankAccountType", "accountType", "accType",
			"bankAccountTypeMXNW", "accountTypeMXNW", "accountTypeMXN", "accountTypeMX"),
	)
	switch code := normalizeBankCodeKey(raw); code {
	case "SA", "SAVINGS", "SAVING", "AHORRO", "AHORROS", "CUENTA_DE_AHORROS":
		return "SA"
	case "CA", "CHECKING", "CURRENT", "CORRIENTE", "CUENTA_CORRIENTE":
		return "CA"
	default:
		return defaultAccountType
	}
}

// normalizeIdentityType 归一秘鲁证件类型，无法识别时返回空串。
func normalizeIdentityType(identityType string) string {
	code := normalizeBankCodeKey(identityType)
	if code == "" {
		return ""
	}
	if _, ok := identityTypes[code]; ok {
		return code
	}
	switch code {
	case "DOCUMENTO", "DOCUMENT", "CEDULA", "ID", "IDCARD", "ID_CARD":
		return "DNI"
	case "PASSPORT", "PASAPORTE", "PASS":
		return "PAS"
	case "CARNET", "CARNET_DE_EXTRANJERIA", "CARNE_DE_EXTRANJERIA", "FOREIGNER", "CEX":
		return "CE"
	case "TAX", "TAX_ID", "TAXID", "RUT":
		return "RUC"
	default:
		return ""
	}
}

// guessIdentityType 未显式指定证件类型时，按号码长度推断：8 位 DNI、11 位 RUC，其余兜底 DNI。
func guessIdentityType(identity string) string {
	if isAllDigits(identity) {
		switch len(identity) {
		case 8:
			return "DNI"
		case 11:
			return "RUC"
		}
	}
	return defaultIdentityType
}

func resolvePayinNotifyURL(cfg base.VcpayPenPayConfig, _ pay.PayRequest) string {
	return strings.TrimSpace(cfg.NotifyURL)
}

func resolvePayinBackURL(cfg base.VcpayPenPayConfig, req pay.PayRequest) string {
	return firstNonEmpty(strings.TrimSpace(req.ReturnURL), strings.TrimSpace(cfg.PageURL))
}

func resolvePayoutNotifyURL(cfg base.VcpayPenPayConfig, _ pay.PayoutRequest) string {
	return firstNonEmpty(strings.TrimSpace(cfg.PayoutNotifyURL), strings.TrimSpace(cfg.NotifyURL))
}

func resolvePayinIdentity(req pay.PayRequest) string {
	return resolveExtraField(req.ExtraFields,
		"identity", "documentmxn", "document", "dni", "idNumber", "identityNo", "documentNumber")
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
