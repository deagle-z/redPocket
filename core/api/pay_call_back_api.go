package api

import (
	"BaseGoUni/core/base"
	"BaseGoUni/core/pay/vcpaymxn"
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/repository"
	"BaseGoUni/core/utils"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"log"
	"strconv"
	"strings"
)

type gctpkConfigResolver func(req gctpkNotifyReq) (base.GctpkPayConfig, string, error)

// GctpkPayinCallback GCTPK 代收异步回调（公开接口，无需 token）
// POST /api/v1/pay/gctpk/notify
func GctpkPayinCallback(ctx *gin.Context) {
	handleGctpkPayinCallback(ctx, resolveAnyGctpkPayConfig)
}

// GctpkPayoutCallback GCTPK 代付/提现异步回调（公开接口，无需 token）
// POST /api/v1/pay/gctpk/payoutNotify
func GctpkPayoutCallback(ctx *gin.Context) {
	handleGctpkPayoutCallback(ctx, func(req gctpkPayoutNotifyReq) (base.GctpkPayConfig, string, error) {
		return validateGctpkPayoutConfig(req, utils.GlobalConfig.Pay.Gctpk, "GCTPK")
	})
}

// GctpkMxnPayinCallback GCTPKMXN 代收异步回调（公开接口，无需 token）
// POST /api/v1/pay/gctpkmxn/notify
func GctpkMxnPayinCallback(ctx *gin.Context) {
	handleGctpkPayinCallback(ctx, func(req gctpkNotifyReq) (base.GctpkPayConfig, string, error) {
		return validateGctpkPayConfig(req, utils.GlobalConfig.Pay.Gctpkmxn, "GCTPKMXN")
	})
}

// GctpkMxnPayoutCallback GCTPKMXN 代付/提现异步回调（公开接口，无需 token）
// POST /api/v1/pay/gctpkmxn/payoutNotify
func GctpkMxnPayoutCallback(ctx *gin.Context) {
	handleGctpkPayoutCallback(ctx, func(req gctpkPayoutNotifyReq) (base.GctpkPayConfig, string, error) {
		return validateGctpkPayoutConfig(req, utils.GlobalConfig.Pay.Gctpkmxn, "GCTPKMXN")
	})
}

// GctpkBrlPayinCallback GCTPKBRL 代收异步回调（公开接口，无需 token）
// POST /api/v1/pay/gctpkbrl/notify
func GctpkBrlPayinCallback(ctx *gin.Context) {
	handleGctpkPayinCallback(ctx, func(req gctpkNotifyReq) (base.GctpkPayConfig, string, error) {
		return validateGctpkPayConfig(req, utils.GlobalConfig.Pay.Gctpkbrl, "GCTPKBRL")
	})
}

// GctpkBrlPayoutCallback GCTPKBRL 代付/提现异步回调（公开接口，无需 token）
// POST /api/v1/pay/gctpkbrl/payoutNotify
func GctpkBrlPayoutCallback(ctx *gin.Context) {
	handleGctpkPayoutCallback(ctx, func(req gctpkPayoutNotifyReq) (base.GctpkPayConfig, string, error) {
		return validateGctpkPayoutConfig(req, utils.GlobalConfig.Pay.Gctpkbrl, "GCTPKBRL")
	})
}

// VcpayMxnPayinCallback VCPAYMXN 代收异步回调（公开接口，无需 token）
// POST /api/v1/pay/vcpaymxn/notify
func VcpayMxnPayinCallback(ctx *gin.Context) {
	var req vcpayMxnNotifyReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.String(400, "FAIL")
		return
	}
	cfg, err := validateVcpayMxnNotifyConfig(req)
	if err != nil {
		log.Printf("[VCPAYMXN Notify] 验签失败 providerOrderNo=%s merchantOrderNo=%s err=%v", req.TradeNo, req.OutTradeNo, err)
		ctx.String(400, "FAIL")
		return
	}

	db := ctx.MustGet("db").(*gorm.DB)
	hostInfo := ctx.MustGet("hostInfo").(pojo.HostInfo)
	localOrderNo := strings.TrimSpace(req.OutTradeNo)
	providerOrderNo := strings.TrimSpace(req.TradeNo)
	if localOrderNo == "" {
		log.Printf("[VCPAYMXN Notify] 商户订单号为空 providerOrderNo=%s", providerOrderNo)
		ctx.String(400, "FAIL")
		return
	}

	if req.Code.String() == "200" && req.TradeState == 1 {
		if err := repository.ProcessRechargeOrderSuccess(
			db, localOrderNo, providerOrderNo, centsToPayAmount(req.OrderAmount.Int64()), hostInfo.TablePrefix,
		); err != nil {
			log.Printf("[VCPAYMXN Notify] 入账失败 localOrderNo=%s providerOrderNo=%s appID=%s err=%v", localOrderNo, providerOrderNo, cfg.AppID, err)
			ctx.String(500, "FAIL")
			return
		}
	}

	ctx.String(200, "SUCCESS")
}

