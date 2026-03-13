package services

import (
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/repository"
	"BaseGoUni/core/utils"
	"context"
	"errors"
	"fmt"
	"github.com/jinzhu/copier"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"log"
	"math/rand/v2"
	"strconv"
	"strings"
	"sync"
	"time"
)

var autoLuckyMaintainMu sync.Mutex

// SendRedPacket 发送红包业务逻辑
func SendRedPacket(db *gorm.DB, senderID int64, senderName string, req pojo.LuckyMoneySend, tablePrefix string) (*pojo.LuckyMoney, error) {
	return sendRedPacket(db, senderID, senderName, req, tablePrefix)
}

func sendRedPacket(db *gorm.DB, senderID int64, senderName string, req pojo.LuckyMoneySend, tablePrefix string) (*pojo.LuckyMoney, error) {
	// 验证金额
	if req.Amount < 5 {
		return nil, errors.New("红包金额不能小于5U")
	}

	// 验证雷号
	if req.Thunder < 0 || req.Thunder > 9 {
		return nil, errors.New("指令有误，雷数应是0~9之间")
	}

	// 获取红包数量配置
	luckyNumConfig := GetLuckyNumConfig(db)
	if luckyNumConfig == "" {
		return nil, errors.New("配置错误：未找到红包数量配置")
	}

	// 确定红包数量
	var luckyTotal int
	if req.Number != nil {
		// 验证用户指定的数量
		valid, msg := utils.ValidateLuckyCount(*req.Number, luckyNumConfig)
		if !valid {
			return nil, errors.New(msg)
		}
		luckyTotal = *req.Number
	} else {
		// 从配置项获取最小值
		luckyTotal = utils.GetLuckyNumMin(luckyNumConfig)
	}

	// 检查余额
	user, err := getTgUserByID(db, senderID)
	if err != nil {
		return nil, err
	}
	if user.Status != 1 {
		return nil, errors.New("用户已禁用，请联系管理员处理")
	}
	if user.Balance < req.Amount {
		return nil, errors.New("您的余额已不足发包")
	}

	// 生成红包金额数组
	minAmount := 0.01
	maxAmount := req.Amount / float64(luckyTotal) * 2
	redEnvelopes := utils.RedEnvelope(req.Amount, luckyTotal, minAmount, maxAmount)

	// 获取中雷倍数
	loseRate := GetLoseRate(db)
	// 获取红包过期分钟配置
	expireMinutes := GetLuckyExpireMinutes(db)
	expireAt := time.Now().Add(time.Duration(expireMinutes) * time.Minute)

	// 创建红包记录
	luckyMoney := &pojo.LuckyMoney{
		SenderID:   senderID,
		SenderName: senderName,
		Amount:     req.Amount,
		Number:     luckyTotal,
		Thunder:    req.Thunder,
		ChatID:     req.ChatID,
		RedList:    "",
		LoseRate:   loseRate,
		Status:     1, // 进行中
		Lucky:      1, // 有效
		Received:   0,
		ExpireTime: expireAt,
	}

	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 创建红包
	if err := repository.CreateLuckyMoney(tx, luckyMoney); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("创建红包失败: %v", err)
	}
	if err := repository.CreateLuckyMoneyItems(tx, luckyMoney.ID, redEnvelopes); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("创建红包明细失败: %v", err)
	}

	// 锁定发送者行，重新校验余额（防止并发发包导致余额为负）
	var lockedSender pojo.TgUser
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", senderID).First(&lockedSender).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("锁定用户失败: %v", err)
	}
	if lockedSender.Balance < req.Amount {
		tx.Rollback()
		return nil, errors.New("您的余额已不足发包")
	}

	// 扣除发送者余额，同时扣除赠送余额（不低于0）
	giftDeduct := req.Amount
	if lockedSender.GiftAmount < giftDeduct {
		giftDeduct = lockedSender.GiftAmount
	}
	if err := tx.Model(&pojo.TgUser{}).
		Where("id = ?", senderID).
		Updates(map[string]interface{}{
			"balance":     gorm.Expr("balance - ?", req.Amount),
			"gift_amount": gorm.Expr("gift_amount - ?", giftDeduct),
		}).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("扣除余额失败: %v", err)
	}

	// 记录余额变动
	normalDeduct := req.Amount - giftDeduct
	awardUniBase := fmt.Sprintf("lucky_%d", luckyMoney.ID)
	runningBalance := lockedSender.Balance

	if giftDeduct > 0 {
		cashHistoryGift := pojo.CashHistory{
			UserId:      senderID,
			AwardUni:    awardUniBase + "_gift",
			Amount:      -giftDeduct,
			StartAmount: runningBalance,
			EndAmount:   runningBalance - giftDeduct,
			CashMark:    "发送红包",
			CashDesc:    fmt.Sprintf("发送红包 %dU（赠送金额%.2fU），雷号%d", int(req.Amount), giftDeduct, req.Thunder),
			Type:        pojo.CashHistoryTypeSendRedPacket,
			IsGift:      1,
			FromUserId:  senderID,
		}
		if err := tx.Create(&cashHistoryGift).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("记录余额变动失败: %v", err)
		}
		runningBalance -= giftDeduct
	}

	if normalDeduct > 0 {
		cashHistoryCash := pojo.CashHistory{
			UserId:      senderID,
			AwardUni:    awardUniBase + "_cash",
			Amount:      -normalDeduct,
			StartAmount: runningBalance,
			EndAmount:   runningBalance - normalDeduct,
			CashMark:    "发送红包",
			CashDesc:    fmt.Sprintf("发送红包 %dU（正常金额%.2fU），雷号%d", int(req.Amount), normalDeduct, req.Thunder),
			Type:        pojo.CashHistoryTypeSendRedPacket,
			IsGift:      0,
			FromUserId:  senderID,
		}
		if err := tx.Create(&cashHistoryCash).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("记录余额变动失败: %v", err)
		}
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("提交事务失败: %v", err)
	}
	if err := EnqueueLuckyExpireTask(tablePrefix, luckyMoney.ID, luckyMoney.ExpireTime); err != nil {
		log.Printf("[lucky] EnqueueLuckyExpireTask failed: luckyID=%d err=%v", luckyMoney.ID, err)
	}
	if err := EnqueueLuckyBotGrabTask(db, tablePrefix, luckyMoney.ID, nil, pickRandomBotGrabCount(luckyMoney.Number)); err != nil {
		log.Printf("[lucky] EnqueueLuckyBotGrabTask failed: luckyID=%d err=%v", luckyMoney.ID, err)
	}

	return luckyMoney, nil
}

