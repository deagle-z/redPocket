package api

import (
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/utils"
	tenantRepo "BaseGoUni/tenant/repository"
	"github.com/gin-gonic/gin"
)

func GetWithdrawOrderBrs(ctx *gin.Context) {
	tenantID, ok := getTenantID(ctx)
	if !ok {
		return
	}
	var search pojo.WithdrawOrderBrSearch
	search.SetPageDefaults()
	if err := ctx.ShouldBindJSON(&search); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	result := tenantRepo.GetWithdrawOrderBrs(getDB(ctx), tenantID, search)
	utils.SuccessObjBack(ctx, result)
}

func GetWithdrawOrderBrById(ctx *gin.Context) {
	tenantID, ok := getTenantID(ctx)
	if !ok {
		return
	}
	id, ok := parsePathID(ctx)
	if !ok {
		return
	}
	result, err := tenantRepo.GetWithdrawOrderBrByID(getDB(ctx), tenantID, id)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

func SetWithdrawOrderBr(ctx *gin.Context) {
	tenantID, ok := getTenantID(ctx)
	if !ok {
		return
	}
	var req pojo.WithdrawOrderBrSet
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	result, err := tenantRepo.SetWithdrawOrderBr(getDB(ctx), tenantID, req)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

func DelWithdrawOrderBr(ctx *gin.Context) {
	tenantID, ok := getTenantID(ctx)
	if !ok {
		return
	}
	id, ok := parsePathID(ctx)
	if !ok {
		return
	}
	result, err := tenantRepo.DelWithdrawOrderBr(getDB(ctx), tenantID, id)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}
