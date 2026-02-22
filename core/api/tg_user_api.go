package api

import (
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/repository"
	"BaseGoUni/core/utils"
	tenantRepo "BaseGoUni/tenant/repository"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"io"
	"strconv"
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
