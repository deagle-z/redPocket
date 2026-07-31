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
	"gorm.io/plugin/dbresolver"
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
	cryptoRechargePlatformCurrency  = "USD"
	cryptoRechargeMaxPlatformAmount = 90071992547409.91
)

var (
	errCryptoRechargeTxAlreadyProcessed = errors.New("crypto_tx_already_processed")
	errCryptoRechargeTxManualReview     = errors.New("crypto_tx_manual_review")
	errCryptoRechargeOrderUnavailable   = errors.New("crypto_order_unavailable")
	errCryptoRechargeOrderMismatch      = errors.New("crypto_order_mismatch")
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
	if !isValidTronMainnetAddress(cfg.ReceiveAddress) {
		return cfg, errors.New("usdt_trc20_receive_address_invalid")
	}
	if !isValidTronMainnetAddress(cfg.ContractAddress) {
		return cfg, errors.New("usdt_trc20_contract_address_invalid")
	}
	if err = validateCryptoReceiveAddressUnique(tablePrefix, pojo.CryptoRechargeNetworkTRC20, cfg.ReceiveAddress); err != nil {
		return cfg, err
	}
	if _, err = parsePositiveRat(cfg.PlatformRate); err != nil {
		return cfg, errors.New("usdt_trc20_platform_rate_invalid")
	}
	return cfg, nil
}

func ValidateUsdtTrc20RuntimeConfig(tablePrefix string) (UsdtTrc20RuntimeConfig, error) {
	return validateUsdtTrc20RuntimeConfig(tablePrefix)
}

