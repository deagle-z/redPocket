package api

import (
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/repository"
	"BaseGoUni/core/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GetAdminDashboardStats(ctx *gin.Context) {
	db := ctx.MustGet("db").(*gorm.DB)
	result := repository.GetAdminDashboardStats(db)
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
