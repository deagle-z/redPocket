package api

import (
	"bytes"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"time"

	"BaseGoUni/core/game"
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/repository"
	"BaseGoUni/core/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	ggrMethodUserBalance       = "user_balance"
	ggrMethodTransaction       = "transaction"
	ggrMessageInternalError    = "INTERNAL_ERROR"
	ggrMessageInsufficientFund = "INSUFFICIENT_USER_FUNDS"
)

type ggrWalletRequest struct {
	Method       string              `json:"method"`
	AgentCode    string              `json:"agent_code"`
	AgentSecret  string              `json:"agent_secret"`
	AgentBalance *json.Number        `json:"agent_balance"`
	UserCode     string              `json:"user_code"`
	UserBalance  *json.Number        `json:"user_balance"`
	GameType     string              `json:"game_type"`
	Info         string              `json:"info"`
	Slot         *ggrTransactionGame `json:"slot"`
	Live         *ggrTransactionGame `json:"live"`
	Sportsbook   *ggrTransactionGame `json:"SB"`
	Mini         *ggrTransactionGame `json:"MN"`
}

type ggrTransactionGame struct {
	ProviderCode string          `json:"provider_code"`
	GameCode     string          `json:"game_code"`
	BetType      string          `json:"type"`
	BetMoney     *json.Number    `json:"bet_money"`
	WinMoney     *json.Number    `json:"win_money"`
	RoundID      json.RawMessage `json:"round_id"`
	TxnID        string          `json:"txn_id"`
	TxnType      string          `json:"txn_type"`
}

type ggrWalletResponse struct {
	Status      int      `json:"status"`
	UserBalance *float64 `json:"user_balance,omitempty"`
	Message     string   `json:"msg,omitempty"`
}

type validatedGGRTransaction struct {
	AgentCode          string
	UserCode           string
	GameType           string
	Info               string
	ProviderCode       string
	GameCode           string
	BetType            string
	BetMoney           float64
	WinMoney           float64
	RoundID            string
	TxnID              string
	TxnType            string
	AgentBalance       *float64
	RequestUserBalance *float64
	Fingerprint        string
	ReceivedAt         time.Time
}

type ggrTransactionFingerprint struct {
	AgentCode          string  `json:"agent_code"`
	AgentBalance       *string `json:"agent_balance"`
	UserCode           string  `json:"user_code"`
	UserToken          string  `json:"user_token"`
	RequestUserBalance *string `json:"user_balance"`
	GameType           string  `json:"game_type"`
	Info               string  `json:"info"`
	ProviderCode       string  `json:"provider_code"`
	GameCode           string  `json:"game_code"`
	BetType            string  `json:"type"`
	BetMoney           string  `json:"bet_money"`
	WinMoney           string  `json:"win_money"`
	RoundID            string  `json:"round_id"`
	TxnID              string  `json:"txn_id"`
	TxnType            string  `json:"txn_type"`
}

type ggrTransactionOutcome struct {
	Status       int
	Message      string
	Decision     string
	Balance      float64
	StartBalance float64
	EndBalance   float64
	AppliedBet   bool
	BetFlow      float64
	UserID       int64
	TenantID     int64
	Idempotent   bool
}

// GGRGoldAPI handles both GGR balance queries and seamless transactions.
// GGR requires its response object directly rather than the application's
// standard response envelope.
func GGRGoldAPI(ctx *gin.Context) {
	startedAt := time.Now()
	logContext := ggrWalletLogContext(ctx)
	req, rawParams, err := decodeGGRWalletRequest(ctx)
	if err != nil {
		log.Printf(
			"[ggr_wallet] callback rejected %s stage=decode content_type=%q content_length=%d raw_params=%s cost=%s err=%v",
			logContext,
			ctx.GetHeader("Content-Type"),
			ctx.Request.ContentLength,
			rawParams,
			time.Since(startedAt).Round(time.Millisecond),
			err,
		)
		ggrWalletFailure(ctx, false, ggrMessageInternalError)
		return
	}
	cfg := game.GetGGRConfig()
	if cfg.AgentCode == "" || cfg.AgentSecret == "" {
		log.Printf(
			"[ggr_wallet] callback rejected %s method=%q stage=configuration agent_code_configured=%t agent_secret_configured=%t cost=%s",
			logContext,
			strings.TrimSpace(req.Method),
			cfg.AgentCode != "",
			cfg.AgentSecret != "",
			time.Since(startedAt).Round(time.Millisecond),
		)
		ggrWalletFailure(ctx, strings.TrimSpace(req.Method) == ggrMethodUserBalance, ggrMessageInternalError)
		return
	}
	if !secureStringEqual(strings.TrimSpace(req.AgentCode), cfg.AgentCode) || !secureStringEqual(req.AgentSecret, cfg.AgentSecret) {
		log.Printf(
			"[ggr_wallet] callback rejected %s method=%q stage=authentication agent_code=%q cost=%s",
			logContext,
			strings.TrimSpace(req.Method),
			strings.TrimSpace(req.AgentCode),
			time.Since(startedAt).Round(time.Millisecond),
		)
		ggrWalletFailure(ctx, strings.TrimSpace(req.Method) == ggrMethodUserBalance, ggrMessageInternalError)
		return
	}

	db := ctx.MustGet("db").(*gorm.DB)
	switch strings.TrimSpace(req.Method) {
	case ggrMethodUserBalance:
		handleGGRUserBalance(ctx, db, req, logContext, startedAt)
	case ggrMethodTransaction:
		handleGGRTransaction(ctx, db, cfg, req, rawParams, logContext, startedAt)
	default:
		log.Printf(
			"[ggr_wallet] callback rejected %s method=%q stage=dispatch cost=%s err=unsupported_method",
			logContext,
			strings.TrimSpace(req.Method),
			time.Since(startedAt).Round(time.Millisecond),
		)
		ggrWalletFailure(ctx, false, ggrMessageInternalError)
	}
}

