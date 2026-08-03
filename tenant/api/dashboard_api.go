package api

import (
	"BaseGoUni/core/pojo"
	coreRepo "BaseGoUni/core/repository"
	"BaseGoUni/core/utils"
	tenantRepo "BaseGoUni/tenant/repository"
	"strconv"

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

func GetDomainVisitStats(ctx *gin.Context) {
	tenantID, ok := getTenantID(ctx)
	if !ok {
		return
	}
	result, err := coreRepo.GetTenantDomainVisitStats(getDB(ctx), tenantID)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

func GetDomainVisits(ctx *gin.Context) {
	tenantID, ok := getTenantID(ctx)
	if !ok {
		return
	}
	var pageInfo pojo.PageInfo
	pageInfo.SetPageDefaults()
	if err := ctx.ShouldBindJSON(&pageInfo); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	result, err := coreRepo.GetTenantDomainVisits(getDB(ctx), tenantID, pageInfo)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

func GetDashboardMonthlyBalances(ctx *gin.Context) {
	tenantID, ok := getTenantID(ctx)
	if !ok {
		return
	}
	year, _ := strconv.Atoi(ctx.Query("year"))
	result := tenantRepo.GetDashboardMonthlyBalances(getDB(ctx), tenantID, year)
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

func GetDashboardAgentRanks(ctx *gin.Context) {
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
	result := tenantRepo.GetDashboardAgentRanks(getDB(ctx), tenantID, search)
	utils.SuccessObjBack(ctx, result)
}

func GetDashboardRechargeOrders(ctx *gin.Context) {
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
	result := tenantRepo.GetDashboardRechargeOrders(getDB(ctx), tenantID, search)
	utils.SuccessObjBack(ctx, result)
}

func GetDashboardWithdrawOrders(ctx *gin.Context) {
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
	result := tenantRepo.GetDashboardWithdrawOrders(getDB(ctx), tenantID, search)
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
