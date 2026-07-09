package game

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"
)

func TestGSCSignMatchesDocumentFormula(t *testing.T) {
	got := GSCSign(1694617425, "XXXX", "CMUT_V2")
	const want = "cf9661ac06ecef181d35403a85e80f73"
	if got != want {
		t.Fatalf("GSCSign() = %q, want %q", got, want)
	}
}

func TestGSCLaunchSignMatchesDocumentFormula(t *testing.T) {
	got := GSCLaunchSign(1694617425, "XXXX", "CMUT_V2")
	const want = "3944f891d3bd2d42f7b1c740c884fc38"
	if got != want {
		t.Fatalf("GSCLaunchSign() = %q, want %q", got, want)
	}
}

func TestGSCClientGameListSignsQueryAndPaginates(t *testing.T) {
	const (
		productCode = 1138
		operator    = "CMUT_V2"
		secret      = "XXXX"
		requestAt   = int64(1694617425)
	)

	seenOffsets := make([]int, 0, 2)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/operators/provider-games" {
			t.Fatalf("path = %q, want /api/operators/provider-games", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Fatalf("method = %q, want GET", r.Method)
		}

		query := r.URL.Query()
		if got := query.Get("product_code"); got != strconv.Itoa(productCode) {
			t.Fatalf("product_code = %q, want %d", got, productCode)
		}
		if got := query.Get("operator_code"); got != operator {
			t.Fatalf("operator_code = %q, want %q", got, operator)
		}
		if got := query.Get("request_time"); got != strconv.FormatInt(requestAt, 10) {
			t.Fatalf("request_time = %q, want %d", got, requestAt)
		}
		if got, want := query.Get("sign"), GSCSign(requestAt, secret, operator); got != want {
			t.Fatalf("sign = %q, want %q", got, want)
		}
		if got := query.Get("game_type"); got != "POKER" {
			t.Fatalf("game_type = %q, want POKER", got)
		}
		if got := query.Get("size"); got != "2" {
			t.Fatalf("size = %q, want 2", got)
		}

		offset, err := strconv.Atoi(query.Get("offset"))
		if err != nil {
			t.Fatalf("offset parse error: %v", err)
		}
		seenOffsets = append(seenOffsets, offset)

		w.Header().Set("Content-Type", "application/json")
		switch offset {
		case 0:
			_ = json.NewEncoder(w).Encode(GSCGameListResponse{
				Code: 0,
				ProviderGames: []GSCProviderGame{
					{GameCode: "aviator", GameName: "Aviator", GameType: "POKER", ProductCode: productCode, SupportCurrency: "MXN", Status: "ACTIVATED"},
					{GameCode: "dice", GameName: "Dice", GameType: "POKER", ProductCode: productCode, SupportCurrency: "MXN", Status: "ACTIVATED"},
				},
				Pagination: GSCPagination{Size: 2, Offset: 0, Total: 3},
			})
		case 2:
			_ = json.NewEncoder(w).Encode(GSCGameListResponse{
				Code: 0,
				ProviderGames: []GSCProviderGame{
					{GameCode: "crash", GameName: "Crash", GameType: "POKER", ProductCode: productCode, SupportCurrency: "MXN", Status: "ACTIVATED"},
				},
				Pagination: GSCPagination{Size: 2, Offset: 2, Total: 3},
			})
		default:
			t.Fatalf("unexpected offset %d", offset)
		}
	}))
	defer server.Close()

	client := NewGSCClientWithConfig(GSCConfig{
		OperatorURL:  server.URL,
		ProductCode:  productCode,
		OperatorCode: operator,
		SecretKey:    secret,
		GameType:     "POKER",
		PageSize:     2,
	})
	client.Now = func() time.Time { return time.Unix(requestAt, 0) }

	games, err := client.GameList()
	if err != nil {
		t.Fatalf("GameList() error = %v", err)
	}
	if len(games) != 3 {
		t.Fatalf("len(games) = %d, want 3", len(games))
	}
	if len(seenOffsets) != 2 || seenOffsets[0] != 0 || seenOffsets[1] != 2 {
		t.Fatalf("seenOffsets = %#v, want [0 2]", seenOffsets)
	}
}

