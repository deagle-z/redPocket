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
	AutoSuccess     bool   // 测试通道使用：创建本地订单后立即按支付成功处理
}

// Provider 代收支付渠道策略接口
type Provider interface {
	// Name 返回渠道名称，需与 sys_pay_channel.channel_code 一致（大写）
	Name() string
	// CreateOrder 调用三方创建代收订单
	CreateOrder(req PayRequest) (PayResponse, error)
}

// PayoutRequest 代付下单请求参数
type PayoutRequest struct {
	OrderNo     string
	Amount      float64
	Currency    string
	AccName     string            // 收款人姓名
	AccNo       string            // 收款账号/卡号
	BankCode    string            // 银行编码
	BusiCode    string            // 支付业务编码
	Email       string            // 邮箱
	Phone       string            // 手机号
	NotifyURL   string            // 回调地址
	ExtraFields map[string]string // extend/province 等扩展字段
}

// PayoutResponse 代付下单响应
type PayoutResponse struct {
	ProviderOrderNo string // 三方订单号（可空）
	Status          int    // 三方返回的订单状态
}

// PayoutProvider 代付渠道策略接口（可选实现，代收渠道不必实现此接口）
type PayoutProvider interface {
	Provider
	// CreatePayoutOrder 调用三方创建代付订单
	CreatePayoutOrder(req PayoutRequest) (PayoutResponse, error)
}

// GetPayout 根据渠道名称获取 PayoutProvider；不支持代付则返回 nil
func GetPayout(channel string) PayoutProvider {
	p := Get(channel)
	if p == nil {
		return nil
	}
	pp, _ := p.(PayoutProvider)
	return pp
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
