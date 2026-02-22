package api

import (
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/utils"
	tenantRepo "BaseGoUni/tenant/repository"
	"github.com/gin-gonic/gin"
)

func GetRechargeOrders(ctx *gin.Context) {
	tenantID, ok := getTenantID(ctx)
	if !ok {
		return
	}
	var search pojo.RechargeOrderSearch
	search.SetPageDefaults()
	if err := ctx.ShouldBindJSON(&search); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	result := tenantRepo.GetRechargeOrders(getDB(ctx), tenantID, search)
	utils.SuccessObjBack(ctx, result)
}

func GetRechargeOrderById(ctx *gin.Context) {
	tenantID, ok := getTenantID(ctx)
	if !ok {
		return
	}
	id, ok := parsePathID(ctx)
	if !ok {
		return
	}
	result, err := tenantRepo.GetRechargeOrderByID(getDB(ctx), tenantID, id)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

func SetRechargeOrder(ctx *gin.Context) {
	tenantID, ok := getTenantID(ctx)
	if !ok {
		return
	}
	var req pojo.RechargeOrderSet
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	result, err := tenantRepo.SetRechargeOrder(getDB(ctx), tenantID, req)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

func DelRechargeOrder(ctx *gin.Context) {
	tenantID, ok := getTenantID(ctx)
	if !ok {
		return
	}
	id, ok := parsePathID(ctx)
	if !ok {
		return
	}
	result, err := tenantRepo.DelRechargeOrder(getDB(ctx), tenantID, id)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}
