package services

import (
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/repository"
	"BaseGoUni/core/utils"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/hibiken/asynq"
)

const (
	usdtTrc20ScanPageLimit     = 10
	usdtTrc20ScanLimit         = 200
	TaskTypeUsdtRechargeExpire = "usdt_recharge:expire"
)

var usdtRechargeScanOnce sync.Once

type UsdtRechargeExpirePayload struct {
	TablePrefix string `json:"tablePrefix"`
	OrderNo     string `json:"orderNo"`
}

type tronGridTRC20Resp struct {
	Data []tronGridTRC20Tx `json:"data"`
	Meta struct {
		Fingerprint string `json:"fingerprint"`
	} `json:"meta"`
}

type tronGridTRC20Tx struct {
	TransactionID string `json:"transaction_id"`
	TokenInfo     struct {
		Address  string `json:"address"`
		Decimals int    `json:"decimals"`
		Symbol   string `json:"symbol"`
	} `json:"token_info"`
	BlockTimestamp int64  `json:"block_timestamp"`
	From           string `json:"from"`
	To             string `json:"to"`
	Type           string `json:"type"`
	Value          string `json:"value"`
}

func StartUsdtRechargeScanTask() {
	usdtRechargeScanOnce.Do(func() {
		if !utils.GlobalConfig.Pay.UsdtTrc20.Enabled {
			return
		}
		go runUsdtRechargeScanLoop()
		log.Printf("[usdt_recharge] scan task started")
	})
}

func runUsdtRechargeScanLoop() {
	for {
		interval := usdtRechargeScanInterval()
		RunUsdtRechargeScanAllHosts()
		time.Sleep(interval)
	}
}

func usdtRechargeScanInterval() time.Duration {
	seconds := utils.GlobalConfig.Pay.UsdtTrc20.ScanIntervalSeconds
	if seconds <= 0 {
		seconds = 10
	}
	if seconds < 10 {
		seconds = 10
	}
	return time.Duration(seconds) * time.Second
}

func RunUsdtRechargeScanAllHosts() {
	utils.LimitPrefix(func(prefix string) {
		if err := RunUsdtRechargeScanForPrefix(prefix); err != nil {
			log.Printf("[usdt_recharge] scan prefix failed prefix=%s err=%v", prefix, err)
		}
	})
}

func RunUsdtRechargeScanForPrefix(tablePrefix string) error {
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
	cfg, err := repository.LoadUsdtTrc20RuntimeConfig(tablePrefix)
	if err != nil {
		return err
	}
	if !cfg.Enabled || strings.TrimSpace(cfg.ReceiveAddress) == "" {
		return nil
	}
	minTimestamp, pendingCount, err := repository.GetUsdtRechargeScanMinTimestamp(db)
	if err != nil {
		return err
	}
	if pendingCount == 0 {
		return nil
	}
	txs, err := fetchTronGridTRC20Transactions(cfg, minTimestamp)
	if err != nil {
		return err
	}
	for _, txRecord := range txs {
		if err := repository.ProcessUsdtRechargeTx(db, tablePrefix, txRecord); err != nil {
			log.Printf("[usdt_recharge] process tx failed prefix=%s tx=%s err=%v", tablePrefix, txRecord.TxID, err)
		}
	}
	return nil
}

func fetchTronGridTRC20Transactions(cfg repository.UsdtTrc20RuntimeConfig, minTimestamp int64) ([]pojo.UsdtRechargeTx, error) {
	result := make([]pojo.UsdtRechargeTx, 0)
	fingerprint := ""
	for page := 0; page < usdtTrc20ScanPageLimit; page++ {
		resp, err := requestTronGridTRC20Page(cfg, minTimestamp, fingerprint)
		if err != nil {
			return result, err
		}
		for _, tx := range resp.Data {
			record, ok := buildUsdtRechargeTxRecord(cfg, tx)
			if ok {
				result = append(result, record)
			}
		}
		fingerprint = strings.TrimSpace(resp.Meta.Fingerprint)
		if fingerprint == "" {
			break
		}
	}
	return result, nil
}

