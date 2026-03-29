package api

import (
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/repository"
	"BaseGoUni/core/utils"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)


// GetUserLotteryRecords 抽奖记录列表
func GetUserLotteryRecords(ctx *gin.Context) {
	var search pojo.UserLotteryRecordSearch
	search.SetPageDefaults()
	if err := ctx.ShouldBindJSON(&search); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result := repository.GetUserLotteryRecords(db, search)
	utils.SuccessObjBack(ctx, result)
}

// GetUserLotteryRecordById 抽奖记录详情
func GetUserLotteryRecordById(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		utils.ErrorBack(ctx, "参数格式错误")
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result, err := repository.GetUserLotteryRecordById(db, id)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

