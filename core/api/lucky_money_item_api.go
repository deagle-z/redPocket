package api

import (
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/repository"
	"BaseGoUni/core/utils"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GetLuckyMoneyItems godoc
//
//	@Summary		获取红包明细列表
//	@Tags			红包明细
//	@Accept			json
//	@Produce		json
//	@Param			data body		pojo.LuckyMoneyItemSearch	true	"查询条件"
//	@Success		200	{object}		pojo.LuckyMoneyItemResp
//	@Router			/api/v1/admin/luckyItem/list [post]
func GetLuckyMoneyItems(ctx *gin.Context) {
	var search pojo.LuckyMoneyItemSearch
	search.SetPageDefaults()
	if err := ctx.ShouldBindJSON(&search); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result := repository.GetLuckyMoneyItems(db, search)
	utils.SuccessObjBack(ctx, result)
}

// SetLuckyMoneyItem godoc
//
//	@Summary		创建或更新红包明细
//	@Tags			红包明细
//	@Accept			json
//	@Produce		json
//	@Param			data body		pojo.LuckyMoneyItemSet	true	"红包明细信息"
//	@Success		200	{object}		pojo.LuckyMoneyItemBack
//	@Router			/api/v1/admin/luckyItem [post]
func SetLuckyMoneyItem(ctx *gin.Context) {
	var req pojo.LuckyMoneyItemSet
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result, err := repository.SetLuckyMoneyItem(db, req)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

// DelLuckyMoneyItem godoc
//
//	@Summary		删除红包明细
//	@Tags			红包明细
//	@Accept			json
//	@Produce		json
//	@Param			id path		int	true	"明细ID"
//	@Success		200	{object}		string
//	@Router			/api/v1/admin/luckyItem/:id [delete]
func DelLuckyMoneyItem(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		utils.ErrorBack(ctx, "参数格式错误")
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result, err := repository.DelLuckyMoneyItem(db, id)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

// GetLuckyMoneyItemById godoc
//
//	@Summary		根据ID获取红包明细
//	@Tags			红包明细
//	@Accept			json
//	@Produce		json
//	@Param			id path		int	true	"明细ID"
//	@Success		200	{object}		pojo.LuckyMoneyItemBack
//	@Router			/api/v1/admin/luckyItem/:id [get]
func GetLuckyMoneyItemById(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		utils.ErrorBack(ctx, "参数格式错误")
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result, err := repository.GetLuckyMoneyItemById(db, id)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}
