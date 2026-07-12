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
	"math"
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

type firstRechargeGiftConfigV2 struct {
	Rates []float64
}

const (
	RechargeActivityCodeFirstRecharge3Day = "first_recharge_3day"
	RechargeActivityCodeTodayFirst        = "today_first_recharge"
	RechargeActivityCodeV2Gift            = "recharge_v2_gift"
	rechargePromotionTimeFormat           = "2006-01-02 15:04:05"
)

const (
	rechargeActivityTypeFirstRecharge3Day int8 = 1
	rechargeActivityTypeTodayFirst        int8 = 2
	rechargeActivityTypeV2Gift            int8 = 3

	rechargeV2MinAmount      float64 = 29.99 // v2版本充值最低金额
	AdminRechargeV2MinAmount float64 = 1     // 后台手动拉起v2充值最低金额
	rechargeV2GiftMinAmount  float64 = 50    // v2版本充值赠送门槛（满此金额才赠送）
)

func normalizeRechargeWalletType(value string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", pojo.RechargeWalletTypeBalance:
		return pojo.RechargeWalletTypeBalance, nil
	case pojo.RechargeWalletTypeSport, "sports", "sport_balance":
		return pojo.RechargeWalletTypeSport, nil
	default:
		return "", errors.New("invalid_recharge_wallet_type")
	}
}

func rechargeWalletColumn(walletType string) string {
	if walletType == pojo.RechargeWalletTypeSport {
		return "sport_balance"
	}
	return "balance"
}

func rechargeWalletBalance(user pojo.TgUser, walletType string) float64 {
	if walletType == pojo.RechargeWalletTypeSport {
		return user.SportBalance
	}
	return user.Balance
}

func rechargeWalletCashMark(walletType string) string {
	if walletType == pojo.RechargeWalletTypeSport {
		return "体育钱包充值到账"
	}
	return "充值到账"
}

func rechargeWalletCashDesc(walletType string, orderNo string) string {
	if walletType == pojo.RechargeWalletTypeSport {
		return fmt.Sprintf("充值订单%s到账体育钱包", orderNo)
	}
	return fmt.Sprintf("充值订单%s到账", orderNo)
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
	fillRechargeOrderUserUIDs(db, result.List)

	result.PageSize = search.PageSize
	result.CurrentPage = search.CurrentPage
	return result
}

