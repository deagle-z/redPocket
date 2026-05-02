package repository

import (
	"BaseGoUni/core/pay"
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/utils"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hibiken/asynq"
	"github.com/jinzhu/copier"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"log"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

// rechargeFieldDef 充值字段配置（对应 sys_country.recharge_fields JSON 元素）
type rechargeFieldDef struct {
	FieldKey   string  `json:"fieldKey"`
	FieldLabel string  `json:"fieldLabel"`
	IsRequired int     `json:"isRequired"`
	RegexRule  *string `json:"regexRule"`
	ErrorTips  *string `json:"errorTips"`
}

type firstRechargeGiftConfig struct {
	Rate      float64
	Ratios    []int
	RatioBase int
}

type firstRechargeGiftInstallment struct {
	Index      int
	Ratio      int
	GiftAmount float64
	ExecuteAt  time.Time
}

var (
	rechargeGiftAsynqOnce   sync.Once
	rechargeGiftAsynqClient *asynq.Client
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
	if req.Amount > 0 {
		req.Amount = utils.Truncate2(req.Amount)
	}
	if req.Fee > 0 {
		req.Fee = utils.Truncate2(req.Fee)
	}
	if req.CreditAmount != nil {
		creditAmount := utils.Truncate2(*req.CreditAmount)
		req.CreditAmount = &creditAmount
	}
	if req.BonusAmount > 0 {
		req.BonusAmount = utils.Truncate2(req.BonusAmount)
	}
	var dbOrder pojo.RechargeOrder
	if req.ID > 0 {
		db.Where("id = ?", req.ID).First(&dbOrder)
		if dbOrder.ID == 0 {
			return result, errors.New("record_not_found_update")
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
		return result, errors.New("record_not_found_delete")
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
		return result, errors.New("record_not_found")
	}
	_ = copier.Copy(&result, &dbOrder)
	return result, nil
}

// AppCreateRechargeOrder app端创建充值订单（dev环境自动回调）
func AppCreateRechargeOrder(db *gorm.DB, userID int64, req pojo.RechargeOrderAppReq, tablePrefix string) (result pojo.RechargeOrderAppBack, err error) {
	req.Channel = strings.TrimSpace(req.Channel)
	req.PayMethod = strings.TrimSpace(req.PayMethod)
	req.Currency = strings.ToUpper(strings.TrimSpace(req.Currency))
	req.MerchantOrderNo = strings.TrimSpace(req.MerchantOrderNo)
	if req.Amount <= 0 {
		return result, errors.New("recharge_amount_positive")
	}
	req.Amount = utils.Truncate2(req.Amount)
	if req.Channel == "" {
		return result, errors.New("recharge_channel_required")
	}
	if req.Currency == "" {
		req.Currency = "BRL"
	}
	req.CountryCode = strings.TrimSpace(req.CountryCode)

	// 若提供了 countryCode，根据国家字段配置校验 extraFields
	if req.CountryCode != "" {
		var country pojo.SysCountry
		db.Where("country_code = ? AND status = 1", req.CountryCode).First(&country)
		if country.ID == 0 {
			return result, errors.New("country_not_available")
		}
		if country.RechargeFields != nil && *country.RechargeFields != "" {
			var fieldDefs []rechargeFieldDef
			if jsonErr := json.Unmarshal([]byte(*country.RechargeFields), &fieldDefs); jsonErr == nil {
				for _, fd := range fieldDefs {
					val := strings.TrimSpace(req.ExtraFields[fd.FieldKey])
					tip := ""
					if fd.ErrorTips != nil && *fd.ErrorTips != "" {
						tip = *fd.ErrorTips
					}

					// 必填校验
					if fd.IsRequired == 1 && val == "" {
						if tip == "" {
							tip = fd.FieldLabel + " 不能为空"
						}
						return result, errors.New(tip)
					}

					// 正则校验（有值才校验）
					if val != "" && fd.RegexRule != nil && *fd.RegexRule != "" {
						matched, regErr := regexp.MatchString(*fd.RegexRule, val)
						if regErr != nil || !matched {
							if tip == "" {
								tip = fd.FieldLabel + " 格式不正确"
							}
							return result, errors.New(tip)
						}
					}
				}
			}
		}
	}

	var tgUser pojo.TgUser
	if err = db.Where("id = ?", userID).First(&tgUser).Error; err != nil || tgUser.ID == 0 {
		log.Printf("[AppCreateRechargeOrder] 用户不存在 userID=%d err=%v", userID, err)
		return result, errors.New("user_not_found")
	}
	if tgUser.Status != 1 {
		log.Printf("[AppCreateRechargeOrder] 用户已禁用 userID=%d status=%d", userID, tgUser.Status)
		return result, errors.New("user_disabled_contact_admin")
	}

	// 延迟写库策略：先解析渠道、调用三方，成功后再落库，失败不留脏数据
	provider, providerErr := pay.MustGet(req.Channel)
	if providerErr != nil {
		log.Printf("[AppCreateRechargeOrder] 渠道未注册 userID=%d channel=%s err=%v", userID, req.Channel, providerErr)
		return result, providerErr
	}

	orderNo := buildRechargeOrderNo()
	payResp, err := provider.CreateOrder(pay.PayRequest{
		OrderNo:     orderNo,
		Amount:      req.Amount,
		Currency:    req.Currency,
		PayMethod:   req.PayMethod,
		CountryCode: req.CountryCode,
		ExtraFields: req.ExtraFields,
	})
	if err != nil {
		log.Printf("[AppCreateRechargeOrder] 三方创建订单失败 userID=%d orderNo=%s channel=%s amount=%.2f err=%v", userID, orderNo, req.Channel, req.Amount, err)
		return result, err
	}

	// 三方调用成功，写库（含三方流水号，无需二次 Update）
	var merchantOrderNo *string
	if req.MerchantOrderNo != "" {
		merchantOrderNo = &req.MerchantOrderNo
	}
	var payMethod *string
	if req.PayMethod != "" {
		payMethod = &req.PayMethod
	}
	var providerTradeNo *string
	if payResp.ProviderTradeNo != "" {
		providerTradeNo = &payResp.ProviderTradeNo
	}

	order := pojo.RechargeOrder{
		TenantId:        tgUser.TenantId,
		UserId:          userID,
		SourceChannelID: tgUser.SourceChannelID,
		OrderNo:         orderNo,
		MerchantOrderNo: merchantOrderNo,
		Channel:         req.Channel,
		PayMethod:       payMethod,
		Currency:        req.Currency,
		Amount:          req.Amount,
		Fee:             0,
		BonusAmount:     0,
		Status:          0, // 待支付
		ProviderTradeNo: providerTradeNo,
		ActivityType:    &req.ActivityType,
	}
	if err = db.Create(&order).Error; err != nil {
		log.Printf("[AppCreateRechargeOrder] 写库失败 userID=%d orderNo=%s err=%v", userID, orderNo, err)
		return result, err
	}

	if payResp.AutoSuccess {
		if err = ProcessRechargeOrderSuccess(db, order.OrderNo, payResp.ProviderTradeNo, order.Amount, tablePrefix); err != nil {
			log.Printf("[AppCreateRechargeOrder] 测试通道自动入账失败 userID=%d orderNo=%s channel=%s err=%v", userID, orderNo, req.Channel, err)
			return result, err
		}
		if err = db.Where("id = ?", order.ID).First(&order).Error; err != nil {
			return result, err
		}
	}
	log.Printf("[AppCreateRechargeOrder] 订单创建成功 userID=%d orderNo=%s channel=%s amount=%.2f activityType=%d", userID, orderNo, req.Channel, req.Amount, req.ActivityType)

	result = pojo.RechargeOrderAppBack{
		OrderNo:         order.OrderNo,
		MerchantOrderNo: order.MerchantOrderNo,
		Channel:         order.Channel,
		PayMethod:       order.PayMethod,
		Currency:        order.Currency,
		Amount:          order.Amount,
		Status:          order.Status,
		CreditAmount:    order.CreditAmount,
		PayURL:          payResp.PayURL,
	}
	return result, nil
}

// ProcessRechargeOrderSuccess 处理代收支付成功回调，入账并更新订单状态
// providerTradeNo: 三方交易号；payAmount: 三方实际支付金额（仅记录，入账按订单 amount 计算）
func ProcessRechargeOrderSuccess(db *gorm.DB, orderNo string, providerTradeNo string, payAmount float64, tablePrefix string) error {
	var successUserID int64
	err := db.Transaction(func(tx *gorm.DB) error {
		var order pojo.RechargeOrder
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("order_no = ?", orderNo).First(&order).Error; err != nil {
			return err
		}
		if order.ID == 0 {
			return errors.New(utils.I18nMessage("order_not_found_with_no", map[string]interface{}{"orderNo": orderNo}))
		}
		// 幂等：已成功则直接返回
		if order.Status == 1 {
			return nil
		}
		if order.Status != 0 {
			return errors.New(utils.I18nMessage("order_status_invalid_credit", map[string]interface{}{"status": order.Status}))
		}

		var user pojo.TgUser
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", order.UserId).First(&user).Error; err != nil {
			return err
		}
		if user.ID == 0 {
			return errors.New(utils.I18nMessage("user_not_found_with_id", map[string]interface{}{"userId": order.UserId}))
		}
		isFirstRecharge := user.RechargeAmount <= 0
		bonusAmount := 0.0

		now := time.Now()
		creditAmount := utils.Truncate2(order.Amount - order.Fee + bonusAmount)
		if creditAmount < 0 {
			creditAmount = 0
		}

		updates := map[string]any{
			"status":            1,
			"credit_amount":     creditAmount,
			"bonus_amount":      bonusAmount,
			"pay_time":          now,
			"notify_time":       now,
			"notify_last_at":    now,
			"notify_count":      gorm.Expr("notify_count + 1"),
			"provider_status":   "SUCCESS",
			"provider_trade_no": providerTradeNo,
		}
		if err := tx.Model(&pojo.RechargeOrder{}).Where("id = ?", order.ID).Updates(updates).Error; err != nil {
			return err
		}

		if err := tx.Model(&pojo.TgUser{}).Where("id = ?", user.ID).Updates(map[string]any{
			"balance":         gorm.Expr("balance + ?", creditAmount),
			"gift_amount":     gorm.Expr("gift_amount + ?", bonusAmount),
			"gift_total":      gorm.Expr("gift_total + ?", bonusAmount),
			"recharge_amount": gorm.Expr("recharge_amount + ?", order.Amount),
		}).Error; err != nil {
			return err
		}
		if err := AddUserWithdrawRestrictedBalance(tx, user, bonusAmount, clampRechargeRestrictedCredit(order.Amount-order.Fee)); err != nil {
			return err
		}

		cashHistory := pojo.CashHistory{
			UserId:          user.ID,
			AwardUni:        fmt.Sprintf("recharge_%s", order.OrderNo),
			Amount:          creditAmount,
			StartAmount:     user.Balance,
			EndAmount:       utils.Truncate2(user.Balance + creditAmount),
			CashMark:        "充值到账",
			CashDesc:        fmt.Sprintf("充值订单%s到账", order.OrderNo),
			Type:            pojo.CashHistoryTypeRechargeCredit,
			IsGift:          0,
			FromUserId:      0,
			SourceChannelID: order.SourceChannelID,
		}
		if err := tx.Create(&cashHistory).Error; err != nil {
			return err
		}

		if bonusAmount > 0 {
			giftHistory := pojo.CashHistory{
				UserId:          user.ID,
				AwardUni:        fmt.Sprintf("recharge_gift_%s", order.OrderNo),
				Amount:          bonusAmount,
				StartAmount:     utils.Truncate2(user.Balance + order.Amount - order.Fee),
				EndAmount:       utils.Truncate2(user.Balance + creditAmount),
				CashMark:        "首充赠送",
				CashDesc:        fmt.Sprintf("首充赠送彩金%s，赠送%.2f", order.OrderNo, bonusAmount),
				Type:            pojo.CashHistoryTypeRechargeCredit,
				IsGift:          1,
				FromUserId:      0,
				SourceChannelID: order.SourceChannelID,
			}
			if err := tx.Create(&giftHistory).Error; err != nil {
				return err
			}
			if err := CreatePlatformProfitLedgerIfAbsent(tx, pojo.PlatformProfitLedger{
				TenantId:        order.TenantId,
				UserId:          user.ID,
				SourceChannelID: order.SourceChannelID,
				SourceType:      pojo.PlatformProfitSourceRechargeGift,
				SourceId:        giftHistory.AwardUni,
				IncomeAmount:    0,
				ExpenseAmount:   bonusAmount,
				Remark:          giftHistory.CashDesc,
			}); err != nil {
				return err
			}
		}

		if isFirstRecharge {
			if err := applyInviteFirstRechargeReward(tx, order, user, tablePrefix, now); err != nil {
				return err
			}
		}
		// 活动赠送：activity_type=1(首充) 或 2(今日首充)
		if order.ActivityType != nil && *order.ActivityType > 0 {
			if err := applyFirstRechargeActivityGift(tx, order, user, tablePrefix, now); err != nil {
				return err
			}
		}
		successUserID = order.UserId
		return nil
	})
	if err == nil && successUserID > 0 {
		go CheckAndUpgradeVipLevel(utils.NewPrefixDb(tablePrefix), successUserID)
	}
	return err
}

// ProcessRechargeOrderClosed 处理支付渠道通知订单关闭/取消
func ProcessRechargeOrderClosed(db *gorm.DB, orderNo string) error {
	return db.Model(&pojo.RechargeOrder{}).
		Where("order_no = ? AND status = 0", orderNo).
		Updates(map[string]any{
			"status":          5, // 关闭/超时
			"provider_status": "CLOSED",
			"notify_count":    gorm.Expr("notify_count + 1"),
			"notify_last_at":  time.Now(),
		}).Error
}

// AdminRechargeOrderCallback 管理员手动触发充值回调（将待支付订单标为成功并入账）
func AdminRechargeOrderCallback(db *gorm.DB, id int64, tablePrefix string) (result pojo.RechargeOrderBack, err error) {
	var order pojo.RechargeOrder
	if err = db.Where("id = ?", id).First(&order).Error; err != nil || order.ID == 0 {
		return result, errors.New("order_not_found")
	}
	if order.Status != 0 {
		return result, errors.New("order_status_not_pending_callback")
	}

	isDev := int8(1)
	if err = rechargeOrderDevCallback(db, order.OrderNo, tablePrefix); err != nil {
		return result, err
	}
	// 标记为手动回调
	_ = db.Model(&pojo.RechargeOrder{}).Where("id = ?", id).Update("is_dev", isDev).Error

	_ = db.Where("id = ?", id).First(&order).Error
	_ = copier.Copy(&result, &order)
	return result, nil
}

func buildRechargeOrderNo() string {
	return fmt.Sprintf("RC%s%s", time.Now().Format("20060102150405"), utils.RandomString(6))
}

// rechargeOrderDevCallback 在dev环境模拟三方回调并完成入账
func rechargeOrderDevCallback(db *gorm.DB, orderNo string, tablePrefix string) error {
	var successUserID int64
	err := db.Transaction(func(tx *gorm.DB) error {
		var order pojo.RechargeOrder
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("order_no = ?", orderNo).First(&order).Error; err != nil {
			return err
		}
		if order.Status == 1 {
			return nil
		}
		if order.Status != 0 {
			return errors.New(utils.I18nMessage("order_status_invalid_callback", map[string]interface{}{"status": order.Status}))
		}

		var user pojo.TgUser
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", order.UserId).First(&user).Error; err != nil {
			return err
		}
		if user.ID == 0 {
			return errors.New("user_not_found")
		}
		isFirstRecharge := user.RechargeAmount <= 0
		bonusAmount := 0.0

		now := time.Now()
		creditAmount := utils.Truncate2(order.Amount - order.Fee + bonusAmount)
		if creditAmount < 0 {
			creditAmount = 0
		}
		providerStatus := "SUCCESS"
		if err := tx.Model(&pojo.RechargeOrder{}).
			Where("id = ?", order.ID).
			Updates(map[string]any{
				"status":          1,
				"credit_amount":   creditAmount,
				"bonus_amount":    bonusAmount,
				"pay_time":        now,
				"notify_time":     now,
				"notify_last_at":  now,
				"notify_count":    gorm.Expr("notify_count + 1"),
				"provider_status": providerStatus,
			}).Error; err != nil {
			return err
		}

		if err := tx.Model(&pojo.TgUser{}).
			Where("id = ?", user.ID).
			Updates(map[string]any{
				"balance":         gorm.Expr("balance + ?", creditAmount),
				"gift_amount":     gorm.Expr("gift_amount + ?", bonusAmount),
				"gift_total":      gorm.Expr("gift_total + ?", bonusAmount),
				"recharge_amount": gorm.Expr("recharge_amount + ?", order.Amount),
			}).Error; err != nil {
			return err
		}
		if err := AddUserWithdrawRestrictedBalance(tx, user, bonusAmount, clampRechargeRestrictedCredit(order.Amount-order.Fee)); err != nil {
			return err
		}

		cashHistory := pojo.CashHistory{
			UserId:          user.ID,
			AwardUni:        fmt.Sprintf("recharge_%s", order.OrderNo),
			Amount:          creditAmount,
			StartAmount:     user.Balance,
			EndAmount:       utils.Truncate2(user.Balance + creditAmount),
			CashMark:        "充值到账",
			CashDesc:        fmt.Sprintf("充值订单%s到账", order.OrderNo),
			Type:            pojo.CashHistoryTypeRechargeCredit,
			IsGift:          0,
			FromUserId:      0,
			SourceChannelID: order.SourceChannelID,
		}
		if err := tx.Create(&cashHistory).Error; err != nil {
			return err
		}

		if bonusAmount > 0 {
			giftCashHistory := pojo.CashHistory{
				UserId:          user.ID,
				AwardUni:        fmt.Sprintf("recharge_gift_%s", order.OrderNo),
				Amount:          bonusAmount,
				StartAmount:     utils.Truncate2(user.Balance + order.Amount - order.Fee),
				EndAmount:       utils.Truncate2(user.Balance + creditAmount),
				CashMark:        "首充赠送",
				CashDesc:        fmt.Sprintf("首充赠送彩金%s，赠送%.2f", order.OrderNo, bonusAmount),
				Type:            pojo.CashHistoryTypeRechargeCredit,
				IsGift:          1,
				FromUserId:      0,
				SourceChannelID: order.SourceChannelID,
			}
			if err := tx.Create(&giftCashHistory).Error; err != nil {
				return err
			}
			if err := CreatePlatformProfitLedgerIfAbsent(tx, pojo.PlatformProfitLedger{
				TenantId:        order.TenantId,
				UserId:          user.ID,
				SourceChannelID: order.SourceChannelID,
				SourceType:      pojo.PlatformProfitSourceRechargeGift,
				SourceId:        giftCashHistory.AwardUni,
				IncomeAmount:    0,
				ExpenseAmount:   bonusAmount,
				Remark:          giftCashHistory.CashDesc,
			}); err != nil {
				return err
			}
		}

		if isFirstRecharge {
			if err := applyInviteFirstRechargeReward(tx, order, user, tablePrefix, now); err != nil {
				return err
			}
		}

		// 活动赠送：activity_type=1(首充) 或 2(今日首充)
		if order.ActivityType != nil && *order.ActivityType > 0 {
			if err := applyFirstRechargeActivityGift(tx, order, user, tablePrefix, now); err != nil {
				return err
			}
		}

		successUserID = order.UserId
		return nil
	})
	if err == nil && successUserID > 0 {
		go CheckAndUpgradeVipLevel(utils.NewPrefixDb(tablePrefix), successUserID)
	}
	return err
}

