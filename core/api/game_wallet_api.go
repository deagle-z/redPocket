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
	"sync"
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
	if code, msg := validateGameCashTransferReq(req); code != game.GameCodeSuccess {
		log.Printf("[game_wallet] TransferInOut validate failed requestId=%s userid=%s tid=%s code=%d msg=%s",
			ctx.GetHeader(game.HeaderRequestID), strings.TrimSpace(req.UserID), strings.TrimSpace(req.TID), code, msg)
		gameErrorBack(ctx, code, msg)
		return
	}

	db := ctx.MustGet("db").(*gorm.DB)
	balance, err := handleGameCashTransfer(db, req)
	if err != nil {
		log.Printf("[game_wallet] TransferInOut failed requestId=%s userid=%s tid=%s err=%v",
			ctx.GetHeader(game.HeaderRequestID), strings.TrimSpace(req.UserID), strings.TrimSpace(req.TID), err)
		writeGameCashTransferError(ctx, err)
		return
	}
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
	case "win":
		return game.GameCodeSuccess, ""
	case "refund":
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

// vipUpgradeCheckThrottle 游戏投注触发 VIP 升级检查的每用户节流窗口
const vipUpgradeCheckThrottle = 10 * time.Second

// gameTransferSlowThreshold 超过该耗时则打印各步骤分解的慢日志（设为 0 可记录每一笔）
const gameTransferSlowThreshold = 300 * time.Millisecond

func msOf(d time.Duration) int64 { return d.Milliseconds() }

