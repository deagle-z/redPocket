package utils

import (
	"BaseGoUni/core/pojo"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestVerifyTelegramLoginWidget_OK(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	req := pojo.TgAuthLoginReq{
		ID:        123456789,
		FirstName: "Alice",
		Username:  "alice_bot",
		AuthDate:  now.Unix(),
	}
	req.Hash = buildTestTgHash("test_bot_token", req)
	if err := VerifyTelegramLoginWidget("test_bot_token", req, now); err != nil {
		t.Fatalf("expect nil error, got %v", err)
	}
}

func TestVerifyTelegramLoginWidget_InvalidHash(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	req := pojo.TgAuthLoginReq{
		ID:        123456789,
		FirstName: "Alice",
		Username:  "alice_bot",
		AuthDate:  now.Unix(),
		Hash:      "deadbeef",
	}
	if err := VerifyTelegramLoginWidget("test_bot_token", req, now); err == nil {
		t.Fatal("expect hash error, got nil")
	}
}

func TestVerifyTelegramLoginWidget_Expired(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	req := pojo.TgAuthLoginReq{
		ID:        123456789,
		FirstName: "Alice",
		Username:  "alice_bot",
		AuthDate:  now.Add(-(TgAuthMaxAgeSeconds + 1) * time.Second).Unix(),
	}
	req.Hash = buildTestTgHash("test_bot_token", req)
	if err := VerifyTelegramLoginWidget("test_bot_token", req, now); err == nil {
		t.Fatal("expect expired error, got nil")
	}
}

func TestVerifyTelegramLoginWidget_EmptyToken(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	req := pojo.TgAuthLoginReq{
		ID:        123456789,
		FirstName: "Alice",
		Username:  "alice_bot",
		AuthDate:  now.Unix(),
		Hash:      "abc",
	}
	if err := VerifyTelegramLoginWidget("", req, now); err == nil {
		t.Fatal("expect empty token error, got nil")
	}
}

func buildTestTgHash(botToken string, req pojo.TgAuthLoginReq) string {
	dataMap := map[string]string{
		"auth_date":  strconv.FormatInt(req.AuthDate, 10),
		"first_name": strings.TrimSpace(req.FirstName),
		"id":         strconv.FormatInt(req.ID, 10),
		"last_name":  strings.TrimSpace(req.LastName),
		"photo_url":  strings.TrimSpace(req.PhotoURL),
		"username":   strings.TrimSpace(req.Username),
	}
	keys := make([]string, 0, len(dataMap))
	for k, v := range dataMap {
		if strings.TrimSpace(v) == "" {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%s", k, dataMap[k]))
	}
	dataCheckString := strings.Join(parts, "\n")

	secret := sha256.Sum256([]byte(botToken))
	mac := hmac.New(sha256.New, secret[:])
	_, _ = mac.Write([]byte(dataCheckString))
	return hex.EncodeToString(mac.Sum(nil))
}
