package api

import (
	"BaseGoUni/core/game"
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/repository"
	"BaseGoUni/core/utils"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// GetGameCash godoc
//
//	@Summary		获取玩家余额
//	@Tags			游戏钱包
//	@Accept			json
//	@Produce		json
//	@Param			X-Sign header string true "签名"
//	@Param			X-Request-Id header string true "请求ID"
//	@Param			X-Appid header string true "商户号"
//	@Param			data body pojo.GameCashGetReq true "玩家ID"
//	@Success		200	{object}	game.APIResponse[pojo.GameCashGetData]
//	@Router			/Cash/Get [post]
func GetGameCash(ctx *gin.Context) {
	body, err := ctx.GetRawData()
	if err != nil {
		log.Printf("[game_wallet] Cash/Get read body failed ip=%s err=%v", utils.GetIPAddress(ctx), err)
		gameErrorBack(ctx, game.GameCodeInvalidMerchantCode, "")
		return
	}
	if !verifyGameRequest(ctx, body) {
		return
	}

	var req pojo.GameCashGetReq
	if err := json.Unmarshal(body, &req); err != nil {
		log.Printf("[game_wallet] Cash/Get json invalid requestId=%s appId=%s ip=%s err=%v body=%s",
			ctx.GetHeader(game.HeaderRequestID), ctx.GetHeader(game.HeaderAppID), utils.GetIPAddress(ctx), err, string(body))
		gameErrorBack(ctx, game.GameCodeInvalidMerchantCode, "")
		return
	}
	userID := strings.TrimSpace(req.UserID)
	log.Printf("[game_wallet] Cash/Get begin requestId=%s appId=%s userid=%s ip=%s",
		ctx.GetHeader(game.HeaderRequestID), ctx.GetHeader(game.HeaderAppID), userID, utils.GetIPAddress(ctx))
	if userID == "" {
		log.Printf("[game_wallet] Cash/Get empty userid requestId=%s appId=%s ip=%s",
			ctx.GetHeader(game.HeaderRequestID), ctx.GetHeader(game.HeaderAppID), utils.GetIPAddress(ctx))
		gameErrorBack(ctx, game.GameCodeEmptyUserID, "")
		return
	}

	db := ctx.MustGet("db").(*gorm.DB)
	var user pojo.TgUser
	err = db.Model(&pojo.TgUser{}).
		Select("id, uid, balance, status").
		Where("uid = ? AND status <> ?", userID, int8(-1)).
		First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("[game_wallet] Cash/Get user not found requestId=%s userid=%s",
				ctx.GetHeader(game.HeaderRequestID), userID)
			gameErrorBack(ctx, game.GameCodePlayerNotFound, "")
			return
		}
		log.Printf("[game_wallet] Cash/Get db error requestId=%s userid=%s err=%v",
			ctx.GetHeader(game.HeaderRequestID), userID, err)
		gameErrorBack(ctx, game.GameCodeInvalidMerchantCode, "")
		return
	}
	if user.Status != 1 {
		log.Printf("[game_wallet] Cash/Get user disabled requestId=%s userid=%s status=%d",
			ctx.GetHeader(game.HeaderRequestID), userID, user.Status)
		gameErrorBack(ctx, game.GameCodePlayerDisabled, "")
		return
	}

	log.Printf("[game_wallet] Cash/Get success requestId=%s userid=%s userId=%d balance=%.2f",
		ctx.GetHeader(game.HeaderRequestID), userID, user.ID, utils.Truncate2(user.Balance))
	gameSuccessBack(ctx, pojo.GameCashGetData{
		Balance: utils.Truncate2(user.Balance),
	})
}

