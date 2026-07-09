package api

import (
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/repository"
	"BaseGoUni/core/utils"
	"io"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GetAppTaskActivityList godoc
//
//	@Summary		App端获取任务活动列表
//	@Tags			App任务活动
//	@Produce		json
//	@Success		200	{object}	[]pojo.TgTaskActivityAppItem
//	@Router			/api/v1/app/taskActivity/list [get]
func GetAppTaskActivityList(ctx *gin.Context) {
	userID, ok := currentAppTaskActivityUserID(ctx)
	if !ok {
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result, err := repository.GetAppTaskActivityList(db, userID, taskActivityRequestLang(ctx))
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

// ClaimAppTaskActivity godoc
//
//	@Summary		App端领取任务活动
//	@Tags			App任务活动
//	@Produce		json
//	@Param			configId	path	int	true	"任务配置ID"
//	@Success		200	{object}	pojo.TgTaskActivityAppItem
//	@Router			/api/v1/app/taskActivity/:configId/claim [post]
func ClaimAppTaskActivity(ctx *gin.Context) {
	userID, ok := currentAppTaskActivityUserID(ctx)
	if !ok {
		return
	}
	configID, err := strconv.ParseInt(ctx.Param("configId"), 10, 64)
	if err != nil || configID <= 0 {
		utils.ErrorBack(ctx, "invalid_params")
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result, err := repository.ClaimAppTaskActivity(db, userID, configID, taskActivityRequestLang(ctx))
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

// GetAppCurrentTaskActivity godoc
//
//	@Summary		App端获取当前任务活动
//	@Tags			App任务活动
//	@Produce		json
//	@Success		200	{object}	pojo.TgTaskActivityAppItem
//	@Router			/api/v1/app/taskActivity/current [get]
func GetAppCurrentTaskActivity(ctx *gin.Context) {
	userID, ok := currentAppTaskActivityUserID(ctx)
	if !ok {
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result, err := repository.GetAppCurrentTaskActivity(db, userID, taskActivityRequestLang(ctx))
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

// GetAppTaskActivityRecords godoc
//
//	@Summary		App端获取我的任务活动历史
//	@Tags			App任务活动
//	@Accept			json
//	@Produce		json
//	@Param			data body		pojo.PageInfo	true	"分页条件"
//	@Success		200	{object}	pojo.TgTaskActivityAppRecordResp
//	@Router			/api/v1/app/taskActivity/records [post]
func GetAppTaskActivityRecords(ctx *gin.Context) {
	userID, ok := currentAppTaskActivityUserID(ctx)
	if !ok {
		return
	}
	var page pojo.PageInfo
	page.SetPageDefaults()
	if err := ctx.ShouldBindJSON(&page); err != nil && err != io.EOF {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result, err := repository.GetAppTaskActivityRecords(db, userID, page, taskActivityRequestLang(ctx))
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

func currentAppTaskActivityUserID(ctx *gin.Context) (int64, bool) {
	userIDRaw, ok := ctx.Get("userId")
	if !ok {
		utils.UnauthorizedBack(ctx, "token_invalid")
		return 0, false
	}
	userID, ok := userIDRaw.(int64)
	if !ok || userID <= 0 {
		utils.UnauthorizedBack(ctx, "token_invalid")
		return 0, false
	}
	return userID, true
}

func taskActivityRequestLang(ctx *gin.Context) string {
	lang := strings.TrimSpace(ctx.Query("lang"))
	if lang != "" {
		return lang
	}
	return strings.TrimSpace(ctx.GetHeader("Accept-Language"))
}