func GetCryptoRechargeOptions(db *gorm.DB, tablePrefix string, amount float64) (pojo.CryptoRechargeOptionsBack, error) {
	amount = normalizeRechargeOrderAmount(amount)
	result := pojo.CryptoRechargeOptionsBack{PlatformAmount: amount}
	if amount <= 0 {
		return result, errors.New("recharge_amount_positive")
	}
	if amount > cryptoRechargeMaxPlatformAmount {
		return result, errors.New("recharge_amount_too_large")
	}
	configs, err := listEnabledCryptoAssetConfigs(tablePrefix)
	if err != nil {
		return result, err
	}
	if len(configs) == 0 {
		return result, errors.New("crypto_recharge_unavailable")
	}
	result.Options = make([]pojo.CryptoRechargeOptionBack, 0, len(configs))
	for _, cfg := range configs {
		baseAmount, calcErr := calculateCryptoBaseAmountAtomic(amount, cfg.PlatformRate, cfg.Decimals)
		if calcErr != nil {
			return result, calcErr
		}
		result.Options = append(result.Options, pojo.CryptoRechargeOptionBack{
			Network:         cfg.Network,
			Token:           cfg.Token,
			ContractAddress: cfg.ContractAddress,
			ReceiveAddress:  cfg.ReceiveAddress,
			EstimatedAmount: formatCryptoAtomicAmount(baseAmount, cfg.Decimals),
			PlatformRate:    cfg.PlatformRate,
			PlatformAmount:  amount,
		})
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
	req.Amount = normalizeRechargeOrderAmount(req.Amount)
	var result pojo.RechargeOrderAppBack
	if req.Amount <= 0 {
		return result, errors.New("recharge_amount_positive")
	}
	if req.Amount > cryptoRechargeMaxPlatformAmount {
		return result, errors.New("recharge_amount_too_large")
	}
	if minAmount > 0 && req.Amount < minAmount {
		return result, errors.New(utils.I18nMessage("recharge_v2_min_amount", map[string]interface{}{
			"min": formatRechargeMinAmount(minAmount),
		}))
	}

	cfg, err := getCryptoAssetConfig(tablePrefix, req.Network, req.Token)
	if err != nil {
		return result, err
	}

	lockKey := fmt.Sprintf("crypto_recharge_create:%s:%s:%s", tablePrefix, cfg.Network, cfg.ReceiveAddress)
	createLock, locked, lockErr := utils.AcquireOwnedLock(lockKey, 30*time.Second)
	if lockErr != nil {
		return result, lockErr
	}
	if !locked {
		return result, errors.New("crypto_recharge_create_busy")
	}
	defer createLock.Release()

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

		orderNo := buildRechargeOrderNo()
		baseAmountAtomic, calcErr := calculateCryptoBaseAmountAtomic(req.Amount, cfg.PlatformRate, cfg.Decimals)
		if calcErr != nil {
			return calcErr
		}
		expectedAmountAtomic, amountErr := allocateCryptoExpectedAmountAtomic(tx, cfg, baseAmountAtomic, orderNo)
		if amountErr != nil {
			return amountErr
		}
		expectedAmount := formatCryptoAtomicAmount(expectedAmountAtomic, cfg.Decimals)
		qrContent, qrErr := buildCryptoQRContent(cfg, expectedAmountAtomic, expectedAmount)
		if qrErr != nil {
			return qrErr
		}
		baseAmountMicro := int64(0)
		expectedAmountMicro := int64(0)
		if cfg.Network == pojo.CryptoRechargeNetworkTRC20 && baseAmountAtomic.IsInt64() && expectedAmountAtomic.IsInt64() {
			baseAmountMicro = baseAmountAtomic.Int64()
			expectedAmountMicro = expectedAmountAtomic.Int64()
		}

		expireTime := time.Now().Add(time.Duration(cfg.OrderExpireMinutes) * time.Minute)
		merchantOrderNo := nullableUsdtString(req.MerchantOrderNo)
		payMethod := cfg.PayMethod
		provider := cfg.Provider
		extra := map[string]any{
			"cryptoPayment": map[string]any{
				"network":              cfg.Network,
				"token":                cfg.Token,
				"contractAddress":      cfg.ContractAddress,
				"receiveAddress":       cfg.ReceiveAddress,
				"platformRate":         cfg.PlatformRate,
				"amountDecimals":       cfg.Decimals,
				"baseAmountAtomic":     baseAmountAtomic.String(),
				"expectedAmountAtomic": expectedAmountAtomic.String(),
				"baseAmountMicro":      baseAmountMicro,
				"expectedAmountMicro":  expectedAmountMicro,
				"expectedAmount":       expectedAmount,
			},
		}
		extraBytes, marshalErr := json.Marshal(extra)
		if marshalErr != nil {
			return marshalErr
		}
		extraStr := string(extraBytes)

		order := pojo.RechargeOrder{
			TenantId:        tgUser.TenantId,
			UserId:          userID,
			SourceChannelID: tgUser.SourceChannelID,
			OrderNo:         orderNo,
			MerchantOrderNo: merchantOrderNo,
			Channel:         cfg.Channel,
			PayMethod:       &payMethod,
			Currency:        cryptoRechargePlatformCurrency,
			Amount:          req.Amount,
			Fee:             0,
			NetAmount:       req.Amount,
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
			TenantId:             tgUser.TenantId,
			UserId:               userID,
			SourceChannelID:      tgUser.SourceChannelID,
			OrderNo:              orderNo,
			Network:              cfg.Network,
			Token:                cfg.Token,
			ContractAddress:      cfg.ContractAddress,
			ReceiveAddress:       cfg.ReceiveAddress,
			PlatformAmount:       req.Amount,
			PlatformRate:         cfg.PlatformRate,
			AmountDecimals:       cfg.Decimals,
			BaseAmountAtomic:     baseAmountAtomic.String(),
			ExpectedAmountAtomic: expectedAmountAtomic.String(),
			BaseAmountMicro:      baseAmountMicro,
			ExpectedAmountMicro:  expectedAmountMicro,
			ExpectedAmount:       expectedAmount,
			Status:               pojo.UsdtRechargeStatusPending,
			ExpireTime:           expireTime,
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
				Network:         cfg.Network,
				Token:           cfg.Token,
				ContractAddress: cfg.ContractAddress,
				ReceiveAddress:  cfg.ReceiveAddress,
				ExpectedAmount:  expectedAmount,
				ExpireTime:      expireTime.Format(time.RFC3339),
				QRContent:       qrContent,
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
	if len(value) > 32 || !positiveDecimalPattern.MatchString(value) {
		return nil, errors.New("invalid_decimal")
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
	qrContent := order.ReceiveAddress
	if order.Network == pojo.CryptoRechargeNetworkBitcoin || order.Network == pojo.CryptoRechargeNetworkEthereum {
		atomicAmount, err := parseCryptoAtomicAmount(order.ExpectedAmountAtomic)
		if err != nil {
			qrContent = ""
		} else {
			cfg := cryptoAssetRuntimeConfig{
				Network:        order.Network,
				ReceiveAddress: order.ReceiveAddress,
			}
			qrContent, _ = buildCryptoQRContent(cfg, atomicAmount, order.ExpectedAmount)
		}
	}
	return &pojo.CryptoPaymentBack{
		Network:         order.Network,
		Token:           order.Token,
		ContractAddress: order.ContractAddress,
		ReceiveAddress:  order.ReceiveAddress,
		ExpectedAmount:  order.ExpectedAmount,
		ExpireTime:      order.ExpireTime.Format(time.RFC3339),
		QRContent:       qrContent,
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
		return expireCryptoRechargeOrder(tx, orderNo, now)
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
			return expireCryptoRechargeOrder(tx, order.OrderNo, now)
		}); err != nil {
			return err
		}
	}
	return nil
}

func expireCryptoRechargeOrder(tx *gorm.DB, orderNo string, now time.Time) error {
	var rechargeOrder pojo.RechargeOrder
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("order_no = ?", orderNo).
		First(&rechargeOrder).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}
	if rechargeOrder.Status != 0 {
		return nil
	}
	var cryptoOrder pojo.UsdtRechargeOrder
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("order_no = ? AND status = ?", orderNo, pojo.UsdtRechargeStatusPending).
		First(&cryptoOrder).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}
	if !cryptoOrder.ExpireTime.IsZero() && now.Before(cryptoOrder.ExpireTime) {
		return nil
	}
	update := tx.Model(&pojo.UsdtRechargeOrder{}).
		Where("id = ? AND status = ?", cryptoOrder.ID, pojo.UsdtRechargeStatusPending).
		Update("status", pojo.UsdtRechargeStatusExpired)
	if update.Error != nil {
		return update.Error
	}
	if update.RowsAffected == 0 {
		return nil
	}
	return tx.Model(&pojo.RechargeOrder{}).
		Where("id = ? AND status = ?", rechargeOrder.ID, 0).
		Updates(map[string]any{
			"status":          5,
			"provider_status": "CLOSED",
			"notify_count":    gorm.Expr("notify_count + 1"),
			"notify_last_at":  now,
		}).Error
}

