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
//	@Success		200	{object}		pojo.TgUserResp
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
//	@Success		200	{object}		pojo.TgUserBack
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
		utils.ErrorBack(ctx, "参数格式错误")
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
//	@Success		200	{object}		pojo.TgUserBack
//	@Router			/api/v1/admin/tgUser/:id [get]
func GetTgUserById(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.ErrorBack(ctx, "参数格式错误")
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
//	@Success		200	{object}		pojo.TgUserBack
//	@Router			/api/v1/admin/tgUser/status [post]
func SetTgUserStatus(ctx *gin.Context) {
	var req pojo.TgUserStatusSet
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	if req.ID <= 0 {
		utils.ErrorBack(ctx, "参数格式错误")
		return
	}
	if req.Status != 1 && req.Status != 0 && req.Status != -1 {
		utils.ErrorBack(ctx, "状态不支持")
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

// GetTgUsersWithSubStats 列表返回所有下级（不限层级）的充值金额、流水、提现金额
func GetTgUsersWithSubStats(ctx *gin.Context) {
	var search pojo.TgUserSearch
	search.SetPageDefaults()
	if err := ctx.ShouldBindJSON(&search); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	result := tenantRepo.GetTgUsersWithSubStats(getDB(ctx), search)
	utils.SuccessObjBack(ctx, result)
}

// GetTgUsersWithSubStatsSummary 展示所有不限层级下级的充值金额之和、流水之和、提现金额之和
func GetTgUsersWithSubStatsSummary(ctx *gin.Context) {
	var search pojo.TgUserSearch
	if err := ctx.ShouldBindJSON(&search); err != nil && err != io.EOF {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	result := tenantRepo.GetTgUsersWithSubStatsSummary(getDB(ctx), search.ParentID)
	utils.SuccessObjBack(ctx, result)
}

// SendTgEmailCode 发送邮箱验证码（按IP每分钟限流）
func SendTgEmailCode(ctx *gin.Context) {
	var req pojo.TgSendEmailCodeReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ErrorBack(ctx, "参数格式错误")
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

// RegisterTgByEmail 邮箱注册
func RegisterTgByEmail(ctx *gin.Context) {
	var req pojo.TgEmailRegisterReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ErrorBack(ctx, "参数格式错误")
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	newUser, err := repository.RegisterTgByEmail(db, req.Email, req.Password, req.Code)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}

	utils.SuccessObjBack(ctx, gin.H{
		"id":    newUser.ID,
		"uid":   newUser.Uid,
		"email": newUser.Email,
	})
}

// LoginTgByEmail 邮箱登录
func LoginTgByEmail(ctx *gin.Context) {
	var req pojo.TgEmailLoginReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ErrorBack(ctx, "参数格式错误")
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

// ForgotPasswordByEmail 忘记密码
func ForgotPasswordByEmail(ctx *gin.Context) {
	var req pojo.TgForgotPasswordReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ErrorBack(ctx, "参数格式错误")
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	if err := repository.ResetTgPasswordByEmail(db, req.Email, req.Code, req.NewPassword, utils.GetIPAddress(ctx)); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessBack(ctx, "success")
}

// BindCurrentTgEmail 绑定当前TG用户邮箱
func BindCurrentTgEmail(ctx *gin.Context) {
	userIDRaw, ok := ctx.Get("userId")
	if !ok {
		utils.UnauthorizedBack(ctx, "token is invalid")
		return
	}
	userID, ok := userIDRaw.(int64)
	if !ok || userID <= 0 {
		utils.UnauthorizedBack(ctx, "token is invalid")
		return
	}

	var req pojo.TgBindEmailReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ErrorBack(ctx, "参数格式错误")
		return
	}
	email := strings.TrimSpace(strings.ToLower(req.Email))
	code := strings.TrimSpace(req.Code)
	if !utils.IsEmail(email) {
		utils.ErrorBack(ctx, "邮箱格式错误")
		return
	}
	if len(code) != 6 {
		utils.ErrorBack(ctx, "验证码格式错误")
		return
	}

	codeKey := fmt.Sprintf("bgu_tg_email_code_%s", email)
	cacheCode, err := utils.RD.Get(context.Background(), codeKey).Result()
	if err != nil || strings.TrimSpace(cacheCode) == "" {
		utils.ErrorBack(ctx, "验证码已失效")
		return
	}
	if strings.TrimSpace(cacheCode) != code {
		utils.ErrorBack(ctx, "验证码错误")
		return
	}

	db := ctx.MustGet("db").(*gorm.DB)
	var exists pojo.TgUser
	if err = db.Where("email = ? AND id <> ? AND status <> ?", email, userID, -1).First(&exists).Error; err == nil && exists.ID > 0 {
		utils.ErrorBack(ctx, "邮箱已被使用")
		return
	}
	if err = db.Model(&pojo.TgUser{}).Where("id = ?", userID).Update("email", email).Error; err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	_ = utils.RD.Del(context.Background(), codeKey).Err()
	utils.SuccessObjBack(ctx, gin.H{
		"email": email,
	})
}

// GetCurrentTgUserInfo 获取当前TG用户信息
func GetCurrentTgUserInfo(ctx *gin.Context) {
	authHeader := strings.TrimSpace(ctx.GetHeader("Authorization"))
	if authHeader == "" {
		utils.UnauthorizedBack(ctx, "Authorization header is missing")
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

// UpdateCurrentTgUserAvatar 更新当前TG用户头像
func UpdateCurrentTgUserAvatar(ctx *gin.Context) {
	userIDRaw, ok := ctx.Get("userId")
	if !ok {
		utils.UnauthorizedBack(ctx, "token is invalid")
		return
	}
	userID, ok := userIDRaw.(int64)
	if !ok || userID <= 0 {
		utils.UnauthorizedBack(ctx, "token is invalid")
		return
	}

	var req pojo.TgUpdateAvatarReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	req.Avatar = strings.TrimSpace(req.Avatar)
	if req.Avatar == "" {
		utils.ErrorBack(ctx, "avatar is required")
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

// GetCurrentTgInviteStats 获取当前TG用户邀请统计
func GetCurrentTgInviteStats(ctx *gin.Context) {
	userIDRaw, ok := ctx.Get("userId")
	if !ok {
		utils.UnauthorizedBack(ctx, "token is invalid")
		return
	}
	userID, ok := userIDRaw.(int64)
	if !ok || userID <= 0 {
		utils.UnauthorizedBack(ctx, "token is invalid")
		return
	}

	db := ctx.MustGet("db").(*gorm.DB)

	var user pojo.TgUser
	if err := db.Where("id = ?", userID).First(&user).Error; err != nil || user.ID == 0 {
		utils.ErrorBack(ctx, "用户不存在")
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
		InviteCode:       inviteCode,
		InviteCount:      inviteCount,
		TodayInviteCount: todayInviteCount,
		RechargeUsers:    rechargeUsers,
		TodayRechargeUsers: todayRechargeUsers,
		TotalCommission:  user.RebateTotalAmount,
		AvailableCommission: user.RebateAmount,
		TodayCommission:  todayCommission,
	})
}

// TransferRebateToBalance 将当前用户 rebate_amount 转移到 balance，并记录 cash_history
func TransferRebateToBalance(ctx *gin.Context) {
	userIDRaw, ok := ctx.Get("userId")
	if !ok {
		utils.UnauthorizedBack(ctx, "token is invalid")
		return
	}
	userID, ok := userIDRaw.(int64)
	if !ok || userID <= 0 {
		utils.UnauthorizedBack(ctx, "token is invalid")
		return
	}

	db := ctx.MustGet("db").(*gorm.DB)
	var transferAmount float64
	var newBalance float64

	err := db.Transaction(func(tx *gorm.DB) error {
		var user pojo.TgUser
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", userID).First(&user).Error; err != nil {
			return err
		}
		if user.Status != 1 {
			return fmt.Errorf("用户已禁用，请联系管理员处理")
		}
		if user.RebateAmount <= 0 {
			return fmt.Errorf("暂无可转移佣金")
		}

		transferAmount = user.RebateAmount
		newBalance = user.Balance + transferAmount

		if err := tx.Model(&pojo.TgUser{}).
			Where("id = ?", user.ID).
			Updates(map[string]any{
				"balance":       gorm.Expr("balance + ?", transferAmount),
				"rebate_amount": 0,
			}).Error; err != nil {
			return err
		}

		cashHistory := pojo.CashHistory{
			UserId:      user.ID,
			AwardUni:    fmt.Sprintf("rebate_transfer_%d_%d", user.ID, time.Now().UnixNano()),
			Amount:      transferAmount,
			StartAmount: user.Balance,
			EndAmount:   newBalance,
			CashMark:    "佣金转余额",
			CashDesc:    fmt.Sprintf("返佣余额转入主余额%.3f", transferAmount),
			Type:        pojo.CashHistoryTypeRebateTransfer,
			IsGift:      0,
			FromUserId:  0,
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
		"rebateAmount":   0,
	})
}

// GetCurrentTgCashHistory 当前TG用户流水列表（分页，排除抽成）
func GetCurrentTgCashHistory(ctx *gin.Context) {
	userIDRaw, ok := ctx.Get("userId")
	if !ok {
		utils.UnauthorizedBack(ctx, "token is invalid")
		return
	}
	userID, ok := userIDRaw.(int64)
	if !ok || userID <= 0 {
		utils.UnauthorizedBack(ctx, "token is invalid")
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
			utils.UnauthorizedBack(ctx, "Authorization header is missing")
			return
		}
		token = strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
	}
	if token == "" {
		utils.UnauthorizedBack(ctx, "token is invalid")
		return
	}
	key := utils.KeyRdTgOnline + utils.MD5(token)
	_ = utils.RD.Del(context.Background(), key).Err()
	utils.SuccessBack(ctx, "success")
}