func decodeGGRWalletRequest(ctx *gin.Context) (ggrWalletRequest, string, error) {
	var req ggrWalletRequest
	body, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		return req, ggrUnreadableRequestBodyForLog(nil), err
	}
	rawParams, rawParamsErr := redactGGRWalletRequestBody(body)
	if rawParamsErr != nil {
		rawParams = ggrUnreadableRequestBodyForLog(body)
	}
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.UseNumber()
	if err := decoder.Decode(&req); err != nil {
		return req, rawParams, err
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			return req, rawParams, errors.New("ggr request contains multiple json values")
		}
		return req, rawParams, err
	}
	return req, rawParams, nil
}

func handleGGRUserBalance(ctx *gin.Context, db *gorm.DB, req ggrWalletRequest, logContext string, startedAt time.Time) {
	userCode, err := validateGGRBalanceUserCode(req)
	if err != nil {
		log.Printf(
			"[ggr_wallet] balance rejected %s stage=validation user_code=%q cost=%s err=%v",
			logContext,
			strings.TrimSpace(req.UserCode),
			time.Since(startedAt).Round(time.Millisecond),
			err,
		)
		ggrWalletFailure(ctx, true, ggrMessageInternalError)
		return
	}

	var user pojo.TgUser
	err = db.Select("id, uid, balance, status").
		Where("uid = ? AND status <> ?", userCode, int8(-1)).
		First(&user).Error
	if err != nil {
		log.Printf(
			"[ggr_wallet] balance rejected %s stage=load_user user_code=%q cost=%s err=%v",
			logContext,
			userCode,
			time.Since(startedAt).Round(time.Millisecond),
			err,
		)
		ggrWalletFailure(ctx, true, ggrMessageInternalError)
		return
	}
	if user.Status != 1 {
		log.Printf(
			"[ggr_wallet] balance rejected %s stage=user_status user_code=%q user_id=%d user_status=%d cost=%s",
			logContext,
			userCode,
			user.ID,
			user.Status,
			time.Since(startedAt).Round(time.Millisecond),
		)
		ggrWalletFailure(ctx, true, ggrMessageInternalError)
		return
	}
	balance := utils.Truncate2(user.Balance)
	ctx.JSON(http.StatusOK, ggrWalletResponse{Status: 1, UserBalance: &balance})
}

func validateGGRBalanceUserCode(req ggrWalletRequest) (string, error) {
	userCode := strings.TrimSpace(req.UserCode)
	if userCode == "" {
		return "", errors.New("ggr user_code is required")
	}
	return userCode, nil
}

