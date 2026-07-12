package repository

import (
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/utils"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	UsdtTrc20EnabledConfigKey        = "usdt_trc20_enabled"
	UsdtTrc20ReceiveAddressConfigKey = "usdt_trc20_receive_address"
	UsdtTrc20PlatformRateConfigKey   = "usdt_trc20_platform_rate"

	usdtTrc20DefaultAPIBaseURL      = "https://api.trongrid.io"
	usdtTrc20DefaultContractAddress = "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t"
	usdtTrc20DefaultExpireMinutes   = 30
	usdtTrc20DefaultScanSeconds     = 10
	usdtTrc20MicroScale             = int64(1000000)
	usdtTrc20TailMaxMicro           = int64(9999)
)

type UsdtTrc20RuntimeConfig struct {
	Enabled             bool
	APIBaseURL          string
	APIKey              string
	ContractAddress     string
	ReceiveAddress      string
	PlatformRate        string
	OrderExpireMinutes  int
	ScanIntervalSeconds int
}

func LoadUsdtTrc20RuntimeConfig(tablePrefix string) (UsdtTrc20RuntimeConfig, error) {
	global := utils.GlobalConfig.Pay.UsdtTrc20
	cfg := UsdtTrc20RuntimeConfig{
		Enabled:             false,
		APIBaseURL:          strings.TrimSpace(global.APIBaseURL),
		APIKey:              strings.TrimSpace(global.APIKey),
		ContractAddress:     strings.TrimSpace(global.ContractAddress),
		ReceiveAddress:      strings.TrimSpace(global.DefaultReceiveAddress),
		OrderExpireMinutes:  global.OrderExpireMinutes,
		ScanIntervalSeconds: global.ScanIntervalSeconds,
	}
	if cfg.APIBaseURL == "" {
		cfg.APIBaseURL = usdtTrc20DefaultAPIBaseURL
	}
	if cfg.ContractAddress == "" {
		cfg.ContractAddress = usdtTrc20DefaultContractAddress
	}
	if cfg.OrderExpireMinutes <= 0 {
		cfg.OrderExpireMinutes = usdtTrc20DefaultExpireMinutes
	}
	if cfg.ScanIntervalSeconds <= 0 {
		cfg.ScanIntervalSeconds = usdtTrc20DefaultScanSeconds
	}

	if !global.Enabled {
		return cfg, nil
	}
	cfg.Enabled = parseUsdtBoolConfig(loadUsdtTenantConfig(tablePrefix, UsdtTrc20EnabledConfigKey, "0"))
	cfg.ReceiveAddress = strings.TrimSpace(loadUsdtTenantConfig(tablePrefix, UsdtTrc20ReceiveAddressConfigKey, cfg.ReceiveAddress))
	cfg.PlatformRate = strings.TrimSpace(loadUsdtTenantConfig(tablePrefix, UsdtTrc20PlatformRateConfigKey, ""))
	return cfg, nil
}

func validateUsdtTrc20RuntimeConfig(tablePrefix string) (UsdtTrc20RuntimeConfig, error) {
	cfg, err := LoadUsdtTrc20RuntimeConfig(tablePrefix)
	if err != nil {
		return cfg, err
	}
	if !cfg.Enabled {
		return cfg, errors.New("usdt_trc20_disabled")
	}
	if cfg.ReceiveAddress == "" {
		return cfg, errors.New("usdt_trc20_receive_address_required")
	}
	if _, err = parsePositiveRat(cfg.PlatformRate); err != nil {
		return cfg, errors.New("usdt_trc20_platform_rate_invalid")
	}
	return cfg, nil
}

func GetCryptoRechargeOptions(db *gorm.DB, tablePrefix string, amount float64) (pojo.CryptoRechargeOptionsBack, error) {
	amount = floorRechargeAmount(amount)
	result := pojo.CryptoRechargeOptionsBack{PlatformAmount: amount}
	if amount <= 0 {
		return result, errors.New("recharge_amount_positive")
	}
	cfg, err := validateUsdtTrc20RuntimeConfig(tablePrefix)
	if err != nil {
		return result, err
	}
	baseAmountMicro, err := calculateUsdtBaseAmountMicro(amount, cfg.PlatformRate)
	if err != nil {
		return result, err
	}
	result.Options = []pojo.CryptoRechargeOptionBack{
		{
			Network:         pojo.CryptoRechargeNetworkTRC20,
			Token:           pojo.CryptoRechargeTokenUSDT,
			ContractAddress: cfg.ContractAddress,
			ReceiveAddress:  cfg.ReceiveAddress,
			EstimatedAmount: FormatUsdtMicroAmount(baseAmountMicro),
			PlatformRate:    cfg.PlatformRate,
			PlatformAmount:  amount,
		},
	}
	return result, nil
}

