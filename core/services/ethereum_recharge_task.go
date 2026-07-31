package services

import (
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/repository"
	"BaseGoUni/core/utils"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"gorm.io/gorm"
)

const (
	ethereumBlockBatchSize           = int64(20)
	ethereumMaxScanBlocks            = int64(1000)
	ethereumReorgMinimumRewindBlocks = int64(64)
)

var (
	ethereumRechargeScanOnce sync.Once
)

type ethereumRPCRequest struct {
	JSONRPC string `json:"jsonrpc"`
	ID      int64  `json:"id"`
	Method  string `json:"method"`
	Params  []any  `json:"params"`
}

type ethereumRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type ethereumRPCResponse struct {
	JSONRPC string            `json:"jsonrpc"`
	ID      int64             `json:"id"`
	Result  json.RawMessage   `json:"result"`
	Error   *ethereumRPCError `json:"error"`
}

type ethereumBlock struct {
	Number       string                `json:"number"`
	Hash         string                `json:"hash"`
	Timestamp    string                `json:"timestamp"`
	Transactions []ethereumTransaction `json:"transactions"`
}

type ethereumTransaction struct {
	Hash  string  `json:"hash"`
	From  string  `json:"from"`
	To    *string `json:"to"`
	Value string  `json:"value"`
}

type ethereumReceipt struct {
	Status          string `json:"status"`
	TransactionHash string `json:"transactionHash"`
	BlockHash       string `json:"blockHash"`
	BlockNumber     string `json:"blockNumber"`
}

func StartEthereumRechargeScanTask() {
	ethereumRechargeScanOnce.Do(func() {
		if !utils.GlobalConfig.Pay.Ethereum.Enabled {
			return
		}
		go runEthereumRechargeScanLoop()
		log.Printf("[eth_recharge] scan task started")
	})
}

func runEthereumRechargeScanLoop() {
	for {
		RunEthereumRechargeScanAllHosts()
		interval := utils.GlobalConfig.Pay.Ethereum.ScanIntervalSeconds
		if interval < 10 {
			interval = 10
		}
		time.Sleep(time.Duration(interval) * time.Second)
	}
}

func RunEthereumRechargeScanAllHosts() {
	utils.LimitPrefix(func(prefix string) {
		if err := RunEthereumRechargeScanForPrefix(prefix); err != nil {
			log.Printf("[eth_recharge] scan prefix failed prefix=%s err=%v", prefix, err)
		}
	})
}

func RunEthereumRechargeScanForPrefix(tablePrefix string) error {
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
	cfg, err := repository.LoadEthereumRuntimeConfig(tablePrefix)
	if err != nil {
		return err
	}
	currentReceiveAddress := ""
	if cfg.Enabled {
		cfg, err = repository.ValidateEthereumRuntimeConfig(tablePrefix)
		if err != nil {
			return err
		}
		currentReceiveAddress = cfg.ReceiveAddress
	}
	targets, err := repository.ListCryptoRechargeScanTargets(
		db,
		pojo.CryptoRechargeNetworkEthereum,
		pojo.CryptoRechargeTokenETH,
		currentReceiveAddress,
		"",
	)
	if err != nil {
		return err
	}
	var scanErrors []error
	for _, target := range targets {
		if ownershipErr := repository.ValidateCryptoScanTargetOwnership(tablePrefix, pojo.CryptoRechargeNetworkEthereum, target.ReceiveAddress); ownershipErr != nil {
			scanErrors = append(scanErrors, fmt.Errorf("address %s ownership: %w", target.ReceiveAddress, ownershipErr))
			continue
		}
		targetCfg := cfg
		targetCfg.ReceiveAddress = target.ReceiveAddress
		if targetErr := runEthereumRechargeScanTarget(db, tablePrefix, targetCfg); targetErr != nil {
			scanErrors = append(scanErrors, fmt.Errorf("address %s: %w", target.ReceiveAddress, targetErr))
		}
	}
	return errors.Join(scanErrors...)
}