// TransferGameCashInOut godoc
//
//	@Summary		修改玩家余额
//	@Tags			游戏钱包
//	@Accept			json
//	@Produce		json
//	@Param			X-Sign header string true "签名"
//	@Param			X-Request-Id header string true "请求ID"
//	@Param			X-Appid header string true "商户号"
//	@Param			data body pojo.GameCashTransferInOutReq true "余额变更"
//	@Success		200	{object}	game.APIResponse[pojo.GameCashGetData]
//	@Router			/Cash/TransferInOut [post]
func TransferGameCashInOut(ctx *gin.Context) {
	body, err := ctx.GetRawData()
	if err != nil {
		log.Printf("[game_wallet] TransferInOut read body failed ip=%s err=%v", utils.GetIPAddress(ctx), err)
		gameErrorBack(ctx, game.GameCodeInvalidMerchantCode, "")
		return
	}
	if !verifyGameRequest(ctx, body) {
		return
	}

	var req pojo.GameCashTransferInOutReq
	if err := json.Unmarshal(body, &req); err != nil {
		log.Printf("[game_wallet] TransferInOut json invalid requestId=%s appId=%s ip=%s err=%v body=%s",
			ctx.GetHeader(game.HeaderRequestID), ctx.GetHeader(game.HeaderAppID), utils.GetIPAddress(ctx), err, string(body))
		gameErrorBack(ctx, game.GameCodeInvalidMerchantCode, "")
		return
	}
	log.Printf("[game_wallet] TransferInOut begin requestId=%s appId=%s userid=%s tid=%s roundid=%s gameid=%s reason=%s amount=%.2f bet=%.2f isEnd=%t isBuy=%t ip=%s",
		ctx.GetHeader(game.HeaderRequestID), ctx.GetHeader(game.HeaderAppID), strings.TrimSpace(req.UserID), strings.TrimSpace(req.TID),
		strings.TrimSpace(req.RoundID), strings.TrimSpace(req.GameID), strings.TrimSpace(req.Reason), req.Amount, req.Bet, req.IsEnd, req.IsBuy, utils.GetIPAddress(ctx))
	if code, msg := validateGameCashTransferReq(req); code != game.GameCodeSuccess {
		log.Printf("[game_wallet] TransferInOut validate failed requestId=%s userid=%s tid=%s code=%d msg=%s",
			ctx.GetHeader(game.HeaderRequestID), strings.TrimSpace(req.UserID), strings.TrimSpace(req.TID), code, msg)
		gameErrorBack(ctx, code, msg)
		return
	}

	lockKey := fmt.Sprintf("bgu_game_cash_transfer_%s", utils.MD5(strings.TrimSpace(req.UserID)))
	acquired, lockErr := utils.AcquireLock(lockKey, 20*time.Second)
	if lockErr != nil {
		log.Printf("[game_wallet] TransferInOut lock error requestId=%s userid=%s tid=%s err=%v",
			ctx.GetHeader(game.HeaderRequestID), strings.TrimSpace(req.UserID), strings.TrimSpace(req.TID), lockErr)
		gameErrorBack(ctx, game.GameCodeTooFrequent, "request too frequent")
		return
	}
	if !acquired {
		log.Printf("[game_wallet] TransferInOut lock busy requestId=%s userid=%s tid=%s",
			ctx.GetHeader(game.HeaderRequestID), strings.TrimSpace(req.UserID), strings.TrimSpace(req.TID))
		gameErrorBack(ctx, game.GameCodeTooFrequent, game.ErrorMessage(game.GameCodeTooFrequent))
		return
	}
	defer utils.ReleaseLock(lockKey)

	db := ctx.MustGet("db").(*gorm.DB)
	balance, err := handleGameCashTransfer(db, req)
	if err != nil {
		log.Printf("[game_wallet] TransferInOut failed requestId=%s userid=%s tid=%s err=%v",
			ctx.GetHeader(game.HeaderRequestID), strings.TrimSpace(req.UserID), strings.TrimSpace(req.TID), err)
		writeGameCashTransferError(ctx, err)
		return
	}
	log.Printf("[game_wallet] TransferInOut success requestId=%s userid=%s tid=%s balance=%.2f",
		ctx.GetHeader(game.HeaderRequestID), strings.TrimSpace(req.UserID), strings.TrimSpace(req.TID), balance)
	gameSuccessBack(ctx, pojo.GameCashGetData{Balance: balance})
}

func verifyGameRequest(ctx *gin.Context, body []byte) bool {
	cfg := game.GetConfig()
	appID := strings.TrimSpace(ctx.GetHeader(game.HeaderAppID))
	requestID := strings.TrimSpace(ctx.GetHeader(game.HeaderRequestID))
	sign := strings.TrimSpace(ctx.GetHeader(game.HeaderSign))
	if cfg.AppID == "" || cfg.AppSecret == "" || appID == "" || appID != cfg.AppID {
		log.Printf("[game_wallet] verify failed invalid appid path=%s requestId=%s appId=%s configuredAppId=%s ip=%s",
			ctx.Request.URL.Path, requestID, appID, cfg.AppID, utils.GetIPAddress(ctx))
		gameErrorBack(ctx, game.GameCodeInvalidAppID, "")
		return false
	}
	if requestID == "" || sign == "" || !game.VerifySign(requestID, body, cfg.AppSecret, sign) {
		log.Printf("[game_wallet] verify failed invalid sign path=%s requestId=%s appId=%s hasSign=%t body=%s ip=%s",
			ctx.Request.URL.Path, requestID, appID, sign != "", string(body), utils.GetIPAddress(ctx))
		gameErrorBack(ctx, game.GameCodeInvalidMerchantCode, "")
		return false
	}
	return true
}

