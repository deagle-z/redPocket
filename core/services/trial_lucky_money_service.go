package services

import (
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/utils"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"log"
	"math"
	"math/rand/v2"
	"strconv"
	"strings"
	"sync"
	"time"
)

var trialAutoLuckyMaintainMu sync.Mutex

const (
	trialUserWinRateDefault = 0.80
	trialUserWinRateConfig  = "trial_user_win_rate"
	trialUserWinRateCache   = "bgu_trial_user_win_rate"
)

func GetTrialMe(db *gorm.DB, userID int64) (pojo.TrialMeResp, error) {
	var user pojo.TgUser
	if err := db.Select("id, trial_balance, status").Where("id = ?", userID).First(&user).Error; err != nil {
		return pojo.TrialMeResp{}, errors.New("user_not_found")
	}
	if user.Status != 1 {
		return pojo.TrialMeResp{}, errors.New("user_disabled_contact_admin")
	}
	return pojo.TrialMeResp{TrialBalance: utils.Truncate2(user.TrialBalance)}, nil
}

func RefreshTrialBalanceDaily(db *gorm.DB, userID int64, now time.Time) (pojo.TrialBalanceRefreshResp, error) {
	if now.IsZero() {
		now = time.Now()
	}
	var result pojo.TrialBalanceRefreshResp
	err := db.Transaction(func(tx *gorm.DB) error {
		var user pojo.TgUser
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Select("id, trial_balance, trial_balance_refreshed_at, status").
			Where("id = ?", userID).
			First(&user).Error; err != nil {
			return errors.New("user_not_found")
		}
		if user.Status != 1 {
			return errors.New("user_disabled_contact_admin")
		}

		balance, refreshed, shouldMark := resolveTrialDailyBalanceRefresh(user.TrialBalance, user.TrialBalanceRefreshedAt, now)
		updates := map[string]any{}
		if shouldMark {
			updates["trial_balance_refreshed_at"] = now
		}
		if balance != utils.Truncate2(user.TrialBalance) {
			updates["trial_balance"] = balance
		}
		if len(updates) > 0 {
			if err := tx.Model(&pojo.TgUser{}).Where("id = ?", user.ID).Updates(updates).Error; err != nil {
				return err
			}
		}
		result = pojo.TrialBalanceRefreshResp{
			TrialBalance: balance,
			Refreshed:    refreshed,
		}
		return nil
	})
	return result, err
}

func resolveTrialDailyBalanceRefresh(balance float64, refreshedAt *time.Time, now time.Time) (float64, bool, bool) {
	currentBalance := utils.Truncate2(balance)
	if refreshedAt != nil && sameTrialRefreshDay(*refreshedAt, now) {
		return currentBalance, false, false
	}
	if currentBalance < pojo.TrialUserDefaultBalance {
		return pojo.TrialUserDefaultBalance, true, true
	}
	return currentBalance, false, true
}

func sameTrialRefreshDay(a time.Time, b time.Time) bool {
	if a.IsZero() || b.IsZero() {
		return false
	}
	localA := a.In(b.Location())
	return localA.Year() == b.Year() && localA.YearDay() == b.YearDay()
}

func validateTrialSendAmount(amount float64) error {
	if amount < 5 {
		return errors.New("amount_min_5")
	}
	if amount > pojo.TrialLuckySendMaxAmount {
		return errors.New("amount_max_500")
	}
	return nil
}

func capTrialLuckySendAmount(amount float64) float64 {
	amount = utils.Truncate2(amount)
	if amount > pojo.TrialLuckySendMaxAmount {
		return pojo.TrialLuckySendMaxAmount
	}
	return amount
}

func SendTrialRedPacket(db *gorm.DB, userID int64, req pojo.TrialLuckyMoneySend, tablePrefix string) (pojo.TrialLuckyMoney, error) {
	req.Amount = utils.Truncate2(req.Amount)
	if err := validateTrialSendAmount(req.Amount); err != nil {
		return pojo.TrialLuckyMoney{}, err
	}
	if req.Thunder < 0 || req.Thunder > 9 {
		return pojo.TrialLuckyMoney{}, errors.New("thunder_range_0_9")
	}
	req.GameMode = normalizeTrialGameMode(req.GameMode)
	number := resolveTrialLuckyNumber(db, req.Number)
	if number <= 0 {
		return pojo.TrialLuckyMoney{}, errors.New("lucky_number_invalid")
	}
	if req.Amount < float64(number)*0.01 {
		return pojo.TrialLuckyMoney{}, errors.New("amount_too_small")
	}
	loseRate := GetLoseRate(db)
	if req.GameMode == 1 {
		loseRate = GetGame2LoseRate(db)
	}
	if loseRate <= 0 {
		loseRate = 1.8
	}
	expireTime := time.Now().Add(time.Duration(GetLuckyExpireMinutes(db)) * time.Minute)
	redList := utils.RedEnvelope(req.Amount, number, 0.01, math.Max(0.01, req.Amount/float64(number)*2))
	redListJSON, _ := json.Marshal(redList)

	var lucky pojo.TrialLuckyMoney
	err := db.Transaction(func(tx *gorm.DB) error {
		var user pojo.TgUser
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", userID).First(&user).Error; err != nil {
			return errors.New("user_not_found")
		}
		if user.Status != 1 {
			return errors.New("user_disabled_contact_admin")
		}
		if user.TrialBalance < req.Amount {
			return errors.New("trial_balance_insufficient")
		}
		startBalance := utils.Truncate2(user.TrialBalance)
		endBalance := utils.Truncate2(startBalance - req.Amount)
		displayName := trialTgDisplayName(user)
		lucky = pojo.TrialLuckyMoney{
			SenderID:     user.ID,
			SenderType:   pojo.TrialActorUser,
			SenderName:   displayName,
			SenderAvatar: user.Avatar,
			Amount:       req.Amount,
			Received:     0,
			Number:       number,
			Thunder:      req.Thunder,
			GameMode:     req.GameMode,
			ChatID:       req.ChatID,
			RedList:      string(redListJSON),
			LoseRate:     loseRate,
			Status:       1,
			TenantId:     req.TenantId,
			ExpireTime:   expireTime,
		}
		if lucky.TenantId == 0 {
			lucky.TenantId = user.TenantId
		}
		if err := tx.Create(&lucky).Error; err != nil {
			return err
		}
		for i, amount := range redList {
			item := pojo.TrialLuckyMoneyItem{
				RedPacketID: lucky.ID,
				SeqNo:       uint(i + 1),
				Amount:      utils.Truncate2(amount),
				IsGrabbed:   0,
				Thunder:     0,
			}
			if err := tx.Create(&item).Error; err != nil {
				return err
			}
		}
		if err := tx.Model(&pojo.TgUser{}).Where("id = ?", user.ID).Update("trial_balance", endBalance).Error; err != nil {
			return err
		}
		return createTrialCashHistory(tx, pojo.TrialCashHistory{
			UserId:      user.ID,
			ActorType:   pojo.TrialActorUser,
			AwardUni:    fmt.Sprintf("trial_send_%d", lucky.ID),
			Amount:      -req.Amount,
			StartAmount: startBalance,
			EndAmount:   endBalance,
			CashMark:    "trial_lucky_send",
			CashDesc:    "试玩发包",
			Type:        pojo.TrialCashTypeSendLucky,
			LuckyID:     lucky.ID,
			TenantId:    lucky.TenantId,
		})
	})
	if err == nil {
		BroadcastTrialLuckySent(db, lucky.ID)
		if err := EnqueueTrialLuckyBotGrabTask(db, tablePrefix, lucky.ID, nil, pickRandomTrialBotGrabCount(lucky.Number)); err != nil {
			log.Printf("[trial_lucky] enqueue bot grab failed luckyID=%d err=%v", lucky.ID, err)
		}
	}
	return lucky, err
}

