package api

import (
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/repository"
	"BaseGoUni/core/utils"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// CreateAttributionEvent godoc
//
//	@Summary		上报事件归因
//	@Tags			事件归因
//	@Accept			json
//	@Produce		json
//	@Param			data body		pojo.AttributionEventCreateReq	true	"事件信息"
//	@Success		200	{object}		pojo.AttributionEventBack
//	@Router			/api/v1/app/attribution/event [post]
func CreateAttributionEvent(ctx *gin.Context) {
	var req pojo.AttributionEventCreateReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}

	hostInfo := ctx.MustGet("hostInfo").(pojo.HostInfo)
	db := ctx.MustGet("db").(*gorm.DB)
	userID, tenantID := parseOptionalAppToken(ctx, hostInfo)

	result, err := repository.CreateAttributionEvent(
		db,
		req,
		userID,
		tenantID,
		utils.GetIPAddress(ctx),
		ctx.GetHeader("User-Agent"),
	)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

// GetAttributionEvents godoc
//
//	@Summary		事件归因明细列表
//	@Tags			事件归因
//	@Accept			json
//	@Produce		json
//	@Param			data body		pojo.AttributionEventSearch	true	"查询条件"
//	@Success		200	{object}		pojo.AttributionEventResp
//	@Router			/api/v1/admin/attributionEvent/list [post]
func GetAttributionEvents(ctx *gin.Context) {
	var search pojo.AttributionEventSearch
	search.SetPageDefaults()
	if err := ctx.ShouldBindJSON(&search); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	utils.SuccessObjBack(ctx, repository.GetAttributionEvents(db, search))
}

// GetAttributionEventSummary godoc
//
//	@Summary		事件归因汇总
//	@Tags			事件归因
//	@Accept			json
//	@Produce		json
//	@Param			data body		pojo.AttributionEventSummarySearch	true	"查询条件"
//	@Success		200	{array}		pojo.AttributionEventSummaryBack
//	@Router			/api/v1/admin/attributionEvent/summary [post]
func GetAttributionEventSummary(ctx *gin.Context) {
	var search pojo.AttributionEventSummarySearch
	if err := ctx.ShouldBindJSON(&search); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result, err := repository.GetAttributionEventSummary(db, search)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

func parseOptionalAppToken(ctx *gin.Context, hostInfo pojo.HostInfo) (int64, int64) {
	authHeader := strings.TrimSpace(ctx.GetHeader("Authorization"))
	if authHeader == "" {
		return 0, 0
	}
	if strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
		authHeader = strings.TrimSpace(authHeader[7:])
	}
	userID, hostName, tenantID, err := utils.ParseAppToken(hostInfo.AccessSecret, authHeader)
	if err != nil || userID <= 0 || hostName != hostInfo.HostName {
		return 0, 0
	}
	return userID, tenantID
}
