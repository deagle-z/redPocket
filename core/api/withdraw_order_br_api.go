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

const (
	appWithdrawMinAmount      = 10.0
	appWithdrawFreeDailyCount = int64(3)
	appWithdrawFeeRate        = 0.05
	appWithdrawSourceRebate   = "rebate"
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
	if req.Amount < appWithdrawMinAmount {
		utils.ErrorBack(ctx, "invalid_withdraw_amount")
		return
	}

	db := ctx.MustGet("db").(*gorm.DB)
	var user pojo.TgUser
	if err := db.Where("id = ?", userID).First(&user).Error; err != nil || user.ID == 0 {
		utils.ErrorBack(ctx, "user_not_found")
		return
	}

	countryCode := pojo.NormalizeWithdrawCountryCode(req.CountryCode)
	if countryCode == "" && user.Country != nil {
		countryCode = pojo.NormalizeWithdrawCountryCode(*user.Country)
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
	todayWithdrawCount, err := countTodayAppWithdrawOrders(db, user.ID, time.Now())
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	withdrawFee := calculateAppWithdrawFee(todayWithdrawCount, req.Amount)
	orderReq := pojo.WithdrawOrderBrSet{
		TenantId:        user.TenantId,
		UserId:          user.ID,
		AccountId:       optionalString(accountID),
		OrderNo:         orderNo,
		Currency:        country.CurrencyCode,
		CountryCode:     countryCode,
		Amount:          req.Amount,
		Fee:             withdrawFee,
		Channel:         "pix",
		Status:          0,
		Extra:           &extra,
		IdempotencyKey:  optionalString(orderNo),
		SourceChannelID: user.SourceChannelID,
	}
	applyWithdrawReceiverSnapshot(&orderReq, countryCode, account, req.FieldValues)

	result, err := repository.SetWithdrawOrderBr(db, orderReq)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, pojo.AppCreateWithdrawOrderResp{OrderNo: result.OrderNo, Fee: result.Fee})
}

// AppCreateRebateWithdrawOrder App端创建佣金提现订单，佣金直接提现不需要流水要求
func AppCreateRebateWithdrawOrder(ctx *gin.Context) {
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
	if user.Status != 1 {
		utils.ErrorBack(ctx, "user_disabled_contact_admin")
		return
	}
	if req.Amount > utils.Truncate2(user.RebateAmount) {
		utils.ErrorBack(ctx, "rebate_amount_insufficient")
		return
	}

	countryCode := pojo.NormalizeWithdrawCountryCode(req.CountryCode)
	if countryCode == "" && user.Country != nil {
		countryCode = pojo.NormalizeWithdrawCountryCode(*user.Country)
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
		"countryCode":   countryCode,
		"fieldValues":   req.FieldValues,
		"source":        "rebate",
		"balanceSource": "rebate",
	})
	extra := string(extraBytes)
	accountID := ""
	if account != nil {
		accountID = strconv.FormatInt(account.ID, 10)
	}
	remark := "佣金提现"
	orderReq := pojo.WithdrawOrderBrSet{
		TenantId:        user.TenantId,
		UserId:          user.ID,
		AccountId:       optionalString(accountID),
		OrderNo:         orderNo,
		Currency:        country.CurrencyCode,
		CountryCode:     countryCode,
		Amount:          req.Amount,
		Fee:             0,
		Channel:         "pix",
		Status:          0,
		Remark:          &remark,
		Extra:           &extra,
		IdempotencyKey:  optionalString(orderNo),
		SourceChannelID: user.SourceChannelID,
	}
	applyWithdrawReceiverSnapshot(&orderReq, countryCode, account, req.FieldValues)

	result, err := repository.SetRebateWithdrawOrder(db, orderReq)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}

	var updated pojo.TgUser
	_ = db.Select("rebate_amount").Where("id = ?", user.ID).First(&updated).Error
	utils.SuccessObjBack(ctx, pojo.AppCreateWithdrawOrderResp{
		OrderNo:      result.OrderNo,
		RebateAmount: utils.Truncate2(updated.RebateAmount),
	})
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

func calculateAppWithdrawFee(todayWithdrawCount int64, amount float64) float64 {
	if todayWithdrawCount < appWithdrawFreeDailyCount {
		return 0
	}
	return utils.Truncate2(amount * appWithdrawFeeRate)
}

