package api

import (
	"BaseGoUni/core/pojo"
	coreRepo "BaseGoUni/core/repository"
	"BaseGoUni/core/utils"
	"github.com/gin-gonic/gin"
)

func GetSysPayChannels(ctx *gin.Context) {
	if _, ok := getTenantID(ctx); !ok {
		return
	}
	var search pojo.SysPayChannelSearch
	search.SetPageDefaults()
	if err := ctx.ShouldBindJSON(&search); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	result := coreRepo.GetSysPayChannels(getDB(ctx), search)
	utils.SuccessObjBack(ctx, result)
}
