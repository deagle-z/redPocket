package api

import (
	"BaseGoUni/core/game"
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/repository"
	"BaseGoUni/core/utils"
	"crypto/sha1"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	gscOpenPlatformCode = "gsc"
	gscOpenTraceMaxLen  = 96
)

// GSCOpenBalance godoc
//
//	@Summary		GSC 查询玩家余额
//	@Tags			GSC
//	@Accept			json
//	@Produce		json
//	@Param			data body pojo.GSCOpenBalanceReq true "余额查询"
//	@Success		200	{object}	pojo.GSCOpenBalanceResp
//	@Router			/v1/api/seamless/balance [post]
func GSCOpenBalance(ctx *gin.Context) {
	var req pojo.GSCOpenBalanceReq
	if !bindAndVerifyGSCOpenRequest(ctx, game.GSCOpenActionGetBalance, &req, func() (string, string, string) {
		return req.OperatorCode, req.RequestTime.String(), req.Sign
	}) {
		return
	}

	db := ctx.MustGet("db").(*gorm.DB)
	resp := pojo.GSCOpenBalanceResp{Data: make([]pojo.GSCOpenBalanceData, 0, len(req.BatchRequests))}
	for _, item := range req.BatchRequests {
		memberAccount := strings.TrimSpace(item.MemberAccount)
		row := pojo.GSCOpenBalanceData{
			MemberAccount: memberAccount,
			ProductCode:   item.ProductCode.Int(),
		}
		user, code, message, err := findGSCOpenUser(db, memberAccount)
		if err != nil {
			log.Printf("[gsc_open] balance user lookup failed member=%s err=%v", memberAccount, err)
		}
		if code != game.GameCodeSuccess {
			row.Code = code
			row.Message = gscOpenMessage(code, message)
			resp.Data = append(resp.Data, row)
			continue
		}
		row.Balance = utils.Truncate2(user.SportBalance)
		row.Code = game.GameCodeSuccess
		resp.Data = append(resp.Data, row)
	}
	ctx.JSON(http.StatusOK, resp)
}

// GSCOpenWithdraw godoc
//
//	@Summary		GSC 玩家下注扣款
//	@Tags			GSC
//	@Accept			json
//	@Produce		json
//	@Param			data body pojo.GSCOpenTransferReq true "扣款请求"
//	@Success		200	{object}	pojo.GSCOpenTransferResp
//	@Router			/v1/api/seamless/withdraw [post]
func GSCOpenWithdraw(ctx *gin.Context) {
	var req pojo.GSCOpenTransferReq
	if !bindAndVerifyGSCOpenRequest(ctx, game.GSCOpenActionWithdraw, &req, func() (string, string, string) {
		return req.OperatorCode, req.RequestTime.String(), req.Sign
	}) {
		return
	}
	handleGSCOpenTransfer(ctx, req, game.GSCOpenActionWithdraw)
}

// GSCOpenDeposit godoc
//
//	@Summary		GSC 玩家派奖加款
//	@Tags			GSC
//	@Accept			json
//	@Produce		json
//	@Param			data body pojo.GSCOpenTransferReq true "加款请求"
//	@Success		200	{object}	pojo.GSCOpenTransferResp
//	@Router			/v1/api/seamless/deposit [post]
func GSCOpenDeposit(ctx *gin.Context) {
	var req pojo.GSCOpenTransferReq
	if !bindAndVerifyGSCOpenRequest(ctx, game.GSCOpenActionDeposit, &req, func() (string, string, string) {
		return req.OperatorCode, req.RequestTime.String(), req.Sign
	}) {
		return
	}
	handleGSCOpenTransfer(ctx, req, game.GSCOpenActionDeposit)
}

// GSCOpenPushBetData godoc
//
//	@Summary		GSC 同步注单数据
//	@Tags			GSC
//	@Accept			json
//	@Produce		json
//	@Param			data body pojo.GSCOpenPushBetDataReq true "注单推送"
//	@Success		200	{object}	pojo.GSCOpenBaseResp
//	@Router			/v1/api/seamless/pushbetdata [post]
func GSCOpenPushBetData(ctx *gin.Context) {
	var req pojo.GSCOpenPushBetDataReq
	if !bindAndVerifyGSCOpenRequest(ctx, game.GSCOpenActionPushBetData, &req, func() (string, string, string) {
		return req.OperatorCode, req.RequestTime.String(), req.Sign
	}) {
		return
	}

	log.Printf("[gsc_open] pushbetdata received wagers=%d operator=%s", len(req.Wagers), strings.TrimSpace(req.OperatorCode))
	gscOpenBaseBack(ctx, game.GameCodeSuccess, "")
}

