package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

const (
	DefaultITNioSMSBaseURL = "https://api.itniotech.com/sms"
	DefaultITNioSMSAPIKey  = "fWT85DCxgICxYnOmjBOVqIjWsLdewBTY"
	DefaultITNioSMSSecret  = "CqHjpubw7haracSJWeoqBaiu2K36ruh6"
	DefaultITNioSMSAppID   = "vBL27gMi"
	itNioSMSLimitPrefix    = "bgu_sms_limit_"
)

var smsCountryDialCodeMap = map[string]string{
	"BR": "55",
	"CN": "86",
	"ID": "62",
	"IN": "91",
	"MX": "52",
	"MY": "60",
	"PH": "63",
	"TH": "66",
	"US": "1",
	"VN": "84",
}

type ITNioSMSClient struct {
	BaseURL    string
	APIKey     string
	APIPwd     string
	AppID      string
	HTTPClient *http.Client
}

type ITNioSendSMSRequest struct {
	AppID    string `json:"appId"`
	Numbers  string `json:"numbers"`
	Content  string `json:"content"`
	SenderID string `json:"senderId,omitempty"`
	OrderID  string `json:"orderId,omitempty"`
}

func NewITNioSMSClient() *ITNioSMSClient {
	return &ITNioSMSClient{
		BaseURL: strings.TrimRight(DefaultITNioSMSBaseURL, "/"),
		APIKey:  DefaultITNioSMSAPIKey,
		APIPwd:  DefaultITNioSMSSecret,
		AppID:   DefaultITNioSMSAppID,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *ITNioSMSClient) SendSMS(numbers, content, senderID, orderID string) ([]byte, error) {
	req := ITNioSendSMSRequest{
		AppID:    c.AppID,
		Numbers:  numbers,
		Content:  content,
		SenderID: senderID,
		OrderID:  orderID,
	}
	return c.SendSMSWithRequest(req)
}

func (c *ITNioSMSClient) SendSMSWithRequest(payload ITNioSendSMSRequest) ([]byte, error) {
	if c == nil {
		return nil, fmt.Errorf("sms client is nil")
	}
	if payload.AppID == "" {
		payload.AppID = c.AppID
	}
	if payload.AppID == "" {
		return nil, fmt.Errorf("appId is required")
	}
	if payload.Numbers == "" {
		return nil, fmt.Errorf("numbers is required")
	}
	if payload.Content == "" {
		return nil, fmt.Errorf("content is required")
	}
	if c.APIKey == "" || c.APIPwd == "" {
		return nil, fmt.Errorf("api credentials are required")
	}
	if err := checkSMSRateLimit(payload.Numbers); err != nil {
		return nil, err
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	sign := GetMd5(c.APIKey + c.APIPwd + timestamp)
	url := strings.TrimRight(c.BaseURL, "/") + "/sendSms"

	httpClient := c.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}

	request, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	request.Header.Set("Connection", "Keep-Alive")
	request.Header.Set("Content-Type", "application/json;charset=UTF-8")
	request.Header.Set("Sign", sign)
	request.Header.Set("Timestamp", timestamp)
	request.Header.Set("Api-Key", c.APIKey)

	response, err := httpClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	respBody, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	log.Printf("sms response: %s", string(respBody))
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return respBody, fmt.Errorf("send sms failed: status=%d body=%s", response.StatusCode, string(respBody))
	}
	return respBody, nil
}

func checkSMSRateLimit(numbers string) error {
	if RD == nil {
		return fmt.Errorf("redis is not initialized")
	}
	for _, phone := range parseSMSNumbers(numbers) {
		limitKey := itNioSMSLimitPrefix + phone
		ok, err := RD.SetNX(context.Background(), limitKey, "1", time.Minute).Result()
		if err != nil {
			return fmt.Errorf("sms rate limit check failed: %w", err)
		}
		if !ok {
			return fmt.Errorf("phone %s requested too frequently, retry after 1 minute", phone)
		}
	}
	return nil
}

func parseSMSNumbers(numbers string) []string {
	replacer := strings.NewReplacer("\n", ",", "\r", ",", "\t", ",", ";", ",", "|", ",", " ", ",")
	normalized := replacer.Replace(strings.TrimSpace(numbers))
	parts := strings.Split(normalized, ",")
	result := make([]string, 0, len(parts))
	seen := make(map[string]struct{}, len(parts))
	for _, part := range parts {
		phone := strings.TrimSpace(part)
		if phone == "" {
			continue
		}
		if _, ok := seen[phone]; ok {
			continue
		}
		seen[phone] = struct{}{}
		result = append(result, phone)
	}
	return result
}

func NormalizeSMSPhone(country string, phone string) string {
	country = strings.TrimSpace(strings.ToUpper(country))
	phone = strings.TrimSpace(phone)
	if phone == "" {
		return phone
	}

	digits := onlyDigits(phone)
	if digits == "" {
		return phone
	}

	if strings.HasPrefix(digits, "00") {
		return strings.TrimPrefix(digits, "00")
	}

	dialCode := smsCountryDialCodeMap[country]
	if dialCode == "" {
		return digits
	}
	if strings.HasPrefix(digits, dialCode) {
		return digits
	}
	return dialCode + digits
}

func NormalizePhoneDigits(phone string) string {
	phone = strings.TrimSpace(phone)
	if phone == "" {
		return phone
	}

	digits := onlyDigits(phone)
	if strings.HasPrefix(digits, "00") {
		return strings.TrimPrefix(digits, "00")
	}
	return digits
}

func HasSupportedRegisterPhoneDialCode(phone string) bool {
	digits := NormalizePhoneDigits(phone)
	return strings.HasPrefix(digits, "55") || strings.HasPrefix(digits, "62") || strings.HasPrefix(digits, "52")
}

// InferCountryByPhone 根据手机号中的国际区号推断国家，推断失败时回退 fallbackCountry。
func InferCountryByPhone(phone string, fallbackCountry string) string {
	fallbackCountry = strings.TrimSpace(strings.ToUpper(fallbackCountry))
	trimmedPhone := strings.TrimSpace(phone)
	digits := onlyDigits(trimmedPhone)
	if digits == "" {
		return fallbackCountry
	}

	explicitIntl := strings.HasPrefix(trimmedPhone, "+") || strings.HasPrefix(digits, "00")
	if strings.HasPrefix(digits, "00") {
		digits = strings.TrimPrefix(digits, "00")
	}

	if country := matchCountryDialCode(digits); country != "" {
		if explicitIntl || fallbackCountry == "" {
			return country
		}
	}

	return fallbackCountry
}

func matchCountryDialCode(digits string) string {
	matchCountry := ""
	matchLen := 0
	for country, dialCode := range smsCountryDialCodeMap {
		if strings.HasPrefix(digits, dialCode) && len(dialCode) > matchLen {
			matchCountry = country
			matchLen = len(dialCode)
		}
	}
	return matchCountry
}

func onlyDigits(value string) string {
	var builder strings.Builder
	builder.Grow(len(value))
	for _, r := range value {
		if r >= '0' && r <= '9' {
			builder.WriteRune(r)
		}
	}
	return builder.String()
}