func AppCreateCryptoRechargeOrder(db *gorm.DB, userID int64, req pojo.CryptoRechargeOrderReq, tablePrefix string) (pojo.RechargeOrderAppBack, error) {
	return createCryptoRechargeOrder(db, userID, req, tablePrefix, rechargeV2MinAmount)
}

func AdminCreateCryptoRechargeOrder(db *gorm.DB, userID int64, req pojo.CryptoRechargeOrderReq, tablePrefix string) (pojo.RechargeOrderAppBack, error) {
	return createCryptoRechargeOrder(db, userID, req, tablePrefix, AdminRechargeV2MinAmount)
}

func createCryptoRechargeOrder(db *gorm.DB, userID int64, req pojo.CryptoRechargeOrderReq, tablePrefix string, minAmount float64) (pojo.RechargeOrderAppBack, error) {
	req.Network = strings.ToUpper(strings.TrimSpace(req.Network))
	req.Token = strings.ToUpper(strings.TrimSpace(req.Token))
	req.MerchantOrderNo = strings.TrimSpace(req.MerchantOrderNo)
	req.ActivityCode = strings.TrimSpace(req.ActivityCode)
	req.Amount = floorRechargeAmount(req.Amount)
	var result pojo.RechargeOrderAppBack
	if req.Amount <= 0 {
		return result, errors.New("recharge_amount_positive")
	}
	if req.Network != "" && req.Network != pojo.CryptoRechargeNetworkTRC20 {
		return result, errors.New("unsupported_crypto_network")
	}
	if req.Token != "" && req.Token != pojo.CryptoRechargeTokenUSDT {
		return result, errors.New("unsupported_crypto_token")
	}
	if minAmount > 0 && req.Amount < minAmount {
		return result, errors.New(utils.I18nMessage("recharge_v2_min_amount", map[string]interface{}{
			"min": formatRechargeMinAmount(minAmount),
		}))
	}

	cfg, err := validateUsdtTrc20RuntimeConfig(tablePrefix)
	if err != nil {
		return result, err
	}

	lockKey := fmt.Sprintf("usdt_recharge_create:%s:%s", tablePrefix, cfg.ReceiveAddress)
	locked, lockErr := utils.AcquireLock(lockKey, 10*time.Second)
	if lockErr != nil {
		return result, lockErr
	}
	if !locked {
		return result, errors.New("usdt_recharge_create_busy")
	}
	defer utils.ReleaseLock(lockKey)

	err = db.Transaction(func(tx *gorm.DB) error {
		var tgUser pojo.TgUser
		if err := tx.Where("id = ?", userID).First(&tgUser).Error; err != nil || tgUser.ID == 0 {
			return errors.New("user_not_found")
		}
		if tgUser.Status != 1 {
			return errors.New("user_disabled_contact_admin")
		}
		activityType := rechargeActivityTypeV2Gift
		if shouldConfirmUnfinishedActivityCycleForRecharge(activityType) && !req.ConfirmUnfinishedActivityCycle {
			activeCycle, cycleErr := GetActiveWithdrawActivityCycle(tx, userID)
			if cycleErr != nil {
				return cycleErr
			}
			if activeCycle.ID > 0 {
				if CanBypassWithdrawActivityCycleByBalance(tgUser.Balance, activeCycle) {
					if endErr := EndWithdrawActivityCycle(tx, userID, pojo.WithdrawActivityCycleEndReasonBalanceBelowLimit); endErr != nil {
						return endErr
					}
					if resetErr := ResetUserWithdrawLimitAfterActivityEnd(tx, userID, tgUser.Balance, false); resetErr != nil {
						return resetErr
					}
				} else {
					result.NeedConfirmUnfinishedActivityCycle = true
					result.ActiveActivityMultiplier = activeCycle.Multiplier
					return nil
				}
			}
		}

		baseAmountMicro, calcErr := calculateUsdtBaseAmountMicro(req.Amount, cfg.PlatformRate)
		if calcErr != nil {
			return calcErr
		}
		expectedAmountMicro, amountErr := allocateUsdtExpectedAmountMicro(tx, cfg.ReceiveAddress, baseAmountMicro)
		if amountErr != nil {
			return amountErr
		}
		expectedAmount := FormatUsdtMicroAmount(expectedAmountMicro)

		orderNo := buildRechargeOrderNo()
		expireTime := time.Now().Add(time.Duration(cfg.OrderExpireMinutes) * time.Minute)
		merchantOrderNo := nullableUsdtString(req.MerchantOrderNo)
		payMethod := pojo.CryptoRechargePayMethodTRC20
		provider := "native_usdt_trc20"
		extra := map[string]any{
			"cryptoPayment": map[string]any{
				"network":             pojo.CryptoRechargeNetworkTRC20,
				"token":               pojo.CryptoRechargeTokenUSDT,
				"contractAddress":     cfg.ContractAddress,
				"receiveAddress":      cfg.ReceiveAddress,
				"platformRate":        cfg.PlatformRate,
				"baseAmountMicro":     baseAmountMicro,
				"expectedAmountMicro": expectedAmountMicro,
				"expectedAmount":      expectedAmount,
			},
		}
		extraBytes, _ := json.Marshal(extra)
		extraStr := string(extraBytes)

		order := pojo.RechargeOrder{
			TenantId:        tgUser.TenantId,
			UserId:          userID,
			SourceChannelID: tgUser.SourceChannelID,
			OrderNo:         orderNo,
			MerchantOrderNo: merchantOrderNo,
			Channel:         pojo.CryptoRechargeChannelUSDTTRC20,
			PayMethod:       &payMethod,
			Currency:        pojo.CryptoRechargeCurrencyUSDT,
			Amount:          req.Amount,
			Fee:             0,
			NetAmount:       UsdtMicroAmountToFloat(expectedAmountMicro),
			BonusAmount:     0,
			WalletType:      pojo.RechargeWalletTypeBalance,
			Status:          0,
			ExpireTime:      &expireTime,
			Provider:        &provider,
			Extra:           &extraStr,
			ActivityType:    &activityType,
		}
		if err := tx.Create(&order).Error; err != nil {
			return err
		}

		usdtOrder := pojo.UsdtRechargeOrder{
			TenantId:            tgUser.TenantId,
			UserId:              userID,
			SourceChannelID:     tgUser.SourceChannelID,
			OrderNo:             orderNo,
			Network:             pojo.CryptoRechargeNetworkTRC20,
			Token:               pojo.CryptoRechargeTokenUSDT,
			ContractAddress:     cfg.ContractAddress,
			ReceiveAddress:      cfg.ReceiveAddress,
			PlatformAmount:      req.Amount,
			PlatformRate:        cfg.PlatformRate,
			BaseAmountMicro:     baseAmountMicro,
			ExpectedAmountMicro: expectedAmountMicro,
			ExpectedAmount:      expectedAmount,
			Status:              pojo.UsdtRechargeStatusPending,
			ExpireTime:          expireTime,
		}
		if err := tx.Create(&usdtOrder).Error; err != nil {
			return err
		}

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
			CryptoPayment: &pojo.CryptoPaymentBack{
				Network:         pojo.CryptoRechargeNetworkTRC20,
				Token:           pojo.CryptoRechargeTokenUSDT,
				ContractAddress: cfg.ContractAddress,
				ReceiveAddress:  cfg.ReceiveAddress,
				ExpectedAmount:  expectedAmount,
				ExpireTime:      expireTime.Format(time.RFC3339),
				QRContent:       cfg.ReceiveAddress,
				PlatformRate:    cfg.PlatformRate,
			},
		}
		return nil
	})
	return result, err
}

