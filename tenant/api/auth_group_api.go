package api

import (
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/utils"
	tenantRepo "BaseGoUni/tenant/repository"
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
)

func GetAuthGroups(ctx *gin.Context) {
	tenantID, ok := getTenantID(ctx)
	if !ok {
		return
	}
	var search pojo.AuthGroupSearch
	search.SetPageDefaults()
	if err := ctx.ShouldBindJSON(&search); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	result := tenantRepo.GetAuthGroups(getDB(ctx), tenantID, search)
	utils.SuccessObjBack(ctx, result)
}

func SetAuthGroup(ctx *gin.Context) {
	tenantID, ok := getTenantID(ctx)
	if !ok {
		return
	}
	var req pojo.AuthGroup
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	result, err := tenantRepo.SetAuthGroup(getDB(ctx), tenantID, req)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	clearAuthGroupCache(result.GroupID)
	utils.SuccessObjBack(ctx, result)
}

func DelAuthGroup(ctx *gin.Context) {
	tenantID, ok := getTenantID(ctx)
	if !ok {
		return
	}
	id, ok := parsePathID(ctx)
	if !ok {
		return
	}
	groupID, result, err := tenantRepo.DelAuthGroup(getDB(ctx), tenantID, id)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	clearAuthGroupCache(groupID)
	utils.SuccessObjBack(ctx, result)
}

func clearAuthGroupCache(groupID int64) {
	if groupID <= 0 {
		return
	}
	ctx := context.Background()
	redisKeys := []string{
		fmt.Sprintf("bgu_auth_group_lose_rate_%d", groupID),
		fmt.Sprintf("bgu_auth_group_num_config_%d", groupID),
		fmt.Sprintf("bgu_auth_group_send_commission_%d", groupID),
		fmt.Sprintf("bgu_auth_group_grabbing_commission_%d", groupID),
	}
	for _, key := range redisKeys {
		_ = utils.RD.Del(ctx, key).Err()
	}
}
