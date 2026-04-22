package pay

import "fmt"

// YoyopayProvider YOYOPAY 支付渠道
// TODO: 填写真实的 API 地址、签名方式、请求结构
type YoyopayProvider struct {
	// AppID  string
	// Secret string
	// APIURL string
}

func init() {
	Register(&YoyopayProvider{})
}

func (y *YoyopayProvider) Name() string {
	return "YOYOPAY"
}

func (y *YoyopayProvider) CreateOrder(req PayRequest) (PayResponse, error) {
	// TODO: 构造 YOYOPAY 签名请求，发起 HTTP 调用，解析返回的 payUrl
	//
	// 示例（伪代码）：
	//   body := map[string]any{
	//       "mch_id":    y.AppID,
	//       "order_no":  req.OrderNo,
	//       "amount":    req.Amount,
	//       "currency":  req.Currency,
	//       "pay_type":  req.PayMethod,
	//       "notify_url": req.NotifyURL,
	//   }
	//   sign := yoyopaySign(body, y.Secret)
	//   resp := callYoyopayAPI(y.APIURL+"/createOrder", body, sign)
	//   return PayResponse{PayURL: resp.PayURL, ProviderTradeNo: resp.TradeNo}, nil

	return PayResponse{}, fmt.Errorf("YOYOPAY provider 未实现，禁止使用该支付渠道")
}