func loadUsdtTenantConfig(tablePrefix string, key string, defaultValue string) string {
	value := defaultValue
	cached := utils.GetStringCache(tablePrefix, key, &value)
	if cached == nil {
		return strings.TrimSpace(defaultValue)
	}
	return strings.TrimSpace(*cached)
}

func parseUsdtBoolConfig(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "on", "enabled":
		return true
	default:
		return false
	}
}

func parsePositiveRat(value string) (*big.Rat, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, errors.New("empty_decimal")
	}
	rat, ok := new(big.Rat).SetString(value)
	if !ok || rat.Sign() <= 0 {
		return nil, errors.New("invalid_decimal")
	}
	return rat, nil
}

func calculateUsdtBaseAmountMicro(platformAmount float64, platformRate string) (int64, error) {
	amountRat, err := parsePositiveRat(strconv.FormatFloat(platformAmount, 'f', 2, 64))
	if err != nil {
		return 0, err
	}
	rateRat, err := parsePositiveRat(platformRate)
	if err != nil {
		return 0, err
	}
	scaled := new(big.Rat).Mul(amountRat, big.NewRat(usdtTrc20MicroScale, 1))
	scaled.Quo(scaled, rateRat)
	return ceilPositiveRatToInt64(scaled)
}

func ceilPositiveRatToInt64(rat *big.Rat) (int64, error) {
	if rat == nil || rat.Sign() <= 0 {
		return 0, errors.New("invalid_decimal")
	}
	num := new(big.Int).Set(rat.Num())
	den := new(big.Int).Set(rat.Denom())
	quotient, remainder := new(big.Int), new(big.Int)
	quotient.QuoRem(num, den, remainder)
	if remainder.Sign() > 0 {
		quotient.Add(quotient, big.NewInt(1))
	}
	if !quotient.IsInt64() {
		return 0, errors.New("amount_overflow")
	}
	return quotient.Int64(), nil
}