func handleGGRTransaction(ctx *gin.Context, db *gorm.DB, cfg game.GGRConfig, req ggrWalletRequest, rawParams string, logContext string, startedAt time.Time) {
	validated, err := validateGGRTransaction(req, cfg)
	if err != nil {
		log.Printf(
			"[ggr_wallet] transaction rejected %s stage=validation %s raw_params=%s cost=%s err=%v",
			logContext,
			ggrWalletRequestSummary(req),
			rawParams,
			time.Since(startedAt).Round(time.Millisecond),
			err,
		)
		ggrWalletFailure(ctx, false, ggrMessageInternalError)
		return
	}
	appGame, err := getGGRAppGame(db, validated.ProviderCode, validated.GameCode)
	if err != nil {
		log.Printf(
			"[ggr_wallet] transaction rejected %s stage=load_game user_code=%q provider=%q game=%q txn_id=%q txn_type=%q bet_money=%.2f win_money=%.2f has_payout=%t balance_delta=%.2f cost=%s err=%v",
			logContext,
			validated.UserCode,
			validated.ProviderCode,
			validated.GameCode,
			validated.TxnID,
			validated.TxnType,
			validated.BetMoney,
			validated.WinMoney,
			validated.WinMoney > 0,
			ggrBalanceDelta(validated.BetMoney, validated.WinMoney),
			time.Since(startedAt).Round(time.Millisecond),
			err,
		)
		ggrWalletFailure(ctx, false, ggrMessageInternalError)
		return
	}
	outcome, err := applyGGRTransaction(db, validated, appGame)
	if err != nil {
		log.Printf(
			"[ggr_wallet] transaction rolled back %s user_code=%q user_id=%d provider=%q game=%q txn_id=%q txn_type=%q start_balance=%.2f bet_money=%.2f win_money=%.2f has_payout=%t balance_delta=%.2f expected_end_balance=%.2f cost=%s err=%v",
			logContext,
			validated.UserCode,
			outcome.UserID,
			validated.ProviderCode,
			validated.GameCode,
			validated.TxnID,
			validated.TxnType,
			outcome.StartBalance,
			validated.BetMoney,
			validated.WinMoney,
			validated.WinMoney > 0,
			ggrBalanceDelta(validated.BetMoney, validated.WinMoney),
			outcome.EndBalance,
			time.Since(startedAt).Round(time.Millisecond),
			err,
		)
		ggrWalletFailure(ctx, false, ggrMessageInternalError)
		return
	}
	if outcome.Status != pojo.GGRTransactionResultSuccess {
		log.Printf(
			"[ggr_wallet] transaction rejected %s decision=%s user_code=%q user_id=%d provider=%q game=%q txn_id=%q txn_type=%q start_balance=%.2f bet_money=%.2f win_money=%.2f has_payout=%t balance_delta=%.2f end_balance=%.2f idempotent=%t status=%d msg=%q cost=%s",
			logContext,
			outcome.Decision,
			validated.UserCode,
			outcome.UserID,
			validated.ProviderCode,
			validated.GameCode,
			validated.TxnID,
			validated.TxnType,
			outcome.StartBalance,
			validated.BetMoney,
			validated.WinMoney,
			validated.WinMoney > 0,
			ggrBalanceDelta(validated.BetMoney, validated.WinMoney),
			outcome.EndBalance,
			outcome.Idempotent,
			outcome.Status,
			outcome.Message,
			time.Since(startedAt).Round(time.Millisecond),
		)
		ggrWalletFailure(ctx, false, outcome.Message)
		return
	}
	if outcome.AppliedBet && outcome.BetFlow > 0 && outcome.UserID > 0 {
		postProcessGGRBet(db, outcome)
	}
	balance := utils.Truncate2(outcome.Balance)
	if validated.WinMoney > 0 {
		log.Printf(
			"[ggr_wallet] payout success %s decision=%s user_code=%q user_id=%d provider=%q game=%q txn_id=%q txn_type=%q start_balance=%.2f bet_money=%.2f win_money=%.2f balance_delta=%.2f end_balance=%.2f response_balance=%.2f idempotent=%t cost=%s",
			logContext,
			outcome.Decision,
			validated.UserCode,
			outcome.UserID,
			validated.ProviderCode,
			validated.GameCode,
			validated.TxnID,
			validated.TxnType,
			outcome.StartBalance,
			validated.BetMoney,
			validated.WinMoney,
			ggrBalanceDelta(validated.BetMoney, validated.WinMoney),
			outcome.EndBalance,
			balance,
			outcome.Idempotent,
			time.Since(startedAt).Round(time.Millisecond),
		)
	}
	ctx.JSON(http.StatusOK, ggrWalletResponse{Status: 1, UserBalance: &balance})
}

