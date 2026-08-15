package game

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"BaseGoUni/core/utils"
)

const (
	GGRMethodProviderList = "provider_list"
	GGRMethodGameList     = "game_list"
	GGRMethodGameLaunch   = "game_launch"
)

type GGRConfig struct {
	APIURL      string
	AgentCode   string
	AgentToken  string
	AgentSecret string
	CategoryMap map[string]string
}

type GGRClient struct {
	HTTPClient *http.Client
	Config     GGRConfig
}

type GGRProvider struct {
	Code   string `json:"code"`
	Name   string `json:"name"`
	Status int    `json:"status"`
}

type GGRProviderListResponse struct {
	Status    int           `json:"status"`
	Message   string        `json:"msg"`
	Providers []GGRProvider `json:"providers"`
}

type GGRGame struct {
	GameCode string `json:"game_code"`
	GameName string `json:"game_name"`
	Banner   string `json:"banner"`
	Status   int    `json:"status"`
}

type GGRGameListResponse struct {
	Status  int       `json:"status"`
	Message string    `json:"msg"`
	Games   []GGRGame `json:"games"`
}

type GGRGameLaunchInput struct {
	UserCode     string
	ProviderCode string
	GameCode     string
	Language     string
	LobbyURL     string
}

type GGRGameLaunchResponse struct {
	Status    int    `json:"status"`
	Message   string `json:"msg"`
	LaunchURL string `json:"launch_url"`
}

type ggrProviderListRequest struct {
	Method     string `json:"method"`
	AgentCode  string `json:"agent_code"`
	AgentToken string `json:"agent_token"`
}

type ggrGameListRequest struct {
	Method       string `json:"method"`
	AgentCode    string `json:"agent_code"`
	AgentToken   string `json:"agent_token"`
	ProviderCode string `json:"provider_code"`
}

type ggrGameLaunchRequest struct {
	Method       string `json:"method"`
	AgentCode    string `json:"agent_code"`
	AgentToken   string `json:"agent_token"`
	UserCode     string `json:"user_code"`
	ProviderCode string `json:"provider_code"`
	GameCode     string `json:"game_code"`
	Language     string `json:"lang"`
	LobbyURL     string `json:"lobby_url,omitempty"`
}

func GetGGRConfig() GGRConfig {
	cfg := utils.GlobalConfig.GGRGame
	return normalizeGGRConfig(GGRConfig{
		APIURL:      cfg.APIURL,
		AgentCode:   cfg.AgentCode,
		AgentToken:  cfg.AgentToken,
		AgentSecret: cfg.AgentSecret,
		CategoryMap: cfg.CategoryMap,
	})
}

func NewGGRClient() GGRClient {
	return NewGGRClientWithConfig(GetGGRConfig())
}

func NewGGRClientWithConfig(cfg GGRConfig) GGRClient {
	return GGRClient{
		HTTPClient: http.DefaultClient,
		Config:     normalizeGGRConfig(cfg),
	}
}

func normalizeGGRConfig(cfg GGRConfig) GGRConfig {
	cfg.APIURL = strings.TrimRight(strings.TrimSpace(cfg.APIURL), "/")
	cfg.AgentCode = strings.TrimSpace(cfg.AgentCode)
	cfg.AgentToken = strings.TrimSpace(cfg.AgentToken)
	cfg.AgentSecret = strings.TrimSpace(cfg.AgentSecret)
	if cfg.CategoryMap == nil {
		cfg.CategoryMap = map[string]string{}
	}
	return cfg
}

func (c GGRClient) ProviderList() (GGRProviderListResponse, error) {
	var result GGRProviderListResponse
	err := c.postJSON(ggrProviderListRequest{
		Method:     GGRMethodProviderList,
		AgentCode:  c.Config.AgentCode,
		AgentToken: c.Config.AgentToken,
	}, &result)
	if err != nil {
		return result, err
	}
	if result.Status != 1 {
		return result, ggrStatusError(result.Status, result.Message)
	}
	return result, nil
}

