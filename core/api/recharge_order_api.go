package api

import (
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/repository"
	"BaseGoUni/core/services"
	"BaseGoUni/core/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"log"
	"strconv"
	"strings"
	"time"
)

const defaultAdminRechargeReturnURL = "https://example.com"

// GetRechargeOrders godoc
//
//	@Summary		获取充值订单列表
//	@Tags			充值订单
//	@Accept			json
//	@Produce		json
//	@Param			data body		pojo.RechargeOrderSearch	true	"查询条件"
//	@Success		200	{object}		pojo.RechargeOrderResp
//	@Router			/api/v1/admin/rechargeOrder/list [post]
func GetRechargeOrders(ctx *gin.Context) {
	var search pojo.RechargeOrderSearch
	search.SetPageDefaults()
	if err := ctx.ShouldBindJSON(&search); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result := repository.GetRechargeOrders(db, search)
	utils.SuccessObjBack(ctx, result)
}

// SetRechargeOrder godoc
//
//	@Summary		创建或更新充值订单
//	@Tags			充值订单
//	@Accept			json
//	@Produce		json
//	@Param			data body		pojo.RechargeOrderSet	true	"充值订单信息"
//	@Success		200	{object}		pojo.RechargeOrderBack
//	@Router			/api/v1/admin/rechargeOrder [post]
func SetRechargeOrder(ctx *gin.Context) {
	var req pojo.RechargeOrderSet
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result, err := repository.SetRechargeOrder(db, req)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

// DelRechargeOrder godoc
//
//	@Summary		删除充值订单
//	@Tags			充值订单
//	@Accept			json
//	@Produce		json
//	@Param			id path		int	true	"充值订单ID"
//	@Success		200	{object}		string
//	@Router			/api/v1/admin/rechargeOrder/:id [delete]
func DelRechargeOrder(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.ErrorBack(ctx, "参数格式错误")
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result, err := repository.DelRechargeOrder(db, id)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

// GetRechargeOrderById godoc
//
//	@Summary		根据ID获取充值订单
//	@Tags			充值订单
//	@Accept			json
//	@Produce		json
//	@Param			id path		int	true	"充值订单ID"
//	@Success		200	{object}		pojo.RechargeOrderBack
//	@Router			/api/v1/admin/rechargeOrder/:id [get]
func GetRechargeOrderById(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.ErrorBack(ctx, "参数格式错误")
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result, err := repository.GetRechargeOrderById(db, id)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

func GetFrontendUnackedRechargeOrders(ctx *gin.Context) {
	var search pojo.RechargeOrderSearch
	search.SetPageDefaults()
	if err := ctx.ShouldBindJSON(&search); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result := repository.GetFrontendUnackedRechargeOrders(db, search)
	utils.SuccessObjBack(ctx, result)
}

// AdminRechargeOrderCallback 管理员手动触发充值回调
func AdminRechargeOrderCallback(ctx *gin.Context) {
	idStr := ctx.Param("id")
	log.Printf("[recharge] admin callback api hit method=%s path=%s id=%s clientIP=%s host=%s",
		ctx.Request.Method, ctx.Request.URL.Path, idStr, ctx.ClientIP(), ctx.Request.Host)
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		log.Printf("[recharge] admin callback api invalid id id=%s err=%v", idStr, err)
		utils.ErrorBack(ctx, "参数格式错误")
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	hostInfo := ctx.MustGet("hostInfo").(pojo.HostInfo)
	log.Printf("[recharge] admin callback api dispatch orderID=%d tablePrefix=%q hostID=%d host=%s",
		id, hostInfo.TablePrefix, hostInfo.ID, hostInfo.HostName)
	result, err := repository.AdminRechargeOrderCallback(db, id, hostInfo.TablePrefix)
	if err != nil {
		log.Printf("[recharge] admin callback api failed orderID=%d tablePrefix=%q err=%v", id, hostInfo.TablePrefix, err)
		utils.ErrorBack(ctx, err.Error())
		return
	}
	log.Printf("[recharge] admin callback api success orderID=%d tablePrefix=%q", id, hostInfo.TablePrefix)
	utils.SuccessObjBack(ctx, result)
}

// AppCreateRechargeOrder godoc
//
//	@Summary		app端创建充值订单
//	@Tags			充值订单
//	@Accept			json
//	@Produce		json
//	@Param			data body		pojo.RechargeOrderAppReq	true	"充值下单参数"
//	@Success		200	{object}		pojo.RechargeOrderAppBack
//	@Router			/api/v1/app/rechargeOrder [post]
func AppCreateRechargeOrder(ctx *gin.Context) {
	var req pojo.RechargeOrderAppReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ErrorBack(ctx, "参数格式错误")
		return
	}
	req.ClientIP = ctx.ClientIP()
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
	result, err := repository.AppCreateRechargeOrder(db, userID, req, hostInfo.TablePrefix)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

// AppCreateRechargeOrderV2 godoc
//
//	@Summary		app端创建v2充值订单
//	@Tags			充值订单
//	@Accept			json
//	@Produce		json
//	@Param			data body		pojo.RechargeOrderAppReq	true	"充值下单参数"
//	@Success		200	{object}		pojo.RechargeOrderAppBack
//	@Router			/api/v1/app/rechargeOrder/v2 [post]
func AppCreateRechargeOrderV2(ctx *gin.Context) {
	var req pojo.RechargeOrderAppReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ErrorBack(ctx, "参数格式错误")
		return
	}
	req.ClientIP = ctx.ClientIP()
	req.ReturnURL = ctx.GetHeader("Referer")
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
	result, err := repository.AppCreateRechargeOrderV2(db, userID, req, hostInfo.TablePrefix)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

func GetCryptoRechargeOptions(ctx *gin.Context) {
	amount, err := strconv.ParseFloat(strings.TrimSpace(ctx.Query("amount")), 64)
	if err != nil || amount <= 0 {
		utils.ErrorBack(ctx, "invalid_params")
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	hostInfo := ctx.MustGet("hostInfo").(pojo.HostInfo)
	result, err := repository.GetCryptoRechargeOptions(db, hostInfo.TablePrefix, amount)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

func AppCreateCryptoRechargeOrder(ctx *gin.Context) {
	var req pojo.CryptoRechargeOrderReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ErrorBack(ctx, "参数格式错误")
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
	result, err := repository.AppCreateCryptoRechargeOrder(db, userID, req, hostInfo.TablePrefix)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	enqueueCryptoRechargeExpire(hostInfo.TablePrefix, result)
	utils.SuccessObjBack(ctx, result)
}

func GetCryptoRechargeOrderStatus(ctx *gin.Context) {
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
	orderNo := strings.TrimSpace(ctx.Param("orderNo"))
	if orderNo == "" {
		utils.ErrorBack(ctx, "order_no_required")
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result, err := repository.GetCryptoRechargeOrderStatus(db, userID, orderNo)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

func CancelCryptoRechargeOrder(ctx *gin.Context) {
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
	orderNo := strings.TrimSpace(ctx.Param("orderNo"))
	if orderNo == "" {
		utils.ErrorBack(ctx, "order_no_required")
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result, err := repository.CancelCryptoRechargeOrder(db, userID, orderNo)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

func enqueueCryptoRechargeExpire(tablePrefix string, result pojo.RechargeOrderAppBack) {
	if result.OrderNo == "" || result.CryptoPayment == nil || strings.TrimSpace(result.CryptoPayment.ExpireTime) == "" {
		return
	}
	expireAt, err := time.Parse(time.RFC3339, result.CryptoPayment.ExpireTime)
	if err != nil {
		log.Printf("[usdt_recharge] parse expire time failed orderNo=%s expireTime=%s err=%v", result.OrderNo, result.CryptoPayment.ExpireTime, err)
		return
	}
	if err = services.EnqueueUsdtRechargeExpireTask(tablePrefix, result.OrderNo, expireAt); err != nil {
		log.Printf("[usdt_recharge] enqueue expire task failed prefix=%s orderNo=%s err=%v", tablePrefix, result.OrderNo, err)
	}
}

// AdminCreateRechargeOrderV2 管理后台为指定TG用户手动拉起v2充值订单
func AdminCreateRechargeOrderV2(ctx *gin.Context) {
	var req pojo.AdminCreateRechargeOrderV2Req
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ErrorBack(ctx, "参数格式错误")
		return
	}
	if req.UserID <= 0 {
		utils.ErrorBack(ctx, "invalid_params")
		return
	}
	if strings.TrimSpace(req.ReturnURL) == "" {
		req.ReturnURL = defaultAdminRechargeReturnURL
	}
	req.ClientIP = ctx.ClientIP()

	db := ctx.MustGet("db").(*gorm.DB)
	hostInfo := ctx.MustGet("hostInfo").(pojo.HostInfo)
	result, err := repository.AdminCreateRechargeOrderV2(db, req.UserID, req.RechargeOrderAppReq, hostInfo.TablePrefix)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

// AdminCreateCryptoRechargeOrder 管理后台为指定TG用户创建虚拟货币充值订单
func AdminCreateCryptoRechargeOrder(ctx *gin.Context) {
	var req pojo.CryptoRechargeOrderReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ErrorBack(ctx, "参数格式错误")
		return
	}
	if req.UserID <= 0 {
		utils.ErrorBack(ctx, "invalid_params")
		return
	}

	db := ctx.MustGet("db").(*gorm.DB)
	hostInfo := ctx.MustGet("hostInfo").(pojo.HostInfo)
	result, err := repository.AdminCreateCryptoRechargeOrder(db, req.UserID, req, hostInfo.TablePrefix)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	enqueueCryptoRechargeExpire(hostInfo.TablePrefix, result)
	utils.SuccessObjBack(ctx, result)
}

// GetAppRechargeOrderHistory godoc
//
//	@Summary		app端充值记录
//	@Tags			充值订单
//	@Accept			json
//	@Produce		json
//	@Param			data body		pojo.AppOrderHistorySearch	true	"分页参数"
//	@Success		200	{object}		pojo.AppOrderHistoryResp
//	@Router			/api/v1/app/rechargeOrder/list [post]
func GetAppRechargeOrderHistory(ctx *gin.Context) {
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

	var search pojo.AppOrderHistorySearch
	search.SetPageDefaults()
	if err := ctx.ShouldBindJSON(&search); err != nil {
		utils.ErrorBack(ctx, "参数格式错误")
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	utils.SuccessObjBack(ctx, repository.GetAppRechargeOrderHistory(db, userID, search))
}

// GetCurrentUserRechargeCount godoc
//
//	@Summary		查询当前用户充值次数
//	@Tags			充值订单
//	@Produce		json
//	@Success		200	{object}	pojo.UserRechargeCountBack
//	@Router			/api/v1/app/recharge/count [get]
func GetCurrentUserRechargeCount(ctx *gin.Context) {
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
	utils.SuccessObjBack(ctx, repository.GetUserRechargeCount(db, userID))
}

func GetCurrentUserPendingRechargeNotifications(ctx *gin.Context) {
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
	result, err := repository.GetCurrentUserPendingRechargeNotifications(db, userID)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

func AckRechargeFrontendNotification(ctx *gin.Context) {
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
	var req pojo.RechargeOrderNotifyAckReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ErrorBack(ctx, "invalid_params")
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	if err := repository.AckRechargeFrontendNotification(db, userID, req.OrderNo); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessBack(ctx, "success")
}

// CheckIsFirstRecharge godoc
//
//	@Summary		检查用户活动参与状态
//	@Tags			充值
//	@Produce		json
//	@Success		200	{object}	map[string]bool	"hasFirst: 是否已参加首充活动; hasTodayFirst: 24h内是否已参加今日首充活动"
//	@Router			/api/v1/app/recharge/isFirst [get]
func CheckIsFirstRecharge(ctx *gin.Context) {
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
	hasFirst, hasTodayFirst := repository.CheckActivityStatus(db, userID)
	utils.SuccessObjBack(ctx, gin.H{
		"hasFirst":      !hasFirst,
		"hasTodayFirst": !hasTodayFirst,
	})
}

// CheckIsFirstRechargeV2 godoc
//
//	@Summary		检查用户v2充值是否实际首充
//	@Tags			充值
//	@Produce		json
//	@Success		200	{object}	map[string]bool	"isFirstRecharge: 是否实际首充; hasFirst: 兼容字段，同isFirstRecharge"
//	@Router			/api/v1/app/recharge/isFirst/v2 [get]
func CheckIsFirstRechargeV2(ctx *gin.Context) {
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
	isFirstRecharge, err := repository.CheckRechargeV2IsFirst(db, userID)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	// 赠送预览：下一次充值的次序 + 各档位比例(%) + 门槛，供前端展示「充X送Y」
	rechargeNumber := repository.RechargeV2NextRechargeNumber(db, userID)
	rates := repository.GetRechargeV2GiftRates(db)
	utils.SuccessObjBack(ctx, gin.H{
		"isFirstRecharge": isFirstRecharge,
		"hasFirst":        isFirstRecharge,
		"rechargeNumber":  rechargeNumber, // 1=首充 2=二充 3=三充 ≥4=后续
		"giftRates":       rates[:],       // [首充,二充,三充,第4次及以后] 单位%
		"giftMinAmount":   repository.RechargeV2GiftMinAmount(),
	})
}

func GetRechargePromotions(ctx *gin.Context) {
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
	utils.SuccessObjBack(ctx, repository.GetRechargePromotions(db, userID, hostInfo.TablePrefix))
}
