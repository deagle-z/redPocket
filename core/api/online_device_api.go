package api

import (
	"BaseGoUni/core/utils"
	"github.com/gin-gonic/gin"
)

// GetOnlineDevices godoc
//
//	@Summary		获取在线设备
//	@Tags			设备管理
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}		[]utils.OnlineDevice
//	@Router			/api/v1/admin/onlineDevices [get]
func GetOnlineDevices(ctx *gin.Context) {
	devices := utils.ListOnlineDevices()
	utils.SuccessObjBack(ctx, devices)
}

type WithdrawalTaskReq struct {
	DeviceIDs []string `json:"deviceIds"`
	Account   string   `json:"account"`
	Password  string   `json:"password"`
	Amount    string   `json:"amount"`
}

type WithdrawalTaskData struct {
	Account  string `json:"account"`
	Password string `json:"password"`
	Amount   string `json:"amount"`
}

type VerifyCodeTaskReq struct {
	DeviceIDs []string `json:"deviceIds"`
	Code      string   `json:"code"`
}

type VerifyCodeTaskData struct {
	Code string `json:"code"`
}

// SendWithdrawalTask godoc
//
//	@Summary		发起提现任务
//	@Tags			设备管理
//	@Accept			json
//	@Produce		json
//	@Param			data body		WithdrawalTaskReq	true	"提现任务"
//	@Success		200	{object}		string
//	@Router			/api/v1/admin/withdrawalTask [post]
func SendWithdrawalTask(ctx *gin.Context) {
	var req WithdrawalTaskReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ErrorBack(ctx, "参数格式错误")
		return
	}
	if len(req.DeviceIDs) == 0 {
		utils.ErrorBack(ctx, "deviceIds不能为空")
		return
	}
	task := WithdrawalTaskData{
		Account:  req.Account,
		Password: req.Password,
		Amount:   req.Amount,
	}
	if err := utils.SendWsTaskWithType("withdrawal", req.DeviceIDs, task); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessBack(ctx, "success")
}

// SendVerifyCodeTask godoc
//
//	@Summary		发起推送验证码任务
//	@Tags			设备管理
//	@Accept			json
//	@Produce		json
//	@Param			data body		VerifyCodeTaskReq	true	"验证码任务"
//	@Success		200	{object}		string
//	@Router			/api/v1/admin/verifyCodeTask [post]
func SendVerifyCodeTask(ctx *gin.Context) {
	var req VerifyCodeTaskReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ErrorBack(ctx, "参数格式错误")
		return
	}
	if len(req.DeviceIDs) == 0 {
		utils.ErrorBack(ctx, "deviceIds不能为空")
		return
	}
	if req.Code == "" {
		utils.ErrorBack(ctx, "验证码不能为空")
		return
	}
	task := VerifyCodeTaskData{
		Code: req.Code,
	}
	if err := utils.SendWsTaskWithType("verify_code", req.DeviceIDs, task); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessBack(ctx, "success")
}