func GrabTrialRedPacket(db *gorm.DB, luckyID int64, userID int64, tablePrefix string, grabIndex int, oddEvenGuess *int) (map[string]any, error) {
	if luckyID <= 0 || userID <= 0 {
		return nil, errors.New("invalid_params")
	}
	result := map[string]any{}
	finished := false
	err := db.Transaction(func(tx *gorm.DB) error {
		var lucky pojo.TrialLuckyMoney
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", luckyID).First(&lucky).Error; err != nil {
			return errors.New("lucky_not_found")
		}
		if lucky.Status != 1 {
			return errors.New("lucky_finished")
		}
		if !lucky.ExpireTime.IsZero() && time.Now().After(lucky.ExpireTime) {
			if err := refundExpiredTrialLucky(tx, &lucky); err != nil {
				return err
			}
			return errors.New("lucky_expired")
		}
		if !canTrialUserGrabSender(lucky.SenderType, lucky.SenderID, userID) {
			return errors.New("cannot_grab_self")
		}
		if lucky.GameMode == 1 {
			if oddEvenGuess == nil || (*oddEvenGuess != 0 && *oddEvenGuess != 1) {
				return errors.New("odd_even_guess_required")
			}
		}
		var user pojo.TgUser
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", userID).First(&user).Error; err != nil {
			return errors.New("user_not_found")
		}
		if user.Status != 1 {
			return errors.New("user_disabled_contact_admin")
		}
		requiredBalance := utils.Truncate2(lucky.Amount * lucky.LoseRate)
		if lucky.GameMode == 1 {
			requiredBalance = utils.Truncate2((lucky.Amount / float64(maxInt(1, lucky.Number))) * lucky.LoseRate)
		}
		if user.TrialBalance < requiredBalance {
			return errors.New("trial_balance_insufficient")
		}

		item, err := pickTrialLuckyItemForUser(tx, lucky, grabIndex, oddEvenGuess, trialTargetUserWin(tx))
		if err != nil {
			return err
		}
		now := time.Now()
		openNum := trialOpenNum(item.Amount)
		isThunder := trialIsThunder(lucky, item.Amount, oddEvenGuess)
		loseMoney := 0.0
		if isThunder {
			if lucky.GameMode == 1 {
				loseMoney = utils.Truncate2(item.Amount * lucky.LoseRate)
			} else {
				loseMoney = utils.Truncate2(lucky.Amount * lucky.LoseRate)
			}
		}
		actualAmount := utils.Truncate2(item.Amount - loseMoney)
		startBalance := utils.Truncate2(user.TrialBalance)
		endBalance := utils.Truncate2(startBalance + actualAmount)
		thunderFlag := int8(0)
		if isThunder {
			thunderFlag = 1
		}

		if err := tx.Model(&pojo.TrialLuckyMoneyItem{}).
			Where("id = ? AND is_grabbed = ?", item.ID, 0).
			Updates(map[string]any{
				"is_grabbed":   1,
				"thunder":      thunderFlag,
				"grabbed_uid":  user.ID,
				"grabbed_type": pojo.TrialActorUser,
				"grabbed_at":   now,
			}).Error; err != nil {
			return err
		}
		if err := tx.Model(&pojo.TgUser{}).Where("id = ?", user.ID).Update("trial_balance", endBalance).Error; err != nil {
			return err
		}
		if err := tx.Model(&pojo.TrialLuckyMoney{}).Where("id = ?", lucky.ID).
			Update("received", gorm.Expr("received + ?", item.Amount)).Error; err != nil {
			return err
		}

		history := pojo.TrialLuckyHistory{
			UserID:       user.ID,
			ActorType:    pojo.TrialActorUser,
			FirstName:    trialTgDisplayName(user),
			LuckyID:      lucky.ID,
			IsThunder:    int(thunderFlag),
			GrabType:     1,
			Amount:       utils.Truncate2(item.Amount),
			ActualAmount: actualAmount,
			LoseMoney:    loseMoney,
			Guess:        oddEvenGuess,
			TenantId:     lucky.TenantId,
		}
		if err := tx.Create(&history).Error; err != nil {
			return err
		}
		if err := createTrialCashHistory(tx, pojo.TrialCashHistory{
			UserId:      user.ID,
			ActorType:   pojo.TrialActorUser,
			AwardUni:    fmt.Sprintf("trial_grab_%d_%d_%d_%d", lucky.ID, user.ID, item.SeqNo, now.UnixNano()),
			Amount:      actualAmount,
			StartAmount: startBalance,
			EndAmount:   endBalance,
			CashMark:    "trial_lucky_grab",
			CashDesc:    "试玩抢包",
			Type:        mapTrialGrabCashType(isThunder),
			IsThunder:   thunderFlag,
			LuckyID:     lucky.ID,
			TenantId:    lucky.TenantId,
		}); err != nil {
			return err
		}
		lotteryRewardCount, trialFlowLotteryRewarded, err := awardTrialLuckyFlowLotteryIfNeeded(tx, user.ID, lucky.TenantId, lucky.ID, history.ID, tablePrefix)
		if err != nil {
			return err
		}
		if isThunder && loseMoney > 0 {
			if err := addTrialSenderThunderIncome(tx, lucky, loseMoney); err != nil {
				return err
			}
		}

		var grabbedCount int64
		if err := tx.Model(&pojo.TrialLuckyMoneyItem{}).Where("red_packet_id = ? AND is_grabbed = ?", lucky.ID, 1).Count(&grabbedCount).Error; err != nil {
			return err
		}
		if int(grabbedCount) >= lucky.Number {
			if err := tx.Model(&pojo.TrialLuckyMoney{}).Where("id = ?", lucky.ID).Update("status", 2).Error; err != nil {
				return err
			}
			finished = true
		}
		result = map[string]any{
			"luckyId":                  lucky.ID,
			"actorType":                pojo.TrialActorUser,
			"userId":                   user.ID,
			"firstName":                trialTgDisplayName(user),
			"grabIndex":                item.SeqNo,
			"amount":                   utils.Truncate2(item.Amount),
			"actualAmount":             actualAmount,
			"loseMoney":                loseMoney,
			"isThunder":                int(thunderFlag),
			"openNum":                  openNum,
			"balance":                  endBalance,
			"message":                  "success",
			"lotteryRewardCount":       lotteryRewardCount,
			"trialFlowLotteryRewarded": trialFlowLotteryRewarded,
		}
		return nil
	})
	if err == nil {
		BroadcastTrialLuckyGrabResult(db, luckyID, result)
		if finished {
			_ = EnsureMinActiveTrialLuckyPackets(db, tablePrefix)
			BroadcastTrialLuckyFinished(db, luckyID)
		}
	}
	return result, err
}

