package api

import (
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/repository"
	"BaseGoUni/core/services"
	"BaseGoUni/core/utils"
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
	"gorm.io/gorm"
	"io"
	"strings"
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

	// 发包成功后，向所有在线用户广播最新红包信息（包含 LuckyMoneyBack）
	_ = utils.BroadcastWsWithType("lucky_sent", result)

	utils.SuccessObjBack(ctx, result)
}

// SendRedPacketApp app端发送红包（TG用户）
func SendRedPacketApp(ctx *gin.Context) {
	var req pojo.LuckyMoneySend
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}

	userIDRaw, ok := ctx.Get("userId")
	if !ok {
		utils.UnauthorizedBack(ctx, "token is invalid")
		return
	}
	userID, ok := userIDRaw.(int64)
	if !ok || userID <= 0 {
		utils.UnauthorizedBack(ctx, "token is invalid")
		return
	}

	db := ctx.MustGet("db").(*gorm.DB)
	hostInfo := ctx.MustGet("hostInfo").(pojo.HostInfo)

	var tgUser pojo.TgUser
	if err := db.Where("id = ?", userID).First(&tgUser).Error; err != nil || tgUser.ID == 0 {
		utils.ErrorBack(ctx, "用户不存在")
		return
	}
	if tgUser.Status != 1 {
		utils.ErrorBack(ctx, "用户已禁用，请联系管理员处理")
		return
	}

	userName := ""
	if tgUser.FirstName != nil && strings.TrimSpace(*tgUser.FirstName) != "" {
		userName = strings.TrimSpace(*tgUser.FirstName)
	} else if tgUser.Username != nil && strings.TrimSpace(*tgUser.Username) != "" {
		userName = strings.TrimSpace(*tgUser.Username)
	} else {
		userName = fmt.Sprintf("User_%d", tgUser.ID)
	}

	luckyMoney, err := services.SendRedPacket(db, userID, userName, req, hostInfo.TablePrefix)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}

	var result pojo.LuckyMoneyBack
	_ = copier.Copy(&result, luckyMoney)

	// app发包成功后，向所有在线用户广播最新红包信息（包含 LuckyMoneyBack）
	_ = utils.BroadcastWsWithType("lucky_sent", result)

	utils.SuccessObjBack(ctx, result)
}

// GetRedPacketListApp app端红包大厅列表
func GetRedPacketListApp(ctx *gin.Context) {
	var search pojo.LuckyMoneyAppListSearch
	search.SetPageDefaults()
	if err := ctx.ShouldBindJSON(&search); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}

	userIDRaw, ok := ctx.Get("userId")
	if !ok {
		utils.UnauthorizedBack(ctx, "token is invalid")
		return
	}
	userID, ok := userIDRaw.(int64)
	if !ok || userID <= 0 {
		utils.UnauthorizedBack(ctx, "token is invalid")
		return
	}

	db := ctx.MustGet("db").(*gorm.DB)
	result := repository.GetLuckyMoneyAppList(db, search, userID)
	utils.SuccessObjBack(ctx, result)
}

// GetLuckyDetailApp app端红包详情
func GetLuckyDetailApp(ctx *gin.Context) {
	var req pojo.LuckyMoneyAppDetailReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}

	userIDRaw, ok := ctx.Get("userId")
	if !ok {
		utils.UnauthorizedBack(ctx, "token is invalid")
		return
	}
	userID, ok := userIDRaw.(int64)
	if !ok || userID <= 0 {
		utils.UnauthorizedBack(ctx, "token is invalid")
		return
	}

	db := ctx.MustGet("db").(*gorm.DB)
	result, err := repository.GetLuckyMoneyAppDetail(db, req.LuckyID, userID)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

// GetRecentLuckyWinnersApp app端最近中奖列表
func GetRecentLuckyWinnersApp(ctx *gin.Context) {
	var search pojo.LuckyRecentWinnerSearch
	if err := ctx.ShouldBindJSON(&search); err != nil && err != io.EOF {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result := repository.GetRecentLuckyWinners(db, search)
	utils.SuccessObjBack(ctx, result)
}

// GetLuckyAppHistory app端发包+抢包历史（union）
func GetLuckyAppHistory(ctx *gin.Context) {
	var search pojo.LuckyAppHistorySearch
	search.SetPageDefaults()
	if err := ctx.ShouldBindJSON(&search); err != nil && err != io.EOF {
		utils.ErrorBack(ctx, err.Error())
		return
	}

	userIDRaw, ok := ctx.Get("userId")
	if !ok {
		utils.UnauthorizedBack(ctx, "token is invalid")
		return
	}
	userID, ok := userIDRaw.(int64)
	if !ok || userID <= 0 {
		utils.UnauthorizedBack(ctx, "token is invalid")
		return
	}

	db := ctx.MustGet("db").(*gorm.DB)
	result := repository.GetLuckyAppHistoryUnion(db, userID, search)
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

// GrabRedPacketApp app端抢红包（TG用户）
func GrabRedPacketApp(ctx *gin.Context) {
	authHeader := strings.TrimSpace(ctx.GetHeader("Authorization"))
	if authHeader == "" {
		utils.UnauthorizedBack(ctx, "Authorization header is missing")
		return
	}
	token := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
	hostInfo := ctx.MustGet("hostInfo").(pojo.HostInfo)
	userID, userType, _, _, err := utils.ParseToken(hostInfo.AccessSecret, token)
	if err != nil || userID <= 0 || userType != 5 {
		utils.UnauthorizedBack(ctx, "token is invalid")
		return
	}
	key := utils.KeyRdTgOnline + utils.MD5(token)
	data := utils.RD.Get(context.Background(), key)
	if data == nil || data.Err() != nil || data.Val() == "" {
		utils.UnauthorizedBack(ctx, "token is passed")
		return
	}

	var req pojo.LuckyMoneyGrab
	if err = ctx.ShouldBindJSON(&req); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	grabIndex := 0
	if req.GrabIndex != nil {
		grabIndex = *req.GrabIndex
		if grabIndex <= 0 {
			utils.ErrorBack(ctx, "grabIndex must be greater than 0")
			return
		}
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result, err := services.GrabRedPacket(db, req.LuckyID, userID, hostInfo.TablePrefix, grabIndex)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}

	_ = services.BroadcastLuckyGrabResult(db, req.LuckyID, result)

	utils.SuccessObjBack(ctx, result)
}
