package pay

import (
	"fmt"
	"strings"
	"sync"
)

// PayRequest 发起支付的请求参数
type PayRequest struct {
	OrderNo     string
	Amount      float64
	Currency    string
	PayMethod   string
	CountryCode string
	ExtraFields map[string]string
	NotifyURL   string
	ReturnURL   string
}

// PayResponse 三方支付创单响应
type PayResponse struct {
	PayURL          string // 跳转支付页面的 URL，为空表示无需跳转（手动入账）
	ProviderTradeNo string // 三方交易号，可空
}

// Provider 支付渠道策略接口
type Provider interface {
	// Name 返回渠道名称，需与 sys_pay_channel.channel_code 一致（大写）
	Name() string
	// CreateOrder 调用三方创建支付订单
	CreateOrder(req PayRequest) (PayResponse, error)
}

var (
	mu        sync.RWMutex
	providers = make(map[string]Provider)
)

// Register 注册一个支付渠道 Provider（建议在各 provider 文件的 init() 中调用）
func Register(p Provider) {
	mu.Lock()
	defer mu.Unlock()
	providers[strings.ToUpper(p.Name())] = p
}

// Get 根据渠道名称（不区分大小写）获取 Provider；未找到时返回 nil
func Get(channel string) Provider {
	mu.RLock()
	defer mu.RUnlock()
	return providers[strings.ToUpper(channel)]
}

// MustGet 获取 Provider，未注册则返回 error
func MustGet(channel string) (Provider, error) {
	p := Get(channel)
	if p == nil {
		return nil, fmt.Errorf("未找到支付渠道: %s", channel)
	}
	return p, nil
}