func runEthereumRechargeScanTarget(db *gorm.DB, tablePrefix string, cfg repository.EthereumRuntimeConfig) error {
	lockKey := fmt.Sprintf("crypto_recharge_scan:%s:%s:%s", tablePrefix, pojo.CryptoRechargeNetworkEthereum, strings.ToLower(cfg.ReceiveAddress))
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
		pojo.CryptoRechargeNetworkEthereum,
		pojo.CryptoRechargeTokenETH,
		strings.ToLower(cfg.ReceiveAddress),
		"",
	)
	if err != nil {
		return err
	}
	cursor, err := repository.GetCryptoRechargeScanCursor(db, pojo.CryptoRechargeNetworkEthereum, pojo.CryptoRechargeTokenETH, cfg.ReceiveAddress)
	if err != nil {
		return err
	}
	txs, confirmedHeight, confirmedHash, err := fetchEthereumTransactions(cfg, minTimestamp, cursor)
	if err != nil {
		return err
	}
	for _, txRecord := range txs {
		if err := repository.ProcessCryptoRechargeTx(db, tablePrefix, txRecord); err != nil {
			return fmt.Errorf("process tx %s: %w", txRecord.TxID, err)
		}
	}
	if confirmedHeight < cursor.LastBlockHeight {
		if err = repository.MarkCryptoRechargeReorgWindowForManualReview(
			db,
			pojo.CryptoRechargeNetworkEthereum,
			confirmedHeight,
		); err != nil {
			return err
		}
	}
	if err = heartbeat.StopAndVerify(); err != nil {
		return err
	}
	cursor.LastBlockHeight = confirmedHeight
	cursor.LastBlockHash = confirmedHash
	cursor.PaginationCursor = ""
	cursor.TargetBlockHeight = 0
	return repository.SaveCryptoRechargeScanCursor(db, cursor)
}

