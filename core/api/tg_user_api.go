package api

import (
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/repository"
	"BaseGoUni/core/utils"
	tenantRepo "BaseGoUni/tenant/repository"
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"io"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// GetTgUsers godoc
//
//	@Summary		获取Telegram用户列表
//	@Tags			Telegram用户
//	@Accept			json
//	@Produce		json
//	@Param			data body		pojo.TgUserSearch	true	"查询条件"
//	@Success		200	{object}		pojo.TgUserAdminResp
//	@Router			/api/v1/admin/tgUser/list [post]
func GetTgUsers(ctx *gin.Context) {
	var search pojo.TgUserSearch
	search.SetPageDefaults()
	if err := ctx.ShouldBindJSON(&search); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result := repository.GetTgUsers(db, search)
	utils.SuccessObjBack(ctx, result)
}

// SetTgUser godoc
//
//	@Summary		创建或更新Telegram用户
//	@Tags			Telegram用户
//	@Accept			json
//	@Produce		json
//	@Param			data body		pojo.TgUserSet	true	"Telegram用户信息"
//	@Success		200	{object}		pojo.TgUserAdminBack
//	@Router			/api/v1/admin/tgUser [post]
func SetTgUser(ctx *gin.Context) {
	var req pojo.TgUserSet
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result, err := repository.SetTgUser(db, req)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

// DelTgUser godoc
//
//	@Summary		删除Telegram用户
//	@Tags			Telegram用户
//	@Accept			json
//	@Produce		json
//	@Param			id path		int	true	"Telegram用户ID"
//	@Success		200	{object}		string
//	@Router			/api/v1/admin/tgUser/:id [delete]
func DelTgUser(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.ErrorBack(ctx, "invalid_params")
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result, err := repository.DelTgUser(db, id)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

// GetTgUserById godoc
//
//	@Summary		根据ID获取Telegram用户
//	@Tags			Telegram用户
//	@Accept			json
//	@Produce		json
//	@Param			id path		int	true	"Telegram用户ID"
//	@Success		200	{object}		pojo.TgUserAdminBack
//	@Router			/api/v1/admin/tgUser/:id [get]
func GetTgUserById(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.ErrorBack(ctx, "invalid_params")
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result, err := repository.GetTgUserById(db, id)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

// SetTgUserStatus godoc
//
//	@Summary		封禁/解封Telegram用户
//	@Tags			Telegram用户
//	@Accept			json
//	@Produce		json
//	@Param			data body		pojo.TgUserStatusSet	true	"状态信息"
//	@Success		200	{object}		pojo.TgUserAdminBack
//	@Router			/api/v1/admin/tgUser/status [post]
func SetTgUserStatus(ctx *gin.Context) {
	var req pojo.TgUserStatusSet
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	if req.ID <= 0 {
		utils.ErrorBack(ctx, "invalid_params")
		return
	}
	if req.Status != 1 && req.Status != 0 && req.Status != -1 {
		utils.ErrorBack(ctx, "invalid_status")
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result, err := repository.SetTgUserStatus(db, req.ID, req.Status)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

// SetTgUserRebateRate godoc
//
//	@Summary		修改Telegram用户返佣比例
//	@Tags			Telegram用户
//	@Accept			json
//	@Produce		json
//	@Param			data body		pojo.TgUserRebateRateSet	true	"返佣比例信息"
//	@Success		200	{object}		pojo.TgUserAdminBack
//	@Router			/api/v1/admin/tgUser/rebateRate [post]
func SetTgUserRebateRate(ctx *gin.Context) {
	var req pojo.TgUserRebateRateSet
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	if req.ID <= 0 || req.RebateRate < 0 || req.RebateRate > 100 {
		utils.ErrorBack(ctx, "invalid_params")
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result, err := repository.SetTgUserRebateRate(db, req.ID, req.RebateRate)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

// AddTgUserRebateAmount godoc
//
//	@Summary		后台给Telegram用户增加佣金金额
//	@Tags			Telegram用户
//	@Accept			json
//	@Produce		json
//	@Param			data body		pojo.TgUserRebateAmountAdd	true	"加佣金金额"
//	@Success		200	{object}		pojo.TgUserAdminBack
//	@Router			/api/v1/admin/tgUser/rebateAmount [post]
func AddTgUserRebateAmount(ctx *gin.Context) {
	var req pojo.TgUserRebateAmountAdd
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	req.Amount = utils.Truncate2(req.Amount)
	if req.ID <= 0 || req.Amount <= 0 {
		utils.ErrorBack(ctx, "invalid_params")
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result, err := repository.AddTgUserRebateAmount(db, req.ID, req.Amount)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

// SetTgUserRemark godoc
//
//	@Summary		修改Telegram用户备注
//	@Tags			Telegram用户
//	@Accept			json
//	@Produce		json
//	@Param			data body		pojo.TgUserRemarkSet	true	"备注信息"
//	@Success		200	{object}		pojo.TgUserAdminBack
//	@Router			/api/v1/admin/tgUser/remark [post]
func SetTgUserRemark(ctx *gin.Context) {
	var req pojo.TgUserRemarkSet
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	if req.ID <= 0 || len([]rune(req.Remark)) > 255 {
		utils.ErrorBack(ctx, "invalid_params")
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result, err := repository.SetTgUserRemark(db, req.ID, req.Remark)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

// BatchCreateBotTgUsers godoc
//
//	@Summary		批量创建机器人 Telegram 用户
//	@Tags			Telegram用户
//	@Accept			json
//	@Produce		json
//	@Param			data body		pojo.TgUserBatchCreateBotReq	true	"批量创建机器人参数"
//	@Success		200	{object}		pojo.TgUserBatchCreateBotResp
//	@Router			/api/v1/admin/tgUser/batchCreateBot [post]
func BatchCreateBotTgUsers(ctx *gin.Context) {
	var req pojo.TgUserBatchCreateBotReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}

	db := ctx.MustGet("db").(*gorm.DB)
	result, err := repository.BatchCreateBotTgUsers(db, req)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

// BatchUpdateBotTgUsers godoc
//
//	@Summary		批量修改机器人 Telegram 用户
//	@Tags			Telegram用户
//	@Accept			json
//	@Produce		json
//	@Param			data body		pojo.TgUserBatchUpdateBotReq	true	"批量修改机器人参数"
//	@Success		200	{object}		pojo.TgUserBatchUpdateBotResp
//	@Router			/api/v1/admin/tgUser/batchUpdateBot [post]
func BatchUpdateBotTgUsers(ctx *gin.Context) {
	var req pojo.TgUserBatchUpdateBotReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}

	db := ctx.MustGet("db").(*gorm.DB)
	result, err := repository.BatchUpdateBotTgUsers(db, req)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

// GetTgUsersWithSubStats 列表返回所有下级（不限层级）的充值金额、流水、提现金额
func GetTgUsersWithSubStats(ctx *gin.Context) {
	var search pojo.TgUserSearch
	search.SetPageDefaults()
	if err := ctx.ShouldBindJSON(&search); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	result := tenantRepo.GetTgUsersWithSubStats(getDB(ctx), 0, search)
	utils.SuccessObjBack(ctx, result)
}

// GetTgUsersWithSubStatsSummary 展示所有不限层级下级的充值金额之和、流水之和、提现金额之和
func GetTgUsersWithSubStatsSummary(ctx *gin.Context) {
	var search pojo.TgUserSearch
	if err := ctx.ShouldBindJSON(&search); err != nil && err != io.EOF {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	result := tenantRepo.GetTgUsersWithSubStatsSummary(getDB(ctx), 0, search)
	utils.SuccessObjBack(ctx, result)
}

// SendTgEmailCode 发送邮箱验证码（按IP每分钟限流）
func SendTgEmailCode(ctx *gin.Context) {
	var req pojo.TgSendEmailCodeReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ErrorBack(ctx, "invalid_params")
		return
	}
	code, err := repository.SendTgEmailCode(req.Email, utils.GetIPAddress(ctx), utils.IsDev())
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	if !utils.IsDev() {
		utils.SuccessBack(ctx, "success")
		return
	}

	utils.SuccessObjBack(ctx, gin.H{
		"code": code,
	})
}

// SendTgSMSCode 发送短信验证码（按IP每分钟限流）
func SendTgSMSCode(ctx *gin.Context) {
	var req pojo.TgSendSMSCodeReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ErrorBack(ctx, "invalid_params")
		return
	}
	code, err := repository.SendTgSMSCode(req.Phone, req.Country, utils.GetIPAddress(ctx), utils.IsDev())
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	if !utils.IsDev() {
		utils.SuccessBack(ctx, "success")
		return
	}

	utils.SuccessObjBack(ctx, gin.H{
		"code":    code,
		"phone":   strings.TrimSpace(req.Phone),
		"country": strings.TrimSpace(strings.ToUpper(req.Country)),
	})
}

// RegisterTgByEmail 邮箱注册
func RegisterTgByEmail(ctx *gin.Context) {
	var req pojo.TgEmailRegisterReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ErrorBack(ctx, "invalid_params")
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	sourceChannelCode := repository.FirstSourceChannelCode(req.SourceChannelCode, req.ChannelCode)
	tenantID := resolveTenantIDByRegisterReferrer(db, registerReferrer(ctx, req.Referrer))
	newUser, err := repository.RegisterTgByEmail(
		db,
		req.Email,
		req.FirstName,
		req.Password,
		req.Code,
		sourceChannelCode,
		tenantID,
		req.InviteCode,
		utils.GetIPAddress(ctx),
		utils.GetIPCountry(ctx),
	)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}

	utils.SuccessObjBack(ctx, gin.H{
		"id":        newUser.ID,
		"uid":       newUser.Uid,
		"email":     newUser.Email,
		"firstName": newUser.FirstName,
	})
}

// RegisterTgByPhone 手机号注册
func RegisterTgByPhone(ctx *gin.Context) {
	var req pojo.TgPhoneRegisterReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ErrorBack(ctx, "invalid_params")
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	sourceChannelCode := repository.FirstSourceChannelCode(req.SourceChannelCode, req.ChannelCode)
	tenantID := resolveTenantIDByRegisterReferrer(db, registerReferrer(ctx, req.Referrer))
	inviteCode := req.InviteCode
	if strings.TrimSpace(inviteCode) == "" {
		inviteCode = req.Code
	}
	newUser, err := repository.RegisterTgByPhone(
		db,
		req.Phone,
		req.Country,
		req.FirstName,
		req.Password,
		sourceChannelCode,
		tenantID,
		inviteCode,
		utils.GetIPAddress(ctx),
		utils.GetIPCountry(ctx),
	)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}

	utils.SuccessObjBack(ctx, gin.H{
		"id":        newUser.ID,
		"uid":       newUser.Uid,
		"phone":     newUser.Phone,
		"country":   newUser.Country,
		"firstName": newUser.FirstName,
	})
}

func registerReferrer(ctx *gin.Context, bodyReferrer string) string {
	if referrer := strings.TrimSpace(bodyReferrer); referrer != "" {
		return referrer
	}
	if referrer := strings.TrimSpace(ctx.GetHeader("Referer")); referrer != "" {
		return referrer
	}
	return strings.TrimSpace(ctx.GetHeader("Referrer"))
}

func resolveTenantIDByRegisterReferrer(db *gorm.DB, referrer string) int64 {
	host := normalizeRegisterReferrerHost(referrer)
	if host == "" {
		return 0
	}

	candidates := []string{host}
	if strings.HasPrefix(host, "www.") {
		candidates = append(candidates, strings.TrimPrefix(host, "www."))
	}
	parts := strings.Split(host, ".")
	for i := 1; i < len(parts)-1; i++ {
		candidates = append(candidates, "*."+strings.Join(parts[i:], "."))
	}

	for _, candidate := range candidates {
		var tenant pojo.SysTenant
		err := db.Model(&pojo.SysTenant{}).
			Where("status = ? AND bind_domain = ?", 1, candidate).
			First(&tenant).Error
		if err == nil && tenant.ID > 0 {
			return tenant.ID
		}
	}
	return 0
}

func normalizeRegisterReferrerHost(referrer string) string {
	referrer = strings.TrimSpace(strings.ToLower(referrer))
	if referrer == "" {
		return ""
	}
	if !strings.Contains(referrer, "://") {
		referrer = "https://" + referrer
	}
	parsed, err := url.Parse(referrer)
	if err != nil {
		return ""
	}
	host := strings.TrimSpace(parsed.Hostname())
	return strings.TrimSuffix(host, ".")
}

// LoginTgByEmail 邮箱登录
func LoginTgByEmail(ctx *gin.Context) {
	var req pojo.TgEmailLoginReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ErrorBack(ctx, "invalid_params")
		return
	}
	tempHostInfo := ctx.MustGet("hostInfo").(pojo.HostInfo)
	db := ctx.MustGet("db").(*gorm.DB)
	onlineUser := pojo.OnlineUser{
		Username:  req.Email,
		Browser:   ctx.GetHeader("User-Agent"),
		Ip:        utils.GetIPAddress(ctx),
		LoginTime: time.Now(),
	}
	data, err := repository.TgEmailLogin(db, tempHostInfo, req, onlineUser)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, data)
}

// LoginTgByPhone 手机号登录
func LoginTgByPhone(ctx *gin.Context) {
	var req pojo.TgPhoneLoginReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ErrorBack(ctx, "invalid_params")
		return
	}
	tempHostInfo := ctx.MustGet("hostInfo").(pojo.HostInfo)
	db := ctx.MustGet("db").(*gorm.DB)
	onlineUser := pojo.OnlineUser{
		Username:  req.Phone,
		Browser:   ctx.GetHeader("User-Agent"),
		Ip:        utils.GetIPAddress(ctx),
		LoginTime: time.Now(),
	}
	data, err := repository.TgPhoneLogin(db, tempHostInfo, req, onlineUser)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, data)
}

// ForgotPasswordByEmail 忘记密码
func ForgotPasswordByEmail(ctx *gin.Context) {
	var req pojo.TgForgotPasswordReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ErrorBack(ctx, "invalid_params")
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	if err := repository.ResetTgPasswordByEmail(db, req.Email, req.Code, req.NewPassword, utils.GetIPAddress(ctx)); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessBack(ctx, "success")
}

// ForgotPasswordByPhone 手机号忘记密码
func ForgotPasswordByPhone(ctx *gin.Context) {
	var req pojo.TgForgotPasswordByPhoneReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ErrorBack(ctx, "invalid_params")
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	if err := repository.ResetTgPasswordByPhone(db, req.Phone, req.Country, req.Code, req.NewPassword, utils.GetIPAddress(ctx)); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessBack(ctx, "success")
}

// BindCurrentTgEmail 绑定当前TG用户邮箱
func BindCurrentTgEmail(ctx *gin.Context) {
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

	var req pojo.TgBindEmailReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ErrorBack(ctx, "invalid_params")
		return
	}
	email := strings.TrimSpace(strings.ToLower(req.Email))
	if !utils.IsEmail(email) {
		utils.ErrorBack(ctx, "email_format_error")
		return
	}

	db := ctx.MustGet("db").(*gorm.DB)
	var exists pojo.TgUser
	if err := db.Where("email = ? AND id <> ? AND status <> ?", email, userID, -1).First(&exists).Error; err == nil && exists.ID > 0 {
		utils.ErrorBack(ctx, "email_already_used")
		return
	}
	if err := db.Model(&pojo.TgUser{}).Where("id = ?", userID).Update("email", email).Error; err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, gin.H{
		"email": email,
	})
}

// BindCurrentTgPhone 绑定或换绑当前TG用户手机号
func BindCurrentTgPhone(ctx *gin.Context) {
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

	var req pojo.TgBindPhoneReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ErrorBack(ctx, "invalid_params")
		return
	}

	db := ctx.MustGet("db").(*gorm.DB)
	if err := repository.BindCurrentTgPhone(db, userID, req.Phone, req.Country, req.Code); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	phone := strings.TrimSpace(req.Phone)
	country := utils.InferCountryByPhone(phone, strings.TrimSpace(strings.ToUpper(req.Country)))
	utils.SuccessObjBack(ctx, gin.H{
		"phone":   phone,
		"country": country,
	})
}

// BindCurrentTgChannelName 绑定已加入指定频道的 Telegram @用户名。
func BindCurrentTgChannelName(ctx *gin.Context) {
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

	var req pojo.TgBindChannelNameReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ErrorBack(ctx, "invalid_params")
		return
	}

	db := ctx.MustGet("db").(*gorm.DB)
	hostInfo := ctx.MustGet("hostInfo").(pojo.HostInfo)
	data, err := repository.BindCurrentTgChannelName(ctx.Request.Context(), db, hostInfo.TablePrefix, userID, req.TgName)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, data)
}