func handleGSCOpenTransfer(ctx *gin.Context, req pojo.GSCOpenTransferReq, action string) {
	db := ctx.MustGet("db").(*gorm.DB)
	resp := pojo.GSCOpenTransferResp{Data: make([]pojo.GSCOpenTransferData, 0, len(req.BatchRequests))}
	for _, batch := range req.BatchRequests {
		memberAccount := strings.TrimSpace(batch.MemberAccount)
		row := pojo.GSCOpenTransferData{
			MemberAccount: memberAccount,
			ProductCode:   batch.ProductCode.Int(),
		}

		user, code, message, err := findGSCOpenUser(db, memberAccount)
		if err != nil {
			log.Printf("[gsc_open] %s user lookup failed member=%s err=%v", action, memberAccount, err)
		}
		if code != game.GameCodeSuccess {
			row.Code = code
			row.Message = gscOpenMessage(code, message)
			resp.Data = append(resp.Data, row)
			continue
		}
		row.BeforeBalance = utils.Truncate2(user.SportBalance)
		row.Balance = row.BeforeBalance

		for _, tx := range batch.Transactions {
			transferReq, code, message := buildGSCOpenTransferReq(memberAccount, batch, tx, action)
			if code != game.GameCodeSuccess {
				row.Code = code
				row.Message = gscOpenMessage(code, message)
				break
			}
			balance, err := handleGSCSportCashTransfer(db, transferReq)
			if err != nil {
				code, message = gscOpenTransferError(err)
				row.Code = code
				row.Message = gscOpenMessage(code, message)
				log.Printf("[gsc_open] %s transfer failed member=%s trace=%s err=%v", action, memberAccount, transferReq.TID, err)
				break
			}
			row.Balance = balance
		}
		if row.Code == 0 && row.Message == "" {
			row.Code = game.GameCodeSuccess
		}
		resp.Data = append(resp.Data, row)
	}
	ctx.JSON(http.StatusOK, resp)
}

func buildGSCOpenTransferReq(memberAccount string, batch pojo.GSCOpenTransferBatchRequest, tx pojo.GSCOpenWalletTransaction, action string) (pojo.GameCashTransferInOutReq, int, string) {
	gameCode := strings.TrimSpace(tx.GameCode)
	if gameCode == "" {
		return pojo.GameCashTransferInOutReq{}, game.GameCodeGameNotFound, "game_code_required"
	}

	roundID := firstNonEmpty(tx.RoundID, tx.WagerCode, tx.ID)
	if roundID == "" {
		return pojo.GameCashTransferInOutReq{}, game.GameCodeInvalidMerchantCode, "round_id_required"
	}

	amount := 0.0
	reason := "bet"
	isEnd := false
	switch action {
	case game.GSCOpenActionWithdraw:
		amount = gscOpenPositiveAmount(tx.Amount.Float64(), tx.BetAmount.Float64(), tx.ValidBetAmount.Float64())
		amount = -amount
		reason = "bet"
	case game.GSCOpenActionDeposit:
		amount = gscOpenPositiveAmount(tx.Amount.Float64(), tx.PrizeAmount.Float64())
		reason = gscOpenDepositReason(tx)
		isEnd = gscOpenIsRoundEnd(tx)
	default:
		return pojo.GameCashTransferInOutReq{}, game.GameCodeInvalidMerchantCode, "invalid_transfer_action"
	}
	if amount == 0 {
		return pojo.GameCashTransferInOutReq{}, game.GameCodeInvalidTransferAmount, game.ErrorMessage(game.GameCodeInvalidTransferAmount)
	}

	return pojo.GameCashTransferInOutReq{
		UserID:       memberAccount,
		TID:          gscOpenTraceID(action, tx),
		Amount:       utils.Truncate2(amount),
		RoundID:      roundID,
		GameID:       gameCode,
		ReqTime:      gscOpenTimestampMillis(tx.SettledAt.Int64()),
		Reason:       reason,
		IsEnd:        isEnd,
		Bet:          gscOpenBetAmount(tx),
		PlatformCode: gscOpenPlatformCode,
	}, game.GameCodeSuccess, ""
}