func fillRechargeOrderUserUIDs(db *gorm.DB, orders []pojo.RechargeOrderBack) {
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

// SetRechargeOrder 创建或更新充值订单
func SetRechargeOrder(db *gorm.DB, req pojo.RechargeOrderSet) (result pojo.RechargeOrderBack, err error) {
	if req.Amount > 0 {
		req.Amount = utils.Truncate2(req.Amount)
	}
	if req.Fee > 0 {
		req.Fee = utils.Truncate2(req.Fee)
	}
	if req.NetAmount > 0 {
		req.NetAmount = utils.Truncate2(req.NetAmount)
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
		if strings.TrimSpace(req.WalletType) == "" {
			req.WalletType = dbOrder.WalletType
		}
	}
	walletType, walletErr := normalizeRechargeWalletType(req.WalletType)
	if walletErr != nil {
		return result, walletErr
	}
	req.WalletType = walletType
	if req.ID > 0 {
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

func GetFrontendUnackedRechargeOrders(db *gorm.DB, search pojo.RechargeOrderSearch) (result pojo.RechargeOrderResp) {
	var orders []pojo.RechargeOrder
	query := db.Model(&pojo.RechargeOrder{}).
		Where("status = ? AND frontend_notify_status <> ?", 1, pojo.RechargeFrontendNotifyAcked)
	if search.TenantId > 0 {
		query = query.Where("tenant_id = ?", search.TenantId)
	}
	if search.UserId > 0 {
		query = query.Where("user_id = ?", search.UserId)
	}
	if search.OrderNo != "" {
		query = query.Where("order_no = ?", search.OrderNo)
	}
	if search.FrontendNotifyStatus != nil {
		query = query.Where("frontend_notify_status = ?", *search.FrontendNotifyStatus)
	}
	query.Count(&result.Total)
	query = query.Order("pay_time desc, id desc").Limit(search.PageSize).Offset(search.PageSize * search.CurrentPage)
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

func GetCurrentUserPendingRechargeNotifications(db *gorm.DB, userID int64) ([]pojo.RechargeOrderFrontendNotifyItem, error) {
	var orders []pojo.RechargeOrder
	if err := db.Where("user_id = ? AND status = ? AND frontend_notify_status <> ?", userID, 1, pojo.RechargeFrontendNotifyAcked).
		Order("pay_time desc, id desc").
		Limit(50).
		Find(&orders).Error; err != nil {
		return nil, err
	}
	result := make([]pojo.RechargeOrderFrontendNotifyItem, 0, len(orders))
	for _, order := range orders {
		result = append(result, pojo.RechargeOrderFrontendNotifyItem{
			OrderNo:              order.OrderNo,
			Channel:              order.Channel,
			Currency:             order.Currency,
			Amount:               utils.Truncate2(order.Amount),
			CreditAmount:         order.CreditAmount,
			BonusAmount:          utils.Truncate2(order.BonusAmount),
			WalletType:           order.WalletType,
			Status:               order.Status,
			IsFirstRecharge:      order.IsFirstRecharge,
			PayTime:              order.PayTime,
			FrontendNotifyStatus: order.FrontendNotifyStatus,
			FrontendNotifyCount:  order.FrontendNotifyCount,
			FrontendNotifyAt:     order.FrontendNotifyAt,
			FrontendNotifyAckAt:  order.FrontendNotifyAckAt,
		})
	}
	return result, nil
}

// GetUserRechargeCount 统计当前用户成功充值次数（status=1，包含手动回调）
func GetUserRechargeCount(db *gorm.DB, userID int64) pojo.UserRechargeCountBack {
	var count int64
	db.Model(&pojo.RechargeOrder{}).Where("user_id = ? AND status = 1", userID).Count(&count)
	// 是否曾把佣金转入余额（cash_history type=佣金转余额）
	var transferCount int64
	db.Model(&pojo.CashHistory{}).
		Where("user_id = ? AND type = ?", userID, pojo.CashHistoryTypeRebateTransfer).
		Count(&transferCount)
	return pojo.UserRechargeCountBack{RechargeCount: count, RebateTransferred: transferCount > 0}
}

func AckRechargeFrontendNotification(db *gorm.DB, userID int64, orderNo string) error {
	orderNo = strings.TrimSpace(orderNo)
	if userID <= 0 || orderNo == "" {
		return errors.New("invalid_params")
	}
	now := time.Now()
	result := db.Model(&pojo.RechargeOrder{}).
		Where("user_id = ? AND order_no = ? AND status = ?", userID, orderNo, 1).
		Updates(map[string]any{
			"frontend_notify_status": pojo.RechargeFrontendNotifyAcked,
			"frontend_notify_ack_at": now,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("order_not_found")
	}
	return nil
}

func AppCreateRechargeOrder(db *gorm.DB, userID int64, req pojo.RechargeOrderAppReq, tablePrefix string) (result pojo.RechargeOrderAppBack, err error) {
	return appCreateRechargeOrder(db, userID, req, tablePrefix, nil, 0)
}

func AppCreateRechargeOrderV2(db *gorm.DB, userID int64, req pojo.RechargeOrderAppReq, tablePrefix string) (result pojo.RechargeOrderAppBack, err error) {
	activityType := rechargeActivityTypeV2Gift
	req = applyRechargeV2DefaultProvider(req)
	return appCreateRechargeOrder(db, userID, req, tablePrefix, &activityType, rechargeV2MinAmount)
}

func AdminCreateRechargeOrderV2(db *gorm.DB, userID int64, req pojo.RechargeOrderAppReq, tablePrefix string) (result pojo.RechargeOrderAppBack, err error) {
	activityType := rechargeActivityTypeV2Gift
	req = applyRechargeV2DefaultProvider(req)
	return appCreateRechargeOrder(db, userID, req, tablePrefix, &activityType, AdminRechargeV2MinAmount)
}

func applyRechargeV2DefaultProvider(req pojo.RechargeOrderAppReq) pojo.RechargeOrderAppReq {
	if strings.TrimSpace(req.Channel) != "" {
		return req
	}
	if pojo.NormalizeWithdrawCountryCode(req.CountryCode) != "MX" {
		return req
	}
	req.Channel = "VCPAYMXN"
	if strings.TrimSpace(req.Currency) == "" {
		req.Currency = "MXN"
	}
	return req
}

// appCreateRechargeOrder app端创建充值订单（dev环境自动回调）
func appCreateRechargeOrder(db *gorm.DB, userID int64, req pojo.RechargeOrderAppReq, tablePrefix string, forcedActivityType *int8, minAmount float64) (result pojo.RechargeOrderAppBack, err error) {
	req.Channel = strings.TrimSpace(req.Channel)
	req.PayMethod = strings.TrimSpace(req.PayMethod)
	req.Currency = strings.ToUpper(strings.TrimSpace(req.Currency))
	req.MerchantOrderNo = strings.TrimSpace(req.MerchantOrderNo)
	req.Amount = normalizeRechargeOrderAmount(req.Amount)
	walletType, walletErr := normalizeRechargeWalletType(req.WalletType)
	if walletErr != nil {
		return result, walletErr
	}
	req.WalletType = walletType
	if req.Amount <= 0 {
		return result, errors.New("recharge_amount_positive")
	}
	if forcedActivityType != nil && *forcedActivityType == rechargeActivityTypeV2Gift && minAmount > 0 && req.Amount < minAmount {
		return result, errors.New(utils.I18nMessage("recharge_v2_min_amount", map[string]interface{}{
			"min": formatRechargeMinAmount(minAmount),
		}))
	}
	if req.Channel == "" {
		return result, errors.New("recharge_channel_required")
	}
	if req.Currency == "" {
		req.Currency = "BRL"
	}
	req.CountryCode = strings.TrimSpace(req.CountryCode)
	var country pojo.SysCountry

	// 若提供了 countryCode，根据国家字段配置校验 extraFields
	if req.CountryCode != "" {
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
	if err = validateRechargeChannelAvailable(db, req.Channel, req.CountryCode); err != nil {
		return result, err
	}
	providerAmount := req.Amount
	if country.ID > 0 && country.Rate > 0 {
		providerAmount = calculateRechargeProviderAmount(req.Amount, country.Rate)
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
	activityType, err := resolveRechargeOrderCreateActivityType(db, userID, req, tablePrefix, forcedActivityType)
	if err != nil {
		return result, err
	}
	if req.WalletType == pojo.RechargeWalletTypeSport && activityType == rechargeActivityTypeV2Gift {
		activityType = 0
	}
	activityCode := strings.TrimSpace(req.ActivityCode)
	if forcedActivityType != nil {
		activityCode = rechargeActivityCodeByType(activityType)
	}
	if req.WalletType == pojo.RechargeWalletTypeBalance && shouldConfirmUnfinishedActivityCycleForRecharge(activityType) && !req.ConfirmUnfinishedActivityCycle {
		activeCycle, cycleErr := GetActiveWithdrawActivityCycle(db, userID)
		if cycleErr != nil {
			return result, cycleErr
		}
		if activeCycle.ID > 0 {
			if CanBypassWithdrawActivityCycleByBalance(tgUser.Balance, activeCycle) {
				if endErr := EndWithdrawActivityCycle(db, userID, pojo.WithdrawActivityCycleEndReasonBalanceBelowLimit); endErr != nil {
					return result, endErr
				}
				if resetErr := ResetUserWithdrawLimitAfterActivityEnd(db, userID, tgUser.Balance, false); resetErr != nil {
					return result, resetErr
				}
			} else {
				result.NeedConfirmUnfinishedActivityCycle = true
				result.ActiveActivityMultiplier = activeCycle.Multiplier
				return result, nil
			}
		}
	}

	orderNo := buildRechargeOrderNo()
	isDev := utils.IsDev()
	payResp := pay.PayResponse{}
	if isDev {
		payResp.ProviderTradeNo = "DEV_" + orderNo
		payResp.AutoSuccess = true
	} else {
		// 延迟写库策略：先解析渠道、调用三方，成功后再落库，失败不留脏数据
		provider, providerErr := pay.MustGet(req.Channel)
		if providerErr != nil {
			log.Printf("[AppCreateRechargeOrder] 渠道未注册 userID=%d channel=%s err=%v", userID, req.Channel, providerErr)
			return result, providerErr
		}

		payResp, err = provider.CreateOrder(pay.PayRequest{
			UserID:         userID,
			UserUID:        tgUser.Uid,
			OrderNo:        orderNo,
			Amount:         req.Amount,
			ProviderAmount: providerAmount,
			Currency:       req.Currency,
			PayMethod:      req.PayMethod,
			CountryCode:    req.CountryCode,
			ClientIP:       req.ClientIP,
			UserIP:         ptrValue(tgUser.Ip),
			ExtraFields:    req.ExtraFields,
			ReturnURL:      req.ReturnURL,
		})
		if err != nil {
			log.Printf("[AppCreateRechargeOrder] 三方创建订单失败 userID=%d orderNo=%s channel=%s amount=%.2f providerAmount=%.2f currency=%s country=%s err=%v", userID, orderNo, req.Channel, req.Amount, providerAmount, req.Currency, req.CountryCode, err)
			return result, err
		}
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
	var devFlag *int8
	if isDev {
		flag := int8(1)
		devFlag = &flag
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
		NetAmount:       providerAmount,
		BonusAmount:     0,
		WalletType:      req.WalletType,
		Status:          0, // 待支付
		ProviderTradeNo: providerTradeNo,
		IsDev:           devFlag,
		ActivityType:    &activityType,
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
	log.Printf("[AppCreateRechargeOrder] 订单创建成功 userID=%d orderNo=%s channel=%s amount=%.2f activityType=%d activityCode=%s", userID, orderNo, req.Channel, req.Amount, activityType, activityCode)

	result = pojo.RechargeOrderAppBack{
		OrderNo:         order.OrderNo,
		MerchantOrderNo: order.MerchantOrderNo,
		Channel:         order.Channel,
		PayMethod:       order.PayMethod,
		Currency:        order.Currency,
		Amount:          order.Amount,
		NetAmount:       order.NetAmount,
		Status:          order.Status,
		CreditAmount:    order.CreditAmount,
		BonusAmount:     order.BonusAmount,
		WalletType:      order.WalletType,
		PayURL:          payResp.PayURL,
	}
	return result, nil
}

func validateRechargeChannelAvailable(db *gorm.DB, channel string, countryCode string) error {
	channel = strings.TrimSpace(channel)
	if channel == "" {
		return errors.New("recharge_channel_required")
	}

	var entity pojo.SysPayChannel
	err := db.
		Where("channel_code = ? AND deleted_at = 0", channel).
		First(&entity).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}

	if entity.Status != 1 {
		return errors.New(utils.I18nMessage("recharge_channel_disabled", nil))
	}
	channelType := strings.ToLower(strings.TrimSpace(entity.ChannelType))
	if channelType != "deposit" && channelType != "both" {
		return errors.New(utils.I18nMessage("recharge_channel_disabled", nil))
	}
	if countryCode != "" && entity.CountryCode != nil {
		channelCountry := pojo.NormalizeWithdrawCountryCode(*entity.CountryCode)
		if channelCountry != "" && channelCountry != pojo.NormalizeWithdrawCountryCode(countryCode) {
			return errors.New(utils.I18nMessage("recharge_channel_disabled", nil))
		}
	}
	return nil
}

func ceilProviderAmount(amount float64) float64 {
	if amount <= 0 || math.IsNaN(amount) || math.IsInf(amount, 0) {
		return 0
	}
	return math.Ceil(amount)
}

func calculateRechargeProviderAmount(amount float64, rate float64) float64 {
	if rate <= 0 {
		return amount
	}
	if rate == 1 {
		return normalizeRechargeOrderAmount(amount)
	}
	return floorRechargeAmount(amount * rate)
}

func floorRechargeAmount(amount float64) float64 {
	if amount <= 0 || math.IsNaN(amount) || math.IsInf(amount, 0) {
		return 0
	}
	return math.Floor(amount)
}

func normalizeRechargeOrderAmount(amount float64) float64 {
	if amount <= 0 || math.IsNaN(amount) || math.IsInf(amount, 0) {
		return 0
	}
	return utils.Truncate2(amount)
}

func formatRechargeMinAmount(amount float64) string {
	amount = utils.Truncate2(amount)
	if math.Mod(amount, 1) == 0 {
		return fmt.Sprintf("%.0f", amount)
	}
	return fmt.Sprintf("%.2f", amount)
}

func addRechargeOrderBonusAmount(tx *gorm.DB, orderID int64, bonusAmount float64) error {
	bonusAmount = utils.Truncate2(bonusAmount)
	if tx == nil || orderID <= 0 || bonusAmount <= 0 {
		return nil
	}
	return tx.Model(&pojo.RechargeOrder{}).
		Where("id = ?", orderID).
		Update("bonus_amount", gorm.Expr("bonus_amount + ?", bonusAmount)).Error
}

func rechargeCallbackCreditBaseAmount(orderAmount float64, payAmount float64) float64 {
	orderAmount = utils.Truncate2(orderAmount)
	if payAmount > 0 {
		payAmount = utils.Truncate2(payAmount)
		if rechargeDisplayAmountRequiresOneCentCredit(orderAmount) && payAmount == orderAmount {
			return rechargeOneCentCreditAmount(orderAmount)
		}
		return payAmount
	}
	return rechargeCreditBaseAmountFromOrderAmount(orderAmount)
}

func rechargeCreditBaseAmountFromOrderAmount(orderAmount float64) float64 {
	orderAmount = utils.Truncate2(orderAmount)
	if rechargeDisplayAmountRequiresOneCentCredit(orderAmount) {
		return rechargeOneCentCreditAmount(orderAmount)
	}
	return orderAmount
}

func rechargeOneCentCreditAmount(amount float64) float64 {
	return utils.Truncate2(amount + 0.01)
}

func rechargeDisplayAmountRequiresOneCentCredit(amount float64) bool {
	switch utils.Truncate2(amount) {
	case 29.99, 39.99, 49.99, 99.99, 149.99, 199.99, 249.99, 299.99, 399.99, 499.99:
		return true
	default:
		return false
	}
}

func calculateRechargeFreeLotteryCount(amount float64) int {
	if amount < 200 || math.IsNaN(amount) || math.IsInf(amount, 0) {
		return 0
	}
	return int(math.Floor(amount / 200))
}

// ProcessRechargeOrderSuccess 处理代收支付成功回调，入账并更新订单状态
// providerTradeNo: 三方交易号；payAmount: 三方实际支付金额（元），为空时回退订单 amount。
func ProcessRechargeOrderSuccess(db *gorm.DB, orderNo string, providerTradeNo string, payAmount float64, tablePrefix string) error {
	var successUserID int64
	var successOrderNo string
	notifyRecharge := false
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
			log.Printf("[recharge] pay callback idempotent success orderNo=%s userID=%d tablePrefix=%q activityType=%s creditAmount=%.2f bonusAmount=%.2f providerTradeNo=%s",
				order.OrderNo, order.UserId, tablePrefix, formatRechargeActivityType(order.ActivityType), order.CreditAmount, order.BonusAmount, providerTradeNo)
			successUserID = order.UserId
			successOrderNo = order.OrderNo
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
		walletType, walletErr := normalizeRechargeWalletType(order.WalletType)
		if walletErr != nil {
			return walletErr
		}
		order.WalletType = walletType
		startWalletBalance := rechargeWalletBalance(user, walletType)
		isFirstRecharge := user.RechargeAmount <= 0
		bonusAmount := 0.0
		log.Printf("[recharge] pay callback begin orderNo=%s userID=%d tablePrefix=%q amount=%.2f fee=%.2f status=%d activityType=%s walletType=%s walletBalance=%.2f userRechargeAmount=%.2f isFirstRecharge=%t providerTradeNo=%s payAmount=%.2f",
			order.OrderNo, user.ID, tablePrefix, order.Amount, order.Fee, order.Status, formatRechargeActivityType(order.ActivityType), walletType, startWalletBalance, user.RechargeAmount, isFirstRecharge, providerTradeNo, payAmount)

		now := time.Now()
		creditBaseAmount := rechargeCallbackCreditBaseAmount(order.Amount, payAmount)
		creditAmount := utils.Truncate2(creditBaseAmount - order.Fee + bonusAmount)
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
			"is_first_recharge": isFirstRecharge,
			"wallet_type":       walletType,
		}
		if err := tx.Model(&pojo.RechargeOrder{}).Where("id = ?", order.ID).Updates(updates).Error; err != nil {
			return err
		}
		order.Amount = creditBaseAmount
		order.CreditAmount = &creditAmount
		order.BonusAmount = bonusAmount
		order.PayTime = &now
		order.Status = 1
		log.Printf("[recharge] pay callback order marked success orderNo=%s userID=%d tablePrefix=%q creditAmount=%.2f bonusAmount=%.2f isFirstRecharge=%t activityType=%s",
			order.OrderNo, user.ID, tablePrefix, creditAmount, bonusAmount, isFirstRecharge, formatRechargeActivityType(order.ActivityType))

		walletColumn := rechargeWalletColumn(walletType)
		userUpdates := map[string]any{
			walletColumn:      gorm.Expr(walletColumn+" + ?", creditAmount),
			"recharge_amount": gorm.Expr("recharge_amount + ?", creditBaseAmount),
			"free_lottery_count": gorm.Expr(
				"free_lottery_count + ?",
				calculateRechargeFreeLotteryCount(creditBaseAmount),
			),
		}
		if walletType == pojo.RechargeWalletTypeBalance {
			userUpdates["gift_amount"] = gorm.Expr("gift_amount + ?", bonusAmount)
			userUpdates["gift_total"] = gorm.Expr("gift_total + ?", bonusAmount)
		}
		if err := tx.Model(&pojo.TgUser{}).Where("id = ?", user.ID).Updates(userUpdates).Error; err != nil {
			return err
		}
		log.Printf("[recharge] pay callback user credited orderNo=%s userID=%d tablePrefix=%q walletType=%s rechargeCredit=%.2f activityBaseGift=%.2f startBalance=%.2f lotteryCount=%d",
			order.OrderNo, user.ID, tablePrefix, walletType, creditAmount, bonusAmount, startWalletBalance, calculateRechargeFreeLotteryCount(creditBaseAmount))
		if err := ApplyInviteRechargeRebate(tx, order, now); err != nil {
			return err
		}
		if _, err := EnsureInviteValidUser(tx, user.ID, now); err != nil {
			return err
		}
		if err := ApplyTaskActivityRechargeProgress(tx, order, now); err != nil {
			return err
		}
		if walletType == pojo.RechargeWalletTypeBalance && (order.ActivityType == nil || *order.ActivityType != rechargeActivityTypeV2Gift) {
			if err := AddUserWithdrawRestrictedBalance(tx, user, bonusAmount, clampRechargeRestrictedCredit(creditBaseAmount-order.Fee)); err != nil {
				return err
			}
		}

		cashHistory := pojo.CashHistory{
			UserId:          user.ID,
			AwardUni:        fmt.Sprintf("recharge_%s", order.OrderNo),
			Amount:          creditAmount,
			StartAmount:     startWalletBalance,
			EndAmount:       utils.Truncate2(startWalletBalance + creditAmount),
			CashMark:        rechargeWalletCashMark(walletType),
			CashDesc:        rechargeWalletCashDesc(walletType, order.OrderNo),
			Type:            pojo.CashHistoryTypeRechargeCredit,
			IsGift:          0,
			FromUserId:      0,
			SourceChannelID: order.SourceChannelID,
		}
		if err := tx.Create(&cashHistory).Error; err != nil {
			return err
		}
		log.Printf("[recharge] pay callback recharge cash history created orderNo=%s userID=%d cashHistoryID=%d awardUni=%s amount=%.2f",
			order.OrderNo, user.ID, cashHistory.ID, cashHistory.AwardUni, cashHistory.Amount)

		if bonusAmount > 0 {
			giftHistory := pojo.CashHistory{
				UserId:          user.ID,
				AwardUni:        fmt.Sprintf("recharge_gift_%s", order.OrderNo),
				Amount:          bonusAmount,
				StartAmount:     utils.Truncate2(user.Balance + creditBaseAmount - order.Fee),
				EndAmount:       utils.Truncate2(user.Balance + creditAmount),
				CashMark:        "首充赠送",
				CashDesc:        fmt.Sprintf("首充赠送彩金%s，赠送%.2f", order.OrderNo, bonusAmount),
				Type:            pojo.CashHistoryTypeFirstRechargeGift,
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

		if err := applyFirstLevelRechargeRebate(tx, order, user, tablePrefix, now); err != nil {
			return err
		}
		// 活动赠送：activity_type=1(三日首充)、2(今日首充)、3(v2充值赠送)
		if walletType == pojo.RechargeWalletTypeBalance && order.ActivityType != nil && *order.ActivityType > 0 {
			log.Printf("[recharge] pay callback apply activity gift orderNo=%s userID=%d activityType=%d tablePrefix=%q", order.OrderNo, user.ID, *order.ActivityType, tablePrefix)
			if err := applyFirstRechargeActivityGift(tx, order, user, tablePrefix, now); err != nil {
				return err
			}
		} else {
			log.Printf("[recharge] pay callback skip activity gift orderNo=%s userID=%d activityType=%s walletType=%s", order.OrderNo, user.ID, formatRechargeActivityType(order.ActivityType), walletType)
		}
		if walletType == pojo.RechargeWalletTypeBalance {
			var latestUser pojo.TgUser
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", user.ID).First(&latestUser).Error; err != nil {
				return err
			}
			if order.ActivityType != nil && *order.ActivityType == rechargeActivityTypeV2Gift {
				if err := EnsureWithdrawFlowBatchForRechargeV2(tx, latestUser, order, user.Balance); err != nil {
					return err
				}
			} else if order.ActivityType != nil && *order.ActivityType > 0 {
				if err := EnsureWithdrawActivityCycleForRecharge(
					tx,
					latestUser,
					order,
					GetWithdrawActivityCycleMultiplier(tx),
					GetWithdrawActivityBalanceThreshold(tx),
					rechargeActivityCodeByType(*order.ActivityType),
				); err != nil {
					return err
				}
			} else if err := RefreshActiveWithdrawActivityCycleForRecharge(tx, latestUser, order); err != nil {
				return err
			}
		}
		successUserID = order.UserId
		successOrderNo = order.OrderNo
		notifyRecharge = true
		return nil
	})
	if err == nil && successUserID > 0 {
		go CheckAndUpgradeVipLevel(utils.NewPrefixDb(tablePrefix), successUserID)
		go pushRechargeSuccessFrontendNotification(utils.NewPrefixDb(tablePrefix), successOrderNo)
		if notifyRecharge {
			go notifyTelegramRechargeByOrderNo(utils.NewPrefixDb(tablePrefix), successOrderNo)
		}
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
	log.Printf("[recharge] admin manual callback request orderID=%d orderNo=%s userID=%d tablePrefix=%q amount=%.2f status=%d activityType=%s",
		order.ID, order.OrderNo, order.UserId, tablePrefix, order.Amount, order.Status, formatRechargeActivityType(order.ActivityType))

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
	var successOrderNo string
	notifyRecharge := false
	err := db.Transaction(func(tx *gorm.DB) error {
		var order pojo.RechargeOrder
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("order_no = ?", orderNo).First(&order).Error; err != nil {
			return err
		}
		if order.Status == 1 {
			successUserID = order.UserId
			successOrderNo = order.OrderNo
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
		walletType, walletErr := normalizeRechargeWalletType(order.WalletType)
		if walletErr != nil {
			return walletErr
		}
		order.WalletType = walletType
		startWalletBalance := rechargeWalletBalance(user, walletType)
		isFirstRecharge := user.RechargeAmount <= 0
		bonusAmount := 0.0
		log.Printf("[recharge] manual callback begin orderNo=%s userID=%d tablePrefix=%q amount=%.2f fee=%.2f status=%d activityType=%s walletType=%s walletBalance=%.2f userRechargeAmount=%.2f isFirstRecharge=%t",
			order.OrderNo, user.ID, tablePrefix, order.Amount, order.Fee, order.Status, formatRechargeActivityType(order.ActivityType), walletType, startWalletBalance, user.RechargeAmount, isFirstRecharge)

		now := time.Now()
		creditBaseAmount := rechargeCreditBaseAmountFromOrderAmount(order.Amount)
		creditAmount := utils.Truncate2(creditBaseAmount - order.Fee + bonusAmount)
		if creditAmount < 0 {
			creditAmount = 0
		}
		providerStatus := "SUCCESS"
		if err := tx.Model(&pojo.RechargeOrder{}).
			Where("id = ?", order.ID).
			Updates(map[string]any{
				"status":            1,
				"credit_amount":     creditAmount,
				"bonus_amount":      bonusAmount,
				"pay_time":          now,
				"notify_time":       now,
				"notify_last_at":    now,
				"notify_count":      gorm.Expr("notify_count + 1"),
				"provider_status":   providerStatus,
				"is_first_recharge": isFirstRecharge,
				"wallet_type":       walletType,
			}).Error; err != nil {
			return err
		}
		order.Amount = creditBaseAmount
		order.CreditAmount = &creditAmount
		order.BonusAmount = bonusAmount
		order.PayTime = &now
		order.Status = 1

		walletColumn := rechargeWalletColumn(walletType)
		userUpdates := map[string]any{
			walletColumn:      gorm.Expr(walletColumn+" + ?", creditAmount),
			"recharge_amount": gorm.Expr("recharge_amount + ?", creditBaseAmount),
			"free_lottery_count": gorm.Expr(
				"free_lottery_count + ?",
				calculateRechargeFreeLotteryCount(creditBaseAmount),
			),
		}
		if walletType == pojo.RechargeWalletTypeBalance {
			userUpdates["gift_amount"] = gorm.Expr("gift_amount + ?", bonusAmount)
			userUpdates["gift_total"] = gorm.Expr("gift_total + ?", bonusAmount)
		}
		if err := tx.Model(&pojo.TgUser{}).
			Where("id = ?", user.ID).
			Updates(userUpdates).Error; err != nil {
			return err
		}
		if walletType == pojo.RechargeWalletTypeBalance && (order.ActivityType == nil || *order.ActivityType != rechargeActivityTypeV2Gift) {
			if err := AddUserWithdrawRestrictedBalance(tx, user, bonusAmount, clampRechargeRestrictedCredit(creditBaseAmount-order.Fee)); err != nil {
				return err
			}
		}

		cashHistory := pojo.CashHistory{
			UserId:          user.ID,
			AwardUni:        fmt.Sprintf("recharge_%s", order.OrderNo),
			Amount:          creditAmount,
			StartAmount:     startWalletBalance,
			EndAmount:       utils.Truncate2(startWalletBalance + creditAmount),
			CashMark:        rechargeWalletCashMark(walletType),
			CashDesc:        rechargeWalletCashDesc(walletType, order.OrderNo),
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
				StartAmount:     utils.Truncate2(user.Balance + creditBaseAmount - order.Fee),
				EndAmount:       utils.Truncate2(user.Balance + creditAmount),
				CashMark:        "首充赠送",
				CashDesc:        fmt.Sprintf("首充赠送彩金%s，赠送%.2f", order.OrderNo, bonusAmount),
				Type:            pojo.CashHistoryTypeFirstRechargeGift,
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

		if err := applyFirstLevelRechargeRebate(tx, order, user, tablePrefix, now); err != nil {
			return err
		}

		// 邀请充值返佣（手动回调同样发放；此时 is_dev 尚未置位，且发放/计次均不排除手动单）
		if err := ApplyInviteRechargeRebate(tx, order, now); err != nil {
			return err
		}
		if _, err := EnsureInviteValidUser(tx, user.ID, now); err != nil {
			return err
		}

		// 活动赠送：activity_type=1(三日首充)、2(今日首充)、3(v2充值赠送)
		if walletType == pojo.RechargeWalletTypeBalance && order.ActivityType != nil && *order.ActivityType > 0 {
			log.Printf("[recharge] manual callback apply activity gift orderNo=%s userID=%d activityType=%d tablePrefix=%q", order.OrderNo, user.ID, *order.ActivityType, tablePrefix)
			if err := applyFirstRechargeActivityGift(tx, order, user, tablePrefix, now); err != nil {
				return err
			}
		} else {
			log.Printf("[recharge] manual callback skip activity gift orderNo=%s userID=%d activityType=%s walletType=%s", order.OrderNo, user.ID, formatRechargeActivityType(order.ActivityType), walletType)
		}
		if walletType == pojo.RechargeWalletTypeBalance && order.ActivityType != nil && *order.ActivityType == rechargeActivityTypeV2Gift {
			var latestUser pojo.TgUser
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", user.ID).First(&latestUser).Error; err != nil {
				return err
			}
			if err := EnsureWithdrawFlowBatchForRechargeV2(tx, latestUser, order, user.Balance); err != nil {
				return err
			}
		}

		successUserID = order.UserId
		successOrderNo = order.OrderNo
		notifyRecharge = true
		return nil
	})
	if err == nil && successUserID > 0 {
		go CheckAndUpgradeVipLevel(utils.NewPrefixDb(tablePrefix), successUserID)
		go pushRechargeSuccessFrontendNotification(utils.NewPrefixDb(tablePrefix), successOrderNo)
		if notifyRecharge {
			go notifyTelegramRechargeByOrderNo(utils.NewPrefixDb(tablePrefix), successOrderNo)
		}
	}
	return err
}

func pushRechargeSuccessFrontendNotification(db *gorm.DB, orderNo string) {
	orderNo = strings.TrimSpace(orderNo)
	if orderNo == "" {
		return
	}
	var order pojo.RechargeOrder
	if err := db.Where("order_no = ?", orderNo).First(&order).Error; err != nil || order.ID == 0 {
		return
	}
	if order.Status != 1 || order.FrontendNotifyStatus == pojo.RechargeFrontendNotifyAcked {
		return
	}
	payload := pojo.RechargeOrderFrontendNotifyItem{
		OrderNo:              order.OrderNo,
		Channel:              order.Channel,
		Currency:             order.Currency,
		Amount:               utils.Truncate2(order.Amount),
		CreditAmount:         order.CreditAmount,
		BonusAmount:          utils.Truncate2(order.BonusAmount),
		WalletType:           order.WalletType,
		Status:               order.Status,
		IsFirstRecharge:      order.IsFirstRecharge,
		PayTime:              order.PayTime,
		FrontendNotifyStatus: order.FrontendNotifyStatus,
		FrontendNotifyCount:  order.FrontendNotifyCount,
		FrontendNotifyAt:     order.FrontendNotifyAt,
		FrontendNotifyAckAt:  order.FrontendNotifyAckAt,
	}
	delivered, err := utils.SendWsUserWithType(5, order.TenantId, order.UserId, "recharge_success", payload)
	if err != nil {
		return
	}
	if !delivered {
		return
	}
	now := time.Now()
	_ = db.Model(&pojo.RechargeOrder{}).
		Where("id = ? AND frontend_notify_status <> ?", order.ID, pojo.RechargeFrontendNotifyAcked).
		Updates(map[string]any{
			"frontend_notify_status": pojo.RechargeFrontendNotifySent,
			"frontend_notify_count":  gorm.Expr("frontend_notify_count + 1"),
			"frontend_notify_at":     now,
		}).Error
}

// applyFirstRechargeActivityGift 活动赠送：activity_type=1 三日首充；2 今日首充；3 v2充值赠送。
func applyFirstRechargeActivityGift(tx *gorm.DB, order pojo.RechargeOrder, user pojo.TgUser, tablePrefix string, now time.Time) error {
	if order.ActivityType == nil || *order.ActivityType <= 0 {
		log.Printf("[recharge] activity gift skip: invalid activity type orderNo=%s userID=%d activityType=%s tablePrefix=%q",
			order.OrderNo, user.ID, formatRechargeActivityType(order.ActivityType), tablePrefix)
		return nil
	}

	log.Printf("[recharge] activity gift begin orderNo=%s userID=%d activityType=%d tablePrefix=%q amount=%.2f payTime=%s now=%s",
		order.OrderNo, user.ID, *order.ActivityType, tablePrefix, order.Amount, formatRechargeTime(order.PayTime), now.Format(time.RFC3339))
	switch *order.ActivityType {
	case rechargeActivityTypeFirstRecharge3Day:
		return applyFirstRechargeGiftV2(tx, order, user, tablePrefix, now)
	case rechargeActivityTypeV2Gift:
		return applyRechargeV2Gift(tx, order, user)
	}

	var completedCount int64
	if err := tx.Model(&pojo.RechargeOrder{}).
		Where("user_id = ? AND activity_type = ? AND status = 1 AND id != ?", user.ID, *order.ActivityType, order.ID).
		Count(&completedCount).Error; err != nil {
		log.Printf("[recharge] today first gift count failed orderNo=%s userID=%d activityType=%d err=%v", order.OrderNo, user.ID, *order.ActivityType, err)
		return err
	}
	if completedCount > 0 {
		log.Printf("[recharge] today first gift skip: completed order exists orderNo=%s userID=%d activityType=%d completedCount=%d", order.OrderNo, user.ID, *order.ActivityType, completedCount)
		return nil
	}

	return applyTodayFirstRechargeGift(tx, order, user)
}

func applyFirstRechargeGiftV2(tx *gorm.DB, order pojo.RechargeOrder, user pojo.TgUser, tablePrefix string, now time.Time) error {
	cfg, err := getFirstRechargeGiftConfigV2(tablePrefix)
	if err != nil {
		log.Printf("[recharge] first recharge gift v2 skip: config parse failed orderNo=%s userID=%d tablePrefix=%q err=%v", order.OrderNo, user.ID, tablePrefix, err)
		return nil
	}
	log.Printf("[recharge] first recharge gift v2 config loaded orderNo=%s userID=%d tablePrefix=%q rates=%v", order.OrderNo, user.ID, tablePrefix, cfg.Rates)

	payAt := now
	if order.PayTime != nil {
		payAt = *order.PayTime
	}

	startAt := payAt
	var firstOrder pojo.RechargeOrder
	if err = tx.Where("user_id = ? AND activity_type = 1 AND status = 1 AND pay_time IS NOT NULL", user.ID).
		Order("pay_time asc, id asc").
		First(&firstOrder).Error; err == nil && firstOrder.ID > 0 && firstOrder.PayTime != nil {
		startAt = *firstOrder.PayTime
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Printf("[recharge] first recharge gift v2 first order query failed orderNo=%s userID=%d err=%v", order.OrderNo, user.ID, err)
		return err
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		log.Printf("[recharge] first recharge gift v2 first order not found orderNo=%s userID=%d usePayAtAsStart=%s", order.OrderNo, user.ID, payAt.Format(time.RFC3339))
	}
	log.Printf("[recharge] first recharge gift v2 date context orderNo=%s userID=%d payAt=%s startAt=%s firstOrderID=%d firstOrderNo=%s firstOrderPayTime=%s",
		order.OrderNo, user.ID, payAt.Format(time.RFC3339), startAt.Format(time.RFC3339), firstOrder.ID, firstOrder.OrderNo, formatRechargeTime(firstOrder.PayTime))

	dayIndex, ok := firstRechargeGiftV2DayIndex(startAt, payAt)
	if !ok || dayIndex > len(cfg.Rates) {
		log.Printf("[recharge] first recharge gift v2 skip: invalid day orderNo=%s userID=%d dayIndex=%d ok=%t ratesLen=%d startAt=%s payAt=%s",
			order.OrderNo, user.ID, dayIndex, ok, len(cfg.Rates), startAt.Format(time.RFC3339), payAt.Format(time.RFC3339))
		return nil
	}

	var activityOrders []pojo.RechargeOrder
	if err = tx.Where("user_id = ? AND activity_type = 1 AND status = 1 AND pay_time IS NOT NULL", user.ID).
		Order("pay_time asc, id asc").
		Find(&activityOrders).Error; err != nil {
		log.Printf("[recharge] first recharge gift v2 activity orders query failed orderNo=%s userID=%d err=%v", order.OrderNo, user.ID, err)
		return err
	}
	completedMap := completedFirstRecharge3DayMap(startAt, activityOrders)
	log.Printf("[recharge] first recharge gift v2 activity orders orderNo=%s userID=%d dayIndex=%d completed=%v orderCount=%d", order.OrderNo, user.ID, dayIndex, completedMap, len(activityOrders))
	if hasMissedFirstRechargeDay(completedMap, dayIndex) {
		log.Printf("[recharge] first recharge gift v2 skip: missed previous day orderNo=%s userID=%d dayIndex=%d completed=%v", order.OrderNo, user.ID, dayIndex, completedMap)
		return nil
	}

	dayStart, dayEnd := naturalDayRange(payAt)
	var firstDayOrder pojo.RechargeOrder
	if err = tx.Where("user_id = ? AND activity_type = 1 AND status = 1 AND pay_time >= ? AND pay_time < ?", user.ID, dayStart, dayEnd).
		Order("pay_time asc, id asc").
		First(&firstDayOrder).Error; err == nil && firstDayOrder.ID > 0 && firstDayOrder.ID != order.ID {
		log.Printf("[recharge] first recharge gift v2 skip: same-day first order is different orderNo=%s userID=%d currentOrderID=%d firstDayOrderID=%d firstDayOrderNo=%s firstDayPayTime=%s dayStart=%s dayEnd=%s",
			order.OrderNo, user.ID, order.ID, firstDayOrder.ID, firstDayOrder.OrderNo, formatRechargeTime(firstDayOrder.PayTime), dayStart.Format(time.RFC3339), dayEnd.Format(time.RFC3339))
		return nil
	}
	if firstDayOrder.ID > 0 {
		log.Printf("[recharge] first recharge gift v2 same-day first order orderNo=%s userID=%d firstDayOrderID=%d firstDayOrderNo=%s", order.OrderNo, user.ID, firstDayOrder.ID, firstDayOrder.OrderNo)
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Printf("[recharge] first recharge gift v2 same-day order query failed orderNo=%s userID=%d err=%v", order.OrderNo, user.ID, err)
		return err
	}

	rate := cfg.Rates[dayIndex-1]
	giftAmount := calculateFirstRechargeGiftV2Amount(order.Amount, rate)
	if giftAmount <= 0 {
		log.Printf("[recharge] first recharge gift v2 skip: gift amount not positive orderNo=%s userID=%d amount=%.2f rate=%.2f giftAmount=%.2f", order.OrderNo, user.ID, order.Amount, rate, giftAmount)
		return nil
	}
	log.Printf("[recharge] first recharge gift v2 apply orderNo=%s userID=%d dayIndex=%d amount=%.2f rate=%.2f giftAmount=%.2f", order.OrderNo, user.ID, dayIndex, order.Amount, rate, giftAmount)
	return applyFirstRechargeGiftV2OrderGift(tx, order, user.ID, dayIndex, giftAmount, rate)
}

func resolveRechargeActivityType(db *gorm.DB, userID int64, req pojo.RechargeOrderAppReq, tablePrefix string) (int8, error) {
	code := strings.TrimSpace(req.ActivityCode)
	if code == "" {
		switch req.ActivityType {
		case 0:
			return 0, nil
		case rechargeActivityTypeFirstRecharge3Day:
			code = RechargeActivityCodeFirstRecharge3Day
		case rechargeActivityTypeTodayFirst:
			code = RechargeActivityCodeTodayFirst
		default:
			return 0, errors.New("activity_unavailable")
		}
	}
	switch code {
	case RechargeActivityCodeFirstRecharge3Day:
		cfg, err := getFirstRechargeGiftConfigV2(tablePrefix)
		if err != nil {
			log.Printf("[recharge] first_recharge_3day unavailable: parse first_recharge_gift_config_v2 failed userID=%d err=%v", userID, err)
			return 0, errors.New("first_recharge_activity_unavailable")
		}
		status := getFirstRecharge3DayPromotionStatus(db, userID, cfg, time.Now())
		if !status.Visible || !status.Selectable {
			return 0, errors.New("first_recharge_activity_unavailable")
		}
		return rechargeActivityTypeFirstRecharge3Day, nil
	case RechargeActivityCodeTodayFirst:
		hasTodayFirst := hasTodayFirstRechargeUsed(db, userID, time.Now())
		if hasTodayFirst {
			return 0, errors.New("today_first_recharge_activity_unavailable")
		}
		return rechargeActivityTypeTodayFirst, nil
	default:
		return 0, errors.New("activity_unavailable")
	}
}

func resolveRechargeOrderCreateActivityType(db *gorm.DB, userID int64, req pojo.RechargeOrderAppReq, tablePrefix string, forcedActivityType *int8) (int8, error) {
	if forcedActivityType != nil {
		return *forcedActivityType, nil
	}
	return resolveRechargeActivityType(db, userID, req, tablePrefix)
}

func shouldConfirmUnfinishedActivityCycleForRecharge(activityType int8) bool {
	return activityType == 0
}

func rechargeActivityCodeByType(activityType int8) string {
	switch activityType {
	case rechargeActivityTypeFirstRecharge3Day:
		return RechargeActivityCodeFirstRecharge3Day
	case rechargeActivityTypeTodayFirst:
		return RechargeActivityCodeTodayFirst
	case rechargeActivityTypeV2Gift:
		return RechargeActivityCodeV2Gift
	default:
		return ""
	}
}

func GetRechargePromotions(db *gorm.DB, userID int64, tablePrefix string) pojo.RechargePromotionStatusResp {
	now := time.Now()
	cfg, err := getFirstRechargeGiftConfigV2(tablePrefix)
	if err != nil {
		cfg = firstRechargeGiftConfigV2{}
	}
	todayRate := getTodayFirstRechargeRate(db)
	hasTodayFirst := hasTodayFirstRechargeUsed(db, userID, now)
	return pojo.RechargePromotionStatusResp{
		FirstRecharge3Day:  getFirstRecharge3DayPromotionStatus(db, userID, cfg, now),
		TodayFirstRecharge: pojo.RechargeTodayFirstPromotion{Visible: todayRate > 0 && !hasTodayFirst, Selectable: todayRate > 0 && !hasTodayFirst, ActivityCode: RechargeActivityCodeTodayFirst, Rate: todayRate},
	}
}

func getFirstRecharge3DayPromotionStatus(db *gorm.DB, userID int64, cfg firstRechargeGiftConfigV2, now time.Time) pojo.RechargeFirstRecharge3DayPromotion {
	resp := pojo.RechargeFirstRecharge3DayPromotion{
		ActivityCode: RechargeActivityCodeFirstRecharge3Day,
		Title:        "firstRechargeTitle",
		Rates:        buildFirstRecharge3DayRates(cfg, 0, nil),
	}
	if len(cfg.Rates) == 0 {
		return resp
	}

	var orders []pojo.RechargeOrder
	_ = db.Where("user_id = ? AND activity_type = 1 AND status = 1 AND pay_time IS NOT NULL", userID).
		Order("pay_time asc, id asc").
		Find(&orders).Error

	if len(orders) == 0 || orders[0].PayTime == nil {
		todayStart, _ := naturalDayRange(now)
		resp.Visible = true
		resp.Selectable = true
		resp.CurrentDay = 1
		resp.ValidFrom = todayStart.Format(rechargePromotionTimeFormat)
		resp.ValidTo = todayStart.AddDate(0, 0, 3).Add(-time.Second).Format(rechargePromotionTimeFormat)
		resp.TodayRate = cfg.Rates[0]
		resp.Rates = buildFirstRecharge3DayRates(cfg, 1, map[int]bool{})
		return resp
	}

	startAt := *orders[0].PayTime
	startDay, _ := naturalDayRange(startAt)
	dayIndex, available := firstRechargeGiftV2DayIndex(startAt, now)
	completed := completedFirstRecharge3DayMap(startAt, orders)
	resp.ValidFrom = startDay.Format(rechargePromotionTimeFormat)
	resp.ValidTo = startDay.AddDate(0, 0, 3).Add(-time.Second).Format(rechargePromotionTimeFormat)
	if !available || hasMissedFirstRechargeDay(completed, dayIndex) {
		resp.Rates = buildFirstRecharge3DayRates(cfg, dayIndex, completed)
		return resp
	}

	todayDone := completed[dayIndex]
	resp.Visible = !todayDone
	resp.Selectable = !todayDone
	resp.CurrentDay = dayIndex
	resp.TodayRate = cfg.Rates[dayIndex-1]
	resp.Rates = buildFirstRecharge3DayRates(cfg, dayIndex, completed)
	return resp
}

func completedFirstRecharge3DayMap(startAt time.Time, orders []pojo.RechargeOrder) map[int]bool {
	completed := map[int]bool{}
	for _, order := range orders {
		if order.PayTime == nil {
			continue
		}
		day, ok := firstRechargeGiftV2DayIndex(startAt, *order.PayTime)
		if ok {
			completed[day] = true
		}
	}
	return completed
}

func hasMissedFirstRechargeDay(completed map[int]bool, currentDay int) bool {
	for day := 1; day < currentDay; day++ {
		if !completed[day] {
			return true
		}
	}
	return false
}

func buildFirstRecharge3DayRates(cfg firstRechargeGiftConfigV2, currentDay int, completed map[int]bool) []pojo.RechargePromotionDayRate {
	rates := make([]pojo.RechargePromotionDayRate, 0, len(cfg.Rates))
	for idx, rate := range cfg.Rates {
		day := idx + 1
		status := "pending"
		if completed != nil && completed[day] {
			status = "done"
		} else if currentDay > 0 {
			if day < currentDay || hasMissedFirstRechargeDay(completed, currentDay) {
				status = "expired"
			} else if day == currentDay {
				status = "available"
			}
		}
		rates = append(rates, pojo.RechargePromotionDayRate{Day: day, Rate: rate, Status: status})
	}
	return rates
}

func hasTodayFirstRechargeUsed(db *gorm.DB, userID int64, now time.Time) bool {
	var todayCount int64
	db.Model(&pojo.RechargeOrder{}).
		Where("user_id = ? AND activity_type = 2 AND status = 1 AND created_at >= ?", userID, now.Add(-24*time.Hour)).
		Count(&todayCount)
	return todayCount > 0
}

func getTodayFirstRechargeRate(db *gorm.DB) float64 {
	var cfg pojo.SysConfig
	db.Where("config_key = ?", "today_first_recharge_gift").First(&cfg)
	rate, err := strconv.ParseFloat(strings.TrimSpace(cfg.ConfigValue), 64)
	if cfg.ID == 0 || err != nil || rate <= 0 {
		return 0
	}
	return rate
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

func getFirstRechargeGiftConfigV2(tablePrefix string) (firstRechargeGiftConfigV2, error) {
	defaultValue := ""
	val := utils.GetStringCache(tablePrefix, "first_recharge_gift_config_v2", &defaultValue)
	if val == nil {
		return firstRechargeGiftConfigV2{}, fmt.Errorf("config is empty")
	}
	log.Printf("[recharge] first recharge gift v2 raw config tablePrefix=%q value=%q", tablePrefix, *val)
	return parseFirstRechargeGiftConfigV2(*val)
}

func parseFirstRechargeGiftConfigV2(raw string) (firstRechargeGiftConfigV2, error) {
	text := strings.TrimSpace(raw)
	if text == "" {
		return firstRechargeGiftConfigV2{}, fmt.Errorf("config is empty")
	}
	text = strings.NewReplacer("｜", "|", "，", "|", ",", "|", "％", "%").Replace(text)
	text = strings.ReplaceAll(text, "%", "")
	parts := strings.Split(text, "|")
	if len(parts) != 3 {
		return firstRechargeGiftConfigV2{}, fmt.Errorf("invalid config format: %s", raw)
	}

	rates := make([]float64, 0, 3)
	for _, part := range parts {
		rate, err := strconv.ParseFloat(strings.TrimSpace(part), 64)
		if err != nil || rate <= 0 {
			return firstRechargeGiftConfigV2{}, fmt.Errorf("invalid rate: %s", part)
		}
		rates = append(rates, rate)
	}
	return firstRechargeGiftConfigV2{Rates: rates}, nil
}

func firstRechargeGiftV2DayIndex(startAt time.Time, payAt time.Time) (int, bool) {
	startAt = startAt.In(payAt.Location())
	startDay, _ := naturalDayRange(startAt)
	payDay, _ := naturalDayRange(payAt)
	if payDay.Before(startDay) {
		return 0, false
	}
	dayIndex := int(payDay.Sub(startDay).Hours()/24) + 1
	if dayIndex < 1 || dayIndex > 3 {
		return 0, false
	}
	return dayIndex, true
}

func naturalDayRange(at time.Time) (time.Time, time.Time) {
	start := time.Date(at.Year(), at.Month(), at.Day(), 0, 0, 0, 0, at.Location())
	return start, start.Add(24 * time.Hour)
}

func calculateFirstRechargeGiftV2Amount(orderAmount float64, rate float64) float64 {
	return utils.Truncate2(utils.ToMoney(orderAmount).Multiply(rate / 100).ToDollars())
}

func rechargeGiftBaseAmount(order pojo.RechargeOrder) float64 {
	if order.CreditAmount != nil && *order.CreditAmount > 0 {
		return utils.Truncate2(*order.CreditAmount)
	}
	return utils.Truncate2(order.Amount)
}

// RechargeV2GiftRatesConfigKey v2充值赠送比例配置键（sys_config，逗号分隔：首充,二充,三充,第4次及以后）。
const RechargeV2GiftRatesConfigKey = "recharge_v2_gift_rates"

// defaultRechargeV2GiftRates 默认 v2 充值赠送比例(%)：首充18 / 二充10 / 三充10 / 第4次及以后10。
var defaultRechargeV2GiftRates = [4]float64{18, 10, 10, 10}

// GetRechargeV2GiftRates 读取 v2 充值赠送比例 [首充, 二充, 三充, 第4次及以后]；
// 配置缺失自动初始化默认值，解析失败回退默认。
func GetRechargeV2GiftRates(db *gorm.DB) [4]float64 {
	rates := defaultRechargeV2GiftRates
	if db == nil {
		return rates
	}
	var cfg pojo.SysConfig
	db.Where("config_key = ?", RechargeV2GiftRatesConfigKey).First(&cfg)
	if cfg.ID == 0 {
		_ = db.Create(&pojo.SysConfig{
			ConfigKey:   RechargeV2GiftRatesConfigKey,
			ConfigValue: fmt.Sprintf("%g,%g,%g,%g", rates[0], rates[1], rates[2], rates[3]),
			ConfigDesc:  "v2充值赠送比例(%)，逗号分隔：首充,二充,三充,第4次及以后",
		}).Error
		return rates
	}
	if parsed, ok := parseRechargeV2GiftRates(cfg.ConfigValue); ok {
		return parsed
	}
	return rates
}

func parseRechargeV2GiftRates(raw string) ([4]float64, bool) {
	parts := strings.Split(strings.TrimSpace(raw), ",")
	if len(parts) < 4 {
		return [4]float64{}, false
	}
	var rates [4]float64
	for i := 0; i < 4; i++ {
		v, err := strconv.ParseFloat(strings.TrimSpace(parts[i]), 64)
		if err != nil || v < 0 {
			return [4]float64{}, false
		}
		rates[i] = v
	}
	return rates, true
}

// rechargeV2GiftRateByNumber 按"该用户第几次充值"返回赠送比例(%)：1=首充 2=二充 3=三充 ≥4=后续。
func rechargeV2GiftRateByNumber(db *gorm.DB, rechargeNumber int) float64 {
	rates := GetRechargeV2GiftRates(db)
	idx := rechargeNumber - 1
	if idx < 0 {
		idx = 0
	}
	if idx > 3 {
		idx = 3
	}
	return rates[idx]
}

// RechargeV2GiftMinAmount v2 充值赠送门槛（满此金额才赠送），供 API 预览使用。
func RechargeV2GiftMinAmount() float64 {
	return rechargeV2GiftMinAmount
}

// RechargeV2NextRechargeNumber 返回该用户「下一次」充值的次序（已成功次数+1），用于赠送预览。
func RechargeV2NextRechargeNumber(db *gorm.DB, userID int64) int {
	if db == nil || userID <= 0 {
		return 1
	}
	var cnt int64
	_ = db.Model(&pojo.RechargeOrder{}).Where("user_id = ? AND status = ?", userID, 1).Count(&cnt).Error
	return int(cnt) + 1
}

// rechargeV2RechargeNumber 返回该用户的成功充值次数（含本次，本次在回调中已置 status=1）。
func rechargeV2RechargeNumber(tx *gorm.DB, userID int64) int {
	if tx == nil || userID <= 0 {
		return 1
	}
	var cnt int64
	_ = tx.Model(&pojo.RechargeOrder{}).Where("user_id = ? AND status = ?", userID, 1).Count(&cnt).Error
	if cnt < 1 {
		return 1
	}
	return int(cnt)
}

// rechargeV2RechargeKindLabel 充值次数的展示标签。
func rechargeV2RechargeKindLabel(rechargeNumber int) string {
	switch rechargeNumber {
	case 1:
		return "首充"
	case 2:
		return "二充"
	case 3:
		return "三充"
	default:
		return fmt.Sprintf("第%d次充值", rechargeNumber)
	}
}

func calculateRechargeV2GiftAmount(db *gorm.DB, orderAmount float64, rechargeNumber int) float64 {
	if orderAmount < rechargeV2GiftMinAmount {
		return 0
	}
	rate := rechargeV2GiftRateByNumber(db, rechargeNumber)
	if rate <= 0 {
		return 0
	}
	return utils.Truncate2(utils.ToMoney(orderAmount).Multiply(rate / 100).ToDollars())
}

func isRechargeV2FirstRechargeAmount(rechargeAmount float64) bool {
	return rechargeAmount <= 0
}

func CheckRechargeV2IsFirst(db *gorm.DB, userID int64) (bool, error) {
	var user pojo.TgUser
	if err := db.Select("id", "recharge_amount").Where("id = ?", userID).First(&user).Error; err != nil {
		return false, err
	}
	if user.ID == 0 {
		return false, errors.New("user_not_found")
	}
	return isRechargeV2FirstRechargeAmount(user.RechargeAmount), nil
}

func applyRechargeV2Gift(tx *gorm.DB, order pojo.RechargeOrder, user pojo.TgUser) error {
	rechargeNumber := rechargeV2RechargeNumber(tx, user.ID)
	rate := rechargeV2GiftRateByNumber(tx, rechargeNumber)
	giftBaseAmount := rechargeGiftBaseAmount(order)
	giftAmount := calculateRechargeV2GiftAmount(tx, giftBaseAmount, rechargeNumber)
	if giftAmount <= 0 {
		log.Printf("[recharge] v2 gift skip: non-positive gift orderNo=%s userID=%d amount=%.2f giftBaseAmount=%.2f rechargeNumber=%d rate=%.2f giftAmount=%.2f",
			order.OrderNo, user.ID, order.Amount, giftBaseAmount, rechargeNumber, rate, giftAmount)
		return nil
	}

	awardUni := fmt.Sprintf("recharge_v2_gift_%s", order.OrderNo)
	var existing pojo.CashHistory
	if err := tx.Where("user_id = ? AND award_uni = ?", user.ID, awardUni).First(&existing).Error; err == nil && existing.ID > 0 {
		log.Printf("[recharge] v2 gift skip: cash history exists orderNo=%s userID=%d awardUni=%s existingID=%d", order.OrderNo, user.ID, awardUni, existing.ID)
		return nil
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	var lockedUser pojo.TgUser
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", user.ID).First(&lockedUser).Error; err != nil {
		return err
	}
	if err := tx.Model(&pojo.TgUser{}).Where("id = ?", user.ID).Updates(map[string]any{
		"balance":     gorm.Expr("balance + ?", giftAmount),
		"gift_amount": gorm.Expr("gift_amount + ?", giftAmount),
		"gift_total":  gorm.Expr("gift_total + ?", giftAmount),
	}).Error; err != nil {
		return err
	}
	rechargeKind := rechargeV2RechargeKindLabel(rechargeNumber)
	desc := fmt.Sprintf("v2充值赠送，%s，订单%s，充值金额%.2f，赠送比例%.2f%%，赠送%.2f", rechargeKind, order.OrderNo, giftBaseAmount, rate, giftAmount)
	history := pojo.CashHistory{
		UserId:          user.ID,
		AwardUni:        awardUni,
		Amount:          giftAmount,
		StartAmount:     lockedUser.Balance,
		EndAmount:       utils.Truncate2(lockedUser.Balance + giftAmount),
		CashMark:        "v2充值赠送",
		CashDesc:        desc,
		Type:            pojo.CashHistoryTypeFirstRechargeGift,
		IsGift:          1,
		FromUserId:      0,
		SourceChannelID: order.SourceChannelID,
	}
	if err := tx.Create(&history).Error; err != nil {
		return err
	}
	if err := addRechargeOrderBonusAmount(tx, order.ID, giftAmount); err != nil {
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

func applyFirstRechargeGiftV2OrderGift(tx *gorm.DB, order pojo.RechargeOrder, userID int64, dayIndex int, giftAmount float64, rate float64) error {
	if giftAmount <= 0 {
		log.Printf("[recharge] first recharge gift v2 order gift skip: non-positive gift orderNo=%s userID=%d dayIndex=%d giftAmount=%.2f", order.OrderNo, userID, dayIndex, giftAmount)
		return nil
	}

	awardUni := fmt.Sprintf("first_recharge_gift_v2_%s_%d", order.OrderNo, dayIndex)
	var existing pojo.CashHistory
	if err := tx.Where("user_id = ? AND award_uni = ?", userID, awardUni).First(&existing).Error; err == nil && existing.ID > 0 {
		log.Printf("[recharge] first recharge gift v2 order gift skip: cash history exists orderNo=%s userID=%d awardUni=%s existingID=%d", order.OrderNo, userID, awardUni, existing.ID)
		return nil
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Printf("[recharge] first recharge gift v2 order gift existing query failed orderNo=%s userID=%d awardUni=%s err=%v", order.OrderNo, userID, awardUni, err)
		return err
	}

	var user pojo.TgUser
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", userID).First(&user).Error; err != nil {
		log.Printf("[recharge] first recharge gift v2 order gift user lock failed orderNo=%s userID=%d err=%v", order.OrderNo, userID, err)
		return err
	}
	log.Printf("[recharge] first recharge gift v2 order gift updating user orderNo=%s userID=%d awardUni=%s startBalance=%.2f startGiftAmount=%.2f giftAmount=%.2f",
		order.OrderNo, userID, awardUni, user.Balance, user.GiftAmount, giftAmount)

	if err := tx.Model(&pojo.TgUser{}).Where("id = ?", userID).Updates(map[string]any{
		"balance":     gorm.Expr("balance + ?", giftAmount),
		"gift_amount": gorm.Expr("gift_amount + ?", giftAmount),
		"gift_total":  gorm.Expr("gift_total + ?", giftAmount),
	}).Error; err != nil {
		log.Printf("[recharge] first recharge gift v2 order gift update user failed orderNo=%s userID=%d awardUni=%s err=%v", order.OrderNo, userID, awardUni, err)
		return err
	}
	if err := AddUserWithdrawRestrictedBalance(tx, user, giftAmount, 0); err != nil {
		log.Printf("[recharge] first recharge gift v2 order gift add withdraw limit failed orderNo=%s userID=%d awardUni=%s err=%v", order.OrderNo, userID, awardUni, err)
		return err
	}

	desc := fmt.Sprintf("首充活动V2第%d天首笔赠送，订单%s，充值金额%.2f，赠送比例%.2f%%，赠送%.2f", dayIndex, order.OrderNo, order.Amount, rate, giftAmount)
	history := pojo.CashHistory{
		UserId:          userID,
		AwardUni:        awardUni,
		Amount:          giftAmount,
		StartAmount:     user.Balance,
		EndAmount:       utils.Truncate2(user.Balance + giftAmount),
		CashMark:        "首充活动赠送",
		CashDesc:        desc,
		Type:            pojo.CashHistoryTypeFirstRechargeGift,
		IsGift:          1,
		FromUserId:      0,
		SourceChannelID: order.SourceChannelID,
	}
	if err := tx.Create(&history).Error; err != nil {
		log.Printf("[recharge] first recharge gift v2 order gift create cash history failed orderNo=%s userID=%d awardUni=%s err=%v", order.OrderNo, userID, awardUni, err)
		return err
	}
	if err := addRechargeOrderBonusAmount(tx, order.ID, giftAmount); err != nil {
		log.Printf("[recharge] first recharge gift v2 order gift update order bonus failed orderNo=%s userID=%d awardUni=%s err=%v", order.OrderNo, userID, awardUni, err)
		return err
	}
	log.Printf("[recharge] first recharge gift v2 order gift cash history created orderNo=%s userID=%d awardUni=%s cashHistoryID=%d giftAmount=%.2f",
		order.OrderNo, userID, awardUni, history.ID, giftAmount)

	if err := CreatePlatformProfitLedgerIfAbsent(tx, pojo.PlatformProfitLedger{
		TenantId:        order.TenantId,
		UserId:          userID,
		SourceChannelID: order.SourceChannelID,
		SourceType:      pojo.PlatformProfitSourceRechargeGift,
		SourceId:        awardUni,
		IncomeAmount:    0,
		ExpenseAmount:   giftAmount,
		Remark:          desc,
	}); err != nil {
		log.Printf("[recharge] first recharge gift v2 order gift create platform ledger failed orderNo=%s userID=%d awardUni=%s err=%v", order.OrderNo, userID, awardUni, err)
		return err
	}
	log.Printf("[recharge] first recharge gift v2 order gift done orderNo=%s userID=%d awardUni=%s giftAmount=%.2f", order.OrderNo, userID, awardUni, giftAmount)
	return nil
}

func formatRechargeActivityType(activityType *int8) string {
	if activityType == nil {
		return "<nil>"
	}
	switch *activityType {
	case rechargeActivityTypeFirstRecharge3Day:
		return "1(first_recharge_3day)"
	case rechargeActivityTypeTodayFirst:
		return "2(today_first_recharge)"
	case rechargeActivityTypeV2Gift:
		return "3(recharge_v2_gift)"
	}
	return strconv.FormatInt(int64(*activityType), 10)
}

func formatRechargeTime(value *time.Time) string {
	if value == nil {
		return "<nil>"
	}
	return value.Format(time.RFC3339)
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
		Type:            pojo.CashHistoryTypeTodayFirstRechargeGift,
		IsGift:          1,
		FromUserId:      0,
		SourceChannelID: order.SourceChannelID,
	}
	if err = tx.Create(&history).Error; err != nil {
		return err
	}
	if err = addRechargeOrderBonusAmount(tx, order.ID, giftAmount); err != nil {
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
		Type:            pojo.CashHistoryTypeFirstRechargeGift,
		IsGift:          1,
		FromUserId:      0,
		SourceChannelID: order.SourceChannelID,
	}
	if err := tx.Create(&history).Error; err != nil {
		return err
	}
	if err := addRechargeOrderBonusAmount(tx, order.ID, installment.GiftAmount); err != nil {
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

func applyFirstLevelRechargeRebate(tx *gorm.DB, order pojo.RechargeOrder, subUser pojo.TgUser, tablePrefix string, now time.Time) error {
	if subUser.ParentID == nil || *subUser.ParentID <= 0 {
		return nil
	}

	rate := getFirstLevelRechargeRebateRate(tablePrefix)
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

	idempotencyKey := fmt.Sprintf("%d:%s:%d", pojo.TgUserRebateSourceTypeRecharge, order.OrderNo, parentID)
	var existing pojo.TgUserRebateRecord
	if err := tx.Where("idempotency_key = ?", idempotencyKey).First(&existing).Error; err == nil && existing.ID > 0 {
		return nil
	}

	currency := strings.TrimSpace(order.Currency)
	if currency == "" {
		currency = "USDT"
	}
	remark := "level_1_rebate"
	record := pojo.TgUserRebateRecord{
		TenantId:        &order.TenantId,
		SubUserId:       subUser.ID,
		ParentUserId:    parentID,
		SourceChannelID: order.SourceChannelID,
		SourceType:      pojo.TgUserRebateSourceTypeRecharge,
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
// hasFirst: 当前是否不可参加首充活动（api 层会取反返回可参加状态）
// hasTodayFirst: 24小时内是否已参加过今日首充活动（activity_type=2 且 status=1）
func CheckActivityStatus(db *gorm.DB, userID int64) (hasFirst bool, hasTodayFirst bool) {
	promotions := GetRechargePromotions(db, userID, utils.GetDbPrefix(db))
	hasFirst = !promotions.FirstRecharge3Day.Visible || !promotions.FirstRecharge3Day.Selectable
	hasTodayFirst = !promotions.TodayFirstRecharge.Visible || !promotions.TodayFirstRecharge.Selectable
	return
}

func getFirstLevelRechargeRebateRate(tablePrefix string) float64 {
	defaultValue := "0"
	val := utils.GetStringCache(tablePrefix, "first_level_rebate", &defaultValue)
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
