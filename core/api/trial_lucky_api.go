package api

import (
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/repository"
	"BaseGoUni/core/services"
	"BaseGoUni/core/utils"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
	"gorm.io/gorm"
	"io"
)

func GetTrialMeApp(ctx *gin.Context) {
	userID := ctx.MustGet("userId").(int64)
	db := ctx.MustGet("db").(*gorm.DB)
	result, err := services.GetTrialMe(db, userID)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

func GetTrialLuckyListApp(ctx *gin.Context) {
	var search pojo.TrialLuckyMoneyAppListSearch
	search.SetPageDefaults()
	if err := ctx.ShouldBindJSON(&search); err != nil && err != io.EOF {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	var userID int64
	if raw, ok := ctx.Get("userId"); ok {
		if id, ok := raw.(int64); ok {
			userID = id
		}
	}
	db := ctx.MustGet("db").(*gorm.DB)
	hostInfo := ctx.MustGet("hostInfo").(pojo.HostInfo)
	_ = services.EnsureMinActiveTrialLuckyPackets(db, hostInfo.TablePrefix)
	utils.SuccessObjBack(ctx, services.GetTrialLuckyAppList(db, search, userID))
}

func SendTrialLuckyApp(ctx *gin.Context) {
	var req pojo.TrialLuckyMoneySend
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	userID := ctx.MustGet("userId").(int64)
	db := ctx.MustGet("db").(*gorm.DB)
	hostInfo := ctx.MustGet("hostInfo").(pojo.HostInfo)
	lucky, err := services.SendTrialRedPacket(db, userID, req, hostInfo.TablePrefix)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	var result pojo.TrialLuckyMoneyBack
	_ = copier.Copy(&result, &lucky)
	utils.SuccessObjBack(ctx, result)
}

func GrabTrialLuckyApp(ctx *gin.Context) {
	var req pojo.TrialLuckyMoneyGrab
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	grabIndex := 0
	if req.GrabIndex != nil {
		grabIndex = *req.GrabIndex
		if grabIndex <= 0 {
			utils.ErrorBack(ctx, "grab_index_positive")
			return
		}
	}
	userID := ctx.MustGet("userId").(int64)
	db := ctx.MustGet("db").(*gorm.DB)
	hostInfo := ctx.MustGet("hostInfo").(pojo.HostInfo)
	result, err := services.GrabTrialRedPacket(db, req.LuckyID, userID, hostInfo.TablePrefix, grabIndex, req.OddEvenGuess)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

func GetTrialLuckyHistoryApp(ctx *gin.Context) {
	var search pojo.TrialCashHistorySearch
	search.SetPageDefaults()
	if err := ctx.ShouldBindJSON(&search); err != nil && err != io.EOF {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	userID := ctx.MustGet("userId").(int64)
	db := ctx.MustGet("db").(*gorm.DB)
	utils.SuccessObjBack(ctx, services.GetTrialAppHistory(db, userID, search))
}

func GetTrialLuckyListAdmin(ctx *gin.Context) {
	var search pojo.TrialLuckyMoneySearch
	search.SetPageDefaults()
	if err := ctx.ShouldBindJSON(&search); err != nil && err != io.EOF {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	utils.SuccessObjBack(ctx, repository.GetTrialLuckyMoneyList(db, search))
}

func GetTrialCashHistoryListAdmin(ctx *gin.Context) {
	var search pojo.TrialCashHistorySearch
	search.SetPageDefaults()
	if err := ctx.ShouldBindJSON(&search); err != nil && err != io.EOF {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	utils.SuccessObjBack(ctx, repository.GetTrialCashHistoryList(db, search))
}
