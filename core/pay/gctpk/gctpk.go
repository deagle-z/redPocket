package gctpk

import (
	"BaseGoUni/core/base"
	"BaseGoUni/core/pay"
	"BaseGoUni/core/utils"
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
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
	return "GCTPK"
}

// CreateOrder 调用 GCTPK 代收下单接口
// POST https://naskl.gctpk.com/payin/createOrder
func (g *Provider) CreateOrder(req pay.PayRequest) (pay.PayResponse, error) {
	cfg := utils.GlobalConfig.Pay.Gctpk
	if cfg.MerNo == "" || cfg.Secret == "" {
		return pay.PayResponse{}, fmt.Errorf("GCTPK 配置不完整，请检查 core.yaml pay.gctpk 节点")
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
		"email":       req.ExtraFields["email"],
		"phone":       req.ExtraFields["phone"],
		"orderAmount": fmt.Sprintf("%.2f", req.Amount),
		"currency":    req.Currency,
		"busiCode":    "104004",
		"pageUrl":     resolvePageURL(cfg, req),
		"notifyUrl":   cfg.NotifyURL,
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
	if apiResp.Data != nil {
		payURL = apiResp.Data.PayURL
	}
	return pay.PayResponse{PayURL: payURL}, nil
}

// resolveBusiCode 从 ExtraFields["busiCode"] 取，或用 PayMethod 兜底
func resolveBusiCode(req pay.PayRequest) string {
	if v := strings.TrimSpace(req.ExtraFields["busiCode"]); v != "" {
		return v
	}
	return strings.TrimSpace(req.PayMethod)
}

// resolvePageURL 优先请求级 ReturnURL，其次配置
func resolvePageURL(cfg base.GctpkPayConfig, req pay.PayRequest) string {
	if req.ReturnURL != "" {
		return req.ReturnURL
	}
	return cfg.PageURL
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
	resp, err := http.Post(url, "application/json", bytes.NewReader(body)) //nolint:noctx
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

// ---- 响应结构 ----

type createOrderResp struct {
	Code int              `json:"code"`
	Msg  string           `json:"msg"`
	Data *createOrderData `json:"data"`
}

type createOrderData struct {
	PayURL string `json:"payUrl"`
}