func fetchEthereumTransactions(
	cfg repository.EthereumRuntimeConfig,
	minTimestamp int64,
	cursor pojo.CryptoRechargeScanCursor,
) ([]pojo.UsdtRechargeTx, int64, string, error) {
	chainID, err := ethereumRPCString(cfg.RPCURL, "eth_chainId", []any{})
	if err != nil {
		return nil, 0, "", err
	}
	if !strings.EqualFold(strings.TrimSpace(chainID), "0x1") {
		return nil, 0, "", errors.New("eth_rpc_not_mainnet")
	}
	latestHex, err := ethereumRPCString(cfg.RPCURL, "eth_blockNumber", []any{})
	if err != nil {
		return nil, 0, "", err
	}
	latestHeight, err := parseEthereumHexInt64(latestHex)
	if err != nil {
		return nil, 0, "", err
	}
	confirmedHeight := latestHeight - int64(cfg.MinConfirmations) + 1
	if confirmedHeight <= 0 {
		return nil, confirmedHeight, "", nil
	}
	if cursor.LastBlockHeight > 0 && cursor.LastBlockHash != "" {
		cursorBlock, blockErr := requestEthereumBlock(cfg.RPCURL, cursor.LastBlockHeight)
		if blockErr != nil {
			return nil, 0, "", blockErr
		}
		if !strings.EqualFold(strings.TrimSpace(cursorBlock.Hash), strings.TrimSpace(cursor.LastBlockHash)) {
			rewindDepth := int64(cfg.MinConfirmations * 2)
			if rewindDepth < ethereumReorgMinimumRewindBlocks {
				rewindDepth = ethereumReorgMinimumRewindBlocks
			}
			rewindHeight := cursor.LastBlockHeight - rewindDepth
			rewindHash := ""
			if rewindHeight > 0 {
				rewindBlock, rewindErr := requestEthereumBlock(cfg.RPCURL, rewindHeight)
				if rewindErr != nil {
					return nil, 0, "", rewindErr
				}
				rewindHash = strings.ToLower(strings.TrimSpace(rewindBlock.Hash))
			} else {
				rewindHeight = 0
			}
			log.Printf("[eth_recharge] reorg detected height=%d rewindHeight=%d", cursor.LastBlockHeight, rewindHeight)
			return nil, rewindHeight, rewindHash, nil
		}
	}
	if cursor.LastBlockHeight <= 0 && minTimestamp <= 0 {
		confirmedBlock, blockErr := requestEthereumBlock(cfg.RPCURL, confirmedHeight)
		if blockErr != nil {
			return nil, 0, "", blockErr
		}
		return nil, confirmedHeight, confirmedBlock.Hash, nil
	}

	var blocks []ethereumBlock
	scannedHeight := confirmedHeight
	if cursor.LastBlockHeight > 0 {
		if cursor.LastBlockHeight > confirmedHeight {
			return nil, 0, "", errors.New("eth_confirmed_height_behind_cursor")
		}
		if cursor.LastBlockHeight == confirmedHeight {
			return nil, confirmedHeight, cursor.LastBlockHash, nil
		}
		if confirmedHeight-cursor.LastBlockHeight > ethereumMaxScanBlocks {
			scannedHeight = cursor.LastBlockHeight + ethereumMaxScanBlocks
		}
		blocks, err = requestEthereumBlockRange(cfg.RPCURL, cursor.LastBlockHeight+1, scannedHeight)
	} else {
		blocks, err = requestEthereumBlocksSince(cfg.RPCURL, confirmedHeight, minTimestamp)
	}
	if err != nil {
		return nil, 0, "", err
	}
	result := make([]pojo.UsdtRechargeTx, 0)
	receiveAddress := strings.ToLower(strings.TrimSpace(cfg.ReceiveAddress))
	for _, block := range blocks {
		blockTimestamp, timestampErr := parseEthereumHexInt64(block.Timestamp)
		if timestampErr != nil {
			continue
		}
		blockHeight, heightErr := parseEthereumHexInt64(block.Number)
		if heightErr != nil {
			continue
		}
		for _, tx := range block.Transactions {
			if tx.To == nil || strings.ToLower(strings.TrimSpace(*tx.To)) != receiveAddress {
				continue
			}
			if strings.ToLower(strings.TrimSpace(tx.From)) == receiveAddress {
				continue
			}
			atomicAmount, amountErr := parseEthereumHexBigInt(tx.Value)
			if amountErr != nil || atomicAmount.Sign() <= 0 {
				continue
			}
			receipt, receiptErr := requestEthereumReceipt(cfg.RPCURL, tx.Hash)
			if receiptErr != nil {
				return result, 0, "", receiptErr
			}
			if !strings.EqualFold(receipt.Status, "0x1") {
				continue
			}
			if !strings.EqualFold(strings.TrimSpace(receipt.TransactionHash), strings.TrimSpace(tx.Hash)) ||
				!strings.EqualFold(strings.TrimSpace(receipt.BlockHash), strings.TrimSpace(block.Hash)) ||
				!strings.EqualFold(strings.TrimSpace(receipt.BlockNumber), strings.TrimSpace(block.Number)) {
				return result, 0, "", errors.New("eth_receipt_block_mismatch")
			}
			raw, _ := json.Marshal(tx)
			result = append(result, pojo.UsdtRechargeTx{
				TxID:           strings.TrimSpace(tx.Hash),
				Network:        pojo.CryptoRechargeNetworkEthereum,
				Token:          pojo.CryptoRechargeTokenETH,
				FromAddress:    strings.ToLower(strings.TrimSpace(tx.From)),
				ToAddress:      receiveAddress,
				AmountDecimals: pojo.CryptoRechargeDecimalsETH,
				AmountAtomic:   atomicAmount.String(),
				Amount:         formatEthereumWei(atomicAmount),
				BlockHeight:    blockHeight,
				BlockHash:      strings.ToLower(strings.TrimSpace(block.Hash)),
				Confirmations:  int(latestHeight-blockHeight) + 1,
				BlockTimestamp: blockTimestamp * 1000,
				RawJSON:        string(raw),
			})
		}
	}
	confirmedHash := ""
	for i := len(blocks) - 1; i >= 0; i-- {
		height, heightErr := parseEthereumHexInt64(blocks[i].Number)
		if heightErr == nil && height == scannedHeight {
			confirmedHash = strings.ToLower(strings.TrimSpace(blocks[i].Hash))
			break
		}
	}
	if confirmedHash == "" {
		confirmedBlock, blockErr := requestEthereumBlock(cfg.RPCURL, scannedHeight)
		if blockErr != nil {
			return result, 0, "", blockErr
		}
		confirmedHash = strings.ToLower(strings.TrimSpace(confirmedBlock.Hash))
	}
	return result, scannedHeight, confirmedHash, nil
}