// applyFirstRechargeActivityGift 活动赠送：activity_type=1 首充分 3 天赠送；activity_type=2 今日首充仍一次性赠送
func applyFirstRechargeActivityGift(tx *gorm.DB, order pojo.RechargeOrder, user pojo.TgUser, tablePrefix string, now time.Time) error {
	if order.ActivityType == nil || *order.ActivityType <= 0 {
		return nil
	}

	var completedCount int64
	tx.Model(&pojo.RechargeOrder{}).
		Where("user_id = ? AND activity_type = ? AND status = 1 AND id != ?", user.ID, *order.ActivityType, order.ID).
		Count(&completedCount)
	if completedCount > 0 {
		return nil
	}

	if *order.ActivityType == 1 {
		return applySplitFirstRechargeGift(tx, order, user, tablePrefix, now)
	}
	return applyTodayFirstRechargeGift(tx, order, user)
}

func applySplitFirstRechargeGift(tx *gorm.DB, order pojo.RechargeOrder, user pojo.TgUser, tablePrefix string, now time.Time) error {
	cfg, err := getFirstRechargeGiftConfig(tablePrefix)
	if err != nil {
		log.Printf("[recharge] parse first_recharge_gift_config failed: orderNo=%s err=%v", order.OrderNo, err)
		return nil
	}

	installments := buildFirstRechargeGiftInstallments(order.Amount, now, cfg)
	if len(installments) == 0 {
		return nil
	}

	for _, installment := range installments {
		if installment.GiftAmount <= 0 {
			continue
		}
		if installment.Index == 1 {
			if err = applyFirstRechargeGiftInstallment(tx, order, user.ID, installment, cfg.Rate, cfg.RatioBase); err != nil {
				return err
			}
			continue
		}
		if enqueueErr := enqueueFirstRechargeGiftInstallmentTask(tablePrefix, order, user.ID, installment, cfg.Rate, cfg.RatioBase); enqueueErr != nil {
			log.Printf("[recharge] enqueue first recharge gift installment failed: orderNo=%s installment=%d err=%v", order.OrderNo, installment.Index, enqueueErr)
		}
	}
	return nil
}