// VcpayMxnPayoutCallback VCPAYMXN 代付/提现异步回调（公开接口，无需 token）
// POST /api/v1/pay/vcpaymxn/payoutNotify
func VcpayMxnPayoutCallback(ctx *gin.Context) {
	var req vcpayMxnNotifyReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.String(400, "FAIL")
		return
	}
	if _, err := validateVcpayMxnNotifyConfig(req); err != nil {
		log.Printf("[VCPAYMXN Payout Notify] 验签失败 providerOrderNo=%s merchantOrderNo=%s err=%v", req.TradeNo, req.OutTradeNo, err)
		ctx.String(400, "FAIL")
		return
	}

	db := ctx.MustGet("db").(*gorm.DB)
	localOrderNo := strings.TrimSpace(req.OutTradeNo)
	providerOrderNo := strings.TrimSpace(req.TradeNo)
	if localOrderNo == "" {
		log.Printf("[VCPAYMXN Payout Notify] 商户订单号为空 providerOrderNo=%s", providerOrderNo)
		ctx.String(400, "FAIL")
		return
	}

	success := req.Code.String() == "200" && req.TradeState == 1
	failed := req.Code.String() != "200" || (req.TradeState != 0 && req.TradeState != 1)
	if success || failed {
		if err := repository.ProcessWithdrawOrderPayoutCallback(db, repository.WithdrawPayoutCallback{
			LocalOrderNo:     localOrderNo,
			ProviderPayoutNo: providerOrderNo,
			ProviderStatus:   fmt.Sprintf("%d", req.TradeState),
			ProviderAmount:   centsToPayAmount(req.OrderAmount.Int64()),
			ResultCode:       req.Code.String(),
			ResultMsg:        strings.TrimSpace(req.Msg),
			Success:          success,
			Failed:           failed,
		}); err != nil {
			log.Printf("[VCPAYMXN Payout Notify] 处理提现回调失败 localOrderNo=%s providerOrderNo=%s status=%d err=%v", localOrderNo, providerOrderNo, req.TradeState, err)
			ctx.String(500, "FAIL")
			return
		}
	}

	ctx.String(200, "SUCCESS")
}

func handleGctpkPayinCallback(ctx *gin.Context, resolveConfig gctpkConfigResolver) {
	var req gctpkNotifyReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.String(400, "FAIL")
		return
	}

	_, providerName, err := resolveConfig(req)
	if err != nil {
		log.Printf("[GCTPK Notify] 验签失败 provider=%s providerOrderNo=%s merchantOrderNo=%s err=%v", providerName, req.OrderNo, req.MerOrderNo, err)
		ctx.String(400, "FAIL")
		return
	}

	db := ctx.MustGet("db").(*gorm.DB)
	hostInfo := ctx.MustGet("hostInfo").(pojo.HostInfo)
	localOrderNo := strings.TrimSpace(req.MerOrderNo)
	providerOrderNo := strings.TrimSpace(req.OrderNo)

	if localOrderNo == "" {
		log.Printf("[GCTPK Notify] 商户订单号为空 provider=%s providerOrderNo=%s", providerName, providerOrderNo)
		ctx.String(400, "FAIL")
		return
	}

	switch req.Status {
	case 5: // 支付成功
		if err := repository.ProcessRechargeOrderSuccess(
			db, localOrderNo, providerOrderNo, parsePayFloat(req.PayAmount), hostInfo.TablePrefix,
		); err != nil {
			log.Printf("[GCTPK Notify] 入账失败 provider=%s localOrderNo=%s providerOrderNo=%s err=%v", providerName, localOrderNo, providerOrderNo, err)
			ctx.String(500, "FAIL")
			return
		}
	case 4: // 订单已关闭
		if err := repository.ProcessRechargeOrderClosed(db, localOrderNo); err != nil {
			log.Printf("[GCTPK Notify] 关闭订单失败 provider=%s localOrderNo=%s providerOrderNo=%s err=%v", providerName, localOrderNo, providerOrderNo, err)
			ctx.String(500, "FAIL")
			return
		}
	default:
		// 其他状态不处理，直接返回 SUCCESS 防止重复通知
	}

	ctx.String(200, "SUCCESS")
}

