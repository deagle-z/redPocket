package repository

import (
	"BaseGoUni/core/pojo"
	"testing"
)

func TestBuildTgAuthLoginBackMatchesLoginPayload(t *testing.T) {
	username := "alice"
	user := pojo.TgUser{
		ID:         7,
		Uid:        "U0000007",
		TgID:       7001,
		Username:   &username,
		TenantId:   3,
		Status:     1,
		RebateRate: 40,
	}

	got := buildTgAuthLoginBack("signed-token", 3600, user)

	if got.AccessToken != "signed-token" {
		t.Fatalf("AccessToken = %q, want %q", got.AccessToken, "signed-token")
	}
	if got.UserType != 5 {
		t.Fatalf("UserType = %d, want 5", got.UserType)
	}
	if got.ExpiresIn != 3600 {
		t.Fatalf("ExpiresIn = %d, want 3600", got.ExpiresIn)
	}
	if got.TgUser.ID != user.ID {
		t.Fatalf("TgUser.ID = %d, want %d", got.TgUser.ID, user.ID)
	}
	if got.TgUser.Uid != user.Uid {
		t.Fatalf("TgUser.Uid = %q, want %q", got.TgUser.Uid, user.Uid)
	}
	if got.TgUser.Username == nil || *got.TgUser.Username != username {
		t.Fatalf("TgUser.Username = %v, want %q", got.TgUser.Username, username)
	}
	if got.TgUser.TgID != user.TgID {
		t.Fatalf("TgUser.TgID = %d, want %d", got.TgUser.TgID, user.TgID)
	}
}

func TestTgLoginClaimUsernamePrefersStoredUsername(t *testing.T) {
	username := "profile_name"
	user := pojo.TgUser{Username: &username}

	got := tgLoginClaimUsername(user, "email@example.com")

	if got != username {
		t.Fatalf("tgLoginClaimUsername() = %q, want %q", got, username)
	}
}

func TestTgLoginClaimUsernameFallsBackToLoginIdentifier(t *testing.T) {
	user := pojo.TgUser{}

	got := tgLoginClaimUsername(user, " email@example.com ")

	if got != "email@example.com" {
		t.Fatalf("tgLoginClaimUsername() = %q, want %q", got, "email@example.com")
	}
}
