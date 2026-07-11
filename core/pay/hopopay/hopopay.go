package hopopay

import (
	"BaseGoUni/core/base"
	"BaseGoUni/core/pay"
	"BaseGoUni/core/utils"
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	providerName    = "HOPOPAY"
	fixedPlatform   = "android"
	payTypeCashapp  = "cashapp"
	payTypePaypal   = "paypal"
	payTypeApplepay = "applepay"
	payTypeGoogle   = "googlepay"
)

var allowedPayTypes = map[string]struct{}{
	payTypeCashapp:  {},
	payTypePaypal:   {},
	payTypeApplepay: {},
	payTypeGoogle:   {},
}

type Provider struct{}

func init() {
	pay.Register(&Provider{})
}

func (p *Provider) Name() string {
	return providerName
}

func (p *Provider) CreateOrder(req pay.PayRequest) (pay.PayResponse, error) {
	cfg := utils.GlobalConfig.Pay.Hopopay
	if missing := missingPayinConfigFields(cfg); len(missing) > 0 {
		return pay.PayResponse{}, fmt.Errorf("HOPOPAY 配置不完整，core.yaml pay.hopopay 缺少: %s", strings.Join(missing, ", "))
	}

	params, err := buildPayinParams(cfg, req, time.Now().Unix(), newNonce())
	if err != nil {
		return pay.PayResponse{}, err
	}
	params["signature"] = BuildSignature(params, cfg.Secret)

	respBody, err := postJSON(baseURL(cfg)+"/api/payment/payin", params)
	if err != nil {
		return pay.PayResponse{}, fmt.Errorf("HOPOPAY 代收请求失败: %w", err)
	}

	var apiResp payinOrderResp
	if err = json.Unmarshal(respBody, &apiResp); err != nil {
		return pay.PayResponse{}, fmt.Errorf("HOPOPAY 代收响应解析失败: %w", err)
	}
	if apiResp.Status != 1 {
		return pay.PayResponse{}, fmt.Errorf("HOPOPAY 代收下单失败 status=%d msg=%s", apiResp.Status, apiErrorMessage(apiResp.Info, apiResp.Error))
	}
	if apiResp.Data == nil {
		return pay.PayResponse{}, errors.New("HOPOPAY 代收下单失败: data is empty")
	}
	if apiResp.Data.Status != 3 {
		return pay.PayResponse{}, fmt.Errorf("HOPOPAY 代收下单失败 orderStatus=%d appOrderID=%s", apiResp.Data.Status, apiResp.Data.AppOrderID)
	}

	payURL := strings.TrimSpace(apiResp.Data.URL)
	if payURL == "" {
		payURL = strings.TrimSpace(apiResp.Data.QRCode)
	}
	return pay.PayResponse{
		PayURL:          payURL,
		ProviderTradeNo: strings.TrimSpace(apiResp.Data.OrderNo),
	}, nil
}

func (p *Provider) CreatePayoutOrder(req pay.PayoutRequest) (pay.PayoutResponse, error) {
	cfg := utils.GlobalConfig.Pay.Hopopay
	if missing := missingPayoutConfigFields(cfg); len(missing) > 0 {
		return pay.PayoutResponse{}, fmt.Errorf("HOPOPAY 代付配置不完整，core.yaml pay.hopopay 缺少: %s", strings.Join(missing, ", "))
	}

	params, err := buildPayoutParams(cfg, req, time.Now().Unix(), newNonce())
	if err != nil {
		return pay.PayoutResponse{}, err
	}
	params["signature"] = BuildSignature(params, cfg.Secret)

	respBody, err := postJSON(baseURL(cfg)+"/api/payment/payout", params)
	if err != nil {
		return pay.PayoutResponse{}, fmt.Errorf("HOPOPAY 代付请求失败: %w", err)
	}

	var apiResp payoutOrderResp
	if err = json.Unmarshal(respBody, &apiResp); err != nil {
		return pay.PayoutResponse{}, fmt.Errorf("HOPOPAY 代付响应解析失败: %w", err)
	}
	if apiResp.Status != 1 {
		return pay.PayoutResponse{}, fmt.Errorf("HOPOPAY 代付下单失败 status=%d msg=%s", apiResp.Status, apiErrorMessage(apiResp.Info, apiResp.Error))
	}
	if apiResp.Data == nil {
		return pay.PayoutResponse{}, errors.New("HOPOPAY 代付下单失败: data is empty")
	}

	return pay.PayoutResponse{
		ProviderOrderNo: strings.TrimSpace(apiResp.Data.OrderNo),
		Status:          apiResp.Data.Status,
	}, nil
}

