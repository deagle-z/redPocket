package api

import (
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/repository"
	"BaseGoUni/core/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"strconv"
)

// GetTgUserRebateRecords godoc
//
//	@Summary		获取反水/返佣记录列表
//	@Tags			Telegram反水记录
//	@Accept			json
//	@Produce		json
//	@Param			data body		pojo.TgUserRebateRecordSearch	true	"查询条件"
//	@Success		200	{object}		pojo.TgUserRebateRecordResp
//	@Router			/api/v1/admin/tgUserRebate/list [post]
func GetTgUserRebateRecords(ctx *gin.Context) {
	var search pojo.TgUserRebateRecordSearch
	search.SetPageDefaults()
	if err := ctx.ShouldBindJSON(&search); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result := repository.GetTgUserRebateRecords(db, search)
	utils.SuccessObjBack(ctx, result)
}

// SetTgUserRebateRecord godoc
//
//	@Summary		创建或更新反水/返佣记录
//	@Tags			Telegram反水记录
//	@Accept			json
//	@Produce		json
//	@Param			data body		pojo.TgUserRebateRecordSet	true	"反水记录信息"
//	@Success		200	{object}		pojo.TgUserRebateRecordBack
//	@Router			/api/v1/admin/tgUserRebate [post]
func SetTgUserRebateRecord(ctx *gin.Context) {
	var req pojo.TgUserRebateRecordSet
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result, err := repository.SetTgUserRebateRecord(db, req)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

// DelTgUserRebateRecord godoc
//
//	@Summary		删除反水/返佣记录
//	@Tags			Telegram反水记录
//	@Accept			json
//	@Produce		json
//	@Param			id path		int	true	"记录ID"
//	@Success		200	{object}		string
//	@Router			/api/v1/admin/tgUserRebate/:id [delete]
func DelTgUserRebateRecord(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.ErrorBack(ctx, "参数格式错误")
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result, err := repository.DelTgUserRebateRecord(db, id)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

// GetTgUserRebateRecordById godoc
//
//	@Summary		根据ID获取反水/返佣记录
//	@Tags			Telegram反水记录
//	@Accept			json
//	@Produce		json
//	@Param			id path		int	true	"记录ID"
//	@Success		200	{object}		pojo.TgUserRebateRecordBack
//	@Router			/api/v1/admin/tgUserRebate/:id [get]
func GetTgUserRebateRecordById(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.ErrorBack(ctx, "参数格式错误")
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result, err := repository.GetTgUserRebateRecordById(db, id)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}