func GetTrialLuckyAppList(db *gorm.DB, search pojo.TrialLuckyMoneyAppListSearch, userID int64) (result pojo.TrialLuckyMoneyAppResp) {
	var list []pojo.TrialLuckyMoney
	query := db.Model(&pojo.TrialLuckyMoney{})
	if search.LuckyID > 0 {
		query = query.Where("id = ?", search.LuckyID)
	}
	if search.ChatID != 0 {
		query = query.Where("chat_id = ?", search.ChatID)
	}
	if search.Status != nil {
		query = query.Where("status = ?", *search.Status)
	}
	if search.GameMode != nil {
		query = query.Where("game_mode = ?", *search.GameMode)
	}
	query.Count(&result.Total)
	query.Order("id desc").Limit(search.PageSize).Offset(search.PageSize * search.CurrentPage).Find(&list)
	for _, lucky := range list {
		back := buildTrialLuckyAppBack(db, lucky, userID)
		result.List = append(result.List, back)
	}
	result.PageSize = search.PageSize
	result.CurrentPage = search.CurrentPage
	return result
}

func GetTrialAppHistory(db *gorm.DB, userID int64, search pojo.TrialCashHistorySearch) pojo.TrialCashHistoryPage {
	search.UserId = userID
	search.ActorType = pojo.TrialActorUser
	return getTrialCashHistoryPage(db, search)
}

func getTrialCashHistoryPage(db *gorm.DB, search pojo.TrialCashHistorySearch) (result pojo.TrialCashHistoryPage) {
	result.PageSize = search.PageSize
	result.CurrentPage = search.CurrentPage
	return result
}

func buildTrialLuckyAppBack(db *gorm.DB, lucky pojo.TrialLuckyMoney, userID int64) pojo.TrialLuckyMoneyAppBack {
	var items []pojo.TrialLuckyMoneyItem
	db.Where("red_packet_id = ?", lucky.ID).Order("seq_no asc").Find(&items)
	var hitCount int64
	db.Model(&pojo.TrialLuckyMoneyItem{}).Where("red_packet_id = ? AND thunder = ?", lucky.ID, 1).Count(&hitCount)
	itemBacks := make([]pojo.TrialLuckyMoneyAppItemBack, 0, len(items))
	grabbedCount := int64(0)
	for _, item := range items {
		if item.IsGrabbed == 1 {
			grabbedCount++
		}
		isMine := int8(0)
		if item.GrabbedType == pojo.TrialActorUser && item.GrabbedUid != nil && *item.GrabbedUid == userID {
			isMine = 1
		}
		amount := 0.0
		if item.IsGrabbed == 1 || isMine == 1 || lucky.Status == 2 {
			amount = item.Amount
		}
		itemBacks = append(itemBacks, pojo.TrialLuckyMoneyAppItemBack{
			SeqNo:      item.SeqNo,
			Amount:     amount,
			IsGrabbed:  item.IsGrabbed,
			Thunder:    item.Thunder,
			IsGrabMine: isMine,
		})
	}
	remainingSeconds := int64(0)
	if !lucky.ExpireTime.IsZero() {
		remainingSeconds = int64(time.Until(lucky.ExpireTime).Seconds())
		if remainingSeconds < 0 {
			remainingSeconds = 0
		}
	}
	return pojo.TrialLuckyMoneyAppBack{
		ID:               lucky.ID,
		SenderID:         lucky.SenderID,
		SenderType:       lucky.SenderType,
		SenderName:       lucky.SenderName,
		SenderAvatar:     lucky.SenderAvatar,
		Amount:           lucky.Amount,
		Received:         lucky.Received,
		Number:           lucky.Number,
		GrabbedCount:     grabbedCount,
		Thunder:          lucky.Thunder,
		GameMode:         lucky.GameMode,
		HitCount:         hitCount,
		LoseRate:         lucky.LoseRate,
		Status:           lucky.Status,
		ExpireTime:       lucky.ExpireTime,
		RemainingSeconds: remainingSeconds,
		RemainingText:    formatTrialRemainingText(remainingSeconds),
		Items:            itemBacks,
		CreatedAt:        lucky.CreatedAt,
	}
}

func BroadcastTrialLuckySent(db *gorm.DB, luckyID int64) {
	if db == nil || luckyID <= 0 {
		return
	}
	detail, err := getTrialLuckyBroadcastDetail(db, luckyID)
	if err != nil {
		log.Printf("[trial_lucky] broadcast sent detail failed luckyID=%d err=%v", luckyID, err)
		return
	}
	if err := utils.BroadcastWsWithType("trial_lucky_sent", detail); err != nil {
		log.Printf("[trial_lucky] broadcast trial_lucky_sent failed luckyID=%d err=%v", luckyID, err)
	}
}

