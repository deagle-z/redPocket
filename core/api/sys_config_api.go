package api

import (
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/repository"
	"BaseGoUni/core/utils"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GetSysConfigs godoc
//
//	@Summary		获取系统配置列表
//	@Tags			系统配置
//	@Accept			json
//	@Produce		json
//	@Param			data body		pojo.SysConfigSearch	true	"查询条件"
//	@Success		200	{object}	pojo.SysConfigResp
//	@Router			/api/v1/admin/sysConfig/list [post]
func GetSysConfigs(ctx *gin.Context) {
	var search pojo.SysConfigSearch
	search.SetPageDefaults()
	if err := ctx.ShouldBindJSON(&search); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	utils.SuccessObjBack(ctx, repository.GetSysConfigs(db, search))
}

// SetSysConfig godoc
//
//	@Summary		创建或更新系统配置
//	@Tags			系统配置
//	@Accept			json
//	@Produce		json
//	@Param			data body		pojo.SysConfigSet	true	"配置信息"
//	@Success		200	{object}	pojo.SysConfigBack
//	@Router			/api/v1/admin/sysConfig [post]
func SetSysConfig(ctx *gin.Context) {
	var req pojo.SysConfigSet
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result, err := repository.SetSysConfig(db, req)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

// DelSysConfig godoc
//
//	@Summary		删除系统配置
//	@Tags			系统配置
//	@Produce		json
//	@Param			id path	int	true	"配置ID"
//	@Success		200	{object}	string
//	@Router			/api/v1/admin/sysConfig/:id [delete]
func DelSysConfig(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		utils.ErrorBack(ctx, "参数格式错误")
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	if err = repository.DelSysConfig(db, id); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, "success")
}

// GetAppSysConfig godoc
//
//	@Summary		App端根据key获取系统配置
//	@Tags			系统配置
//	@Produce		json
//	@Param			key path	string	true	"配置key"
//	@Success		200	{object}	pojo.SysConfigBack
//	@Router			/api/v1/app/config/:key [get]
func GetAppSysConfig(ctx *gin.Context) {
	key := ctx.Param("key")
	if key == "" {
		utils.ErrorBack(ctx, "key不能为空")
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result, err := repository.GetSysConfigByKey(db, key)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}