func applyTodayFirstRechargeGift(tx *gorm.DB, order pojo.RechargeOrder, user pojo.TgUser) error {
	configKey := "today_first_recharge_gift"
	awardUniPrefix := "today_first_recharge_gift"
	cashMark := "今日首充活动赠送"

	var cfg pojo.SysConfig
	tx.Where("config_key = ?", configKey).First(&cfg)
	if cfg.ID == 0 || strings.TrimSpace(cfg.ConfigValue) == "" {
		return nil
	}
	rate, err := strconv.ParseFloat(strings.TrimSpace(cfg.ConfigValue), 64)
	if err != nil || rate <= 0 {
		return nil
	}

	giftAmount := utils.Truncate2(utils.ToMoney(order.Amount).Multiply(rate / 100).ToDollars())
	if giftAmount <= 0 {
		return nil
	}

	if err = tx.Model(&pojo.TgUser{}).Where("id = ?", user.ID).Updates(map[string]any{
		"balance":     gorm.Expr("balance + ?", giftAmount),
		"gift_amount": gorm.Expr("gift_amount + ?", giftAmount),
		"gift_total":  gorm.Expr("gift_total + ?", giftAmount),
	}).Error; err != nil {
		return err
	}
	if err = AddUserWithdrawRestrictedBalance(tx, user, giftAmount, 0); err != nil {
		return err
	}

	awardUni := fmt.Sprintf("%s_%s", awardUniPrefix, order.OrderNo)
	desc := fmt.Sprintf("%s%.2f，订单%s，充值金额%.2f，赠送比例%.2f%%", cashMark, giftAmount, order.OrderNo, order.Amount, rate)
	history := pojo.CashHistory{
		UserId:          user.ID,
		AwardUni:        awardUni,
		Amount:          giftAmount,
		StartAmount:     user.Balance,
		EndAmount:       utils.Truncate2(user.Balance + giftAmount),
		CashMark:        cashMark,
		CashDesc:        desc,
		Type:            pojo.CashHistoryTypeRechargeCredit,
		IsGift:          1,
		FromUserId:      0,
		SourceChannelID: order.SourceChannelID,
	}
	if err = tx.Create(&history).Error; err != nil {
		return err
	}

	return CreatePlatformProfitLedgerIfAbsent(tx, pojo.PlatformProfitLedger{
		TenantId:        order.TenantId,
		UserId:          user.ID,
		SourceChannelID: order.SourceChannelID,
		SourceType:      pojo.PlatformProfitSourceRechargeGift,
		SourceId:        awardUni,
		IncomeAmount:    0,
		ExpenseAmount:   giftAmount,
		Remark:          desc,
	})
}

