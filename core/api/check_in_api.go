package api

import (
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/repository"
	"BaseGoUni/core/utils"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GetCurrentCheckInStatus 获取当前用户签到状态。
func GetCurrentCheckInStatus(ctx *gin.Context) {
	userID := ctx.MustGet("userId").(int64)
	db := ctx.MustGet("db").(*gorm.DB)

	result, err := repository.GetCurrentCheckInStatus(db, userID, time.Now())
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

// DoCurrentUserCheckIn 当前用户签到。
func DoCurrentUserCheckIn(ctx *gin.Context) {
	userID := ctx.MustGet("userId").(int64)
	tenantID := ctx.MustGet("tenantId").(int64)
	db := ctx.MustGet("db").(*gorm.DB)

	result, err := repository.DoCurrentUserCheckIn(db, tenantID, userID, time.Now())
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

// GetCurrentUserCheckInRecords 获取当前用户最近签到记录。
func GetCurrentUserCheckInRecords(ctx *gin.Context) {
	userID := ctx.MustGet("userId").(int64)
	db := ctx.MustGet("db").(*gorm.DB)

	limit, err := strconv.Atoi(ctx.DefaultQuery("limit", "30"))
	if err != nil || limit <= 0 {
		limit = 30
	}

	result, err := repository.GetCurrentUserCheckInRecords(db, userID, limit)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

// GetCheckInRecordsAdmin 获取管理端签到记录列表。
func GetCheckInRecordsAdmin(ctx *gin.Context) {
	var search pojo.TgUserCheckInRecordSearch
	search.SetPageDefaults()
	if err := ctx.ShouldBindJSON(&search); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}

	db := ctx.MustGet("db").(*gorm.DB)
	result, err := repository.GetCheckInRecordsAdmin(db, search)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}
