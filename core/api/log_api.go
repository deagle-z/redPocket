package api

import (
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/repository"
	"BaseGoUni/core/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GetManageLogs godoc
//
//	@Summary		获取管理员操作日志
//	@Tags			系统设置
//	@Accept			json
//	@Produce		json
//	@Param			data body		pojo.ManageLogSearch	true	"获取管理员操作日志"
//	@Success		200	{object}		pojo.ManageLogResp
//	@Router			/api/v1/admin/manage/logs [post]
func GetManageLogs(ctx *gin.Context) {
	currentUser, err := utils.GetCurrentUser(ctx)
	if err != nil {
		utils.UnauthorizedBack(ctx, err.Error())
		return
	}
	var searchData pojo.ManageLogSearch
	err = ctx.BindJSON(&searchData)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result, err := repository.GetManageLogs(db, currentUser, searchData)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}
