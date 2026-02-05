package api

import (
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/repository"
	"BaseGoUni/core/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"strconv"
)

// GetWithdrawOrderBrs godoc
//
//	@Summary		获取巴西提现订单列表
//	@Tags			巴西提现订单
//	@Accept			json
//	@Produce		json
//	@Param			data body		pojo.WithdrawOrderBrSearch	true	"查询条件"
//	@Success		200	{object}		pojo.WithdrawOrderBrResp
//	@Router			/api/v1/admin/withdrawOrderBr/list [post]
func GetWithdrawOrderBrs(ctx *gin.Context) {
	var search pojo.WithdrawOrderBrSearch
	search.SetPageDefaults()
	if err := ctx.ShouldBindJSON(&search); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result := repository.GetWithdrawOrderBrs(db, search)
	utils.SuccessObjBack(ctx, result)
}

// SetWithdrawOrderBr godoc
//
//	@Summary		创建或更新巴西提现订单
//	@Tags			巴西提现订单
//	@Accept			json
//	@Produce		json
//	@Param			data body		pojo.WithdrawOrderBrSet	true	"提现订单信息"
//	@Success		200	{object}		pojo.WithdrawOrderBrBack
//	@Router			/api/v1/admin/withdrawOrderBr [post]
func SetWithdrawOrderBr(ctx *gin.Context) {
	var req pojo.WithdrawOrderBrSet
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result, err := repository.SetWithdrawOrderBr(db, req)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

// DelWithdrawOrderBr godoc
//
//	@Summary		删除巴西提现订单
//	@Tags			巴西提现订单
//	@Accept			json
//	@Produce		json
//	@Param			id path		int	true	"提现订单ID"
//	@Success		200	{object}		string
//	@Router			/api/v1/admin/withdrawOrderBr/:id [delete]
func DelWithdrawOrderBr(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.ErrorBack(ctx, "参数格式错误")
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result, err := repository.DelWithdrawOrderBr(db, id)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

// GetWithdrawOrderBrById godoc
//
//	@Summary		根据ID获取巴西提现订单
//	@Tags			巴西提现订单
//	@Accept			json
//	@Produce		json
//	@Param			id path		int	true	"提现订单ID"
//	@Success		200	{object}		pojo.WithdrawOrderBrBack
//	@Router			/api/v1/admin/withdrawOrderBr/:id [get]
func GetWithdrawOrderBrById(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.ErrorBack(ctx, "参数格式错误")
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result, err := repository.GetWithdrawOrderBrById(db, id)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}
