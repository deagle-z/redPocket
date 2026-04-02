package api

import (
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/repository"
	"BaseGoUni/core/utils"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GetSysBanners godoc
//
//	@Summary		获取轮播图列表
//	@Tags			轮播图
//	@Accept			json
//	@Produce		json
//	@Param			data body		pojo.SysBannerSearch	true	"查询条件"
//	@Success		200	{object}		pojo.SysBannerResp
//	@Router			/api/v1/admin/sysBanner/list [post]
func GetSysBanners(ctx *gin.Context) {
	var search pojo.SysBannerSearch
	search.SetPageDefaults()
	if err := ctx.ShouldBindJSON(&search); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result := repository.GetSysBanners(db, search)
	utils.SuccessObjBack(ctx, result)
}

// SetSysBanner godoc
//
//	@Summary		创建或更新轮播图
//	@Tags			轮播图
//	@Accept			json
//	@Produce		json
//	@Param			data body		pojo.SysBannerSet	true	"轮播图信息"
//	@Success		200	{object}		pojo.SysBannerBack
//	@Router			/api/v1/admin/sysBanner [post]
func SetSysBanner(ctx *gin.Context) {
	var req pojo.SysBannerSet
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result, err := repository.SetSysBanner(db, req)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

// DelSysBanner godoc
//
//	@Summary		删除轮播图
//	@Tags			轮播图
//	@Accept			json
//	@Produce		json
//	@Param			id path		int	true	"轮播图ID"
//	@Success		200	{object}		string
//	@Router			/api/v1/admin/sysBanner/:id [delete]
func DelSysBanner(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.ErrorBack(ctx, "参数格式错误")
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result, err := repository.DelSysBanner(db, id)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

// GetAppBanners godoc
//
//	@Summary		App端获取轮播图（按position分组）
//	@Tags			轮播图
//	@Accept			json
//	@Produce		json
//	@Param			data body		pojo.SysBannerAppReq	false	"平台"
//	@Success		200	{object}		pojo.SysBannerGroupedResp
//	@Router			/api/v1/app/banners [post]
func GetAppBanners(ctx *gin.Context) {
	var req pojo.SysBannerAppReq
	_ = ctx.ShouldBindJSON(&req)
	db := ctx.MustGet("db").(*gorm.DB)
	result := repository.GetSysBannersGroupedByPosition(db, req.Platform)
	utils.SuccessObjBack(ctx, result)
}

// GetSysBannerById godoc
//
//	@Summary		根据ID获取轮播图
//	@Tags			轮播图
//	@Accept			json
//	@Produce		json
//	@Param			id path		int	true	"轮播图ID"
//	@Success		200	{object}		pojo.SysBannerBack
//	@Router			/api/v1/admin/sysBanner/:id [get]
func GetSysBannerById(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.ErrorBack(ctx, "参数格式错误")
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result, err := repository.GetSysBannerById(db, id)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}
