package api

import (
	"BaseGoUni/core/game"
	"testing"
)

func TestCanLaunchAppGameAfterMinimumRecharge(t *testing.T) {
	if !canLaunchAppGame(50, false) {
		t.Fatalf("canLaunchAppGame() = false, want true")
	}
}

func TestCanLaunchAppGameRejectsRechargeBelowMinimumWithoutRebateTransfer(t *testing.T) {
	if canLaunchAppGame(49.99, false) {
		t.Fatalf("canLaunchAppGame() = true, want false")
	}
}

func TestCanLaunchAppGameAfterRebateTransfer(t *testing.T) {
	if !canLaunchAppGame(0, true) {
		t.Fatalf("canLaunchAppGame() = false, want true")
	}
}

func TestCanLaunchAppGameAfterRebateTransferBelowMinimumRecharge(t *testing.T) {
	if !canLaunchAppGame(10, true) {
		t.Fatalf("canLaunchAppGame() = false, want true")
	}
}

func TestCanLaunchAppGameRejectsWithoutRechargeOrRebateTransfer(t *testing.T) {
	if canLaunchAppGame(0, false) {
		t.Fatalf("canLaunchAppGame() = true, want false")
	}
}

func TestShouldLaunchWithHGClientAllowsOnlyHGOrLegacyEmptyPlatform(t *testing.T) {
	if !shouldLaunchWithHGClient("hg") {
		t.Fatalf("shouldLaunchWithHGClient(hg) = false, want true")
	}
	if !shouldLaunchWithHGClient("") {
		t.Fatalf("shouldLaunchWithHGClient(empty) = false, want true")
	}
	if shouldLaunchWithHGClient("gsc") {
		t.Fatalf("shouldLaunchWithHGClient(gsc) = true, want false")
	}
}

func TestShouldLaunchWithGSCClientAllowsOnlyGSC(t *testing.T) {
	if !shouldLaunchWithGSCClient("gsc") {
		t.Fatalf("shouldLaunchWithGSCClient(gsc) = false, want true")
	}
	if shouldLaunchWithGSCClient("hg") {
		t.Fatalf("shouldLaunchWithGSCClient(hg) = true, want false")
	}
	if shouldLaunchWithGSCClient("") {
		t.Fatalf("shouldLaunchWithGSCClient(empty) = true, want false")
	}
}

func TestBuildAppGameSetFromGSCProviderGameUsesConfiguredMapping(t *testing.T) {
	item := game.GSCProviderGame{
		GameCode:        "aviator",
		GameName:        "Fallback Aviator",
		GameType:        "POKER",
		ImageURL:        "https://img.example/default.png",
		SupportCurrency: "PEN",
		Status:          "ACTIVATED",
		LangName: map[string]string{
			"0": "Aviator",
		},
		LangIcon: map[string]string{
			"0": "https://img.example/lang.png",
		},
	}

	got := buildAppGameSetFromGSCProviderGame(
		"gsc",
		7,
		item,
		map[string]string{"POKER": "mini"},
	)

	if got.ThirdGameID == nil || *got.ThirdGameID != "aviator" {
		t.Fatalf("ThirdGameID = %#v, want aviator", got.ThirdGameID)
	}
	if got.GameName == nil || *got.GameName != "Aviator" {
		t.Fatalf("GameName = %#v, want Aviator", got.GameName)
	}
	if got.CategoryCode == nil || *got.CategoryCode != "mini" {
		t.Fatalf("CategoryCode = %#v, want mini", got.CategoryCode)
	}
	if got.Type == nil || *got.Type != 1 {
		t.Fatalf("Type = %#v, want 1", got.Type)
	}
	if got.GameIcon == nil || *got.GameIcon != "https://img.example/default.png" {
		t.Fatalf("GameIcon = %#v, want image_url icon", got.GameIcon)
	}
	if got.DisabledFlag == nil || *got.DisabledFlag != 0 {
		t.Fatalf("DisabledFlag = %#v, want 0", got.DisabledFlag)
	}
}

func TestBuildAppGameSetFromGSCProviderGameDisablesInactiveStatus(t *testing.T) {
	item := game.GSCProviderGame{
		GameCode:        "closed_game",
		GameName:        "Closed Game",
		GameType:        "UNKNOWN",
		ImageURL:        "https://img.example/closed.png",
		SupportCurrency: "PEN",
		Status:          "MAINTAINING",
	}

	got := buildAppGameSetFromGSCProviderGame("gsc", 1, item, nil)

	if got.CategoryCode == nil || *got.CategoryCode != "mini" {
		t.Fatalf("CategoryCode = %#v, want mini fallback", got.CategoryCode)
	}
	if got.DisabledFlag == nil || *got.DisabledFlag != 1 {
		t.Fatalf("DisabledFlag = %#v, want 1", got.DisabledFlag)
	}
}