func getFirstRechargeGiftConfig(tablePrefix string) (firstRechargeGiftConfig, error) {
	defaultValue := ""
	val := utils.GetStringCache(tablePrefix, "first_recharge_gift_config", &defaultValue)
	if val == nil || strings.TrimSpace(*val) == "" {
		return firstRechargeGiftConfig{}, fmt.Errorf("config is empty")
	}

	raw := strings.TrimSpace(*val)
	matches := regexp.MustCompile(`^([0-9]+(?:\.[0-9]+)?)\((\d+)\|(\d+)\|(\d+)\)$`).FindStringSubmatch(raw)
	if len(matches) != 5 {
		return firstRechargeGiftConfig{}, fmt.Errorf("invalid config format: %s", raw)
	}

	rate, err := strconv.ParseFloat(matches[1], 64)
	if err != nil || rate <= 0 {
		return firstRechargeGiftConfig{}, fmt.Errorf("invalid rate: %s", matches[1])
	}

	ratios := make([]int, 0, 3)
	sum := 0
	for _, item := range matches[2:] {
		v, convErr := strconv.Atoi(item)
		if convErr != nil || v <= 0 {
			return firstRechargeGiftConfig{}, fmt.Errorf("invalid ratio: %s", item)
		}
		sum += v
		ratios = append(ratios, v)
	}
	if sum != 10 && sum != 100 {
		return firstRechargeGiftConfig{}, fmt.Errorf("ratio sum must be 10 or 100: %s", raw)
	}

	return firstRechargeGiftConfig{
		Rate:      rate,
		Ratios:    ratios,
		RatioBase: sum,
	}, nil
}