func allocateUsdtExpectedAmountMicro(tx *gorm.DB, receiveAddress string, baseAmountMicro int64) (int64, error) {
	if baseAmountMicro <= 0 {
		return 0, errors.New("usdt_amount_invalid")
	}
	seed := time.Now().UnixNano()
	for i := int64(0); i < usdtTrc20TailMaxMicro; i++ {
		tail := (seed + i*7919) % usdtTrc20TailMaxMicro
		if tail < 0 {
			tail = -tail
		}
		tail++
		expected := baseAmountMicro + tail
		var count int64
		if err := tx.Model(&pojo.UsdtRechargeOrder{}).
			Where("receive_address = ? AND expected_amount_micro = ? AND status = ?", receiveAddress, expected, pojo.UsdtRechargeStatusPending).
			Count(&count).Error; err != nil {
			return 0, err
		}
		if count == 0 {
			return expected, nil
		}
	}
	return 0, errors.New("usdt_expected_amount_unavailable")
}

func FormatUsdtMicroAmount(amountMicro int64) string {
	if amountMicro < 0 {
		return "-" + FormatUsdtMicroAmount(-amountMicro)
	}
	return fmt.Sprintf("%d.%06d", amountMicro/usdtTrc20MicroScale, amountMicro%usdtTrc20MicroScale)
}

func UsdtMicroAmountToFloat(amountMicro int64) float64 {
	value, _ := strconv.ParseFloat(FormatUsdtMicroAmount(amountMicro), 64)
	return value
}

func nullableUsdtString(value string) *string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return &value
}

func GetCryptoRechargeOrderStatus(db *gorm.DB, userID int64, orderNo string) (pojo.CryptoRechargeOrderStatusBack, error) {
	orderNo = strings.TrimSpace(orderNo)
	var result pojo.CryptoRechargeOrderStatusBack
	if orderNo == "" {
		return result, errors.New("order_no_required")
	}

	var rechargeOrder pojo.RechargeOrder
	if err := db.Where("order_no = ? AND user_id = ?", orderNo, userID).First(&rechargeOrder).Error; err != nil {
		return result, err
	}
	var usdtOrder pojo.UsdtRechargeOrder
	if err := db.Where("order_no = ? AND user_id = ?", orderNo, userID).First(&usdtOrder).Error; err != nil {
		return result, err
	}
	return buildCryptoRechargeOrderStatusBack(rechargeOrder, usdtOrder), nil
}

