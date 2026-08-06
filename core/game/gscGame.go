package game

import (
	"BaseGoUni/core/utils"
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	gscGameListPath        = "/api/operators/provider-games"
	gscLaunchGamePath      = "/api/operators/launch-game"
	gscGameListSignAction  = "gamelist"
	gscLaunchGameAction    = "launchgame"
	gscLaunchGameType      = "SPORT_BOOK"
	gscDefaultGameListSize = 500

	GSCOpenActionGetBalance  = "getbalance"
	GSCOpenActionWithdraw    = "withdraw"
	GSCOpenActionDeposit     = "deposit"
	GSCOpenActionPushBetData = "pushbetdata"
)

type GSCConfig struct {
	OperatorURL           string
	ProductCode           int
	OperatorCode          string
	SecretKey             string
	GameType              string
	SupportCurrency       string
	Currency              string
	Password              string
	LaunchPlatform        string
	LaunchLanguage        string
	OperatorLobbyURL      string
	DefaultMemberPassword string
	PageSize              int
	CategoryMap           map[string]string
}

type GSCClient struct {
	HTTPClient *http.Client
	Config     GSCConfig
	Now        func() time.Time
}

type GSCGameListResponse struct {
	Code          int               `json:"code"`
	Message       string            `json:"message"`
	ProviderGames []GSCProviderGame `json:"provider_games"`
	Pagination    GSCPagination     `json:"pagination"`
}

type GSCProviderGame struct {
	GameCode        string            `json:"game_code"`
	GameName        string            `json:"game_name"`
	GameType        string            `json:"game_type"`
	ImageURL        string            `json:"image_url"`
	ProductID       int               `json:"product_id"`
	ProductCode     int               `json:"product_code"`
	SupportCurrency string            `json:"support_currency"`
	Status          string            `json:"status"`
	AllowFreeRound  bool              `json:"allow_free_round"`
	LangName        map[string]string `json:"lang_name"`
	LangIcon        map[string]string `json:"lang_icon"`
	CreatedAt       int64             `json:"created_at"`
}

type GSCPagination struct {
	Size   int    `json:"size"`
	Offset int    `json:"offset"`
	Total  GSCInt `json:"total"`
}

type GSCLaunchGameInput struct {
	MemberAccount    string
	Password         string
	Nickname         string
	GameCode         *string
	GameType         string
	IP               string
	OperatorLobbyURL string
}

type GSCLaunchGameRequest struct {
	OperatorCode     string  `json:"operator_code"`
	MemberAccount    string  `json:"member_account"`
	Password         string  `json:"password"`
	Nickname         string  `json:"nickname,omitempty"`
	Currency         string  `json:"currency"`
	GameCode         *string `json:"game_code,omitempty"`
	ProductCode      int     `json:"product_code"`
	GameType         string  `json:"game_type"`
	LanguageCode     string  `json:"language_code"`
	IP               string  `json:"ip"`
	Platform         string  `json:"platform"`
	Sign             string  `json:"sign"`
	RequestTime      int64   `json:"request_time"`
	OperatorLobbyURL string  `json:"operator_lobby_url"`
}

type GSCLaunchGameResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	URL     string `json:"url"`
	Content string `json:"content"`
}

type GSCInt int

func (v *GSCInt) UnmarshalJSON(data []byte) error {
	raw := strings.TrimSpace(string(data))
	raw = strings.Trim(raw, `"`)
	if raw == "" || strings.EqualFold(raw, "null") {
		*v = 0
		return nil
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		return err
	}
	*v = GSCInt(n)
	return nil
}