func buildFirstRechargeGiftInstallments(orderAmount float64, baseTime time.Time, cfg firstRechargeGiftConfig) []firstRechargeGiftInstallment {
	totalGiftMoney := utils.ToMoney(orderAmount).Multiply(cfg.Rate / 100)
	if totalGiftMoney <= 0 {
		return nil
	}

	result := make([]firstRechargeGiftInstallment, 0, len(cfg.Ratios))
	distributed := utils.Money(0)
	for i, ratio := range cfg.Ratios {
		amountMoney := utils.Money(0)
		if i == len(cfg.Ratios)-1 {
			amountMoney = totalGiftMoney.Subtract(distributed)
		} else {
			amountMoney = totalGiftMoney.Multiply(float64(ratio) / float64(cfg.RatioBase))
			distributed = distributed.Add(amountMoney)
		}
		if amountMoney <= 0 {
			continue
		}
		result = append(result, firstRechargeGiftInstallment{
			Index:      i + 1,
			Ratio:      ratio,
			GiftAmount: utils.Truncate2(amountMoney.ToDollars()),
			ExecuteAt:  baseTime.Add(time.Duration(i) * 24 * time.Hour),
		})
	}
	return result
}

func applyFirstRechargeGiftInstallment(tx *gorm.DB, order pojo.RechargeOrder, userID int64, installment firstRechargeGiftInstallment, totalRate float64, ratioBase int) error {
	if installment.GiftAmount <= 0 {
		return nil
	}

	awardUni := fmt.Sprintf("first_recharge_gift_%s_%d", order.OrderNo, installment.Index)
	var existing pojo.CashHistory
	if err := tx.Where("user_id = ? AND award_uni = ?", userID, awardUni).First(&existing).Error; err == nil && existing.ID > 0 {
		return nil
	}

	var user pojo.TgUser
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", userID).First(&user).Error; err != nil {
		return err
	}

	if err := tx.Model(&pojo.TgUser{}).Where("id = ?", userID).Updates(map[string]any{
		"balance":     gorm.Expr("balance + ?", installment.GiftAmount),
		"gift_amount": gorm.Expr("gift_amount + ?", installment.GiftAmount),
		"gift_total":  gorm.Expr("gift_total + ?", installment.GiftAmount),
	}).Error; err != nil {
		return err
	}
	if err := AddUserWithdrawRestrictedBalance(tx, user, installment.GiftAmount, 0); err != nil {
		return err
	}

	desc := fmt.Sprintf(
		"首充活动赠送第%d天，订单%s，充值金额%.2f，赠送比例%.2f%%，分段比例%d/%d，赠送%.2f",
		installment.Index,
		order.OrderNo,
		order.Amount,
		totalRate,
		installment.Ratio,
		ratioBase,
		installment.GiftAmount,
	)
	history := pojo.CashHistory{
		UserId:          userID,
		AwardUni:        awardUni,
		Amount:          installment.GiftAmount,
		StartAmount:     user.Balance,
		EndAmount:       utils.Truncate2(user.Balance + installment.GiftAmount),
		CashMark:        "首充活动赠送",
		CashDesc:        desc,
		Type:            pojo.CashHistoryTypeRechargeCredit,
		IsGift:          1,
		FromUserId:      0,
		SourceChannelID: order.SourceChannelID,
	}
	if err := tx.Create(&history).Error; err != nil {
		return err
	}

	return CreatePlatformProfitLedgerIfAbsent(tx, pojo.PlatformProfitLedger{
		TenantId:        order.TenantId,
		UserId:          userID,
		SourceChannelID: order.SourceChannelID,
		SourceType:      pojo.PlatformProfitSourceRechargeGift,
		SourceId:        awardUni,
		IncomeAmount:    0,
		ExpenseAmount:   installment.GiftAmount,
		Remark:          desc,
	})
}