func BroadcastTrialLuckyGrabResult(db *gorm.DB, luckyID int64, result map[string]any) {
	if db == nil || luckyID <= 0 {
		return
	}
	detail, err := getTrialLuckyBroadcastDetail(db, luckyID)
	if err != nil {
		log.Printf("[trial_lucky] broadcast grabbed detail failed luckyID=%d err=%v", luckyID, err)
		return
	}
	payload := map[string]any{
		"id":           detail.ID,
		"luckyId":      detail.ID,
		"senderId":     detail.SenderID,
		"senderType":   detail.SenderType,
		"senderName":   detail.SenderName,
		"amount":       detail.Amount,
		"received":     detail.Received,
		"number":       detail.Number,
		"grabbedCount": detail.GrabbedCount,
		"thunder":      detail.Thunder,
		"gameMode":     detail.GameMode,
		"hitCount":     detail.HitCount,
		"loseRate":     detail.LoseRate,
		"status":       detail.Status,
		"expireTime":   detail.ExpireTime,
		"items":        detail.Items,
		"lucky":        detail,
	}
	for key, val := range result {
		payload[key] = val
	}
	if err := utils.BroadcastWsWithType("trial_lucky_grabbed", payload); err != nil {
		log.Printf("[trial_lucky] broadcast trial_lucky_grabbed failed luckyID=%d err=%v", luckyID, err)
	}
}

func BroadcastTrialLuckyFinished(db *gorm.DB, luckyID int64) {
	if db == nil || luckyID <= 0 {
		return
	}
	detail, err := getTrialLuckyBroadcastDetail(db, luckyID)
	if err != nil {
		log.Printf("[trial_lucky] broadcast finished detail failed luckyID=%d err=%v", luckyID, err)
		return
	}
	if err := utils.BroadcastWsWithType("trial_lucky_finished", detail); err != nil {
		log.Printf("[trial_lucky] broadcast trial_lucky_finished failed luckyID=%d err=%v", luckyID, err)
	}
}

func getTrialLuckyBroadcastDetail(db *gorm.DB, luckyID int64) (pojo.TrialLuckyMoneyAppBack, error) {
	var lucky pojo.TrialLuckyMoney
	if err := db.Where("id = ?", luckyID).First(&lucky).Error; err != nil {
		return pojo.TrialLuckyMoneyAppBack{}, err
	}
	return buildTrialLuckyAppBack(db, lucky, 0), nil
}

func EnsureMinActiveTrialLuckyPackets(db *gorm.DB, tablePrefix string) error {
	if db == nil {
		return nil
	}
	trialAutoLuckyMaintainMu.Lock()
	defer trialAutoLuckyMaintainMu.Unlock()
	if err := ensureMinTrialForMode(db, tablePrefix, 0); err != nil {
		return err
	}
	return ensureMinTrialForMode(db, tablePrefix, 1)
}

func EnsureMinActiveTrialLuckyPacketsAllHosts() {
	var hostInfos []pojo.HostInfo
	if err := utils.Db.Model(&pojo.HostInfo{}).
		Where("enabled = ?", true).
		Where("table_prefix <> ''").
		Find(&hostInfos).Error; err != nil {
		log.Printf("[trial_lucky] ensure all hosts query failed: %v", err)
		return
	}
	seenPrefixes := make(map[string]struct{}, len(hostInfos))
	for _, hostInfo := range hostInfos {
		if hostInfo.TablePrefix == "" {
			continue
		}
		if _, ok := seenPrefixes[hostInfo.TablePrefix]; ok {
			continue
		}
		seenPrefixes[hostInfo.TablePrefix] = struct{}{}
		if err := EnsureMinActiveTrialLuckyPackets(utils.NewPrefixDb(hostInfo.TablePrefix), hostInfo.TablePrefix); err != nil {
			log.Printf("[trial_lucky] ensure prefix=%s failed: %v", hostInfo.TablePrefix, err)
		}
	}
}

func SweepExpiredTrialLuckyPacketsAllHosts() {
	var hostInfos []pojo.HostInfo
	if err := utils.Db.Model(&pojo.HostInfo{}).
		Where("enabled = ?", true).
		Where("table_prefix <> ''").
		Find(&hostInfos).Error; err != nil {
		log.Printf("[trial_lucky] sweep host infos failed: %v", err)
		return
	}
	seenPrefixes := make(map[string]struct{}, len(hostInfos))
	for _, hostInfo := range hostInfos {
		if hostInfo.TablePrefix == "" {
			continue
		}
		if _, ok := seenPrefixes[hostInfo.TablePrefix]; ok {
			continue
		}
		seenPrefixes[hostInfo.TablePrefix] = struct{}{}
		db := utils.NewPrefixDb(hostInfo.TablePrefix)
		if db == nil {
			continue
		}
		var luckyIDs []int64
		if err := db.Model(&pojo.TrialLuckyMoney{}).
			Where("status = ?", 1).
			Where("expire_time > ? AND expire_time <= ?", time.Time{}, time.Now()).
			Limit(200).
			Pluck("id", &luckyIDs).Error; err != nil {
			log.Printf("[trial_lucky] sweep query failed prefix=%s err=%v", hostInfo.TablePrefix, err)
			continue
		}
		for _, luckyID := range luckyIDs {
			if err := refundExpiredTrialLuckyByID(db, luckyID); err != nil {
				log.Printf("[trial_lucky] sweep refund failed prefix=%s luckyID=%d err=%v", hostInfo.TablePrefix, luckyID, err)
				continue
			}
			BroadcastTrialLuckyFinished(db, luckyID)
		}
		if err := EnsureMinActiveTrialLuckyPackets(db, hostInfo.TablePrefix); err != nil {
			log.Printf("[trial_lucky] sweep ensure failed prefix=%s err=%v", hostInfo.TablePrefix, err)
		}
	}
}

func ensureMinTrialForMode(db *gorm.DB, tablePrefix string, gameMode int) error {
	for i := 0; i < 3; i++ {
		activeCount, err := countActiveTrialLuckyMoneyByMode(db, gameMode)
		if err != nil {
			return err
		}
		if activeCount >= 3 {
			return nil
		}
		bot, err := pickRandomTrialSendBotUser(db)
		if err != nil || bot.ID == 0 {
			return err
		}
		amount := pickRandomTrialBotAmount(tablePrefix)
		if err = ensureTrialBotBalance(db, &bot, amount); err != nil {
			return err
		}
		if _, err = sendTrialRedPacketByBot(db, bot, amount, gameMode, tablePrefix); err != nil {
			return err
		}
	}
	return nil
}

