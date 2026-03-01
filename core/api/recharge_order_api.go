package api

import (
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/repository"
	"BaseGoUni/core/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"strconv"
)

// GetRechargeOrders godoc
//
//	@Summary		获取充值订单列表
//	@Tags			充值订单
//	@Accept			json
//	@Produce		json
//	@Param			data body		pojo.RechargeOrderSearch	true	"查询条件"
//	@Success		200	{object}		pojo.RechargeOrderResp
//	@Router			/api/v1/admin/rechargeOrder/list [post]
func GetRechargeOrders(ctx *gin.Context) {
	var search pojo.RechargeOrderSearch
	search.SetPageDefaults()
	if err := ctx.ShouldBindJSON(&search); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result := repository.GetRechargeOrders(db, search)
	utils.SuccessObjBack(ctx, result)
}

// SetRechargeOrder godoc
//
//	@Summary		创建或更新充值订单
//	@Tags			充值订单
//	@Accept			json
//	@Produce		json
//	@Param			data body		pojo.RechargeOrderSet	true	"充值订单信息"
//	@Success		200	{object}		pojo.RechargeOrderBack
//	@Router			/api/v1/admin/rechargeOrder [post]
func SetRechargeOrder(ctx *gin.Context) {
	var req pojo.RechargeOrderSet
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result, err := repository.SetRechargeOrder(db, req)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

// DelRechargeOrder godoc
//
//	@Summary		删除充值订单
//	@Tags			充值订单
//	@Accept			json
//	@Produce		json
//	@Param			id path		int	true	"充值订单ID"
//	@Success		200	{object}		string
//	@Router			/api/v1/admin/rechargeOrder/:id [delete]
func DelRechargeOrder(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.ErrorBack(ctx, "参数格式错误")
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result, err := repository.DelRechargeOrder(db, id)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

// GetRechargeOrderById godoc
//
//	@Summary		根据ID获取充值订单
//	@Tags			充值订单
//	@Accept			json
//	@Produce		json
//	@Param			id path		int	true	"充值订单ID"
//	@Success		200	{object}		pojo.RechargeOrderBack
//	@Router			/api/v1/admin/rechargeOrder/:id [get]
func GetRechargeOrderById(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.ErrorBack(ctx, "参数格式错误")
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result, err := repository.GetRechargeOrderById(db, id)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

// AppCreateRechargeOrder godoc
//
//	@Summary		app端创建充值订单
//	@Tags			充值订单
//	@Accept			json
//	@Produce		json
//	@Param			data body		pojo.RechargeOrderAppReq	true	"充值下单参数"
//	@Success		200	{object}		pojo.RechargeOrderAppBack
//	@Router			/api/v1/app/rechargeOrder [post]
func AppCreateRechargeOrder(ctx *gin.Context) {
	var req pojo.RechargeOrderAppReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ErrorBack(ctx, "参数格式错误")
		return
	}
	userIDRaw, ok := ctx.Get("userId")
	if !ok {
		utils.UnauthorizedBack(ctx, "token is invalid")
		return
	}
	userID, ok := userIDRaw.(int64)
	if !ok || userID <= 0 {
		utils.UnauthorizedBack(ctx, "token is invalid")
		return
	}

	db := ctx.MustGet("db").(*gorm.DB)
	result, err := repository.AppCreateRechargeOrder(db, userID, req)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}
