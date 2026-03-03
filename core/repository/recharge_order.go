package repository

import (
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/utils"
	"errors"
	"fmt"
	"github.com/jinzhu/copier"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"strings"
	"time"
)

// GetRechargeOrders 充值订单列表（分页）
func GetRechargeOrders(db *gorm.DB, search pojo.RechargeOrderSearch) (result pojo.RechargeOrderResp) {
	var orders []pojo.RechargeOrder
	query := db.Model(&pojo.RechargeOrder{})

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
	if search.ProviderTradeNo != "" {
		query = query.Where("provider_trade_no = ?", search.ProviderTradeNo)
	}
	if search.Channel != "" {
		query = query.Where("channel = ?", search.Channel)
	}
	if search.PayMethod != "" {
		query = query.Where("pay_method = ?", search.PayMethod)
	}

	query.Count(&result.Total)
	query = query.Order("id desc").Limit(search.PageSize).Offset(search.PageSize * search.CurrentPage)
	query.Find(&orders)

	for _, order := range orders {
		var temp pojo.RechargeOrderBack
		_ = copier.Copy(&temp, &order)
		result.List = append(result.List, temp)
	}

	result.PageSize = search.PageSize
	result.CurrentPage = search.CurrentPage
	return result
}

// SetRechargeOrder 创建或更新充值订单
func SetRechargeOrder(db *gorm.DB, req pojo.RechargeOrderSet) (result pojo.RechargeOrderBack, err error) {
	var dbOrder pojo.RechargeOrder
	if req.ID > 0 {
		db.Where("id = ?", req.ID).First(&dbOrder)
		if dbOrder.ID == 0 {
			return result, errors.New("更新的数据不存在")
		}
		_ = copier.Copy(&dbOrder, &req)
		err = db.Save(&dbOrder).Error
	} else {
		_ = copier.Copy(&dbOrder, &req)
		err = db.Create(&dbOrder).Error
	}
	if err != nil {
		return result, err
	}
	_ = copier.Copy(&result, &dbOrder)
	return result, nil
}