func validateGameCashTransferReq(req pojo.GameCashTransferInOutReq) (int, string) {
	if strings.TrimSpace(req.UserID) == "" {
		return game.GameCodeEmptyUserID, ""
	}
	if strings.TrimSpace(req.TID) == "" || strings.TrimSpace(req.RoundID) == "" || strings.TrimSpace(req.GameID) == "" {
		return game.GameCodeInvalidMerchantCode, ""
	}
	if req.ReqTime <= 0 {
		return game.GameCodeInvalidMerchantCode, ""
	}
	switch strings.ToLower(strings.TrimSpace(req.Reason)) {
	case "bet":
		if req.Amount > 0 {
			return game.GameCodeInvalidTransferAmount, game.ErrorMessage(game.GameCodeInvalidTransferAmount)
		}
		return game.GameCodeSuccess, ""
	case "win", "refund":
		if req.Amount < 0 {
			return game.GameCodeInvalidTransferAmount, game.ErrorMessage(game.GameCodeInvalidTransferAmount)
		}
		return game.GameCodeSuccess, ""
	default:
		return game.GameCodeInvalidTransferAmount, "invalid reason"
	}
}

type gameCashTransferError struct {
	code int
	msg  string
}

func (e gameCashTransferError) Error() string {
	return e.msg
}

func newGameCashTransferError(code int, msg string) error {
	if strings.TrimSpace(msg) == "" {
		msg = game.ErrorMessage(code)
	}
	return gameCashTransferError{code: code, msg: msg}
}

func writeGameCashTransferError(ctx *gin.Context, err error) {
	var transferErr gameCashTransferError
	if errors.As(err, &transferErr) {
		gameErrorBack(ctx, transferErr.code, transferErr.msg)
		return
	}
	gameErrorBack(ctx, game.GameCodeInvalidMerchantCode, err.Error())
}

func handleGameCashTransfer(db *gorm.DB, req pojo.GameCashTransferInOutReq) (float64, error) {
	userID := strings.TrimSpace(req.UserID)
	tid := strings.TrimSpace(req.TID)
	awardUni := gameCashTransferAwardUni(tid)
	var balance float64

	err := db.Transaction(func(tx *gorm.DB) error {
		var history pojo.CashHistory
		err := tx.Where("award_uni = ?", awardUni).First(&history).Error
		if err == nil {
			balance = utils.Truncate2(history.EndAmount)
			log.Printf("[game_wallet] TransferInOut idempotent hit userid=%s tid=%s balance=%.2f historyId=%d",
				userID, tid, balance, history.ID)
			return nil
		}
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		var user pojo.TgUser
		err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("uid = ? AND status <> ?", userID, int8(-1)).
			First(&user).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				log.Printf("[game_wallet] TransferInOut user not found userid=%s tid=%s", userID, tid)
				return newGameCashTransferError(game.GameCodePlayerNotFound, "")
			}
			return err
		}
		if user.Status != 1 {
			log.Printf("[game_wallet] TransferInOut user disabled userid=%s tid=%s userId=%d status=%d",
				userID, tid, user.ID, user.Status)
			return newGameCashTransferError(game.GameCodePlayerDisabled, "")
		}

		amount := utils.Truncate2(req.Amount)
		if amount < 0 && utils.Truncate2(user.Balance+amount) < 0 {
			log.Printf("[game_wallet] TransferInOut insufficient balance userid=%s tid=%s userId=%d balance=%.2f amount=%.2f",
				userID, tid, user.ID, utils.Truncate2(user.Balance), amount)
			return newGameCashTransferError(game.GameCodeInsufficientBalance, game.ErrorMessage(game.GameCodeInsufficientBalance))
		}

		startBalance := utils.Truncate2(user.Balance)
		endBalance := utils.Truncate2(startBalance + amount)
		if err := tx.Model(&pojo.TgUser{}).
			Where("id = ?", user.ID).
			Update("balance", endBalance).Error; err != nil {
			return err
		}

		if err := createGameBetRecord(tx, user, req, amount); err != nil {
			return err
		}
		betAmount, _ := gameBetRecordAmounts(req, amount)
		if betAmount > 0 {
			occurredAt := time.Now()
			if req.ReqTime > 0 {
				occurredAt = time.UnixMilli(req.ReqTime)
			}
			if err := repository.RecordWithdrawFlowEvent(
				tx,
				user.ID,
				user.TenantId,
				pojo.WithdrawFlowEventTypeGameBet,
				"game_bet:"+tid,
				0,
				tid,
				betAmount,
				occurredAt,
			); err != nil {
				return err
			}
			if err := repository.ApplyInviteBetRebate(tx, user, betAmount, tid, occurredAt); err != nil {
				return err
			}
		}
		if err := tx.Create(&pojo.CashHistory{
			UserId:      user.ID,
			AwardUni:    awardUni,
			Amount:      amount,
			StartAmount: startBalance,
			EndAmount:   endBalance,
			CashMark:    gameCashTransferMark(req.Reason),
			CashDesc:    gameCashTransferDesc(req),
			Type:        gameCashTransferCashHistoryType(req.Reason),
		}).Error; err != nil {
			return err
		}

		balance = endBalance
		log.Printf("[game_wallet] TransferInOut persisted userid=%s tid=%s userId=%d reason=%s amount=%.2f before=%.2f after=%.2f",
			userID, tid, user.ID, strings.TrimSpace(req.Reason), amount, startBalance, endBalance)
		return nil
	})
	if err != nil {
		return 0, err
	}
	return balance, nil
}