func countActiveTrialLuckyMoneyByMode(db *gorm.DB, gameMode int) (int64, error) {
	var total int64
	err := db.Model(&pojo.TrialLuckyMoney{}).
		Where("status = ? AND game_mode = ?", 1, gameMode).
		Where("(expire_time = ? OR expire_time > ?)", time.Time{}, time.Now()).
		Count(&total).Error
	return total, err
}

func pickRandomTrialSendBotUser(db *gorm.DB) (pojo.TrialBotUser, error) {
	var bot pojo.TrialBotUser
	err := db.Where("status = ?", 1).Order("RAND()").First(&bot).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return pojo.TrialBotUser{}, nil
	}
	return bot, err
}

func pickRandomTrialBotAmount(tablePrefix string) float64 {
	amounts := getBotRandomSendAmounts(tablePrefix)
	return capTrialLuckySendAmount(amounts[rand.IntN(len(amounts))])
}

func ensureTrialBotBalance(db *gorm.DB, bot *pojo.TrialBotUser, minBalance float64) error {
	if db == nil || bot == nil || bot.ID == 0 || minBalance <= 0 || bot.Balance >= minBalance {
		return nil
	}
	topUpAmount := utils.Truncate2(minBalance - bot.Balance)
	if err := db.Model(&pojo.TrialBotUser{}).Where("id = ?", bot.ID).Update("balance", gorm.Expr("balance + ?", topUpAmount)).Error; err != nil {
		return err
	}
	bot.Balance = utils.Truncate2(minBalance)
	return nil
}

func sendTrialRedPacketByBot(db *gorm.DB, bot pojo.TrialBotUser, amount float64, gameMode int, tablePrefix string) (pojo.TrialLuckyMoney, error) {
	amount = capTrialLuckySendAmount(amount)
	if amount < 5 {
		amount = 5
	}
	gameMode = normalizeTrialGameMode(gameMode)
	number := resolveTrialLuckyNumber(db, nil)
	if number <= 0 {
		number = 3
	}
	thunder := 0
	if gameMode == 0 {
		thunder = rand.IntN(10)
	}
	loseRate := GetLoseRate(db)
	if gameMode == 1 {
		loseRate = GetGame2LoseRate(db)
	}
	if loseRate <= 0 {
		loseRate = 1.8
	}
	expireTime := time.Now().Add(time.Duration(GetLuckyExpireMinutes(db)) * time.Minute)
	redList := utils.RedEnvelope(amount, number, 0.01, math.Max(0.01, amount/float64(number)*2))
	redListJSON, _ := json.Marshal(redList)
	name := trialBotDisplayName(bot)
	var lucky pojo.TrialLuckyMoney
	err := db.Transaction(func(tx *gorm.DB) error {
		var lockedBot pojo.TrialBotUser
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND status = ?", bot.ID, 1).First(&lockedBot).Error; err != nil {
			return err
		}
		if lockedBot.Balance < amount {
			return errors.New("trial_bot_balance_insufficient")
		}
		startBalance := utils.Truncate2(lockedBot.Balance)
		endBalance := utils.Truncate2(startBalance - amount)
		lucky = pojo.TrialLuckyMoney{
			SenderID:     lockedBot.ID,
			SenderType:   pojo.TrialActorBot,
			SenderName:   name,
			SenderAvatar: lockedBot.Avatar,
			Amount:       amount,
			Received:     0,
			Number:       number,
			Thunder:      thunder,
			GameMode:     gameMode,
			RedList:      string(redListJSON),
			LoseRate:     loseRate,
			Status:       1,
			TenantId:     lockedBot.TenantId,
			ExpireTime:   expireTime,
		}
		if err := tx.Create(&lucky).Error; err != nil {
			return err
		}
		for i, itemAmount := range redList {
			if err := tx.Create(&pojo.TrialLuckyMoneyItem{
				RedPacketID: lucky.ID,
				SeqNo:       uint(i + 1),
				Amount:      utils.Truncate2(itemAmount),
				IsGrabbed:   0,
				Thunder:     0,
			}).Error; err != nil {
				return err
			}
		}
		if err := tx.Model(&pojo.TrialBotUser{}).Where("id = ?", lockedBot.ID).Update("balance", endBalance).Error; err != nil {
			return err
		}
		return createTrialCashHistory(tx, pojo.TrialCashHistory{
			UserId:      lockedBot.ID,
			ActorType:   pojo.TrialActorBot,
			AwardUni:    fmt.Sprintf("trial_bot_send_%d", lucky.ID),
			Amount:      -amount,
			StartAmount: startBalance,
			EndAmount:   endBalance,
			CashMark:    "trial_lucky_send",
			CashDesc:    "试玩机器人发包",
			Type:        pojo.TrialCashTypeSendLucky,
			LuckyID:     lucky.ID,
			TenantId:    lockedBot.TenantId,
		})
	})
	if err != nil {
		return lucky, err
	}
	BroadcastTrialLuckySent(db, lucky.ID)
	if err := EnqueueTrialLuckyBotGrabTask(db, tablePrefix, lucky.ID, nil, pickRandomTrialBotGrabCount(lucky.Number)); err != nil {
		log.Printf("[trial_lucky] enqueue bot grab failed luckyID=%d err=%v", lucky.ID, err)
	}
	return lucky, nil
}

func trialBotDisplayName(bot pojo.TrialBotUser) string {
	if bot.FirstName != nil && strings.TrimSpace(*bot.FirstName) != "" {
		return strings.TrimSpace(*bot.FirstName)
	}
	if bot.Username != nil && strings.TrimSpace(*bot.Username) != "" {
		return strings.TrimSpace(*bot.Username)
	}
	return fmt.Sprintf("TrialBot_%d", bot.ID)
}

func AutoTrialBotGrab(db *gorm.DB, luckyID int64) error {
	var lucky pojo.TrialLuckyMoney
	if err := db.Where("id = ? AND status = ?", luckyID, 1).First(&lucky).Error; err != nil {
		return nil
	}
	limit := lucky.Number / 2
	if limit <= 0 {
		return nil
	}
	if limit > 3 {
		limit = 3
	}
	var bots []pojo.TrialBotUser
	if err := db.Where("status = ?", 1).Order("RAND()").Limit(limit).Find(&bots).Error; err != nil {
		return err
	}
	for _, bot := range bots {
		if err := grabTrialRedPacketByBot(db, luckyID, bot.ID, 0); err != nil {
			continue
		}
	}
	return nil
}

