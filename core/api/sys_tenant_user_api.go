package api

import (
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/repository"
	"BaseGoUni/core/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"strconv"
)

// GetSysTenantUsers godoc
//
//	@Summary		获取租户用户列表
//	@Tags			租户用户
//	@Accept			json
//	@Produce		json
//	@Param			data body		pojo.SysTenantUserSearch	true	"查询条件"
//	@Success		200	{object}		pojo.SysTenantUserResp
//	@Router			/api/v1/admin/tenantUser/list [post]
func GetSysTenantUsers(ctx *gin.Context) {
	var search pojo.SysTenantUserSearch
	search.SetPageDefaults()
	if err := ctx.ShouldBindJSON(&search); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result := repository.GetSysTenantUsers(db, search)
	utils.SuccessObjBack(ctx, result)
}

// SetSysTenantUser godoc
//
//	@Summary		创建或更新租户用户
//	@Tags			租户用户
//	@Accept			json
//	@Produce		json
//	@Param			data body		pojo.SysTenantUserSet	true	"租户用户信息"
//	@Success		200	{object}		pojo.SysTenantUserBack
//	@Router			/api/v1/admin/tenantUser [post]
func SetSysTenantUser(ctx *gin.Context) {
	var req pojo.SysTenantUserSet
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result, err := repository.SetSysTenantUser(db, req)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

// DelSysTenantUser godoc
//
//	@Summary		删除租户用户
//	@Tags			租户用户
//	@Accept			json
//	@Produce		json
//	@Param			id path		int	true	"租户用户ID"
//	@Success		200	{object}		string
//	@Router			/api/v1/admin/tenantUser/:id [delete]
func DelSysTenantUser(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.ErrorBack(ctx, "参数格式错误")
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result, err := repository.DelSysTenantUser(db, id)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

// GetSysTenantUserById godoc
//
//	@Summary		根据ID获取租户用户
//	@Tags			租户用户
//	@Accept			json
//	@Produce		json
//	@Param			id path		int	true	"租户用户ID"
//	@Success		200	{object}		pojo.SysTenantUserBack
//	@Router			/api/v1/admin/tenantUser/:id [get]
func GetSysTenantUserById(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.ErrorBack(ctx, "参数格式错误")
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result, err := repository.GetSysTenantUserById(db, id)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}
