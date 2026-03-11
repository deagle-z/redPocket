package api

import (
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/repository"
	"BaseGoUni/core/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"strconv"
)

// GetPlatformProfitLedgers godoc
//
//	@Summary		获取平台盈利流水列表
//	@Tags			平台盈利流水
//	@Accept			json
//	@Produce		json
//	@Param			data body		pojo.PlatformProfitLedgerSearch	true	"查询条件"
//	@Success		200	{object}		pojo.PlatformProfitLedgerResp
//	@Router			/api/v1/admin/platformProfitLedger/list [post]
func GetPlatformProfitLedgers(ctx *gin.Context) {
	var search pojo.PlatformProfitLedgerSearch
	search.SetPageDefaults()
	if err := ctx.ShouldBindJSON(&search); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result := repository.GetPlatformProfitLedgers(db, search)
	utils.SuccessObjBack(ctx, result)
}

// SetPlatformProfitLedger godoc
//
//	@Summary		创建或更新平台盈利流水
//	@Tags			平台盈利流水
//	@Accept			json
//	@Produce		json
//	@Param			data body		pojo.PlatformProfitLedgerSet	true	"平台盈利流水信息"
//	@Success		200	{object}		pojo.PlatformProfitLedgerBack
//	@Router			/api/v1/admin/platformProfitLedger [post]
func SetPlatformProfitLedger(ctx *gin.Context) {
	var req pojo.PlatformProfitLedgerSet
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result, err := repository.SetPlatformProfitLedger(db, req)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

// DelPlatformProfitLedger godoc
//
//	@Summary		删除平台盈利流水
//	@Tags			平台盈利流水
//	@Accept			json
//	@Produce		json
//	@Param			id path		int	true	"平台盈利流水ID"
//	@Success		200	{object}		string
//	@Router			/api/v1/admin/platformProfitLedger/:id [delete]
func DelPlatformProfitLedger(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.ErrorBack(ctx, "参数格式错误")
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result, err := repository.DelPlatformProfitLedger(db, id)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

// GetPlatformProfitLedgerById godoc
//
//	@Summary		根据ID获取平台盈利流水
//	@Tags			平台盈利流水
//	@Accept			json
//	@Produce		json
//	@Param			id path		int	true	"平台盈利流水ID"
//	@Success		200	{object}		pojo.PlatformProfitLedgerBack
//	@Router			/api/v1/admin/platformProfitLedger/:id [get]
func GetPlatformProfitLedgerById(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.ErrorBack(ctx, "参数格式错误")
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result, err := repository.GetPlatformProfitLedgerById(db, id)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}