func grabTrialRedPacketByBot(db *gorm.DB, luckyID int64, botID int64, grabIndex int) error {
	result := map[string]any{}
	finished := false
	err := db.Transaction(func(tx *gorm.DB) error {
		var lucky pojo.TrialLuckyMoney
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", luckyID).First(&lucky).Error; err != nil {
			return err
		}
		if lucky.Status != 1 {
			return errors.New("lucky_finished")
		}
		if !canTrialBotGrabSender(lucky.SenderType, lucky.SenderID, botID) {
			return errors.New("cannot_grab_self")
		}
		var bot pojo.TrialBotUser
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND status = ?", botID, 1).First(&bot).Error; err != nil {
			return err
		}
		var exists int64
		if err := tx.Model(&pojo.TrialLuckyHistory{}).
			Where("lucky_id = ? AND user_id = ? AND actor_type = ? AND grab_type = ?", lucky.ID, bot.ID, pojo.TrialActorBot, 1).
			Count(&exists).Error; err != nil {
			return err
		}
		if exists > 0 {
			return errors.New("already_grabbed")
		}
		requiredBalance := utils.Truncate2(lucky.Amount * lucky.LoseRate)
		if lucky.GameMode == 1 {
			requiredBalance = utils.Truncate2((lucky.Amount / float64(maxInt(1, lucky.Number))) * lucky.LoseRate)
		}
		if bot.Balance < requiredBalance {
			return errors.New("trial_bot_balance_insufficient")
		}
		item, err := pickTrialLuckyItem(tx, lucky.ID, grabIndex)
		if err != nil {
			return err
		}
		guess := rand.IntN(2)
		var guessPtr *int
		if lucky.GameMode == 1 {
			guessPtr = &guess
		}
		now := time.Now()
		openNum := trialOpenNum(item.Amount)
		isThunder := trialIsThunder(lucky, item.Amount, guessPtr)
		loseMoney := 0.0
		if isThunder {
			if lucky.GameMode == 1 {
				loseMoney = utils.Truncate2(item.Amount * lucky.LoseRate)
			} else {
				loseMoney = utils.Truncate2(lucky.Amount * lucky.LoseRate)
			}
		}
		actualAmount := utils.Truncate2(item.Amount - loseMoney)
		startBalance := utils.Truncate2(bot.Balance)
		endBalance := utils.Truncate2(startBalance + actualAmount)
		thunderFlag := int8(0)
		if isThunder {
			thunderFlag = 1
		}
		if err := tx.Model(&pojo.TrialLuckyMoneyItem{}).
			Where("id = ? AND is_grabbed = ?", item.ID, 0).
			Updates(map[string]any{
				"is_grabbed":   1,
				"thunder":      thunderFlag,
				"grabbed_uid":  bot.ID,
				"grabbed_type": pojo.TrialActorBot,
				"grabbed_at":   now,
			}).Error; err != nil {
			return err
		}
		if err := tx.Model(&pojo.TrialBotUser{}).Where("id = ?", bot.ID).Update("balance", endBalance).Error; err != nil {
			return err
		}
		if err := tx.Model(&pojo.TrialLuckyMoney{}).Where("id = ?", lucky.ID).
			Update("received", gorm.Expr("received + ?", item.Amount)).Error; err != nil {
			return err
		}
		name := trialBotDisplayName(bot)
		if err := tx.Create(&pojo.TrialLuckyHistory{
			UserID:       bot.ID,
			ActorType:    pojo.TrialActorBot,
			FirstName:    name,
			LuckyID:      lucky.ID,
			IsThunder:    int(thunderFlag),
			GrabType:     1,
			Amount:       utils.Truncate2(item.Amount),
			ActualAmount: actualAmount,
			LoseMoney:    loseMoney,
			Guess:        guessPtr,
			TenantId:     lucky.TenantId,
		}).Error; err != nil {
			return err
		}
		if err := createTrialCashHistory(tx, pojo.TrialCashHistory{
			UserId:      bot.ID,
			ActorType:   pojo.TrialActorBot,
			AwardUni:    fmt.Sprintf("trial_bot_grab_%d_%d", lucky.ID, bot.ID),
			Amount:      actualAmount,
			StartAmount: startBalance,
			EndAmount:   endBalance,
			CashMark:    "trial_lucky_grab",
			CashDesc:    "试玩机器人抢包",
			Type:        mapTrialGrabCashType(isThunder),
			IsThunder:   thunderFlag,
			LuckyID:     lucky.ID,
			TenantId:    lucky.TenantId,
		}); err != nil {
			return err
		}
		if isThunder && loseMoney > 0 {
			if err := addTrialSenderThunderIncome(tx, lucky, loseMoney); err != nil {
				return err
			}
		}
		var grabbedCount int64
		if err := tx.Model(&pojo.TrialLuckyMoneyItem{}).Where("red_packet_id = ? AND is_grabbed = ?", lucky.ID, 1).Count(&grabbedCount).Error; err != nil {
			return err
		}
		if int(grabbedCount) >= lucky.Number {
			if err := tx.Model(&pojo.TrialLuckyMoney{}).Where("id = ?", lucky.ID).Update("status", 2).Error; err != nil {
				return err
			}
			finished = true
		}
		result = map[string]any{
			"luckyId":      lucky.ID,
			"actorType":    pojo.TrialActorBot,
			"userId":       bot.ID,
			"firstName":    name,
			"grabIndex":    item.SeqNo,
			"amount":       utils.Truncate2(item.Amount),
			"actualAmount": actualAmount,
			"loseMoney":    loseMoney,
			"isThunder":    int(thunderFlag),
			"openNum":      openNum,
			"balance":      endBalance,
			"message":      "success",
		}
		return nil
	})
	if err == nil {
		BroadcastTrialLuckyGrabResult(db, luckyID, result)
		if finished {
			BroadcastTrialLuckyFinished(db, luckyID)
		}
	}
	return err
}

func pickTrialLuckyItem(tx *gorm.DB, luckyID int64, grabIndex int) (pojo.TrialLuckyMoneyItem, error) {
	var item pojo.TrialLuckyMoneyItem
	query := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("red_packet_id = ? AND is_grabbed = ?", luckyID, 0)
	if grabIndex > 0 {
		if err := query.Where("seq_no = ?", grabIndex).First(&item).Error; err != nil {
			return item, errors.New("grab_index_unavailable")
		}
		return item, nil
	}
	if err := query.Order("seq_no asc").First(&item).Error; err != nil {
		return item, errors.New("lucky_empty")
	}
	return item, nil
}

