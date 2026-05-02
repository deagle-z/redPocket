package api

import (
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/utils"
	tenantRepo "BaseGoUni/tenant/repository"

	"github.com/gin-gonic/gin"
)

func GetDashboardStats(ctx *gin.Context) {
	tenantID, ok := getTenantID(ctx)
	if !ok {
		return
	}
	result := tenantRepo.GetDashboardStats(getDB(ctx), tenantID)
	utils.SuccessObjBack(ctx, result)
}

func GetDashboardOnlineUsers(ctx *gin.Context) {
	tenantID, ok := getTenantID(ctx)
	if !ok {
		return
	}
	var search pojo.TenantDashboardDetailSearch
	search.SetPageDefaults()
	if err := ctx.ShouldBindJSON(&search); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	result := tenantRepo.GetDashboardOnlineUsers(getDB(ctx), tenantID, search)
	utils.SuccessObjBack(ctx, result)
}

func GetDashboardRechargeUsers(ctx *gin.Context) {
	tenantID, ok := getTenantID(ctx)
	if !ok {
		return
	}
	var search pojo.TenantDashboardDetailSearch
	search.SetPageDefaults()
	if err := ctx.ShouldBindJSON(&search); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	result := tenantRepo.GetDashboardRechargeUsers(getDB(ctx), tenantID, search)
	utils.SuccessObjBack(ctx, result)
}

func GetDashboardRegisterUsers(ctx *gin.Context) {
	tenantID, ok := getTenantID(ctx)
	if !ok {
		return
	}
	var search pojo.TenantDashboardDetailSearch
	search.SetPageDefaults()
	if err := ctx.ShouldBindJSON(&search); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	result := tenantRepo.GetDashboardRegisterUsers(getDB(ctx), tenantID, search)
	utils.SuccessObjBack(ctx, result)
}
