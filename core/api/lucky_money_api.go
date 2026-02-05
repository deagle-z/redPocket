package api

import (
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/repository"
	"BaseGoUni/core/services"
	"BaseGoUni/core/utils"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
	"gorm.io/gorm"
)

// SendRedPacket 发送红包
// @Summary 发送红包
// @Tags 红包管理
// @Accept json
// @Produce json
// @Param data body pojo.LuckyMoneySend true "发送红包"
// @Success 200 {object} pojo.BaseObjResponse[pojo.LuckyMoneyBack]
// @Router /api/v1/outside/lucky/send [post]
func SendRedPacket(ctx *gin.Context) {
	currentUser, err := utils.GetCurrentUser(ctx)
	if err != nil {
		utils.UnauthorizedBack(ctx, err.Error())
		return
	}

	var req pojo.LuckyMoneySend
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}

	db := ctx.MustGet("db").(*gorm.DB)
	hostInfo := ctx.MustGet("hostInfo").(pojo.HostInfo)

	// 获取用户名称
	userName := currentUser.Username
	if userName == "" {
		userName = fmt.Sprintf("User_%d", currentUser.ID)
	}

	luckyMoney, err := services.SendRedPacket(db, currentUser.ID, userName, req, hostInfo.TablePrefix)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}

	var result pojo.LuckyMoneyBack
	_ = copier.Copy(&result, luckyMoney)
	utils.SuccessObjBack(ctx, result)
}

// GrabRedPacket 抢红包
// @Summary 抢红包
// @Tags 红包管理
// @Accept json
// @Produce json
// @Param data body pojo.LuckyMoneyGrab true "抢红包"
// @Success 200 {object} pojo.BaseResponse
// @Router /api/v1/outside/lucky/grab [post]
//func GrabRedPacket(ctx *gin.Context) {
//	currentUser, err := utils.GetCurrentUser(ctx)
//	if err != nil {
//		utils.UnauthorizedBack(ctx, err.Error())
//		return
//	}
//
//	var req pojo.LuckyMoneyGrab
//	if err := ctx.ShouldBindJSON(&req); err != nil {
//		utils.ErrorBack(ctx, err.Error())
//		return
//	}
//
//	db := ctx.MustGet("db").(*gorm.DB)
//	hostInfo := ctx.MustGet("hostInfo").(pojo.HostInfo)
//
//	result, err := services.GrabRedPacket(db, req.LuckyID, currentUser.ID, hostInfo.TablePrefix)
//	if err != nil {
//		utils.ErrorBack(ctx, err.Error())
//		return
//	}
//
//	utils.SuccessObjBack(ctx, result)
//}

// GetRedPacketList 获取红包列表
// @Summary 获取红包列表
// @Tags 红包管理
// @Accept json
// @Produce json
// @Param data body pojo.LuckyMoneySearch true "查询条件"
// @Success 200 {object} pojo.LuckyMoneyResp
// @Router /api/v1/outside/lucky/list [post]
func GetRedPacketList(ctx *gin.Context) {
	currentUser, err := utils.GetCurrentUser(ctx)
	if err != nil {
		utils.UnauthorizedBack(ctx, err.Error())
		return
	}

	var search pojo.LuckyMoneySearch
	search.SetPageDefaults()
	if err := ctx.ShouldBindJSON(&search); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}

	db := ctx.MustGet("db").(*gorm.DB)

	// 如果用户不是管理员，只能查看自己发送的红包
	if currentUser.UserType != 1 && currentUser.UserType != 2 {
		search.SenderID = currentUser.ID
	}

	result := repository.GetLuckyMoneyList(db, search)
	utils.SuccessObjBack(ctx, result)
}

// GetRedPacketDetail 获取红包详情
// @Summary 获取红包详情
// @Tags 红包管理
// @Accept json
// @Produce json
// @Param id path int true "红包ID"
// @Success 200 {object} pojo.BaseResponse
// @Router /api/v1/outside/lucky/:id [get]
func GetRedPacketDetail(ctx *gin.Context) {
	luckyIDStr := ctx.Param("id")
	luckyID := int64(0)
	_, err := fmt.Sscanf(luckyIDStr, "%d", &luckyID)
	if err != nil || luckyID <= 0 {
		utils.ErrorBack(ctx, "无效的红包ID")
		return
	}

	db := ctx.MustGet("db").(*gorm.DB)

	result, err := services.GetRedPacketDetails(db, luckyID)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}

	utils.SuccessObjBack(ctx, result)
}

// GetRedPacketStatus 获取红包状态
// @Summary 获取红包状态
// @Tags 红包管理
// @Accept json
// @Produce json
// @Param id path int true "红包ID"
// @Success 200 {object} pojo.BaseResponse
// @Router /api/v1/outside/lucky/status/:id [get]
func GetRedPacketStatus(ctx *gin.Context) {
	luckyIDStr := ctx.Param("id")
	luckyID := int64(0)
	_, err := fmt.Sscanf(luckyIDStr, "%d", &luckyID)
	if err != nil || luckyID <= 0 {
		utils.ErrorBack(ctx, "无效的红包ID")
		return
	}

	db := ctx.MustGet("db").(*gorm.DB)

	luckyMoney, grabbedCount, err := services.GetRedPacketStatus(db, luckyID)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}

	result := map[string]interface{}{
		"luckyMoney":   luckyMoney,
		"grabbedCount": grabbedCount,
		"openNum":      int(grabbedCount),
	}

	utils.SuccessObjBack(ctx, result)
}

// CheckGrabBalance 检查抢包余额
// @Summary 检查抢包余额
// @Tags 红包管理
// @Accept json
// @Produce json
// @Param data body pojo.LuckyMoneyGrab true "红包ID"
// @Success 200 {object} pojo.BaseResponse
// @Router /api/v1/outside/lucky/checkBalance [post]
func CheckGrabBalance(ctx *gin.Context) {
	currentUser, err := utils.GetCurrentUser(ctx)
	if err != nil {
		utils.UnauthorizedBack(ctx, err.Error())
		return
	}

	var req pojo.LuckyMoneyGrab
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}

	db := ctx.MustGet("db").(*gorm.DB)
	hostInfo := ctx.MustGet("hostInfo").(pojo.HostInfo)

	err = services.CheckGrabBalance(db, req.LuckyID, currentUser.ID, hostInfo.TablePrefix)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}

	utils.SuccessBack(ctx, "余额充足")
}