func handleGSCSportCashTransfer(db *gorm.DB, req pojo.GameCashTransferInOutReq) (float64, error) {
	userID := strings.TrimSpace(req.UserID)
	tid := strings.TrimSpace(req.TID)
	awardUni := gameCashTransferAwardUni(tid)
	var balance float64
	var vipUserID int64
	var vipTenantID int64
	var vipBetFlow float64

	gameInfo := getGameCashTransferGameInfo(db, req.GameID, req.PlatformCode)
	roundAggregated := gameCashTransferIsRoundAggregated(gameInfo)

	tx := db.Begin()
	var err error
	if tx.Error != nil {
		err = tx.Error
	} else {
		err = func(tx *gorm.DB) error {
			var user pojo.TgUser
			err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
				Where("uid = ? AND status <> ?", userID, int8(-1)).
				First(&user).Error
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					log.Printf("[gsc_open] sport transfer user not found userid=%s tid=%s", userID, tid)
					return newGameCashTransferError(game.GameCodePlayerNotFound, "")
				}
				return err
			}
			if roundAggregated {
				exists, traceErr := gameCashTransferTraceExists(tx, user.ID, tid)
				if traceErr != nil {
					return traceErr
				}
				if exists {
					balance = utils.Truncate2(user.SportBalance)
					return nil
				}
			} else {
				var history pojo.CashHistory
				err = tx.Where("user_id = ? AND award_uni = ?", user.ID, awardUni).First(&history).Error
				if err == nil {
					balance = utils.Truncate2(user.SportBalance)
					return nil
				}
				if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
					return err
				}
			}
			if user.Status != 1 {
				log.Printf("[gsc_open] sport transfer user disabled userid=%s tid=%s userId=%d status=%d",
					userID, tid, user.ID, user.Status)
				return newGameCashTransferError(game.GameCodePlayerDisabled, "")
			}

			amount := utils.Truncate2(req.Amount)
			if code, msg := validateGameCashTransferAmount(req, amount, gameInfo.IsFishing); code != game.GameCodeSuccess {
				return newGameCashTransferError(code, msg)
			}
			if amount < 0 && utils.Truncate2(user.SportBalance+amount) < 0 {
				log.Printf("[gsc_open] sport transfer insufficient balance userid=%s tid=%s userId=%d sportBalance=%.2f amount=%.2f",
					userID, tid, user.ID, utils.Truncate2(user.SportBalance), amount)
				return newGameCashTransferError(game.GameCodeInsufficientBalance, game.ErrorMessage(game.GameCodeInsufficientBalance))
			}

			startBalance := utils.Truncate2(user.SportBalance)
			endBalance := utils.Truncate2(startBalance + amount)
			if err := tx.Model(&pojo.TgUser{}).
				Where("id = ?", user.ID).
				Update("sport_balance", endBalance).Error; err != nil {
				return err
			}

			if err := createGameBetRecord(tx, user, req, amount, gameInfo); err != nil {
				return err
			}
			betAmount, _ := gameBetRecordAmounts(req, amount, gameInfo.IsFishing)
			withdrawFlowAmount := 0.0
			withdrawFlowEventKey := "game_bet:" + tid
			withdrawFlowSourceOrderNo := tid
			if roundAggregated {
				if req.IsEnd {
					roundBetAmount, sumErr := gameCashTransferRoundBetAmount(tx, user.ID, req)
					if sumErr != nil {
						return sumErr
					}
					withdrawFlowAmount = gameWithdrawFlowAmount(roundBetAmount, gameInfo)
					withdrawFlowEventKey = gameCashTransferRoundFlowEventKey(user.ID, req.GameID, req.RoundID)
					withdrawFlowSourceOrderNo = gameCashTransferRoundSourceOrderNo(user.ID, req.GameID, req.RoundID)
				}
			} else {
				withdrawFlowAmount = gameWithdrawFlowAmount(betAmount, gameInfo)
			}
			if withdrawFlowAmount > 0 {
				vipUserID = user.ID
				vipTenantID = user.TenantId
				vipBetFlow = withdrawFlowAmount
				occurredAt := time.Now()
				if req.ReqTime > 0 {
					occurredAt = time.UnixMilli(req.ReqTime)
				}
				if err := repository.RecordWithdrawFlowEvent(
					tx,
					user.ID,
					user.TenantId,
					pojo.WithdrawFlowEventTypeGameBet,
					withdrawFlowEventKey,
					0,
					withdrawFlowSourceOrderNo,
					withdrawFlowAmount,
					occurredAt,
				); err != nil {
					return err
				}
			}

			if roundAggregated {
				if err := upsertGameCashTransferRoundCashHistory(tx, user.ID, req, amount, startBalance, endBalance); err != nil {
					return err
				}
			} else {
				if err := tx.Create(&pojo.CashHistory{
					UserId:      user.ID,
					AwardUni:    awardUni,
					Amount:      amount,
					StartAmount: startBalance,
					EndAmount:   endBalance,
					CashMark:    gameCashTransferMark(req.Reason, amount),
					CashDesc:    gameCashTransferDesc(req),
					Type:        gameCashTransferCashHistoryType(req.Reason, amount),
				}).Error; err != nil {
					return err
				}
			}

			balance = endBalance
			return nil
		}(tx)
		if err != nil {
			if rollbackErr := tx.Rollback().Error; rollbackErr != nil {
				log.Printf("[gsc_open] sport transfer rollback failed userid=%s tid=%s err=%v rollbackErr=%v", userID, tid, err, rollbackErr)
			}
		} else {
			err = tx.Commit().Error
		}
	}
	if err != nil {
		return 0, err
	}
	if vipBetFlow > 0 && vipUserID > 0 {
		throttleKey := fmt.Sprintf("game_bet_post_throttle:%d:%d", vipTenantID, vipUserID)
		if acquired, _ := utils.AcquireLock(throttleKey, vipUpgradeCheckThrottle); acquired {
			uid := vipUserID
			go func() {
				_ = db.Transaction(func(tx2 *gorm.DB) error {
					_, e := repository.EnsureInviteValidUser(tx2, uid, time.Now())
					return e
				})
				repository.CheckAndUpgradeVipLevel(db, uid)
			}()
		}
	}
	return balance, nil
}

