package api

import (
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/utils"
	tenantRepo "BaseGoUni/tenant/repository"
	"github.com/gin-gonic/gin"
	"io"
)

func GetTgUsers(ctx *gin.Context) {
	tenantID, ok := getTenantID(ctx)
	if !ok {
		return
	}
	var search pojo.TgUserSearch
	search.SetPageDefaults()
	if err := ctx.ShouldBindJSON(&search); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	result := tenantRepo.GetTgUsers(getDB(ctx), tenantID, search)
	utils.SuccessObjBack(ctx, result)
}

func GetTgUserById(ctx *gin.Context) {
	tenantID, ok := getTenantID(ctx)
	if !ok {
		return
	}
	id, ok := parsePathID(ctx)
	if !ok {
		return
	}
	result, err := tenantRepo.GetTgUserByID(getDB(ctx), tenantID, id)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

func GetTgUserWithdrawActivityFlow(ctx *gin.Context) {
	tenantID, ok := getTenantID(ctx)
	if !ok {
		return
	}
	id, ok := parsePathID(ctx)
	if !ok {
		return
	}
	result, err := tenantRepo.GetTgUserWithdrawActivityFlow(getDB(ctx), tenantID, id)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

func SetTgUser(ctx *gin.Context) {
	tenantID, ok := getTenantID(ctx)
	if !ok {
		return
	}
	var req pojo.TgUserSet
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	result, err := tenantRepo.SetTgUser(getDB(ctx), tenantID, req)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

func SetTgUserStatus(ctx *gin.Context) {
	tenantID, ok := getTenantID(ctx)
	if !ok {
		return
	}
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
	result, err := tenantRepo.SetTgUserStatus(getDB(ctx), tenantID, req.ID, req.Status)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

func SetTgUserRebateRate(ctx *gin.Context) {
	tenantID, ok := getTenantID(ctx)
	if !ok {
		return
	}
	var req pojo.TgUserRebateRateSet
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	if req.ID <= 0 || req.RebateRate < 0 || req.RebateRate > 100 {
		utils.ErrorBack(ctx, "参数格式错误")
		return
	}
	result, err := tenantRepo.SetTgUserRebateRate(getDB(ctx), tenantID, req.ID, req.RebateRate)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

func SetTgUserRebateType(ctx *gin.Context) {
	tenantID, ok := getTenantID(ctx)
	if !ok {
		return
	}
	var req pojo.TgUserRebateTypeSet
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	if req.ID <= 0 || (req.RebateType != 1 && req.RebateType != 2) {
		utils.ErrorBack(ctx, "参数格式错误")
		return
	}
	result, err := tenantRepo.SetTgUserRebateType(getDB(ctx), tenantID, req.ID, req.RebateType)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

func SetTgUserRebateWithdrawDisabled(ctx *gin.Context) {
	tenantID, ok := getTenantID(ctx)
	if !ok {
		return
	}
	var req pojo.TgUserRebateWithdrawDisabledSet
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	if req.ID <= 0 || (req.RebateWithdrawDisabled != 0 && req.RebateWithdrawDisabled != 1) {
		utils.ErrorBack(ctx, "参数格式错误")
		return
	}
	result, err := tenantRepo.SetTgUserRebateWithdrawDisabled(getDB(ctx), tenantID, req.ID, req.RebateWithdrawDisabled)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

func SetTgUserRechargeRebateRates(ctx *gin.Context) {
	tenantID, ok := getTenantID(ctx)
	if !ok {
		return
	}
	var req pojo.TgUserRechargeRebateRatesSet
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	if req.ID <= 0 {
		utils.ErrorBack(ctx, "参数格式错误")
		return
	}
	result, err := tenantRepo.SetTgUserRechargeRebateRates(getDB(ctx), tenantID, req.ID, req.RechargeRebateRates)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

func AddTgUserRebateAmount(ctx *gin.Context) {
	tenantID, ok := getTenantID(ctx)
	if !ok {
		return
	}
	var req pojo.TgUserRebateAmountAdd
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	req.Amount = utils.Truncate2(req.Amount)
	if req.ID <= 0 || req.Amount <= 0 {
		utils.ErrorBack(ctx, "参数格式错误")
		return
	}
	result, err := tenantRepo.AddTgUserRebateAmount(getDB(ctx), tenantID, req.ID, req.Amount)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

func TenantCreateRechargeOrderV2(ctx *gin.Context) {
	tenantID, ok := getTenantID(ctx)
	if !ok {
		return
	}
	var req pojo.AdminCreateRechargeOrderV2Req
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ErrorBack(ctx, "参数格式错误")
		return
	}
	if req.UserID <= 0 {
		utils.ErrorBack(ctx, "参数格式错误")
		return
	}
	hostInfo := ctx.MustGet("hostInfo").(pojo.HostInfo)
	result, err := tenantRepo.AdminCreateRechargeOrderV2(getDB(ctx), tenantID, req, hostInfo.TablePrefix)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

func SetTgUserRemark(ctx *gin.Context) {
	tenantID, ok := getTenantID(ctx)
	if !ok {
		return
	}
	var req pojo.TgUserRemarkSet
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	if req.ID <= 0 || len([]rune(req.Remark)) > 255 {
		utils.ErrorBack(ctx, "参数格式错误")
		return
	}
	result, err := tenantRepo.SetTgUserRemark(getDB(ctx), tenantID, req.ID, req.Remark)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

func DelTgUser(ctx *gin.Context) {
	tenantID, ok := getTenantID(ctx)
	if !ok {
		return
	}
	id, ok := parsePathID(ctx)
	if !ok {
		return
	}
	result, err := tenantRepo.DelTgUser(getDB(ctx), tenantID, id)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

// GetTgUsersWithSubStats 列表返回所有下级（不限层级）的充值金额、流水、提现金额
func GetTgUsersWithSubStats(ctx *gin.Context) {
	tenantID, ok := getTenantID(ctx)
	if !ok {
		return
	}
	var search pojo.TgUserSearch
	search.SetPageDefaults()
	if err := ctx.ShouldBindJSON(&search); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	result := tenantRepo.GetTgUsersWithSubStats(getDB(ctx), tenantID, search)
	utils.SuccessObjBack(ctx, result)
}

// GetTgUsersWithSubStatsSummary 展示所有不限层级下级的充值金额之和、流水之和、提现金额之和
func GetTgUsersWithSubStatsSummary(ctx *gin.Context) {
	tenantID, ok := getTenantID(ctx)
	if !ok {
		return
	}
	var search pojo.TgUserSearch
	if err := ctx.ShouldBindJSON(&search); err != nil && err != io.EOF {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	result := tenantRepo.GetTgUsersWithSubStatsSummary(getDB(ctx), tenantID, search)
	utils.SuccessObjBack(ctx, result)
}
