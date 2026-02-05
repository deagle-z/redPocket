package services

import (
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/repository"
	"BaseGoUni/core/utils"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"strconv"
	"time"
)

// SendRedPacket 发送红包业务逻辑
func SendRedPacket(db *gorm.DB, senderID int64, senderName string, req pojo.LuckyMoneySend, tablePrefix string) (*pojo.LuckyMoney, error) {
	// 验证金额
	if req.Amount < 5 {
		return nil, errors.New("红包金额不能小于5U")
	}

	// 验证雷号
	if req.Thunder < 0 || req.Thunder > 9 {
		return nil, errors.New("指令有误，雷数应是0~9之间")
	}

	// 获取红包数量配置
	luckyNumConfig := GetLuckyNumConfig(db, req.ChatID)
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

	// 序列化红包列表
	redListJSON, err := json.Marshal(redEnvelopes)
	if err != nil {
		return nil, fmt.Errorf("生成红包列表失败: %v", err)
	}

	// 获取中雷倍数
	loseRate := GetLoseRate(db, req.ChatID)

	// 创建红包记录
	luckyMoney := &pojo.LuckyMoney{
		SenderID:   senderID,
		SenderName: senderName,
		Amount:     req.Amount,
		Number:     luckyTotal,
		Thunder:    req.Thunder,
		ChatID:     req.ChatID,
		RedList:    string(redListJSON),
		LoseRate:   loseRate,
		Status:     1, // 进行中
		Lucky:      1, // 有效
		Received:   0,
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

	// 扣除发送者余额
	if err := tx.Model(&pojo.TgUser{}).
		Where("id = ?", senderID).
		Update("balance", gorm.Expr("balance - ?", req.Amount)).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("扣除余额失败: %v", err)
	}

	// 记录余额变动
	cashHistory := pojo.CashHistory{
		UserId:      senderID,
		AwardUni:    fmt.Sprintf("lucky_%d", luckyMoney.ID),
		Amount:      -req.Amount,
		StartAmount: user.Balance,
		EndAmount:   user.Balance - req.Amount,
		CashMark:    "发送红包",
		CashDesc:    fmt.Sprintf("发送红包 %dU，雷号%d", int(req.Amount), req.Thunder),
		FromUserId:  senderID,
	}
	if err := tx.Create(&cashHistory).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("记录余额变动失败: %v", err)
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("提交事务失败: %v", err)
	}

	return luckyMoney, nil
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
	isThunder := 0
	loseMoney := 0.0
	amountStr := fmt.Sprintf("%.2f", redAmount)
	lastDigit := amountStr[len(amountStr)-1]
	thunderStr := strconv.Itoa(luckyMoney.Thunder)

	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if string(lastDigit) == thunderStr {
		// 中雷
		isThunder = 1
		loseMoney = luckyMoney.Amount * luckyMoney.LoseRate

		// 扣除用户余额
		if err := tx.Model(&pojo.TgUser{}).
			Where("id = ?", userID).
			Update("balance", gorm.Expr("balance - ?", loseMoney)).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("扣除余额失败: %v", err)
		}

		// 获取发包中雷抽成百分比
		sendCommission := GetSendCommission(db, luckyMoney.ChatID)
		// 计算抽成金额
		commissionAmount := loseMoney * float64(sendCommission) / 100.0
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
		cashHistoryUser := pojo.CashHistory{
			UserId:      userID,
			AwardUni:    fmt.Sprintf("lucky_grab_%d_%d_%d_%d", luckyID, userID, awardTs, int64(loseMoney*1000)),
			Amount:      -loseMoney,
			StartAmount: user.Balance,
			EndAmount:   user.Balance - loseMoney,
			CashMark:    "抢红包中雷",
			CashDesc:    fmt.Sprintf("抢红包中雷，损失%.2fU", loseMoney),
			FromUserId:  luckyMoney.SenderID,
		}
		if err := tx.Create(&cashHistoryUser).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("记录余额变动失败: %v", err)
		}

		// 记录余额变动（发送者）- 抽成后金额（实际到账）
		senderUser, err := getTgUserByID(db, luckyMoney.SenderID)
		if err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("获取发送者失败: %v", err)
		}
		cashHistorySender := pojo.CashHistory{
			UserId:      luckyMoney.SenderID,
			AwardUni:    fmt.Sprintf("lucky_thunder_%d_%d_%d_%d", luckyID, userID, awardTs, int64(actualLoseMoney*1000)),
			Amount:      actualLoseMoney,
			StartAmount: senderUser.Balance,
			EndAmount:   senderUser.Balance + actualLoseMoney, // 实际到账金额
			CashMark:    "红包中雷收益",
			CashDesc:    fmt.Sprintf("红包中雷收益，获得%.2fU（抽成后）", actualLoseMoney),
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
				StartAmount: senderUser.Balance + actualLoseMoney, // 抽成后金额
				EndAmount:   senderUser.Balance + actualLoseMoney, // 抽成后金额（不变）
				CashMark:    "红包中雷抽成",
				CashDesc:    fmt.Sprintf("红包中雷抽成%.2f%%，抽成金额%.2fU", float64(sendCommission), commissionAmount),
				FromUserId:  userID,
			}
			if err := tx.Create(&cashHistoryCommission).Error; err != nil {
				tx.Rollback()
				return nil, fmt.Errorf("记录余额变动失败: %v", err)
			}
		}
	} else {
		// 未中雷，增加用户余额（需要扣除抽成）
		// 获取抢红包抽成百分比
		grabbingCommission := GetGrabbingCommission(db, luckyMoney.ChatID)
		// 计算抽成金额
		commissionAmount := redAmount * float64(grabbingCommission) / 100.0
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
			CashDesc:    fmt.Sprintf("抢红包，获得%.2fU（抽成后）", actualAmount),
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
				FromUserId:  luckyMoney.SenderID,
			}
			if err := tx.Create(&cashHistoryCommission).Error; err != nil {
				tx.Rollback()
				return nil, fmt.Errorf("记录余额变动失败: %v", err)
			}
		}
	}

	// 创建领取记录
	history := &pojo.LuckyHistory{
		UserID:    userID,
		FirstName: utils.FormatName(getTgUserDisplayName(&user), 8),
		LuckyID:   luckyID,
		IsThunder: isThunder,
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

	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("提交事务失败: %v", err)
	}

	// 返回结果
	result := map[string]interface{}{
		"amount":    redAmount,
		"isThunder": isThunder,
		"loseMoney": loseMoney,
		"openNum":   grabIndex,
		"luckyInfo": luckyMoney,
		"message":   "",
	}

	if isThunder == 1 {
		result["message"] = fmt.Sprintf("中雷，领取 %.2f U，损失 %.2f U", redAmount, loseMoney)
	} else {
		result["message"] = fmt.Sprintf("未中雷，领取 %.2f U", redAmount)
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

	loseRate := GetLoseRate(db, luckyMoney.ChatID)
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
func GetLoseRate(db *gorm.DB, chatID int64) float64 {
	defaultValue := 1.8
	redisKey := fmt.Sprintf("bgu_auth_group_lose_rate_%d", chatID)
	ctx := context.Background()

	// 先从Redis缓存获取
	cachedValue, err := utils.RD.Get(ctx, redisKey).Result()
	if err == nil && cachedValue != "" {
		if value, parseErr := strconv.ParseFloat(cachedValue, 64); parseErr == nil && value > 0 {
			return value
		}
	}

	// 缓存未命中，从数据库查询
	var authGroup pojo.AuthGroup
	err = db.Where("group_id = ?", chatID).First(&authGroup).Error
	if err != nil {
		// 即使查询失败也缓存默认值，避免频繁查询数据库
		utils.RD.SetEX(ctx, redisKey, strconv.FormatFloat(defaultValue, 'f', 2, 64), utils.GetRandomRangeSecond(20*60, 40*60))
		return defaultValue
	}

	// 如果查询到有效值，存入缓存
	result := defaultValue
	if authGroup.LoseRate > 0 {
		result = authGroup.LoseRate
		// 存入Redis，设置过期时间为20-40分钟随机
		utils.RD.SetEX(ctx, redisKey, strconv.FormatFloat(result, 'f', 2, 64), utils.GetRandomRangeSecond(20*60, 40*60))
	} else {
		// 即使没有值也缓存默认值，避免频繁查询数据库
		utils.RD.SetEX(ctx, redisKey, strconv.FormatFloat(defaultValue, 'f', 2, 64), utils.GetRandomRangeSecond(20*60, 40*60))
	}

	return result
}

// GetLuckyNumConfig 获取红包数量配置（带Redis缓存）
func GetLuckyNumConfig(db *gorm.DB, chatID int64) string {
	defaultValue := "3"
	redisKey := fmt.Sprintf("bgu_auth_group_num_config_%d", chatID)
	ctx := context.Background()

	// 先从Redis缓存获取
	cachedValue, err := utils.RD.Get(ctx, redisKey).Result()
	if err == nil && cachedValue != "" {
		return cachedValue
	}

	// 缓存未命中，从数据库查询
	var authGroup pojo.AuthGroup
	err = db.Where("group_id = ?", chatID).First(&authGroup).Error
	if err != nil {
		// 即使查询失败也缓存默认值，避免频繁查询数据库
		utils.RD.SetEX(ctx, redisKey, defaultValue, utils.GetRandomRangeSecond(20*60, 40*60))
		return defaultValue
	}

	// 如果查询到有效值，存入缓存
	result := defaultValue
	if authGroup.NumConfig != "" {
		result = authGroup.NumConfig
	} else {
		result = defaultValue
	}
	// 存入Redis，设置过期时间为20-40分钟随机
	utils.RD.SetEX(ctx, redisKey, result, utils.GetRandomRangeSecond(20*60, 40*60))

	return result
}

// GetSendCommission 获取发包中雷抽成百分比（带Redis缓存）
func GetSendCommission(db *gorm.DB, chatID int64) int {
	defaultValue := 2
	redisKey := fmt.Sprintf("bgu_auth_group_send_commission_%d", chatID)
	ctx := context.Background()

	// 先从Redis缓存获取
	cachedValue, err := utils.RD.Get(ctx, redisKey).Result()
	if err == nil && cachedValue != "" {
		if value, parseErr := strconv.Atoi(cachedValue); parseErr == nil && value >= 0 {
			return value
		}
	}

	// 缓存未命中，从数据库查询
	var authGroup pojo.AuthGroup
	err = db.Where("group_id = ?", chatID).First(&authGroup).Error
	if err != nil {
		// 即使查询失败也缓存默认值，避免频繁查询数据库
		utils.RD.SetEX(ctx, redisKey, strconv.Itoa(defaultValue), utils.GetRandomRangeSecond(20*60, 40*60))
		return defaultValue
	}

	// 如果查询到有效值，存入缓存
	result := defaultValue
	if authGroup.SendCommission > 0 {
		result = authGroup.SendCommission
	} else {
		result = defaultValue
	}
	// 存入Redis，设置过期时间为20-40分钟随机
	utils.RD.SetEX(ctx, redisKey, strconv.Itoa(result), utils.GetRandomRangeSecond(20*60, 40*60))

	return result
}

// GetGrabbingCommission 获取抢红包抽成百分比（带Redis缓存）
func GetGrabbingCommission(db *gorm.DB, chatID int64) int {
	defaultValue := 3
	redisKey := fmt.Sprintf("bgu_auth_group_grabbing_commission_%d", chatID)
	ctx := context.Background()

	// 先从Redis缓存获取
	cachedValue, err := utils.RD.Get(ctx, redisKey).Result()
	if err == nil && cachedValue != "" {
		if value, parseErr := strconv.Atoi(cachedValue); parseErr == nil && value >= 0 {
			return value
		}
	}

	// 缓存未命中，从数据库查询
	var authGroup pojo.AuthGroup
	err = db.Where("group_id = ?", chatID).First(&authGroup).Error
	if err != nil {
		// 即使查询失败也缓存默认值，避免频繁查询数据库
		utils.RD.SetEX(ctx, redisKey, strconv.Itoa(defaultValue), utils.GetRandomRangeSecond(20*60, 40*60))
		return defaultValue
	}

	// 如果查询到有效值，存入缓存
	result := defaultValue
	if authGroup.GrabbingCommission > 0 {
		result = authGroup.GrabbingCommission
	} else {
		result = defaultValue
	}
	// 存入Redis，设置过期时间为20-40分钟随机
	utils.RD.SetEX(ctx, redisKey, strconv.Itoa(result), utils.GetRandomRangeSecond(20*60, 40*60))

	return result
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