func requestEthereumBlocksSince(rpcURL string, confirmedHeight int64, minTimestamp int64) ([]ethereumBlock, error) {
	result := make([]ethereumBlock, 0)
	remaining := ethereumMaxScanBlocks
	reachedMinimumTimestamp := minTimestamp <= 0
	for end := confirmedHeight; end >= 0 && remaining > 0; {
		start := end - ethereumBlockBatchSize + 1
		if start < 0 {
			start = 0
		}
		blocks, err := requestEthereumBlockRange(rpcURL, start, end)
		if err != nil {
			return result, err
		}
		result = append(result, blocks...)
		remaining -= end - start + 1
		oldestTimestamp := int64(0)
		for _, block := range blocks {
			timestamp, timestampErr := parseEthereumHexInt64(block.Timestamp)
			if timestampErr == nil && (oldestTimestamp == 0 || timestamp < oldestTimestamp) {
				oldestTimestamp = timestamp
			}
		}
		if minTimestamp > 0 && oldestTimestamp > 0 && oldestTimestamp*1000 < minTimestamp {
			reachedMinimumTimestamp = true
			break
		}
		if start == 0 {
			break
		}
		end = start - 1
	}
	if !reachedMinimumTimestamp && remaining <= 0 {
		return result, errors.New("eth_scan_window_exceeded")
	}
	sort.Slice(result, func(i, j int) bool {
		left, _ := parseEthereumHexInt64(result[i].Number)
		right, _ := parseEthereumHexInt64(result[j].Number)
		return left < right
	})
	return result, nil
}

func requestEthereumBlockRange(rpcURL string, start int64, end int64) ([]ethereumBlock, error) {
	if end < start {
		return nil, nil
	}
	requests := make([]ethereumRPCRequest, 0, end-start+1)
	for blockNumber := start; blockNumber <= end; blockNumber++ {
		requests = append(requests, ethereumRPCRequest{
			JSONRPC: "2.0",
			ID:      blockNumber + 1,
			Method:  "eth_getBlockByNumber",
			Params:  []any{fmt.Sprintf("0x%x", blockNumber), true},
		})
	}
	responses, err := requestEthereumRPCBatch(rpcURL, requests)
	if err != nil {
		return nil, err
	}
	result := make([]ethereumBlock, 0, len(responses))
	seenHeights := make(map[int64]struct{}, len(responses))
	for _, response := range responses {
		if response.Error != nil {
			return nil, fmt.Errorf("eth_rpc_%d: %s", response.Error.Code, response.Error.Message)
		}
		var block ethereumBlock
		if len(response.Result) == 0 || string(response.Result) == "null" {
			return nil, errors.New("eth_block_not_found")
		}
		if err = json.Unmarshal(response.Result, &block); err != nil {
			return nil, err
		}
		height, heightErr := parseEthereumHexInt64(block.Number)
		if heightErr != nil || height < start || height > end || strings.TrimSpace(block.Hash) == "" {
			return nil, errors.New("eth_block_response_invalid")
		}
		if _, exists := seenHeights[height]; exists {
			return nil, errors.New("eth_block_response_duplicate")
		}
		seenHeights[height] = struct{}{}
		result = append(result, block)
	}
	if int64(len(result)) != end-start+1 {
		return nil, errors.New("eth_block_response_incomplete")
	}
	sort.Slice(result, func(i, j int) bool {
		left, _ := parseEthereumHexInt64(result[i].Number)
		right, _ := parseEthereumHexInt64(result[j].Number)
		return left < right
	})
	return result, nil
}

