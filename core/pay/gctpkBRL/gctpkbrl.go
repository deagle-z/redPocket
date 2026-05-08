package gctpk

import (
	"BaseGoUni/core/base"
	"BaseGoUni/core/pay"
	"BaseGoUni/core/utils"
	"bytes"
	"crypto"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"
	"strings"
	"time"
)

// Provider GCTPK 代收支付渠道
type Provider struct{}

func init() {
	pay.Register(&Provider{})
}

func (g *Provider) Name() string {
	return "GCTPKBRL"
}

// CreateOrder 调用 GCTPK 代收下单接口
// POST https://naskl.gctpk.com/payin/createOrder
func (g *Provider) CreateOrder(req pay.PayRequest) (pay.PayResponse, error) {
	cfg := utils.GlobalConfig.Pay.Gctpkbrl
	if cfg.MerNo == "" || cfg.Secret == "" {
		return pay.PayResponse{}, fmt.Errorf("GCTPKBRL 配置不完整，请检查 core.yaml pay.gctpkbrl 节点")
	}

	baseURL := strings.TrimRight(cfg.BaseURL, "/")
	if baseURL == "" {
		baseURL = "https://naskl.gctpk.com"
	}

	timestamp := fmt.Sprintf("%d", time.Now().UnixMilli())

	// 必填参数
	params := map[string]string{
		"merNo":       cfg.MerNo,
		"merOrderNo":  req.OrderNo,
		"email":       req.ExtraFields["emailbrl"],
		"phone":       req.ExtraFields["phonebrl"],
		"orderAmount": fmt.Sprintf("%.2f", req.Amount),
		"currency":    "BRL",
		"busiCode":    resolveBusiCode(req, "106001"),
		"pageUrl":     resolvePageURL(cfg, req),
		"notifyUrl":   resolveNotifyURL(cfg, req.NotifyURL),
		"timestamp":   timestamp,
	}

	params["sign"] = BuildSign(params, cfg.Secret)

	respBody, err := postJSON(baseURL+"/payin/createOrder", params)
	if err != nil {
		return pay.PayResponse{}, fmt.Errorf("GCTPK 请求失败: %w", err)
	}

	var apiResp createOrderResp
	if err = json.Unmarshal(respBody, &apiResp); err != nil {
		return pay.PayResponse{}, fmt.Errorf("GCTPK 响应解析失败: %w", err)
	}
	if apiResp.Code != 200 {
		return pay.PayResponse{}, fmt.Errorf("GCTPK 下单失败: %s", apiResp.Msg)
	}

	payURL := ""
	providerOrderNo := ""
	if apiResp.Data != nil {
		payURL = apiResp.Data.PayURL
		if strings.TrimSpace(payURL) == "" {
			payURL = apiResp.Data.OrderData
		}
		providerOrderNo = apiResp.Data.OrderNo
	}
	return pay.PayResponse{
		PayURL:          payURL,
		ProviderTradeNo: providerOrderNo,
	}, nil
}

// resolveBusiCode 从 ExtraFields["busiCode"] 取，或用 PayMethod 兜底，再回退默认值
func resolveBusiCode(req pay.PayRequest, defaultCode string) string {
	if v := strings.TrimSpace(req.ExtraFields["busiCode"]); v != "" {
		return v
	}
	if v := strings.TrimSpace(req.PayMethod); v != "" {
		return v
	}
	return defaultCode
}

// resolvePageURL 优先请求级 ReturnURL，其次配置
func resolvePageURL(cfg base.GctpkPayConfig, req pay.PayRequest) string {
	if req.ReturnURL != "" {
		return req.ReturnURL
	}
	return cfg.PageURL
}

func resolveNotifyURL(cfg base.GctpkPayConfig, notifyURL string) string {
	if strings.TrimSpace(notifyURL) != "" {
		return strings.TrimSpace(notifyURL)
	}
	return cfg.NotifyURL
}

// BuildSign 所有非空参数按 key ASCII 升序排列后 HmacSHA256 签名（供回调验签复用）
// 空值参数和 sign 字段不参与签名（见 readme.md）
func BuildSign(params map[string]string, secret string) string {
	keys := make([]string, 0, len(params))
	for k, v := range params {
		if k == "sign" || strings.TrimSpace(v) == "" {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, k+"="+params[k])
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(strings.Join(parts, "&")))
	return hex.EncodeToString(mac.Sum(nil))
}

func postJSON(url string, payload map[string]string) ([]byte, error) {
	body, _ := json.Marshal(payload)
	log.Printf("[GCTPKBRL] third-party request url=%s params=%s", url, string(body))
	resp, err := http.Post(url, "application/json", bytes.NewReader(body)) //nolint:noctx
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respBody, readErr := io.ReadAll(resp.Body)
	log.Printf("[GCTPKBRL] third-party response status=%d body=%s", resp.StatusCode, string(respBody))
	return respBody, readErr
}