func GetUsdtRechargeScanMinTimestamp(db *gorm.DB) (int64, int64, error) {
	return GetCryptoRechargeScanMinTimestamp(db, pojo.CryptoRechargeNetworkTRC20, pojo.CryptoRechargeTokenUSDT)
}

func GetCryptoRechargeScanMinTimestamp(db *gorm.DB, network string, token string) (int64, int64, error) {
	return GetCryptoRechargeScanMinTimestampForTarget(db, network, token, "", "")
}

func GetCryptoRechargeScanMinTimestampForTarget(
	db *gorm.DB,
	network string,
	token string,
	receiveAddress string,
	contractAddress string,
) (int64, int64, error) {
	now := time.Now()
	var first pojo.UsdtRechargeOrder
	query := db.Clauses(dbresolver.Write).Where("network = ? AND token = ? AND status = ? AND expire_time > ?", network, token, pojo.UsdtRechargeStatusPending, now)
	if receiveAddress = strings.TrimSpace(receiveAddress); receiveAddress != "" {
		query = query.Where("receive_address = ?", receiveAddress)
	}
	if contractAddress = strings.TrimSpace(contractAddress); contractAddress != "" {
		query = query.Where("contract_address = ?", contractAddress)
	}
	err := query.
		Order("created_at asc, id asc").
		First(&first).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, 0, nil
	}
	if err != nil {
		return 0, 0, err
	}
	var count int64
	countQuery := db.Clauses(dbresolver.Write).Model(&pojo.UsdtRechargeOrder{}).
		Where("network = ? AND token = ? AND status = ? AND expire_time > ?", network, token, pojo.UsdtRechargeStatusPending, now)
	if receiveAddress != "" {
		countQuery = countQuery.Where("receive_address = ?", receiveAddress)
	}
	if contractAddress != "" {
		countQuery = countQuery.Where("contract_address = ?", contractAddress)
	}
	if err = countQuery.Count(&count).Error; err != nil {
		return 0, 0, err
	}
	minTime := first.CreatedAt.Add(-15 * time.Minute)
	if minTime.IsZero() {
		minTime = now.Add(-time.Duration(usdtTrc20DefaultExpireMinutes) * time.Minute)
	}
	return minTime.UnixMilli(), count, nil
}