func handleGctpkPayoutCallback(ctx *gin.Context, resolveConfig gctpkPayoutConfigResolver) {
	var req gctpkPayoutNotifyReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.String(400, "FAIL")
		return
	}

	_, providerName, err := resolveConfig(req)
	if err != nil {
		log.Printf("[GCTPK Payout Notify] 验签失败 provider=%s providerOrderNo=%s merchantOrderNo=%s err=%v", providerName, req.OrderNo, req.MerOrderNo, err)
		ctx.String(400, "FAIL")
		return
	}

	db := ctx.MustGet("db").(*gorm.DB)
	localOrderNo := strings.TrimSpace(req.MerOrderNo)
	providerOrderNo := strings.TrimSpace(req.OrderNo)
	if localOrderNo == "" {
		log.Printf("[GCTPK Payout Notify] 商户订单号为空 provider=%s providerOrderNo=%s", providerName, providerOrderNo)
		ctx.String(400, "FAIL")
		return
	}

	success := req.Status == 7
	failed := req.Status == 8 || req.Status == 9 || (!success && strings.TrimSpace(req.ResultCode) != "")
	if success || failed {
		if err := repository.ProcessWithdrawOrderPayoutCallback(db, repository.WithdrawPayoutCallback{
			LocalOrderNo:     localOrderNo,
			ProviderPayoutNo: providerOrderNo,
			ProviderStatus:   fmt.Sprintf("%d", req.Status),
			ProviderAmount:   parsePayFloat(req.OrderAmount),
			ResultCode:       strings.TrimSpace(req.ResultCode),
			ResultMsg:        strings.TrimSpace(req.ResultMsg),
			ProviderPayTime:  strings.TrimSpace(req.PayTime),
			Success:          success,
			Failed:           failed,
		}); err != nil {
			log.Printf("[GCTPK Payout Notify] 处理提现回调失败 provider=%s localOrderNo=%s providerOrderNo=%s status=%d err=%v", providerName, localOrderNo, providerOrderNo, req.Status, err)
			ctx.String(500, "FAIL")
			return
		}
	}

	ctx.String(200, "SUCCESS")
}

func resolveAnyGctpkPayConfig(req gctpkNotifyReq) (base.GctpkPayConfig, string, error) {
	candidates := []struct {
		name string
		cfg  base.GctpkPayConfig
	}{
		{name: "GCTPK", cfg: utils.GlobalConfig.Pay.Gctpk},
		{name: "GCTPKMXN", cfg: utils.GlobalConfig.Pay.Gctpkmxn},
		{name: "GCTPKBRL", cfg: utils.GlobalConfig.Pay.Gctpkbrl},
	}

	var lastErr error
	for _, candidate := range candidates {
		cfg, providerName, err := validateGctpkPayConfig(req, candidate.cfg, candidate.name)
		if err == nil {
			return cfg, providerName, nil
		}
		lastErr = err
	}

	return base.GctpkPayConfig{}, "UNKNOWN", lastErr
}

type gctpkPayoutConfigResolver func(req gctpkPayoutNotifyReq) (base.GctpkPayConfig, string, error)

func validateGctpkPayoutConfig(req gctpkPayoutNotifyReq, cfg base.GctpkPayConfig, providerName string) (base.GctpkPayConfig, string, error) {
	return cfg, providerName, nil
}

func buildGctpkPayoutNotifyParams(req gctpkPayoutNotifyReq) map[string]string {
	return map[string]string{
		"merNo":       req.MerNo,
		"merOrderNo":  req.MerOrderNo,
		"orderNo":     req.OrderNo,
		"orderAmount": req.OrderAmount,
		"payTime":     req.PayTime,
		"status":      fmt.Sprintf("%d", req.Status),
		"resultCode":  req.ResultCode,
		"resultMsg":   req.ResultMsg,
	}
}

func validateGctpkPayConfig(req gctpkNotifyReq, cfg base.GctpkPayConfig, providerName string) (base.GctpkPayConfig, string, error) {
	return cfg, providerName, nil
}

func buildGctpkNotifyParams(req gctpkNotifyReq) map[string]string {
	return map[string]string{
		"payAmount":  req.PayAmount,
		"orderNo":    req.OrderNo,
		"merNo":      req.MerNo,
		"merOrderNo": req.MerOrderNo,
		"payTime":    req.PayTime,
		"busiCode":   req.BusiCode,
		"status":     fmt.Sprintf("%d", req.Status),
	}
}

// gctpkNotifyReq GCTPK 代收回调请求体
type gctpkNotifyReq struct {
	PayAmount  string `json:"payAmount"`  // 实际支付金额
	OrderNo    string `json:"orderNo"`    // 三方订单号
	MerNo      string `json:"merNo"`      // 商户号
	MerOrderNo string `json:"merOrderNo"` // 商户订单号（本地订单号）
	PayTime    string `json:"payTime"`    // 交易时间
	BusiCode   string `json:"busiCode"`   // 支付编码
	Sign       string `json:"sign"`       // 签名
	Status     int    `json:"status"`     // 5=支付成功 4=订单已关闭
}

