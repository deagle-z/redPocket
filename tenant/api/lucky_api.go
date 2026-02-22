package api

import (
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/utils"
	tenantRepo "BaseGoUni/tenant/repository"
	"github.com/gin-gonic/gin"
)

func GetLuckyMoneyList(ctx *gin.Context) {
	tenantID, ok := getTenantID(ctx)
	if !ok {
		return
	}
	var search pojo.LuckyMoneySearch
	search.SetPageDefaults()
	if err := ctx.ShouldBindJSON(&search); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	result := tenantRepo.GetLuckyMoneyList(getDB(ctx), tenantID, search)
	utils.SuccessObjBack(ctx, result)
}

func GetLuckyHistoryList(ctx *gin.Context) {
	tenantID, ok := getTenantID(ctx)
	if !ok {
		return
	}
	var search pojo.LuckyHistorySearch
	search.SetPageDefaults()
	if err := ctx.ShouldBindJSON(&search); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	result := tenantRepo.GetLuckyHistoryList(getDB(ctx), tenantID, search)
	utils.SuccessObjBack(ctx, result)
}

func GetLuckyMoneyDetail(ctx *gin.Context) {
	tenantID, ok := getTenantID(ctx)
	if !ok {
		return
	}
	id, ok := parsePathID(ctx)
	if !ok {
		return
	}
	luckyMoney, err := tenantRepo.GetLuckyMoneyByID(getDB(ctx), tenantID, id)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	historyList, err := tenantRepo.GetLuckyHistoryByLuckyID(getDB(ctx), tenantID, id)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, map[string]any{
		"luckyMoney": luckyMoney,
		"history":    historyList,
	})
}
