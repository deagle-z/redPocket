package api

import (
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/repository"
	"BaseGoUni/core/services"
	"BaseGoUni/core/utils"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GetLuckyMoneyListAdmin 管理员获取红包列表
// @Summary 管理员获取红包列表
// @Tags 红包管理-管理员
// @Accept json
// @Produce json
// @Param data body pojo.LuckyMoneySearch true "查询条件"
// @Success 200 {object} pojo.LuckyMoneyResp
// @Router /api/v1/admin/lucky/list [post]
func GetLuckyMoneyListAdmin(ctx *gin.Context) {
	var search pojo.LuckyMoneySearch
	search.SetPageDefaults()
	if err := ctx.ShouldBindJSON(&search); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}

	db := ctx.MustGet("db").(*gorm.DB)

	result := repository.GetLuckyMoneyList(db, search)
	utils.SuccessObjBack(ctx, result)
}

// GetLuckyHistoryListAdmin 管理员获取领取历史记录
// @Summary 管理员获取领取历史记录
// @Tags 红包管理-管理员
// @Accept json
// @Produce json
// @Param data body pojo.LuckyHistorySearch true "查询条件"
// @Success 200 {object} pojo.LuckyHistoryResp
// @Router /api/v1/admin/lucky/history [post]
func GetLuckyHistoryListAdmin(ctx *gin.Context) {
	var search pojo.LuckyHistorySearch
	search.SetPageDefaults()
	if err := ctx.ShouldBindJSON(&search); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}

	db := ctx.MustGet("db").(*gorm.DB)

	result := repository.GetLuckyHistoryList(db, search)
	utils.SuccessObjBack(ctx, result)
}

// GetLuckyHistoryUserFlowListAdmin 管理员获取按用户汇总的流水金额
// @Summary 管理员获取按用户汇总的流水金额
// @Tags 红包管理-管理员
// @Accept json
// @Produce json
// @Param data body pojo.LuckyHistoryUserFlowSearch true "查询条件"
// @Success 200 {object} pojo.LuckyHistoryUserFlowResp
// @Router /api/v1/admin/lucky/historyUserFlow [post]
func GetLuckyHistoryUserFlowListAdmin(ctx *gin.Context) {
	var search pojo.LuckyHistoryUserFlowSearch
	search.SetPageDefaults()
	if err := ctx.ShouldBindJSON(&search); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}

	db := ctx.MustGet("db").(*gorm.DB)
	result := repository.GetLuckyHistoryUserFlowList(db, search)
	utils.SuccessObjBack(ctx, result)
}

// GetLuckyMoneyDetailAdmin 管理员获取红包详情
// @Summary 管理员获取红包详情
// @Tags 红包管理-管理员
// @Accept json
// @Produce json
// @Param id path int true "红包ID"
// @Success 200 {object} pojo.BaseResponse
// @Router /api/v1/admin/lucky/:id [get]
func GetLuckyMoneyDetailAdmin(ctx *gin.Context) {
	luckyIDStr := ctx.Param("id")
	luckyID := int64(0)
	_, err := fmt.Sscanf(luckyIDStr, "%d", &luckyID)
	if err != nil || luckyID <= 0 {
		utils.ErrorBack(ctx, "无效的红包ID")
		return
	}

	db := ctx.MustGet("db").(*gorm.DB)

	luckyMoney, err := repository.GetLuckyMoney(db, luckyID)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}

	historyList, err := repository.GetLuckyHistoryByLuckyId(db, luckyID)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}

	items, err := repository.GetLuckyMoneyItemsByLuckyID(db, luckyID)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	if luckyMoney.RedList == "" && len(items) > 0 {
		redList := make([]float64, 0, len(items))
		for _, item := range items {
			redList = append(redList, item.Amount)
		}
		if redListJSON, marshalErr := json.Marshal(redList); marshalErr == nil {
			luckyMoney.RedList = string(redListJSON)
		}
	}

	result := map[string]interface{}{
		"luckyMoney": luckyMoney,
		"history":    historyList,
		"items":      items,
	}

	utils.SuccessObjBack(ctx, result)
}

// ManualGrabLuckyMoneyAdmin 管理员指定机器人手动抢未领取红包
// @Summary 管理员指定机器人手动抢未领取红包
// @Tags 红包管理-管理员
// @Accept json
// @Produce json
// @Param data body pojo.LuckyMoneyManualGrab true "手动抢红包"
// @Success 200 {object} pojo.BaseResponse
// @Router /api/v1/admin/lucky/manualGrab [post]
func ManualGrabLuckyMoneyAdmin(ctx *gin.Context) {
	var req pojo.LuckyMoneyManualGrab
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}

	db := ctx.MustGet("db").(*gorm.DB)
	hostInfo := ctx.MustGet("hostInfo").(pojo.HostInfo)

	botUser, err := repository.GetTgUserById(db, req.BotUserID)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	if !botUser.IsBot {
		utils.ErrorBack(ctx, "selected_user_not_bot")
		return
	}
	if botUser.Status != 1 {
		utils.ErrorBack(ctx, "bot_user_disabled")
		return
	}

	if err := validateManualGrabItem(db, req.LuckyID, req.SeqNo); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}

	result, err := services.GrabRedPacket(db, req.LuckyID, req.BotUserID, hostInfo.TablePrefix, req.SeqNo, req.OddEvenGuess)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}

	utils.SuccessObjBack(ctx, result)
}

func validateManualGrabItem(db *gorm.DB, luckyID int64, seqNo int) error {
	luckyMoney, err := repository.GetLuckyMoney(db, luckyID)
	if err != nil {
		return err
	}
	if luckyMoney.Status != 1 {
		return errors.New("lucky_finished")
	}
	if seqNo > luckyMoney.Number {
		return errors.New("invalid_lucky_seq_no")
	}

	var item pojo.LuckyMoneyItem
	if err := db.Where("red_packet_id = ? AND seq_no = ?", luckyID, seqNo).First(&item).Error; err != nil {
		return err
	}
	if item.IsGrabbed == 1 {
		return errors.New("lucky_item_already_grabbed")
	}
	return nil
}

// GetCashHistoryListAdmin 管理员获取余额变动记录列表
// @Summary 管理员获取余额变动记录列表
// @Tags 红包管理-管理员
// @Accept json
// @Produce json
// @Param data body pojo.CashHistorySearch true "查询条件"
// @Success 200 {object} pojo.CashHistoryPage
// @Router /api/v1/admin/cashHistory/list [post]
func GetCashHistoryListAdmin(ctx *gin.Context) {
	var search pojo.CashHistorySearch
	search.SetPageDefaults()
	if err := ctx.ShouldBindJSON(&search); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}

	db := ctx.MustGet("db").(*gorm.DB)

	result := repository.GetCashHistoryListAdmin(db, search)
	utils.SuccessObjBack(ctx, result)
}
