package repository

import (
	"BaseGoUni/core/pay"
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/utils"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jinzhu/copier"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"strings"
	"time"
)

const withdrawOrderSourceRebate = "rebate"

// GetWithdrawOrderBrs 巴西提现订单列表（分页）
func GetWithdrawOrderBrs(db *gorm.DB, search pojo.WithdrawOrderBrSearch) (result pojo.WithdrawOrderBrResp) {
	var orders []pojo.WithdrawOrderBr
	query := db.Model(&pojo.WithdrawOrderBr{})

	if search.TenantId > 0 {
		query = query.Where("tenant_id = ?", search.TenantId)
	}
	if search.UserId > 0 {
		query = query.Where("user_id = ?", search.UserId)
	}
	if userUid := strings.TrimSpace(search.UserUid); userUid != "" {
		userQuery := db.Model(&pojo.TgUser{}).
			Select("id").
			Where("uid = ?", userUid)
		if search.TenantId > 0 {
			userQuery = userQuery.Where("tenant_id = ?", search.TenantId)
		}
		query = query.Where("user_id IN (?)", userQuery)
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
	fillWithdrawOrderBrUserUIDs(db, result.List)

	result.PageSize = search.PageSize
	result.CurrentPage = search.CurrentPage
	return result
}

func fillWithdrawOrderBrUserUIDs(db *gorm.DB, orders []pojo.WithdrawOrderBrBack) {
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
		Where("id IN ?", userIDs).
		Find(&users).Error

	uidMap := make(map[int64]string, len(users))
	for _, user := range users {
		uidMap[user.ID] = user.Uid
	}
	for i := range orders {
		orders[i].UserUid = uidMap[orders[i].UserId]
	}
}

// SetWithdrawOrderBr 创建或更新巴西提现订单
func SetWithdrawOrderBr(db *gorm.DB, req pojo.WithdrawOrderBrSet) (result pojo.WithdrawOrderBrBack, err error) {
	req.NormalizeCountryCodeFromExtra()
	if req.Amount > 0 {
		req.Amount = utils.Truncate2(req.Amount)
	}
	if req.Fee > 0 {
		req.Fee = utils.Truncate2(req.Fee)
	}
	var dbOrder pojo.WithdrawOrderBr
	err = db.Transaction(func(tx *gorm.DB) error {
		if req.ID > 0 {
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", req.ID).First(&dbOrder).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return errors.New("record_not_found_update")
				}
				return err
			}
			oldStatus := dbOrder.Status
			mergeWithdrawOrderUpdate(&dbOrder, req)
			ensureWithdrawMerchantOrderNo(&dbOrder)
			if oldStatus == 0 && dbOrder.Status == 1 {
				if err := submitWithdrawPayout(tx, &dbOrder); err != nil {
					return err
				}
			}
			if err := tx.Save(&dbOrder).Error; err != nil {
				return err
			}
			if needWithdrawRefund(oldStatus, dbOrder.Status) {
				if err := refundWithdrawAmount(tx, dbOrder); err != nil {
					return err
				}
			}
			return nil
		}

		_ = copier.Copy(&dbOrder, &req)
		ensureWithdrawMerchantOrderNo(&dbOrder)
		if dbOrder.SourceChannelID == nil && dbOrder.UserId > 0 {
			sourceChannelID, _, sourceErr := LoadUserSourceChannelSnapshot(tx, dbOrder.UserId)
			if sourceErr != nil {
				return sourceErr
			}
			dbOrder.SourceChannelID = sourceChannelID
		}
		if err := tx.Create(&dbOrder).Error; err != nil {
			return err
		}
		if needWithdrawDeductOnCreate(dbOrder.Status) {
			if err := deductWithdrawAmount(tx, &dbOrder); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return result, err
	}
	_ = copier.Copy(&result, &dbOrder)
	return result, nil
}

// SetRebateWithdrawOrder 创建佣金提现订单，不占用普通提现流水。
func SetRebateWithdrawOrder(db *gorm.DB, req pojo.WithdrawOrderBrSet) (result pojo.WithdrawOrderBrBack, err error) {
	req.NormalizeCountryCodeFromExtra()
	if req.Amount > 0 {
		req.Amount = utils.Truncate2(req.Amount)
	}
	if req.Fee > 0 {
		req.Fee = utils.Truncate2(req.Fee)
	}
	var dbOrder pojo.WithdrawOrderBr
	err = db.Transaction(func(tx *gorm.DB) error {
		_ = copier.Copy(&dbOrder, &req)
		ensureWithdrawMerchantOrderNo(&dbOrder)
		if dbOrder.SourceChannelID == nil && dbOrder.UserId > 0 {
			sourceChannelID, _, sourceErr := LoadUserSourceChannelSnapshot(tx, dbOrder.UserId)
			if sourceErr != nil {
				return sourceErr
			}
			dbOrder.SourceChannelID = sourceChannelID
		}
		if err := tx.Create(&dbOrder).Error; err != nil {
			return err
		}
		if needWithdrawDeductOnCreate(dbOrder.Status) {
			if err := deductRebateWithdrawAmount(tx, &dbOrder); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return result, err
	}
	_ = copier.Copy(&result, &dbOrder)
	return result, nil
}

func mergeWithdrawOrderUpdate(order *pojo.WithdrawOrderBr, req pojo.WithdrawOrderBrSet) {
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

func ensureWithdrawMerchantOrderNo(order *pojo.WithdrawOrderBr) {
	if order.MerchantOrderNo != nil && strings.TrimSpace(*order.MerchantOrderNo) != "" {
		value := strings.TrimSpace(*order.MerchantOrderNo)
		order.MerchantOrderNo = &value
		return
	}
	value := buildWithdrawMerchantOrderNo()
	order.MerchantOrderNo = &value
}

func buildWithdrawMerchantOrderNo() string {
	return fmt.Sprintf("ST%s%s", time.Now().Format("060102150405"), utils.RandomString(8))
}

func submitWithdrawPayout(db *gorm.DB, order *pojo.WithdrawOrderBr) error {
	providerCode := strings.TrimSpace(order.Channel)
	if order.Provider != nil && strings.TrimSpace(*order.Provider) != "" {
		providerCode = strings.TrimSpace(*order.Provider)
	}
	if providerCode == "" {
		return errors.New("withdraw_payout_channel_required")
	}
	provider := pay.GetPayout(providerCode)
	if provider == nil {
		return fmt.Errorf("withdraw_payout_channel_not_supported: %s", providerCode)
	}

	ensureWithdrawMerchantOrderNo(order)
	merchantOrderNo := strings.TrimSpace(*order.MerchantOrderNo)
	extraFields := withdrawOrderExtraFields(order.Extra)
	countryCode := pojo.NormalizeWithdrawCountryCode(order.CountryCode)
	if countryCode == "" {
		countryCode = pojo.NormalizeWithdrawCountryCode(withdrawFieldValue(extraFields, "countryCode"))
	}
	var user pojo.TgUser
	if order.UserId > 0 {
		_ = db.Select("uid, username, first_name, email, phone").
			Where("id = ?", order.UserId).
			First(&user).Error
	}
	amount := order.NetAmount
	if amount <= 0 {
		amount = order.Amount - order.Fee
	}
	amount = utils.Truncate2(amount)
	if amount <= 0 {
		return errors.New("withdraw_payout_amount_invalid")
	}
	providerAmount := amount
	if countryCode != "" {
		var country pojo.SysCountry
		if err := db.Select("rate").Where("country_code = ?", countryCode).First(&country).Error; err == nil && country.Rate > 0 {
			providerAmount = ceilProviderAmount(amount * country.Rate)
		}
	}

	accNo := firstWithdrawValue(
		ptrValue(order.PixKey),
		ptrValue(order.AccountNumber),
		withdrawCountryFieldValue(extraFields, countryCode, "accNo", "accountNumber", "pixKey", "cardNo", "clabe"),
	)
	bankCode := firstWithdrawValue(
		ptrValue(order.BankCode),
		withdrawCountryFieldValue(extraFields, countryCode, "bankCode", "bnakCode", "bank"),
	)
	req := pay.PayoutRequest{
		OrderNo:        merchantOrderNo,
		Amount:         amount,
		ProviderAmount: providerAmount,
		Currency:       strings.TrimSpace(order.Currency),
		AccName: firstWithdrawValue(
			ptrValue(order.ReceiverName),
			withdrawCountryFieldValue(extraFields, countryCode, "accName", "receiverName", "name", "fullName", "accountName"),
			ptrValue(user.FirstName),
			ptrValue(user.Username),
			user.Uid,
		),
		AccNo:        accNo,
		BankCode:     bankCode,
		IdentityType: firstWithdrawValue(ptrValue(order.PixKeyType), ptrValue(order.ReceiverDocumentType), withdrawCountryFieldValue(extraFields, countryCode, "identityType", "documentType", "pixKeyType")),
		IdentityNo:   firstWithdrawValue(ptrValue(order.ReceiverDocument), withdrawCountryFieldValue(extraFields, countryCode, "identityNo", "document", "cpf", "idNumber")),
		BusiCode:     withdrawCountryFieldValue(extraFields, countryCode, "busiCode"),
		Email:        firstWithdrawValue(withdrawCountryFieldValue(extraFields, countryCode, "email"), user.Email),
		Phone:        firstWithdrawValue(withdrawCountryFieldValue(extraFields, countryCode, "phone", "mobile", "mobilePhone"), ptrValue(user.Phone), accNo),
		ExtraFields:  extraFields,
	}
	if req.AccName == "" {
		return errors.New("withdraw_payout_acc_name_required")
	}
	if req.AccNo == "" {
		return errors.New("withdraw_payout_acc_no_required")
	}
	if countryCode == "MX" && req.BankCode == "" {
		return errors.New("withdraw_payout_bank_code_required")
	}
	if req.Email == "" {
		return errors.New("withdraw_payout_email_required")
	}
	if req.Phone == "" {
		return errors.New("withdraw_payout_phone_required")
	}

	resp, err := provider.CreatePayoutOrder(req)
	if err != nil {
		return err
	}
	order.Status = 2
	if resp.ProviderOrderNo != "" {
		order.ProviderPayoutNo = &resp.ProviderOrderNo
	}
	providerStatus := fmt.Sprintf("%d", resp.Status)
	order.ProviderStatus = &providerStatus
	return nil
}

func withdrawOrderExtraFields(extra *string) map[string]string {
	fields := map[string]string{}
	if extra == nil || strings.TrimSpace(*extra) == "" {
		return fields
	}
	var raw map[string]any
	if err := json.Unmarshal([]byte(*extra), &raw); err != nil {
		return fields
	}
	flattenWithdrawExtraFields(fields, raw)
	if nested, ok := raw["fieldValues"].(map[string]any); ok {
		flattenWithdrawExtraFields(fields, nested)
	}
	return fields
}

func flattenWithdrawExtraFields(fields map[string]string, values map[string]any) {
	for key, value := range values {
		switch typed := value.(type) {
		case string:
			if strings.TrimSpace(typed) != "" {
				fields[key] = strings.TrimSpace(typed)
			}
		case float64:
			fields[key] = fmt.Sprintf("%.0f", typed)
		}
	}
}

func withdrawFieldValue(fields map[string]string, aliases ...string) string {
	for _, alias := range aliases {
		if value := strings.TrimSpace(fields[alias]); value != "" {
			return value
		}
	}

	normalized := normalizeWithdrawFields(fields)
	for _, alias := range aliases {
		if value := normalized[normalizeWithdrawFieldKey(alias)]; value != "" {
			return value
		}
	}
	for _, alias := range aliases {
		normalizedAlias := normalizeWithdrawFieldKey(alias)
		if normalizedAlias == "" {
			continue
		}
		for key, value := range normalized {
			if strings.HasPrefix(key, normalizedAlias) {
				return value
			}
		}
	}
	return ""
}

func withdrawCountryFieldValue(fields map[string]string, countryCode string, aliases ...string) string {
	normalized := normalizeWithdrawFields(fields)
	suffixes := withdrawCountryFieldSuffixes(countryCode)
	for _, suffix := range suffixes {
		for _, alias := range aliases {
			normalizedAlias := normalizeWithdrawFieldKey(alias)
			if normalizedAlias == "" {
				continue
			}
			if value := normalized[normalizedAlias+suffix]; value != "" {
				return value
			}
		}
	}
	for _, alias := range aliases {
		if value := normalized[normalizeWithdrawFieldKey(alias)]; value != "" {
			return value
		}
	}
	if pojo.NormalizeWithdrawCountryCode(countryCode) != "" {
		return ""
	}
	for _, alias := range aliases {
		normalizedAlias := normalizeWithdrawFieldKey(alias)
		if normalizedAlias == "" {
			continue
		}
		for key, value := range normalized {
			if strings.HasPrefix(key, normalizedAlias) {
				return value
			}
		}
	}
	return ""
}

func withdrawCountryFieldSuffixes(countryCode string) []string {
	switch pojo.NormalizeWithdrawCountryCode(countryCode) {
	case "BR":
		return []string{"brlw", "brl", "br"}
	case "MX":
		return []string{"mxnw", "mxn", "mx"}
	default:
		return nil
	}
}

func normalizeWithdrawFields(fields map[string]string) map[string]string {
	normalized := make(map[string]string, len(fields))
	for key, value := range fields {
		normalizedKey := normalizeWithdrawFieldKey(key)
		if normalizedKey == "" || strings.TrimSpace(value) == "" {
			continue
		}
		if _, exists := normalized[normalizedKey]; !exists {
			normalized[normalizedKey] = strings.TrimSpace(value)
		}
	}
	return normalized
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

func firstWithdrawValue(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func ptrValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

// DelWithdrawOrderBr 删除巴西提现订单
func DelWithdrawOrderBr(db *gorm.DB, id int64) (result string, err error) {
	var dbOrder pojo.WithdrawOrderBr
	db.Where("id = ?", id).First(&dbOrder)
	if dbOrder.ID == 0 {
		return result, errors.New("record_not_found_delete")
	}
	err = db.Delete(&dbOrder).Error
	if err != nil {
		return result, err
	}
	return "success", nil
}

// GetWithdrawOrderBrById 根据ID获取巴西提现订单
func GetWithdrawOrderBrById(db *gorm.DB, id int64) (result pojo.WithdrawOrderBrBack, err error) {
	var dbOrder pojo.WithdrawOrderBr
	db.Where("id = ?", id).First(&dbOrder)
	if dbOrder.ID == 0 {
		return result, errors.New("record_not_found")
	}
	_ = copier.Copy(&result, &dbOrder)
	return result, nil
}

type WithdrawPayoutCallback struct {
	LocalOrderNo     string
	ProviderPayoutNo string
	ProviderStatus   string
	ProviderAmount   float64
	ResultCode       string
	ResultMsg        string
	ProviderPayTime  string
	Success          bool
	Failed           bool
}

func ProcessWithdrawOrderPayoutCallback(db *gorm.DB, cb WithdrawPayoutCallback) error {
	localOrderNo := strings.TrimSpace(cb.LocalOrderNo)
	if localOrderNo == "" {
		return errors.New("withdraw_order_no_empty")
	}
	now := time.Now()

	return db.Transaction(func(tx *gorm.DB) error {
		var order pojo.WithdrawOrderBr
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("order_no = ? OR merchant_order_no = ?", localOrderNo, localOrderNo).
			First(&order).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("withdraw_order_not_found")
			}
			return err
		}

		oldStatus := order.Status
		updates := map[string]any{
			"notify_time":  now,
			"notify_count": gorm.Expr("notify_count + 1"),
		}
		if providerStatus := strings.TrimSpace(cb.ProviderStatus); providerStatus != "" {
			updates["provider_status"] = providerStatus
		}
		if providerPayoutNo := strings.TrimSpace(cb.ProviderPayoutNo); providerPayoutNo != "" {
			updates["provider_payout_no"] = providerPayoutNo
		}

		switch {
		case cb.Success:
			if oldStatus == 4 || oldStatus == 5 || oldStatus == 6 {
				return errors.New("withdraw_order_status_invalid_success")
			}
			if oldStatus != 3 {
				updates["status"] = 3
			}
			paidAt := now
			if parsedPayTime, ok := parseProviderPayTime(cb.ProviderPayTime); ok {
				paidAt = parsedPayTime
			}
			updates["paid_at"] = paidAt
			updates["fail_code"] = nil
			updates["fail_msg"] = nil
		case cb.Failed:
			failCode := strings.TrimSpace(cb.ResultCode)
			failMsg := strings.TrimSpace(cb.ResultMsg)
			if failCode != "" {
				updates["fail_code"] = failCode
			}
			if failMsg != "" {
				updates["fail_msg"] = failMsg
			}
			if oldStatus != 4 && oldStatus != 5 && oldStatus != 6 {
				updates["status"] = 4
				order.Status = 4
			}
		}

		if err := tx.Model(&pojo.WithdrawOrderBr{}).Where("id = ?", order.ID).Updates(updates).Error; err != nil {
			return err
		}
		if cb.Failed && needWithdrawRefund(oldStatus, 4) {
			return refundWithdrawAmount(tx, order)
		}
		return nil
	})
}

func parseProviderPayTime(payTime string) (time.Time, bool) {
	payTime = strings.TrimSpace(payTime)
	if payTime == "" {
		return time.Time{}, false
	}
	for _, layout := range []string{"2006-01-02 15:04:05", time.RFC3339} {
		if parsed, err := time.ParseInLocation(layout, payTime, time.Local); err == nil {
			return parsed, true
		}
	}
	return time.Time{}, false
}

func needWithdrawDeductOnCreate(status int) bool {
	// 0待审核 1待打款 2打款中 3成功：创建单据时需要先冻结/扣减余额
	return status == 0 || status == 1 || status == 2 || status == 3
}

func needWithdrawRefund(oldStatus int, newStatus int) bool {
	// 活跃状态 -> 失败/取消/退回，执行退款
	oldActive := oldStatus == 0 || oldStatus == 1 || oldStatus == 2 || oldStatus == 3
	newRefund := newStatus == 4 || newStatus == 5 || newStatus == 6
	return oldActive && newRefund
}

func deductWithdrawAmount(tx *gorm.DB, order *pojo.WithdrawOrderBr) error {
	if order == nil || order.UserId <= 0 || order.Amount <= 0 {
		return nil
	}
	if isRebateWithdrawOrder(*order) {
		return deductRebateWithdrawAmount(tx, order)
	}
	var user pojo.TgUser
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", order.UserId).First(&user).Error; err != nil {
		return err
	}
	if user.ID == 0 {
		return errors.New("user_not_found")
	}
	if user.Balance < order.Amount {
		return errors.New("user_balance_insufficient")
	}
	activeCycle, err := GetActiveWithdrawActivityCycle(tx, user.ID)
	if err != nil {
		return err
	}
	hasActiveCycle := activeCycle.ID > 0
	activityCycleLowBalance := CanBypassWithdrawActivityCycleByBalance(user.Balance, activeCycle)
	if err := ReserveWithdrawLimitForOrder(tx, user, order); err != nil {
		return err
	}
	if err := tx.Model(&pojo.TgUser{}).
		Where("id = ?", user.ID).
		Update("balance", gorm.Expr("balance - ?", order.Amount)).Error; err != nil {
		return err
	}
	remainingBalance := utils.Truncate2(user.Balance - order.Amount)
	if hasActiveCycle {
		endReason := pojo.WithdrawActivityCycleEndReasonWithdraw
		restrictRemaining := true
		if activityCycleLowBalance {
			endReason = pojo.WithdrawActivityCycleEndReasonBalanceBelowLimit
			restrictRemaining = false
		}
		if err := EndWithdrawActivityCycle(tx, user.ID, endReason); err != nil {
			return err
		}
		if err := ResetUserWithdrawLimitAfterActivityEnd(tx, user.ID, remainingBalance, restrictRemaining); err != nil {
			return err
		}
	}
	cashHistory := pojo.CashHistory{
		UserId:          user.ID,
		AwardUni:        fmt.Sprintf("withdraw_apply_%s", order.OrderNo),
		Amount:          -order.Amount,
		StartAmount:     user.Balance,
		EndAmount:       utils.Truncate2(user.Balance - order.Amount),
		CashMark:        "提现申请",
		CashDesc:        fmt.Sprintf("提现申请%s，冻结/扣减%.2f", order.OrderNo, order.Amount),
		Type:            pojo.CashHistoryTypeWithdrawApply,
		IsGift:          0,
		FromUserId:      0,
		SourceChannelID: order.SourceChannelID,
	}
	return tx.Create(&cashHistory).Error
}

func deductRebateWithdrawAmount(tx *gorm.DB, order *pojo.WithdrawOrderBr) error {
	if order == nil || order.UserId <= 0 || order.Amount <= 0 {
		return nil
	}
	var user pojo.TgUser
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", order.UserId).First(&user).Error; err != nil {
		return err
	}
	if user.ID == 0 {
		return errors.New("user_not_found")
	}
	if user.Status != 1 {
		return errors.New("user_disabled_contact_admin")
	}
	if user.RebateAmount < order.Amount {
		return errors.New("rebate_amount_insufficient")
	}
	newRebateAmount := utils.Truncate2(user.RebateAmount - order.Amount)
	if err := tx.Model(&pojo.TgUser{}).
		Where("id = ?", user.ID).
		Update("rebate_amount", gorm.Expr("rebate_amount - ?", order.Amount)).Error; err != nil {
		return err
	}
	cashHistory := pojo.CashHistory{
		UserId:          user.ID,
		AwardUni:        fmt.Sprintf("rebate_withdraw_apply_%s", order.OrderNo),
		Amount:          -order.Amount,
		StartAmount:     user.RebateAmount,
		EndAmount:       newRebateAmount,
		CashMark:        "佣金提现申请",
		CashDesc:        fmt.Sprintf("佣金提现申请%s，冻结/扣减%.2f", order.OrderNo, order.Amount),
		Type:            pojo.CashHistoryTypeWithdrawApply,
		IsGift:          0,
		FromUserId:      0,
		SourceChannelID: order.SourceChannelID,
	}
	return tx.Create(&cashHistory).Error
}

func refundWithdrawAmount(tx *gorm.DB, order pojo.WithdrawOrderBr) error {
	if order.UserId <= 0 || order.Amount <= 0 {
		return nil
	}
	var user pojo.TgUser
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", order.UserId).First(&user).Error; err != nil {
		return err
	}
	if user.ID == 0 {
		return errors.New("user_not_found")
	}
	if isRebateWithdrawOrder(order) {
		if err := tx.Model(&pojo.TgUser{}).
			Where("id = ?", user.ID).
			Update("rebate_amount", gorm.Expr("rebate_amount + ?", order.Amount)).Error; err != nil {
			return err
		}
		cashHistory := pojo.CashHistory{
			UserId:          user.ID,
			AwardUni:        fmt.Sprintf("rebate_withdraw_refund_%s", order.OrderNo),
			Amount:          order.Amount,
			StartAmount:     user.RebateAmount,
			EndAmount:       utils.Truncate2(user.RebateAmount + order.Amount),
			CashMark:        "佣金提现退回",
			CashDesc:        fmt.Sprintf("佣金提现订单%s失败/取消/退回，返还%.2f", order.OrderNo, order.Amount),
			Type:            pojo.CashHistoryTypeWithdrawRefund,
			IsGift:          0,
			FromUserId:      0,
			SourceChannelID: order.SourceChannelID,
		}
		return tx.Create(&cashHistory).Error
	}
	if err := RefundWithdrawLimitForOrder(tx, user, order); err != nil {
		return err
	}
	if err := tx.Model(&pojo.TgUser{}).
		Where("id = ?", user.ID).
		Update("balance", gorm.Expr("balance + ?", order.Amount)).Error; err != nil {
		return err
	}
	cashHistory := pojo.CashHistory{
		UserId:          user.ID,
		AwardUni:        fmt.Sprintf("withdraw_refund_%s", order.OrderNo),
		Amount:          order.Amount,
		StartAmount:     user.Balance,
		EndAmount:       utils.Truncate2(user.Balance + order.Amount),
		CashMark:        "提现退回",
		CashDesc:        fmt.Sprintf("提现订单%s失败/取消/退回，返还%.2f", order.OrderNo, order.Amount),
		Type:            pojo.CashHistoryTypeWithdrawRefund,
		IsGift:          0,
		FromUserId:      0,
		SourceChannelID: order.SourceChannelID,
	}
	return tx.Create(&cashHistory).Error
}

func isRebateWithdrawOrder(order pojo.WithdrawOrderBr) bool {
	if order.Extra == nil || strings.TrimSpace(*order.Extra) == "" {
		return false
	}
	var extra map[string]any
	if err := json.Unmarshal([]byte(*order.Extra), &extra); err != nil {
		return false
	}
	for _, key := range []string{"source", "balanceSource", "withdrawSource"} {
		if strings.EqualFold(strings.TrimSpace(fmt.Sprint(extra[key])), withdrawOrderSourceRebate) {
			return true
		}
	}
	return false
}
