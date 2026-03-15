package api

import (
	"BaseGoUni/core/repository"
	"BaseGoUni/core/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"strconv"
)

type SetChannelMethodsReq struct {
	ChannelID int64   `json:"channelId"`
	MethodIds []int64 `json:"methodIds"`
}

// GetSysPayChannelMethods godoc
//
//	@Summary		获取通道绑定的支付方式
//	@Tags			支付通道方式
//	@Accept			json
//	@Produce		json
//	@Param			channelId path	int	true	"支付通道ID"
//	@Router			/api/v1/admin/sysPayChannelMethod/:channelId [get]
func GetSysPayChannelMethods(ctx *gin.Context) {
	idStr := ctx.Param("channelId")
	channelId, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.ErrorBack(ctx, "参数格式错误")
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result, err := repository.GetSysPayChannelMethods(db, channelId)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

// SetSysPayChannelMethods godoc
//
//	@Summary		设置通道支持的支付方式（全量覆盖）
//	@Tags			支付通道方式
//	@Accept			json
//	@Produce		json
//	@Param			data body		SetChannelMethodsReq	true	"通道ID及方式ID列表"
//	@Router			/api/v1/admin/sysPayChannelMethod [post]
func SetSysPayChannelMethods(ctx *gin.Context) {
	var req SetChannelMethodsReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	if err := repository.SetSysPayChannelMethods(db, req.ChannelID, req.MethodIds); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, "success")
}

// DelSysPayChannelMethod godoc
//
//	@Summary		删除单条通道-方式绑定
//	@Tags			支付通道方式
//	@Accept			json
//	@Produce		json
//	@Param			id path		int	true	"绑定记录ID"
//	@Router			/api/v1/admin/sysPayChannelMethod/:id [delete]
func DelSysPayChannelMethod(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.ErrorBack(ctx, "参数格式错误")
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result, err := repository.DelSysPayChannelMethod(db, id)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}
