package pay

// YoyopayProvider YOYOPAY 测试支付渠道
type YoyopayProvider struct {
}

func init() {
	Register(&YoyopayProvider{})
}

func (y *YoyopayProvider) Name() string {
	return "YOYOPAY"
}

func (y *YoyopayProvider) CreateOrder(req PayRequest) (PayResponse, error) {
	return PayResponse{
		ProviderTradeNo: "YOYOPAY_TEST_" + req.OrderNo,
		AutoSuccess:     true,
	}, nil
}
