package game

import (
	"BaseGoUni/core/utils"
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const (
	HeaderSign      = "X-Sign"
	HeaderRequestID = "X-Request-Id"
	HeaderAppID     = "X-Appid"

	ContentTypeJSON = "application/json;charset=UTF-8"
)

const (
	GameCodeSuccess                         = 0
	GameCodeOperatorDisabled                = 1001
	GameCodeInvalidAppID                    = 1002
	GameCodeAgentOrMultiCurrencyDisabled    = 1003
	GameCodeGameNotFound                    = 1004
	GameCodeGameMaintaining                 = 1005
	GameCodeGameClosed                      = 1006
	GameCodeGameHidden                      = 1007
	GameCodeEmptyUserID                     = 1008
	GameCodeInvalidWalletType               = 1009
	GameCodeInvalidMerchantCode             = 1011
	GameCodeCountryOrRegionRestricted       = 1012
	GameCodeAgentOrMultiCurrencyForbidden   = 1013
	GameCodeIPNotAllowed                    = 1014
	GameCodeInvalidRTP                      = 1015
	GameCodeInvalidTransferAmount           = 1016
	GameCodeOrderExists                     = 1017
	GameCodeOrderNotFound                   = 1018
	GameCodeTooFrequent                     = 1019
	GameCodeInvalidGameType                 = 1020
	GameCodeMerchantRTPControlDisabled      = 1021
	GameCodeMerchantNotApproved             = 1022
	GameCodeInsufficientBalance             = 1023
	GameCodeInvalidSwitchValue              = 1024
	GameCodeInvalidRTPEffectiveCount        = 1025
	GameCodeMaxMultiplierLEMinMultiplier    = 1026
	GameCodeBuyRTPPermissionDisabled        = 1027
	GameCodeInvalidPersonalMaxWin           = 1028
	GameCodeInvalidPersonalMaxMultiplier    = 1029
	GameCodeInvalidMonitorTypeOrSwitch      = 1030
	GameCodeInvalidMonitorNewbieRounds      = 1031
	GameCodeInvalidMonitorPlayerRTPRange    = 1032
	GameCodeInvalidMonitorStatsPeriod       = 1033
	GameCodeInvalidMonitorRTPAdjustRange    = 1034
	GameCodeInvalidNewbieControlProbability = 1035
	GameCodeBetHistoryPageSizeTooLarge      = 1036
	GameCodePlayerNotFound                  = 2001
	GameCodePlayerDisabled                  = 2002
)

var GameErrorMessages = map[int]string{
	GameCodeSuccess:                         "成功",
	GameCodeOperatorDisabled:                "运营商被禁用",
	GameCodeInvalidAppID:                    "无效的商户ID",
	GameCodeAgentOrMultiCurrencyDisabled:    "代理或多币种商户被禁用",
	GameCodeGameNotFound:                    "游戏未找到",
	GameCodeGameMaintaining:                 "该游戏正在进行维护",
	GameCodeGameClosed:                      "该游戏已关闭",
	GameCodeGameHidden:                      "该游戏已隐藏",
	GameCodeEmptyUserID:                     "用户ID为空",
	GameCodeInvalidWalletType:               "无效的钱包类型",
	GameCodeInvalidMerchantCode:             "无效的商户编码",
	GameCodeCountryOrRegionRestricted:       "您所在的国家或地区受到限制",
	GameCodeAgentOrMultiCurrencyForbidden:   "代理或多币种商户不允许调用接口",
	GameCodeIPNotAllowed:                    "IP不允许访问",
	GameCodeInvalidRTP:                      "错误的RTP赋值",
	GameCodeInvalidTransferAmount:           "错误的转账金额",
	GameCodeOrderExists:                     "订单已存在",
	GameCodeOrderNotFound:                   "订单不存在",
	GameCodeTooFrequent:                     "请求太频繁",
	GameCodeInvalidGameType:                 "无效的游戏类型",
	GameCodeMerchantRTPControlDisabled:      "未开启商户调控RTP开关",
	GameCodeMerchantNotApproved:             "商户未审核",
	GameCodeInsufficientBalance:             "余额不足",
	GameCodeInvalidSwitchValue:              "开关值错误",
	GameCodeInvalidRTPEffectiveCount:        "RTP生效次数值错误",
	GameCodeMaxMultiplierLEMinMultiplier:    "最大倍数值小于等于最小倍数值",
	GameCodeBuyRTPPermissionDisabled:        "购买RTP开关权限未开启",
	GameCodeInvalidPersonalMaxWin:           "个人最高赢分设置值错误",
	GameCodeInvalidPersonalMaxMultiplier:    "个人最高倍数设置值错误",
	GameCodeInvalidMonitorTypeOrSwitch:      "监控类型或监控开关值错误",
	GameCodeInvalidMonitorNewbieRounds:      "监控新手局数值错误",
	GameCodeInvalidMonitorPlayerRTPRange:    "监控玩家RTP误差范围值错误",
	GameCodeInvalidMonitorStatsPeriod:       "监控游戏内统计数据周期值错误",
	GameCodeInvalidMonitorRTPAdjustRange:    "监控增加RTP范围或减少RTP范围为空",
	GameCodeInvalidNewbieControlProbability: "监控新手非新手游戏调控的触发概率值设置错误",
	GameCodeBetHistoryPageSizeTooLarge:      "获取下注历史每页数据条数需要小于10000",
	GameCodePlayerNotFound:                  "玩家不存在",
	GameCodePlayerDisabled:                  "玩家被禁用",
}

type APIResponse[T any] struct {
	Code  int    `json:"code"`
	Error string `json:"error"`
	Data  T      `json:"data"`
}

func (r APIResponse[T]) Success() bool {
	return r.Code == GameCodeSuccess
}

type RequestHeaders struct {
	AppID     string
	RequestID string
	Sign      string
}

type Config struct {
	APIURL    string
	AppID     string
	AppSecret string
}

func Sign(requestID string, rawJSONBody []byte, secret string) string {
	sum := md5.Sum([]byte(requestID + string(rawJSONBody) + secret))
	return hex.EncodeToString(sum[:])
}

func VerifySign(requestID string, rawJSONBody []byte, secret string, sign string) bool {
	return Sign(requestID, rawJSONBody, secret) == sign
}

func NewRequestID() string {
	return fmt.Sprintf("%d_%06d", time.Now().UTC().UnixMilli(), randomSixDigits())
}

func BuildHeaders(appID string, requestID string, rawJSONBody []byte, secret string) RequestHeaders {
	return RequestHeaders{
		AppID:     appID,
		RequestID: requestID,
		Sign:      Sign(requestID, rawJSONBody, secret),
	}
}

func GetConfig() Config {
	cfg := utils.GlobalConfig.Game
	return Config{
		APIURL:    strings.TrimRight(strings.TrimSpace(cfg.APIURL), "/"),
		AppID:     strings.TrimSpace(cfg.AppID),
		AppSecret: strings.TrimSpace(cfg.AppSecret),
	}
}

func BuildConfiguredHeaders(rawJSONBody []byte) (RequestHeaders, error) {
	return BuildConfiguredHeadersWithRequestID(NewRequestID(), rawJSONBody)
}

func BuildConfiguredHeadersWithRequestID(requestID string, rawJSONBody []byte) (RequestHeaders, error) {
	cfg := GetConfig()
	if cfg.AppID == "" {
		return RequestHeaders{}, errors.New("game appId is empty")
	}
	if cfg.AppSecret == "" {
		return RequestHeaders{}, errors.New("game appSecret is empty")
	}
	return BuildHeaders(cfg.AppID, requestID, rawJSONBody, cfg.AppSecret), nil
}

func BuildURL(path string) (string, error) {
	cfg := GetConfig()
	return BuildURLWithConfig(cfg, path)
}

func BuildURLWithConfig(cfg Config, path string) (string, error) {
	cfg.APIURL = strings.TrimRight(strings.TrimSpace(cfg.APIURL), "/")
	if cfg.APIURL == "" {
		return "", errors.New("game apiUrl is empty")
	}
	if path == "" {
		return cfg.APIURL, nil
	}
	return cfg.APIURL + "/" + strings.TrimLeft(path, "/"), nil
}

func (h RequestHeaders) Apply(header http.Header) {
	header.Set(HeaderAppID, h.AppID)
	header.Set(HeaderRequestID, h.RequestID)
	header.Set(HeaderSign, h.Sign)
	header.Set("Content-Type", ContentTypeJSON)
}

func ErrorMessage(code int) string {
	if msg, ok := GameErrorMessages[code]; ok {
		return msg
	}
	return "未知错误"
}

func randomSixDigits() int {
	var b [4]byte
	if _, err := rand.Read(b[:]); err != nil {
		return int(time.Now().UTC().UnixNano() % 1000000)
	}
	n := int(b[0])<<24 | int(b[1])<<16 | int(b[2])<<8 | int(b[3])
	if n < 0 {
		n = -n
	}
	return n % 1000000
}
