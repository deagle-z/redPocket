package repository

import (
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/utils"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jinzhu/copier"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"math"
	"strings"
	"time"
)

func GetWithdrawOrderBrs(db *gorm.DB, tenantID int64, search pojo.WithdrawOrderBrSearch) (result pojo.WithdrawOrderBrResp) {
	var orders []pojo.WithdrawOrderBr
	query := db.Model(&pojo.WithdrawOrderBr{}).Where("tenant_id = ?", tenantID)
	if search.UserId > 0 {
		query = query.Where("user_id = ?", search.UserId)
	}
	if userUid := strings.TrimSpace(search.UserUid); userUid != "" {
		query = query.Where("user_id IN (?)", db.Model(&pojo.TgUser{}).
			Select("id").
			Where("tenant_id = ? AND uid = ?", tenantID, userUid))
	}
	if search.Status != nil {
		query = query.Where("status = ?", *search.Status)
	}
	if search.OrderNo != "" {
		query = query.Where("order_no = ?", search.OrderNo)
	}
	if search.MerchantOrderNo != "" {
		query = query.Where("merchant_order_no = ?", search.MerchantOrderNo)
	}
	if search.ProviderPayoutNo != "" {
		query = query.Where("provider_payout_no = ?", search.ProviderPayoutNo)
	}
	if countryCode := pojo.NormalizeWithdrawCountryCode(search.CountryCode); countryCode != "" {
		query = query.Where("country_code = ?", countryCode)
	}
	if search.Channel != "" {
		query = query.Where("channel = ?", search.Channel)
	}
	if search.PayMethod != "" {
		query = query.Where("pay_method = ?", search.PayMethod)
	}
	if search.ReceiverDocumentType != "" {
		query = query.Where("receiver_document_type = ?", search.ReceiverDocumentType)
	}
	if search.ReceiverDocument != "" {
		query = query.Where("receiver_document = ?", search.ReceiverDocument)
	}
	query.Count(&result.Total)
	query = query.Order("id desc").Limit(search.PageSize).Offset(search.PageSize * search.CurrentPage)
	query.Find(&orders)
	for _, order := range orders {
		var temp pojo.WithdrawOrderBrBack
		_ = copier.Copy(&temp, &order)
		result.List = append(result.List, temp)
	}
	fillTenantWithdrawOrderBrUserUIDs(db, tenantID, result.List)
	result.PageSize = search.PageSize
	result.CurrentPage = search.CurrentPage
	return result
}

func fillTenantWithdrawOrderBrUserUIDs(db *gorm.DB, tenantID int64, orders []pojo.WithdrawOrderBrBack) {
	userIDs := make([]int64, 0, len(orders))
	seen := make(map[int64]struct{}, len(orders))
	for _, order := range orders {
		if order.UserId <= 0 {
			continue
		}
		if _, ok := seen[order.UserId]; ok {
			continue
		}
		seen[order.UserId] = struct{}{}
		userIDs = append(userIDs, order.UserId)
	}
	if len(userIDs) == 0 {
		return
	}

	var users []pojo.TgUser
	_ = db.Model(&pojo.TgUser{}).
		Select("id, uid").
		Where("tenant_id = ? AND id IN ?", tenantID, userIDs).
		Find(&users).Error

	uidMap := make(map[int64]string, len(users))
	for _, user := range users {
		uidMap[user.ID] = user.Uid
	}
	for i := range orders {
		orders[i].UserUid = uidMap[orders[i].UserId]
	}
}

func GetWithdrawOrderBrByID(db *gorm.DB, tenantID int64, id int64) (result pojo.WithdrawOrderBrBack, err error) {
	var dbOrder pojo.WithdrawOrderBr
	db.Where("id = ? and tenant_id = ?", id, tenantID).First(&dbOrder)
	if dbOrder.ID == 0 {
		return result, errors.New("数据不存在")
	}
	_ = copier.Copy(&result, &dbOrder)
	return result, nil
}

func SetWithdrawOrderBr(db *gorm.DB, tenantID int64, req pojo.WithdrawOrderBrSet) (result pojo.WithdrawOrderBrBack, err error) {
	req.TenantId = tenantID
	req.NormalizeCountryCodeFromExtra()
	var dbOrder pojo.WithdrawOrderBr
	err = db.Transaction(func(tx *gorm.DB) error {
		if req.ID > 0 {
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
				Where("id = ? and tenant_id = ?", req.ID, tenantID).
				First(&dbOrder).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return errors.New("更新的数据不存在")
				}
				return err
			}
			oldStatus := dbOrder.Status
			mergeTenantWithdrawOrderUpdate(&dbOrder, req)
			ensureTenantWithdrawMerchantOrderNo(&dbOrder)
			fillTenantWithdrawOrderNetAmount(tx, &dbOrder)
			if err := tx.Save(&dbOrder).Error; err != nil {
				return err
			}
			if needTenantWithdrawRefund(oldStatus, dbOrder.Status) {
				if err := refundTenantWithdrawAmount(tx, tenantID, dbOrder); err != nil {
					return err
				}
			}
			return nil
		}

		_ = copier.Copy(&dbOrder, &req)
		ensureTenantWithdrawMerchantOrderNo(&dbOrder)
		fillTenantWithdrawOrderNetAmount(tx, &dbOrder)
		return tx.Create(&dbOrder).Error
	})
	if err != nil {
		return result, err
	}
	_ = copier.Copy(&result, &dbOrder)
	return result, nil
}