func EnsureMinActiveLuckyPackets(db *gorm.DB, tablePrefix string) error {
	if db == nil {
		return nil
	}

	autoLuckyMaintainMu.Lock()
	defer autoLuckyMaintainMu.Unlock()

	for i := 0; i < 3; i++ {
		activeCount, err := countActiveLuckyMoney(db)
		if err != nil {
			return err
		}
		if activeCount >= 3 {
			return nil
		}

		botUser, err := pickRandomLuckyBotUser(db)
		if err != nil {
			return err
		}
		if botUser.ID == 0 {
			return nil
		}

		amount := pickRandomLuckyBotAmount(tablePrefix)
		if err = ensureBotBalance(db, &botUser, amount); err != nil {
			return err
		}

		req := pojo.LuckyMoneySend{
			Amount:  amount,
			Thunder: rand.IntN(10),
		}
		luckyMoney, sendErr := sendRedPacket(db, botUser.ID, getTgUserDisplayName(&botUser), req, tablePrefix)
		if sendErr != nil {
			err = sendErr
			return err
		}
		broadcastLuckySent(luckyMoney)
	}

	return nil
}

func broadcastLuckySent(luckyMoney *pojo.LuckyMoney) {
	if luckyMoney == nil || luckyMoney.ID <= 0 {
		return
	}

	var result pojo.LuckyMoneyBack
	_ = copier.Copy(&result, luckyMoney)
	if err := utils.BroadcastWsWithType("lucky_sent", result); err != nil {
		log.Printf("[lucky] Broadcast lucky_sent failed: luckyID=%d err=%v", luckyMoney.ID, err)
	}
}

func countActiveLuckyMoney(db *gorm.DB) (int64, error) {
	var total int64
	err := db.Model(&pojo.LuckyMoney{}).Where("status = ?", 1).Count(&total).Error
	return total, err
}

func pickRandomLuckyBotUser(db *gorm.DB) (pojo.TgUser, error) {
	var botUser pojo.TgUser
	err := db.Where("is_bot = ? AND status = ?", true, 1).Order("RAND()").First(&botUser).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return pojo.TgUser{}, nil
	}
	return botUser, err
}

func pickRandomLuckyBotAmount(tablePrefix string) float64 {
	amounts := getBotRandomSendAmounts(tablePrefix)
	return amounts[rand.IntN(len(amounts))]
}

func getBotRandomSendAmounts(tablePrefix string) []float64 {
	defaultValue := "100|200|500"
	val := utils.GetStringCache(tablePrefix, "bot_rsend_random", &defaultValue)
	raw := defaultValue
	if val != nil && strings.TrimSpace(*val) != "" {
		raw = strings.TrimSpace(*val)
	}

	parts := strings.Split(raw, "|")
	result := make([]float64, 0, len(parts))
	for _, part := range parts {
		amount, err := strconv.ParseFloat(strings.TrimSpace(part), 64)
		if err != nil || amount <= 0 {
			continue
		}
		result = append(result, amount)
	}
	if len(result) == 0 {
		return []float64{100, 200, 500}
	}
	return result
}

func topUpBotBalance(db *gorm.DB, userID int64, amount float64) error {
	if db == nil || userID <= 0 || amount <= 0 {
		return nil
	}
	return db.Model(&pojo.TgUser{}).
		Where("id = ? AND is_bot = ?", userID, true).
		Update("balance", gorm.Expr("balance + ?", amount)).Error
}

func ensureBotBalance(db *gorm.DB, botUser *pojo.TgUser, minBalance float64) error {
	if db == nil || botUser == nil || botUser.ID == 0 || minBalance <= 0 {
		return nil
	}
	if botUser.Balance >= minBalance {
		return nil
	}
	topUpAmount := minBalance - botUser.Balance
	if err := topUpBotBalance(db, botUser.ID, topUpAmount); err != nil {
		return err
	}
	botUser.Balance = minBalance
	return nil
}

