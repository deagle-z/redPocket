package pay

// ManualProvider 手动入账渠道（无需跳转三方，dev 环境自动回调，生产环境由后台手动审核）
type ManualProvider struct{}

func init() {
	Register(&ManualProvider{})
}

func (m *ManualProvider) Name() string {
	return "MANUAL"
}

func (m *ManualProvider) CreateOrder(req PayRequest) (PayResponse, error) {
	// 不调用任何三方，返回空 PayURL
	// dev 环境由 AppCreateRechargeOrder 检测后触发自动回调
	return PayResponse{}, nil
}
