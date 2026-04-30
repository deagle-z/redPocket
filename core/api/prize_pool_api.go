package api

import (
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/repository"
	"BaseGoUni/core/utils"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GetPrizePoolByCodeAdmin 根据奖池编码获取奖池信息
func GetPrizePoolByCodeAdmin(ctx *gin.Context) {
	poolCode := strings.TrimSpace(ctx.Param("poolCode"))
	if poolCode == "" {
		utils.ErrorBack(ctx, "参数格式错误")
		return
	}

	db := ctx.MustGet("db").(*gorm.DB)
	result, err := repository.GetPrizePoolByCode(db, poolCode)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

// SetPrizePoolBalanceAdmin 直接设置奖池余额
func SetPrizePoolBalanceAdmin(ctx *gin.Context) {
	var req pojo.SysTenantPrizePoolBalanceSet
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}

	db := ctx.MustGet("db").(*gorm.DB)
	result, err := repository.SetPrizePoolBalance(db, req)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

// GetPrizePoolOutRecordsApp app端查询奖池消耗流水，固定 change_type=out，不按商户过滤。
func GetPrizePoolOutRecordsApp(ctx *gin.Context) {
	var page pojo.PageInfo
	page.SetPageDefaults()
	if err := ctx.ShouldBindQuery(&page); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	if page.PageSize <= 0 || page.PageSize > 10 {
		page.PageSize = 10
	}
	if page.CurrentPage < 0 {
		page.CurrentPage = 0
	}

	db := ctx.MustGet("db").(*gorm.DB)
	result := repository.GetPrizePoolOutRecordsApp(db, page)
	utils.SuccessObjBack(ctx, result)
}