func BroadcastLuckyGrabResult(db *gorm.DB, luckyID int64, result map[string]interface{}) error {
	if db == nil || luckyID <= 0 {
		return nil
	}

	luckyMoney, err := repository.GetLuckyMoney(db, luckyID)
	if err != nil || luckyMoney.ID == 0 {
		return nil
	}

	broadcast := map[string]interface{}{
		"id":         luckyMoney.ID,
		"createdAt":  luckyMoney.CreatedAt,
		"updatedAt":  luckyMoney.UpdatedAt,
		"senderId":   luckyMoney.SenderID,
		"senderName": luckyMoney.SenderName,
		"amount":     luckyMoney.Amount,
		"received":   luckyMoney.Received,
		"number":     luckyMoney.Number,
		"lucky":      luckyMoney.Lucky,
		"thunder":    luckyMoney.Thunder,
		"chatId":     luckyMoney.ChatID,
		"redList":    luckyMoney.RedList,
		"loseRate":   luckyMoney.LoseRate,
		"status":     luckyMoney.Status,
		"tenantId":   luckyMoney.TenantId,
		"expireTime": luckyMoney.ExpireTime,
	}
	if idx, ok := result["grabIndex"]; ok {
		broadcast["grabIndex"] = idx
	} else if idx, ok := result["openNum"]; ok {
		broadcast["grabIndex"] = idx
	}

	grabbedCount, _ := repository.GetLuckyHistoryCount(db, luckyID)
	hideSecondLast := shouldHideSecondLastInProgressForBroadcast(luckyMoney, grabbedCount, time.Now())

	grabAmount := toFloat64Value(result["amount"])
	if hideSecondLast {
		grabAmount = 0
	}
	broadcast["grabAmount"] = grabAmount

	grabSeqNo := toIntValue(result["grabIndex"])
	if grabSeqNo <= 0 {
		grabSeqNo = toIntValue(result["openNum"])
	}
	// 单次查询同时获取当前包的 thunder_amount 和全局 SUM(thunder_amount)
	var thunderRow struct {
		ThisThunderAmount  float64 `gorm:"column:this_thunder_amount"`
		TotalThunderAmount float64 `gorm:"column:total_thunder_amount"`
	}
	_ = db.Table("lucky_money_item").
		Select("COALESCE(SUM(CASE WHEN seq_no = ? THEN thunder_amount ELSE 0 END), 0) as this_thunder_amount, COALESCE(SUM(thunder_amount), 0) as total_thunder_amount", grabSeqNo).
		Where("red_packet_id = ?", luckyID).
		Scan(&thunderRow).Error
	rawThunderAmount := thunderRow.ThisThunderAmount
	thunderAmount := rawThunderAmount
	if hideSecondLast {
		thunderAmount = 0
	}
	broadcast["thunderAmount"] = thunderAmount

	if isThunder, ok := result["isThunder"]; ok {
		broadcast["isThunder"] = isThunder
	}
	if loseMoney, ok := result["loseMoney"]; ok {
		broadcast["loseMoney"] = loseMoney
	}
	totalThunderAmount := thunderRow.TotalThunderAmount
	if hideSecondLast {
		totalThunderAmount -= rawThunderAmount
		if totalThunderAmount < 0 {
			totalThunderAmount = 0
		}
	}
	broadcast["totalThunderAmount"] = totalThunderAmount

	return utils.BroadcastWsWithType("lucky_grabbed", broadcast)
}

func shouldHideSecondLastInProgressForBroadcast(lucky pojo.LuckyMoney, grabbedCount int64, now time.Time) bool {
	if lucky.Status != 1 {
		return false
	}
	if lucky.Number <= 1 {
		return false
	}
	if int(grabbedCount) != lucky.Number-1 {
		return false
	}
	expireAt := lucky.ExpireTime
	if expireAt.IsZero() {
		expireAt = lucky.CreatedAt.Add(3 * time.Minute)
	}
	return now.Before(expireAt)
}

func toIntValue(v interface{}) int {
	switch n := v.(type) {
	case int:
		return n
	case int8:
		return int(n)
	case int16:
		return int(n)
	case int32:
		return int(n)
	case int64:
		return int(n)
	case uint:
		return int(n)
	case uint8:
		return int(n)
	case uint16:
		return int(n)
	case uint32:
		return int(n)
	case uint64:
		return int(n)
	case float32:
		return int(n)
	case float64:
		return int(n)
	default:
		return 0
	}
}

func toFloat64Value(v interface{}) float64 {
	switch n := v.(type) {
	case float64:
		return n
	case float32:
		return float64(n)
	case int:
		return float64(n)
	case int8:
		return float64(n)
	case int16:
		return float64(n)
	case int32:
		return float64(n)
	case int64:
		return float64(n)
	case uint:
		return float64(n)
	case uint8:
		return float64(n)
	case uint16:
		return float64(n)
	case uint32:
		return float64(n)
	case uint64:
		return float64(n)
	default:
		return 0
	}
}