func (p *Provider) QueryPayoutOrder(req pay.PayoutQueryRequest) (pay.PayoutQueryResponse, error) {
	result, err := QueryPayoutOrder(req)
	if err != nil {
		return pay.PayoutQueryResponse{}, err
	}
	return pay.PayoutQueryResponse{
		ProviderOrderNo: result.OrderNo,
		OrderNo:         result.AppOrderID,
		Status:          normalizePayoutQueryStatus(result.Status),
		Amount:          result.Amount.Float64(),
		SubMsg:          result.Message,
		SubCode:         fmt.Sprintf("%d", result.Status),
		PayTime:         formatUnixSeconds(result.PayAt.Int64()),
	}, nil
}

type PayinQueryRequest struct {
	OrderNo         string
	ProviderOrderNo string
}

type PayinQueryResponse struct {
	OrderNo    string    `json:"order_no"`
	AppOrderID string    `json:"app_order_id"`
	Status     int       `json:"status"`
	Amount     flexFloat `json:"amount"`
	PayAmount  flexFloat `json:"pay_amount"`
	NetAmount  flexFloat `json:"net_amount"`
	Fee        flexFloat `json:"fee"`
	CreateAt   flexInt   `json:"create_at"`
	PayAt      flexInt   `json:"pay_at"`
	Time       flexInt   `json:"time"`
	Message    string    `json:"-"`
}

func QueryPayinOrder(req PayinQueryRequest) (PayinQueryResponse, error) {
	cfg := utils.GlobalConfig.Pay.Hopopay
	if missing := missingBaseConfigFields(cfg); len(missing) > 0 {
		return PayinQueryResponse{}, fmt.Errorf("HOPOPAY 代收查询配置不完整，core.yaml pay.hopopay 缺少: %s", strings.Join(missing, ", "))
	}
	params, err := buildQueryParams(cfg, req.OrderNo, req.ProviderOrderNo, time.Now().Unix(), newNonce())
	if err != nil {
		return PayinQueryResponse{}, err
	}
	params["signature"] = BuildSignature(params, cfg.Secret)

	respBody, err := postJSON(baseURL(cfg)+"/api/payin/order/query", params)
	if err != nil {
		return PayinQueryResponse{}, fmt.Errorf("HOPOPAY 代收查询请求失败: %w", err)
	}
	var apiResp payinQueryResp
	if err = json.Unmarshal(respBody, &apiResp); err != nil {
		return PayinQueryResponse{}, fmt.Errorf("HOPOPAY 代收查询响应解析失败: %w", err)
	}
	if apiResp.Status != 1 {
		return PayinQueryResponse{}, fmt.Errorf("HOPOPAY 代收查询失败 status=%d msg=%s", apiResp.Status, apiErrorMessage(apiResp.Info, apiResp.Error))
	}
	if apiResp.Data == nil {
		return PayinQueryResponse{}, errors.New("HOPOPAY 代收查询失败: data is empty")
	}
	apiResp.Data.Message = apiResp.Info
	return *apiResp.Data, nil
}

type PayoutQueryResponse struct {
	OrderNo    string    `json:"order_no"`
	AppOrderID string    `json:"app_order_id"`
	Status     int       `json:"status"`
	Amount     flexFloat `json:"amount"`
	NetAmount  flexFloat `json:"net_amount"`
	Fee        flexFloat `json:"fee"`
	CreateAt   flexInt   `json:"create_at"`
	PayAt      flexInt   `json:"pay_at"`
	Time       flexInt   `json:"time"`
	Message    string    `json:"-"`
}

func QueryPayoutOrder(req pay.PayoutQueryRequest) (PayoutQueryResponse, error) {
	cfg := utils.GlobalConfig.Pay.Hopopay
	if missing := missingBaseConfigFields(cfg); len(missing) > 0 {
		return PayoutQueryResponse{}, fmt.Errorf("HOPOPAY 代付查询配置不完整，core.yaml pay.hopopay 缺少: %s", strings.Join(missing, ", "))
	}
	params, err := buildQueryParams(cfg, req.OrderNo, req.ProviderOrderNo, time.Now().Unix(), newNonce())
	if err != nil {
		return PayoutQueryResponse{}, err
	}
	params["signature"] = BuildSignature(params, cfg.Secret)

	respBody, err := postJSON(baseURL(cfg)+"/api/payout/order/query", params)
	if err != nil {
		return PayoutQueryResponse{}, fmt.Errorf("HOPOPAY 代付查询请求失败: %w", err)
	}
	var apiResp payoutQueryResp
	if err = json.Unmarshal(respBody, &apiResp); err != nil {
		return PayoutQueryResponse{}, fmt.Errorf("HOPOPAY 代付查询响应解析失败: %w", err)
	}
	if apiResp.Status != 1 {
		return PayoutQueryResponse{}, fmt.Errorf("HOPOPAY 代付查询失败 status=%d msg=%s", apiResp.Status, apiErrorMessage(apiResp.Info, apiResp.Error))
	}
	if apiResp.Data == nil {
		return PayoutQueryResponse{}, errors.New("HOPOPAY 代付查询失败: data is empty")
	}
	apiResp.Data.Message = apiResp.Info
	return *apiResp.Data, nil
}