func createGameBetRecord(tx *gorm.DB, user pojo.TgUser, req pojo.GameCashTransferInOutReq, amount float64) error {
	now := time.Now()
	if req.ReqTime > 0 {
		now = time.UnixMilli(req.ReqTime)
	}
	gameID := strings.TrimSpace(req.GameID)
	roundID := strings.TrimSpace(req.RoundID)
	tid := strings.TrimSpace(req.TID)
	platformCode := "hg"
	gameName := lookupAppGameName(tx, gameID)
	roundEnd := 0
	if req.IsEnd {
		roundEnd = 1
	}
	betAmount, winAmount := gameBetRecordAmounts(req, amount)
	uid := parseGameUserNumericUID(user.Uid)
	remark := gameCashTransferDesc(req)

	return tx.Table(pojo.AppUserBetRecordTableNameByUserID(user.ID)).Create(&pojo.AppUserBetRecord{
		UID:          uid,
		UserID:       &user.ID,
		GameID:       &gameID,
		GameName:     &gameName,
		PlatformCode: &platformCode,
		BetAmount:    &betAmount,
		WinAmount:    &winAmount,
		RoundID:      &roundID,
		TraceID:      &tid,
		RoundEnd:     &roundEnd,
		Date:         &now,
		Remark:       &remark,
		CreateTime:   &now,
		UpdateTime:   &now,
	}).Error
}

func lookupAppGameName(tx *gorm.DB, thirdGameID string) string {
	var appGame pojo.AppGame
	err := tx.Model(&pojo.AppGame{}).
		Select("game_name").
		Where("third_game_id = ? AND COALESCE(deleted_flag, 0) = 0", strings.TrimSpace(thirdGameID)).
		First(&appGame).Error
	if err != nil || appGame.GameName == nil {
		return ""
	}
	return *appGame.GameName
}

func gameBetRecordAmounts(req pojo.GameCashTransferInOutReq, amount float64) (float64, float64) {
	betAmount := utils.Truncate2(req.Bet)
	winAmount := 0.0
	switch strings.ToLower(strings.TrimSpace(req.Reason)) {
	case "bet":
		if betAmount <= 0 {
			betAmount = utils.Truncate2(-amount)
		}
	case "win", "refund":
		if amount > 0 {
			winAmount = amount
		}
	}
	return betAmount, winAmount
}

func parseGameUserNumericUID(uid string) *int64 {
	value, err := strconv.ParseInt(strings.TrimSpace(uid), 10, 64)
	if err != nil {
		return nil
	}
	return &value
}

func gameCashTransferAwardUni(tid string) string {
	return "game_transfer_" + strings.TrimSpace(tid)
}

func gameCashTransferCashHistoryType(reason string) int8 {
	switch strings.ToLower(strings.TrimSpace(reason)) {
	case "bet":
		return pojo.CashHistoryTypeGameBet
	case "win":
		return pojo.CashHistoryTypeGameWin
	case "refund":
		return pojo.CashHistoryTypeGameRefund
	default:
		return pojo.CashHistoryTypeUnknown
	}
}

func gameCashTransferMark(reason string) string {
	switch strings.ToLower(strings.TrimSpace(reason)) {
	case "bet":
		return "游戏下注"
	case "win":
		return "游戏派奖"
	case "refund":
		return "游戏退款"
	default:
		return "游戏余额变动"
	}
}

func gameCashTransferDesc(req pojo.GameCashTransferInOutReq) string {
	return fmt.Sprintf("游戏%s tid=%s round=%s game=%s amount=%.2f isEnd=%t isBuy=%t",
		strings.TrimSpace(req.Reason),
		strings.TrimSpace(req.TID),
		strings.TrimSpace(req.RoundID),
		strings.TrimSpace(req.GameID),
		utils.Truncate2(req.Amount),
		req.IsEnd,
		req.IsBuy,
	)
}

func gameSuccessBack[T any](ctx *gin.Context, data T) {
	ctx.JSON(http.StatusOK, game.APIResponse[T]{
		Code:  game.GameCodeSuccess,
		Error: "",
		Data:  data,
	})
}

func gameErrorBack(ctx *gin.Context, code int, msg string) {
	if strings.TrimSpace(msg) == "" {
		msg = game.ErrorMessage(code)
	}
	ctx.JSON(http.StatusOK, gin.H{
		"code":  code,
		"error": msg,
	})
}