// GrabRedPacket 抢红包业务逻辑
func GrabRedPacket(db *gorm.DB, luckyID int64, userID int64, tablePrefix string, grabIndex int) (map[string]interface{}, error) {
	// 获取红包信息
	luckyMoney, err := repository.GetLuckyMoney(db, luckyID)
	if err != nil {
		return nil, errors.New("数据有误！红包不存在")
	}

	// 检查红包状态
	if luckyMoney.Status != 1 {
		return nil, errors.New("该红包已结束")
	}

	// 获取用户信息
	user, err := getTgUserByID(db, userID)
	if err != nil {
		return nil, errors.New("用户未注册，请先注册")
	}
	if user.Status != 1 {
		return nil, errors.New("用户已禁用，请联系管理员处理")
	}

	// 检查余额（需满足 lose_rate 倍数）
	lowestAmount := luckyMoney.Amount * luckyMoney.LoseRate
	if user.Balance < lowestAmount {
		return nil, fmt.Errorf("你至少需要有%.2fU才能抢这个红包~", lowestAmount)
	}

	// 检查是否已领取
	//grabbed, err := repository.CheckUserGrabbed(db, luckyID, userID)
	//if err != nil {
	//	return nil, fmt.Errorf("检查领取状态失败: %v", err)
	//}
	//if grabbed {
	//	return nil, errors.New("您已领取该红包，无法再领取")
	//}

	// 获取已领取数量
	grabbedCount, err := repository.GetLuckyHistoryCount(db, luckyID)
	if err != nil {
		return nil, fmt.Errorf("获取领取数量失败: %v", err)
	}

	if int(grabbedCount) >= luckyMoney.Number {
		return nil, errors.New("该红包已全部被领取")
	}

	// 获取红包金额列表
	redList, err := repository.GetLuckyMoneyRedList(db, luckyID)
	if err != nil {
		return nil, fmt.Errorf("获取红包列表失败: %v", err)
	}

	if len(redList) == 0 {
		return nil, errors.New("红包数据异常")
	}

	// 选择指定包（1-based）
	if grabIndex <= 0 {
		grabIndex = int(grabbedCount) + 1
	}
	if grabIndex > len(redList) {
		return nil, errors.New("红包数据异常")
	}
	redAmount := redList[grabIndex-1]
	awardTs := time.Now().Unix()
	redAmountMilli := int64(redAmount * 1000)

	// 判断是否中雷
	isThunder := int8(0)
	loseMoney := 0.0
	thunderFee := 0.0 // 中雷手续费（发包者端抽成）
	winFee := 0.0     // 中奖手续费（抢包者端抽成）
	sendCommission := 0
	grabbingCommission := 0
	amountStr := fmt.Sprintf("%.2f", redAmount)
	lastDigit := amountStr[len(amountStr)-1]
	thunderStr := strconv.Itoa(luckyMoney.Thunder)
	hitThunder := string(lastDigit) == thunderStr
	if hitThunder {
		isThunder = 1
		loseMoney = luckyMoney.Amount * luckyMoney.LoseRate
		sendCommission = GetSendCommission(db)
		thunderFee = loseMoney * float64(sendCommission) / 100.0
	} else {
		grabbingCommission = GetGrabbingCommission(db)
		winFee = redAmount * float64(grabbingCommission) / 100.0
	}

	lockKey := fmt.Sprintf("lucky_grab:%d", luckyID)
	acquired, lockErr := utils.AcquireLock(lockKey, 10*time.Second)
	if lockErr != nil || !acquired {
		return nil, errors.New("操作频繁，请稍后再试")
	}
	defer utils.ReleaseLock(lockKey)

	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 标记子红包被抢（防止同一序号被重复抢），并记录是否中雷
	if err := repository.MarkLuckyMoneyItemGrabbed(tx, luckyID, grabIndex, userID, isThunder, loseMoney, thunderFee, winFee, time.Now()); err != nil {
		tx.Rollback()
		return nil, err
	}

	if hitThunder {
		// 中雷
		// 锁定抢包者行，重新校验余额（防止并发抢不同红包导致余额为负）
		var lockedUser pojo.TgUser
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", userID).First(&lockedUser).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("锁定用户失败: %v", err)
		}
		if lockedUser.Balance < lowestAmount {
			tx.Rollback()
			return nil, fmt.Errorf("你至少需要有%.2fU才能抢这个红包~", lowestAmount)
		}
		// 锁定发送者行（在 UPDATE 前读取，得到准确的 StartAmount）
		var lockedSender pojo.TgUser
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", luckyMoney.SenderID).First(&lockedSender).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("锁定发送者失败: %v", err)
		}

		// 扣除用户余额
		giftDeduct := loseMoney
		if lockedUser.GiftAmount < giftDeduct {
			giftDeduct = lockedUser.GiftAmount
		}
		if err := tx.Model(&pojo.TgUser{}).
			Where("id = ?", userID).
			Updates(map[string]interface{}{
				"balance":     gorm.Expr("balance - ?", loseMoney),
				"gift_amount": gorm.Expr("gift_amount - ?", giftDeduct),
			}).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("扣除余额失败: %v", err)
		}

		// 中雷手续费
		commissionAmount := thunderFee
		// 实际到账金额 = 显示金额 - 抽成金额
		actualLoseMoney := loseMoney - commissionAmount

		// 增加发送者余额（实际到账金额）
		if err := tx.Model(&pojo.TgUser{}).
			Where("id = ?", luckyMoney.SenderID).
			Update("balance", gorm.Expr("balance + ?", actualLoseMoney)).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("增加余额失败: %v", err)
		}

		// 记录余额变动（用户）
		normalDeduct := loseMoney - giftDeduct
		awardUniBase := fmt.Sprintf("lucky_grab_%d_%d_%d_%d", luckyID, userID, awardTs, int64(loseMoney*1000))
		runningBalance := lockedUser.Balance

		if giftDeduct > 0 {
			cashHistoryGift := pojo.CashHistory{
				UserId:      userID,
				AwardUni:    awardUniBase + "_gift",
				Amount:      -giftDeduct,
				StartAmount: runningBalance,
				EndAmount:   runningBalance - giftDeduct,
				CashMark:    "抢红包中雷",
				CashDesc:    fmt.Sprintf("抢红包中雷，损失%.2fU（赠送金额%.2fU）", loseMoney, giftDeduct),
				Type:        pojo.CashHistoryTypeGrabRedPacketThunder,
				IsGift:      1,
				FromUserId:  luckyMoney.SenderID,
			}
			if err := tx.Create(&cashHistoryGift).Error; err != nil {
				tx.Rollback()
				return nil, fmt.Errorf("记录余额变动失败: %v", err)
			}
			runningBalance -= giftDeduct
		}

		if normalDeduct > 0 {
			cashHistoryCash := pojo.CashHistory{
				UserId:      userID,
				AwardUni:    awardUniBase + "_cash",
				Amount:      -normalDeduct,
				StartAmount: runningBalance,
				EndAmount:   runningBalance - normalDeduct,
				CashMark:    "抢红包中雷",
				CashDesc:    fmt.Sprintf("抢红包中雷，损失%.2fU（正常金额%.2fU）", loseMoney, normalDeduct),
				Type:        pojo.CashHistoryTypeGrabRedPacketThunder,
				IsGift:      0,
				FromUserId:  luckyMoney.SenderID,
			}
			if err := tx.Create(&cashHistoryCash).Error; err != nil {
				tx.Rollback()
				return nil, fmt.Errorf("记录余额变动失败: %v", err)
			}
		}

		// 记录余额变动（发送者）- 抽成后金额（实际到账）
		// 使用事务内锁定的 lockedSender（UPDATE 前的值，即准确的 StartAmount）
		cashHistorySender := pojo.CashHistory{
			UserId:      luckyMoney.SenderID,
			AwardUni:    fmt.Sprintf("lucky_thunder_%d_%d_%d_%d", luckyID, userID, awardTs, int64(actualLoseMoney*1000)),
			Amount:      actualLoseMoney,
			StartAmount: lockedSender.Balance,
			EndAmount:   lockedSender.Balance + actualLoseMoney, // 实际到账金额
			CashMark:    "红包中雷收益",
			CashDesc:    fmt.Sprintf("红包中雷收益，获得%.2f", actualLoseMoney),
			Type:        pojo.CashHistoryTypeRedPacketThunderIncome,
			IsGift:      0,
			FromUserId:  userID,
		}
		if err := tx.Create(&cashHistorySender).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("记录余额变动失败: %v", err)
		}

		// 记录余额变动（发送者）- 抽成金额
		if commissionAmount > 0 {
			cashHistoryCommission := pojo.CashHistory{
				UserId:      luckyMoney.SenderID,
				AwardUni:    fmt.Sprintf("lucky_thunder_commission_%d_%d_%d_%d", luckyID, userID, awardTs, int64(commissionAmount*1000)),
				Amount:      commissionAmount,
				StartAmount: lockedSender.Balance + actualLoseMoney, // 抽成后金额
				EndAmount:   lockedSender.Balance + actualLoseMoney, // 抽成后金额（不变）
				CashMark:    "红包中雷抽成",
				CashDesc:    fmt.Sprintf("红包中雷抽成%.2f%%，抽成金额%.2fU", float64(sendCommission), commissionAmount),
				Type:        pojo.CashHistoryTypeRedPacketCommission,
				IsGift:      0,
				FromUserId:  userID,
			}
			if err := tx.Create(&cashHistoryCommission).Error; err != nil {
				tx.Rollback()
				return nil, fmt.Errorf("记录余额变动失败: %v", err)
			}
			if err := repository.CreatePlatformProfitLedgerIfAbsent(tx, pojo.PlatformProfitLedger{
				TenantId:      luckyMoney.TenantId,
				UserId:        luckyMoney.SenderID,
				SourceType:    pojo.PlatformProfitSourceLuckyThunderCommission,
				SourceId:      cashHistoryCommission.AwardUni,
				IncomeAmount:  commissionAmount,
				ExpenseAmount: 0,
				Remark:        cashHistoryCommission.CashDesc,
			}); err != nil {
				tx.Rollback()
				return nil, fmt.Errorf("记录平台盈利流水失败: %v", err)
			}

			// 中雷时：给发包用户的上级按平台收益返佣
			if err := applyLuckyInviteRebate(tx, tablePrefix, luckyMoney.SenderID, luckyMoney.TenantId, commissionAmount, getInviteThunderRebateRate(tablePrefix), luckyID, grabIndex, "thunder"); err != nil {
				tx.Rollback()
				return nil, fmt.Errorf("发包上级返佣失败: %v", err)
			}
		}
	} else {
		// 未中雷，增加用户余额（需要扣除抽成）
		// 中奖手续费
		commissionAmount := winFee
		// 实际到账金额 = 显示金额 - 抽成金额
		actualAmount := redAmount - commissionAmount

		// 增加用户余额（实际到账金额）
		if err := tx.Model(&pojo.TgUser{}).
			Where("id = ?", userID).
			Update("balance", gorm.Expr("balance + ?", actualAmount)).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("增加余额失败: %v", err)
		}

		// 记录余额变动 - 抽成后金额（实际到账）
		cashHistoryGrab := pojo.CashHistory{
			UserId:      userID,
			AwardUni:    fmt.Sprintf("lucky_grab_%d_%d_%d_%s", luckyID, userID, redAmountMilli, utils.RandomString(6)),
			Amount:      actualAmount,
			StartAmount: user.Balance,
			EndAmount:   user.Balance + actualAmount, // 实际到账金额
			CashMark:    "抢红包",
			CashDesc:    fmt.Sprintf("抢红包，获得%.2fU", actualAmount),
			Type:        pojo.CashHistoryTypeGrabRedPacketWin,
			IsGift:      0,
			FromUserId:  luckyMoney.SenderID,
		}
		if err := tx.Create(&cashHistoryGrab).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("记录余额变动失败: %v", err)
		}

		// 记录余额变动 - 抽成金额
		if commissionAmount > 0 {
			cashHistoryCommission := pojo.CashHistory{
				UserId:      userID,
				AwardUni:    fmt.Sprintf("l_grab_%d_%d_%d_%s", luckyID, userID, awardTs, utils.RandomString(6)),
				Amount:      commissionAmount,
				StartAmount: user.Balance + actualAmount, // 抽成后金额
				EndAmount:   user.Balance + actualAmount, // 抽成后金额（不变）
				CashMark:    "抢红包抽成",
				CashDesc:    fmt.Sprintf("抢红包抽成%.2f%%，抽成金额%.2fU", float64(grabbingCommission), commissionAmount),
				Type:        pojo.CashHistoryTypeRedPacketCommission,
				IsGift:      0,
				FromUserId:  luckyMoney.SenderID,
			}
			if err := tx.Create(&cashHistoryCommission).Error; err != nil {
				tx.Rollback()
				return nil, fmt.Errorf("记录余额变动失败: %v", err)
			}
			if err := repository.CreatePlatformProfitLedgerIfAbsent(tx, pojo.PlatformProfitLedger{
				TenantId:      luckyMoney.TenantId,
				UserId:        userID,
				SourceType:    pojo.PlatformProfitSourceLuckyGrabCommission,
				SourceId:      cashHistoryCommission.AwardUni,
				IncomeAmount:  commissionAmount,
				ExpenseAmount: 0,
				Remark:        cashHistoryCommission.CashDesc,
			}); err != nil {
				tx.Rollback()
				return nil, fmt.Errorf("记录平台盈利流水失败: %v", err)
			}

			// 中奖时：给中奖用户的上级按平台收益返佣
			if err := applyLuckyInviteRebate(tx, tablePrefix, userID, luckyMoney.TenantId, commissionAmount, getInviteLuckyRebateRate(tablePrefix), luckyID, grabIndex, "lucky"); err != nil {
				tx.Rollback()
				return nil, fmt.Errorf("中奖上级返佣失败: %v", err)
			}
		}
	}

	// 创建领取记录
	history := &pojo.LuckyHistory{
		UserID:    userID,
		FirstName: utils.FormatName(getTgUserDisplayName(&user), 8),
		LuckyID:   luckyID,
		IsThunder: int(isThunder),
		Amount:    redAmount,
		LoseMoney: loseMoney,
	}
	if err := repository.CreateLuckyHistory(tx, history); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("创建领取记录失败: %v", err)
	}

	// 更新红包已领取金额
	newReceived := luckyMoney.Received + redAmount
	updates := map[string]interface{}{
		"received": newReceived,
	}

	// 检查是否全部领取完成
	if int(grabbedCount+1) >= luckyMoney.Number {
		updates["status"] = 2 // 已完成
	}

	if err := repository.UpdateLuckyMoney(tx, luckyID, updates); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("更新红包状态失败: %v", err)
	}

	completedAfterGrab := int(grabbedCount+1) >= luckyMoney.Number
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("提交事务失败: %v", err)
	}
	if completedAfterGrab {
		_ = EnsureMinActiveLuckyPackets(db, tablePrefix)
	}

	displayAmount := redAmount
	displayLoseMoney := loseMoney
	if shouldHideSecondLastInProgressForBroadcast(luckyMoney, grabbedCount+1, time.Now()) {
		displayAmount = 0
		displayLoseMoney = 0
	}

	// 返回结果
	result := map[string]interface{}{
		"amount":    displayAmount,
		"isThunder": isThunder,
		"loseMoney": displayLoseMoney,
		"openNum":   grabIndex,
		"grabIndex": grabIndex,
		"luckyInfo": luckyMoney,
		"message":   "",
	}

	if isThunder == 1 {
		result["message"] = fmt.Sprintf("中雷，领取 %.2f U，损失 %.2f U", displayAmount, displayLoseMoney)
	} else {
		result["message"] = fmt.Sprintf("未中雷，领取 %.2f U", displayAmount)
	}

	return result, nil
}

