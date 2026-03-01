package api

import (
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/repository"
	"BaseGoUni/core/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"strconv"
)

// GetPayChannels godoc
//
//	@Summary		获取支付通道列表
//	@Tags			支付通道
//	@Accept			json
//	@Produce		json
//	@Param			data body		pojo.PayChannelSearch	true	"查询条件"
//	@Success		200	{object}		pojo.PayChannelResp
//	@Router			/api/v1/admin/payChannel/list [post]
func GetPayChannels(ctx *gin.Context) {
	var search pojo.PayChannelSearch
	search.SetPageDefaults()
	if err := ctx.ShouldBindJSON(&search); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result := repository.GetPayChannels(db, search)
	utils.SuccessObjBack(ctx, result)
}

// SetPayChannel godoc
//
//	@Summary		创建或更新支付通道
//	@Tags			支付通道
//	@Accept			json
//	@Produce		json
//	@Param			data body		pojo.PayChannelSet	true	"支付通道信息"
//	@Success		200	{object}		pojo.PayChannelBack
//	@Router			/api/v1/admin/payChannel [post]
func SetPayChannel(ctx *gin.Context) {
	var req pojo.PayChannelSet
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result, err := repository.SetPayChannel(db, req)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

// DelPayChannel godoc
//
//	@Summary		删除支付通道
//	@Tags			支付通道
//	@Accept			json
//	@Produce		json
//	@Param			id path		int	true	"支付通道ID"
//	@Success		200	{object}		string
//	@Router			/api/v1/admin/payChannel/:id [delete]
func DelPayChannel(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.ErrorBack(ctx, "参数格式错误")
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result, err := repository.DelPayChannel(db, id)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

// GetPayChannelById godoc
//
//	@Summary		根据ID获取支付通道
//	@Tags			支付通道
//	@Accept			json
//	@Produce		json
//	@Param			id path		int	true	"支付通道ID"
//	@Success		200	{object}		pojo.PayChannelBack
//	@Router			/api/v1/admin/payChannel/:id [get]
func GetPayChannelById(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.ErrorBack(ctx, "参数格式错误")
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result, err := repository.GetPayChannelById(db, id)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}
