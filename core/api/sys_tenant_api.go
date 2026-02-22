package api

import (
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/repository"
	"BaseGoUni/core/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"strconv"
)

// GetSysTenants godoc
//
//	@Summary		获取租户列表
//	@Tags			租户
//	@Accept			json
//	@Produce		json
//	@Param			data body		pojo.SysTenantSearch	true	"查询条件"
//	@Success		200	{object}		pojo.SysTenantResp
//	@Router			/api/v1/admin/tenant/list [post]
func GetSysTenants(ctx *gin.Context) {
	var search pojo.SysTenantSearch
	search.SetPageDefaults()
	if err := ctx.ShouldBindJSON(&search); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result := repository.GetSysTenants(db, search)
	utils.SuccessObjBack(ctx, result)
}

// SetSysTenant godoc
//
//	@Summary		创建或更新租户
//	@Tags			租户
//	@Accept			json
//	@Produce		json
//	@Param			data body		pojo.SysTenantSet	true	"租户信息"
//	@Success		200	{object}		pojo.SysTenantBack
//	@Router			/api/v1/admin/tenant [post]
func SetSysTenant(ctx *gin.Context) {
	var req pojo.SysTenantSet
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result, err := repository.SetSysTenant(db, req)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

// DelSysTenant godoc
//
//	@Summary		删除租户
//	@Tags			租户
//	@Accept			json
//	@Produce		json
//	@Param			id path		int	true	"租户ID"
//	@Success		200	{object}		string
//	@Router			/api/v1/admin/tenant/:id [delete]
func DelSysTenant(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.ErrorBack(ctx, "参数格式错误")
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result, err := repository.DelSysTenant(db, id)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

// GetSysTenantById godoc
//
//	@Summary		根据ID获取租户
//	@Tags			租户
//	@Accept			json
//	@Produce		json
//	@Param			id path		int	true	"租户ID"
//	@Success		200	{object}		pojo.SysTenantBack
//	@Router			/api/v1/admin/tenant/:id [get]
func GetSysTenantById(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.ErrorBack(ctx, "参数格式错误")
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result, err := repository.GetSysTenantById(db, id)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

// ResetSysTenantPassword godoc
//
//	@Summary		重置租户密码
//	@Tags			租户
//	@Accept			json
//	@Produce		json
//	@Param			data body		pojo.SysTenantResetPassword	true	"租户重置密码参数"
//	@Success		200	{object}		string
//	@Router			/api/v1/admin/tenant/resetPassword [post]
func ResetSysTenantPassword(ctx *gin.Context) {
	var req pojo.SysTenantResetPassword
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result, err := repository.ResetSysTenantPassword(db, req)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}
