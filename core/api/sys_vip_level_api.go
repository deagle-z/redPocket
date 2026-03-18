package api

import (
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/repository"
	"BaseGoUni/core/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"strconv"
)

// GetSysVipLevels godoc
//
//	@Summary		获取VIP等级列表
//	@Tags			VIP等级
//	@Accept			json
//	@Produce		json
//	@Param			data body		pojo.SysVipLevelSearch	true	"查询条件"
//	@Success		200	{object}	pojo.SysVipLevelResp
//	@Router			/api/v1/admin/sysVipLevel/list [post]
func GetSysVipLevels(ctx *gin.Context) {
	var search pojo.SysVipLevelSearch
	search.SetPageDefaults()
	if err := ctx.ShouldBindJSON(&search); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result := repository.GetSysVipLevels(db, search)
	utils.SuccessObjBack(ctx, result)
}

// SetSysVipLevel godoc
//
//	@Summary		创建或更新VIP等级
//	@Tags			VIP等级
//	@Accept			json
//	@Produce		json
//	@Param			data body		pojo.SysVipLevelSet	true	"VIP等级信息"
//	@Success		200	{object}	pojo.SysVipLevelBack
//	@Router			/api/v1/admin/sysVipLevel [post]
func SetSysVipLevel(ctx *gin.Context) {
	var req pojo.SysVipLevelSet
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result, err := repository.SetSysVipLevel(db, req)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

// DelSysVipLevel godoc
//
//	@Summary		删除VIP等级
//	@Tags			VIP等级
//	@Accept			json
//	@Produce		json
//	@Param			id path		int	true	"VIP等级ID"
//	@Success		200	{object}	string
//	@Router			/api/v1/admin/sysVipLevel/:id [delete]
func DelSysVipLevel(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.ErrorBack(ctx, "参数格式错误")
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result, err := repository.DelSysVipLevel(db, id)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

// AppGetClaimableVipRewards godoc
//
//	@Summary		查询当前用户可领取的VIP奖励列表
//	@Tags			VIP等级
//	@Produce		json
//	@Success		200	{object}	[]pojo.SysVipRewardLogBack
//	@Router			/api/v1/app/vip/rewards [get]
func AppGetClaimableVipRewards(ctx *gin.Context) {
	userID := ctx.MustGet("userId").(int64)
	db := ctx.MustGet("db").(*gorm.DB)
	result, err := repository.GetClaimableVipRewards(db, userID)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

// AppClaimVipReward godoc
//
//	@Summary		领取VIP奖励（id=0 领取全部）
//	@Tags			VIP等级
//	@Produce		json
//	@Param			id	path	int	false	"奖励记录ID，0或不传则领取全部"
//	@Success		200	{object}	string
//	@Router			/api/v1/app/vip/rewards/:id/claim [post]
func AppClaimVipReward(ctx *gin.Context) {
	userID := ctx.MustGet("userId").(int64)
	idStr := ctx.Param("id")
	var rewardID int64
	var err error
	if idStr != "" && idStr != "0" {
		rewardID, err = strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			utils.ErrorBack(ctx, "参数格式错误")
			return
		}
	}
	db := ctx.MustGet("db").(*gorm.DB)
	hostInfo := ctx.MustGet("hostInfo").(pojo.HostInfo)
	if err := repository.ClaimVipReward(db, userID, rewardID, hostInfo.TablePrefix); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, "领取成功")
}

// AppGetVipProgress godoc
//
//	@Summary		获取当前用户VIP进度
//	@Tags			VIP等级
//	@Produce		json
//	@Success		200	{object}	pojo.AppVipProgressBack
//	@Router			/api/v1/app/vip/progress [get]
func AppGetVipProgress(ctx *gin.Context) {
	userID := ctx.MustGet("userId").(int64)
	db := ctx.MustGet("db").(*gorm.DB)
	result, err := repository.GetAppVipProgress(db, userID)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

// GetSysVipLevelById godoc
//
//	@Summary		根据ID获取VIP等级
//	@Tags			VIP等级
//	@Accept			json
//	@Produce		json
//	@Param			id path		int	true	"VIP等级ID"
//	@Success		200	{object}	pojo.SysVipLevelBack
//	@Router			/api/v1/admin/sysVipLevel/:id [get]
func GetSysVipLevelById(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.ErrorBack(ctx, "参数格式错误")
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result, err := repository.GetSysVipLevelById(db, id)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}