// CheckGrabBalance 检查抢包余额
func CheckGrabBalance(db *gorm.DB, luckyID int64, userID int64, tablePrefix string) error {
	luckyMoney, err := repository.GetLuckyMoney(db, luckyID)
	if err != nil {
		return errors.New("数据有误！红包不存在")
	}

	user, err := getTgUserByID(db, userID)
	if err != nil {
		return errors.New("用户未注册，请先注册!命令：/register")
	}

	if user.Status != 1 {
		return errors.New("用户已禁用，请联系管理员处理")
	}

	loseRate := GetLoseRate(db)
	lowestAmount := luckyMoney.Amount * loseRate
	if user.Balance < lowestAmount {
		return fmt.Errorf("你至少需要有%.2fU才能抢这个红包~", lowestAmount)
	}

	return nil
}

// GetRedPacketStatus 获取红包状态
func GetRedPacketStatus(db *gorm.DB, luckyID int64) (*pojo.LuckyMoney, int64, error) {
	luckyMoney, err := repository.GetLuckyMoney(db, luckyID)
	if err != nil {
		return nil, 0, err
	}

	grabbedCount, err := repository.GetLuckyHistoryCount(db, luckyID)
	if err != nil {
		return nil, 0, err
	}

	return &luckyMoney, grabbedCount, nil
}