// GetCurrentTgUserInfo 获取当前TG用户信息
func GetCurrentTgUserInfo(ctx *gin.Context) {
	authHeader := strings.TrimSpace(ctx.GetHeader("Authorization"))
	if authHeader == "" {
		utils.UnauthorizedBack(ctx, "auth_header_missing")
		return
	}
	token := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
	tempHostInfo := ctx.MustGet("hostInfo").(pojo.HostInfo)
	db := ctx.MustGet("db").(*gorm.DB)
	data, err := repository.GetCurrentTgUserInfo(db, tempHostInfo.AccessSecret, token)
	if err != nil {
		utils.UnauthorizedBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, data)
}

// GetCurrentTgWithdrawSummary 获取当前TG用户提现可用摘要
func GetCurrentTgWithdrawSummary(ctx *gin.Context) {
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

	db := ctx.MustGet("db").(*gorm.DB)
	data, err := repository.GetUserWithdrawSummary(db, userID)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, data)
}

// GetCurrentTgWithdrawActivityFlow 获取当前TG用户提现活动和流水进度
func GetCurrentTgWithdrawActivityFlow(ctx *gin.Context) {
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

	db := ctx.MustGet("db").(*gorm.DB)
	data, err := repository.GetUserWithdrawActivityFlow(db, userID)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, data)
}