func enqueueFirstRechargeGiftInstallmentTask(tablePrefix string, order pojo.RechargeOrder, userID int64, installment firstRechargeGiftInstallment, totalRate float64, ratioBase int) error {
	client := getRechargeGiftAsynqClient()
	if client == nil {
		return fmt.Errorf("asynq client is nil")
	}

	payload, _ := json.Marshal(pojo.RechargeFirstGiftInstallmentPayload{
		TablePrefix:      tablePrefix,
		OrderNo:          order.OrderNo,
		UserId:           userID,
		InstallmentIndex: installment.Index,
		GiftAmount:       installment.GiftAmount,
		TotalRate:        totalRate,
		Ratio:            installment.Ratio,
		RatioBase:        ratioBase,
		TenantId:         order.TenantId,
		SourceChannelId:  order.SourceChannelID,
	})
	task := asynq.NewTask(pojo.TaskTypeRechargeFirstGiftInstallment, payload)
	_, err := client.Enqueue(task, asynq.ProcessAt(installment.ExecuteAt), asynq.MaxRetry(10))
	return err
}

func getRechargeGiftAsynqClient() *asynq.Client {
	rechargeGiftAsynqOnce.Do(func() {
		rechargeGiftAsynqClient = asynq.NewClient(asynq.RedisClientOpt{
			Addr:     utils.GlobalConfig.Redis.Host,
			Password: utils.GlobalConfig.Redis.Pass,
			DB:       utils.GlobalConfig.Redis.Db,
		})
	})
	return rechargeGiftAsynqClient
}