// GetRedPacketDetails 获取红包详情（包含领取记录）
func GetRedPacketDetails(db *gorm.DB, luckyID int64) (map[string]interface{}, error) {
	luckyMoney, err := repository.GetLuckyMoney(db, luckyID)
	if err != nil {
		return nil, err
	}

	historyList, err := repository.GetLuckyHistoryByLuckyId(db, luckyID)
	if err != nil {
		return nil, err
	}

	result := map[string]interface{}{
		"luckyMoney": luckyMoney,
		"history":    historyList,
	}

	return result, nil
}

// GetLoseRate 获取中雷倍数（带Redis缓存）
func GetLoseRate(db *gorm.DB) float64 {
	defaultValue := 1.8
	redisKey := fmt.Sprintf("bgu_auth_group_lose_rate")
	ctx := context.Background()
	configKey := fmt.Sprintf("lucky_lose_rate")

	// 先从Redis缓存获取
	cachedValue, err := utils.RD.Get(ctx, redisKey).Result()
	if err == nil && cachedValue != "" {
		if value, parseErr := strconv.ParseFloat(cachedValue, 64); parseErr == nil && value > 0 {
			return value
		}
	}

	// 缓存未命中，从sys_config查询；key不存在时写入默认值
	configValue := getOrInitSysConfigValue(db, configKey, strconv.FormatFloat(defaultValue, 'f', 2, 64), "红包中雷倍数")
	result, parseErr := strconv.ParseFloat(configValue, 64)
	if parseErr != nil || result <= 0 {
		result = defaultValue
		_ = db.Model(&pojo.SysConfig{}).Where("config_key = ?", configKey).Update("config_value", strconv.FormatFloat(defaultValue, 'f', 2, 64)).Error
	}
	// 存入Redis，设置过期时间为20-40分钟随机
	utils.RD.SetEX(ctx, redisKey, strconv.FormatFloat(result, 'f', 2, 64), utils.GetRandomRangeSecond(20*60, 40*60))

	return result
}