// GetTgUserWithdrawActivityFlowById 管理端查询指定TG用户提现活动和流水进度
func GetTgUserWithdrawActivityFlowById(ctx *gin.Context) {
	idStr := ctx.Param("id")
	userID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || userID <= 0 {
		utils.ErrorBack(ctx, "invalid_params")
		return
	}

	db := ctx.MustGet("db").(*gorm.DB)
	data, err := repository.GetUserWithdrawActivityFlow(db, userID)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, data)
}

// UpdateCurrentTgUserAvatar 更新当前TG用户头像
func UpdateCurrentTgUserAvatar(ctx *gin.Context) {
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

	var req pojo.TgUpdateAvatarReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	req.Avatar = strings.TrimSpace(req.Avatar)
	if req.Avatar == "" {
		utils.ErrorBack(ctx, "avatar_required")
		return
	}

	db := ctx.MustGet("db").(*gorm.DB)
	if err := db.Model(&pojo.TgUser{}).Where("id = ?", userID).Update("avatar", req.Avatar).Error; err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, gin.H{
		"avatar": req.Avatar,
	})
}

// UpdateCurrentTgUserName 更新当前TG用户用户名
func UpdateCurrentTgUserName(ctx *gin.Context) {
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

	var req pojo.TgUpdateNameReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	username := strings.TrimSpace(req.Username)
	if username == "" {
		utils.ErrorBack(ctx, "username_required")
		return
	}
	if len([]rune(username)) > 64 {
		utils.ErrorBack(ctx, "username_too_long")
		return
	}

	db := ctx.MustGet("db").(*gorm.DB)
	result := db.Model(&pojo.TgUser{}).
		Where("id = ? AND status = ?", userID, 1).
		Update("username", username)
	if result.Error != nil {
		utils.ErrorBack(ctx, result.Error.Error())
		return
	}
	if result.RowsAffected == 0 {
		utils.ErrorBack(ctx, "user_not_found")
		return
	}
	utils.SuccessObjBack(ctx, gin.H{
		"username": username,
	})
}

