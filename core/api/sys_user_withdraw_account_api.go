package api

import (
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/repository"
	"BaseGoUni/core/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GetSysUserWithdrawAccounts godoc
//
//	@Summary		获取提现账户列表
//	@Tags			提现账户
//	@Accept			json
//	@Produce		json
//	@Param			data body		pojo.SysUserWithdrawAccountSearch	true	"查询条件"
//	@Success		200	{object}	pojo.SysUserWithdrawAccountResp
//	@Router			/api/v1/admin/userWithdrawAccount/list [post]
func GetSysUserWithdrawAccounts(ctx *gin.Context) {
	var search pojo.SysUserWithdrawAccountSearch
	search.SetPageDefaults()
	if err := ctx.ShouldBindJSON(&search); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result := repository.GetSysUserWithdrawAccounts(db, search)
	utils.SuccessObjBack(ctx, result)
}

// GetSysUserWithdrawAccountById godoc
//
//	@Summary		根据ID获取提现账户
//	@Tags			提现账户
//	@Produce		json
//	@Param			id path		int	true	"账户ID"
//	@Success		200	{object}	pojo.SysUserWithdrawAccountBack
//	@Router			/api/v1/admin/userWithdrawAccount/:id [get]
func GetSysUserWithdrawAccountById(ctx *gin.Context) {
	id, ok := parsePathID(ctx)
	if !ok {
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result, err := repository.GetSysUserWithdrawAccountById(db, id)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

// AdminSetSysUserWithdrawAccount godoc
//
//	@Summary		管理员创建或更新提现账户
//	@Tags			提现账户
//	@Accept			json
//	@Produce		json
//	@Param			data body		pojo.SysUserWithdrawAccountSet	true	"账户信息"
//	@Success		200	{object}	pojo.SysUserWithdrawAccountBack
//	@Router			/api/v1/admin/userWithdrawAccount [post]
func AdminSetSysUserWithdrawAccount(ctx *gin.Context) {
	var req pojo.SysUserWithdrawAccountSet
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	hostInfo := ctx.MustGet("hostInfo").(pojo.HostInfo)
	result, err := repository.AdminSetSysUserWithdrawAccount(db, req, hostInfo.ID, req.UserID)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

// AdminDelSysUserWithdrawAccount godoc
//
//	@Summary		管理员删除提现账户
//	@Tags			提现账户
//	@Produce		json
//	@Param			id path		int	true	"账户ID"
//	@Success		200	{object}	string
//	@Router			/api/v1/admin/userWithdrawAccount/:id [delete]
func AdminDelSysUserWithdrawAccount(ctx *gin.Context) {
	id, ok := parsePathID(ctx)
	if !ok {
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result, err := repository.AdminDelSysUserWithdrawAccount(db, id)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

// GetAppWithdrawAccounts godoc
//
//	@Summary		App端获取当前用户提现账户列表
//	@Tags			App提现账户
//	@Produce		json
//	@Success		200	{object}	[]pojo.SysUserWithdrawAccountBack
//	@Router			/api/v1/app/withdrawAccount/list [get]
func GetAppWithdrawAccounts(ctx *gin.Context) {
	userID := ctx.MustGet("userId").(int64)
	db := ctx.MustGet("db").(*gorm.DB)
	result := repository.GetAppWithdrawAccounts(db, userID)
	utils.SuccessObjBack(ctx, result)
}

// AppAddWithdrawAccount godoc
//
//	@Summary		App端新增提现账户
//	@Tags			App提现账户
//	@Accept			json
//	@Produce		json
//	@Param			data body		pojo.SysUserWithdrawAccountSet	true	"账户信息"
//	@Success		200	{object}	pojo.SysUserWithdrawAccountBack
//	@Router			/api/v1/app/withdrawAccount [post]
func AppAddWithdrawAccount(ctx *gin.Context) {
	var req pojo.SysUserWithdrawAccountSet
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	userID := ctx.MustGet("userId").(int64)
	hostInfo := ctx.MustGet("hostInfo").(pojo.HostInfo)
	db := ctx.MustGet("db").(*gorm.DB)
	result, err := repository.AppAddWithdrawAccount(db, req, hostInfo.ID, userID)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

// AppUpdateWithdrawAccount godoc
//
//	@Summary		App端修改提现账户
//	@Tags			App提现账户
//	@Accept			json
//	@Produce		json
//	@Param			id path		int	true	"账户ID"
//	@Param			data body	pojo.SysUserWithdrawAccountSet	true	"账户信息"
//	@Success		200	{object}	pojo.SysUserWithdrawAccountBack
//	@Router			/api/v1/app/withdrawAccount/:id/update [post]
func AppUpdateWithdrawAccount(ctx *gin.Context) {
	id, ok := parsePathID(ctx)
	if !ok {
		return
	}
	var req pojo.SysUserWithdrawAccountSet
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	userID := ctx.MustGet("userId").(int64)
	db := ctx.MustGet("db").(*gorm.DB)
	result, err := repository.AppUpdateWithdrawAccount(db, id, userID, req)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

// AppDelWithdrawAccount godoc
//
//	@Summary		App端删除自己的提现账户
//	@Tags			App提现账户
//	@Produce		json
//	@Param			id path		int	true	"账户ID"
//	@Success		200	{object}	string
//	@Router			/api/v1/app/withdrawAccount/:id [delete]
func AppDelWithdrawAccount(ctx *gin.Context) {
	id, ok := parsePathID(ctx)
	if !ok {
		return
	}
	userID := ctx.MustGet("userId").(int64)
	db := ctx.MustGet("db").(*gorm.DB)
	result, err := repository.AppDelWithdrawAccount(db, id, userID)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

// AppSetDefaultWithdrawAccount godoc
//
//	@Summary		App端设置默认提现账户
//	@Tags			App提现账户
//	@Accept			json
//	@Produce		json
//	@Param			id path		int	true	"账户ID"
//	@Success		200	{object}	string
//	@Router			/api/v1/app/withdrawAccount/:id/setDefault [post]
func AppSetDefaultWithdrawAccount(ctx *gin.Context) {
	id, ok := parsePathID(ctx)
	if !ok {
		return
	}
	userID := ctx.MustGet("userId").(int64)
	db := ctx.MustGet("db").(*gorm.DB)
	result, err := repository.AppSetDefaultWithdrawAccount(db, id, userID)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}