func GetGSCConfig() GSCConfig {
	cfg := utils.GlobalConfig.GSCGame
	result := GSCConfig{
		OperatorURL:           strings.TrimRight(strings.TrimSpace(cfg.OperatorURL), "/"),
		ProductCode:           cfg.ProductCode,
		OperatorCode:          strings.TrimSpace(cfg.OperatorCode),
		SecretKey:             strings.TrimSpace(cfg.SecretKey),
		GameType:              strings.TrimSpace(cfg.GameType),
		SupportCurrency:       strings.TrimSpace(cfg.SupportCurrency),
		Currency:              strings.TrimSpace(cfg.Currency),
		Password:              strings.TrimSpace(cfg.Password),
		LaunchPlatform:        strings.TrimSpace(cfg.LaunchPlatform),
		LaunchLanguage:        strings.TrimSpace(cfg.LaunchLanguage),
		OperatorLobbyURL:      strings.TrimSpace(cfg.OperatorLobbyURL),
		DefaultMemberPassword: strings.TrimSpace(cfg.DefaultMemberPassword),
		PageSize:              cfg.PageSize,
		CategoryMap:           cfg.CategoryMap,
	}
	if result.PageSize <= 0 {
		result.PageSize = gscDefaultGameListSize
	}
	if result.Currency == "" {
		result.Currency = result.SupportCurrency
	}
	if result.Currency == "" {
		result.Currency = "PEN"
	}
	if result.LaunchPlatform == "" {
		result.LaunchPlatform = "WEB"
	}
	if result.LaunchLanguage == "" {
		result.LaunchLanguage = "0"
	}
	if result.CategoryMap == nil {
		result.CategoryMap = map[string]string{}
	}
	return result
}

func NewGSCClient() GSCClient {
	return NewGSCClientWithConfig(GetGSCConfig())
}

func NewGSCClientWithConfig(cfg GSCConfig) GSCClient {
	cfg.OperatorURL = strings.TrimRight(strings.TrimSpace(cfg.OperatorURL), "/")
	cfg.OperatorCode = strings.TrimSpace(cfg.OperatorCode)
	cfg.SecretKey = strings.TrimSpace(cfg.SecretKey)
	cfg.GameType = strings.TrimSpace(cfg.GameType)
	cfg.SupportCurrency = strings.TrimSpace(cfg.SupportCurrency)
	cfg.Currency = strings.TrimSpace(cfg.Currency)
	cfg.Password = strings.TrimSpace(cfg.Password)
	cfg.LaunchPlatform = strings.TrimSpace(cfg.LaunchPlatform)
	cfg.LaunchLanguage = strings.TrimSpace(cfg.LaunchLanguage)
	cfg.OperatorLobbyURL = strings.TrimSpace(cfg.OperatorLobbyURL)
	cfg.DefaultMemberPassword = strings.TrimSpace(cfg.DefaultMemberPassword)
	if cfg.PageSize <= 0 {
		cfg.PageSize = gscDefaultGameListSize
	}
	if cfg.Currency == "" {
		cfg.Currency = cfg.SupportCurrency
	}
	if cfg.Currency == "" {
		cfg.Currency = "PEN"
	}
	if cfg.LaunchPlatform == "" {
		cfg.LaunchPlatform = "WEB"
	}
	if cfg.LaunchLanguage == "" {
		cfg.LaunchLanguage = "0"
	}
	if cfg.CategoryMap == nil {
		cfg.CategoryMap = map[string]string{}
	}
	return GSCClient{
		HTTPClient: http.DefaultClient,
		Config:     cfg,
		Now:        time.Now,
	}
}

func GSCSign(requestTime int64, secretKey string, operatorCode string) string {
	return gscActionSign(requestTime, secretKey, gscGameListSignAction, operatorCode)
}

func GSCLaunchSign(requestTime int64, secretKey string, operatorCode string) string {
	return gscActionSign(requestTime, secretKey, gscLaunchGameAction, operatorCode)
}

func GSCOpenSign(operatorCode string, requestTime string, action string, secretKey string) string {
	sum := md5.Sum([]byte(strings.TrimSpace(operatorCode) + strings.TrimSpace(requestTime) + strings.TrimSpace(action) + strings.TrimSpace(secretKey)))
	return hex.EncodeToString(sum[:])
}

func gscActionSign(requestTime int64, secretKey string, action string, operatorCode string) string {
	sum := md5.Sum([]byte(strconv.FormatInt(requestTime, 10) + secretKey + action + operatorCode))
	return hex.EncodeToString(sum[:])
}

