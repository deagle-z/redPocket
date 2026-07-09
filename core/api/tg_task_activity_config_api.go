package api

import (
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/repository"
	"BaseGoUni/core/utils"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GetTgTaskActivityConfigs godoc
//
//	@Summary		获取任务活动配置列表
//	@Tags			任务活动配置
//	@Accept			json
//	@Produce		json
//	@Param			data body		pojo.TgTaskActivityConfigSearch	true	"查询条件"
//	@Success		200	{object}	pojo.TgTaskActivityConfigResp
//	@Router			/api/v1/admin/taskActivityConfig/list [post]
func GetTgTaskActivityConfigs(ctx *gin.Context) {
	var search pojo.TgTaskActivityConfigSearch
	search.SetPageDefaults()
	if err := ctx.ShouldBindJSON(&search); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	utils.SuccessObjBack(ctx, repository.GetTgTaskActivityConfigs(db, search))
}

// GetTgTaskActivityConfigById godoc
//
//	@Summary		根据ID获取任务活动配置
//	@Tags			任务活动配置
//	@Produce		json
//	@Param			id path	int	true	"配置ID"
//	@Success		200	{object}	pojo.TgTaskActivityConfigBack
//	@Router			/api/v1/admin/taskActivityConfig/:id [get]
func GetTgTaskActivityConfigById(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		utils.ErrorBack(ctx, "参数格式错误")
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result, err := repository.GetTgTaskActivityConfigById(db, id)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

// SetTgTaskActivityConfig godoc
//
//	@Summary		创建或更新任务活动配置
//	@Tags			任务活动配置
//	@Accept			json
//	@Produce		json
//	@Param			data body		pojo.TgTaskActivityConfigSet	true	"配置信息"
//	@Success		200	{object}	pojo.TgTaskActivityConfigBack
//	@Router			/api/v1/admin/taskActivityConfig [post]
func SetTgTaskActivityConfig(ctx *gin.Context) {
	var req pojo.TgTaskActivityConfigSet
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result, err := repository.SetTgTaskActivityConfig(db, req)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

// DelTgTaskActivityConfig godoc
//
//	@Summary		删除任务活动配置
//	@Tags			任务活动配置
//	@Produce		json
//	@Param			id path	int	true	"配置ID"
//	@Success		200	{object}	string
//	@Router			/api/v1/admin/taskActivityConfig/:id [delete]
func DelTgTaskActivityConfig(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		utils.ErrorBack(ctx, "参数格式错误")
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result, err := repository.DelTgTaskActivityConfig(db, id)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}