type MerchantWalletResponse struct {
	Money              string `json:"money"`
	TotalMoney         string `json:"total_money"`
	WithdrawMoney      string `json:"withdraw_money"`
	TotalWithdrawMoney string `json:"total_withdraw_money"`
	TotalPayoutMoney   string `json:"total_payout_money"`
}

func QueryMerchantWallet() (MerchantWalletResponse, error) {
	cfg := utils.GlobalConfig.Pay.Hopopay
	if missing := missingBaseConfigFields(cfg); len(missing) > 0 {
		return MerchantWalletResponse{}, fmt.Errorf("HOPOPAY 商户钱包查询配置不完整，core.yaml pay.hopopay 缺少: %s", strings.Join(missing, ", "))
	}
	params := map[string]any{
		"merchant_id": strings.TrimSpace(cfg.MerchantID),
		"timestamp":   time.Now().Unix(),
		"nonce":       newNonce(),
	}
	params["signature"] = BuildSignature(params, cfg.Secret)

	respBody, err := postJSON(baseURL(cfg)+"/api/merchant/wallet/query", params)
	if err != nil {
		return MerchantWalletResponse{}, fmt.Errorf("HOPOPAY 商户钱包查询请求失败: %w", err)
	}
	var apiResp walletQueryResp
	if err = json.Unmarshal(respBody, &apiResp); err != nil {
		return MerchantWalletResponse{}, fmt.Errorf("HOPOPAY 商户钱包查询响应解析失败: %w", err)
	}
	if apiResp.Status != 1 {
		return MerchantWalletResponse{}, fmt.Errorf("HOPOPAY 商户钱包查询失败 status=%d msg=%s", apiResp.Status, apiErrorMessage(apiResp.Info, apiResp.Error))
	}
	if apiResp.Data == nil {
		return MerchantWalletResponse{}, errors.New("HOPOPAY 商户钱包查询失败: data is empty")
	}
	return *apiResp.Data, nil
}

func buildPayinParams(cfg base.HopopayPayConfig, req pay.PayRequest, timestamp int64, nonce string) (map[string]any, error) {
	appUserID := resolveAppUserID(req.UserID)
	if appUserID == "" {
		return nil, errors.New("HOPOPAY 代收参数缺少 app_user_id")
	}
	payType := resolvePayType(req.PayMethod, req.ExtraFields)
	if payType == "" {
		return nil, errors.New("HOPOPAY 代收参数缺少 pay_type")
	}
	clientIP := resolveRequestIP(req.ExtraFields, req.ClientIP, req.UserIP)
	if clientIP == "" {
		return nil, errors.New("HOPOPAY 代收参数缺少 ip")
	}
	returnURL := firstNonEmpty(strings.TrimSpace(req.ReturnURL), strings.TrimSpace(cfg.PageURL))
	if returnURL == "" {
		return nil, errors.New("HOPOPAY 代收参数缺少 return_url")
	}

	extraParams, err := json.Marshal(hopopayPayinExtraParams{PayType: payType, Platform: fixedPlatform})
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"merchant_id":  strings.TrimSpace(cfg.MerchantID),
		"app_user_id":  appUserID,
		"app_order_id": strings.TrimSpace(req.OrderNo),
		"amount":       amountNumber(pay.ResolveOrderAmount(req)),
		"timestamp":    timestamp,
		"nonce":        nonce,
		"ip":           clientIP,
		"extra_params": string(extraParams),
		"callback_url": firstNonEmpty(strings.TrimSpace(req.NotifyURL), strings.TrimSpace(cfg.NotifyURL)),
		"return_url":   returnURL,
	}, nil
}