func (c GGRClient) GameList(providerCode string) (GGRGameListResponse, error) {
	providerCode = strings.ToUpper(strings.TrimSpace(providerCode))
	if providerCode == "" {
		return GGRGameListResponse{}, errors.New("ggr provider_code is empty")
	}

	var result GGRGameListResponse
	err := c.postJSON(ggrGameListRequest{
		Method:       GGRMethodGameList,
		AgentCode:    c.Config.AgentCode,
		AgentToken:   c.Config.AgentToken,
		ProviderCode: providerCode,
	}, &result)
	if err != nil {
		return result, err
	}
	if result.Status != 1 {
		return result, ggrStatusError(result.Status, result.Message)
	}
	return result, nil
}

func (c GGRClient) GameLaunch(input GGRGameLaunchInput) (GGRGameLaunchResponse, error) {
	input.UserCode = strings.TrimSpace(input.UserCode)
	input.ProviderCode = strings.ToUpper(strings.TrimSpace(input.ProviderCode))
	input.GameCode = strings.TrimSpace(input.GameCode)
	input.Language = strings.TrimSpace(input.Language)
	input.LobbyURL = strings.TrimSpace(input.LobbyURL)
	if input.UserCode == "" {
		return GGRGameLaunchResponse{}, errors.New("ggr user_code is empty")
	}
	if input.ProviderCode == "" {
		return GGRGameLaunchResponse{}, errors.New("ggr provider_code is empty")
	}
	if input.GameCode == "" {
		return GGRGameLaunchResponse{}, errors.New("ggr game_code is empty")
	}
	if input.Language == "" {
		return GGRGameLaunchResponse{}, errors.New("ggr lang is empty")
	}

	var result GGRGameLaunchResponse
	err := c.postJSON(ggrGameLaunchRequest{
		Method:       GGRMethodGameLaunch,
		AgentCode:    c.Config.AgentCode,
		AgentToken:   c.Config.AgentToken,
		UserCode:     input.UserCode,
		ProviderCode: input.ProviderCode,
		GameCode:     input.GameCode,
		Language:     input.Language,
		LobbyURL:     input.LobbyURL,
	}, &result)
	if err != nil {
		return result, err
	}
	if result.Status != 1 {
		return result, ggrStatusError(result.Status, result.Message)
	}
	if strings.TrimSpace(result.LaunchURL) == "" {
		return result, errors.New("ggr launch_url is empty")
	}
	return result, nil
}

func (c GGRClient) postJSON(payload any, result any) error {
	if c.Config.APIURL == "" {
		return errors.New("ggr apiUrl is empty")
	}
	if c.Config.AgentCode == "" {
		return errors.New("ggr agentCode is empty")
	}
	if c.Config.AgentToken == "" {
		return errors.New("ggr agentToken is empty")
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("encode ggr request: %w", err)
	}
	req, err := http.NewRequest(http.MethodPost, c.Config.APIURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create ggr request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	httpClient := c.HTTPClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("call ggr api: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		_, _ = io.Copy(io.Discard, resp.Body)
		return fmt.Errorf("ggr api http status=%d", resp.StatusCode)
	}
	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return fmt.Errorf("decode ggr response: %w", err)
	}
	return nil
}

func ggrStatusError(status int, message string) error {
	message = strings.TrimSpace(message)
	if message == "" {
		return fmt.Errorf("ggr api status=%d", status)
	}
	return fmt.Errorf("ggr api status=%d message=%s", status, message)
}

// ResolveGGRProviderCategory returns the local category configured for a GGR provider code.
// Unknown providers and invalid category values are rejected instead of being assigned a fallback category.
func ResolveGGRProviderCategory(providerCode string) (string, error) {
	return ResolveGGRProviderCategoryWithMap(providerCode, GetGGRConfig().CategoryMap)
}

func ResolveGGRProviderCategoryWithMap(providerCode string, categoryMap map[string]string) (string, error) {
	providerCode = strings.ToUpper(strings.TrimSpace(providerCode))
	if providerCode == "" {
		return "", errors.New("ggr provider_code is empty")
	}

	categoryCode := strings.ToLower(strings.TrimSpace(categoryMap[providerCode]))
	if categoryCode == "" {
		return "", fmt.Errorf("ggr provider category is not configured: %s", providerCode)
	}

	switch categoryCode {
	case "slots", "casino", "sports", "mini", "fishing":
		return categoryCode, nil
	default:
		return "", fmt.Errorf("ggr provider category is invalid: provider=%s category=%s", providerCode, categoryCode)
	}
}
