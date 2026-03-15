package api

import (
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/repository"
	"BaseGoUni/core/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"strconv"
)

// GetSysPayChannels godoc
//
//	@Summary		获取支付通道列表
//	@Tags			支付通道
//	@Accept			json
//	@Produce		json
//	@Param			data body		pojo.SysPayChannelSearch	true	"查询条件"
//	@Success		200	{object}		pojo.SysPayChannelResp
//	@Router			/api/v1/admin/sysPayChannel/list [post]
func GetSysPayChannels(ctx *gin.Context) {
	var search pojo.SysPayChannelSearch
	search.SetPageDefaults()
	if err := ctx.ShouldBindJSON(&search); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result := repository.GetSysPayChannels(db, search)
	utils.SuccessObjBack(ctx, result)
}

// SetSysPayChannel godoc
//
//	@Summary		创建或更新支付通道
//	@Tags			支付通道
//	@Accept			json
//	@Produce		json
//	@Param			data body		pojo.SysPayChannelSet	true	"支付通道信息"
//	@Success		200	{object}		pojo.SysPayChannelBack
//	@Router			/api/v1/admin/sysPayChannel [post]
func SetSysPayChannel(ctx *gin.Context) {
	var req pojo.SysPayChannelSet
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result, err := repository.SetSysPayChannel(db, req)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

// DelSysPayChannel godoc
//
//	@Summary		删除支付通道
//	@Tags			支付通道
//	@Accept			json
//	@Produce		json
//	@Param			id path		int	true	"支付通道ID"
//	@Success		200	{object}		string
//	@Router			/api/v1/admin/sysPayChannel/:id [delete]
func DelSysPayChannel(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.ErrorBack(ctx, "参数格式错误")
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result, err := repository.DelSysPayChannel(db, id)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

// GetSysPayChannelById godoc
//
//	@Summary		根据ID获取支付通道
//	@Tags			支付通道
//	@Accept			json
//	@Produce		json
//	@Param			id path		int	true	"支付通道ID"
//	@Success		200	{object}		pojo.SysPayChannelBack
//	@Router			/api/v1/admin/sysPayChannel/:id [get]
func GetSysPayChannelById(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.ErrorBack(ctx, "参数格式错误")
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result, err := repository.GetSysPayChannelById(db, id)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}
