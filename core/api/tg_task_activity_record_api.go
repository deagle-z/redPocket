package api

import (
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/repository"
	"BaseGoUni/core/utils"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GetTgTaskActivityRecords godoc
//
//	@Summary		获取任务活动领取记录列表
//	@Tags			任务活动记录
//	@Accept			json
//	@Produce		json
//	@Param			data body		pojo.TgTaskActivityRecordSearch	true	"查询条件"
//	@Success		200	{object}	pojo.TgTaskActivityRecordResp
//	@Router			/api/v1/admin/taskActivityRecord/list [post]
func GetTgTaskActivityRecords(ctx *gin.Context) {
	var search pojo.TgTaskActivityRecordSearch
	search.SetPageDefaults()
	if err := ctx.ShouldBindJSON(&search); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	utils.SuccessObjBack(ctx, repository.GetTgTaskActivityRecords(db, search))
}

// GetTgTaskActivityRecordById godoc
//
//	@Summary		根据ID获取任务活动领取记录
//	@Tags			任务活动记录
//	@Produce		json
//	@Param			id path	int	true	"记录ID"
//	@Success		200	{object}	pojo.TgTaskActivityRecordBack
//	@Router			/api/v1/admin/taskActivityRecord/:id [get]
func GetTgTaskActivityRecordById(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		utils.ErrorBack(ctx, "参数格式错误")
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result, err := repository.GetTgTaskActivityRecordById(db, id)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}
