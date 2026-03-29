package api

import (
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/repository"
	"BaseGoUni/core/utils"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GetPrizePoolConfig 根据 poolId 获取概率配置
func GetPrizePoolConfig(ctx *gin.Context) {
	poolIdStr := ctx.Param("poolId")
	poolId, err := strconv.ParseInt(poolIdStr, 10, 64)
	if err != nil || poolId <= 0 {
		utils.ErrorBack(ctx, "参数格式错误")
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result, err := repository.GetPrizePoolConfigByPoolId(db, poolId)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

// SetPrizePoolConfig 创建或更新奖池概率配置
func SetPrizePoolConfig(ctx *gin.Context) {
	var req pojo.SysTenantPrizePoolConfigSet
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result, err := repository.SetPrizePoolConfig(db, req)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

// DelPrizePoolConfig 删除奖池概率配置
func DelPrizePoolConfig(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		utils.ErrorBack(ctx, "参数格式错误")
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	if err = repository.DelPrizePoolConfig(db, id); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessBack(ctx, "删除成功")
}