func mergeTenantWithdrawOrderUpdate(order *pojo.WithdrawOrderBr, req pojo.WithdrawOrderBrSet) {
	if !req.HasJSONFields() {
		oldID := order.ID
		oldCreatedAt := order.CreatedAt
		oldDeletedAt := order.DeletedAt
		_ = copier.Copy(order, &req)
		order.ID = oldID
		order.CreatedAt = oldCreatedAt
		order.DeletedAt = oldDeletedAt
		return
	}
	if req.HasJSONField("tenantId") {
		order.TenantId = req.TenantId
	}
	if req.HasJSONField("appId") {
		order.AppId = req.AppId
	}
	if req.HasJSONField("userId") {
		order.UserId = req.UserId
	}
	if req.HasJSONField("sourceChannelId") {
		order.SourceChannelID = req.SourceChannelID
	}
	if req.HasJSONField("accountId") {
		order.AccountId = req.AccountId
	}
	if req.HasJSONField("orderNo") {
		order.OrderNo = req.OrderNo
	}
	if req.HasJSONField("merchantOrderNo") {
		order.MerchantOrderNo = req.MerchantOrderNo
	}
	if req.HasJSONField("currency") {
		order.Currency = req.Currency
	}
	if req.HasJSONField("countryCode") {
		order.CountryCode = req.CountryCode
	}
	if req.HasJSONField("amount") {
		order.Amount = req.Amount
	}
	if req.HasJSONField("fee") {
		order.Fee = req.Fee
	}
	if req.HasJSONField("channel") {
		order.Channel = req.Channel
	}
	if req.HasJSONField("payMethod") {
		order.PayMethod = req.PayMethod
	}
	if req.HasJSONField("status") {
		order.Status = req.Status
	}
	if req.HasJSONField("reviewedBy") {
		order.ReviewedBy = req.ReviewedBy
	}
	if req.HasJSONField("reviewedAt") {
		order.ReviewedAt = req.ReviewedAt
	}
	if req.HasJSONField("paidAt") {
		order.PaidAt = req.PaidAt
	}
	if req.HasJSONField("failCode") {
		order.FailCode = req.FailCode
	}
	if req.HasJSONField("failMsg") {
		order.FailMsg = req.FailMsg
	}
	if req.HasJSONField("receiverName") {
		order.ReceiverName = req.ReceiverName
	}
	if req.HasJSONField("receiverDocument") {
		order.ReceiverDocument = req.ReceiverDocument
	}
	if req.HasJSONField("receiverDocumentType") {
		order.ReceiverDocumentType = req.ReceiverDocumentType
	}
	if req.HasJSONField("pixKeyType") {
		order.PixKeyType = req.PixKeyType
	}
	if req.HasJSONField("pixKey") {
		order.PixKey = req.PixKey
	}
	if req.HasJSONField("bankCode") {
		order.BankCode = req.BankCode
	}
	if req.HasJSONField("bankName") {
		order.BankName = req.BankName
	}
	if req.HasJSONField("branchNumber") {
		order.BranchNumber = req.BranchNumber
	}
	if req.HasJSONField("accountNumber") {
		order.AccountNumber = req.AccountNumber
	}
	if req.HasJSONField("accountType") {
		order.AccountType = req.AccountType
	}
	if req.HasJSONField("provider") {
		order.Provider = req.Provider
	}
	if req.HasJSONField("providerPayoutNo") {
		order.ProviderPayoutNo = req.ProviderPayoutNo
	}
	if req.HasJSONField("providerStatus") {
		order.ProviderStatus = req.ProviderStatus
	}
	if req.HasJSONField("notifyTime") {
		order.NotifyTime = req.NotifyTime
	}
	if req.HasJSONField("notifyCount") {
		order.NotifyCount = req.NotifyCount
	}
	if req.HasJSONField("idempotencyKey") {
		order.IdempotencyKey = req.IdempotencyKey
	}
	if req.HasJSONField("riskLevel") {
		order.RiskLevel = req.RiskLevel
	}
	if req.HasJSONField("remark") {
		order.Remark = req.Remark
	}
	if req.HasJSONField("extra") {
		order.Extra = req.Extra
	}
}

func ensureTenantWithdrawMerchantOrderNo(order *pojo.WithdrawOrderBr) {
	if order.MerchantOrderNo != nil && strings.TrimSpace(*order.MerchantOrderNo) != "" {
		value := strings.TrimSpace(*order.MerchantOrderNo)
		order.MerchantOrderNo = &value
		return
	}
	value := fmt.Sprintf("ST%s%s", time.Now().Format("060102150405"), utils.RandomString(8))
	order.MerchantOrderNo = &value
}