func validateGGRTransaction(req ggrWalletRequest, cfg game.GGRConfig) (validatedGGRTransaction, error) {
	result := validatedGGRTransaction{
		AgentCode:  strings.TrimSpace(req.AgentCode),
		UserCode:   strings.TrimSpace(req.UserCode),
		GameType:   strings.TrimSpace(req.GameType),
		Info:       req.Info,
		ReceivedAt: time.Now(),
	}
	if result.UserCode == "" {
		return result, errors.New("ggr user_code is required")
	}

	transactionGame, err := ggrTransactionGameForType(req)
	if err != nil {
		return result, err
	}
	result.ProviderCode = strings.ToUpper(strings.TrimSpace(transactionGame.ProviderCode))
	result.GameCode = strings.TrimSpace(transactionGame.GameCode)
	result.BetType = strings.TrimSpace(transactionGame.BetType)
	result.TxnID = strings.TrimSpace(transactionGame.TxnID)
	result.TxnType = strings.ToLower(strings.TrimSpace(transactionGame.TxnType))
	if result.ProviderCode == "" || result.GameCode == "" || result.BetType == "" || result.TxnID == "" {
		return result, errors.New("ggr transaction identity fields are required")
	}
	if err := validateGGRTransactionFieldLengths(result); err != nil {
		return result, err
	}

	categoryCode, err := game.ResolveGGRProviderCategoryWithMap(result.ProviderCode, cfg.CategoryMap)
	if err != nil {
		return result, err
	}
	if !ggrCategoryMatchesGameType(categoryCode, result.GameType) {
		return result, fmt.Errorf("ggr provider category does not match game_type: provider=%s category=%s game_type=%s", result.ProviderCode, categoryCode, result.GameType)
	}

	result.BetMoney, err = parseRequiredGGRMoney(transactionGame.BetMoney, "bet_money")
	if err != nil {
		return result, err
	}
	result.WinMoney, err = parseRequiredGGRMoney(transactionGame.WinMoney, "win_money")
	if err != nil {
		return result, err
	}
	if result.BetMoney < 0 || result.WinMoney < 0 {
		return result, errors.New("ggr transaction money must be non-negative")
	}
	switch result.TxnType {
	case "debit":
		if result.BetMoney <= 0 || result.WinMoney != 0 {
			return result, errors.New("ggr debit requires bet_money > 0 and win_money = 0")
		}
	case "credit":
		if result.BetMoney != 0 || result.WinMoney <= 0 {
			return result, errors.New("ggr credit requires bet_money = 0 and win_money > 0")
		}
	case "debit_credit":
		if result.BetMoney == 0 && result.WinMoney == 0 {
			return result, errors.New("ggr debit_credit requires a non-zero amount")
		}
	default:
		return result, fmt.Errorf("ggr txn_type is invalid: %s", result.TxnType)
	}

	result.RoundID, err = parseGGRRoundID(transactionGame.RoundID)
	if err != nil {
		return result, err
	}
	if len(result.RoundID) > 255 {
		return result, errors.New("ggr round_id exceeds 255 bytes")
	}
	result.AgentBalance, err = parseOptionalGGRMoney(req.AgentBalance, "agent_balance")
	if err != nil {
		return result, err
	}
	result.RequestUserBalance, err = parseOptionalGGRMoney(req.UserBalance, "user_balance")
	if err != nil {
		return result, err
	}
	result.Fingerprint, err = fingerprintGGRTransaction(result)
	if err != nil {
		return result, err
	}
	return result, nil
}

func validateGGRTransactionFieldLengths(req validatedGGRTransaction) error {
	fields := []struct {
		name  string
		value string
		max   int
	}{
		{name: "agent_code", value: req.AgentCode, max: 128},
		{name: "user_code", value: req.UserCode, max: 255},
		{name: "game_type", value: req.GameType, max: 16},
		{name: "provider_code", value: req.ProviderCode, max: 64},
		{name: "game_code", value: req.GameCode, max: 255},
		{name: "type", value: req.BetType, max: 128},
		{name: "txn_id", value: req.TxnID, max: 255},
		{name: "txn_type", value: req.TxnType, max: 32},
	}
	for _, field := range fields {
		if len(field.value) > field.max {
			return fmt.Errorf("ggr %s exceeds %d bytes", field.name, field.max)
		}
	}
	return nil
}

func ggrTransactionGameForType(req ggrWalletRequest) (*ggrTransactionGame, error) {
	objects := 0
	for _, item := range []*ggrTransactionGame{req.Slot, req.Live, req.Sportsbook, req.Mini} {
		if item != nil {
			objects++
		}
	}
	if objects != 1 {
		return nil, errors.New("ggr transaction must contain exactly one game object")
	}
	switch strings.TrimSpace(req.GameType) {
	case "slot":
		if req.Slot == nil {
			return nil, errors.New("ggr slot object is required")
		}
		return req.Slot, nil
	case "live":
		if req.Live == nil {
			return nil, errors.New("ggr live object is required")
		}
		return req.Live, nil
	case "SB":
		if req.Sportsbook == nil {
			return nil, errors.New("ggr SB object is required")
		}
		return req.Sportsbook, nil
	case "MN":
		if req.Mini == nil {
			return nil, errors.New("ggr MN object is required")
		}
		return req.Mini, nil
	default:
		return nil, fmt.Errorf("ggr game_type is invalid: %s", strings.TrimSpace(req.GameType))
	}
}