// gctpkPayoutNotifyReq GCTPK 代付/提现回调请求体
type gctpkPayoutNotifyReq struct {
	MerNo       string `json:"merNo"`       // 商户号
	MerOrderNo  string `json:"merOrderNo"`  // 商户订单号（本地提现订单号）
	OrderNo     string `json:"orderNo"`     // 三方代付订单号
	OrderAmount string `json:"orderAmount"` // 代付金额
	Sign        string `json:"sign"`        // 签名
	PayTime     string `json:"payTime"`     // 交易时间
	Status      int    `json:"status"`      // 7=代付成功
	ResultCode  string `json:"resultCode"`  // 失败码
	ResultMsg   string `json:"resultMsg"`   // 失败原因
}

func parsePayFloat(s string) float64 {
	var v float64
	fmt.Sscanf(s, "%f", &v)
	return v
}

func validateVcpayMxnNotifyConfig(req vcpayMxnNotifyReq) (base.VcpayMxnPayConfig, error) {
	cfg := utils.GlobalConfig.Pay.Vcpaymxn
	if strings.TrimSpace(cfg.AppID) == "" || strings.TrimSpace(cfg.AppKey) == "" {
		return cfg, fmt.Errorf("VCPAYMXN callback config missing appId/appKey")
	}
	if strings.TrimSpace(req.AppID) != strings.TrimSpace(cfg.AppID) {
		return cfg, fmt.Errorf("app_id mismatch")
	}
	if strings.TrimSpace(req.Sign) == "" {
		return cfg, fmt.Errorf("sign is empty")
	}
	if !vcpaymxn.VerifyCallbackSignAny(buildVcpayMxnNotifyParams(req), cfg.AppKey) {
		return cfg, fmt.Errorf("sign invalid")
	}
	return cfg, nil
}

func buildVcpayMxnNotifyParams(req vcpayMxnNotifyReq) map[string]any {
	return map[string]any{
		"code":         req.Code.String(),
		"msg":          req.Msg,
		"app_id":       req.AppID,
		"nonce_str":    req.NonceStr,
		"notify_url":   req.NotifyURL,
		"order_amount": req.OrderAmount.Int64(),
		"order_fee":    req.OrderFee.Int64(),
		"out_trade_no": req.OutTradeNo,
		"trade_no":     req.TradeNo,
		"trade_state":  req.TradeState,
		"trade_type":   req.TradeType,
		"sign":         req.Sign,
	}
}

func centsToPayAmount(cents int64) float64 {
	return utils.Truncate2(float64(cents) / 100)
}

type vcpayMxnNotifyReq struct {
	Code        vcpayFlexString `json:"code"`
	Msg         string          `json:"msg"`
	AppID       string          `json:"app_id"`
	NonceStr    string          `json:"nonce_str"`
	NotifyURL   string          `json:"notify_url"`
	OrderAmount vcpayFlexInt    `json:"order_amount"`
	OrderFee    vcpayFlexInt    `json:"order_fee"`
	OutTradeNo  string          `json:"out_trade_no"`
	TradeNo     string          `json:"trade_no"`
	TradeState  int             `json:"trade_state"`
	TradeType   string          `json:"trade_type"`
	Sign        string          `json:"sign"`
}

type vcpayFlexString string

func (v *vcpayFlexString) UnmarshalJSON(data []byte) error {
	data = bytes.TrimSpace(data)
	if bytes.Equal(data, []byte("null")) {
		*v = ""
		return nil
	}
	var text string
	if err := json.Unmarshal(data, &text); err == nil {
		*v = vcpayFlexString(text)
		return nil
	}
	*v = vcpayFlexString(string(data))
	return nil
}

func (v vcpayFlexString) String() string {
	return strings.TrimSpace(string(v))
}

type vcpayFlexInt int64

func (v *vcpayFlexInt) UnmarshalJSON(data []byte) error {
	data = bytes.TrimSpace(data)
	if bytes.Equal(data, []byte("null")) || len(data) == 0 {
		*v = 0
		return nil
	}
	var n int64
	if err := json.Unmarshal(data, &n); err == nil {
		*v = vcpayFlexInt(n)
		return nil
	}
	var text string
	if err := json.Unmarshal(data, &text); err == nil {
		parsed, _ := strconv.ParseInt(strings.TrimSpace(text), 10, 64)
		*v = vcpayFlexInt(parsed)
	}
	return nil
}

func (v vcpayFlexInt) Int64() int64 {
	return int64(v)
}