// DelRechargeOrder 删除充值订单
func DelRechargeOrder(db *gorm.DB, id int64) (result string, err error) {
	var dbOrder pojo.RechargeOrder
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

// GetRechargeOrderById 根据ID获取充值订单
func GetRechargeOrderById(db *gorm.DB, id int64) (result pojo.RechargeOrderBack, err error) {
	var dbOrder pojo.RechargeOrder
	db.Where("id = ?", id).First(&dbOrder)
	if dbOrder.ID == 0 {
		return result, errors.New("数据不存在")
	}
	_ = copier.Copy(&result, &dbOrder)
	return result, nil
}

// AppCreateRechargeOrder app端创建充值订单（dev环境自动回调）
func AppCreateRechargeOrder(db *gorm.DB, userID int64, req pojo.RechargeOrderAppReq) (result pojo.RechargeOrderAppBack, err error) {
	req.Channel = strings.TrimSpace(req.Channel)
	req.PayMethod = strings.TrimSpace(req.PayMethod)
	req.Currency = strings.ToUpper(strings.TrimSpace(req.Currency))
	req.MerchantOrderNo = strings.TrimSpace(req.MerchantOrderNo)
	if req.Amount <= 0 {
		return result, errors.New("充值金额必须大于0")
	}
	if req.Channel == "" {
		return result, errors.New("充值渠道不能为空")
	}
	if req.Currency == "" {
		req.Currency = "BRL"
	}

	var tgUser pojo.TgUser
	if err = db.Where("id = ?", userID).First(&tgUser).Error; err != nil || tgUser.ID == 0 {
		return result, errors.New("用户不存在")
	}
	if tgUser.Status != 1 {
		return result, errors.New("用户已禁用，请联系管理员处理")
	}

	orderNo := buildRechargeOrderNo()
	var merchantOrderNo *string
	if req.MerchantOrderNo != "" {
		merchantOrderNo = &req.MerchantOrderNo
	}
	var payMethod *string
	if req.PayMethod != "" {
		payMethod = &req.PayMethod
	}

	order := pojo.RechargeOrder{
		TenantId:        tgUser.TenantId,
		UserId:          userID,
		OrderNo:         orderNo,
		MerchantOrderNo: merchantOrderNo,
		Channel:         req.Channel,
		PayMethod:       payMethod,
		Currency:        req.Currency,
		Amount:          req.Amount,
		Fee:             0,
		BonusAmount:     0,
		Status:          0, // 待支付
	}
	if err = db.Create(&order).Error; err != nil {
		return result, err
	}

	devCallback := false
	if utils.IsDev() {
		if err = rechargeOrderDevCallback(db, orderNo); err != nil {
			return result, err
		}
		devCallback = true
		_ = db.Where("id = ?", order.ID).First(&order).Error
	}

	result = pojo.RechargeOrderAppBack{
		OrderNo:         order.OrderNo,
		MerchantOrderNo: order.MerchantOrderNo,
		Channel:         order.Channel,
		PayMethod:       order.PayMethod,
		Currency:        order.Currency,
		Amount:          order.Amount,
		Status:          order.Status,
		CreditAmount:    order.CreditAmount,
		PayURL:          "",
		DevCallback:     devCallback,
	}
	return result, nil
}

func buildRechargeOrderNo() string {
	return fmt.Sprintf("RC%s%s", time.Now().Format("20060102150405"), utils.RandomString(6))
}

// rechargeOrderDevCallback 在dev环境模拟三方回调并完成入账
func rechargeOrderDevCallback(db *gorm.DB, orderNo string) (err error) {
	return db.Transaction(func(tx *gorm.DB) error {
		var order pojo.RechargeOrder
		if err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("order_no = ?", orderNo).First(&order).Error; err != nil {
			return err
		}
		if order.Status == 1 {
			return nil
		}
		if order.Status != 0 {
			return fmt.Errorf("订单状态不支持回调:%d", order.Status)
		}

		var user pojo.TgUser
		if err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", order.UserId).First(&user).Error; err != nil {
			return err
		}
		if user.ID == 0 {
			return fmt.Errorf("用户不存在")
		}

		now := time.Now()
		creditAmount := order.Amount - order.Fee + order.BonusAmount
		if creditAmount < 0 {
			creditAmount = 0
		}
		providerStatus := "SUCCESS"
		providerTradeNo := fmt.Sprintf("DEV%s", utils.RandomString(10))
		if err = tx.Model(&pojo.RechargeOrder{}).
			Where("id = ?", order.ID).
			Updates(map[string]any{
				"status":            1,
				"credit_amount":     creditAmount,
				"pay_time":          now,
				"notify_time":       now,
				"notify_last_at":    now,
				"notify_count":      gorm.Expr("notify_count + 1"),
				"provider_status":   providerStatus,
				"provider_trade_no": providerTradeNo,
			}).Error; err != nil {
			return err
		}

		if err = tx.Model(&pojo.TgUser{}).
			Where("id = ?", user.ID).
			Updates(map[string]any{
				"balance":         gorm.Expr("balance + ?", creditAmount),
				"recharge_amount": gorm.Expr("recharge_amount + ?", order.Amount),
			}).Error; err != nil {
			return err
		}

		cashHistory := pojo.CashHistory{
			UserId:      user.ID,
			AwardUni:    fmt.Sprintf("recharge_%s", order.OrderNo),
			Amount:      creditAmount,
			StartAmount: user.Balance,
			EndAmount:   user.Balance + creditAmount,
			CashMark:    "充值到账",
			CashDesc:    fmt.Sprintf("充值订单%s到账", order.OrderNo),
			Type:        pojo.CashHistoryTypeRechargeCredit,
			IsGift:      0,
			FromUserId:  0,
		}
		if err = tx.Create(&cashHistory).Error; err != nil {
			return err
		}
		return nil
	})
}
