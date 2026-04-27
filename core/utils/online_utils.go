package utils

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
)

const (
	OnlineUserTTL = 5 * time.Minute

	KeyOnlineAdminUsers     = "bgu_online_users:admin"
	KeyOnlineTenantUsersAll = "bgu_online_users:tenant:all"
	KeyOnlineTenantUsers    = "bgu_online_users:tenant:%d"
	KeyOnlineTgUsersAll     = "bgu_online_users:tg:all"
	KeyOnlineTgUsers        = "bgu_online_users:tg:%d"
)

type OnlineUserItem struct {
	UserID     int64 `json:"userId"`
	LastActive int64 `json:"lastActive"`
}

func OnlineTenantUsersKey(tenantID int64) string {
	return fmt.Sprintf(KeyOnlineTenantUsers, tenantID)
}

func OnlineTgUsersKey(tenantID int64) string {
	return fmt.Sprintf(KeyOnlineTgUsers, tenantID)
}

func TouchAdminOnlineUser(userID int64) {
	touchOnlineUser(KeyOnlineAdminUsers, userID)
}

func TouchTenantOnlineUser(tenantID int64, userID int64) {
	touchOnlineUser(KeyOnlineTenantUsersAll, userID)
	if tenantID > 0 {
		touchOnlineUser(OnlineTenantUsersKey(tenantID), userID)
	}
}

func TouchTgOnlineUser(tenantID int64, userID int64) {
	touchOnlineUser(KeyOnlineTgUsersAll, userID)
	touchOnlineUser(OnlineTgUsersKey(tenantID), userID)
}

func CountOnlineUsers(key string) int64 {
	if RD == nil || key == "" {
		return 0
	}
	ctx := context.Background()
	cleanExpiredOnlineUsers(ctx, key)
	count, _ := RD.ZCard(ctx, key).Result()
	return count
}

func ListOnlineUsers(key string, offset int64, limit int64) ([]OnlineUserItem, int64) {
	if RD == nil || key == "" || limit <= 0 {
		return nil, 0
	}
	ctx := context.Background()
	cleanExpiredOnlineUsers(ctx, key)
	total, _ := RD.ZCard(ctx, key).Result()
	rows, err := RD.ZRevRangeWithScores(ctx, key, offset, offset+limit-1).Result()
	if err != nil {
		return nil, total
	}
	result := make([]OnlineUserItem, 0, len(rows))
	for _, row := range rows {
		userID, err := strconv.ParseInt(fmt.Sprintf("%v", row.Member), 10, 64)
		if err != nil || userID <= 0 {
			continue
		}
		result = append(result, OnlineUserItem{
			UserID:     userID,
			LastActive: int64(row.Score),
		})
	}
	return result, total
}

func cleanExpiredOnlineUsers(ctx context.Context, key string) {
	expiredBefore := time.Now().Add(-OnlineUserTTL).Unix()
	_ = RD.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", expiredBefore)).Err()
}

func touchOnlineUser(key string, userID int64) {
	if RD == nil || key == "" || userID <= 0 {
		return
	}
	ctx := context.Background()
	member := fmt.Sprintf("%d", userID)
	now := float64(time.Now().Unix())
	pipe := RD.TxPipeline()
	pipe.ZAdd(ctx, key, &redis.Z{
		Score:  now,
		Member: member,
	})
	pipe.Expire(ctx, key, OnlineUserTTL*2)
	_, _ = pipe.Exec(ctx)
}
