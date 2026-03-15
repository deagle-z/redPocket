package api

import (
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/repository"
	"BaseGoUni/core/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"strconv"
)

// GetSysPayMethods godoc
//
//	@Summary		获取支付方式列表
//	@Tags			支付方式
//	@Accept			json
//	@Produce		json
//	@Param			data body		pojo.SysPayMethodSearch	true	"查询条件"
//	@Success		200	{object}		pojo.SysPayMethodResp
//	@Router			/api/v1/admin/sysPayMethod/list [post]
func GetSysPayMethods(ctx *gin.Context) {
	var search pojo.SysPayMethodSearch
	search.SetPageDefaults()
	if err := ctx.ShouldBindJSON(&search); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result := repository.GetSysPayMethods(db, search)
	utils.SuccessObjBack(ctx, result)
}

// SetSysPayMethod godoc
//
//	@Summary		创建或更新支付方式
//	@Tags			支付方式
//	@Accept			json
//	@Produce		json
//	@Param			data body		pojo.SysPayMethodSet	true	"支付方式信息"
//	@Success		200	{object}		pojo.SysPayMethodBack
//	@Router			/api/v1/admin/sysPayMethod [post]
func SetSysPayMethod(ctx *gin.Context) {
	var req pojo.SysPayMethodSet
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result, err := repository.SetSysPayMethod(db, req)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

// DelSysPayMethod godoc
//
//	@Summary		删除支付方式
//	@Tags			支付方式
//	@Accept			json
//	@Produce		json
//	@Param			id path		int	true	"支付方式ID"
//	@Success		200	{object}		string
//	@Router			/api/v1/admin/sysPayMethod/:id [delete]
func DelSysPayMethod(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.ErrorBack(ctx, "参数格式错误")
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result, err := repository.DelSysPayMethod(db, id)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

// GetSysPayMethodById godoc
//
//	@Summary		根据ID获取支付方式
//	@Tags			支付方式
//	@Accept			json
//	@Produce		json
//	@Param			id path		int	true	"支付方式ID"
//	@Success		200	{object}		pojo.SysPayMethodBack
//	@Router			/api/v1/admin/sysPayMethod/:id [get]
func GetSysPayMethodById(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.ErrorBack(ctx, "参数格式错误")
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result, err := repository.GetSysPayMethodById(db, id)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}
