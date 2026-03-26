package api

import (
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/repository"
	"BaseGoUni/core/utils"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GetSysSourceChannels godoc
//
//	@Summary		获取投流来源渠道列表
//	@Tags			投流来源渠道
//	@Accept			json
//	@Produce		json
//	@Param			data body		pojo.SysSourceChannelSearch	true	"查询条件"
//	@Success		200	{object}		pojo.SysSourceChannelResp
//	@Router			/api/v1/admin/sysSourceChannel/list [post]
func GetSysSourceChannels(ctx *gin.Context) {
	var search pojo.SysSourceChannelSearch
	search.SetPageDefaults()
	if err := ctx.ShouldBindJSON(&search); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result := repository.GetSysSourceChannels(db, search)
	utils.SuccessObjBack(ctx, result)
}

// SetSysSourceChannel godoc
//
//	@Summary		创建或更新投流来源渠道
//	@Tags			投流来源渠道
//	@Accept			json
//	@Produce		json
//	@Param			data body		pojo.SysSourceChannelSet	true	"渠道信息"
//	@Success		200	{object}		pojo.SysSourceChannelBack
//	@Router			/api/v1/admin/sysSourceChannel [post]
func SetSysSourceChannel(ctx *gin.Context) {
	var req pojo.SysSourceChannelSet
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result, err := repository.SetSysSourceChannel(db, req)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

// DelSysSourceChannel godoc
//
//	@Summary		删除投流来源渠道
//	@Tags			投流来源渠道
//	@Accept			json
//	@Produce		json
//	@Param			id path		int	true	"渠道ID"
//	@Success		200	{object}		string
//	@Router			/api/v1/admin/sysSourceChannel/:id [delete]
func DelSysSourceChannel(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.ErrorBack(ctx, "参数格式错误")
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result, err := repository.DelSysSourceChannel(db, id)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

// GetSysSourceChannelById godoc
//
//	@Summary		根据ID获取投流来源渠道
//	@Tags			投流来源渠道
//	@Accept			json
//	@Produce		json
//	@Param			id path		int	true	"渠道ID"
//	@Success		200	{object}		pojo.SysSourceChannelBack
//	@Router			/api/v1/admin/sysSourceChannel/:id [get]
func GetSysSourceChannelById(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.ErrorBack(ctx, "参数格式错误")
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result, err := repository.GetSysSourceChannelById(db, id)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}