func bindAndVerifyGSCOpenRequest(ctx *gin.Context, action string, req any, meta func() (operatorCode string, requestTime string, sign string)) bool {
	body, err := ctx.GetRawData()
	if err != nil {
		log.Printf("[gsc_open] read body failed path=%s ip=%s err=%v", ctx.Request.URL.Path, utils.GetIPAddress(ctx), err)
		gscOpenBaseBack(ctx, game.GameCodeInvalidMerchantCode, game.ErrorMessage(game.GameCodeInvalidMerchantCode))
		return false
	}
	if err := json.Unmarshal(body, req); err != nil {
		log.Printf("[gsc_open] json invalid path=%s ip=%s err=%v body=%s", ctx.Request.URL.Path, utils.GetIPAddress(ctx), err, string(body))
		gscOpenBaseBack(ctx, game.GameCodeInvalidMerchantCode, game.ErrorMessage(game.GameCodeInvalidMerchantCode))
		return false
	}

	operatorCode, requestTime, sign := meta()
	if !verifyGSCOpenSignature(ctx, action, operatorCode, requestTime, sign) {
		return false
	}
	return true
}

func verifyGSCOpenSignature(ctx *gin.Context, action string, operatorCode string, requestTime string, sign string) bool {
	cfg := game.GetGSCConfig()
	operatorCode = strings.TrimSpace(operatorCode)
	requestTime = strings.TrimSpace(requestTime)
	sign = strings.ToLower(strings.TrimSpace(sign))
	if cfg.OperatorCode == "" || cfg.SecretKey == "" {
		log.Printf("[gsc_open] config invalid path=%s operatorCodeConfigured=%t secretConfigured=%t", ctx.Request.URL.Path, cfg.OperatorCode != "", cfg.SecretKey != "")
		gscOpenBaseBack(ctx, game.GameCodeInvalidMerchantCode, "gsc_config_invalid")
		return false
	}
	if operatorCode == "" || operatorCode != cfg.OperatorCode {
		log.Printf("[gsc_open] invalid operator path=%s operator=%s configured=%s ip=%s", ctx.Request.URL.Path, operatorCode, cfg.OperatorCode, utils.GetIPAddress(ctx))
		gscOpenBaseBack(ctx, game.GameCodeInvalidAppID, game.ErrorMessage(game.GameCodeInvalidAppID))
		return false
	}
	if requestTime == "" || sign == "" {
		log.Printf("[gsc_open] empty sign fields path=%s operator=%s requestTime=%s hasSign=%t", ctx.Request.URL.Path, operatorCode, requestTime, sign != "")
		gscOpenBaseBack(ctx, game.GameCodeInvalidMerchantCode, game.ErrorMessage(game.GameCodeInvalidMerchantCode))
		return false
	}
	expected := game.GSCOpenSign(operatorCode, requestTime, action, cfg.SecretKey)
	if subtle.ConstantTimeCompare([]byte(sign), []byte(expected)) != 1 {
		log.Printf("[gsc_open] invalid sign path=%s operator=%s action=%s requestTime=%s ip=%s", ctx.Request.URL.Path, operatorCode, action, requestTime, utils.GetIPAddress(ctx))
		gscOpenBaseBack(ctx, game.GameCodeInvalidMerchantCode, game.ErrorMessage(game.GameCodeInvalidMerchantCode))
		return false
	}
	return true
}

