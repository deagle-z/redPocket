package api

import (
	"encoding/json"
	"strings"
	"testing"

	"BaseGoUni/core/game"
)

func TestValidateGGRTransactionAcceptsAllGameTypes(t *testing.T) {
	cfg := game.GGRConfig{CategoryMap: map[string]string{
		"PRAGMATIC":  "slots",
		"EVOLUTION":  "casino",
		"SPORTSBOOK": "sports",
		"SPRIBE":     "mini",
	}}
	tests := []struct {
		name     string
		gameType string
		provider string
		txnType  string
		bet      string
		win      string
		setGame  func(*ggrWalletRequest, *ggrTransactionGame)
	}{
		{name: "slot debit", gameType: "slot", provider: "PRAGMATIC", txnType: "debit", bet: "10", win: "0", setGame: func(req *ggrWalletRequest, item *ggrTransactionGame) { req.Slot = item }},
		{name: "live credit", gameType: "live", provider: "EVOLUTION", txnType: "credit", bet: "0", win: "20", setGame: func(req *ggrWalletRequest, item *ggrTransactionGame) { req.Live = item }},
		{name: "sports combined", gameType: "SB", provider: "SPORTSBOOK", txnType: "debit_credit", bet: "5", win: "0", setGame: func(req *ggrWalletRequest, item *ggrTransactionGame) { req.Sportsbook = item }},
		{name: "mini combined", gameType: "MN", provider: "SPRIBE", txnType: "debit_credit", bet: "1", win: "2", setGame: func(req *ggrWalletRequest, item *ggrTransactionGame) { req.Mini = item }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			bet := json.Number(test.bet)
			win := json.Number(test.win)
			req := ggrWalletRequest{
				Method:    ggrMethodTransaction,
				AgentCode: "agent-1",
				UserCode:  "12345678",
				GameType:  test.gameType,
			}
			item := &ggrTransactionGame{
				ProviderCode: test.provider,
				GameCode:     "game-1",
				BetType:      "BASE",
				BetMoney:     &bet,
				WinMoney:     &win,
				RoundID:      json.RawMessage("123456789"),
				TxnID:        "txn-1-" + test.name,
				TxnType:      test.txnType,
			}
			test.setGame(&req, item)
			got, err := validateGGRTransaction(req, cfg)
			if err != nil {
				t.Fatalf("validateGGRTransaction() error = %v", err)
			}
			if got.ProviderCode != test.provider || got.TxnType != test.txnType {
				t.Fatalf("validated transaction = %#v", got)
			}
		})
	}
}

func TestValidateGGRTransactionStrictMoneyRules(t *testing.T) {
	cfg := game.GGRConfig{CategoryMap: map[string]string{"PRAGMATIC": "slots"}}
	tests := []struct {
		name    string
		txnType string
		bet     string
		win     string
	}{
		{name: "debit with win", txnType: "debit", bet: "10", win: "1"},
		{name: "credit with bet", txnType: "credit", bet: "1", win: "10"},
		{name: "empty combined", txnType: "debit_credit", bet: "0", win: "0"},
		{name: "negative amount", txnType: "debit", bet: "-1", win: "0"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			bet := json.Number(test.bet)
			win := json.Number(test.win)
			req := ggrWalletRequest{
				Method:    ggrMethodTransaction,
				AgentCode: "agent-1",
				UserCode:  "12345678",
				GameType:  "slot",
				Slot: &ggrTransactionGame{
					ProviderCode: "PRAGMATIC",
					GameCode:     "game-1",
					BetType:      "BASE",
					BetMoney:     &bet,
					WinMoney:     &win,
					TxnID:        "txn-invalid",
					TxnType:      test.txnType,
				},
			}
			if _, err := validateGGRTransaction(req, cfg); err == nil {
				t.Fatal("validateGGRTransaction() error = nil")
			}
		})
	}
}

func TestValidateGGRTransactionRejectsObjectMismatch(t *testing.T) {
	bet := json.Number("1")
	win := json.Number("0")
	req := ggrWalletRequest{
		Method:    ggrMethodTransaction,
		AgentCode: "agent-1",
		UserCode:  "12345678",
		GameType:  "live",
		Slot: &ggrTransactionGame{
			ProviderCode: "PRAGMATIC",
			GameCode:     "game-1",
			BetType:      "BASE",
			BetMoney:     &bet,
			WinMoney:     &win,
			TxnID:        "txn-1",
			TxnType:      "debit",
		},
	}
	cfg := game.GGRConfig{CategoryMap: map[string]string{"PRAGMATIC": "slots"}}
	if _, err := validateGGRTransaction(req, cfg); err == nil {
		t.Fatal("game object mismatch error = nil")
	}
}

func TestGGRMoneyFormulaAndFingerprint(t *testing.T) {
	if got := calculateGGREndBalance(100, 30.129, 5.999); got != 75.87 {
		t.Fatalf("calculateGGREndBalance() = %.2f, want 75.87", got)
	}
	if got := calculateGGREndBalance(1, 2, 0); got >= 0 {
		t.Fatalf("insufficient end balance = %.2f, want negative", got)
	}

	base := validatedGGRTransaction{
		AgentCode: "agent-1", UserCode: "12345678", GameType: "slot",
		ProviderCode: "PRAGMATIC", GameCode: "game-1", BetType: "BASE", BetMoney: 10,
		WinMoney: 2, RoundID: "123", TxnID: "txn-1", TxnType: "debit_credit",
	}
	first, err := fingerprintGGRTransaction(base)
	if err != nil {
		t.Fatalf("fingerprintGGRTransaction() error = %v", err)
	}
	base.WinMoney = 3
	second, err := fingerprintGGRTransaction(base)
	if err != nil {
		t.Fatalf("fingerprintGGRTransaction() error = %v", err)
	}
	if first == second {
		t.Fatal("fingerprint did not change with win_money")
	}
	if ggrTransactionStorageKey("txn-1") != ggrTransactionStorageKey("txn-1") {
		t.Fatal("transaction storage key is not deterministic")
	}
}

func TestGGRWalletResponseKeepsZeroBalance(t *testing.T) {
	balance := 0.0
	body, err := json.Marshal(ggrWalletResponse{Status: 1, UserBalance: &balance})
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	if !strings.Contains(string(body), `"user_balance":0`) {
		t.Fatalf("response = %s, want explicit zero user_balance", body)
	}
}

func TestValidateGGRBalanceUserCodeOnlyRequiresUserCode(t *testing.T) {
	got, err := validateGGRBalanceUserCode(ggrWalletRequest{UserCode: " 12345678 "})
	if err != nil {
		t.Fatalf("validateGGRBalanceUserCode() error = %v", err)
	}
	if got != "12345678" {
		t.Fatalf("validateGGRBalanceUserCode() = %q, want %q", got, "12345678")
	}

	if _, err := validateGGRBalanceUserCode(ggrWalletRequest{}); err == nil {
		t.Fatal("validateGGRBalanceUserCode() accepted an empty user_code")
	}
}