func requestTronGridTRC20Page(cfg repository.UsdtTrc20RuntimeConfig, minTimestamp int64, fingerprint string) (tronGridTRC20Resp, error) {
	var result tronGridTRC20Resp
	baseURL := strings.TrimRight(cfg.APIBaseURL, "/")
	reqURL := fmt.Sprintf("%s/v1/accounts/%s/transactions/trc20", baseURL, url.PathEscape(cfg.ReceiveAddress))
	query := url.Values{}
	query.Set("only_confirmed", "true")
	query.Set("only_to", "true")
	query.Set("contract_address", cfg.ContractAddress)
	query.Set("limit", strconv.Itoa(usdtTrc20ScanLimit))
	query.Set("order_by", "block_timestamp,desc")
	if minTimestamp > 0 {
		query.Set("min_timestamp", strconv.FormatInt(minTimestamp, 10))
	}
	if strings.TrimSpace(fingerprint) != "" {
		query.Set("fingerprint", strings.TrimSpace(fingerprint))
	}
	reqURL += "?" + query.Encode()

	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return result, err
	}
	if cfg.APIKey != "" {
		req.Header.Set("TRON-PRO-API-KEY", cfg.APIKey)
	}
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return result, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return result, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return result, fmt.Errorf("trongrid_status_%d: %s", resp.StatusCode, string(body))
	}
	if err = json.Unmarshal(body, &result); err != nil {
		return result, err
	}
	return result, nil
}

func buildUsdtRechargeTxRecord(cfg repository.UsdtTrc20RuntimeConfig, tx tronGridTRC20Tx) (pojo.UsdtRechargeTx, bool) {
	if strings.TrimSpace(tx.TransactionID) == "" {
		return pojo.UsdtRechargeTx{}, false
	}
	if !strings.EqualFold(strings.TrimSpace(tx.To), strings.TrimSpace(cfg.ReceiveAddress)) {
		return pojo.UsdtRechargeTx{}, false
	}
	if !strings.EqualFold(strings.TrimSpace(tx.TokenInfo.Address), strings.TrimSpace(cfg.ContractAddress)) {
		return pojo.UsdtRechargeTx{}, false
	}
	amountMicro, err := normalizeTronTokenValueToMicro(tx.Value, tx.TokenInfo.Decimals)
	if err != nil || amountMicro <= 0 {
		return pojo.UsdtRechargeTx{}, false
	}
	raw, _ := json.Marshal(tx)
	return pojo.UsdtRechargeTx{
		TxID:            strings.TrimSpace(tx.TransactionID),
		Network:         pojo.CryptoRechargeNetworkTRC20,
		Token:           pojo.CryptoRechargeTokenUSDT,
		ContractAddress: strings.TrimSpace(tx.TokenInfo.Address),
		FromAddress:     strings.TrimSpace(tx.From),
		ToAddress:       strings.TrimSpace(tx.To),
		AmountMicro:     amountMicro,
		Amount:          repository.FormatUsdtMicroAmount(amountMicro),
		BlockTimestamp:  tx.BlockTimestamp,
		RawJSON:         string(raw),
	}, true
}

func normalizeTronTokenValueToMicro(value string, decimals int) (int64, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, errors.New("empty_token_value")
	}
	tokenValue, ok := new(big.Int).SetString(value, 10)
	if !ok || tokenValue.Sign() <= 0 {
		return 0, errors.New("invalid_token_value")
	}
	if decimals < 0 {
		return 0, errors.New("invalid_token_decimals")
	}
	if decimals > 6 {
		divisor := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals-6)), nil)
		tokenValue.Quo(tokenValue, divisor)
	} else if decimals < 6 {
		multiplier := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(6-decimals)), nil)
		tokenValue.Mul(tokenValue, multiplier)
	}
	if !tokenValue.IsInt64() {
		return 0, errors.New("token_value_overflow")
	}
	return tokenValue.Int64(), nil
}

func EnqueueUsdtRechargeExpireTask(tablePrefix string, orderNo string, expireAt time.Time) error {
	tablePrefix = strings.TrimSpace(tablePrefix)
	orderNo = strings.TrimSpace(orderNo)
	if asynqClient == nil || tablePrefix == "" || orderNo == "" || expireAt.IsZero() {
		return nil
	}
	payload, _ := json.Marshal(UsdtRechargeExpirePayload{
		TablePrefix: tablePrefix,
		OrderNo:     orderNo,
	})
	task := asynq.NewTask(TaskTypeUsdtRechargeExpire, payload)
	_, err := asynqClient.Enqueue(task, asynq.ProcessAt(expireAt), asynq.MaxRetry(5))
	return err
}

func handleUsdtRechargeExpireTask(ctx context.Context, task *asynq.Task) error {
	var payload UsdtRechargeExpirePayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return err
	}
	payload.TablePrefix = strings.TrimSpace(payload.TablePrefix)
	payload.OrderNo = strings.TrimSpace(payload.OrderNo)
	if payload.TablePrefix == "" || payload.OrderNo == "" {
		return nil
	}
	db := utils.NewPrefixDb(payload.TablePrefix)
	if db == nil {
		return fmt.Errorf("db not ready for prefix=%s", payload.TablePrefix)
	}
	return repository.ExpireUsdtRechargeOrderByOrderNo(db, payload.OrderNo)
}
