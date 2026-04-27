package api

import (
	"BaseGoUni/core/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

type OnlineUserStatsBack struct {
	AdminUsers       int64 `json:"adminUsers"`
	TenantUsers      int64 `json:"tenantUsers"`
	TgUsers          int64 `json:"tgUsers"`
	TenantID         int64 `json:"tenantId,omitempty"`
	TotalTenantUsers int64 `json:"totalTenantUsers,omitempty"`
	TotalTgUsers     int64 `json:"totalTgUsers,omitempty"`
}

func GetAdminOnlineUserStats(ctx *gin.Context) {
	tenantID, _ := strconv.ParseInt(ctx.Query("tenantId"), 10, 64)
	result := OnlineUserStatsBack{
		AdminUsers:       utils.CountOnlineUsers(utils.KeyOnlineAdminUsers),
		TotalTenantUsers: utils.CountOnlineUsers(utils.KeyOnlineTenantUsersAll),
		TotalTgUsers:     utils.CountOnlineUsers(utils.KeyOnlineTgUsersAll),
	}
	if tenantID > 0 {
		result.TenantID = tenantID
		result.TenantUsers = utils.CountOnlineUsers(utils.OnlineTenantUsersKey(tenantID))
		result.TgUsers = utils.CountOnlineUsers(utils.OnlineTgUsersKey(tenantID))
	}
	utils.SuccessObjBack(ctx, result)
}

func GetTenantOnlineUserStats(ctx *gin.Context) {
	tenantID, ok := getTenantID(ctx)
	if !ok {
		return
	}
	result := OnlineUserStatsBack{
		TenantID:    tenantID,
		TenantUsers: utils.CountOnlineUsers(utils.OnlineTenantUsersKey(tenantID)),
		TgUsers:     utils.CountOnlineUsers(utils.OnlineTgUsersKey(tenantID)),
	}
	utils.SuccessObjBack(ctx, result)
}
