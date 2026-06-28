package api

import (
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/repository"
	"BaseGoUni/core/utils"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GetAdminDashboardStats(ctx *gin.Context) {
	db := ctx.MustGet("db").(*gorm.DB)
	tenantID, _ := strconv.ParseInt(ctx.Query("tenantId"), 10, 64)
	result := repository.GetAdminDashboardStats(db, tenantID)
	utils.SuccessObjBack(ctx, result)
}

func GetAdminDashboardMonthlyBalances(ctx *gin.Context) {
	db := ctx.MustGet("db").(*gorm.DB)
	tenantID, _ := strconv.ParseInt(ctx.Query("tenantId"), 10, 64)
	year, _ := strconv.Atoi(ctx.Query("year"))
	result := repository.GetAdminDashboardMonthlyBalances(db, tenantID, year)
	utils.SuccessObjBack(ctx, result)
}

func GetAdminDashboardOnlineUsers(ctx *gin.Context) {
	var search pojo.TenantDashboardDetailSearch
	search.SetPageDefaults()
	if err := ctx.ShouldBindJSON(&search); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result := repository.GetAdminDashboardOnlineUsers(db, search)
	utils.SuccessObjBack(ctx, result)
}

func GetAdminDashboardRechargeUsers(ctx *gin.Context) {
	var search pojo.TenantDashboardDetailSearch
	search.SetPageDefaults()
	if err := ctx.ShouldBindJSON(&search); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result := repository.GetAdminDashboardRechargeUsers(db, search)
	utils.SuccessObjBack(ctx, result)
}

func GetAdminDashboardAgentRanks(ctx *gin.Context) {
	var search pojo.TenantDashboardDetailSearch
	search.SetPageDefaults()
	if err := ctx.ShouldBindJSON(&search); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result := repository.GetAdminDashboardAgentRanks(db, search)
	utils.SuccessObjBack(ctx, result)
}

func GetAdminDashboardRechargeOrders(ctx *gin.Context) {
	var search pojo.TenantDashboardDetailSearch
	search.SetPageDefaults()
	if err := ctx.ShouldBindJSON(&search); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result := repository.GetAdminDashboardRechargeOrders(db, search)
	utils.SuccessObjBack(ctx, result)
}

func GetAdminDashboardWithdrawOrders(ctx *gin.Context) {
	var search pojo.TenantDashboardDetailSearch
	search.SetPageDefaults()
	if err := ctx.ShouldBindJSON(&search); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result := repository.GetAdminDashboardWithdrawOrders(db, search)
	utils.SuccessObjBack(ctx, result)
}

func GetAdminDashboardRegisterUsers(ctx *gin.Context) {
	var search pojo.TenantDashboardDetailSearch
	search.SetPageDefaults()
	if err := ctx.ShouldBindJSON(&search); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result := repository.GetAdminDashboardRegisterUsers(db, search)
	utils.SuccessObjBack(ctx, result)
}
