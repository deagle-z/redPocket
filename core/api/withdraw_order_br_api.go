package api

import (
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/repository"
	"BaseGoUni/core/utils"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"strconv"
	"strings"
	"time"
)

// GetWithdrawOrderBrs godoc
//
//	@Summary		获取巴西提现订单列表
//	@Tags			巴西提现订单
//	@Accept			json
//	@Produce		json
//	@Param			data body		pojo.WithdrawOrderBrSearch	true	"查询条件"
//	@Success		200	{object}		pojo.WithdrawOrderBrResp
//	@Router			/api/v1/admin/withdrawOrderBr/list [post]
func GetWithdrawOrderBrs(ctx *gin.Context) {
	var search pojo.WithdrawOrderBrSearch
	search.SetPageDefaults()
	if err := ctx.ShouldBindJSON(&search); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result := repository.GetWithdrawOrderBrs(db, search)
	utils.SuccessObjBack(ctx, result)
}

// SetWithdrawOrderBr godoc
//
//	@Summary		创建或更新巴西提现订单
//	@Tags			巴西提现订单
//	@Accept			json
//	@Produce		json
//	@Param			data body		pojo.WithdrawOrderBrSet	true	"提现订单信息"
//	@Success		200	{object}		pojo.WithdrawOrderBrBack
//	@Router			/api/v1/admin/withdrawOrderBr [post]
func SetWithdrawOrderBr(ctx *gin.Context) {
	var req pojo.WithdrawOrderBrSet
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result, err := repository.SetWithdrawOrderBr(db, req)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

// DelWithdrawOrderBr godoc
//
//	@Summary		删除巴西提现订单
//	@Tags			巴西提现订单
//	@Accept			json
//	@Produce		json
//	@Param			id path		int	true	"提现订单ID"
//	@Success		200	{object}		string
//	@Router			/api/v1/admin/withdrawOrderBr/:id [delete]
func DelWithdrawOrderBr(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.ErrorBack(ctx, "参数格式错误")
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result, err := repository.DelWithdrawOrderBr(db, id)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

// GetWithdrawOrderBrById godoc
//
//	@Summary		根据ID获取巴西提现订单
//	@Tags			巴西提现订单
//	@Accept			json
//	@Produce		json
//	@Param			id path		int	true	"提现订单ID"
//	@Success		200	{object}		pojo.WithdrawOrderBrBack
//	@Router			/api/v1/admin/withdrawOrderBr/:id [get]
func GetWithdrawOrderBrById(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.ErrorBack(ctx, "参数格式错误")
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result, err := repository.GetWithdrawOrderBrById(db, id)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

// AppCreateWithdrawOrder App端创建提现订单
func AppCreateWithdrawOrder(ctx *gin.Context) {
	userIDRaw, ok := ctx.Get("userId")
	if !ok {
		utils.UnauthorizedBack(ctx, "token_invalid")
		return
	}
	userID, ok := userIDRaw.(int64)
	if !ok || userID <= 0 {
		utils.UnauthorizedBack(ctx, "token_invalid")
		return
	}

	var req pojo.AppCreateWithdrawOrderReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	req.Amount = utils.Truncate2(req.Amount)
	if req.Amount <= 0 {
		utils.ErrorBack(ctx, "invalid_withdraw_amount")
		return
	}

	db := ctx.MustGet("db").(*gorm.DB)
	var user pojo.TgUser
	if err := db.Where("id = ?", userID).First(&user).Error; err != nil || user.ID == 0 {
		utils.ErrorBack(ctx, "user_not_found")
		return
	}

	countryCode := strings.TrimSpace(req.CountryCode)
	if countryCode == "" && user.Country != nil {
		countryCode = strings.TrimSpace(*user.Country)
	}
	if countryCode == "" {
		utils.ErrorBack(ctx, "country_required")
		return
	}

	var country pojo.SysCountry
	if err := db.Where("country_code = ? AND status = 1", countryCode).First(&country).Error; err != nil || country.ID == 0 {
		utils.ErrorBack(ctx, "country_not_available")
		return
	}

	var account *pojo.SysUserWithdrawAccount
	if req.AccountID != nil && *req.AccountID > 0 {
		var dbAccount pojo.SysUserWithdrawAccount
		if err := db.Where("id = ? AND user_id = ? AND status = 1", *req.AccountID, userID).First(&dbAccount).Error; err != nil || dbAccount.ID == 0 {
			utils.ErrorBack(ctx, "account_not_found")
			return
		}
		if !strings.EqualFold(dbAccount.CountryCode, countryCode) {
			utils.ErrorBack(ctx, "account_country_mismatch")
			return
		}
		account = &dbAccount
	}

	orderNo := buildWithdrawOrderNo()
	extraBytes, _ := json.Marshal(map[string]any{
		"countryCode": countryCode,
		"fieldValues": req.FieldValues,
	})
	extra := string(extraBytes)
	accountID := ""
	if account != nil {
		accountID = strconv.FormatInt(account.ID, 10)
	}
	orderReq := pojo.WithdrawOrderBrSet{
		TenantId:        user.TenantId,
		UserId:          user.ID,
		AccountId:       optionalString(accountID),
		OrderNo:         orderNo,
		Currency:        country.CurrencyCode,
		Amount:          req.Amount,
		Fee:             0,
		Channel:         "pix",
		Status:          0,
		Extra:           &extra,
		IdempotencyKey:  optionalString(orderNo),
		SourceChannelID: user.SourceChannelID,
	}
	applyWithdrawReceiverSnapshot(&orderReq, account, req.FieldValues)

	result, err := repository.SetWithdrawOrderBr(db, orderReq)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, pojo.AppCreateWithdrawOrderResp{OrderNo: result.OrderNo})
}

func buildWithdrawOrderNo() string {
	return fmt.Sprintf("WD%s%s", time.Now().Format("20060102150405"), utils.RandomString(6))
}

func optionalString(value string) *string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return &value
}

func applyWithdrawReceiverSnapshot(req *pojo.WithdrawOrderBrSet, account *pojo.SysUserWithdrawAccount, fieldValues map[string]string) {
	values := map[string]string{}
	if account != nil && strings.TrimSpace(account.AccountData) != "" {
		_ = json.Unmarshal([]byte(account.AccountData), &values)
	}
	for k, v := range fieldValues {
		values[k] = v
	}
	req.ReceiverName = firstOptional(values, "receiverName", "name", "fullName", "accountName")
	req.ReceiverDocument = firstOptional(values, "receiverDocument", "document", "cpf", "idNumber")
	req.ReceiverDocumentType = firstOptional(values, "receiverDocumentType", "documentType")
	req.PixKeyType = firstOptional(values, "pixKeyType")
	req.PixKey = firstOptional(values, "pixKey")
	req.BankCode = firstOptional(values, "bankCode")
	req.BankName = firstOptional(values, "bankName")
	req.BranchNumber = firstOptional(values, "branchNumber", "agency")
	req.AccountNumber = firstOptional(values, "accountNumber")
	req.AccountType = firstOptional(values, "accountType")
}

func firstOptional(values map[string]string, keys ...string) *string {
	for _, key := range keys {
		if value := strings.TrimSpace(values[key]); value != "" {
			return &value
		}
	}
	return nil
}