func ggrCategoryMatchesGameType(categoryCode string, gameType string) bool {
	switch categoryCode {
	case "slots":
		return gameType == "slot"
	case "casino":
		return gameType == "live"
	case "sports":
		return gameType == "SB"
	case "mini":
		return gameType == "MN"
	default:
		return false
	}
}

func parseRequiredGGRMoney(number *json.Number, field string) (float64, error) {
	if number == nil {
		return 0, fmt.Errorf("ggr %s is required", field)
	}
	value, err := number.Float64()
	if err != nil || math.IsNaN(value) || math.IsInf(value, 0) {
		return 0, fmt.Errorf("ggr %s is invalid", field)
	}
	value = utils.Truncate2(value)
	if value == 0 {
		value = 0
	}
	return value, nil
}

func parseOptionalGGRMoney(number *json.Number, field string) (*float64, error) {
	if number == nil {
		return nil, nil
	}
	value, err := parseRequiredGGRMoney(number, field)
	if err != nil {
		return nil, err
	}
	return &value, nil
}

func parseGGRRoundID(raw json.RawMessage) (string, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return "", nil
	}
	text := strings.TrimSpace(string(raw))
	if text == "" || strings.HasPrefix(text, `"`) {
		return "", errors.New("ggr round_id must be an integer")
	}
	value := new(big.Int)
	if _, ok := value.SetString(text, 10); !ok {
		return "", errors.New("ggr round_id must be an integer")
	}
	return value.String(), nil
}

func fingerprintGGRTransaction(req validatedGGRTransaction) (string, error) {
	canonical := ggrTransactionFingerprint{
		AgentCode: req.AgentCode,
		UserCode:  req.UserCode,
		// Preserve the previous canonical fingerprint for in-flight retries.
		// Before removal, user_token was required to equal user_code.
		UserToken:    req.UserCode,
		GameType:     req.GameType,
		Info:         req.Info,
		ProviderCode: req.ProviderCode,
		GameCode:     req.GameCode,
		BetType:      req.BetType,
		BetMoney:     formatGGRMoney(req.BetMoney),
		WinMoney:     formatGGRMoney(req.WinMoney),
		RoundID:      req.RoundID,
		TxnID:        req.TxnID,
		TxnType:      req.TxnType,
	}
	if req.AgentBalance != nil {
		value := formatGGRMoney(*req.AgentBalance)
		canonical.AgentBalance = &value
	}
	if req.RequestUserBalance != nil {
		value := formatGGRMoney(*req.RequestUserBalance)
		canonical.RequestUserBalance = &value
	}
	body, err := json.Marshal(canonical)
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256(body)
	return hex.EncodeToString(digest[:]), nil
}

func formatGGRMoney(value float64) string {
	return strconv.FormatFloat(utils.Truncate2(value), 'f', 2, 64)
}

func getGGRAppGame(db *gorm.DB, providerCode string, gameCode string) (pojo.AppGame, error) {
	var games []pojo.AppGame
	err := db.Model(&pojo.AppGame{}).
		Where("LOWER(platform_code) = ? AND UPPER(third_game_category) = ? AND BINARY third_game_id = BINARY ?", "ggr", providerCode, gameCode).
		Limit(2).
		Find(&games).Error
	if err != nil {
		return pojo.AppGame{}, err
	}
	if len(games) != 1 {
		return pojo.AppGame{}, fmt.Errorf("ggr local game count=%d", len(games))
	}
	return games[0], nil
}