func countTodayAppWithdrawOrders(db *gorm.DB, userID int64, now time.Time) (int64, error) {
	if db == nil || userID <= 0 {
		return 0, nil
	}
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	end := start.AddDate(0, 0, 1)
	var count int64
	err := db.Model(&pojo.WithdrawOrderBr{}).
		Where("user_id = ? AND created_at >= ? AND created_at < ?", userID, start, end).
		Where(`COALESCE(JSON_UNQUOTE(JSON_EXTRACT(extra, '$.source')), '') <> ?`, appWithdrawSourceRebate).
		Where(`COALESCE(JSON_UNQUOTE(JSON_EXTRACT(extra, '$.balanceSource')), '') <> ?`, appWithdrawSourceRebate).
		Where(`COALESCE(JSON_UNQUOTE(JSON_EXTRACT(extra, '$.withdrawSource')), '') <> ?`, appWithdrawSourceRebate).
		Count(&count).Error
	return count, err
}

func applyWithdrawReceiverSnapshot(req *pojo.WithdrawOrderBrSet, countryCode string, account *pojo.SysUserWithdrawAccount, fieldValues map[string]string) {
	values := map[string]string{}
	if account != nil && strings.TrimSpace(account.AccountData) != "" {
		_ = json.Unmarshal([]byte(account.AccountData), &values)
	}
	for k, v := range fieldValues {
		values[k] = v
	}
	req.ReceiverName = firstOptional(countryCode, values, "receiverName", "accName", "name", "fullName", "accountName")
	req.ReceiverDocument = firstOptional(countryCode, values, "receiverDocument", "identityNo", "document", "cpf", "idNumber")
	req.ReceiverDocumentType = firstOptional(countryCode, values, "receiverDocumentType", "identityType", "documentType")
	req.PixKeyType = firstOptional(countryCode, values, "pixKeyType", "identityType")
	req.PixKey = firstOptional(countryCode, values, "pixKey", "accNo", "accountNumber")
	req.BankCode = firstOptional(countryCode, values, "bankCode", "bnakCode")
	req.BankName = firstOptional(countryCode, values, "bankName")
	req.BranchNumber = firstOptional(countryCode, values, "branchNumber", "agency")
	req.AccountNumber = firstOptional(countryCode, values, "accountNumber")
	req.AccountType = firstOptional(countryCode, values, "accountType")
}

func firstOptional(countryCode string, values map[string]string, keys ...string) *string {
	if value := countryFieldValue(countryCode, values, keys...); value != "" {
		return &value
	}
	return nil
}

func countryFieldValue(countryCode string, values map[string]string, keys ...string) string {
	normalizedValues := make(map[string]string, len(values))
	for key, value := range values {
		normalizedKey := normalizeWithdrawFieldKey(key)
		if normalizedKey == "" || strings.TrimSpace(value) == "" {
			continue
		}
		if _, exists := normalizedValues[normalizedKey]; !exists {
			normalizedValues[normalizedKey] = strings.TrimSpace(value)
		}
	}
	for _, suffix := range countryFieldSuffixes(countryCode) {
		for _, key := range keys {
			normalizedKey := normalizeWithdrawFieldKey(key)
			if normalizedKey == "" {
				continue
			}
			if value := normalizedValues[normalizedKey+suffix]; value != "" {
				return value
			}
		}
	}
	for _, key := range keys {
		if value := normalizedValues[normalizeWithdrawFieldKey(key)]; value != "" {
			return value
		}
	}
	if pojo.NormalizeWithdrawCountryCode(countryCode) != "" {
		return ""
	}
	for _, key := range keys {
		normalizedKey := normalizeWithdrawFieldKey(key)
		if normalizedKey == "" {
			continue
		}
		for fieldKey, value := range normalizedValues {
			if strings.HasPrefix(fieldKey, normalizedKey) {
				return value
			}
		}
	}
	return ""
}

func countryFieldSuffixes(countryCode string) []string {
	switch pojo.NormalizeWithdrawCountryCode(countryCode) {
	case "BR":
		return []string{"brlw", "brl", "br"}
	case "MX":
		return []string{"mxnw", "mxn", "mx"}
	default:
		return nil
	}
}

func normalizeWithdrawFieldKey(key string) string {
	var builder strings.Builder
	for _, r := range strings.ToLower(strings.TrimSpace(key)) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			builder.WriteRune(r)
		}
	}
	return builder.String()
}
