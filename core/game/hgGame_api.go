package game

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type Client struct {
	HTTPClient *http.Client
	Config     Config
}

type GameListRequest struct {
	Language string `json:"language"`
}

type GameInfo struct {
	ID              string `json:"id"`
	GameID          string `json:"gameid"`
	Name            string `json:"name"`
	Type            int    `json:"type"`
	IconURL         string `json:"icon_url"`
	Manufacturer    string `json:"manufacturer"`
	PortraitIconURL string `json:"portrait_icon_url"`
}

type GameListData struct {
	List  []GameInfo `json:"list"`
	GList []GameInfo `json:"glist"`
}

func (d GameListData) Games() []GameInfo {
	if len(d.List) > 0 {
		return d.List
	}
	return d.GList
}

type GameLaunchRequest struct {
	UserID   string `json:"userid"`
	GameID   string `json:"gameid"`
	Language string `json:"language"`
}

type GameLaunchData struct {
	URL string `json:"url"`
}

func NewClient() Client {
	return Client{
		HTTPClient: http.DefaultClient,
		Config:     GetConfig(),
	}
}

func NewClientWithConfig(cfg Config) Client {
	cfg.APIURL = strings.TrimRight(strings.TrimSpace(cfg.APIURL), "/")
	cfg.AppID = strings.TrimSpace(cfg.AppID)
	cfg.AppSecret = strings.TrimSpace(cfg.AppSecret)
	return Client{
		HTTPClient: http.DefaultClient,
		Config:     cfg,
	}
}

func (c Client) GameList(language string) (APIResponse[GameListData], error) {
	resp, err := c.PostJSON("/api/v1/game/list", GameListRequest{Language: language})
	if err != nil {
		return APIResponse[GameListData]{}, err
	}
	return DecodeData[GameListData](resp)
}

func (c Client) GameLaunch(userID string, gameID string, language string) (APIResponse[GameLaunchData], error) {
	resp, err := c.PostJSON("/api/v1/game/launch", GameLaunchRequest{
		UserID:   userID,
		GameID:   gameID,
		Language: language,
	})
	if err != nil {
		return APIResponse[GameLaunchData]{}, err
	}
	return DecodeData[GameLaunchData](resp)
}

func (c Client) PostJSON(path string, payload any) (APIResponse[json.RawMessage], error) {
	var empty APIResponse[json.RawMessage]
	body, err := json.Marshal(payload)
	if err != nil {
		return empty, err
	}

	url, err := BuildURLWithConfig(c.Config, path)
	if err != nil {
		return empty, err
	}
	if c.Config.AppID == "" {
		return empty, errors.New("game appId is empty")
	}
	if c.Config.AppSecret == "" {
		return empty, errors.New("game appSecret is empty")
	}

	httpClient := c.HTTPClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return empty, err
	}
	headers := BuildHeaders(c.Config.AppID, NewRequestID(), body, c.Config.AppSecret)
	headers.Apply(req.Header)

	resp, err := httpClient.Do(req)
	if err != nil {
		return empty, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return empty, err
	}
	if resp.StatusCode != http.StatusOK {
		return empty, fmt.Errorf("game api http status=%d body=%s", resp.StatusCode, string(respBody))
	}

	var apiResp APIResponse[json.RawMessage]
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return empty, err
	}
	return apiResp, nil
}

func DecodeData[T any](resp APIResponse[json.RawMessage]) (APIResponse[T], error) {
	result := APIResponse[T]{
		Code:  resp.Code,
		Error: resp.Error,
	}
	if len(resp.Data) == 0 {
		return result, nil
	}
	if err := json.Unmarshal(resp.Data, &result.Data); err != nil {
		return result, err
	}
	return result, nil
}