func ApplyFirstRechargeGiftInstallmentByOrderNo(db *gorm.DB, orderNo string, installmentIndex int, giftAmount float64, totalRate float64, ratio int, ratioBase int) error {
	if strings.TrimSpace(orderNo) == "" || installmentIndex <= 0 || giftAmount <= 0 {
		return nil
	}
	giftAmount = utils.Truncate2(giftAmount)

	return db.Transaction(func(tx *gorm.DB) error {
		var order pojo.RechargeOrder
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("order_no = ?", orderNo).First(&order).Error; err != nil {
			return err
		}
		if order.ID == 0 || order.Status != 1 || order.ActivityType == nil || *order.ActivityType != 1 {
			return nil
		}
		return applyFirstRechargeGiftInstallment(tx, order, order.UserId, firstRechargeGiftInstallment{
			Index:      installmentIndex,
			Ratio:      ratio,
			GiftAmount: giftAmount,
		}, totalRate, ratioBase)
	})
}

func applyInviteFirstRechargeReward(tx *gorm.DB, order pojo.RechargeOrder, subUser pojo.TgUser, tablePrefix string, now time.Time) error {
	if subUser.ParentID == nil || *subUser.ParentID <= 0 {
		return nil
	}

	rate := getInviteFirstRechargeRewardRate(tablePrefix)
	if rate <= 0 {
		return nil
	}

	rebateAmount := utils.Truncate2(utils.ToMoney(order.Amount).Multiply(rate / 100).ToDollars())
	if rebateAmount < 0.01 {
		return nil
	}

	parentID := *subUser.ParentID
	var parent pojo.TgUser
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", parentID).First(&parent).Error; err != nil || parent.ID == 0 {
		return nil
	}
	if parent.Status != 1 {
		return nil
	}

	idempotencyKey := fmt.Sprintf("first_recharge_reward:%s:%d", order.OrderNo, parentID)
	var existing pojo.TgUserRebateRecord
	if err := tx.Where("idempotency_key = ?", idempotencyKey).First(&existing).Error; err == nil && existing.ID > 0 {
		return nil
	}

	currency := strings.TrimSpace(order.Currency)
	if currency == "" {
		currency = "USDT"
	}
	remark := "first_recharge_reward"
	record := pojo.TgUserRebateRecord{
		TenantId:        &order.TenantId,
		SubUserId:       subUser.ID,
		ParentUserId:    parentID,
		SourceChannelID: order.SourceChannelID,
		SourceType:      5,
		SourceOrderId:   order.OrderNo,
		SourceAmount:    order.Amount,
		RebateRate:      rate,
		RebateAmount:    rebateAmount,
		Currency:        currency,
		Status:          1,
		SettledAt:       &now,
		IdempotencyKey:  idempotencyKey,
		Remark:          &remark,
	}
	if err := tx.Create(&record).Error; err != nil {
		return err
	}

	return tx.Model(&pojo.TgUser{}).Where("id = ?", parentID).Updates(map[string]any{
		"rebate_amount":       gorm.Expr("rebate_amount + ?", rebateAmount),
		"rebate_total_amount": gorm.Expr("rebate_total_amount + ?", rebateAmount),
	}).Error
}