// GetLuckyNumConfig 获取红包数量配置（带Redis缓存）
func GetLuckyNumConfig(db *gorm.DB) string {
	defaultValue := "3"
	redisKey := fmt.Sprintf("bgu_auth_group_num_config")
	ctx := context.Background()
	configKey := fmt.Sprintf("lucky_num_config")

	// 先从Redis缓存获取
	cachedValue, err := utils.RD.Get(ctx, redisKey).Result()
	if err == nil && cachedValue != "" {
		return cachedValue
	}

	// 缓存未命中，从sys_config查询；key不存在时写入默认值
	result := getOrInitSysConfigValue(db, configKey, defaultValue, "红包数量配置")
	if result == "" {
		result = defaultValue
		_ = db.Model(&pojo.SysConfig{}).Where("config_key = ?", configKey).Update("config_value", defaultValue).Error
	}
	// 存入Redis，设置过期时间为20-40分钟随机
	utils.RD.SetEX(ctx, redisKey, result, utils.GetRandomRangeSecond(20*60, 40*60))

	return result
}

// GetSendCommission 获取发包中雷抽成百分比（带Redis缓存）
func GetSendCommission(db *gorm.DB) int {
	defaultValue := 2
	redisKey := fmt.Sprintf("bgu_auth_group_send_commission")
	ctx := context.Background()
	configKey := fmt.Sprintf("lucky_send_commission")

	// 先从Redis缓存获取
	cachedValue, err := utils.RD.Get(ctx, redisKey).Result()
	if err == nil && cachedValue != "" {
		if value, parseErr := strconv.Atoi(cachedValue); parseErr == nil && value >= 0 {
			return value
		}
	}

	// 缓存未命中，从sys_config查询；key不存在时写入默认值
	configValue := getOrInitSysConfigValue(db, configKey, strconv.Itoa(defaultValue), "红包中雷抽成百分比")
	result, parseErr := strconv.Atoi(configValue)
	if parseErr != nil || result < 0 {
		result = defaultValue
		_ = db.Model(&pojo.SysConfig{}).Where("config_key = ?", configKey).Update("config_value", strconv.Itoa(defaultValue)).Error
	}
	// 存入Redis，设置过期时间为20-40分钟随机
	utils.RD.SetEX(ctx, redisKey, strconv.Itoa(result), utils.GetRandomRangeSecond(20*60, 40*60))

	return result
}

// GetGrabbingCommission 获取抢红包抽成百分比（带Redis缓存）
func GetGrabbingCommission(db *gorm.DB) int {
	defaultValue := 3
	redisKey := fmt.Sprintf("bgu_auth_group_grabbing_commission")
	ctx := context.Background()
	configKey := fmt.Sprintf("lucky_grabbing_commission")

	// 先从Redis缓存获取
	cachedValue, err := utils.RD.Get(ctx, redisKey).Result()
	if err == nil && cachedValue != "" {
		if value, parseErr := strconv.Atoi(cachedValue); parseErr == nil && value >= 0 {
			return value
		}
	}

	// 缓存未命中，从sys_config查询；key不存在时写入默认值
	configValue := getOrInitSysConfigValue(db, configKey, strconv.Itoa(defaultValue), "抢红包抽成百分比")
	result, parseErr := strconv.Atoi(configValue)
	if parseErr != nil || result < 0 {
		result = defaultValue
		_ = db.Model(&pojo.SysConfig{}).Where("config_key = ?", configKey).Update("config_value", strconv.Itoa(defaultValue)).Error
	}
	// 存入Redis，设置过期时间为20-40分钟随机
	utils.RD.SetEX(ctx, redisKey, strconv.Itoa(result), utils.GetRandomRangeSecond(20*60, 40*60))

	return result
}

// GetLuckyExpireMinutes 获取红包过期分钟配置（带Redis缓存）
func GetLuckyExpireMinutes(db *gorm.DB) int {
	defaultValue := 3
	redisKey := fmt.Sprintf("bgu_lucky_expire_time")
	ctx := context.Background()
	configKey := fmt.Sprintf("lucky_expire_time")

	// 先从Redis缓存获取
	cachedValue, err := utils.RD.Get(ctx, redisKey).Result()
	if err == nil && cachedValue != "" {
		if value, parseErr := strconv.Atoi(cachedValue); parseErr == nil && value > 0 {
			return value
		}
	}

	// 缓存未命中，从sys_config查询；key不存在时写入默认值
	configValue := getOrInitSysConfigValue(db, configKey, strconv.Itoa(defaultValue), "红包过期时间(分钟)")
	result, parseErr := strconv.Atoi(configValue)
	if parseErr != nil || result <= 0 {
		result = defaultValue
		_ = db.Model(&pojo.SysConfig{}).Where("config_key = ?", configKey).Update("config_value", strconv.Itoa(defaultValue)).Error
	}
	// 存入Redis，设置过期时间为20-40分钟随机
	utils.RD.SetEX(ctx, redisKey, strconv.Itoa(result), utils.GetRandomRangeSecond(20*60, 40*60))

	return result
}