func handleGameCashTransfer(db *gorm.DB, req pojo.GameCashTransferInOutReq) (float64, error) {
	userID := strings.TrimSpace(req.UserID)
	tid := strings.TrimSpace(req.TID)
	awardUni := gameCashTransferAwardUni(tid)
	var balance float64
	var vipUserID int64
	var vipTenantID int64
	var vipBetFlow float64

	// ===== 耗时打点 =====
	reqStart := time.Now()
	var tGameInfo, tTx, tTxBegin, tTxCommit, tTxRollback, tLock, tIdem, tBalance, tBetRec, tFlow, tCash time.Duration
	idempotent := false

	// 游戏元数据查询(带缓存)放在事务外，避免占用行锁时间
	g0 := time.Now()
	gameInfo := getGameCashTransferGameInfo(db, req.GameID)
	tGameInfo = time.Since(g0)

	txStart := time.Now()
	sBegin := time.Now()
	tx := db.Begin()
	tTxBegin = time.Since(sBegin)
	var err error
	if tx.Error != nil {
		err = tx.Error
	} else {
		err = func(tx *gorm.DB) error {
			s := time.Now()
			var user pojo.TgUser
			err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
				Where("uid = ? AND status <> ?", userID, int8(-1)).
				First(&user).Error
			tLock = time.Since(s)
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					log.Printf("[game_wallet] TransferInOut user not found userid=%s tid=%s", userID, tid)
					return newGameCashTransferError(game.GameCodePlayerNotFound, "")
				}
				return err
			}
			s = time.Now()
			var history pojo.CashHistory
			err = tx.Where("user_id = ? AND award_uni = ?", user.ID, awardUni).First(&history).Error
			tIdem = time.Since(s)
			if err == nil {
				balance = utils.Truncate2(history.EndAmount)
				idempotent = true
				return nil
			}
			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
			if user.Status != 1 {
				log.Printf("[game_wallet] TransferInOut user disabled userid=%s tid=%s userId=%d status=%d",
					userID, tid, user.ID, user.Status)
				return newGameCashTransferError(game.GameCodePlayerDisabled, "")
			}

			amount := utils.Truncate2(req.Amount)
			if code, msg := validateGameCashTransferAmount(req, amount, gameInfo.IsFishing); code != game.GameCodeSuccess {
				return newGameCashTransferError(code, msg)
			}
			if amount < 0 && utils.Truncate2(user.Balance+amount) < 0 {
				log.Printf("[game_wallet] TransferInOut insufficient balance userid=%s tid=%s userId=%d balance=%.2f amount=%.2f",
					userID, tid, user.ID, utils.Truncate2(user.Balance), amount)
				return newGameCashTransferError(game.GameCodeInsufficientBalance, game.ErrorMessage(game.GameCodeInsufficientBalance))
			}

			startBalance := utils.Truncate2(user.Balance)
			endBalance := utils.Truncate2(startBalance + amount)
			s = time.Now()
			if err := tx.Model(&pojo.TgUser{}).
				Where("id = ?", user.ID).
				Update("balance", endBalance).Error; err != nil {
				return err
			}
			tBalance = time.Since(s)

			s = time.Now()
			if err := createGameBetRecord(tx, user, req, amount, gameInfo); err != nil {
				return err
			}
			tBetRec = time.Since(s)
			betAmount, _ := gameBetRecordAmounts(req, amount, gameInfo.IsFishing)
			if betAmount > 0 {
				vipUserID = user.ID
				vipTenantID = user.TenantId
				vipBetFlow = betAmount
				occurredAt := time.Now()
				if req.ReqTime > 0 {
					occurredAt = time.UnixMilli(req.ReqTime)
				}
				s = time.Now()
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
				tFlow = time.Since(s)
				// 邀请有效用户标记不再同步发生（仅统计用、非资金关键），移到事务提交后异步处理
			}
			s = time.Now()
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
			tCash = time.Since(s)

			balance = endBalance
			return nil
		}(tx)
		if err != nil {
			sRollback := time.Now()
			if rollbackErr := tx.Rollback().Error; rollbackErr != nil {
				log.Printf("[game_wallet] TransferInOut rollback failed userid=%s tid=%s err=%v rollbackErr=%v", userID, tid, err, rollbackErr)
			}
			tTxRollback = time.Since(sRollback)
		} else {
			sCommit := time.Now()
			err = tx.Commit().Error
			tTxCommit = time.Since(sCommit)
		}
	}
	tTx = time.Since(txStart)
	// 慢日志：定位各步骤耗时（lock=行锁+读用户, idem=幂等查, balance=改余额, betRec=下注记录, flow=提现流水, cash=账变）
	if total := time.Since(reqStart); total >= gameTransferSlowThreshold {
		measuredTx := tTxBegin + tLock + tIdem + tBalance + tBetRec + tFlow + tCash + tTxCommit + tTxRollback
		txOther := tTx - measuredTx
		if txOther < 0 {
			txOther = 0
		}
		log.Printf("[game_wallet][slow] TransferInOut tid=%s userid=%s reason=%s idempotent=%t total=%dms gameInfo=%dms tx=%dms txBegin=%dms lock=%dms idem=%dms balance=%dms betRec=%dms flow=%dms cash=%dms txCommit=%dms txRollback=%dms txOther=%dms",
			tid, userID, strings.ToLower(strings.TrimSpace(req.Reason)), idempotent,
			msOf(total), msOf(tGameInfo), msOf(tTx), msOf(tTxBegin), msOf(tLock), msOf(tIdem), msOf(tBalance), msOf(tBetRec), msOf(tFlow), msOf(tCash), msOf(tTxCommit), msOf(tTxRollback), msOf(txOther))
	}
	if err != nil {
		return 0, err
	}
	// 投注结算后的非资金关键处理（VIP 升级检查 + 邀请有效用户标记）异步执行，不阻塞下注响应。
	// 每用户短 TTL 节流：高频投注下同一用户在 vipUpgradeCheckThrottle 窗口内只跑一次，降低 DB 压力。
	if vipBetFlow > 0 && vipUserID > 0 {
		throttleKey := fmt.Sprintf("game_bet_post_throttle:%d:%d", vipTenantID, vipUserID)
		if acquired, _ := utils.AcquireLock(throttleKey, vipUpgradeCheckThrottle); acquired {
			uid := vipUserID
			go func() {
				// 邀请有效用户标记（含分表流水/充值聚合查询，原本在同步事务里，移出热点路径）
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

func validateGameCashTransferAmount(req pojo.GameCashTransferInOutReq, amount float64, isFishingGame bool) (int, string) {
	reason := strings.ToLower(strings.TrimSpace(req.Reason))
	if reason == "win" && amount < 0 && !isFishingGame {
		return game.GameCodeInvalidTransferAmount, game.ErrorMessage(game.GameCodeInvalidTransferAmount)
	}
	return game.GameCodeSuccess, ""
}

type gameCashTransferGameInfo struct {
	GameName  string
	IsFishing bool
}

type cachedGameInfo struct {
	info gameCashTransferGameInfo
	at   time.Time
}

var gameInfoCache sync.Map // thirdGameID(string) -> cachedGameInfo

const gameInfoCacheTTL = 5 * time.Minute

// getGameCashTransferGameInfo 带 TTL 缓存的游戏信息查询（游戏元数据基本不变，避免每次下注都查库）。
// 仅缓存命中(找到)的结果，未找到不缓存以便新游戏及时生效。
func getGameCashTransferGameInfo(db *gorm.DB, thirdGameID string) gameCashTransferGameInfo {
	key := strings.TrimSpace(thirdGameID)
	if key == "" {
		return gameCashTransferGameInfo{}
	}
	if v, ok := gameInfoCache.Load(key); ok {
		if c, ok2 := v.(cachedGameInfo); ok2 && time.Since(c.at) < gameInfoCacheTTL {
			return c.info
		}
	}
	info, found := lookupGameCashTransferGameInfo(db, key)
	if found {
		gameInfoCache.Store(key, cachedGameInfo{info: info, at: time.Now()})
	}
	return info
}

func lookupGameCashTransferGameInfo(db *gorm.DB, thirdGameID string) (gameCashTransferGameInfo, bool) {
	var appGame pojo.AppGame
	err := db.Model(&pojo.AppGame{}).
		Select("game_name, category_code, type").
		Where("third_game_id = ? AND COALESCE(deleted_flag, 0) = 0", strings.TrimSpace(thirdGameID)).
		First(&appGame).Error
	if err != nil {
		return gameCashTransferGameInfo{}, false
	}
	categoryCode := strings.ToLower(strings.TrimSpace(appGameStringValue(appGame.CategoryCode)))
	isFishing := categoryCode == "fishing" || (appGame.Type != nil && *appGame.Type == 3)
	return gameCashTransferGameInfo{
		GameName:  appGameStringValue(appGame.GameName),
		IsFishing: isFishing,
	}, true
}

func createGameBetRecord(tx *gorm.DB, user pojo.TgUser, req pojo.GameCashTransferInOutReq, amount float64, gameInfo gameCashTransferGameInfo) error {
	now := time.Now()
	if req.ReqTime > 0 {
		now = time.UnixMilli(req.ReqTime)
	}
	gameID := strings.TrimSpace(req.GameID)
	roundID := strings.TrimSpace(req.RoundID)
	tid := strings.TrimSpace(req.TID)
	platformCode := "hg"
	gameName := gameInfo.GameName
	roundEnd := 0
	if req.IsEnd {
		roundEnd = 1
	}
	betAmount, winAmount := gameBetRecordAmounts(req, amount, gameInfo.IsFishing)
	uid := parseGameUserNumericUID(user.Uid)
	remark := gameCashTransferDesc(req)

	return tx.Exec(
		"INSERT INTO `"+pojo.AppUserBetRecordTableNameByUserID(user.ID)+"` "+
			"(`uid`, `user_id`, `game_id`, `game_name`, `platform_code`, `bet_amount`, `win_amount`, `round_id`, `trace_id`, `round_end`, `date`, `remark`, `create_time`, `update_time`) "+
			"VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		uid,
		user.ID,
		gameID,
		gameName,
		platformCode,
		betAmount,
		winAmount,
		roundID,
		tid,
		roundEnd,
		now,
		remark,
		now,
		now,
	).Error
}

func gameBetRecordAmounts(req pojo.GameCashTransferInOutReq, amount float64, isFishingGame bool) (float64, float64) {
	betAmount := utils.Truncate2(req.Bet)
	winAmount := 0.0
	switch strings.ToLower(strings.TrimSpace(req.Reason)) {
	case "bet":
		if betAmount <= 0 {
			betAmount = utils.Truncate2(-amount)
		}
	case "win":
		if isFishingGame && amount < 0 {
			if betAmount <= 0 {
				betAmount = utils.Truncate2(-amount)
			}
			break
		}
		if amount > 0 {
			winAmount = amount
		}
	case "refund":
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

func gameCashTransferCashHistoryType(reason string, amount float64) int8 {
	switch strings.ToLower(strings.TrimSpace(reason)) {
	case "bet":
		return pojo.CashHistoryTypeGameBet
	case "win":
		if amount < 0 {
			return pojo.CashHistoryTypeGameBet
		}
		return pojo.CashHistoryTypeGameWin
	case "refund":
		return pojo.CashHistoryTypeGameRefund
	default:
		return pojo.CashHistoryTypeUnknown
	}
}

func gameCashTransferMark(reason string, amount float64) string {
	switch strings.ToLower(strings.TrimSpace(reason)) {
	case "bet":
		return "游戏下注"
	case "win":
		if amount < 0 {
			return "游戏下注"
		}
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