// GetCurrentTgInviteStats 获取当前TG用户邀请统计
func GetCurrentTgInviteStats(ctx *gin.Context) {
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

	db := ctx.MustGet("db").(*gorm.DB)

	var user pojo.TgUser
	if err := db.Where("id = ?", userID).First(&user).Error; err != nil || user.ID == 0 {
		utils.ErrorBack(ctx, "user_not_found")
		return
	}

	var inviteCount int64
	_ = db.Model(&pojo.TgUser{}).
		Where("parent_id = ? AND status <> ?", userID, -1).
		Count(&inviteCount).Error

	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)
	var todayInviteCount int64
	_ = db.Model(&pojo.TgUser{}).
		Where("parent_id = ? AND status <> ? AND created_at >= ? AND created_at < ?", userID, -1, startOfDay, endOfDay).
		Count(&todayInviteCount).Error

	var rechargeUsers int64
	_ = db.Model(&pojo.TgUser{}).
		Where("parent_id = ? AND status <> ? AND recharge_amount > 0", userID, -1).
		Count(&rechargeUsers).Error

	var todayRechargeUsers int64
	_ = db.Table("recharge_order ro").
		Joins("inner join tg_user tu on tu.id = ro.user_id").
		Where("tu.parent_id = ? AND tu.status <> ? AND ro.status = ? AND ro.pay_time >= ? AND ro.pay_time < ?", userID, -1, 1, startOfDay, endOfDay).
		Select("COUNT(DISTINCT ro.user_id)").
		Scan(&todayRechargeUsers).Error

	var todayCommission float64
	_ = db.Model(&pojo.TgUserRebateRecord{}).
		Where("parent_user_id = ? AND status = ? AND created_at >= ? AND created_at < ?", userID, 1, startOfDay, endOfDay).
		Select("COALESCE(SUM(rebate_amount), 0)").
		Scan(&todayCommission).Error

	inviteCode := ""
	if user.InviteCode != nil {
		inviteCode = strings.TrimSpace(*user.InviteCode)
	}

	utils.SuccessObjBack(ctx, pojo.TgInviteStatsBack{
		InviteCode:          inviteCode,
		InviteCount:         inviteCount,
		TodayInviteCount:    todayInviteCount,
		RechargeUsers:       rechargeUsers,
		TodayRechargeUsers:  todayRechargeUsers,
		TotalCommission:     utils.Truncate2(user.RebateTotalAmount),
		AvailableCommission: utils.Truncate2(user.RebateAmount),
		TodayCommission:     utils.Truncate2(todayCommission),
	})
}

