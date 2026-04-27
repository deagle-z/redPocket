package api

import (
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/utils"
	tenantRepo "BaseGoUni/tenant/repository"

	"github.com/gin-gonic/gin"
)

func GetCashHistoryList(ctx *gin.Context) {
	tenantID, ok := getTenantID(ctx)
	if !ok {
		return
	}

	var search pojo.CashHistorySearch
	search.SetPageDefaults()
	if err := ctx.ShouldBindJSON(&search); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}

	result := tenantRepo.GetCashHistoryList(getDB(ctx), tenantID, search)
	utils.SuccessObjBack(ctx, result)
}