func fillTenantWithdrawOrderNetAmount(db *gorm.DB, order *pojo.WithdrawOrderBr) {
	if order == nil {
		return
	}
	order.Amount = utils.Truncate2(order.Amount)
	order.Fee = utils.Truncate2(order.Fee)
	order.NetAmount = calculateTenantWithdrawNetAmount(order.Amount, order.Fee, tenantWithdrawOrderCountryRate(db, *order))
}

func calculateTenantWithdrawNetAmount(amount float64, fee float64, rate float64) float64 {
	baseAmount := utils.Truncate2(amount) - utils.Truncate2(fee)
	if baseAmount <= 0 {
		return 0
	}
	if rate <= 0 {
		rate = 1
	}
	converted := baseAmount * rate
	if converted <= 0 || math.IsNaN(converted) || math.IsInf(converted, 0) {
		return 0
	}
	return math.Ceil(converted)
}

func tenantWithdrawOrderCountryRate(db *gorm.DB, order pojo.WithdrawOrderBr) float64 {
	countryCode := pojo.NormalizeWithdrawCountryCode(order.CountryCode)
	if countryCode == "" {
		countryCode = pojo.NormalizeWithdrawCountryCode(tenantWithdrawFieldValue(tenantWithdrawOrderExtraFields(order.Extra), "countryCode"))
	}
	if db == nil || countryCode == "" {
		return 1
	}

	var country pojo.SysCountry
	if err := db.Select("rate").
		Where("country_code = ? AND status = 1", countryCode).
		First(&country).Error; err != nil || country.Rate <= 0 {
		return 1
	}
	return country.Rate
}

func tenantWithdrawOrderExtraFields(extra *string) map[string]string {
	fields := map[string]string{}
	if extra == nil || strings.TrimSpace(*extra) == "" {
		return fields
	}
	var raw map[string]any
	if err := json.Unmarshal([]byte(*extra), &raw); err != nil {
		return fields
	}
	flattenTenantWithdrawExtraFields(fields, raw)
	if nested, ok := raw["fieldValues"].(map[string]any); ok {
		flattenTenantWithdrawExtraFields(fields, nested)
	}
	return fields
}

func flattenTenantWithdrawExtraFields(fields map[string]string, raw map[string]any) {
	for key, value := range raw {
		switch typed := value.(type) {
		case string:
			fields[key] = strings.TrimSpace(typed)
		case float64, bool:
			fields[key] = strings.TrimSpace(fmt.Sprint(typed))
		}
	}
}

func tenantWithdrawFieldValue(fields map[string]string, aliases ...string) string {
	for _, alias := range aliases {
		if value := strings.TrimSpace(fields[alias]); value != "" {
			return value
		}
	}
	return ""
}

func needTenantWithdrawRefund(oldStatus int, newStatus int) bool {
	oldActive := oldStatus == 0 || oldStatus == 1 || oldStatus == 2 || oldStatus == 3
	newRefund := newStatus == 4 || newStatus == 5 || newStatus == 6
	return oldActive && newRefund
}

func refundTenantWithdrawAmount(tx *gorm.DB, tenantID int64, order pojo.WithdrawOrderBr) error {
	if order.UserId <= 0 || order.Amount <= 0 {
		return nil
	}

	var user pojo.TgUser
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ? and tenant_id = ?", order.UserId, tenantID).
		First(&user).Error; err != nil {
		return err
	}
	if user.ID == 0 {
		return errors.New("用户不存在")
	}

	if err := tx.Model(&pojo.TgUser{}).
		Where("id = ? and tenant_id = ?", user.ID, tenantID).
		Update("balance", gorm.Expr("balance + ?", order.Amount)).Error; err != nil {
		return err
	}

	cashHistory := pojo.CashHistory{
		UserId:          user.ID,
		AwardUni:        fmt.Sprintf("withdraw_refund_%s", order.OrderNo),
		Amount:          order.Amount,
		StartAmount:     user.Balance,
		EndAmount:       user.Balance + order.Amount,
		CashMark:        "提现退回",
		CashDesc:        fmt.Sprintf("提现订单%s退回，返还%.2f", order.OrderNo, order.Amount),
		Type:            pojo.CashHistoryTypeWithdrawRefund,
		IsGift:          0,
		FromUserId:      0,
		SourceChannelID: order.SourceChannelID,
	}
	return tx.Create(&cashHistory).Error
}

func DelWithdrawOrderBr(db *gorm.DB, tenantID int64, id int64) (result string, err error) {
	var dbOrder pojo.WithdrawOrderBr
	db.Where("id = ? and tenant_id = ?", id, tenantID).First(&dbOrder)
	if dbOrder.ID == 0 {
		return result, errors.New("删除的数据不存在")
	}
	err = db.Delete(&dbOrder).Error
	if err != nil {
		return result, err
	}
	return "success", nil
}
