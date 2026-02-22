package api

import (
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/utils"
	tenantRepo "BaseGoUni/tenant/repository"
	"github.com/gin-gonic/gin"
)

func GetTgUserRebateRecords(ctx *gin.Context) {
	tenantID, ok := getTenantID(ctx)
	if !ok {
		return
	}
	var search pojo.TgUserRebateRecordSearch
	search.SetPageDefaults()
	if err := ctx.ShouldBindJSON(&search); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	result := tenantRepo.GetTgUserRebateRecords(getDB(ctx), tenantID, search)
	utils.SuccessObjBack(ctx, result)
}

func GetTgUserRebateRecordById(ctx *gin.Context) {
	tenantID, ok := getTenantID(ctx)
	if !ok {
		return
	}
	id, ok := parsePathID(ctx)
	if !ok {
		return
	}
	result, err := tenantRepo.GetTgUserRebateRecordByID(getDB(ctx), tenantID, id)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

func SetTgUserRebateRecord(ctx *gin.Context) {
	tenantID, ok := getTenantID(ctx)
	if !ok {
		return
	}
	var req pojo.TgUserRebateRecordSet
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	result, err := tenantRepo.SetTgUserRebateRecord(getDB(ctx), tenantID, req)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

func DelTgUserRebateRecord(ctx *gin.Context) {
	tenantID, ok := getTenantID(ctx)
	if !ok {
		return
	}
	id, ok := parsePathID(ctx)
	if !ok {
		return
	}
	result, err := tenantRepo.DelTgUserRebateRecord(getDB(ctx), tenantID, id)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}