func (c GSCClient) GameList() ([]GSCProviderGame, error) {
	if err := c.validateConfig(); err != nil {
		return nil, err
	}

	size := c.Config.PageSize
	if size <= 0 {
		size = gscDefaultGameListSize
	}

	offset := 0
	result := make([]GSCProviderGame, 0)
	for {
		page, err := c.GameListPage(offset, size)
		if err != nil {
			return nil, err
		}
		if page.Code != 0 {
			msg := strings.TrimSpace(page.Message)
			if msg == "" {
				msg = "gsc game api error"
			}
			return nil, fmt.Errorf("gsc game api error code=%d message=%s", page.Code, msg)
		}

		result = append(result, page.ProviderGames...)
		returned := len(page.ProviderGames)
		total := int(page.Pagination.Total)
		if returned == 0 || returned < size {
			break
		}
		if total > 0 && len(result) >= total {
			break
		}
		offset += size
	}
	return result, nil
}

func (c GSCClient) GameListPage(offset int, size int) (GSCGameListResponse, error) {
	var empty GSCGameListResponse
	if err := c.validateConfig(); err != nil {
		return empty, err
	}
	if size <= 0 {
		size = gscDefaultGameListSize
	}
	if offset < 0 {
		offset = 0
	}

	requestTime := c.now().Unix()
	endpoint, err := c.buildGameListURL(offset, size, requestTime)
	if err != nil {
		return empty, err
	}
	log.Printf("[gsc_game] GameListPage request url=%s product_code=%d operator_code=%s request_time=%d sign=%s offset=%d size=%d game_type=%s",
		endpoint,
		c.Config.ProductCode,
		strings.TrimSpace(c.Config.OperatorCode),
		requestTime,
		GSCSign(requestTime, strings.TrimSpace(c.Config.SecretKey), strings.TrimSpace(c.Config.OperatorCode)),
		offset,
		size,
		strings.TrimSpace(c.Config.GameType),
	)

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return empty, err
	}

	httpClient := c.HTTPClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return empty, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return empty, err
	}
	log.Printf("[gsc_game] GameListPage response status=%d body=%s", resp.StatusCode, string(body))
	if resp.StatusCode != http.StatusOK {
		return empty, fmt.Errorf("gsc game api http status=%d body=%s", resp.StatusCode, string(body))
	}

	var result GSCGameListResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return empty, err
	}
	return result, nil
}

func (c GSCClient) LaunchGame(input GSCLaunchGameInput) (GSCLaunchGameResponse, error) {
	var empty GSCLaunchGameResponse
	if err := c.validateLaunchConfig(); err != nil {
		return empty, err
	}

	input.MemberAccount = strings.TrimSpace(input.MemberAccount)
	input.Password = strings.TrimSpace(c.Config.Password)
	input.Nickname = strings.TrimSpace(input.Nickname)
	input.GameType = gscLaunchGameType
	input.IP = strings.TrimSpace(input.IP)
	operatorLobbyURL := strings.TrimSpace(c.Config.OperatorLobbyURL)
	if operatorLobbyURL == "" {
		operatorLobbyURL = strings.TrimSpace(input.OperatorLobbyURL)
	}
	if input.MemberAccount == "" {
		return empty, errors.New("gsc launch memberAccount is empty")
	}
	if input.Password == "" {
		input.Password = strings.TrimSpace(c.Config.DefaultMemberPassword)
	}
	if input.Password == "" {
		return empty, errors.New("gsc launch password is empty")
	}
	if input.GameType == "" {
		return empty, errors.New("gsc launch gameType is empty")
	}
	if input.IP == "" {
		return empty, errors.New("gsc launch ip is empty")
	}
	if operatorLobbyURL == "" {
		return empty, errors.New("gscGame operatorLobbyUrl is empty")
	}

	requestTime := c.now().Unix()
	payload := GSCLaunchGameRequest{
		OperatorCode:     strings.TrimSpace(c.Config.OperatorCode),
		MemberAccount:    input.MemberAccount,
		Password:         input.Password,
		Nickname:         input.Nickname,
		Currency:         strings.TrimSpace(c.Config.Currency),
		GameCode:         nil,
		ProductCode:      c.Config.ProductCode,
		GameType:         input.GameType,
		LanguageCode:     strings.TrimSpace(c.Config.LaunchLanguage),
		IP:               input.IP,
		Platform:         strings.TrimSpace(c.Config.LaunchPlatform),
		RequestTime:      requestTime,
		Sign:             GSCLaunchSign(requestTime, strings.TrimSpace(c.Config.SecretKey), strings.TrimSpace(c.Config.OperatorCode)),
		OperatorLobbyURL: operatorLobbyURL,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return empty, err
	}
	endpoint, err := c.buildLaunchGameURL()
	if err != nil {
		return empty, err
	}
	log.Printf("[gsc_game] LaunchGame request url=%s body=%s", endpoint, gscLaunchLogBody(payload))
	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return empty, err
	}
	req.Header.Set("Content-Type", "application/json")

	httpClient := c.HTTPClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return empty, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return empty, err
	}
	log.Printf("[gsc_game] LaunchGame response status=%d body=%s", resp.StatusCode, string(respBody))
	if resp.StatusCode != http.StatusOK {
		return empty, fmt.Errorf("gsc launch api http status=%d body=%s", resp.StatusCode, string(respBody))
	}

	var result GSCLaunchGameResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return empty, err
	}
	return result, nil
}

