package api

import (
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
	Status     int
	Message    string
	Balance    float64
	AppliedBet bool
	BetFlow    float64
	UserID     int64
	TenantID   int64
	Idempotent bool
}

// GGRGoldAPI handles both GGR balance queries and seamless transactions.
// GGR requires its response object directly rather than the application's
// standard response envelope.
func GGRGoldAPI(ctx *gin.Context) {
	req, err := decodeGGRWalletRequest(ctx)
	if err != nil {
		ggrWalletFailure(ctx, false, ggrMessageInternalError)
		return
	}
	cfg := game.GetGGRConfig()
	if cfg.AgentCode == "" || cfg.AgentSecret == "" {
		log.Printf("[ggr] callback configuration is incomplete method=%s", strings.TrimSpace(req.Method))
		ggrWalletFailure(ctx, strings.TrimSpace(req.Method) == ggrMethodUserBalance, ggrMessageInternalError)
		return
	}
	if !secureStringEqual(strings.TrimSpace(req.AgentCode), cfg.AgentCode) || !secureStringEqual(req.AgentSecret, cfg.AgentSecret) {
		log.Printf("[ggr] callback authentication failed method=%s agent_code=%s", strings.TrimSpace(req.Method), strings.TrimSpace(req.AgentCode))
		ggrWalletFailure(ctx, strings.TrimSpace(req.Method) == ggrMethodUserBalance, ggrMessageInternalError)
		return
	}

	db := ctx.MustGet("db").(*gorm.DB)
	switch strings.TrimSpace(req.Method) {
	case ggrMethodUserBalance:
		handleGGRUserBalance(ctx, db, req)
	case ggrMethodTransaction:
		handleGGRTransaction(ctx, db, cfg, req)
	default:
		ggrWalletFailure(ctx, false, ggrMessageInternalError)
	}
}

func decodeGGRWalletRequest(ctx *gin.Context) (ggrWalletRequest, error) {
	var req ggrWalletRequest
	decoder := json.NewDecoder(ctx.Request.Body)
	decoder.UseNumber()
	if err := decoder.Decode(&req); err != nil {
		return req, err
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			return req, errors.New("ggr request contains multiple json values")
		}
		return req, err
	}
	return req, nil
}

func handleGGRUserBalance(ctx *gin.Context, db *gorm.DB, req ggrWalletRequest) {
	userCode, err := validateGGRBalanceUserCode(req)
	if err != nil {
		ggrWalletFailure(ctx, true, ggrMessageInternalError)
		return
	}

	var user pojo.TgUser
	if err = db.Select("id, uid, balance, status").
		Where("uid = ? AND status <> ?", userCode, int8(-1)).
		First(&user).Error; err != nil || user.Status != 1 {
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

func handleGGRTransaction(ctx *gin.Context, db *gorm.DB, cfg game.GGRConfig, req ggrWalletRequest) {
	validated, err := validateGGRTransaction(req, cfg)
	if err != nil {
		log.Printf("[ggr] transaction validation failed txn_id=%s err=%v", ggrRequestTxnID(req), err)
		ggrWalletFailure(ctx, false, ggrMessageInternalError)
		return
	}

	appGame, err := getGGRAppGame(db, validated.ProviderCode, validated.GameCode)
	if err != nil {
		log.Printf("[ggr] transaction game not found provider=%s game=%s txn_id=%s err=%v", validated.ProviderCode, validated.GameCode, validated.TxnID, err)
		ggrWalletFailure(ctx, false, ggrMessageInternalError)
		return
	}

	outcome, err := applyGGRTransaction(db, validated, appGame)
	if err != nil {
		log.Printf("[ggr] transaction failed txn_id=%s err=%v", validated.TxnID, err)
		ggrWalletFailure(ctx, false, ggrMessageInternalError)
		return
	}
	if outcome.Status != pojo.GGRTransactionResultSuccess {
		ggrWalletFailure(ctx, false, outcome.Message)
		return
	}
	if outcome.AppliedBet && outcome.BetFlow > 0 && outcome.UserID > 0 {
		postProcessGGRBet(db, outcome)
	}
	balance := utils.Truncate2(outcome.Balance)
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
			return err
		}
		outcome.UserID = user.ID
		outcome.TenantID = user.TenantId

		var existing pojo.GGRTransaction
		err := tx.Where("txn_id = ?", req.TxnID).First(&existing).Error
		if err == nil {
			outcome.Idempotent = true
			if existing.RequestFingerprint != req.Fingerprint || existing.UserID != user.ID {
				outcome.Status = pojo.GGRTransactionResultFailed
				outcome.Message = ggrMessageInternalError
				return nil
			}
			outcome.Status = existing.ResultStatus
			outcome.Message = existing.ResultMessage
			outcome.Balance = utils.Truncate2(user.Balance)
			return nil
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if user.Status != 1 {
			return errors.New("ggr player is disabled")
		}

		startBalance := utils.Truncate2(user.Balance)
		endBalance := calculateGGREndBalance(startBalance, req.BetMoney, req.WinMoney)
		if endBalance < 0 {
			failed := buildGGRTransactionRecord(req, user, startBalance, startBalance, pojo.GGRTransactionResultFailed, ggrMessageInsufficientFund)
			if err := tx.Create(&failed).Error; err != nil {
				return err
			}
			outcome.Status = pojo.GGRTransactionResultFailed
			outcome.Message = ggrMessageInsufficientFund
			outcome.Balance = startBalance
			return nil
		}

		if err := tx.Model(&pojo.TgUser{}).Where("id = ?", user.ID).Update("balance", endBalance).Error; err != nil {
			return err
		}
		record := buildGGRTransactionRecord(req, user, startBalance, endBalance, pojo.GGRTransactionResultSuccess, "")
		if err := tx.Create(&record).Error; err != nil {
			return err
		}
		if err := createGGRCashHistory(tx, user.ID, req, startBalance, endBalance); err != nil {
			return err
		}
		if err := createGGRBetRecord(tx, user, req, appGame); err != nil {
			return err
		}

		gameInfo := gameCashTransferGameInfo{
			ThirdGameID:  req.GameCode,
			PlatformCode: "ggr",
			GameName:     appGameStringValue(appGame.GameName),
			IsLiveCasino: strings.EqualFold(strings.TrimSpace(appGameStringValue(appGame.CategoryCode)), "casino"),
		}
		betFlow := gameWithdrawFlowAmount(req.BetMoney, gameInfo)
		if betFlow > 0 {
			key := ggrTransactionStorageKey(req.TxnID)
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
				return err
			}
			outcome.AppliedBet = true
			outcome.BetFlow = betFlow
		}
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
		AwardUni:    "ggr_txn:" + ggrTransactionStorageKey(req.TxnID),
		Amount:      netAmount,
		StartAmount: startBalance,
		EndAmount:   endBalance,
		CashMark:    cashMark,
		CashDesc:    fmt.Sprintf("GGR %s provider=%s txn=%s", req.TxnType, req.ProviderCode, ggrTransactionStorageKey(req.TxnID)),
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

func ggrTransactionStorageKey(txnID string) string {
	digest := sha256.Sum256([]byte(strings.TrimSpace(txnID)))
	return hex.EncodeToString(digest[:])
}

func ggrRequestTxnID(req ggrWalletRequest) string {
	for _, item := range []*ggrTransactionGame{req.Slot, req.Live, req.Sportsbook, req.Mini} {
		if item != nil {
			return strings.TrimSpace(item.TxnID)
		}
	}
	return ""
}