func buildPayoutParams(cfg base.HopopayPayConfig, req pay.PayoutRequest, timestamp int64, nonce string) (map[string]any, error) {
	appUserID := resolveAppUserID(req.UserID)
	if appUserID == "" {
		return nil, errors.New("HOPOPAY 代付参数缺少 app_user_id")
	}
	bankCode := resolvePayoutBankCode(req.BankCode)
	if bankCode == "" {
		return nil, errors.New("HOPOPAY 代付参数缺少 bank_code")
	}
	if strings.TrimSpace(req.AccNo) == "" {
		return nil, errors.New("HOPOPAY 代付参数缺少 bank_number")
	}
	if strings.TrimSpace(req.AccName) == "" {
		return nil, errors.New("HOPOPAY 代付参数缺少 bank_account_name")
	}
	clientIP := resolveRequestIP(req.ExtraFields, req.ClientIP, req.UserIP)
	if clientIP == "" {
		return nil, errors.New("HOPOPAY 代付参数缺少 ip")
	}

	extraParams, err := json.Marshal(hopopayPayoutExtraParams{Platform: fixedPlatform})
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"merchant_id":       strings.TrimSpace(cfg.MerchantID),
		"app_user_id":       appUserID,
		"app_order_id":      strings.TrimSpace(req.OrderNo),
		"amount":            amountNumber(pay.ResolvePayoutAmount(req)),
		"timestamp":         timestamp,
		"nonce":             nonce,
		"ip":                clientIP,
		"bank_code":         bankCode,
		"bank_number":       strings.TrimSpace(req.AccNo),
		"bank_account_name": strings.TrimSpace(req.AccName),
		"extra_params":      string(extraParams),
		"callback_url":      firstNonEmpty(strings.TrimSpace(req.NotifyURL), strings.TrimSpace(cfg.PayoutNotifyURL), strings.TrimSpace(cfg.NotifyURL)),
	}, nil
}

func buildQueryParams(cfg base.HopopayPayConfig, appOrderID string, providerOrderNo string, timestamp int64, nonce string) (map[string]any, error) {
	appOrderID = strings.TrimSpace(appOrderID)
	providerOrderNo = strings.TrimSpace(providerOrderNo)
	if appOrderID == "" {
		return nil, errors.New("HOPOPAY 查询参数缺少 app_order_id")
	}
	if providerOrderNo == "" {
		return nil, errors.New("HOPOPAY 查询参数缺少 order_no")
	}
	return map[string]any{
		"merchant_id":  strings.TrimSpace(cfg.MerchantID),
		"app_order_id": appOrderID,
		"order_no":     providerOrderNo,
		"timestamp":    timestamp,
		"nonce":        nonce,
	}, nil
}

func missingPayinConfigFields(cfg base.HopopayPayConfig) []string {
	missing := missingBaseConfigFields(cfg)
	if strings.TrimSpace(cfg.NotifyURL) == "" {
		missing = append(missing, "notifyUrl")
	}
	return missing
}

func missingPayoutConfigFields(cfg base.HopopayPayConfig) []string {
	missing := missingBaseConfigFields(cfg)
	if strings.TrimSpace(cfg.PayoutNotifyURL) == "" && strings.TrimSpace(cfg.NotifyURL) == "" {
		missing = append(missing, "payoutNotifyUrl")
	}
	return missing
}

func missingBaseConfigFields(cfg base.HopopayPayConfig) []string {
	missing := make([]string, 0, 3)
	if strings.TrimSpace(cfg.BaseURL) == "" {
		missing = append(missing, "baseUrl")
	}
	if strings.TrimSpace(cfg.MerchantID) == "" {
		missing = append(missing, "merchantId")
	}
	if strings.TrimSpace(cfg.Secret) == "" {
		missing = append(missing, "secret")
	}
	return missing
}

func baseURL(cfg base.HopopayPayConfig) string {
	return strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/")
}

func resolveAppUserID(userID int64) string {
	if userID <= 0 {
		return ""
	}
	return strconv.FormatInt(userID, 10)
}

func resolvePayType(payMethod string, fields map[string]string) string {
	value := firstNonEmpty(payMethod, extraField(fields, "pay_type", "payType"))
	value = strings.ToLower(strings.TrimSpace(value))
	if _, ok := allowedPayTypes[value]; ok {
		return value
	}
	return ""
}

func resolvePayoutBankCode(bankCode string) string {
	value := strings.ToLower(strings.TrimSpace(bankCode))
	switch value {
	case payTypeCashapp, payTypePaypal:
		return value
	default:
		return ""
	}
}