func findGSCOpenUser(db *gorm.DB, memberAccount string) (pojo.TgUser, int, string, error) {
	memberAccount = strings.TrimSpace(memberAccount)
	if memberAccount == "" {
		return pojo.TgUser{}, game.GameCodeEmptyUserID, "", nil
	}
	var user pojo.TgUser
	err := db.Model(&pojo.TgUser{}).
		Select("id, uid, sport_balance, status, tenant_id").
		Where("uid = ? AND status <> ?", memberAccount, int8(-1)).
		First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return pojo.TgUser{}, game.GameCodePlayerNotFound, "", nil
	}
	if err != nil {
		return pojo.TgUser{}, game.GameCodeInvalidMerchantCode, err.Error(), err
	}
	if user.Status != 1 {
		return user, game.GameCodePlayerDisabled, "", nil
	}
	return user, game.GameCodeSuccess, "", nil
}

func gscOpenTransferError(err error) (int, string) {
	var transferErr gameCashTransferError
	if errors.As(err, &transferErr) {
		return transferErr.code, transferErr.msg
	}
	return game.GameCodeInvalidMerchantCode, err.Error()
}

func gscOpenTraceID(action string, tx pojo.GSCOpenWalletTransaction) string {
	raw := firstNonEmpty(tx.ID, tx.WagerCode, tx.RoundID, tx.GameCode)
	if raw == "" {
		raw = fmt.Sprintf("%d", time.Now().UnixNano())
	}
	value := "gsc_" + strings.TrimSpace(action) + "_" + strings.TrimSpace(raw)
	if len(value) <= gscOpenTraceMaxLen {
		return value
	}
	sum := sha1.Sum([]byte(value))
	return "gsc_" + strings.TrimSpace(action) + "_" + hex.EncodeToString(sum[:])
}

func gscOpenPositiveAmount(values ...float64) float64 {
	for _, value := range values {
		value = utils.Truncate2(value)
		if value > 0 {
			return value
		}
		if value < 0 {
			return utils.Truncate2(-value)
		}
	}
	return 0
}

func gscOpenBetAmount(tx pojo.GSCOpenWalletTransaction) float64 {
	return gscOpenPositiveAmount(tx.BetAmount.Float64(), tx.ValidBetAmount.Float64())
}

func gscOpenDepositReason(tx pojo.GSCOpenWalletTransaction) string {
	action := strings.ToLower(strings.TrimSpace(tx.Action))
	status := strings.ToLower(strings.TrimSpace(tx.WagerStatus))
	if strings.Contains(action, "refund") || strings.Contains(action, "rollback") || strings.Contains(status, "refund") || strings.Contains(status, "cancel") || strings.Contains(status, "void") {
		return "refund"
	}
	return "win"
}

func gscOpenIsRoundEnd(tx pojo.GSCOpenWalletTransaction) bool {
	action := strings.ToLower(strings.TrimSpace(tx.Action))
	status := strings.ToLower(strings.TrimSpace(tx.WagerStatus))
	return strings.Contains(action, "settled") || strings.Contains(status, "settled") || strings.Contains(status, "cancel") || strings.Contains(status, "void")
}

func gscOpenTimestampMillis(value int64) int64 {
	if value <= 0 {
		return time.Now().UnixMilli()
	}
	if value < 1_000_000_000_000 {
		return value * 1000
	}
	return value
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func gscOpenMessage(code int, message string) string {
	message = strings.TrimSpace(message)
	if message != "" {
		return message
	}
	return game.ErrorMessage(code)
}

func gscOpenBaseBack(ctx *gin.Context, code int, message string) {
	if code == game.GameCodeSuccess {
		message = ""
	} else {
		message = gscOpenMessage(code, message)
	}
	ctx.JSON(http.StatusOK, pojo.GSCOpenBaseResp{
		Code:    code,
		Message: message,
	})
}