// GetCurrentTgInviteRuleConfig 获取邀请规则配置
func GetCurrentTgInviteRuleConfig(ctx *gin.Context) {
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

	hostInfo := ctx.MustGet("hostInfo").(pojo.HostInfo)
	tablePrefix := hostInfo.TablePrefix
	db := ctx.MustGet("db").(*gorm.DB)

	parseConfigFloat := func(key string, defaultValue string) float64 {
		val := utils.GetStringCache(tablePrefix, key, &defaultValue)
		if val == nil || strings.TrimSpace(*val) == "" {
			f, _ := strconv.ParseFloat(defaultValue, 64)
			return f
		}
		f, err := strconv.ParseFloat(strings.TrimSpace(*val), 64)
		if err != nil {
			f, _ = strconv.ParseFloat(defaultValue, 64)
			return f
		}
		return f
	}

	parseSendMinMax := func(defaultValue string) (float64, float64) {
		val := utils.GetStringCache(tablePrefix, "send_min_max", &defaultValue)
		raw := defaultValue
		if val != nil && strings.TrimSpace(*val) != "" {
			raw = strings.TrimSpace(*val)
		}

		parts := strings.Split(raw, "|")
		if len(parts) != 2 {
			return 10, 5000
		}

		minValue, errMin := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
		maxValue, errMax := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
		if errMin != nil || errMax != nil || minValue <= 0 || maxValue <= 0 || minValue > maxValue {
			return 10, 5000
		}
		return utils.Truncate2(minValue), utils.Truncate2(maxValue)
	}

	sendMinAmount, sendMaxAmount := parseSendMinMax("10|5000")
	inviteLuckyRebateRate := parseConfigFloat("invite_lucky_rebate_rate", "40")
	var user pojo.TgUser
	if err := db.Select("id", "rebate_rate").Where("id = ?", userID).First(&user).Error; err == nil && user.RebateRate > 0 {
		inviteLuckyRebateRate = user.RebateRate
	}

	utils.SuccessObjBack(ctx, pojo.TgInviteRuleConfigBack{
		LuckySendCommission:       parseConfigFloat("lucky_send_commission", "5"),
		LuckyGrabbingCommission:   parseConfigFloat("lucky_grabbing_commission", "5"),
		InviteFirstRechargeReward: parseConfigFloat("invite_first_recharge_reward", "10"),
		InviteLuckyRebateRate:     inviteLuckyRebateRate,
		InviteThunderRebateRate:   parseConfigFloat("invite_thunder_rebate_rate", "40"),
		SendMinAmount:             sendMinAmount,
		SendMaxAmount:             sendMaxAmount,
	})
}