func (c GSCClient) validateConfig() error {
	if strings.TrimSpace(c.Config.OperatorURL) == "" {
		return errors.New("gscGame operatorUrl is empty")
	}
	if c.Config.ProductCode <= 0 {
		return errors.New("gscGame productCode is empty")
	}
	if strings.TrimSpace(c.Config.OperatorCode) == "" {
		return errors.New("gscGame operatorCode is empty")
	}
	if strings.TrimSpace(c.Config.SecretKey) == "" {
		return errors.New("gscGame secretKey is empty")
	}
	return nil
}

func gscLaunchLogBody(payload GSCLaunchGameRequest) string {
	payload.Password = "***"
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Sprintf("marshal_error=%v", err)
	}
	return string(body)
}

func (c GSCClient) validateLaunchConfig() error {
	if err := c.validateConfig(); err != nil {
		return err
	}
	if strings.TrimSpace(c.Config.Currency) == "" {
		return errors.New("gscGame currency is empty")
	}
	if strings.TrimSpace(c.Config.LaunchPlatform) == "" {
		return errors.New("gscGame launchPlatform is empty")
	}
	if strings.TrimSpace(c.Config.LaunchLanguage) == "" {
		return errors.New("gscGame launchLanguage is empty")
	}
	return nil
}

func (c GSCClient) buildGameListURL(offset int, size int, requestTime int64) (string, error) {
	endpoint := strings.TrimRight(strings.TrimSpace(c.Config.OperatorURL), "/") + gscGameListPath
	parsed, err := url.Parse(endpoint)
	if err != nil {
		return "", err
	}

	query := parsed.Query()
	query.Set("product_code", strconv.Itoa(c.Config.ProductCode))
	query.Set("operator_code", strings.TrimSpace(c.Config.OperatorCode))
	query.Set("request_time", strconv.FormatInt(requestTime, 10))
	query.Set("sign", GSCSign(requestTime, strings.TrimSpace(c.Config.SecretKey), strings.TrimSpace(c.Config.OperatorCode)))
	query.Set("offset", strconv.Itoa(offset))
	query.Set("size", strconv.Itoa(size))
	if gameType := strings.TrimSpace(c.Config.GameType); gameType != "" {
		query.Set("game_type", gameType)
	}
	parsed.RawQuery = query.Encode()
	return parsed.String(), nil
}

func (c GSCClient) buildLaunchGameURL() (string, error) {
	endpoint := strings.TrimRight(strings.TrimSpace(c.Config.OperatorURL), "/") + gscLaunchGamePath
	parsed, err := url.Parse(endpoint)
	if err != nil {
		return "", err
	}
	return parsed.String(), nil
}

func (c GSCClient) now() time.Time {
	if c.Now != nil {
		return c.Now()
	}
	return time.Now()
}
