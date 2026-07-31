package services

import (
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/repository"
	"BaseGoUni/core/utils"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"gorm.io/gorm"
)

const bitcoinScanPageLimit = 10
const bitcoinEsploraPageSize = 25
const bitcoinReorgMinimumRewindBlocks = int64(12)
const bitcoinScanOverlapBlocks = int64(12)

var bitcoinRechargeScanOnce sync.Once

type esploraTransaction struct {
	TxID   string `json:"txid"`
	Status struct {
		Confirmed   bool   `json:"confirmed"`
		BlockHeight int64  `json:"block_height"`
		BlockHash   string `json:"block_hash"`
		BlockTime   int64  `json:"block_time"`
	} `json:"status"`
	Vin []struct {
		Prevout *struct {
			Address string `json:"scriptpubkey_address"`
		} `json:"prevout"`
	} `json:"vin"`
	Vout []struct {
		Address string `json:"scriptpubkey_address"`
		Value   int64  `json:"value"`
	} `json:"vout"`
}

func StartBitcoinRechargeScanTask() {
	bitcoinRechargeScanOnce.Do(func() {
		if !utils.GlobalConfig.Pay.Bitcoin.Enabled {
			return
		}
		go runBitcoinRechargeScanLoop()
		log.Printf("[btc_recharge] scan task started")
	})
}

func runBitcoinRechargeScanLoop() {
	for {
		RunBitcoinRechargeScanAllHosts()
		interval := utils.GlobalConfig.Pay.Bitcoin.ScanIntervalSeconds
		if interval < 10 {
			interval = 10
		}
		time.Sleep(time.Duration(interval) * time.Second)
	}
}

func RunBitcoinRechargeScanAllHosts() {
	utils.LimitPrefix(func(prefix string) {
		if err := RunBitcoinRechargeScanForPrefix(prefix); err != nil {
			log.Printf("[btc_recharge] scan prefix failed prefix=%s err=%v", prefix, err)
		}
	})
}

func RunBitcoinRechargeScanForPrefix(tablePrefix string) error {
	tablePrefix = strings.TrimSpace(tablePrefix)
	if tablePrefix == "" {
		return nil
	}
	db := utils.NewPrefixDb(tablePrefix)
	if db == nil {
		return errors.New("db_not_available")
	}
	if err := repository.ExpirePendingUsdtRechargeOrders(db); err != nil {
		return err
	}
	cfg, err := repository.LoadBitcoinRuntimeConfig(tablePrefix)
	if err != nil {
		return err
	}
	currentReceiveAddress := ""
	if cfg.Enabled {
		cfg, err = repository.ValidateBitcoinRuntimeConfig(tablePrefix)
		if err != nil {
			return err
		}
		currentReceiveAddress = cfg.ReceiveAddress
	}
	targets, err := repository.ListCryptoRechargeScanTargets(
		db,
		pojo.CryptoRechargeNetworkBitcoin,
		pojo.CryptoRechargeTokenBTC,
		currentReceiveAddress,
		"",
	)
	if err != nil {
		return err
	}
	var scanErrors []error
	for _, target := range targets {
		if ownershipErr := repository.ValidateCryptoScanTargetOwnership(tablePrefix, pojo.CryptoRechargeNetworkBitcoin, target.ReceiveAddress); ownershipErr != nil {
			scanErrors = append(scanErrors, fmt.Errorf("address %s ownership: %w", target.ReceiveAddress, ownershipErr))
			continue
		}
		targetCfg := cfg
		targetCfg.ReceiveAddress = target.ReceiveAddress
		if targetErr := runBitcoinRechargeScanTarget(db, tablePrefix, targetCfg); targetErr != nil {
			scanErrors = append(scanErrors, fmt.Errorf("address %s: %w", target.ReceiveAddress, targetErr))
		}
	}
	return errors.Join(scanErrors...)
}

func runBitcoinRechargeScanTarget(db *gorm.DB, tablePrefix string, cfg repository.BitcoinRuntimeConfig) error {
	lockKey := fmt.Sprintf("crypto_recharge_scan:%s:%s:%s", tablePrefix, pojo.CryptoRechargeNetworkBitcoin, cfg.ReceiveAddress)
	scanLock, acquired, err := utils.AcquireOwnedLock(lockKey, 5*time.Minute)
	if err != nil {
		return err
	}
	if !acquired {
		return nil
	}
	defer scanLock.Release()
	heartbeat := startCryptoScanLockHeartbeat(scanLock)
	defer heartbeat.Stop()

	minTimestamp, _, err := repository.GetCryptoRechargeScanMinTimestampForTarget(
		db,
		pojo.CryptoRechargeNetworkBitcoin,
		pojo.CryptoRechargeTokenBTC,
		cfg.ReceiveAddress,
		"",
	)
	if err != nil {
		return err
	}
	cursor, err := repository.GetCryptoRechargeScanCursor(db, pojo.CryptoRechargeNetworkBitcoin, pojo.CryptoRechargeTokenBTC, cfg.ReceiveAddress)
	if err != nil {
		return err
	}
	txs, nextCursor, err := fetchEsploraBitcoinTransactions(cfg, minTimestamp, cursor)
	if err != nil {
		return err
	}
	for _, txRecord := range txs {
		if err := repository.ProcessCryptoRechargeTx(db, tablePrefix, txRecord); err != nil {
			return fmt.Errorf("process tx %s: %w", txRecord.TxID, err)
		}
	}
	if nextCursor.LastBlockHeight < cursor.LastBlockHeight {
		if err = repository.MarkCryptoRechargeReorgWindowForManualReview(
			db,
			pojo.CryptoRechargeNetworkBitcoin,
			nextCursor.LastBlockHeight,
		); err != nil {
			return err
		}
	}
	if err = heartbeat.StopAndVerify(); err != nil {
		return err
	}
	return repository.SaveCryptoRechargeScanCursor(db, nextCursor)
}

