package api

import (
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/repository"
	"BaseGoUni/core/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"log"
	"strconv"
)

// GetSysCountries godoc
//
//	@Summary		获取国家列表
//	@Tags			国家
//	@Accept			json
//	@Produce		json
//	@Param			data body		pojo.SysCountrySearch	true	"查询条件"
//	@Success		200	{object}		pojo.SysCountryResp
//	@Router			/api/v1/admin/sysCountry/list [post]
func GetSysCountries(ctx *gin.Context) {
	var search pojo.SysCountrySearch
	search.SetPageDefaults()
	if err := ctx.ShouldBindJSON(&search); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result := repository.GetSysCountries(db, search)
	utils.SuccessObjBack(ctx, result)
}

// SetSysCountry godoc
//
//	@Summary		创建或更新国家
//	@Tags			国家
//	@Accept			json
//	@Produce		json
//	@Param			data body		pojo.SysCountrySet	true	"国家信息"
//	@Success		200	{object}		pojo.SysCountryBack
//	@Router			/api/v1/admin/sysCountry [post]
func SetSysCountry(ctx *gin.Context) {
	var req pojo.SysCountrySet
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result, err := repository.SetSysCountry(db, req)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

// DelSysCountry godoc
//
//	@Summary		删除国家
//	@Tags			国家
//	@Accept			json
//	@Produce		json
//	@Param			id path		int	true	"国家ID"
//	@Success		200	{object}		string
//	@Router			/api/v1/admin/sysCountry/:id [delete]
func DelSysCountry(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.ErrorBack(ctx, "参数格式错误")
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result, err := repository.DelSysCountry(db, id)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

// GetSysCountryById godoc
//
//	@Summary		根据ID获取国家
//	@Tags			国家
//	@Accept			json
//	@Produce		json
//	@Param			id path		int	true	"国家ID"
//	@Success		200	{object}		pojo.SysCountryBack
//	@Router			/api/v1/admin/sysCountry/:id [get]
func GetSysCountryById(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.ErrorBack(ctx, "参数格式错误")
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result, err := repository.GetSysCountryById(db, id)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

// GetAppCountries godoc
//
//	@Summary		App端获取所有可用国家（IP自动定位，匹配国家置顶）
//	@Tags			App国家
//	@Produce		json
//	@Success		200	{object}		[]pojo.AppCountryItem
//	@Router			/api/v1/app/countries [get]
func GetAppCountries(ctx *gin.Context) {
	ip := utils.GetIPAddress(ctx)
	detectedCode := utils.GetCountryCodeByIP(ip)
	db := ctx.MustGet("db").(*gorm.DB)
	result := repository.GetAppCountries(db, detectedCode)
	utils.SuccessObjBack(ctx, result)
}

// GetCountryRechargeInfo godoc
//
//	@Summary		App端获取国家充值信息（充值字段+通道+支付方式）
//	@Tags			App国家
//	@Produce		json
//	@Param			code path	string	true	"国家码，如 IN / BR"
//	@Success		200	{object}		pojo.AppCountryRechargeInfo
//	@Router			/api/v1/app/country/:code/recharge [get]
func GetCountryRechargeInfo(ctx *gin.Context) {
	code := ctx.Param("code")
	db := ctx.MustGet("db").(*gorm.DB)
	result, err := repository.GetCountryRechargeInfo(db, code)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

// GetCountryWithdrawFields godoc
//
//	@Summary		App端获取国家提现字段配置
//	@Tags			App国家
//	@Produce		json
//	@Param			code path	string	true	"国家码，如 IN / BR"
//	@Success		200	{object}		interface{}
//	@Router			/api/v1/app/country/:code/withdrawFields [get]
func GetCountryWithdrawFields(ctx *gin.Context) {
	code := ctx.Param("code")
	db := ctx.MustGet("db").(*gorm.DB)
	result, err := repository.GetCountryWithdrawFields(db, code)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

// GetCountryRechargeFields godoc
//
//	@Summary		App端获取国家充值字段配置
//	@Tags			App国家
//	@Produce		json
//	@Param			code path	string	true	"国家码，如 IN / BR"
//	@Success		200	{object}		interface{}
//	@Router			/api/v1/app/country/:code/rechargeFields [get]
func GetCountryRechargeFields(ctx *gin.Context) {
	code := ctx.Param("code")
	db := ctx.MustGet("db").(*gorm.DB)
	result, err := repository.GetCountryRechargeFields(db, code)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	log.Printf("result: %v", result)
	utils.SuccessObjBack(ctx, result)
}
