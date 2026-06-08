package api

import (
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/repository"
	"BaseGoUni/core/utils"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GetExchangeCodeListAdmin 获取兑换码列表。
func GetExchangeCodeListAdmin(ctx *gin.Context) {
	var search pojo.ExchangeCodeSearch
	search.SetPageDefaults()
	if err := ctx.ShouldBindJSON(&search); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}

	db := ctx.MustGet("db").(*gorm.DB)
	result, err := repository.GetExchangeCodes(db, search)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

// GetExchangeCodeByIDAdmin 获取兑换码详情。
func GetExchangeCodeByIDAdmin(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		utils.ErrorBack(ctx, "invalid_params")
		return
	}

	db := ctx.MustGet("db").(*gorm.DB)
	result, err := repository.GetExchangeCodeByID(db, id)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

// SetExchangeCodeAdmin 创建或更新兑换码。
func SetExchangeCodeAdmin(ctx *gin.Context) {
	var req pojo.ExchangeCodeSet
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}

	currentUser, err := utils.GetCurrentUser(ctx)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result, err := repository.SetExchangeCode(db, currentUser, req)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

// DelExchangeCodeAdmin 删除兑换码。
func DelExchangeCodeAdmin(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		utils.ErrorBack(ctx, "invalid_params")
		return
	}

	db := ctx.MustGet("db").(*gorm.DB)
	if err := repository.DelExchangeCode(db, id); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, "success")
}

// RedeemExchangeCodeApp H5 用户兑换兑换码。
func RedeemExchangeCodeApp(ctx *gin.Context) {
	var req pojo.ExchangeCodeRedeemReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}

	userID := ctx.MustGet("userId").(int64)
	tenantID := ctx.MustGet("tenantId").(int64)
	db := ctx.MustGet("db").(*gorm.DB)
	result, err := repository.RedeemExchangeCode(db, tenantID, userID, req.Code)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}
