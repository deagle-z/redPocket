package api

import (
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/repository"
	"BaseGoUni/core/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"strconv"
)

// GetPkManagers godoc
//
//	@Summary		获取PK管理器列表
//	@Tags			PK管理器
//	@Accept			json
//	@Produce		json
//	@Param			data body		pojo.PkManagerSearch	true	"查询条件"
//	@Success		200	{object}		pojo.PkManagerResp
//	@Router			/api/v1/outside/pkManagers [post]
func GetPkManagers(ctx *gin.Context) {
	var search pojo.PkManagerSearch
	search.SetPageDefaults()
	err := ctx.BindJSON(&search)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result := repository.GetPkManagers(db, search)
	utils.SuccessObjBack(ctx, result)
}

// SetPkManager godoc
//
//	@Summary		创建或更新PK管理器
//	@Tags			PK管理器
//	@Accept			json
//	@Produce		json
//	@Param			data body		pojo.PkManagerSet	true	"PK管理器信息"
//	@Success		200	{object}		pojo.PkManagerBack
//	@Router			/api/v1/manager/pkManager [post]
func SetPkManager(ctx *gin.Context) {
	var pkManagerSet pojo.PkManagerSet
	err := ctx.ShouldBindJSON(&pkManagerSet)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result, err := repository.SetPkManager(db, pkManagerSet)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

// DelPkManager godoc
//
//	@Summary		删除PK管理器
//	@Tags			PK管理器
//	@Accept			json
//	@Produce		json
//	@Param			id path		int	true	"PK管理器ID"
//	@Success		200	{object}		string
//	@Router			/api/v1/manager/pkManager/:id [delete]
func DelPkManager(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.ErrorBack(ctx, "参数格式错误")
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result, err := repository.DelPkManager(db, id)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

// GetPkManagerById godoc
//
//	@Summary		根据ID获取PK管理器
//	@Tags			PK管理器
//	@Accept			json
//	@Produce		json
//	@Param			id path		int	true	"PK管理器ID"
//	@Success		200	{object}		pojo.PkManagerBack
//	@Router			/api/v1/outside/pkManager/:id [get]
func GetPkManagerById(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.ErrorBack(ctx, "参数格式错误")
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result, err := repository.GetPkManagerById(db, id)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

// GetPkManagerUrlByName godoc
//
//	@Summary		根据名称获取PK管理器的URL
//	@Tags			PK管理器
//	@Accept			json
//	@Produce		json
//	@Param			name path		string	true	"PK管理器名称"
//	@Success		200	{object}		map[string]string
//	@Router			/api/v1/app/getPkByName/:name [get]
func GetPkManagerUrlByName(ctx *gin.Context) {
	name := ctx.Param("name")
	if name == "" {
		utils.ErrorBack(ctx, "名称参数不能为空")
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	url, err := repository.GetPkManagerUrlByName(db, name)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, map[string]string{"url": url})
}
