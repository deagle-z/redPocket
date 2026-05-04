package repository

import (
	"BaseGoUni/core/pojo"
	"errors"
	"fmt"
	"github.com/jinzhu/copier"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"strings"
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
			_ = copier.Copy(&dbOrder, &req)
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
		return tx.Create(&dbOrder).Error
	})
	if err != nil {
		return result, err
	}
	_ = copier.Copy(&result, &dbOrder)
	return result, nil
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
