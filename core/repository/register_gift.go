package repository

import (
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/utils"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	RegisterGiftAmountConfigKey = "register_gift_amount"
	defaultRegisterGiftAmount   = 58.0
)

// GetRegisterGiftAmount returns the configured registration gift amount.
// Missing config is initialized to the default 58. A configured value of 0 disables the gift.
func GetRegisterGiftAmount(db *gorm.DB) float64 {
	if db == nil {
		return defaultRegisterGiftAmount
	}

	defaultValue := strconv.FormatFloat(defaultRegisterGiftAmount, 'f', -1, 64)
	var cfg pojo.SysConfig
	err := db.Where("config_key = ?", RegisterGiftAmountConfigKey).First(&cfg).Error
	if errors.Is(err, gorm.ErrRecordNotFound) || (err == nil && cfg.ID == 0) {
		_ = db.Clauses(clause.OnConflict{DoNothing: true}).Create(&pojo.SysConfig{
			ConfigKey:   RegisterGiftAmountConfigKey,
			ConfigValue: defaultValue,
			ConfigDesc:  "注册自动赠送金额，配置为0则关闭",
		}).Error
		return defaultRegisterGiftAmount
	}
	if err != nil {
		return defaultRegisterGiftAmount
	}

	amount, parseErr := strconv.ParseFloat(strings.TrimSpace(cfg.ConfigValue), 64)
	if parseErr != nil || amount < 0 {
		return defaultRegisterGiftAmount
	}
	return utils.Truncate2(amount)
}

// ApplyRegisterGift credits the registration gift once per user and creates the related ledgers.
func ApplyRegisterGift(tx *gorm.DB, user *pojo.TgUser) error {
	if tx == nil || user == nil || user.ID <= 0 {
		return nil
	}

	giftAmount := GetRegisterGiftAmount(tx)
	if giftAmount <= 0 {
		return nil
	}
	awardUni := buildRegisterGiftAwardUni(user.ID)
	cashHistoryTable := registerGiftCashHistoryTableName(user.ID)

	var existing pojo.CashHistory
	err := tx.Table(cashHistoryTable).
		Where("user_id = ? AND award_uni = ?", user.ID, awardUni).
		First(&existing).Error
	if err == nil && existing.ID > 0 {
		return nil
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	var lockedUser pojo.TgUser
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", user.ID).First(&lockedUser).Error; err != nil {
		return err
	}
	beforeBalance := utils.Truncate2(lockedUser.Balance)
	afterBalance := utils.Truncate2(beforeBalance + giftAmount)

	if err := tx.Model(&pojo.TgUser{}).Where("id = ?", user.ID).Updates(map[string]any{
		"balance": gorm.Expr("balance + ?", giftAmount),
	}).Error; err != nil {
		return err
	}

	desc := fmt.Sprintf("注册自动赠送%.2f", giftAmount)
	history := pojo.CashHistory{
		UserId:          user.ID,
		AwardUni:        awardUni,
		Amount:          giftAmount,
		StartAmount:     beforeBalance,
		EndAmount:       afterBalance,
		CashMark:        "注册赠送",
		CashDesc:        desc,
		Type:            pojo.CashHistoryTypeRegisterGift,
		IsGift:          1,
		FromUserId:      0,
		SourceChannelID: lockedUser.SourceChannelID,
	}
	if err := tx.Table(cashHistoryTable).Create(&history).Error; err != nil {
		return err
	}

	if err := EnsureWithdrawFlowBatchForGift(
		tx,
		lockedUser,
		pojo.WithdrawFlowBatchSourceRegisterGift,
		history.ID,
		awardUni,
		pojo.WithdrawFlowBatchSourceRegisterGift,
		giftAmount,
	); err != nil {
		return err
	}

	if err := CreatePlatformProfitLedgerIfAbsent(tx, pojo.PlatformProfitLedger{
		TenantId:        lockedUser.TenantId,
		UserId:          lockedUser.ID,
		SourceChannelID: lockedUser.SourceChannelID,
		SourceType:      pojo.PlatformProfitSourceRegisterGift,
		SourceId:        awardUni,
		IncomeAmount:    0,
		ExpenseAmount:   giftAmount,
		Remark:          desc,
	}); err != nil {
		return err
	}

	user.Balance = afterBalance
	return nil
}

func buildRegisterGiftAwardUni(userID int64) string {
	return fmt.Sprintf("register_gift_%d", userID)
}

func registerGiftCashHistoryTableName(userID int64) string {
	shardIndex := int(userID % int64(pojo.CashHistoryShards))
	return cashHistoryShardTableName(shardIndex)
}