func CancelCryptoRechargeOrder(db *gorm.DB, userID int64, orderNo string) (pojo.CryptoRechargeOrderStatusBack, error) {
	orderNo = strings.TrimSpace(orderNo)
	var result pojo.CryptoRechargeOrderStatusBack
	if orderNo == "" {
		return result, errors.New("order_no_required")
	}
	err := db.Transaction(func(tx *gorm.DB) error {
		var rechargeOrder pojo.RechargeOrder
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("order_no = ? AND user_id = ?", orderNo, userID).
			First(&rechargeOrder).Error; err != nil {
			return err
		}
		var usdtOrder pojo.UsdtRechargeOrder
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("order_no = ? AND user_id = ?", orderNo, userID).
			First(&usdtOrder).Error; err != nil {
			return err
		}
		if usdtOrder.Status == pojo.UsdtRechargeStatusPending && rechargeOrder.Status == 0 {
			now := time.Now()
			if err := tx.Model(&pojo.UsdtRechargeOrder{}).
				Where("id = ? AND status = ?", usdtOrder.ID, pojo.UsdtRechargeStatusPending).
				Update("status", pojo.UsdtRechargeStatusCanceled).Error; err != nil {
				return err
			}
			if err := tx.Model(&pojo.RechargeOrder{}).
				Where("id = ? AND status = ?", rechargeOrder.ID, 0).
				Updates(map[string]any{
					"status":          5,
					"provider_status": "CANCELED",
					"notify_count":    gorm.Expr("notify_count + 1"),
					"notify_last_at":  now,
				}).Error; err != nil {
				return err
			}
			usdtOrder.Status = pojo.UsdtRechargeStatusCanceled
			rechargeOrder.Status = 5
			rechargeOrder.ProviderStatus = nullableUsdtString("CANCELED")
			rechargeOrder.NotifyLastAt = &now
		}
		result = buildCryptoRechargeOrderStatusBack(rechargeOrder, usdtOrder)
		return nil
	})
	return result, err
}

func buildCryptoRechargeOrderStatusBack(rechargeOrder pojo.RechargeOrder, usdtOrder pojo.UsdtRechargeOrder) pojo.CryptoRechargeOrderStatusBack {
	status := usdtOrder.Status
	if rechargeOrder.Status == 1 {
		status = pojo.UsdtRechargeStatusPaid
	}
	txID := ""
	if usdtOrder.TxID != nil {
		txID = strings.TrimSpace(*usdtOrder.TxID)
	}
	remainingSeconds := int64(0)
	if status == pojo.UsdtRechargeStatusPending && !usdtOrder.ExpireTime.IsZero() {
		remainingSeconds = int64(time.Until(usdtOrder.ExpireTime).Seconds())
		if remainingSeconds < 0 {
			remainingSeconds = 0
		}
	}
	return pojo.CryptoRechargeOrderStatusBack{
		OrderNo:          rechargeOrder.OrderNo,
		Status:           status,
		RechargeStatus:   rechargeOrder.Status,
		TxID:             txID,
		PayTime:          rechargeOrder.PayTime,
		ExpireTime:       usdtOrder.ExpireTime.Format(time.RFC3339),
		RemainingSeconds: remainingSeconds,
		CryptoPayment:    buildCryptoPaymentBack(usdtOrder),
	}
}

func buildCryptoPaymentBack(order pojo.UsdtRechargeOrder) *pojo.CryptoPaymentBack {
	if order.OrderNo == "" {
		return nil
	}
	return &pojo.CryptoPaymentBack{
		Network:         order.Network,
		Token:           order.Token,
		ContractAddress: order.ContractAddress,
		ReceiveAddress:  order.ReceiveAddress,
		ExpectedAmount:  order.ExpectedAmount,
		ExpireTime:      order.ExpireTime.Format(time.RFC3339),
		QRContent:       order.ReceiveAddress,
		PlatformRate:    order.PlatformRate,
	}
}

func ExpireUsdtRechargeOrderByOrderNo(db *gorm.DB, orderNo string) error {
	orderNo = strings.TrimSpace(orderNo)
	if orderNo == "" {
		return nil
	}
	now := time.Now()
	return db.Transaction(func(tx *gorm.DB) error {
		var order pojo.UsdtRechargeOrder
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("order_no = ? AND status = ?", orderNo, pojo.UsdtRechargeStatusPending).
			First(&order).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil
			}
			return err
		}
		if !order.ExpireTime.IsZero() && now.Before(order.ExpireTime) {
			return nil
		}
		return expireLockedUsdtRechargeOrder(tx, order, now)
	})
}

func ExpirePendingUsdtRechargeOrders(db *gorm.DB) error {
	now := time.Now()
	var orders []pojo.UsdtRechargeOrder
	if err := db.Where("status = ? AND expire_time < ?", pojo.UsdtRechargeStatusPending, now).
		Limit(200).
		Find(&orders).Error; err != nil {
		return err
	}
	for _, order := range orders {
		if err := db.Transaction(func(tx *gorm.DB) error {
			return expireLockedUsdtRechargeOrder(tx, order, now)
		}); err != nil {
			return err
		}
	}
	return nil
}