// TransferRebateToBalance 将当前用户 rebate_amount 转移到 balance，并记录 cash_history
func TransferRebateToBalance(ctx *gin.Context) {
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

	var req pojo.TgRebateTransferReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	req.Amount = utils.Truncate2(req.Amount)
	if req.Amount <= 0 {
		utils.ErrorBack(ctx, "invalid_amount")
		return
	}

	db := ctx.MustGet("db").(*gorm.DB)
	var transferAmount float64
	var newBalance float64
	var newRebateAmount float64

	err := db.Transaction(func(tx *gorm.DB) error {
		var user pojo.TgUser
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", userID).First(&user).Error; err != nil {
			return err
		}
		if user.Status != 1 {
			return fmt.Errorf("user_disabled_contact_admin")
		}
		if user.RebateAmount <= 0 {
			return fmt.Errorf("no_rebate_to_transfer")
		}
		if req.Amount > utils.Truncate2(user.RebateAmount) {
			return fmt.Errorf("rebate_amount_insufficient")
		}

		transferAmount = req.Amount
		newBalance = utils.Truncate2(user.Balance + transferAmount)
		newRebateAmount = utils.Truncate2(user.RebateAmount - transferAmount)

		if err := repository.EnsureUserWithdrawLimitState(tx, user); err != nil {
			return err
		}
		if err := tx.Model(&pojo.TgUser{}).
			Where("id = ?", user.ID).
			Updates(map[string]any{
				"balance":       gorm.Expr("balance + ?", transferAmount),
				"rebate_amount": gorm.Expr("rebate_amount - ?", transferAmount),
			}).Error; err != nil {
			return err
		}

		cashHistory := pojo.CashHistory{
			UserId:          user.ID,
			AwardUni:        fmt.Sprintf("rebate_transfer_%d_%d", user.ID, time.Now().UnixNano()),
			Amount:          transferAmount,
			StartAmount:     user.Balance,
			EndAmount:       newBalance,
			CashMark:        "佣金转余额",
			CashDesc:        fmt.Sprintf("返佣余额转入主余额%.2f", transferAmount),
			Type:            pojo.CashHistoryTypeRebateTransfer,
			IsGift:          0,
			FromUserId:      0,
			SourceChannelID: user.SourceChannelID,
		}
		return tx.Create(&cashHistory).Error
	})
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}

	utils.SuccessObjBack(ctx, gin.H{
		"transferAmount": transferAmount,
		"balance":        newBalance,
		"rebateAmount":   newRebateAmount,
	})
}

