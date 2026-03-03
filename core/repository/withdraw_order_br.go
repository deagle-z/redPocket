package repository

import (
	"BaseGoUni/core/pojo"
	"errors"
	"fmt"
	"github.com/jinzhu/copier"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

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

	result.PageSize = search.PageSize
	result.CurrentPage = search.CurrentPage
	return result
}

// SetWithdrawOrderBr 创建或更新巴西提现订单
func SetWithdrawOrderBr(db *gorm.DB, req pojo.WithdrawOrderBrSet) (result pojo.WithdrawOrderBrBack, err error) {
	var dbOrder pojo.WithdrawOrderBr
	err = db.Transaction(func(tx *gorm.DB) error {
		if req.ID > 0 {
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", req.ID).First(&dbOrder).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return errors.New("更新的数据不存在")
				}
				return err
			}
			oldStatus := dbOrder.Status
			_ = copier.Copy(&dbOrder, &req)
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
		if err := tx.Create(&dbOrder).Error; err != nil {
			return err
		}
		if needWithdrawDeductOnCreate(dbOrder.Status) {
			if err := deductWithdrawAmount(tx, dbOrder); err != nil {
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
		return result, errors.New("删除的数据不存在")
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
		return result, errors.New("数据不存在")
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

func deductWithdrawAmount(tx *gorm.DB, order pojo.WithdrawOrderBr) error {
	if order.UserId <= 0 || order.Amount <= 0 {
		return nil
	}
	var user pojo.TgUser
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", order.UserId).First(&user).Error; err != nil {
		return err
	}
	if user.ID == 0 {
		return errors.New("用户不存在")
	}
	if user.Balance < order.Amount {
		return errors.New("用户余额不足")
	}
	if err := tx.Model(&pojo.TgUser{}).
		Where("id = ?", user.ID).
		Update("balance", gorm.Expr("balance - ?", order.Amount)).Error; err != nil {
		return err
	}
	cashHistory := pojo.CashHistory{
		UserId:      user.ID,
		AwardUni:    fmt.Sprintf("withdraw_apply_%s", order.OrderNo),
		Amount:      -order.Amount,
		StartAmount: user.Balance,
		EndAmount:   user.Balance - order.Amount,
		CashMark:    "提现申请",
		CashDesc:    fmt.Sprintf("提现申请%s，冻结/扣减%.2f", order.OrderNo, order.Amount),
		Type:        pojo.CashHistoryTypeWithdrawApply,
		IsGift:      0,
		FromUserId:  0,
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
		return errors.New("用户不存在")
	}
	if err := tx.Model(&pojo.TgUser{}).
		Where("id = ?", user.ID).
		Update("balance", gorm.Expr("balance + ?", order.Amount)).Error; err != nil {
		return err
	}
	cashHistory := pojo.CashHistory{
		UserId:      user.ID,
		AwardUni:    fmt.Sprintf("withdraw_refund_%s", order.OrderNo),
		Amount:      order.Amount,
		StartAmount: user.Balance,
		EndAmount:   user.Balance + order.Amount,
		CashMark:    "提现退回",
		CashDesc:    fmt.Sprintf("提现订单%s失败/取消/退回，返还%.2f", order.OrderNo, order.Amount),
		Type:        pojo.CashHistoryTypeWithdrawRefund,
		IsGift:      0,
		FromUserId:  0,
	}
	return tx.Create(&cashHistory).Error
}
