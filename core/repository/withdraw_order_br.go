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
	"strings"
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
			oldExtra := dbOrder.Extra
			_ = copier.Copy(&dbOrder, &req)
			if req.Extra == nil {
				dbOrder.Extra = oldExtra
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
	if req.Amount > 0 {
		req.Amount = utils.Truncate2(req.Amount)
	}
	if req.Fee > 0 {
		req.Fee = utils.Truncate2(req.Fee)
	}
	var dbOrder pojo.WithdrawOrderBr
	err = db.Transaction(func(tx *gorm.DB) error {
		_ = copier.Copy(&dbOrder, &req)
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
	if err := ReserveWithdrawLimitForOrder(tx, user, order); err != nil {
		return err
	}
	if err := tx.Model(&pojo.TgUser{}).
		Where("id = ?", user.ID).
		Update("balance", gorm.Expr("balance - ?", order.Amount)).Error; err != nil {
		return err
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