func TestGSCClientLaunchGameSignsJSONBodyAndParsesResponse(t *testing.T) {
	const (
		productCode = 1001
		operator    = "CMUT_V2"
		secret      = "XXXX"
		requestAt   = int64(1694617425)
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/operators/launch-game" {
			t.Fatalf("path = %q, want /api/operators/launch-game", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Fatalf("method = %q, want POST", r.Method)
		}
		if got := r.Header.Get("Content-Type"); got != "application/json" {
			t.Fatalf("Content-Type = %q, want application/json", got)
		}

		var payload GSCLaunchGameRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if payload.OperatorCode != operator {
			t.Fatalf("operator_code = %q, want %q", payload.OperatorCode, operator)
		}
		if payload.MemberAccount != "T9XPKVGB" {
			t.Fatalf("member_account = %q, want T9XPKVGB", payload.MemberAccount)
		}
		if payload.Password != "Qwer1234" {
			t.Fatalf("password = %q, want Qwer1234", payload.Password)
		}
		if payload.Nickname != "Player One" {
			t.Fatalf("nickname = %q, want Player One", payload.Nickname)
		}
		if payload.Currency != "MXN" {
			t.Fatalf("currency = %q, want MXN", payload.Currency)
		}
		if payload.GameCode != nil {
			t.Fatalf("game_code = %#v, want nil", payload.GameCode)
		}
		if payload.ProductCode != productCode {
			t.Fatalf("product_code = %d, want %d", payload.ProductCode, productCode)
		}
		if payload.GameType != "SPORT_BOOK" {
			t.Fatalf("game_type = %q, want SPORT_BOOK", payload.GameType)
		}
		if payload.LanguageCode != "0" {
			t.Fatalf("language_code = %q, want 0", payload.LanguageCode)
		}
		if payload.IP != "127.0.0.1" {
			t.Fatalf("ip = %q, want 127.0.0.1", payload.IP)
		}
		if payload.Platform != "WEB" {
			t.Fatalf("platform = %q, want WEB", payload.Platform)
		}
		if payload.OperatorLobbyURL != "https://app.example.test" {
			t.Fatalf("operator_lobby_url = %q, want lobby url", payload.OperatorLobbyURL)
		}
		if payload.RequestTime != requestAt {
			t.Fatalf("request_time = %d, want %d", payload.RequestTime, requestAt)
		}
		if got, want := payload.Sign, GSCLaunchSign(requestAt, secret, operator); got != want {
			t.Fatalf("sign = %q, want %q", got, want)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(GSCLaunchGameResponse{
			Code:    200,
			Message: "",
			URL:     "https://gsc.example/launch/token",
		})
	}))
	defer server.Close()

	client := NewGSCClientWithConfig(GSCConfig{
		OperatorURL:      server.URL,
		ProductCode:      productCode,
		OperatorCode:     operator,
		SecretKey:        secret,
		Currency:         "MXN",
		Password:         "Qwer1234",
		LaunchPlatform:   "WEB",
		LaunchLanguage:   "0",
		OperatorLobbyURL: "https://app.example.test",
	})
	client.Now = func() time.Time { return time.Unix(requestAt, 0) }

	resp, err := client.LaunchGame(GSCLaunchGameInput{
		MemberAccount: "T9XPKVGB",
		Nickname:      "Player One",
		IP:            "127.0.0.1",
	})
	if err != nil {
		t.Fatalf("LaunchGame() error = %v", err)
	}
	if resp.URL != "https://gsc.example/launch/token" {
		t.Fatalf("launch url = %q, want https://gsc.example/launch/token", resp.URL)
	}
}