// CheckActivityStatus 检查用户活动参与状态
// hasFirst: 是否已参加过首充活动（activity_type=1 且 status=1）
// hasTodayFirst: 24小时内是否已参加过今日首充活动（activity_type=2 且 status=1）
func CheckActivityStatus(db *gorm.DB, userID int64) (hasFirst bool, hasTodayFirst bool) {
	var firstCount int64
	db.Model(&pojo.RechargeOrder{}).
		Where("user_id = ? AND activity_type = 1 AND status = 1", userID).
		Count(&firstCount)
	hasFirst = firstCount > 0

	var todayCount int64
	db.Model(&pojo.RechargeOrder{}).
		Where("user_id = ? AND activity_type = 2 AND status = 1 AND created_at >= ?", userID, time.Now().Add(-24*time.Hour)).
		Count(&todayCount)
	hasTodayFirst = todayCount > 0
	return
}

func getInviteFirstRechargeRewardRate(tablePrefix string) float64 {
	defaultValue := "10"
	val := utils.GetStringCache(tablePrefix, "invite_first_recharge_reward", &defaultValue)
	if val == nil || strings.TrimSpace(*val) == "" {
		r, _ := strconv.ParseFloat(defaultValue, 64)
		return r
	}
	r, err := strconv.ParseFloat(strings.TrimSpace(*val), 64)
	if err != nil {
		r, _ = strconv.ParseFloat(defaultValue, 64)
		return r
	}
	return r
}

func getRechargeGiftAmount(tablePrefix string) float64 {
	defaultValue := "0"
	val := utils.GetStringCache(tablePrefix, "recharge_gift_amount", &defaultValue)
	if val == nil || strings.TrimSpace(*val) == "" {
		return 0
	}
	amount, err := strconv.ParseFloat(strings.TrimSpace(*val), 64)
	if err != nil || amount < 0 {
		return 0
	}
	return utils.Truncate2(amount)
}

func clampRechargeRestrictedCredit(amount float64) float64 {
	if amount <= 0 {
		return 0
	}
	return utils.Truncate2(amount)
}
