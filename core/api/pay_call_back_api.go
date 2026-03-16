package api

import (
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

// GctpkPayinCallback GCTPK 代收异步回调（公开接口，无需 token）
// POST /api/v1/pay/gctpk/notify
func GctpkPayinCallback(ctx *gin.Context) {
	var req gctpkNotifyReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.String(400, "FAIL")
		return
	}

	cfg := utils.GlobalConfig.Pay.Gctpk

	// 验签：将回调参数（除 sign 外所有非空字段）按规则签名后比对
	params := map[string]string{
		"payAmount":  req.PayAmount,
		"orderNo":    req.OrderNo,
		"merNo":      req.MerNo,
		"merOrderNo": req.MerOrderNo,
		"payTime":    req.PayTime,
		"busiCode":   req.BusiCode,
		"status":     fmt.Sprintf("%d", req.Status),
	}
	expectSign := gctpk.BuildSign(params, cfg.Secret)
	if !strings.EqualFold(expectSign, strings.TrimSpace(req.Sign)) {
		log.Printf("[GCTPK Notify] 验签失败 orderNo=%s expect=%s got=%s", req.OrderNo, expectSign, req.Sign)
		ctx.String(400, "FAIL")
		return
	}

	db := ctx.MustGet("db").(*gorm.DB)
	hostInfo := ctx.MustGet("hostInfo").(pojo.HostInfo)

	switch req.Status {
	case 5: // 支付成功
		if err := repository.ProcessRechargeOrderSuccess(
			db, req.OrderNo, req.MerOrderNo, parsePayFloat(req.PayAmount), hostInfo.TablePrefix,
		); err != nil {
			log.Printf("[GCTPK Notify] 入账失败 orderNo=%s err=%v", req.OrderNo, err)
			ctx.String(500, "FAIL")
			return
		}
	case 4: // 订单已关闭
		if err := repository.ProcessRechargeOrderClosed(db, req.OrderNo); err != nil {
			log.Printf("[GCTPK Notify] 关闭订单失败 orderNo=%s err=%v", req.OrderNo, err)
			ctx.String(500, "FAIL")
			return
		}
	default:
		// 其他状态不处理，直接返回 SUCCESS 防止重复通知
	}

	ctx.String(200, "SUCCESS")
}

// gctpkNotifyReq GCTPK 代收回调请求体
type gctpkNotifyReq struct {
	PayAmount  string `json:"payAmount"`  // 实际支付金额
	OrderNo    string `json:"orderNo"`    // 平台订单号
	MerNo      string `json:"merNo"`      // 商户号
	MerOrderNo string `json:"merOrderNo"` // 商户订单号
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
