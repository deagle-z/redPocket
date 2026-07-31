package repository

import (
	"BaseGoUni/core/pojo"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/plugin/dbresolver"
)

type CryptoRechargeScanTarget struct {
	ReceiveAddress  string `gorm:"column:receive_address"`
	ContractAddress string `gorm:"column:contract_address"`
}

func ListCryptoRechargeScanTargets(
	db *gorm.DB,
	network string,
	token string,
	currentReceiveAddress string,
	currentContractAddress string,
) ([]CryptoRechargeScanTarget, error) {
	if db == nil {
		return nil, errors.New("db_not_available")
	}
	network = strings.ToUpper(strings.TrimSpace(network))
	token = strings.ToUpper(strings.TrimSpace(token))
	if network == "" || token == "" {
		return nil, errors.New("crypto_scan_target_fields_required")
	}
	rows := make([]CryptoRechargeScanTarget, 0)
	if err := db.Clauses(dbresolver.Write).
		Model(&pojo.UsdtRechargeOrder{}).
		Select("receive_address, contract_address").
		Where("network = ? AND token = ? AND receive_address <> ''", network, token).
		Group("receive_address, contract_address").
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	if strings.TrimSpace(currentReceiveAddress) != "" {
		rows = append(rows, CryptoRechargeScanTarget{
			ReceiveAddress:  currentReceiveAddress,
			ContractAddress: currentContractAddress,
		})
	}
	result := make([]CryptoRechargeScanTarget, 0, len(rows))
	seen := make(map[string]struct{}, len(rows))
	for _, row := range rows {
		row.ReceiveAddress = normalizeCryptoRechargeAddress(network, row.ReceiveAddress)
		row.ContractAddress = strings.TrimSpace(row.ContractAddress)
		if row.ReceiveAddress == "" {
			continue
		}
		key := strings.Join([]string{network, token, row.ReceiveAddress, row.ContractAddress}, "|")
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, row)
	}
	return result, nil
}

func GetCryptoRechargeScanCursor(db *gorm.DB, network string, token string, receiveAddress string, contractAddresses ...string) (pojo.CryptoRechargeScanCursor, error) {
	var cursor pojo.CryptoRechargeScanCursor
	if db == nil {
		return cursor, errors.New("db_not_available")
	}
	network = strings.ToUpper(strings.TrimSpace(network))
	token = strings.ToUpper(strings.TrimSpace(token))
	receiveAddress = normalizeCryptoRechargeAddress(network, receiveAddress)
	contractAddress := ""
	if len(contractAddresses) > 0 {
		contractAddress = strings.TrimSpace(contractAddresses[0])
	}
	if network == "" || token == "" || receiveAddress == "" {
		return cursor, errors.New("crypto_cursor_fields_required")
	}
	cursor.CursorKey = buildCryptoScanCursorKey(network, token, receiveAddress, contractAddress)
	cursor.Network = network
	cursor.Token = token
	cursor.ReceiveAddress = receiveAddress
	cursor.ContractAddress = contractAddress
	err := db.Clauses(dbresolver.Write).Where("cursor_key = ?", cursor.CursorKey).First(&cursor).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return cursor, nil
	}
	return cursor, err
}

func SaveCryptoRechargeScanCursor(db *gorm.DB, cursor pojo.CryptoRechargeScanCursor) error {
	if db == nil {
		return errors.New("db_not_available")
	}
	cursor.Network = strings.ToUpper(strings.TrimSpace(cursor.Network))
	cursor.Token = strings.ToUpper(strings.TrimSpace(cursor.Token))
	cursor.ReceiveAddress = normalizeCryptoRechargeAddress(cursor.Network, cursor.ReceiveAddress)
	cursor.ContractAddress = strings.TrimSpace(cursor.ContractAddress)
	if cursor.CursorKey == "" {
		cursor.CursorKey = buildCryptoScanCursorKey(cursor.Network, cursor.Token, cursor.ReceiveAddress, cursor.ContractAddress)
	}
	if cursor.Network == "" || cursor.Token == "" || cursor.ReceiveAddress == "" || cursor.CursorKey == "" {
		return errors.New("crypto_cursor_fields_required")
	}
	now := time.Now()
	return db.Clauses(dbresolver.Write, clause.OnConflict{
		Columns: []clause.Column{{Name: "cursor_key"}},
		DoUpdates: clause.Assignments(map[string]any{
			"network":                cursor.Network,
			"token":                  cursor.Token,
			"receive_address":        cursor.ReceiveAddress,
			"contract_address":       cursor.ContractAddress,
			"last_block_height":      cursor.LastBlockHeight,
			"last_block_hash":        cursor.LastBlockHash,
			"last_block_timestamp":   cursor.LastBlockTimestamp,
			"pagination_cursor":      strings.TrimSpace(cursor.PaginationCursor),
			"target_block_height":    cursor.TargetBlockHeight,
			"target_block_timestamp": cursor.TargetBlockTimestamp,
			"updated_at":             now,
		}),
	}).Create(&cursor).Error
}

func normalizeCryptoRechargeAddress(network string, address string) string {
	address = strings.TrimSpace(address)
	if strings.EqualFold(strings.TrimSpace(network), pojo.CryptoRechargeNetworkEthereum) {
		return strings.ToLower(address)
	}
	return address
}

func buildCryptoScanCursorKey(network string, token string, receiveAddress string, contractAddress string) string {
	parts := []string{network, token, receiveAddress}
	if strings.TrimSpace(contractAddress) != "" {
		parts = append(parts, strings.TrimSpace(contractAddress))
	}
	return strings.Join(parts, "|")
}

func MarkCryptoRechargeReorgWindowForManualReview(db *gorm.DB, network string, afterBlockHeight int64) error {
	if db == nil {
		return errors.New("db_not_available")
	}
	network = strings.ToUpper(strings.TrimSpace(network))
	if network == "" || afterBlockHeight < 0 {
		return errors.New("crypto_reorg_fields_required")
	}
	return db.Clauses(dbresolver.Write).Model(&pojo.UsdtRechargeTx{}).
		Where("network = ? AND block_height > ? AND review_status IN ?", network, afterBlockHeight, []int8{
			pojo.CryptoRechargeTxReviewNew,
			pojo.CryptoRechargeTxReviewAutoMatched,
		}).
		Updates(map[string]any{
			"review_status":  pojo.CryptoRechargeTxReviewManualRequired,
			"exception_code": pojo.CryptoRechargeExceptionChainReorg,
		}).Error
}
