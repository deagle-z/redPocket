package api

import (
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/repository"
	"BaseGoUni/core/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"io"
	"strconv"
)

func GetTrialBotUsers(ctx *gin.Context) {
	var search pojo.TrialBotUserSearch
	search.SetPageDefaults()
	if err := ctx.ShouldBindJSON(&search); err != nil && err != io.EOF {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	utils.SuccessObjBack(ctx, repository.GetTrialBotUsers(db, search))
}

func BatchCreateTrialBotUsers(ctx *gin.Context) {
	var req pojo.TrialBotBatchCreateReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result, err := repository.BatchCreateTrialBotUsers(db, req)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

func BatchUpdateTrialBotUsers(ctx *gin.Context) {
	var req pojo.TrialBotBatchUpdateReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result, err := repository.BatchUpdateTrialBotUsers(db, req)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

func SetTrialBotUserStatus(ctx *gin.Context) {
	var req pojo.TrialBotStatusSet
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result, err := repository.SetTrialBotUserStatus(db, req.ID, req.Status)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

func DelTrialBotUser(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		utils.ErrorBack(ctx, "invalid_params")
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result, err := repository.DelTrialBotUser(db, id)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}
