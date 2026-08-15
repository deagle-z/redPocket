package game

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGGRClientProviderAndGameList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %q, want POST", r.Method)
		}
		if got := r.Header.Get("Content-Type"); got != "application/json" {
			t.Fatalf("Content-Type = %q, want application/json", got)
		}
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if payload["agent_code"] != "agent-1" || payload["agent_token"] != "token-1" {
			t.Fatalf("credentials = %#v", payload)
		}
		w.Header().Set("Content-Type", "application/json")
		switch payload["method"] {
		case GGRMethodProviderList:
			_, _ = w.Write([]byte(`{"status":1,"msg":"SUCCESS","providers":[{"code":"PRAGMATIC","name":"Pragmatic Play","status":1}]}`))
		case GGRMethodGameList:
			if payload["provider_code"] != "PRAGMATIC" {
				t.Fatalf("provider_code = %#v, want PRAGMATIC", payload["provider_code"])
			}
			_, _ = w.Write([]byte(`{"status":1,"msg":"SUCCESS","games":[{"game_code":"vs20doghouse","game_name":"The Dog House","banner":"https://img.example/game.png","status":1}]}`))
		default:
			t.Fatalf("unexpected method %#v", payload["method"])
		}
	}))
	defer server.Close()

	client := NewGGRClientWithConfig(GGRConfig{APIURL: server.URL, AgentCode: "agent-1", AgentToken: "token-1"})
	providers, err := client.ProviderList()
	if err != nil {
		t.Fatalf("ProviderList() error = %v", err)
	}
	if len(providers.Providers) != 1 || providers.Providers[0].Code != "PRAGMATIC" {
		t.Fatalf("providers = %#v", providers.Providers)
	}
	games, err := client.GameList(" pragmatic ")
	if err != nil {
		t.Fatalf("GameList() error = %v", err)
	}
	if len(games.Games) != 1 || games.Games[0].GameCode != "vs20doghouse" {
		t.Fatalf("games = %#v", games.Games)
	}
}

func TestGGRClientGameLaunchUsesDocumentedFieldsOnly(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		want := map[string]string{
			"method":        GGRMethodGameLaunch,
			"agent_code":    "agent-1",
			"agent_token":   "token-1",
			"user_code":     "12345678",
			"provider_code": "EVOLUTION",
			"game_code":     "game-100",
			"lang":          "en",
			"lobby_url":     "https://tenant.example",
		}
		for key, value := range want {
			if payload[key] != value {
				t.Fatalf("%s = %#v, want %q", key, payload[key], value)
			}
		}
		if _, exists := payload["rtp"]; exists {
			t.Fatalf("rtp must not be sent when it is not configured")
		}
		_, _ = w.Write([]byte(`{"status":1,"msg":"SUCCESS","launch_url":"https://games.example/launch"}`))
	}))
	defer server.Close()

	client := NewGGRClientWithConfig(GGRConfig{APIURL: server.URL, AgentCode: "agent-1", AgentToken: "token-1"})
	resp, err := client.GameLaunch(GGRGameLaunchInput{
		UserCode:     "12345678",
		ProviderCode: " evolution ",
		GameCode:     "game-100",
		Language:     "en",
		LobbyURL:     "https://tenant.example",
	})
	if err != nil {
		t.Fatalf("GameLaunch() error = %v", err)
	}
	if resp.LaunchURL != "https://games.example/launch" {
		t.Fatalf("LaunchURL = %q", resp.LaunchURL)
	}
}

func TestGGRClientRejectsMissingConfigWithoutLeakingToken(t *testing.T) {
	client := NewGGRClientWithConfig(GGRConfig{AgentToken: "sensitive-token"})
	_, err := client.ProviderList()
	if err == nil {
		t.Fatal("ProviderList() error = nil, want configuration error")
	}
	if strings.Contains(err.Error(), "sensitive-token") {
		t.Fatalf("error leaks agent token: %v", err)
	}
}

func TestResolveGGRProviderCategoryMappings(t *testing.T) {
	mapping := map[string]string{
		"EVOLUTION": "casino", "PP_LIVE_PRO": "casino", "SPORTSBOOK": "sports",
		"PRAGMATIC": "slots", "PGSOFT": "slots", "HABANERO": "slots", "BOOONGO": "slots",
		"PLAYSON": "slots", "CQ9": "slots", "EVOPLAY": "slots", "TOPTREND": "slots",
		"DREAMTECH": "slots", "SPRIBE": "mini", "HACKSAW": "slots", "FACHAI": "slots",
		"PLAYNGO": "slots", "AMUSNET": "slots", "EGT": "slots", "SPADEGAMING": "slots",
		"FASTSPIN": "slots", "JOKERGAMING": "slots", "RUBYPLAY": "slots", "AMATIC": "slots",
		"REELKINGDOM": "slots", "FATPANDA": "slots", "FISHHUNTER": "fishing",
	}
	if len(mapping) != 26 {
		t.Fatalf("mapping count = %d, want 26", len(mapping))
	}
	for provider, want := range mapping {
		got, err := ResolveGGRProviderCategoryWithMap(provider, mapping)
		if err != nil {
			t.Fatalf("provider=%s error=%v", provider, err)
		}
		if got != want {
			t.Fatalf("provider=%s category=%s, want %s", provider, got, want)
		}
	}
	if _, err := ResolveGGRProviderCategoryWithMap("UNKNOWN", mapping); err == nil {
		t.Fatal("unknown provider error = nil")
	}
}
