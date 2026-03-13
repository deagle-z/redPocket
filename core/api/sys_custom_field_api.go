package api

import (
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/repository"
	"BaseGoUni/core/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"strconv"
)

// GetSysCustomFields godoc
//
//	@Summary		获取自定义字段列表
//	@Tags			自定义字段
//	@Accept			json
//	@Produce		json
//	@Param			data body		pojo.SysCustomFieldSearch	true	"查询条件"
//	@Success		200	{object}		pojo.SysCustomFieldResp
//	@Router			/api/v1/admin/sysCustomField/list [post]
func GetSysCustomFields(ctx *gin.Context) {
	var search pojo.SysCustomFieldSearch
	search.SetPageDefaults()
	if err := ctx.ShouldBindJSON(&search); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result := repository.GetSysCustomFields(db, search)
	utils.SuccessObjBack(ctx, result)
}

// SetSysCustomField godoc
//
//	@Summary		创建或更新自定义字段
//	@Tags			自定义字段
//	@Accept			json
//	@Produce		json
//	@Param			data body		pojo.SysCustomFieldSet	true	"自定义字段信息"
//	@Success		200	{object}		pojo.SysCustomFieldBack
//	@Router			/api/v1/admin/sysCustomField [post]
func SetSysCustomField(ctx *gin.Context) {
	var req pojo.SysCustomFieldSet
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result, err := repository.SetSysCustomField(db, req)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

// DelSysCustomField godoc
//
//	@Summary		删除自定义字段
//	@Tags			自定义字段
//	@Accept			json
//	@Produce		json
//	@Param			id path		int	true	"自定义字段ID"
//	@Success		200	{object}		string
//	@Router			/api/v1/admin/sysCustomField/:id [delete]
func DelSysCustomField(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.ErrorBack(ctx, "参数格式错误")
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result, err := repository.DelSysCustomField(db, id)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

// GetSysCustomFieldById godoc
//
//	@Summary		根据ID获取自定义字段
//	@Tags			自定义字段
//	@Accept			json
//	@Produce		json
//	@Param			id path		int	true	"自定义字段ID"
//	@Success		200	{object}		pojo.SysCustomFieldBack
//	@Router			/api/v1/admin/sysCustomField/:id [get]
func GetSysCustomFieldById(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.ErrorBack(ctx, "参数格式错误")
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result, err := repository.GetSysCustomFieldById(db, id)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}