func fetchEsploraBitcoinTransactions(
	cfg repository.BitcoinRuntimeConfig,
	minTimestamp int64,
	cursor pojo.CryptoRechargeScanCursor,
) ([]pojo.UsdtRechargeTx, pojo.CryptoRechargeScanCursor, error) {
	tipHeight, err := requestEsploraTipHeight(cfg)
	if err != nil {
		return nil, cursor, err
	}
	confirmedHeight := tipHeight - int64(cfg.MinConfirmations) + 1
	if confirmedHeight <= 0 {
		return nil, cursor, nil
	}
	if cursor.LastBlockHeight > 0 && cursor.LastBlockHash != "" {
		storedHash, hashErr := requestEsploraBlockHash(cfg, cursor.LastBlockHeight)
		if hashErr != nil {
			return nil, cursor, hashErr
		}
		if !strings.EqualFold(storedHash, cursor.LastBlockHash) {
			rewindDepth := int64(cfg.MinConfirmations * 2)
			if rewindDepth < bitcoinReorgMinimumRewindBlocks {
				rewindDepth = bitcoinReorgMinimumRewindBlocks
			}
			rewindHeight := cursor.LastBlockHeight - rewindDepth
			rewindHash := ""
			if rewindHeight > 0 {
				rewindHash, hashErr = requestEsploraBlockHash(cfg, rewindHeight)
				if hashErr != nil {
					return nil, cursor, hashErr
				}
			} else {
				rewindHeight = 0
			}
			log.Printf("[btc_recharge] reorg detected height=%d rewindHeight=%d", cursor.LastBlockHeight, rewindHeight)
			cursor.LastBlockHeight = rewindHeight
			cursor.LastBlockHash = rewindHash
			cursor.PaginationCursor = ""
			cursor.TargetBlockHeight = 0
			return nil, cursor, nil
		}
	}
	if cursor.LastBlockHeight <= 0 && minTimestamp <= 0 {
		tipHash, hashErr := requestEsploraBlockHash(cfg, confirmedHeight)
		if hashErr != nil {
			return nil, cursor, hashErr
		}
		cursor.LastBlockHeight = confirmedHeight
		cursor.LastBlockHash = tipHash
		return nil, cursor, nil
	}
	result := make([]pojo.UsdtRechargeTx, 0)
	targetHeight := cursor.TargetBlockHeight
	if targetHeight <= 0 {
		targetHeight = confirmedHeight
	}
	lastSeenTxID := strings.TrimSpace(cursor.PaginationCursor)
	completed := false
	for page := 0; page < bitcoinScanPageLimit; page++ {
		txs, requestErr := requestEsploraAddressTransactions(cfg, lastSeenTxID)
		if requestErr != nil {
			return result, cursor, requestErr
		}
		if len(txs) == 0 {
			completed = true
			break
		}
		lastConfirmedTxID := ""
		for _, tx := range txs {
			if !tx.Status.Confirmed || tx.Status.BlockHeight <= 0 {
				continue
			}
			lastConfirmedTxID = strings.TrimSpace(tx.TxID)
			if tx.Status.BlockHeight > targetHeight {
				continue
			}
			if cursor.LastBlockHeight > 0 && tx.Status.BlockHeight < cursor.LastBlockHeight {
				completed = true
				continue
			}
			confirmations := tipHeight - tx.Status.BlockHeight + 1
			if confirmations < int64(cfg.MinConfirmations) {
				continue
			}
			if minTimestamp > 0 && tx.Status.BlockTime*1000 < minTimestamp {
				completed = true
				continue
			}
			if record, ok := buildBitcoinRechargeTxRecord(cfg, tx, confirmations); ok {
				result = append(result, record)
			}
		}
		if completed || len(txs) < bitcoinEsploraPageSize {
			completed = true
			break
		}
		if lastConfirmedTxID == "" {
			return result, cursor, errors.New("btc_pagination_cursor_unavailable")
		}
		lastSeenTxID = lastConfirmedTxID
		cursor.PaginationCursor = lastSeenTxID
		cursor.TargetBlockHeight = targetHeight
	}
	if completed {
		checkpointHeight := targetHeight - bitcoinScanOverlapBlocks
		if checkpointHeight < cursor.LastBlockHeight {
			checkpointHeight = cursor.LastBlockHeight
		}
		checkpointHash := ""
		if checkpointHeight > 0 {
			var hashErr error
			checkpointHash, hashErr = requestEsploraBlockHash(cfg, checkpointHeight)
			if hashErr != nil {
				return result, cursor, hashErr
			}
		}
		cursor.LastBlockHeight = checkpointHeight
		cursor.LastBlockHash = checkpointHash
		cursor.PaginationCursor = ""
		cursor.TargetBlockHeight = 0
	} else if cursor.PaginationCursor == "" {
		return result, cursor, errors.New("btc_scan_page_limit_reached_without_cursor")
	}
	return result, cursor, nil
}