type trialLuckyItemAmountSwap struct {
	PrimaryID     uint
	MatchID       uint
	PrimaryAmount float64
	MatchAmount   float64
}

func pickTrialLuckyItemForUser(tx *gorm.DB, lucky pojo.TrialLuckyMoney, grabIndex int, oddEvenGuess *int, targetWin bool) (pojo.TrialLuckyMoneyItem, error) {
	var items []pojo.TrialLuckyMoneyItem
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("red_packet_id = ? AND is_grabbed = ?", lucky.ID, 0).
		Order("seq_no asc").
		Find(&items).Error; err != nil {
		return pojo.TrialLuckyMoneyItem{}, err
	}
	item, swap, ok := pickTrialLuckyItemForUserTarget(items, lucky, grabIndex, oddEvenGuess, targetWin)
	if !ok {
		if grabIndex > 0 {
			return pojo.TrialLuckyMoneyItem{}, errors.New("grab_index_unavailable")
		}
		return pojo.TrialLuckyMoneyItem{}, errors.New("lucky_empty")
	}
	if swap != nil {
		if err := tx.Model(&pojo.TrialLuckyMoneyItem{}).Where("id = ?", swap.PrimaryID).Update("amount", swap.PrimaryAmount).Error; err != nil {
			return pojo.TrialLuckyMoneyItem{}, err
		}
		if err := tx.Model(&pojo.TrialLuckyMoneyItem{}).Where("id = ?", swap.MatchID).Update("amount", swap.MatchAmount).Error; err != nil {
			return pojo.TrialLuckyMoneyItem{}, err
		}
	}
	return item, nil
}

func pickTrialLuckyItemForUserTarget(items []pojo.TrialLuckyMoneyItem, lucky pojo.TrialLuckyMoney, grabIndex int, oddEvenGuess *int, targetWin bool) (pojo.TrialLuckyMoneyItem, *trialLuckyItemAmountSwap, bool) {
	if len(items) == 0 {
		return pojo.TrialLuckyMoneyItem{}, nil, false
	}
	matchesTarget := func(item pojo.TrialLuckyMoneyItem) bool {
		isThunder := trialIsThunder(lucky, item.Amount, oddEvenGuess)
		return isThunder != targetWin
	}
	if grabIndex > 0 {
		var fixed *pojo.TrialLuckyMoneyItem
		for i := range items {
			if int(items[i].SeqNo) == grabIndex {
				fixed = &items[i]
				break
			}
		}
		if fixed == nil {
			return pojo.TrialLuckyMoneyItem{}, nil, false
		}
		if matchesTarget(*fixed) {
			return *fixed, nil, true
		}
		for _, item := range items {
			if item.ID == fixed.ID {
				continue
			}
			if matchesTarget(item) {
				selected := *fixed
				selected.Amount = item.Amount
				return selected, &trialLuckyItemAmountSwap{
					PrimaryID:     fixed.ID,
					MatchID:       item.ID,
					PrimaryAmount: item.Amount,
					MatchAmount:   fixed.Amount,
				}, true
			}
		}
		return *fixed, nil, true
	}
	for _, item := range items {
		if matchesTarget(item) {
			return item, nil, true
		}
	}
	return items[0], nil, true
}

func trialTargetUserWin(db *gorm.DB) bool {
	return rand.Float64() < GetTrialUserWinRate(db)
}

func GetTrialUserWinRate(db *gorm.DB) float64 {
	ctx := context.Background()
	if utils.RD != nil {
		if cachedValue, err := utils.RD.Get(ctx, trialUserWinRateCache).Result(); err == nil && cachedValue != "" {
			return normalizeTrialUserWinRate(cachedValue)
		}
	}

	defaultValue := strconv.FormatFloat(trialUserWinRateDefault, 'f', 2, 64)
	configValue := getOrInitSysConfigValue(db, trialUserWinRateConfig, defaultValue, "试玩真实用户赢率")
	result := normalizeTrialUserWinRate(configValue)
	if result == trialUserWinRateDefault && strings.TrimSpace(configValue) != defaultValue {
		_ = db.Model(&pojo.SysConfig{}).Where("config_key = ?", trialUserWinRateConfig).Update("config_value", defaultValue).Error
	}
	if utils.RD != nil {
		utils.RD.SetEX(ctx, trialUserWinRateCache, strconv.FormatFloat(result, 'f', 2, 64), utils.GetRandomRangeSecond(20*60, 40*60))
	}
	return result
}

func normalizeTrialUserWinRate(raw string) float64 {
	value, err := strconv.ParseFloat(strings.TrimSpace(raw), 64)
	if err != nil || value < 0 || value > 1 {
		return trialUserWinRateDefault
	}
	return value
}

func canTrialUserGrabSender(senderType string, senderID int64, userID int64) bool {
	return true
}

func canTrialBotGrabSender(senderType string, senderID int64, botID int64) bool {
	return !(senderType == pojo.TrialActorBot && senderID == botID)
}

func addTrialSenderThunderIncome(tx *gorm.DB, lucky pojo.TrialLuckyMoney, amount float64) error {
	if lucky.SenderType == pojo.TrialActorBot {
		var bot pojo.TrialBotUser
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", lucky.SenderID).First(&bot).Error; err != nil {
			return nil
		}
		start := utils.Truncate2(bot.Balance)
		end := utils.Truncate2(start + amount)
		if err := tx.Model(&pojo.TrialBotUser{}).Where("id = ?", bot.ID).Update("balance", end).Error; err != nil {
			return err
		}
		return createTrialCashHistory(tx, pojo.TrialCashHistory{
			UserId:      bot.ID,
			ActorType:   pojo.TrialActorBot,
			AwardUni:    fmt.Sprintf("trial_sender_income_%d_%d", lucky.ID, time.Now().UnixNano()),
			Amount:      amount,
			StartAmount: start,
			EndAmount:   end,
			CashMark:    "trial_lucky_thunder_income",
			CashDesc:    "试玩发包中雷收益",
			Type:        pojo.TrialCashTypeThunderIncome,
			IsThunder:   1,
			LuckyID:     lucky.ID,
			TenantId:    lucky.TenantId,
		})
	}
	var sender pojo.TgUser
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", lucky.SenderID).First(&sender).Error; err != nil {
		return nil
	}
	start := utils.Truncate2(sender.TrialBalance)
	end := utils.Truncate2(start + amount)
	if err := tx.Model(&pojo.TgUser{}).Where("id = ?", sender.ID).Update("trial_balance", end).Error; err != nil {
		return err
	}
	return createTrialCashHistory(tx, pojo.TrialCashHistory{
		UserId:      sender.ID,
		ActorType:   pojo.TrialActorUser,
		AwardUni:    fmt.Sprintf("trial_sender_income_%d_%d", lucky.ID, time.Now().UnixNano()),
		Amount:      amount,
		StartAmount: start,
		EndAmount:   end,
		CashMark:    "trial_lucky_thunder_income",
		CashDesc:    "试玩发包中雷收益",
		Type:        pojo.TrialCashTypeThunderIncome,
		IsThunder:   1,
		LuckyID:     lucky.ID,
		TenantId:    lucky.TenantId,
	})
}

