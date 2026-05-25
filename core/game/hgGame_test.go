package game

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSignMatchesDocumentExample(t *testing.T) {
	requestID := "1760060260227_224451"
	secret := "970cb4e4-9ed3-4fc0-802c-8dbedb8b5e85"
	body := []byte(`{"language":"en"}`)
	want := "cdb2ea5d7b5186cff285b6f9607a02ce"

	if got := Sign(requestID, body, secret); got != want {
		t.Fatalf("Sign() = %q, want %q", got, want)
	}
}

func TestClientGameListSignsRequestAndParsesResponse(t *testing.T) {
	const appID = "qwe456_USD_1"
	const secret = "970cb4e4-9ed3-4fc0-802c-8dbedb8b5e85"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/game/list" {
			t.Fatalf("path = %q, want /api/v1/game/list", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Fatalf("method = %q, want POST", r.Method)
		}
		body := []byte(`{"language":"en"}`)
		if r.Header.Get(HeaderAppID) != appID {
			t.Fatalf("%s = %q", HeaderAppID, r.Header.Get(HeaderAppID))
		}
		requestID := r.Header.Get(HeaderRequestID)
		if requestID == "" {
			t.Fatalf("%s is empty", HeaderRequestID)
		}
		if got, want := r.Header.Get(HeaderSign), Sign(requestID, body, secret); got != want {
			t.Fatalf("%s = %q, want %q", HeaderSign, got, want)
		}

		w.Header().Set("Content-Type", ContentTypeJSON)
		_ = json.NewEncoder(w).Encode(APIResponse[GameListData]{
			Code: GameCodeSuccess,
			Data: GameListData{List: []GameInfo{{
				ID:              "bg_1069564",
				Name:            "Gold Rush With Johnny",
				Type:            0,
				IconURL:         "/BHdownload/BG-1069564.webp",
				Manufacturer:    "bg",
				PortraitIconURL: "/BHdownload/bg_1069564_vertical.webp",
			}}},
		})
	}))
	defer server.Close()

	client := NewClientWithConfig(Config{
		APIURL:    server.URL,
		AppID:     appID,
		AppSecret: secret,
	})
	resp, err := client.GameList("en")
	if err != nil {
		t.Fatalf("GameList() error = %v", err)
	}
	if !resp.Success() {
		t.Fatalf("response success = false, code=%d error=%q", resp.Code, resp.Error)
	}
	if len(resp.Data.List) != 1 || resp.Data.List[0].ID != "bg_1069564" {
		t.Fatalf("unexpected game list: %#v", resp.Data.List)
	}
}

func TestClientGameLaunchSignsRequestAndParsesResponse(t *testing.T) {
	const appID = "qwe456_USD_1"
	const secret = "970cb4e4-9ed3-4fc0-802c-8dbedb8b5e85"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/game/launch" {
			t.Fatalf("path = %q, want /api/v1/game/launch", r.URL.Path)
		}
		body := []byte(`{"userid":"abc","gameid":"pg_98","language":"th"}`)
		requestID := r.Header.Get(HeaderRequestID)
		if got, want := r.Header.Get(HeaderSign), Sign(requestID, body, secret); got != want {
			t.Fatalf("%s = %q, want %q", HeaderSign, got, want)
		}

		w.Header().Set("Content-Type", ContentTypeJSON)
		_ = json.NewEncoder(w).Encode(APIResponse[GameLaunchData]{
			Code: GameCodeSuccess,
			Data: GameLaunchData{URL: "https://demoapi.cc/pg_98/index.html?l=th&t=token"},
		})
	}))
	defer server.Close()

	client := NewClientWithConfig(Config{
		APIURL:    server.URL,
		AppID:     appID,
		AppSecret: secret,
	})
	resp, err := client.GameLaunch("abc", "pg_98", "th")
	if err != nil {
		t.Fatalf("GameLaunch() error = %v", err)
	}
	if got := resp.Data.URL; got != "https://demoapi.cc/pg_98/index.html?l=th&t=token" {
		t.Fatalf("launch url = %q", got)
	}
}