func getOrInitSysConfigValue(db *gorm.DB, configKey string, defaultValue string, configDesc string) string {
	var sysConfig pojo.SysConfig
	err := db.Where("config_key = ?", configKey).First(&sysConfig).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			sysConfig = pojo.SysConfig{
				ConfigKey:   configKey,
				ConfigValue: defaultValue,
				ConfigDesc:  configDesc,
			}
			if createErr := db.Create(&sysConfig).Error; createErr == nil {
				return defaultValue
			}
			return defaultValue
		}
		return defaultValue
	}
	if strings.TrimSpace(sysConfig.ConfigValue) == "" {
		_ = db.Model(&sysConfig).Update("config_value", defaultValue).Error
		return defaultValue
	}
	return sysConfig.ConfigValue
}

func getTgUserByID(db *gorm.DB, userID int64) (pojo.TgUser, error) {
	var user pojo.TgUser
	err := db.Where("id = ?", userID).First(&user).Error
	if err != nil || user.ID == 0 {
		return user, errors.New("用户不存在")
	}
	return user, nil
}

func getTgUserDisplayName(user *pojo.TgUser) string {
	if user == nil {
		return ""
	}
	if user.FirstName != nil && *user.FirstName != "" {
		return *user.FirstName
	}
	if user.Username != nil && *user.Username != "" {
		return *user.Username
	}
	return fmt.Sprintf("User_%d", user.TgID)
}

// GetDefaultBalance 获取默认余额
func GetDefaultBalance(tablePrefix string, chatID int64) float64 {
	configKey := fmt.Sprintf("default_balance_%d", chatID)
	defaultValue := int64(1000) // 默认1000，单位是分*10，所以是1000 = 1.000
	value := utils.GetInt64Cache(tablePrefix, configKey, defaultValue)
	return float64(value) / 1000.0
}

func getInviteLuckyRebateRate(tablePrefix string) float64 {
	defaultValue := "40"
	val := utils.GetStringCache(tablePrefix, "invite_lucky_rebate_rate", &defaultValue)
	if val == nil || strings.TrimSpace(*val) == "" {
		r, _ := strconv.ParseFloat(defaultValue, 64)
		return r
	}
	r, err := strconv.ParseFloat(strings.TrimSpace(*val), 64)
	if err != nil {
		r, _ = strconv.ParseFloat(defaultValue, 64)
		return r
	}
	return r
}

func getInviteThunderRebateRate(tablePrefix string) float64 {
	defaultValue := "40"
	val := utils.GetStringCache(tablePrefix, "invite_thunder_rebate_rate", &defaultValue)
	if val == nil || strings.TrimSpace(*val) == "" {
		r, _ := strconv.ParseFloat(defaultValue, 64)
		return r
	}
	r, err := strconv.ParseFloat(strings.TrimSpace(*val), 64)
	if err != nil {
		r, _ = strconv.ParseFloat(defaultValue, 64)
		return r
	}
	return r
}

func applyLuckyInviteRebate(
	tx *gorm.DB,
	tablePrefix string,
	subUserID int64,
	tenantID int64,
	platformIncome float64,
	rate float64,
	luckyID int64,
	grabIndex int,
	scene string,
) error {
	if platformIncome <= 0 || rate <= 0 || subUserID <= 0 {
		return nil
	}

	var subUser pojo.TgUser
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", subUserID).First(&subUser).Error; err != nil || subUser.ID == 0 {
		return nil
	}
	if subUser.ParentID == nil || *subUser.ParentID <= 0 {
		return nil
	}

	parentID := *subUser.ParentID
	var parent pojo.TgUser
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", parentID).First(&parent).Error; err != nil || parent.ID == 0 {
		return nil
	}
	if parent.Status != 1 {
		return nil
	}

	rebateAmount := utils.ToMoney(platformIncome).Multiply(rate / 100).ToDollars()
	if rebateAmount <= 0 {
		return nil
	}

	idempotencyKey := fmt.Sprintf("invite_rebate:%s:%d:%d:%d", scene, luckyID, grabIndex, parentID)
	remark := fmt.Sprintf("lucky_%s_invite_rebate", scene)
	sourceOrderID := fmt.Sprintf("lucky_%d_%d", luckyID, grabIndex)
	tenant := tenantID
	record := pojo.TgUserRebateRecord{
		TenantId:       &tenant,
		SubUserId:      subUserID,
		ParentUserId:   parentID,
		SourceType:     3,
		SourceOrderId:  sourceOrderID,
		SourceAmount:   platformIncome,
		RebateRate:     rate,
		RebateAmount:   rebateAmount,
		Currency:       "USDT",
		Status:         1,
		SettledAt:      ptrTime(time.Now()),
		IdempotencyKey: idempotencyKey,
		Remark:         &remark,
	}
	// OnConflict DoNothing：若 idempotency_key 唯一索引冲突（并发重入），静默跳过而非报错回滚
	result := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&record)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		// 已存在，幂等跳过
		return nil
	}

	return tx.Model(&pojo.TgUser{}).
		Where("id = ?", parentID).
		Updates(map[string]any{
			"rebate_amount":       gorm.Expr("rebate_amount + ?", rebateAmount),
			"rebate_total_amount": gorm.Expr("rebate_total_amount + ?", rebateAmount),
		}).Error
}

func ptrTime(t time.Time) *time.Time {
	return &t
}