func requestEsploraTipHeight(cfg repository.BitcoinRuntimeConfig) (int64, error) {
	body, err := requestBitcoinEsplora(cfg.EsploraAPIBaseURL + "/blocks/tip/height")
	if err != nil {
		return 0, err
	}
	height, err := strconv.ParseInt(strings.TrimSpace(string(body)), 10, 64)
	if err != nil || height <= 0 {
		return 0, errors.New("btc_tip_height_invalid")
	}
	return height, nil
}

func requestEsploraBlockHash(cfg repository.BitcoinRuntimeConfig, height int64) (string, error) {
	body, err := requestBitcoinEsplora(cfg.EsploraAPIBaseURL + "/block-height/" + strconv.FormatInt(height, 10))
	if err != nil {
		return "", err
	}
	hash := strings.ToLower(strings.TrimSpace(string(body)))
	if len(hash) != 64 {
		return "", errors.New("btc_block_hash_invalid")
	}
	return hash, nil
}

func requestEsploraAddressTransactions(cfg repository.BitcoinRuntimeConfig, lastSeenTxID string) ([]esploraTransaction, error) {
	base := cfg.EsploraAPIBaseURL + "/address/" + url.PathEscape(cfg.ReceiveAddress) + "/txs"
	if strings.TrimSpace(lastSeenTxID) != "" {
		base += "/chain/" + url.PathEscape(strings.TrimSpace(lastSeenTxID))
	}
	body, err := requestBitcoinEsplora(base)
	if err != nil {
		return nil, err
	}
	var result []esploraTransaction
	if err = json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func requestBitcoinEsplora(requestURL string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := (&http.Client{Timeout: 15 * time.Second}).Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 16<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("btc_esplora_status_%d: %s", resp.StatusCode, string(body))
	}
	return body, nil
}

func buildBitcoinRechargeTxRecord(cfg repository.BitcoinRuntimeConfig, tx esploraTransaction, confirmations int64) (pojo.UsdtRechargeTx, bool) {
	txID := strings.TrimSpace(tx.TxID)
	if len(txID) != 64 {
		return pojo.UsdtRechargeTx{}, false
	}
	blockHash := strings.TrimSpace(tx.Status.BlockHash)
	if _, err := hex.DecodeString(txID); err != nil || len(blockHash) != 64 {
		return pojo.UsdtRechargeTx{}, false
	}
	if _, err := hex.DecodeString(blockHash); err != nil {
		return pojo.UsdtRechargeTx{}, false
	}
	// A platform sweep can return change to the receive address. Never treat a
	// transaction spending that address as a new external deposit.
	for _, input := range tx.Vin {
		if input.Prevout != nil && strings.TrimSpace(input.Prevout.Address) == strings.TrimSpace(cfg.ReceiveAddress) {
			return pojo.UsdtRechargeTx{}, false
		}
	}
	amountSatoshi := int64(0)
	for _, output := range tx.Vout {
		if strings.TrimSpace(output.Address) == strings.TrimSpace(cfg.ReceiveAddress) && output.Value > 0 {
			amountSatoshi += output.Value
		}
	}
	if amountSatoshi <= 0 {
		return pojo.UsdtRechargeTx{}, false
	}
	fromAddress := ""
	for _, input := range tx.Vin {
		if input.Prevout != nil && strings.TrimSpace(input.Prevout.Address) != "" {
			fromAddress = strings.TrimSpace(input.Prevout.Address)
			break
		}
	}
	raw, _ := json.Marshal(tx)
	return pojo.UsdtRechargeTx{
		TxID:           txID,
		Network:        pojo.CryptoRechargeNetworkBitcoin,
		Token:          pojo.CryptoRechargeTokenBTC,
		FromAddress:    fromAddress,
		ToAddress:      strings.TrimSpace(cfg.ReceiveAddress),
		AmountDecimals: pojo.CryptoRechargeDecimalsBTC,
		AmountAtomic:   strconv.FormatInt(amountSatoshi, 10),
		Amount:         formatBitcoinSatoshi(amountSatoshi),
		BlockHeight:    tx.Status.BlockHeight,
		BlockHash:      strings.ToLower(strings.TrimSpace(tx.Status.BlockHash)),
		Confirmations:  int(confirmations),
		BlockTimestamp: tx.Status.BlockTime * 1000,
		RawJSON:        string(raw),
	}, true
}

func formatBitcoinSatoshi(value int64) string {
	return fmt.Sprintf("%d.%08d", value/100000000, value%100000000)
}