func requestEthereumBlock(rpcURL string, blockNumber int64) (ethereumBlock, error) {
	var result ethereumBlock
	blocks, err := requestEthereumBlockRange(rpcURL, blockNumber, blockNumber)
	if err != nil {
		return result, err
	}
	if len(blocks) != 1 || strings.TrimSpace(blocks[0].Hash) == "" {
		return result, errors.New("eth_block_not_found")
	}
	return blocks[0], nil
}

func requestEthereumReceipt(rpcURL string, txHash string) (ethereumReceipt, error) {
	var result ethereumReceipt
	raw, err := requestEthereumRPC(rpcURL, "eth_getTransactionReceipt", []any{txHash})
	if err != nil {
		return result, err
	}
	if len(raw) == 0 || string(raw) == "null" {
		return result, errors.New("eth_receipt_not_found")
	}
	err = json.Unmarshal(raw, &result)
	return result, err
}

func ethereumRPCString(rpcURL string, method string, params []any) (string, error) {
	raw, err := requestEthereumRPC(rpcURL, method, params)
	if err != nil {
		return "", err
	}
	var result string
	if err = json.Unmarshal(raw, &result); err != nil {
		return "", err
	}
	return result, nil
}

func requestEthereumRPC(rpcURL string, method string, params []any) (json.RawMessage, error) {
	responses, err := requestEthereumRPCBatch(rpcURL, []ethereumRPCRequest{{
		JSONRPC: "2.0",
		ID:      1,
		Method:  method,
		Params:  params,
	}})
	if err != nil {
		return nil, err
	}
	if len(responses) != 1 {
		return nil, errors.New("eth_rpc_response_invalid")
	}
	if responses[0].Error != nil {
		return nil, fmt.Errorf("eth_rpc_%d: %s", responses[0].Error.Code, responses[0].Error.Message)
	}
	return responses[0].Result, nil
}

func requestEthereumRPCBatch(rpcURL string, requests []ethereumRPCRequest) ([]ethereumRPCResponse, error) {
	payload, err := json.Marshal(requests)
	if err != nil {
		return nil, err
	}
	if len(requests) == 1 {
		payload, err = json.Marshal(requests[0])
		if err != nil {
			return nil, err
		}
	}
	req, err := http.NewRequest(http.MethodPost, rpcURL, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := (&http.Client{Timeout: 30 * time.Second}).Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 64<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("eth_rpc_status_%d: %s", resp.StatusCode, string(body))
	}
	if len(requests) == 1 {
		var response ethereumRPCResponse
		if err = json.Unmarshal(body, &response); err != nil {
			return nil, err
		}
		return []ethereumRPCResponse{response}, nil
	}
	var responses []ethereumRPCResponse
	if err = json.Unmarshal(body, &responses); err != nil {
		return nil, err
	}
	return responses, nil
}

func parseEthereumHexInt64(value string) (int64, error) {
	value = strings.TrimPrefix(strings.TrimSpace(value), "0x")
	if value == "" {
		return 0, errors.New("eth_hex_invalid")
	}
	return strconv.ParseInt(value, 16, 64)
}

func parseEthereumHexBigInt(value string) (*big.Int, error) {
	value = strings.TrimPrefix(strings.TrimSpace(value), "0x")
	if value == "" {
		return nil, errors.New("eth_hex_invalid")
	}
	result, ok := new(big.Int).SetString(value, 16)
	if !ok {
		return nil, errors.New("eth_hex_invalid")
	}
	return result, nil
}

func formatEthereumWei(value *big.Int) string {
	if value == nil {
		return ""
	}
	digits := value.String()
	if len(digits) <= pojo.CryptoRechargeDecimalsETH {
		digits = strings.Repeat("0", pojo.CryptoRechargeDecimalsETH-len(digits)+1) + digits
	}
	return digits[:len(digits)-pojo.CryptoRechargeDecimalsETH] + "." + digits[len(digits)-pojo.CryptoRechargeDecimalsETH:]
}
