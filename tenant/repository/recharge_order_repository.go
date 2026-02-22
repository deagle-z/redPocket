package repository

import (
	"BaseGoUni/core/pojo"
	"errors"
	"github.com/jinzhu/copier"
	"gorm.io/gorm"
)

func GetRechargeOrders(db *gorm.DB, tenantID int64, search pojo.RechargeOrderSearch) (result pojo.RechargeOrderResp) {
	var orders []pojo.RechargeOrder
	query := db.Model(&pojo.RechargeOrder{}).Where("tenant_id = ?", tenantID)
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

func GetRechargeOrderByID(db *gorm.DB, tenantID int64, id int64) (result pojo.RechargeOrderBack, err error) {
	var dbOrder pojo.RechargeOrder
	db.Where("id = ? and tenant_id = ?", id, tenantID).First(&dbOrder)
	if dbOrder.ID == 0 {
		return result, errors.New("数据不存在")
	}
	_ = copier.Copy(&result, &dbOrder)
	return result, nil
}

func SetRechargeOrder(db *gorm.DB, tenantID int64, req pojo.RechargeOrderSet) (result pojo.RechargeOrderBack, err error) {
	req.TenantId = tenantID
	var dbOrder pojo.RechargeOrder
	if req.ID > 0 {
		db.Where("id = ? and tenant_id = ?", req.ID, tenantID).First(&dbOrder)
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

func DelRechargeOrder(db *gorm.DB, tenantID int64, id int64) (result string, err error) {
	var dbOrder pojo.RechargeOrder
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