func resolveRequestIP(fields map[string]string, candidates ...string) string {
	if value := extraField(fields, "ip", "clientIp", "client_ip", "realIp", "real_ip"); value != "" {
		return value
	}
	return firstNonEmpty(candidates...)
}

func extraField(fields map[string]string, aliases ...string) string {
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

func amountNumber(amount float64) float64 {
	return utils.Truncate2(amount)
}

func newNonce() string {
	var b [8]byte
	if _, err := rand.Read(b[:]); err == nil {
		return hex.EncodeToString(b[:])
	}
	return fmt.Sprintf("%016d", time.Now().UnixNano()%1_0000_0000_0000_0000)
}

func postJSON(url string, payload map[string]any) ([]byte, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		log.Printf("[%s] third-party request marshal failed url=%s err=%v", providerName, url, err)
		return nil, err
	}
	log.Printf("[%s] third-party request url=%s params=%s", providerName, url, string(body))

	resp, err := http.Post(url, "application/json; charset=utf-8", bytes.NewReader(body)) //nolint:noctx
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

func apiErrorMessage(info string, apiErr *hopopayAPIError) string {
	if apiErr != nil {
		return strings.TrimSpace(apiErr.ErrorCode.String() + " " + apiErr.ErrorMsg)
	}
	return strings.TrimSpace(info)
}

func normalizePayoutQueryStatus(status int) int {
	switch status {
	case 3:
		return 7
	case 2, 5, 6, 9:
		return 8
	default:
		return 0
	}
}

func formatUnixSeconds(seconds int64) string {
	if seconds <= 0 {
		return ""
	}
	return time.Unix(seconds, 0).Format(time.RFC3339)
}

type hopopayPayinExtraParams struct {
	PayType  string `json:"pay_type"`
	Platform string `json:"platform"`
}

type hopopayPayoutExtraParams struct {
	Platform string `json:"platform"`
}

type hopopayAPIError struct {
	ErrorCode flexString `json:"error_code"`
	ErrorMsg  string     `json:"error_msg"`
}

type payinOrderResp struct {
	Info      string           `json:"info"`
	Status    int              `json:"status"`
	Data      *payinOrderData  `json:"data"`
	Error     *hopopayAPIError `json:"error"`
	RequestID string           `json:"request_id"`
}

type payinOrderData struct {
	Step       string `json:"step"`
	OrderNo    string `json:"order_no"`
	AppOrderID string `json:"app_order_id"`
	Status     int    `json:"status"`
	URL        string `json:"url"`
	QRCode     string `json:"qrcode"`
}

type payoutOrderResp struct {
	Info      string           `json:"info"`
	Status    int              `json:"status"`
	Data      *payoutOrderData `json:"data"`
	Error     *hopopayAPIError `json:"error"`
	RequestID string           `json:"request_id"`
}

type payoutOrderData struct {
	Step       string `json:"step"`
	OrderNo    string `json:"order_no"`
	AppOrderID string `json:"app_order_id"`
	Status     int    `json:"status"`
}

type payinQueryResp struct {
	Info      string              `json:"info"`
	Status    int                 `json:"status"`
	Data      *PayinQueryResponse `json:"data"`
	Error     *hopopayAPIError    `json:"error"`
	RequestID string              `json:"request_id"`
}

type payoutQueryResp struct {
	Info      string               `json:"info"`
	Status    int                  `json:"status"`
	Data      *PayoutQueryResponse `json:"data"`
	Error     *hopopayAPIError     `json:"error"`
	RequestID string               `json:"request_id"`
}

type walletQueryResp struct {
	Info      string                  `json:"info"`
	Status    int                     `json:"status"`
	Data      *MerchantWalletResponse `json:"data"`
	Error     *hopopayAPIError        `json:"error"`
	RequestID string                  `json:"request_id"`
}

type flexFloat float64

func (v *flexFloat) UnmarshalJSON(data []byte) error {
	data = bytes.TrimSpace(data)
	if bytes.Equal(data, []byte("null")) || len(data) == 0 {
		*v = 0
		return nil
	}
	var n float64
	if err := json.Unmarshal(data, &n); err == nil {
		*v = flexFloat(n)
		return nil
	}
	var text string
	if err := json.Unmarshal(data, &text); err == nil {
		parsed, _ := strconv.ParseFloat(strings.TrimSpace(text), 64)
		*v = flexFloat(parsed)
	}
	return nil
}

func (v flexFloat) Float64() float64 {
	return float64(v)
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
