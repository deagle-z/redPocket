package api

import (
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/repository"
	"BaseGoUni/core/utils"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GetLuckyMoneyListAdmin 管理员获取红包列表
// @Summary 管理员获取红包列表
// @Tags 红包管理-管理员
// @Accept json
// @Produce json
// @Param data body pojo.LuckyMoneySearch true "查询条件"
// @Success 200 {object} pojo.LuckyMoneyResp
// @Router /api/v1/admin/lucky/list [post]
func GetLuckyMoneyListAdmin(ctx *gin.Context) {
	var search pojo.LuckyMoneySearch
	search.SetPageDefaults()
	if err := ctx.ShouldBindJSON(&search); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}

	db := ctx.MustGet("db").(*gorm.DB)

	result := repository.GetLuckyMoneyList(db, search)
	utils.SuccessObjBack(ctx, result)
}

// GetLuckyHistoryListAdmin 管理员获取领取历史记录
// @Summary 管理员获取领取历史记录
// @Tags 红包管理-管理员
// @Accept json
// @Produce json
// @Param data body pojo.LuckyHistorySearch true "查询条件"
// @Success 200 {object} pojo.LuckyHistoryResp
// @Router /api/v1/admin/lucky/history [post]
func GetLuckyHistoryListAdmin(ctx *gin.Context) {
	var search pojo.LuckyHistorySearch
	search.SetPageDefaults()
	if err := ctx.ShouldBindJSON(&search); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}

	db := ctx.MustGet("db").(*gorm.DB)

	result := repository.GetLuckyHistoryList(db, search)
	utils.SuccessObjBack(ctx, result)
}

// GetLuckyMoneyDetailAdmin 管理员获取红包详情
// @Summary 管理员获取红包详情
// @Tags 红包管理-管理员
// @Accept json
// @Produce json
// @Param id path int true "红包ID"
// @Success 200 {object} pojo.BaseResponse
// @Router /api/v1/admin/lucky/:id [get]
func GetLuckyMoneyDetailAdmin(ctx *gin.Context) {
	luckyIDStr := ctx.Param("id")
	luckyID := int64(0)
	_, err := fmt.Sscanf(luckyIDStr, "%d", &luckyID)
	if err != nil || luckyID <= 0 {
		utils.ErrorBack(ctx, "无效的红包ID")
		return
	}

	db := ctx.MustGet("db").(*gorm.DB)

	luckyMoney, err := repository.GetLuckyMoney(db, luckyID)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}

	historyList, err := repository.GetLuckyHistoryByLuckyId(db, luckyID)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}

	result := map[string]interface{}{
		"luckyMoney": luckyMoney,
		"history":    historyList,
	}

	utils.SuccessObjBack(ctx, result)
}

// GetCashHistoryListAdmin 管理员获取余额变动记录列表
// @Summary 管理员获取余额变动记录列表
// @Tags 红包管理-管理员
// @Accept json
// @Produce json
// @Param data body pojo.CashHistorySearch true "查询条件"
// @Success 200 {object} pojo.CashHistoryPage
// @Router /api/v1/admin/cashHistory/list [post]
func GetCashHistoryListAdmin(ctx *gin.Context) {
	var search pojo.CashHistorySearch
	search.SetPageDefaults()
	if err := ctx.ShouldBindJSON(&search); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}

	db := ctx.MustGet("db").(*gorm.DB)

	result := repository.GetCashHistoryListAdmin(db, search)
	utils.SuccessObjBack(ctx, result)
}
