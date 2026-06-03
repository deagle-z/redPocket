package api

import (
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/repository"
	"BaseGoUni/core/utils"
	"io"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GetAppUserBetRecords godoc
//
//	@Summary		获取三方游戏下注记录
//	@Tags			游戏下注记录
//	@Accept			json
//	@Produce		json
//	@Param			data body		pojo.AppUserBetRecordSearch	true	"查询条件"
//	@Success		200	{object}		pojo.AppUserBetRecordResp
//	@Router			/api/v1/admin/appUserBetRecord/list [post]
func GetAppUserBetRecords(ctx *gin.Context) {
	var search pojo.AppUserBetRecordSearch
	search.SetPageDefaults()
	if err := ctx.ShouldBindJSON(&search); err != nil && err != io.EOF {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	if search.UserID == nil && search.UID == nil && search.ID <= 0 && search.TraceID == "" && (search.StartTime <= 0 || search.EndTime <= 0) {
		utils.ErrorBack(ctx, "非用户维度查询必须选择时间范围")
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	utils.SuccessObjBack(ctx, repository.GetAppUserBetRecords(db, search))
}

// GetAppCurrentUserBetRecords app端获取当前用户三方游戏下注记录
//
//	@Summary		App端获取当前用户三方游戏下注记录
//	@Tags			游戏下注记录
//	@Accept			json
//	@Produce		json
//	@Param			data body		pojo.AppUserBetRecordAppSearch	true	"查询条件"
//	@Success		200	{object}		pojo.AppUserBetRecordResp
//	@Router			/api/v1/app/appUserBetRecord/list [post]
func GetAppCurrentUserBetRecords(ctx *gin.Context) {
	userIDRaw, ok := ctx.Get("userId")
	if !ok {
		utils.UnauthorizedBack(ctx, "token_invalid")
		return
	}
	userID, ok := userIDRaw.(int64)
	if !ok || userID <= 0 {
		utils.UnauthorizedBack(ctx, "token_invalid")
		return
	}

	var search pojo.AppUserBetRecordAppSearch
	search.SetPageDefaults()
	if err := ctx.ShouldBindJSON(&search); err != nil && err != io.EOF {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result, err := repository.GetAppUserBetRecordsApp(db, userID, search)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}