func applyGGRTransaction(db *gorm.DB, req validatedGGRTransaction, appGame pojo.AppGame) (ggrTransactionOutcome, error) {
	outcome := ggrTransactionOutcome{Status: pojo.GGRTransactionResultFailed, Message: ggrMessageInternalError}
	err := db.Transaction(func(tx *gorm.DB) error {
		var user pojo.TgUser
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("uid = ? AND status <> ?", req.UserCode, int8(-1)).
			First(&user).Error; err != nil {
			return fmt.Errorf("stage=lock_user user_code=%q: %w", req.UserCode, err)
		}
		outcome.UserID = user.ID
		outcome.TenantID = user.TenantId
		outcome.StartBalance = utils.Truncate2(user.Balance)
		outcome.EndBalance = outcome.StartBalance

		var existing pojo.GGRTransaction
		err := tx.Where("txn_id = ? AND txn_type = ?", req.TxnID, req.TxnType).First(&existing).Error
		if err == nil {
			outcome.Idempotent = true
			if existing.RequestFingerprint != req.Fingerprint || existing.UserID != user.ID {
				outcome.Decision = "idempotency_conflict"
				outcome.Status = pojo.GGRTransactionResultFailed
				outcome.Message = ggrMessageInternalError
				return nil
			}
			outcome.Decision = "idempotent_replay"
			outcome.Status = existing.ResultStatus
			outcome.Message = existing.ResultMessage
			outcome.Balance = utils.Truncate2(user.Balance)
			outcome.StartBalance = existing.StartBalance
			outcome.EndBalance = outcome.Balance
			return nil
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("stage=load_idempotency txn_id=%q txn_type=%q: %w", req.TxnID, req.TxnType, err)
		}
		if user.Status != 1 {
			return fmt.Errorf("stage=validate_user_status user_id=%d user_status=%d: ggr player is disabled", user.ID, user.Status)
		}

		startBalance := utils.Truncate2(user.Balance)
		endBalance := calculateGGREndBalance(startBalance, req.BetMoney, req.WinMoney)
		outcome.StartBalance = startBalance
		outcome.EndBalance = endBalance
		if endBalance < 0 {
			failed := buildGGRTransactionRecord(req, user, startBalance, startBalance, pojo.GGRTransactionResultFailed, ggrMessageInsufficientFund)
			if err := tx.Create(&failed).Error; err != nil {
				return fmt.Errorf("stage=create_insufficient_transaction txn_id=%q: %w", req.TxnID, err)
			}
			outcome.Decision = "insufficient_funds"
			outcome.Status = pojo.GGRTransactionResultFailed
			outcome.Message = ggrMessageInsufficientFund
			outcome.Balance = startBalance
			outcome.EndBalance = startBalance
			return nil
		}

		balanceUpdate := tx.Model(&pojo.TgUser{}).Where("id = ?", user.ID).Update("balance", endBalance)
		if balanceUpdate.Error != nil {
			return fmt.Errorf("stage=update_balance user_id=%d start_balance=%.2f end_balance=%.2f: %w", user.ID, startBalance, endBalance, balanceUpdate.Error)
		}
		if balanceUpdate.RowsAffected != 1 {
			return fmt.Errorf("stage=update_balance user_id=%d rows_affected=%d", user.ID, balanceUpdate.RowsAffected)
		}
		record := buildGGRTransactionRecord(req, user, startBalance, endBalance, pojo.GGRTransactionResultSuccess, "")
		if err := tx.Create(&record).Error; err != nil {
			return fmt.Errorf("stage=create_transaction txn_id=%q: %w", req.TxnID, err)
		}
		if err := createGGRCashHistory(tx, user.ID, req, startBalance, endBalance); err != nil {
			return fmt.Errorf("stage=create_cash_history txn_id=%q: %w", req.TxnID, err)
		}
		if err := createGGRBetRecord(tx, user, req, appGame); err != nil {
			return fmt.Errorf("stage=create_bet_record txn_id=%q user_id=%d: %w", req.TxnID, user.ID, err)
		}

		gameInfo := gameCashTransferGameInfo{
			ThirdGameID:  req.GameCode,
			PlatformCode: "ggr",
			GameName:     appGameStringValue(appGame.GameName),
			IsLiveCasino: strings.EqualFold(strings.TrimSpace(appGameStringValue(appGame.CategoryCode)), "casino"),
		}
		betFlow := gameWithdrawFlowAmount(req.BetMoney, gameInfo)
		if betFlow > 0 {
			key := ggrTransactionStorageKey(req.TxnID, req.TxnType)
			if err := repository.RecordWithdrawFlowEvent(
				tx,
				user.ID,
				user.TenantId,
				pojo.WithdrawFlowEventTypeGameBet,
				"ggr_game_bet:"+key,
				0,
				key,
				betFlow,
				req.ReceivedAt,
			); err != nil {
				return fmt.Errorf("stage=create_withdraw_flow txn_id=%q user_id=%d: %w", req.TxnID, user.ID, err)
			}
			outcome.AppliedBet = true
			outcome.BetFlow = betFlow
		}
		outcome.Decision = "applied"
		outcome.Status = pojo.GGRTransactionResultSuccess
		outcome.Message = ""
		outcome.Balance = endBalance
		return nil
	})
	return outcome, err
}

func calculateGGREndBalance(startBalance float64, betMoney float64, winMoney float64) float64 {
	return utils.Truncate2(utils.Truncate2(startBalance) - utils.Truncate2(betMoney) + utils.Truncate2(winMoney))
}