func expireLockedUsdtRechargeOrder(tx *gorm.DB, order pojo.UsdtRechargeOrder, now time.Time) error {
	update := tx.Model(&pojo.UsdtRechargeOrder{}).
		Where("id = ? AND status = ?", order.ID, pojo.UsdtRechargeStatusPending).
		Update("status", pojo.UsdtRechargeStatusExpired)
	if update.Error != nil {
		return update.Error
	}
	if update.RowsAffected == 0 {
		return nil
	}
	return tx.Model(&pojo.RechargeOrder{}).
		Where("order_no = ? AND status = ?", order.OrderNo, 0).
		Updates(map[string]any{
			"status":          5,
			"provider_status": "CLOSED",
			"notify_count":    gorm.Expr("notify_count + 1"),
			"notify_last_at":  now,
		}).Error
}

func GetUsdtRechargeScanMinTimestamp(db *gorm.DB) (int64, int64, error) {
	now := time.Now()
	var first pojo.UsdtRechargeOrder
	err := db.Where("status = ? AND expire_time > ?", pojo.UsdtRechargeStatusPending, now).
		Order("created_at asc, id asc").
		First(&first).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, 0, nil
	}
	if err != nil {
		return 0, 0, err
	}
	var count int64
	if err = db.Model(&pojo.UsdtRechargeOrder{}).
		Where("status = ? AND expire_time > ?", pojo.UsdtRechargeStatusPending, now).
		Count(&count).Error; err != nil {
		return 0, 0, err
	}
	minTime := first.CreatedAt.Add(-15 * time.Minute)
	if minTime.IsZero() {
		minTime = now.Add(-time.Duration(usdtTrc20DefaultExpireMinutes) * time.Minute)
	}
	return minTime.UnixMilli(), count, nil
}

func ProcessUsdtRechargeTx(db *gorm.DB, tablePrefix string, txRecord pojo.UsdtRechargeTx) error {
	txRecord.TxID = strings.TrimSpace(txRecord.TxID)
	txRecord.ToAddress = strings.TrimSpace(txRecord.ToAddress)
	txRecord.FromAddress = strings.TrimSpace(txRecord.FromAddress)
	txRecord.ContractAddress = strings.TrimSpace(txRecord.ContractAddress)
	if txRecord.TxID == "" || txRecord.ToAddress == "" || txRecord.AmountMicro <= 0 {
		return nil
	}
	if txRecord.Network == "" {
		txRecord.Network = pojo.CryptoRechargeNetworkTRC20
	}
	if txRecord.Token == "" {
		txRecord.Token = pojo.CryptoRechargeTokenUSDT
	}
	if txRecord.Amount == "" {
		txRecord.Amount = FormatUsdtMicroAmount(txRecord.AmountMicro)
	}
	if err := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&txRecord).Error; err != nil {
		return err
	}

	var existingTx pojo.UsdtRechargeTx
	if err := db.Where("tx_id = ?", txRecord.TxID).First(&existingTx).Error; err != nil {
		return err
	}
	if existingTx.MatchedOrderNo != nil && strings.TrimSpace(*existingTx.MatchedOrderNo) != "" {
		return nil
	}

	now := time.Now()
	var order pojo.UsdtRechargeOrder
	err := db.Where(
		"receive_address = ? AND expected_amount_micro = ? AND status = ? AND expire_time > ?",
		txRecord.ToAddress,
		txRecord.AmountMicro,
		pojo.UsdtRechargeStatusPending,
		now,
	).Order("created_at asc, id asc").First(&order).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	if err != nil {
		return err
	}

	if err = ProcessRechargeOrderSuccess(db, order.OrderNo, txRecord.TxID, order.PlatformAmount, tablePrefix); err != nil {
		return err
	}

	paidAmountMicro := txRecord.AmountMicro
	return db.Transaction(func(tx *gorm.DB) error {
		updates := map[string]any{
			"status":            pojo.UsdtRechargeStatusPaid,
			"paid_amount_micro": paidAmountMicro,
			"paid_at":           now,
			"tx_id":             txRecord.TxID,
		}
		if err := tx.Model(&pojo.UsdtRechargeOrder{}).
			Where("id = ? AND status = ?", order.ID, pojo.UsdtRechargeStatusPending).
			Updates(updates).Error; err != nil {
			return err
		}
		matchedOrderNo := order.OrderNo
		return tx.Model(&pojo.UsdtRechargeTx{}).
			Where("tx_id = ?", txRecord.TxID).
			Updates(map[string]any{
				"matched_order_no": matchedOrderNo,
				"matched_at":       now,
			}).Error
	})
}
