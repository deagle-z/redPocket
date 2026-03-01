package utils

import (
	"BaseGoUni/core/pojo"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

const TgAuthMaxAgeSeconds int64 = 300

func VerifyTelegramLoginWidget(botToken string, req pojo.TgAuthLoginReq, now time.Time) error {
	if strings.TrimSpace(botToken) == "" {
		return errors.New("telegram bot token is empty")
	}
	if req.ID <= 0 || req.AuthDate <= 0 || strings.TrimSpace(req.Hash) == "" {
		return errors.New("invalid tg auth params")
	}

	nowUnix := now.Unix()
	if nowUnix-req.AuthDate > TgAuthMaxAgeSeconds || req.AuthDate-nowUnix > TgAuthMaxAgeSeconds {
		return errors.New("tg auth expired")
	}

	dataMap := map[string]string{
		"auth_date":  strconv.FormatInt(req.AuthDate, 10),
		"first_name": strings.TrimSpace(req.FirstName),
		"id":         strconv.FormatInt(req.ID, 10),
		"last_name":  strings.TrimSpace(req.LastName),
		"photo_url":  strings.TrimSpace(req.PhotoURL),
		"username":   strings.TrimSpace(req.Username),
	}
	dataCheckString := buildTgDataCheckString(dataMap)
	if dataCheckString == "" {
		return errors.New("invalid tg auth payload")
	}

	secret := sha256.Sum256([]byte(botToken))
	mac := hmac.New(sha256.New, secret[:])
	_, _ = mac.Write([]byte(dataCheckString))
	expectedHash := hex.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(strings.ToLower(expectedHash)), []byte(strings.ToLower(strings.TrimSpace(req.Hash)))) {
		return errors.New("invalid tg auth hash")
	}
	return nil
}

func buildTgDataCheckString(data map[string]string) string {
	keys := make([]string, 0, len(data))
	for k, v := range data {
		if strings.TrimSpace(v) == "" {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%s", k, data[k]))
	}
	return strings.Join(parts, "\n")
}