func ProcessUsdtRechargeTx(db *gorm.DB, tablePrefix string, txRecord pojo.UsdtRechargeTx) error {
	if txRecord.Network == "" {
		txRecord.Network = pojo.CryptoRechargeNetworkTRC20
	}
	if txRecord.Token == "" {
		txRecord.Token = pojo.CryptoRechargeTokenUSDT
	}
	return ProcessCryptoRechargeTx(db, tablePrefix, txRecord)
}

func ProcessCryptoRechargeTx(db *gorm.DB, tablePrefix string, txRecord pojo.UsdtRechargeTx) error {
	txRecord.TxID = strings.ToLower(strings.TrimSpace(txRecord.TxID))
	txRecord.ToAddress = strings.TrimSpace(txRecord.ToAddress)
	txRecord.FromAddress = strings.TrimSpace(txRecord.FromAddress)
	txRecord.ContractAddress = strings.TrimSpace(txRecord.ContractAddress)
	txRecord.Network = strings.ToUpper(strings.TrimSpace(txRecord.Network))
	txRecord.Token = strings.ToUpper(strings.TrimSpace(txRecord.Token))
	if txRecord.Network == pojo.CryptoRechargeNetworkEthereum {
		txRecord.ToAddress = strings.ToLower(txRecord.ToAddress)
		txRecord.FromAddress = strings.ToLower(txRecord.FromAddress)
	}
	if txRecord.TxID == "" || txRecord.ToAddress == "" || txRecord.Network == "" || txRecord.Token == "" {
		return errors.New("crypto_tx_fields_required")
	}
	if txRecord.AmountAtomic == "" && txRecord.AmountMicro > 0 {
		txRecord.AmountAtomic = strconv.FormatInt(txRecord.AmountMicro, 10)
	}
	atomicAmount, err := parseCryptoAtomicAmount(txRecord.AmountAtomic)
	if err != nil {
		return err
	}
	if txRecord.BlockTimestamp <= 0 {
		return errors.New("crypto_tx_block_timestamp_required")
	}
	if txRecord.AmountDecimals <= 0 {
		switch txRecord.Token {
		case pojo.CryptoRechargeTokenBTC:
			txRecord.AmountDecimals = pojo.CryptoRechargeDecimalsBTC
		case pojo.CryptoRechargeTokenETH:
			txRecord.AmountDecimals = pojo.CryptoRechargeDecimalsETH
		default:
			txRecord.AmountDecimals = pojo.CryptoRechargeDecimalsUSDT
		}
	}
	if txRecord.Amount == "" {
		txRecord.Amount = formatCryptoAtomicAmount(atomicAmount, txRecord.AmountDecimals)
	}
	txRecord.ReviewStatus = pojo.CryptoRechargeTxReviewNew
	if err := db.Clauses(dbresolver.Write, clause.OnConflict{DoNothing: true}).Create(&txRecord).Error; err != nil {
		return err
	}

	var existingTx pojo.UsdtRechargeTx
	if err := db.Clauses(dbresolver.Write).Where("tx_id = ?", txRecord.TxID).First(&existingTx).Error; err != nil {
		return err
	}
	if existingTx.MatchedOrderNo != nil && strings.TrimSpace(*existingTx.MatchedOrderNo) != "" {
		return nil
	}
	if existingTx.ReviewStatus != pojo.CryptoRechargeTxReviewNew {
		return nil
	}

	now := time.Now()
	txTime := time.UnixMilli(txRecord.BlockTimestamp)
	var order pojo.UsdtRechargeOrder
	query := db.Clauses(dbresolver.Write).Where(
		"network = ? AND token = ? AND receive_address = ? AND expected_amount_atomic = ? AND status IN ? AND expire_time >= ? AND created_at <= ?",
		txRecord.Network,
		txRecord.Token,
		txRecord.ToAddress,
		atomicAmount.String(),
		[]int8{pojo.UsdtRechargeStatusPending, pojo.UsdtRechargeStatusExpired},
		txTime,
		txTime,
	)
	err = query.Order("created_at asc, id asc").First(&order).Error
	if errors.Is(err, gorm.ErrRecordNotFound) && txRecord.Network == pojo.CryptoRechargeNetworkTRC20 && txRecord.AmountMicro > 0 {
		err = db.Clauses(dbresolver.Write).Where(
			"network = ? AND token = ? AND receive_address = ? AND (expected_amount_atomic IS NULL OR expected_amount_atomic = '') AND expected_amount_micro = ? AND status IN ? AND expire_time >= ? AND created_at <= ?",
			txRecord.Network,
			txRecord.Token,
			txRecord.ToAddress,
			txRecord.AmountMicro,
			[]int8{pojo.UsdtRechargeStatusPending, pojo.UsdtRechargeStatusExpired},
			txTime,
			txTime,
		).Order("created_at asc, id asc").First(&order).Error
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		exceptionCode, classifyErr := classifyCryptoRechargeException(db, txRecord, atomicAmount, txTime)
		if classifyErr != nil {
			return classifyErr
		}
		return markCryptoRechargeTxForManualReview(db, existingTx.ID, exceptionCode)
	}
	if err != nil {
		return err
	}

	paidAmountAtomic := atomicAmount.String()
	var lockedCryptoOrder pojo.UsdtRechargeOrder
	err = processRechargeOrderSuccessWithHooks(
		db,
		order.OrderNo,
		txRecord.TxID,
		order.PlatformAmount,
		tablePrefix,
		func(tx *gorm.DB, rechargeOrder *pojo.RechargeOrder) error {
			if rechargeOrder.Status == 1 {
				return errCryptoRechargeOrderMismatch
			}
			var lockedTx pojo.UsdtRechargeTx
			if lockErr := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", existingTx.ID).First(&lockedTx).Error; lockErr != nil {
				return lockErr
			}
			if lockedTx.MatchedOrderNo != nil && strings.TrimSpace(*lockedTx.MatchedOrderNo) != "" {
				return errCryptoRechargeTxAlreadyProcessed
			}
			if lockedTx.ReviewStatus != pojo.CryptoRechargeTxReviewNew {
				return errCryptoRechargeTxManualReview
			}
			if lockErr := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", order.ID).First(&lockedCryptoOrder).Error; lockErr != nil {
				return lockErr
			}
			if lockedCryptoOrder.Status != pojo.UsdtRechargeStatusPending && lockedCryptoOrder.Status != pojo.UsdtRechargeStatusExpired {
				return errCryptoRechargeOrderUnavailable
			}
			if txTime.After(lockedCryptoOrder.ExpireTime) {
				return errCryptoRechargeOrderUnavailable
			}
			if lockedCryptoOrder.Status == pojo.UsdtRechargeStatusExpired {
				if rechargeOrder.Status != 5 {
					return errCryptoRechargeOrderUnavailable
				}
				reopened := tx.Model(&pojo.RechargeOrder{}).
					Where("id = ? AND status = ?", rechargeOrder.ID, 5).
					Updates(map[string]any{"status": 0, "provider_status": "CHAIN_CONFIRMED"})
				if reopened.Error != nil {
					return reopened.Error
				}
				if reopened.RowsAffected != 1 {
					return errCryptoRechargeOrderUnavailable
				}
				rechargeOrder.Status = 0
			} else if rechargeOrder.Status != 0 {
				return errCryptoRechargeOrderMismatch
			}
			amountMatches := lockedCryptoOrder.ExpectedAmountAtomic == atomicAmount.String()
			if !amountMatches && txRecord.Network == pojo.CryptoRechargeNetworkTRC20 && lockedCryptoOrder.ExpectedAmountAtomic == "" {
				amountMatches = lockedCryptoOrder.ExpectedAmountMicro == txRecord.AmountMicro
			}
			if rechargeOrder.OrderNo != lockedCryptoOrder.OrderNo ||
				lockedCryptoOrder.Network != txRecord.Network ||
				lockedCryptoOrder.Token != txRecord.Token ||
				lockedCryptoOrder.ReceiveAddress != txRecord.ToAddress ||
				!amountMatches ||
				txTime.Before(lockedCryptoOrder.CreatedAt) {
				return errCryptoRechargeOrderMismatch
			}
			return nil
		},
		func(tx *gorm.DB, _ *pojo.RechargeOrder) error {
			updates := map[string]any{
				"status":             pojo.UsdtRechargeStatusPaid,
				"paid_amount_atomic": paidAmountAtomic,
				"paid_at":            now,
				"tx_id":              txRecord.TxID,
			}
			if txRecord.Network == pojo.CryptoRechargeNetworkTRC20 && txRecord.AmountMicro > 0 {
				updates["paid_amount_micro"] = txRecord.AmountMicro
			}
			cryptoUpdate := tx.Model(&pojo.UsdtRechargeOrder{}).
				Where("id = ? AND status IN ?", order.ID, []int8{pojo.UsdtRechargeStatusPending, pojo.UsdtRechargeStatusExpired}).
				Updates(updates)
			if cryptoUpdate.Error != nil {
				return cryptoUpdate.Error
			}
			if cryptoUpdate.RowsAffected != 1 {
				return errCryptoRechargeOrderUnavailable
			}
			matchedOrderNo := order.OrderNo
			txUpdate := tx.Model(&pojo.UsdtRechargeTx{}).
				Where("id = ? AND matched_order_no IS NULL AND review_status = ?", existingTx.ID, pojo.CryptoRechargeTxReviewNew).
				Updates(map[string]any{
					"matched_order_no": matchedOrderNo,
					"matched_at":       now,
					"review_status":    pojo.CryptoRechargeTxReviewAutoMatched,
					"exception_code":   "",
				})
			if txUpdate.Error != nil {
				return txUpdate.Error
			}
			if txUpdate.RowsAffected != 1 {
				return errCryptoRechargeTxAlreadyProcessed
			}
			return nil
		},
	)
	if errors.Is(err, errCryptoRechargeTxAlreadyProcessed) || errors.Is(err, errCryptoRechargeTxManualReview) {
		return nil
	}
	if errors.Is(err, errCryptoRechargeOrderUnavailable) || errors.Is(err, errCryptoRechargeOrderMismatch) {
		exceptionCode := pojo.CryptoRechargeExceptionDuplicate
		if errors.Is(err, errCryptoRechargeOrderMismatch) {
			exceptionCode = pojo.CryptoRechargeExceptionOrderState
		}
		markErr := markCryptoRechargeTxForManualReview(db, existingTx.ID, exceptionCode)
		if markErr != nil {
			return markErr
		}
		return nil
	}
	return err
}

