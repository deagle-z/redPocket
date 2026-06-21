package api

import (
	"BaseGoUni/core/utils"
	tenantRepo "BaseGoUni/tenant/repository"

	"github.com/gin-gonic/gin"
)

type changeTenantPasswordReq struct {
	OldPassword string `json:"oldPassword"`
	NewPassword string `json:"newPassword"`
}

// ChangeTenantPassword 当前租户用户修改密码
func ChangeTenantPassword(ctx *gin.Context) {
	tenantID, ok := getTenantID(ctx)
	if !ok {
		return
	}
	userID := ctx.MustGet("userId").(int64)
	var req changeTenantPasswordReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	if err := tenantRepo.ChangeTenantUserPassword(getDB(ctx), tenantID, userID, req.OldPassword, req.NewPassword); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, "success")
}

// GetTenantTwofaStatus 查询当前用户 Google 验证码绑定状态
func GetTenantTwofaStatus(ctx *gin.Context) {
	tenantID, ok := getTenantID(ctx)
	if !ok {
		return
	}
	userID := ctx.MustGet("userId").(int64)
	bound, err := tenantRepo.GetTenantUserTwofaStatus(getDB(ctx), tenantID, userID)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, gin.H{"bound": bound})
}

// SetupTenantTwofa 生成待绑定的 Google 验证码密钥与 otpauth 链接
func SetupTenantTwofa(ctx *gin.Context) {
	tenantID, ok := getTenantID(ctx)
	if !ok {
		return
	}
	userID := ctx.MustGet("userId").(int64)
	secret, otpauthURL, err := tenantRepo.GenerateTenantUserTwofa(getDB(ctx), tenantID, userID)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, gin.H{"secret": secret, "otpauthUrl": otpauthURL})
}

type tenantTwofaBindReq struct {
	Secret string `json:"secret"`
	Code   string `json:"code"`
}

// BindTenantTwofa 校验动态码并绑定 Google 验证码
func BindTenantTwofa(ctx *gin.Context) {
	tenantID, ok := getTenantID(ctx)
	if !ok {
		return
	}
	userID := ctx.MustGet("userId").(int64)
	var req tenantTwofaBindReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	if err := tenantRepo.BindTenantUserTwofa(getDB(ctx), tenantID, userID, req.Secret, req.Code); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, "success")
}

type tenantTwofaUnbindReq struct {
	Code string `json:"code"`
}

// UnbindTenantTwofa 校验动态码并解绑 Google 验证码
func UnbindTenantTwofa(ctx *gin.Context) {
	tenantID, ok := getTenantID(ctx)
	if !ok {
		return
	}
	userID := ctx.MustGet("userId").(int64)
	var req tenantTwofaUnbindReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	if err := tenantRepo.UnbindTenantUserTwofa(getDB(ctx), tenantID, userID, req.Code); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, "success")
}