// GetCurrentTgCashHistory 当前TG用户流水列表（分页，排除抽成）
func GetCurrentTgCashHistory(ctx *gin.Context) {
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

	var search pojo.CashHistorySearch
	search.SetPageDefaults()
	if err := ctx.ShouldBindJSON(&search); err != nil && err != io.EOF {
		utils.ErrorBack(ctx, err.Error())
		return
	}

	db := ctx.MustGet("db").(*gorm.DB)
	result := repository.GetCashHistoryListApp(db, userID, search)
	utils.SuccessObjBack(ctx, result)
}

// TgLogout TG用户登出
// SetAudioOpen 设置音效开关
//
//	@Summary		设置音效开关
//	@Tags			TG用户
//	@Accept			json
//	@Produce		json
//	@Param			data body		pojo.TgSetAudioOpenReq	true	"audioOpen 0=关 1=开"
//	@Success		200	{object}		string
//	@Router			/api/v1/app/tg/audioOpen [post]
func SetAudioOpen(ctx *gin.Context) {
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
	var req pojo.TgSetAudioOpenReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	if err := db.Model(&pojo.TgUser{}).Where("id = ?", userID).Update("audio_open", req.AudioOpen).Error; err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessBack(ctx, "success")
}

func TgLogout(ctx *gin.Context) {
	token := ""
	if tokenVal, ok := ctx.Get("token"); ok {
		if v, okCast := tokenVal.(string); okCast {
			token = strings.TrimSpace(v)
		}
	}
	if token == "" {
		authHeader := strings.TrimSpace(ctx.GetHeader("Authorization"))
		if authHeader == "" {
			utils.UnauthorizedBack(ctx, "auth_header_missing")
			return
		}
		token = strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
	}
	if token == "" {
		utils.UnauthorizedBack(ctx, "token_invalid")
		return
	}
	key := utils.KeyRdTgOnline + utils.MD5(token)
	_ = utils.RD.Del(context.Background(), key).Err()
	utils.SuccessBack(ctx, "success")
}