// CreatePayoutOrder 调用 GCTPK 代付下单接口
// POST https://taslk.gctpk.com/payout/singleOrder
// 签名：HmacSHA256 → signA → RSA-1024 私钥加密 → Base64
func (g *Provider) CreatePayoutOrder(req pay.PayoutRequest) (pay.PayoutResponse, error) {
	cfg := utils.GlobalConfig.Pay.Gctpkbrl
	if cfg.MerNo == "" || cfg.Secret == "" || cfg.PrivateKey == "" {
		return pay.PayoutResponse{}, fmt.Errorf("GCTPKBRL 代付配置不完整，请检查 core.yaml pay.gctpkbrl 节点")
	}

	baseURL := strings.TrimRight(cfg.PayoutBaseURL, "/")
	if baseURL == "" {
		baseURL = "https://taslk.gctpk.com"
	}
	notifyURL := cfg.PayoutNotifyURL
	if notifyURL == "" {
		notifyURL = cfg.NotifyURL
	}
	identityType := resolvePayoutIdentityType(req)
	if identityType == "" {
		return pay.PayoutResponse{}, fmt.Errorf("GCTPKBRL 代付参数缺少 identityType")
	}

	timestamp := fmt.Sprintf("%d", time.Now().UnixMilli())

	params := buildPayoutParams(cfg, req, notifyURL, timestamp, identityType)

	// 签名：HmacSHA256 → signA → RSA 私钥加密 → Base64
	signA := BuildSign(params, cfg.Secret)
	rsaSign, err := rsaPrivateKeyEncrypt([]byte(signA), cfg.PrivateKey)
	if err != nil {
		return pay.PayoutResponse{}, fmt.Errorf("GCTPK 代付签名失败: %w", err)
	}
	params["sign"] = rsaSign

	respBody, err := postJSON(baseURL+"/payout/singleOrder", params)
	if err != nil {
		return pay.PayoutResponse{}, fmt.Errorf("GCTPK 代付请求失败: %w", err)
	}

	var apiResp payoutOrderResp
	if err = json.Unmarshal(respBody, &apiResp); err != nil {
		return pay.PayoutResponse{}, fmt.Errorf("GCTPK 代付响应解析失败: %w", err)
	}
	// code=200 或 code=500 均表示请求成功（订单已入库），以 data.status 为准
	if apiResp.Code != 200 && apiResp.Code != 500 {
		return pay.PayoutResponse{}, fmt.Errorf("GCTPK 代付下单失败 code=%d msg=%s", apiResp.Code, apiResp.Msg)
	}

	resp := pay.PayoutResponse{}
	if apiResp.Data != nil {
		resp.ProviderOrderNo = apiResp.Data.OrderNo
		resp.Status = apiResp.Data.Status
	}
	return resp, nil
}

func buildPayoutParams(cfg base.GctpkPayConfig, req pay.PayoutRequest, notifyURL string, timestamp string, identityType string) map[string]string {
	params := map[string]string{
		"merNo":        cfg.MerNo,
		"merOrderNo":   req.OrderNo,
		"identityType": identityType,
		"identityNo":   resolvePayoutIdentityNo(req),
		"accName":      req.AccName,
		"accNo":        req.AccNo,
		"bankCode":     "PIX",
		"busiCode":     resolvePayoutBusiCode(req, "206001"),
		"currency":     "BRL",
		"email":        req.Email,
		"notifyUrl":    notifyURL,
		"orderAmount":  fmt.Sprintf("%.2f", req.Amount),
		"phone":        req.Phone,
		"timestamp":    timestamp,
	}
	// 可选扩展字段（extend / province 等）
	for _, k := range []string{"extend", "province"} {
		if v := strings.TrimSpace(req.ExtraFields[k]); v != "" {
			params[k] = v
		}
	}
	return params
}

func resolvePayoutIdentityType(req pay.PayoutRequest) string {
	if v := strings.TrimSpace(req.IdentityType); v != "" {
		return v
	}
	return strings.TrimSpace(req.ExtraFields["identityType"])
}

func resolvePayoutIdentityNo(req pay.PayoutRequest) string {
	if v := strings.TrimSpace(req.IdentityNo); v != "" {
		return v
	}
	return strings.TrimSpace(req.ExtraFields["identityNo"])
}

func resolvePayoutBusiCode(req pay.PayoutRequest, defaultCode string) string {
	if v := strings.TrimSpace(req.BusiCode); v != "" {
		return v
	}
	return defaultCode
}

// rsaPrivateKeyEncrypt 使用 PKCS#8 RSA-1024 私钥对 data 进行 PKCS1v15 加密（即私钥签名），返回 Base64 字符串
func rsaPrivateKeyEncrypt(data []byte, pemKey string) (string, error) {
	block, _ := pem.Decode([]byte(pemKey))
	if block == nil {
		return "", fmt.Errorf("PEM 解码失败，请确认私钥格式正确（PKCS#8）")
	}

	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return "", fmt.Errorf("解析 PKCS#8 私钥失败: %w", err)
	}
	rsaKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return "", fmt.Errorf("私钥不是 RSA 类型")
	}

	// 代付签名：RSA PKCS1v15，hash=0 表示 data 直接作为消息体（不再 hash）
	sig, err := rsa.SignPKCS1v15(rand.Reader, rsaKey, crypto.Hash(0), data)
	if err != nil {
		return "", fmt.Errorf("RSA 签名失败: %w", err)
	}
	return base64.StdEncoding.EncodeToString(sig), nil
}

// ---- 响应结构 ----

type createOrderResp struct {
	Code int              `json:"code"`
	Msg  string           `json:"msg"`
	Data *createOrderData `json:"data"`
}

type createOrderData struct {
	PayURL    string `json:"payUrl"`
	OrderData string `json:"orderData"`
	OrderNo   string `json:"orderNo"`
	Status    int    `json:"status"`
	Common    string `json:"common"`
}

type payoutOrderResp struct {
	Code int              `json:"code"`
	Msg  string           `json:"msg"`
	Data *payoutOrderData `json:"data"`
}

type payoutOrderData struct {
	OrderNo string `json:"orderNo"`
	Status  int    `json:"status"`
}