func refundExpiredTrialLucky(tx *gorm.DB, lucky *pojo.TrialLuckyMoney) error {
	var grabbed float64
	if err := tx.Model(&pojo.TrialLuckyMoneyItem{}).Where("red_packet_id = ? AND is_grabbed = ?", lucky.ID, 1).
		Select("COALESCE(SUM(amount), 0)").Scan(&grabbed).Error; err != nil {
		return err
	}
	refund := utils.Truncate2(lucky.Amount - grabbed)
	if refund > 0 && lucky.SenderType == pojo.TrialActorUser {
		var sender pojo.TgUser
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", lucky.SenderID).First(&sender).Error; err == nil {
			start := utils.Truncate2(sender.TrialBalance)
			end := utils.Truncate2(start + refund)
			if err := tx.Model(&pojo.TgUser{}).Where("id = ?", sender.ID).Update("trial_balance", end).Error; err != nil {
				return err
			}
			if err := createTrialCashHistory(tx, pojo.TrialCashHistory{
				UserId:      sender.ID,
				ActorType:   pojo.TrialActorUser,
				AwardUni:    fmt.Sprintf("trial_refund_%d", lucky.ID),
				Amount:      refund,
				StartAmount: start,
				EndAmount:   end,
				CashMark:    "trial_lucky_refund",
				CashDesc:    "试玩红包过期退回",
				Type:        pojo.TrialCashTypeExpireRefund,
				LuckyID:     lucky.ID,
				TenantId:    lucky.TenantId,
			}); err != nil {
				return err
			}
		}
	}
	if refund > 0 && lucky.SenderType == pojo.TrialActorBot {
		var bot pojo.TrialBotUser
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", lucky.SenderID).First(&bot).Error; err == nil {
			start := utils.Truncate2(bot.Balance)
			end := utils.Truncate2(start + refund)
			if err := tx.Model(&pojo.TrialBotUser{}).Where("id = ?", bot.ID).Update("balance", end).Error; err != nil {
				return err
			}
			if err := createTrialCashHistory(tx, pojo.TrialCashHistory{
				UserId:      bot.ID,
				ActorType:   pojo.TrialActorBot,
				AwardUni:    fmt.Sprintf("trial_bot_refund_%d", lucky.ID),
				Amount:      refund,
				StartAmount: start,
				EndAmount:   end,
				CashMark:    "trial_lucky_refund",
				CashDesc:    "试玩机器人红包过期退回",
				Type:        pojo.TrialCashTypeExpireRefund,
				LuckyID:     lucky.ID,
				TenantId:    lucky.TenantId,
			}); err != nil {
				return err
			}
		}
	}
	lucky.Status = 3
	return tx.Model(&pojo.TrialLuckyMoney{}).Where("id = ?", lucky.ID).Update("status", 3).Error
}

func refundExpiredTrialLuckyByID(db *gorm.DB, luckyID int64) error {
	return db.Transaction(func(tx *gorm.DB) error {
		var lucky pojo.TrialLuckyMoney
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", luckyID).First(&lucky).Error; err != nil {
			return err
		}
		if lucky.Status != 1 {
			return nil
		}
		if !lucky.ExpireTime.IsZero() && time.Now().Before(lucky.ExpireTime) {
			return nil
		}
		return refundExpiredTrialLucky(tx, &lucky)
	})
}

func createTrialCashHistory(tx *gorm.DB, item pojo.TrialCashHistory) error {
	return nil
}

func trialTgDisplayName(user pojo.TgUser) string {
	if user.FirstName != nil && strings.TrimSpace(*user.FirstName) != "" {
		return strings.TrimSpace(*user.FirstName)
	}
	if user.Username != nil && strings.TrimSpace(*user.Username) != "" {
		return strings.TrimSpace(*user.Username)
	}
	return fmt.Sprintf("User_%d", user.ID)
}

func resolveTrialLuckyNumber(db *gorm.DB, requested *int) int {
	if requested != nil && *requested > 0 {
		return *requested
	}
	config := strings.TrimSpace(GetLuckyNumConfig(db))
	fields := strings.FieldsFunc(config, func(r rune) bool {
		return r < '0' || r > '9'
	})
	for _, field := range fields {
		if n, err := strconv.Atoi(field); err == nil && n > 0 {
			return n
		}
	}
	return 3
}

func normalizeTrialGameMode(gameMode int) int {
	if gameMode == 1 {
		return 1
	}
	return 0
}

func trialOpenNum(amount float64) int {
	units := int(math.Round(amount * 100))
	return units % 10
}

func trialIsThunder(lucky pojo.TrialLuckyMoney, amount float64, oddEvenGuess *int) bool {
	openNum := trialOpenNum(amount)
	if lucky.GameMode == 1 {
		if oddEvenGuess == nil {
			return false
		}
		actual := openNum % 2
		return actual != *oddEvenGuess
	}
	return openNum == lucky.Thunder
}

func mapTrialGrabCashType(isThunder bool) int8 {
	if isThunder {
		return pojo.TrialCashTypeGrabLuckyLose
	}
	return pojo.TrialCashTypeGrabLuckyWin
}

func formatTrialRemainingText(seconds int64) string {
	if seconds <= 0 {
		return "已结束"
	}
	if seconds < 60 {
		return fmt.Sprintf("%d秒", seconds)
	}
	return fmt.Sprintf("%d分%d秒", seconds/60, seconds%60)
}

func maxInt(a int, b int) int {
	if a > b {
		return a
	}
	return b
}