func ggrBalanceDelta(betMoney float64, winMoney float64) float64 {
	return utils.Truncate2(utils.Truncate2(winMoney) - utils.Truncate2(betMoney))
}

func buildGGRTransactionRecord(req validatedGGRTransaction, user pojo.TgUser, startBalance float64, endBalance float64, status int, message string) pojo.GGRTransaction {
	return pojo.GGRTransaction{
		TxnID:     req.TxnID,
		UserID:    user.ID,
		UID:       strings.TrimSpace(user.Uid),
		AgentCode: req.AgentCode,
		UserCode:  req.UserCode,
		// Keep the legacy non-null column compatible using the canonical UID.
		UserToken:          req.UserCode,
		GameType:           req.GameType,
		ProviderCode:       req.ProviderCode,
		GameCode:           req.GameCode,
		BetType:            req.BetType,
		BetMoney:           req.BetMoney,
		WinMoney:           req.WinMoney,
		RoundID:            req.RoundID,
		TxnType:            req.TxnType,
		Info:               req.Info,
		AgentBalance:       req.AgentBalance,
		RequestUserBalance: req.RequestUserBalance,
		StartBalance:       startBalance,
		EndBalance:         endBalance,
		RequestFingerprint: req.Fingerprint,
		ResultStatus:       status,
		ResultMessage:      message,
		ReceivedAt:         req.ReceivedAt,
	}
}

func createGGRCashHistory(tx *gorm.DB, userID int64, req validatedGGRTransaction, startBalance float64, endBalance float64) error {
	netAmount := utils.Truncate2(req.WinMoney - req.BetMoney)
	cashType := pojo.CashHistoryTypeGameWin
	cashMark := "ggr_game_win"
	if netAmount < 0 || (netAmount == 0 && req.BetMoney > 0) {
		cashType = pojo.CashHistoryTypeGameBet
		cashMark = "ggr_game_bet"
	}
	if req.TxnType == "debit_credit" {
		cashMark = "ggr_game_settlement"
	}
	return tx.Create(&pojo.CashHistory{
		UserId:      userID,
		AwardUni:    "ggr_txn:" + ggrTransactionStorageKey(req.TxnID, req.TxnType),
		Amount:      netAmount,
		StartAmount: startBalance,
		EndAmount:   endBalance,
		CashMark:    cashMark,
		CashDesc:    fmt.Sprintf("GGR %s provider=%s txn=%s", req.TxnType, req.ProviderCode, ggrTransactionStorageKey(req.TxnID, req.TxnType)),
		Type:        cashType,
	}).Error
}

func createGGRBetRecord(tx *gorm.DB, user pojo.TgUser, req validatedGGRTransaction, appGame pojo.AppGame) error {
	uid := parseGameUserNumericUID(user.Uid)
	userID := user.ID
	gameID := req.GameCode
	gameName := appGameStringValue(appGame.GameName)
	platformCode := "ggr"
	betAmount := req.BetMoney
	winAmount := req.WinMoney
	roundID := req.RoundID
	txnID := req.TxnID
	roundEnd := 0
	date := req.ReceivedAt
	remark := fmt.Sprintf("GGR txn_type=%s provider=%s type=%s", req.TxnType, req.ProviderCode, req.BetType)
	now := req.ReceivedAt
	return tx.Table(pojo.AppUserBetRecordTableNameByUserID(user.ID)).Create(&pojo.AppUserBetRecord{
		UID:          uid,
		UserID:       &userID,
		GameID:       &gameID,
		GameName:     &gameName,
		PlatformCode: &platformCode,
		BetAmount:    &betAmount,
		WinAmount:    &winAmount,
		RoundID:      &roundID,
		TraceID:      &txnID,
		RoundEnd:     &roundEnd,
		Date:         &date,
		Remark:       &remark,
		CreateTime:   &now,
		UpdateTime:   &now,
	}).Error
}

func postProcessGGRBet(db *gorm.DB, outcome ggrTransactionOutcome) {
	throttleKey := fmt.Sprintf("game_bet_post_throttle:%d:%d", outcome.TenantID, outcome.UserID)
	acquired, err := utils.AcquireLock(throttleKey, vipUpgradeCheckThrottle)
	if err != nil {
		log.Printf("[ggr] post-process throttle failed user_id=%d err=%v", outcome.UserID, err)
		return
	}
	if !acquired {
		return
	}
	userID := outcome.UserID
	go func() {
		if err := db.Transaction(func(tx *gorm.DB) error {
			_, err := repository.EnsureInviteValidUser(tx, userID, time.Now())
			return err
		}); err != nil {
			log.Printf("[ggr] invite valid post-process failed user_id=%d err=%v", userID, err)
		}
		repository.CheckAndUpgradeVipLevel(db, userID)
	}()
}

