package api

import (
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/repository"
	"BaseGoUni/core/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/url"
	"strconv"
	"strings"
)

// GetSysTenants godoc
//
//	@Summary		获取租户列表
//	@Tags			租户
//	@Accept			json
//	@Produce		json
//	@Param			data body		pojo.SysTenantSearch	true	"查询条件"
//	@Success		200	{object}		pojo.SysTenantResp
//	@Router			/api/v1/admin/tenant/list [post]
func GetSysTenants(ctx *gin.Context) {
	var search pojo.SysTenantSearch
	search.SetPageDefaults()
	if err := ctx.ShouldBindJSON(&search); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result := repository.GetSysTenants(db, search)
	utils.SuccessObjBack(ctx, result)
}

// SetSysTenant godoc
//
//	@Summary		创建或更新租户
//	@Tags			租户
//	@Accept			json
//	@Produce		json
//	@Param			data body		pojo.SysTenantSet	true	"租户信息"
//	@Success		200	{object}		pojo.SysTenantBack
//	@Router			/api/v1/admin/tenant [post]
func SetSysTenant(ctx *gin.Context) {
	var req pojo.SysTenantSet
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result, err := repository.SetSysTenant(db, req)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

// DelSysTenant godoc
//
//	@Summary		删除租户
//	@Tags			租户
//	@Accept			json
//	@Produce		json
//	@Param			id path		int	true	"租户ID"
//	@Success		200	{object}		string
//	@Router			/api/v1/admin/tenant/:id [delete]
func DelSysTenant(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.ErrorBack(ctx, "参数格式错误")
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result, err := repository.DelSysTenant(db, id)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

// GetSysTenantById godoc
//
//	@Summary		根据ID获取租户
//	@Tags			租户
//	@Accept			json
//	@Produce		json
//	@Param			id path		int	true	"租户ID"
//	@Success		200	{object}		pojo.SysTenantBack
//	@Router			/api/v1/admin/tenant/:id [get]
func GetSysTenantById(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.ErrorBack(ctx, "参数格式错误")
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result, err := repository.GetSysTenantById(db, id)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

// GetAppTenantServiceLinks godoc
//
//	@Summary		App端获取当前域名绑定客服链接
//	@Tags			租户
//	@Produce		json
//	@Success		200	{object}		pojo.SysTenantServiceLinksBack
//	@Router			/api/v1/app/domain/serviceLinks [get]
func GetAppTenantServiceLinks(ctx *gin.Context) {
	db := ctx.MustGet("db").(*gorm.DB)
	// 已登录：用其所属商户配置的客服；未登录：按域名匹配商户客服
	tenantID := appServiceLinksLoginTenantID(ctx)
	result, err := repository.GetCurrentTenantServiceLinks(db, tenantID, getAppServiceLinksHost(ctx))
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

// appServiceLinksLoginTenantID 可选解析 App token（该接口本身不校验登录）：
// 已登录且 token 有效返回其商户ID；未登录或 token 无效返回 0（回退按域名查）。
func appServiceLinksLoginTenantID(ctx *gin.Context) int64 {
	authHeader := strings.TrimSpace(ctx.GetHeader("Authorization"))
	if authHeader == "" {
		return 0
	}
	authHeader = strings.TrimPrefix(authHeader, "Bearer ")
	hostInfo := utils.GetTempHostInfo(utils.GetRequestHost(ctx))
	userID, hostName, tenantID, parseErr := utils.ParseAppToken(hostInfo.AccessSecret, authHeader)
	if parseErr != nil || userID == 0 || hostInfo.HostName != hostName {
		return 0
	}
	return tenantID
}

func getAppServiceLinksHost(ctx *gin.Context) string {
	for _, rawHost := range []string{
		ctx.GetHeader("Origin"),
		ctx.GetHeader("Referer"),
		ctx.Request.Host,
	} {
		if host := parseHeaderHost(rawHost); host != "" {
			return host
		}
	}

	return ""
}

func parseHeaderHost(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Host == "" {
		parsed, err = url.Parse("//" + raw)
	}
	if err != nil {
		return ""
	}
	return parsed.Hostname()
}

// ResetSysTenantPassword godoc
//
//	@Summary		重置租户密码
//	@Tags			租户
//	@Accept			json
//	@Produce		json
//	@Param			data body		pojo.SysTenantResetPassword	true	"租户重置密码参数"
//	@Success		200	{object}		string
//	@Router			/api/v1/admin/tenant/resetPassword [post]
func ResetSysTenantPassword(ctx *gin.Context) {
	var req pojo.SysTenantResetPassword
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result, err := repository.ResetSysTenantPassword(db, req)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}