func markCryptoRechargeTxForManualReview(db *gorm.DB, txID int64, exceptionCode string) error {
	if txID <= 0 || strings.TrimSpace(exceptionCode) == "" {
		return errors.New("crypto_exception_fields_required")
	}
	return db.Clauses(dbresolver.Write).Model(&pojo.UsdtRechargeTx{}).
		Where("id = ? AND matched_order_no IS NULL AND review_status = ?", txID, pojo.CryptoRechargeTxReviewNew).
		Updates(map[string]any{
			"review_status":  pojo.CryptoRechargeTxReviewManualRequired,
			"exception_code": exceptionCode,
		}).Error
}

func classifyCryptoRechargeException(
	db *gorm.DB,
	txRecord pojo.UsdtRechargeTx,
	atomicAmount *big.Int,
	txTime time.Time,
) (string, error) {
	var exactOrder pojo.UsdtRechargeOrder
	err := db.Clauses(dbresolver.Write).
		Where("network = ? AND token = ? AND receive_address = ? AND expected_amount_atomic = ?",
			txRecord.Network, txRecord.Token, txRecord.ToAddress, atomicAmount.String()).
		Order("created_at desc, id desc").
		First(&exactOrder).Error
	if err == nil && exactOrder.ID > 0 {
		if txTime.Before(exactOrder.CreatedAt) {
			return pojo.CryptoRechargeExceptionPreOrder, nil
		}
		if txTime.After(exactOrder.ExpireTime) || exactOrder.Status == pojo.UsdtRechargeStatusCanceled {
			return pojo.CryptoRechargeExceptionLatePayment, nil
		}
		return pojo.CryptoRechargeExceptionDuplicate, nil
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return "", err
	}

	var activeOrderCount int64
	if err = db.Clauses(dbresolver.Write).Model(&pojo.UsdtRechargeOrder{}).
		Where("network = ? AND token = ? AND receive_address = ? AND status IN ? AND expire_time >= ? AND created_at <= ?",
			txRecord.Network,
			txRecord.Token,
			txRecord.ToAddress,
			[]int8{pojo.UsdtRechargeStatusPending, pojo.UsdtRechargeStatusExpired},
			txTime,
			txTime,
		).
		Count(&activeOrderCount).Error; err != nil {
		return "", err
	}
	if activeOrderCount > 0 {
		return pojo.CryptoRechargeExceptionAmountMismatch, nil
	}
	return pojo.CryptoRechargeExceptionNoActiveOrder, nil
}