func ggrWalletFailure(ctx *gin.Context, includeBalance bool, message string) {
	response := ggrWalletResponse{Status: 0, Message: message}
	if includeBalance {
		balance := 0.0
		response.UserBalance = &balance
	}
	ctx.JSON(http.StatusOK, response)
}

func secureStringEqual(value string, expected string) bool {
	return subtle.ConstantTimeCompare([]byte(value), []byte(expected)) == 1
}

func ggrTransactionStorageKey(txnID string, txnType string) string {
	digest := sha256.Sum256([]byte(strings.TrimSpace(txnID) + "\x00" + strings.TrimSpace(txnType)))
	return hex.EncodeToString(digest[:])
}

func ggrWalletLogContext(ctx *gin.Context) string {
	host := utils.GetRequestHost(ctx)
	prefix := ""
	if value, exists := ctx.Get("hostInfo"); exists {
		if hostInfo, ok := value.(pojo.HostInfo); ok {
			prefix = strings.TrimSpace(hostInfo.TablePrefix)
		}
	}
	return fmt.Sprintf("host=%q prefix=%q client_ip=%q", host, prefix, utils.GetIPAddress(ctx))
}

func ggrWalletRequestSummary(req ggrWalletRequest) string {
	objectKey, transactionGame := ggrRequestGameForLog(req)
	providerCode := ""
	gameCode := ""
	betType := ""
	txnID := ""
	txnType := ""
	betMoney := "<missing>"
	winMoney := "<missing>"
	roundID := ""
	if transactionGame != nil {
		providerCode = strings.TrimSpace(transactionGame.ProviderCode)
		gameCode = strings.TrimSpace(transactionGame.GameCode)
		betType = strings.TrimSpace(transactionGame.BetType)
		txnID = strings.TrimSpace(transactionGame.TxnID)
		txnType = strings.TrimSpace(transactionGame.TxnType)
		betMoney = ggrJSONNumberForLog(transactionGame.BetMoney)
		winMoney = ggrJSONNumberForLog(transactionGame.WinMoney)
		roundID = strings.TrimSpace(string(transactionGame.RoundID))
	}
	return fmt.Sprintf(
		"method=%q agent_code=%q user_code=%q game_type=%q game_object=%q provider=%q game=%q bet_type=%q txn_id=%q txn_type=%q bet_money=%s win_money=%s round_id=%q agent_balance=%s request_user_balance=%s info_bytes=%d",
		strings.TrimSpace(req.Method),
		strings.TrimSpace(req.AgentCode),
		strings.TrimSpace(req.UserCode),
		strings.TrimSpace(req.GameType),
		objectKey,
		providerCode,
		gameCode,
		betType,
		txnID,
		txnType,
		betMoney,
		winMoney,
		roundID,
		ggrJSONNumberForLog(req.AgentBalance),
		ggrJSONNumberForLog(req.UserBalance),
		len(req.Info),
	)
}

func ggrRequestGameForLog(req ggrWalletRequest) (string, *ggrTransactionGame) {
	objects := []struct {
		key  string
		game *ggrTransactionGame
	}{
		{key: "slot", game: req.Slot},
		{key: "live", game: req.Live},
		{key: "SB", game: req.Sportsbook},
		{key: "MN", game: req.Mini},
	}
	for _, object := range objects {
		if object.game != nil {
			return object.key, object.game
		}
	}
	return "", nil
}

func ggrJSONNumberForLog(number *json.Number) string {
	if number == nil {
		return "<missing>"
	}
	return string(*number)
}

func redactGGRWalletRequestBody(body []byte) (string, error) {
	var object map[string]json.RawMessage
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.UseNumber()
	if err := decoder.Decode(&object); err != nil {
		return "", err
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			return "", errors.New("ggr request contains multiple json values")
		}
		return "", err
	}
	if _, exists := object["agent_secret"]; exists {
		object["agent_secret"] = json.RawMessage(`"<redacted>"`)
	}
	redacted, err := json.Marshal(object)
	if err != nil {
		return "", err
	}
	return string(redacted), nil
}

func ggrUnreadableRequestBodyForLog(body []byte) string {
	digest := sha256.Sum256(body)
	return fmt.Sprintf("<unavailable bytes=%d sha256=%s>", len(body), hex.EncodeToString(digest[:]))
}
