package api

import (
	"BaseGoUni/core/base"
	"BaseGoUni/core/pay/gctpk"
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/repository"
	"BaseGoUni/core/utils"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"log"
	"strings"
)

type gctpkConfigResolver func(req gctpkNotifyReq) (base.GctpkPayConfig, string, error)

// GctpkPayinCallback GCTPK 代收异步回调（公开接口，无需 token）
// POST /api/v1/pay/gctpk/notify
func GctpkPayinCallback(ctx *gin.Context) {
	handleGctpkPayinCallback(ctx, resolveAnyGctpkPayConfig)
}

// GctpkMxnPayinCallback GCTPKMXN 代收异步回调（公开接口，无需 token）
// POST /api/v1/pay/gctpkmxn/notify
func GctpkMxnPayinCallback(ctx *gin.Context) {
	handleGctpkPayinCallback(ctx, func(req gctpkNotifyReq) (base.GctpkPayConfig, string, error) {
		return validateGctpkPayConfig(req, utils.GlobalConfig.Pay.Gctpkmxn, "GCTPKMXN")
	})
}

// GctpkBrlPayinCallback GCTPKBRL 代收异步回调（公开接口，无需 token）
// POST /api/v1/pay/gctpkbrl/notify
func GctpkBrlPayinCallback(ctx *gin.Context) {
	handleGctpkPayinCallback(ctx, func(req gctpkNotifyReq) (base.GctpkPayConfig, string, error) {
		return validateGctpkPayConfig(req, utils.GlobalConfig.Pay.Gctpkbrl, "GCTPKBRL")
	})
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

func validateGctpkPayConfig(req gctpkNotifyReq, cfg base.GctpkPayConfig, providerName string) (base.GctpkPayConfig, string, error) {
	if strings.TrimSpace(cfg.Secret) == "" {
		return base.GctpkPayConfig{}, providerName, fmt.Errorf("%s 回调密钥为空", providerName)
	}
	if merNo := strings.TrimSpace(req.MerNo); merNo != "" && strings.TrimSpace(cfg.MerNo) != "" && !strings.EqualFold(strings.TrimSpace(cfg.MerNo), merNo) {
		return base.GctpkPayConfig{}, providerName, fmt.Errorf("%s 商户号不匹配", providerName)
	}

	expectSign := gctpk.BuildSign(buildGctpkNotifyParams(req), cfg.Secret)
	if !strings.EqualFold(expectSign, strings.TrimSpace(req.Sign)) {
		return base.GctpkPayConfig{}, providerName, fmt.Errorf("%s 签名不匹配", providerName)
	}

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

func parsePayFloat(s string) float64 {
	var v float64
	fmt.Sscanf(s, "%f", &v)
	return v
}
