package services

import (
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/repository"
	"BaseGoUni/core/services"
	"BaseGoUni/core/utils"
	"context"
	"errors"
	"fmt"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"gorm.io/gorm"
	"log"
	"math/rand/v2"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// TelegramBotService Telegram Bot 服务
type TelegramBotService struct {
	DB          *gorm.DB
	TablePrefix string
	BotToken    string
	Bot         *bot.Bot
}

// InitTelegramBot 初始化 Telegram Bot
func InitTelegramBot(db *gorm.DB, tablePrefix string, botToken string) error {
	if botToken == "" {
		log.Println("Telegram Bot Token 未配置，跳过初始化")
		return nil
	}

	botService := &TelegramBotService{
		DB:          db,
		TablePrefix: tablePrefix,
		BotToken:    botToken,
	}

	ctx := context.Background()

	// 创建 Bot 实例
	b, err := bot.New(botToken,
		bot.WithDefaultHandler(botService.handleDefault),
		bot.WithMessageTextHandler("/start", bot.MatchTypeExact, botService.handleStartCommand),
		bot.WithMessageTextHandler("/help", bot.MatchTypeExact, botService.handleHelpCommand),
		bot.WithMessageTextHandler("/register", bot.MatchTypeExact, botService.handleRegisterCommand),
		bot.WithMessageTextHandler("/recharge", bot.MatchTypeExact, botService.handleRechargeCommand),
		bot.WithMessageTextHandler("/withdraw", bot.MatchTypeExact, botService.handleWithdrawCommand),
		bot.WithMessageTextHandler("/team", bot.MatchTypeExact, botService.handleTeamCommand),
		bot.WithMessageTextHandler("/invite", bot.MatchTypeExact, botService.handleInviteCommand),
		bot.WithMessageTextHandler("/rebate", bot.MatchTypeExact, botService.handleRebateCommand),
		bot.WithCallbackQueryDataHandler("qiang-", bot.MatchTypePrefix, botService.handleGrabCallback),
		bot.WithCallbackQueryDataHandler("balance", bot.MatchTypeExact, botService.handleBalanceCallback),
		bot.WithCallbackQueryDataHandler("balance_", bot.MatchTypePrefix, botService.handleBalanceActionCallback),
		bot.WithCallbackQueryDataHandler("recharge_amount_", bot.MatchTypePrefix, botService.handleRechargeAmountCallback),
		bot.WithCallbackQueryDataHandler("recharge_custom", bot.MatchTypeExact, botService.handleRechargeCustomCallback),
		bot.WithCallbackQueryDataHandler("recharge_usdt", bot.MatchTypeExact, botService.handleRechargeUsdtCallback),
		bot.WithCallbackQueryDataHandler("recharge_close", bot.MatchTypeExact, botService.handleRechargeCloseCallback),
		bot.WithCallbackQueryDataHandler("today_data", bot.MatchTypeExact, botService.handleTodayDataCallback),
		bot.WithCallbackQueryDataHandler("share_data", bot.MatchTypeExact, botService.handleShareDataCallback),
	)
	if err != nil {
		return fmt.Errorf("初始化 Telegram Bot 失败: %v", err)
	}

	botService.Bot = b

	// 获取 Bot 信息
	botUser, err := b.GetMe(ctx)
	if err != nil {
		return fmt.Errorf("获取 Bot 信息失败: %v", err)
	}

	log.Printf("Telegram Bot 启动成功: @%s (ID: %d)", botUser.Username, botUser.ID)

	// 在 goroutine 中启动 Bot
	go func() {
		log.Println("Telegram Bot 开始监听消息...")
		b.Start(ctx)
	}()

	log.Printf("Telegram Bot 服务初始化完成 (Token: %s...)", botToken[:10])
	return nil
}

// handleDefault 默认处理器
func (s *TelegramBotService) handleDefault(ctx context.Context, b *bot.Bot, update *models.Update) {
	// 处理消息
	if update.Message != nil {
		s.handleMessage(ctx, b, update.Message)
		return
	}

	// 处理回调查询（如果没有匹配到特定的 callback handler）
	if update.CallbackQuery != nil {
		// 默认回调处理
		_, _ = b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: update.CallbackQuery.ID,
			Text:            "未知操作",
		})
		return
	}
}

// handleMessage 处理消息
func (s *TelegramBotService) handleMessage(ctx context.Context, b *bot.Bot, message *models.Message) {
	chatID := message.Chat.ID
	text := message.Text
	if text == "" {
		return
	}

	// 私聊自定义充值金额
	if message.Chat.Type == "private" || chatID > 0 {
		if s.tryHandleCustomRecharge(ctx, b, message, text) {
			return
		}
	}

	// 私聊充值提示
	if matched, _ := regexp.MatchString(`(?i)^(充值|recharge|钱包)$`, text); matched {
		if message.Chat.Type == "private" || chatID > 0 {
			s.handleRechargePrivateMessage(ctx, b, message)
			return
		}
	}

	// 私聊注册
	if matched, _ := regexp.MatchString(`(?i)^(注册|register)$`, text); matched {
		if message.Chat.Type == "private" || chatID > 0 {
			s.handleRegisterTextMessage(ctx, b, message)
			return
		}
	}

	// 私聊提现
	if matched, _ := regexp.MatchString(`(?i)^(提现|withdraw)$`, text); matched {
		if message.Chat.Type == "private" || chatID > 0 {
			s.handleWithdrawMessage(ctx, b, message)
			return
		}
	}

	// 私聊余额查询
	if matched, _ := regexp.MatchString(`(?i)^(1|查|余额|查余额)$`, text); matched {
		if message.Chat.Type == "private" || chatID > 0 {
			s.handleBalancePrivateMessage(ctx, b, message)
		} else {
			s.handleBalanceGroupMessage(ctx, b, message)
		}
		return
	}

	// 团队信息查询
	if matched, _ := regexp.MatchString(`(?i)^(团队|我的团队)$`, text); matched {
		if message.Chat.Type == "private" || chatID > 0 {
			s.handleTeamMessage(ctx, b, message)
		}
		return
	}

	// 邀请信息查询
	if matched, _ := regexp.MatchString(`(?i)^(邀请|邀请好友|invite)$`, text); matched {
		if message.Chat.Type == "private" || chatID > 0 {
			s.handleInviteMessage(ctx, b, message)
		}
		return
	}

	// 反水/佣金信息查询
	if matched, _ := regexp.MatchString(`(?i)^(反水|佣金|佣金明细|rebate|commission)$`, text); matched {
		if message.Chat.Type == "private" || chatID > 0 {
			s.handleRebateMessage(ctx, b, message)
		}
		return
	}

	// 只处理群组消息
	if message.Chat.Type != "group" && message.Chat.Type != "supergroup" {
		return
	}

	// 检查群组授权
	if !s.CheckGroupAuth(chatID) {
		_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   "未授权",
		})
		return
	}

	// 处理红包命令（3参数格式：发10-3-1）
	pattern3 := regexp.MustCompile(`(?i)(?:发[包]*)?(\d+)[-/](\d+)[-/](\d+)`)
	if pattern3.MatchString(text) {
		s.handleRedPacketMessage3(ctx, b, &models.Update{Message: message})
		return
	}

	// 处理红包命令（2参数格式：发10-1）
	pattern2 := regexp.MustCompile(`(?i)(?:发[包]*)?(\d+)[-/]([0-9])`)
	if pattern2.MatchString(text) {
		s.handleRedPacketMessage2(ctx, b, &models.Update{Message: message})
		return
	}

	// 处理群信息查询
	if matched, _ := regexp.MatchString(`(?i)(群信息|获取群信息|查看群信息)`, text); matched {
		s.handleGroupInfoMessage(ctx, b, &models.Update{Message: message})
		return
	}

	// 未匹配任何指令：群组消息按配置决定是否删除
	if message.Chat.Type == "group" || message.Chat.Type == "supergroup" {
		groupInfo, err := s.GetGroupInfo(chatID)
		if err == nil && groupInfo != nil && groupInfo.DeleteMsg == 1 {
			userIDStr := fmt.Sprintf("%d", message.From.ID)
			if !isWhiteId(groupInfo.WhiteIds, userIDStr) {
				_, _ = b.DeleteMessage(ctx, &bot.DeleteMessageParams{
					ChatID:    chatID,
					MessageID: message.ID,
				})
			}
		}
	}
}

func isWhiteId(whiteIds string, userID string) bool {
	if whiteIds == "" || userID == "" {
		return false
	}
	parts := strings.FieldsFunc(whiteIds, func(r rune) bool {
		return r == ',' || r == ' ' || r == ';' || r == '|' || r == '\n' || r == '\t'
	})
	for _, part := range parts {
		if strings.TrimSpace(part) == userID {
			return true
		}
	}
	return false
}

// handleStartCommand 处理 /start 命令
func (s *TelegramBotService) handleStartCommand(ctx context.Context, b *bot.Bot, update *models.Update) {
	message := update.Message
	if message == nil {
		return
	}

	inviteCode := parseStartInviteCode(message.Text)
	if inviteCode != "" && message.From != nil {
		redisKey := fmt.Sprintf("tg_start_invite:%d", message.From.ID)
		_ = utils.RD.Set(ctx, redisKey, inviteCode, 24*time.Hour).Err()
	}

	_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      message.Chat.ID,
		Text:        "开始游戏",
		ParseMode:   models.ParseModeHTML,
		ReplyMarkup: s.buildPrivateEntryKeyboard(),
	})
}

// handleHelpCommand 处理 /help 命令
func (s *TelegramBotService) handleHelpCommand(ctx context.Context, b *bot.Bot, update *models.Update) {
	message := update.Message
	if message == nil {
		return
	}

	helpText := `
注册：<code>/register</code>
发红包：<code>发10-1</code>或<code>10-1</code>
发红包（指定数量）：<code>发10-3-1</code>或<code>10-3-1</code>
查余额：<code>查</code>或<code>1</code>或<code>余额</code>
获取ID：<code>获取群信息</code>
`

	_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    message.Chat.ID,
		Text:      helpText,
		ParseMode: models.ParseModeHTML,
	})
}

// handleRegisterCommand 处理 /register 命令
func (s *TelegramBotService) handleRegisterCommand(ctx context.Context, b *bot.Bot, update *models.Update) {
	message := update.Message
	if message == nil || message.From == nil {
		return
	}

	chatID := message.Chat.ID
	userID := message.From.ID
	userName := message.From.FirstName
	userUsername := message.From.Username
	if userName == "" {
		userName = userUsername
	}
	if userName == "" {
		userName = fmt.Sprintf("User_%d", userID)
	}

	if message.Chat.Type == "group" || message.Chat.Type == "supergroup" {
		if !s.CheckGroupAuth(chatID) {
			_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: chatID,
				Text:   "未授权",
			})
			return
		}
	}

	result, err := s.HandleRegisterCommand(chatID, userID, userName, userUsername)
	if err != nil {
		result = err.Error()
	}

	_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      chatID,
		Text:        result,
		ParseMode:   models.ParseModeHTML,
		ReplyMarkup: s.buildPrivateEntryKeyboard(),
	})
}

func (s *TelegramBotService) handleRechargeCommand(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update == nil || update.Message == nil {
		return
	}
	s.handleRechargePrivateMessage(ctx, b, update.Message)
}

func (s *TelegramBotService) handleWithdrawCommand(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update == nil || update.Message == nil {
		return
	}
	s.handleWithdrawMessage(ctx, b, update.Message)
}

func (s *TelegramBotService) handleTeamCommand(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update == nil || update.Message == nil {
		return
	}
	s.handleTeamMessage(ctx, b, update.Message)
}

func (s *TelegramBotService) handleInviteCommand(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update == nil || update.Message == nil {
		return
	}
	s.handleInviteMessage(ctx, b, update.Message)
}

func (s *TelegramBotService) handleRebateCommand(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update == nil || update.Message == nil {
		return
	}
	s.handleRebateMessage(ctx, b, update.Message)
}

// handleRedPacketMessage3 处理红包消息（格式：发10-3-1）
func (s *TelegramBotService) handleRedPacketMessage3(ctx context.Context, b *bot.Bot, update *models.Update) {
	message := update.Message
	if message == nil || message.From == nil {
		return
	}

	// 只处理群组消息
	if message.Chat.Type != "group" && message.Chat.Type != "supergroup" {
		return
	}

	chatID := message.Chat.ID
	userID := int64(message.From.ID)
	userName := message.From.FirstName
	if userName == "" {
		userName = message.From.Username
	}
	if userName == "" {
		userName = fmt.Sprintf("User_%d", userID)
	}

	// 检查群组授权
	if !s.CheckGroupAuth(chatID) {
		return
	}

	text := message.Text
	amount, number, thunder, ok := ParseRedPacketCommand(text)
	if !ok || number == nil {
		return // 不是3参数格式，让其他处理器处理
	}

	// 处理发送红包
	s.processRedPacket(ctx, b, message, chatID, userID, userName, amount, number, thunder)
}

// handleRedPacketMessage2 处理红包消息（格式：发10-1）
func (s *TelegramBotService) handleRedPacketMessage2(ctx context.Context, b *bot.Bot, update *models.Update) {
	message := update.Message
	if message == nil || message.From == nil {
		return
	}

	// 只处理群组消息
	if message.Chat.Type != "group" && message.Chat.Type != "supergroup" {
		return
	}

	chatID := message.Chat.ID
	userID := int64(message.From.ID)
	userName := message.From.FirstName
	if userName == "" {
		userName = message.From.Username
	}
	if userName == "" {
		userName = fmt.Sprintf("User_%d", userID)
	}

	// 检查群组授权
	if !s.CheckGroupAuth(chatID) {
		return
	}

	text := message.Text
	amount, number, thunder, ok := ParseRedPacketCommand(text)
	if !ok || number != nil {
		return // 不是2参数格式，跳过
	}

	// 处理发送红包
	s.processRedPacket(ctx, b, message, chatID, userID, userName, amount, nil, thunder)
}

// processRedPacket 处理发送红包的通用逻辑
func (s *TelegramBotService) processRedPacket(ctx context.Context, b *bot.Bot, message *models.Message, chatID, userID int64, userName string, amount float64, number *int, thunder int) {
	// 解析命令
	result, err := s.HandleRedPacketCommand(chatID, userID, userName, message.Text)
	if err != nil {
		_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   err.Error(),
			ReplyParameters: &models.ReplyParameters{
				MessageID: message.ID,
				ChatID:    chatID,
			},
		})
		return
	}

	// 获取群组信息
	groupInfo, _ := s.GetGroupInfo(chatID)

	// 格式化消息
	messageText := FormatRedPacketMessage(result["senderName"].(string), result["amount"].(float64))
	imageURL := s.getSendPacketImage(chatID)

	// 构建内联键盘
	inlineKeyboard := s.buildInlineKeyboard(
		result["luckyId"].(int64),
		result["number"].(int),
		0,
		result["amount"].(float64),
		result["thunder"].(int),
		groupInfo,
	)

	// 发送消息
	_, err = b.SendPhoto(ctx, &bot.SendPhotoParams{
		ChatID:      chatID,
		Photo:       &models.InputFileString{Data: imageURL},
		Caption:     messageText,
		ParseMode:   models.ParseModeHTML,
		ReplyMarkup: inlineKeyboard,
	})
	if err != nil {
		log.Printf("发送红包消息失败: %v", err)
	}
}

// handleGroupInfoMessage 处理群信息查询
func (s *TelegramBotService) handleGroupInfoMessage(ctx context.Context, b *bot.Bot, update *models.Update) {
	message := update.Message
	if message == nil || message.From == nil {
		return
	}

	chatID := message.Chat.ID
	userID := int64(message.From.ID)

	// 检查群组授权
	if !s.CheckGroupAuth(chatID) {
		return
	}

	infoText := s.HandleGroupInfoCommand(chatID, userID)
	_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    chatID,
		Text:      infoText,
		ParseMode: models.ParseModeHTML,
		ReplyParameters: &models.ReplyParameters{
			MessageID: message.ID,
			ChatID:    chatID,
		},
	})
}

// handleBalanceMessage 处理余额查询
func (s *TelegramBotService) handleBalancePrivateMessage(ctx context.Context, b *bot.Bot, message *models.Message) {
	if message == nil || message.From == nil {
		return
	}

	userID := int64(message.From.ID)
	balanceText, err := s.HandleBalanceCommand(userID)
	if err != nil {
		balanceText = "未注册用户"
	}

	_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      message.Chat.ID,
		Text:        balanceText,
		ParseMode:   models.ParseModeHTML,
		ReplyMarkup: s.buildBalanceKeyboard(),
	})
}

func (s *TelegramBotService) handleBalanceGroupMessage(ctx context.Context, b *bot.Bot, message *models.Message) {
	if message == nil || message.From == nil {
		return
	}

	chatID := message.Chat.ID
	userID := int64(message.From.ID)

	// 检查群组授权
	if !s.CheckGroupAuth(chatID) {
		_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   "未授权",
		})
		return
	}

	balanceText, err := s.HandleBalanceCommand(userID)
	if err != nil {
		balanceText = "未注册用户"
	}

	_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    chatID,
		Text:      balanceText,
		ParseMode: models.ParseModeHTML,
		ReplyParameters: &models.ReplyParameters{
			MessageID: message.ID,
			ChatID:    chatID,
		},
		ReplyMarkup: s.buildBalanceKeyboard(),
	})
}

func (s *TelegramBotService) handleTeamMessage(ctx context.Context, b *bot.Bot, message *models.Message) {
	if message == nil || message.From == nil {
		return
	}

	chatID := message.Chat.ID
	if message.Chat.Type == "group" || message.Chat.Type == "supergroup" {
		if !s.CheckGroupAuth(chatID) {
			_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: chatID,
				Text:   "未授权",
			})
			return
		}
	}

	userID := int64(message.From.ID)
	user, err := s.GetTgUserByTelegramID(userID)
	if err != nil {
		_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   "未注册用户",
		})
		return
	}

	text, err := s.buildTeamMessage(user)
	if err != nil {
		_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   "获取团队信息失败",
		})
		return
	}

	_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      chatID,
		Text:        text,
		ReplyMarkup: s.buildPrivateEntryKeyboard(),
	})
}

func (s *TelegramBotService) handleRechargePrivateMessage(ctx context.Context, b *bot.Bot, message *models.Message) {
	if message == nil || message.From == nil {
		return
	}

	userID := int64(message.From.ID)

	user, err := s.GetTgUserByTelegramID(userID)
	if err != nil {
		_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: message.Chat.ID,
			Text:   "未注册用户",
		})
		return
	}

	displayName := formatTgUserDisplayName(user)
	text := fmt.Sprintf(`
用户名: %s
余额: %.3f R$`, displayName, user.Balance)

	_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      message.Chat.ID,
		Text:        text,
		ReplyMarkup: s.buildRechargeKeyboard(),
	})
}

func (s *TelegramBotService) handleRegisterTextMessage(ctx context.Context, b *bot.Bot, message *models.Message) {
	if message == nil || message.From == nil {
		return
	}
	update := &models.Update{Message: message}
	s.handleRegisterCommand(ctx, b, update)
}

func (s *TelegramBotService) handleWithdrawMessage(ctx context.Context, b *bot.Bot, message *models.Message) {
	if message == nil || message.From == nil {
		return
	}
	_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      message.Chat.ID,
		Text:        "提现功能开发中",
		ReplyMarkup: s.buildPrivateEntryKeyboard(),
	})
}

func (s *TelegramBotService) handleInviteMessage(ctx context.Context, b *bot.Bot, message *models.Message) {
	if message == nil || message.From == nil {
		return
	}
	user, err := s.GetTgUserByTelegramID(message.From.ID)
	if err != nil {
		_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:      message.Chat.ID,
			Text:        "未注册用户",
			ReplyMarkup: s.buildPrivateEntryKeyboard(),
		})
		return
	}
	_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      message.Chat.ID,
		Text:        s.buildInviteText(user),
		ReplyMarkup: s.buildPrivateEntryKeyboard(),
	})
}

func (s *TelegramBotService) handleRebateMessage(ctx context.Context, b *bot.Bot, message *models.Message) {
	if message == nil || message.From == nil {
		return
	}
	user, err := s.GetTgUserByTelegramID(int64(message.From.ID))
	if err != nil {
		_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:      message.Chat.ID,
			Text:        "未注册用户",
			ReplyMarkup: s.buildPrivateEntryKeyboard(),
		})
		return
	}
	_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      message.Chat.ID,
		Text:        s.buildCommissionText(user),
		ReplyMarkup: s.buildPrivateEntryKeyboard(),
	})
}

func (s *TelegramBotService) buildPrivateEntryKeyboard() *models.ReplyKeyboardMarkup {
	rows := [][]models.KeyboardButton{
		{
			{Text: "注册"},
			{Text: "充值"},
			{Text: "提现"},
		},
		{
			{Text: "团队"},
			{Text: "邀请"},
			{Text: "反水"},
		},
	}
	return &models.ReplyKeyboardMarkup{
		Keyboard:       rows,
		IsPersistent:   true,
		ResizeKeyboard: true,
	}
}

func (s *TelegramBotService) buildRechargeKeyboard() *models.InlineKeyboardMarkup {
	rows := [][]models.InlineKeyboardButton{
		{
			{Text: "10R$", CallbackData: "recharge_amount_10"},
			{Text: "20R$", CallbackData: "recharge_amount_20"},
			{Text: "50R$", CallbackData: "recharge_amount_50"},
		},
		{
			{Text: "100R$", CallbackData: "recharge_amount_100"},
			{Text: "200R$", CallbackData: "recharge_amount_200"},
			{Text: "300R$", CallbackData: "recharge_amount_300"},
		},
		{
			{Text: "1000R$", CallbackData: "recharge_amount_1000"},
			{Text: "2000R$", CallbackData: "recharge_amount_2000"},
			{Text: "3000R$", CallbackData: "recharge_amount_3000"},
		},
		{
			{Text: "自定义金额", CallbackData: "recharge_custom"},
		},
		{
			{Text: "❌ 关闭", CallbackData: "recharge_close"},
		},
	}

	return &models.InlineKeyboardMarkup{InlineKeyboard: rows}
}

func (s *TelegramBotService) handleRechargeAmountCallback(ctx context.Context, b *bot.Bot, update *models.Update) {
	query := update.CallbackQuery
	if query == nil {
		return
	}

	amount := strings.TrimPrefix(query.Data, "recharge_amount_")
	amountValue, err := strconv.ParseFloat(amount, 64)
	if err != nil || amountValue <= 0 {
		_, _ = b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: query.ID,
			Text:            "金额格式错误",
			ShowAlert:       true,
		})
		return
	}

	orderNo, err := s.processRechargeAmount(ctx, int64(query.From.ID), amountValue)
	if err != nil {
		_, _ = b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: query.ID,
			Text:            err.Error(),
			ShowAlert:       true,
		})
		return
	}

	_, _ = b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: query.ID,
		Text:            fmt.Sprintf("充值成功，订单号：%s", orderNo),
		ShowAlert:       true,
	})
}

func (s *TelegramBotService) tryHandleCustomRecharge(ctx context.Context, b *bot.Bot, message *models.Message, text string) bool {
	amountValue, err := strconv.ParseFloat(strings.TrimSpace(text), 64)
	if err != nil || amountValue <= 0 {
		return false
	}

	redisKey := fmt.Sprintf("tg_recharge_custom:%d", message.From.ID)
	ts, err := utils.RD.Get(ctx, redisKey).Int64()
	if err != nil {
		return false
	}
	if time.Now().Unix()-ts > 15*60 {
		_ = utils.RD.Del(ctx, redisKey).Err()
		return false
	}
	_ = utils.RD.Del(ctx, redisKey).Err()

	orderNo, err := s.processRechargeAmount(ctx, int64(message.From.ID), amountValue)
	if err != nil {
		_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: message.Chat.ID,
			Text:   err.Error(),
		})
		return true
	}

	_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: message.Chat.ID,
		Text:   fmt.Sprintf("充值成功，订单号：%s", orderNo),
	})
	return true
}

func (s *TelegramBotService) processRechargeAmount(ctx context.Context, telegramUserID int64, amountValue float64) (string, error) {
	user, err := s.GetTgUserByTelegramID(telegramUserID)
	if err != nil {
		return "", fmt.Errorf("未注册用户")
	}

	db := utils.NewPrefixDb(s.TablePrefix)
	orderNo := s.generateRechargeOrderNo(user.ID)
	rechargeOrder := pojo.RechargeOrder{
		TenantId:    0,
		UserId:      user.ID,
		OrderNo:     orderNo,
		Channel:     "telegram",
		PayMethod:   strPtr("manual"),
		Currency:    "BRL",
		Amount:      amountValue,
		Fee:         0,
		BonusAmount: 0,
		Status:      0,
		Title:       strPtr("Telegram 充值"),
		Remark:      strPtr(fmt.Sprintf("tg:%d", telegramUserID)),
	}

	if err := db.Create(&rechargeOrder).Error; err != nil {
		return "", fmt.Errorf("创建充值订单失败")
	}

	if err := s.manualRechargeCallback(db, rechargeOrder.ID); err != nil {
		return "", fmt.Errorf("充值回调失败")
	}

	return orderNo, nil
}

func (s *TelegramBotService) manualRechargeCallback(db *gorm.DB, orderID int64) error {
	tx := db.Begin()
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
		}
	}()

	var order pojo.RechargeOrder
	if err := tx.Where("id = ?", orderID).First(&order).Error; err != nil {
		tx.Rollback()
		return err
	}
	if order.Status != 0 {
		tx.Rollback()
		return fmt.Errorf("订单状态异常")
	}

	var user pojo.TgUser
	if err := tx.Where("id = ?", order.UserId).First(&user).Error; err != nil {
		tx.Rollback()
		return err
	}
	isFirstRecharge := user.RechargeAmount <= 0

	startAmount := user.Balance
	endAmount := utils.ToMoney(startAmount).Add(utils.ToMoney(order.Amount)).ToDollars()
	cashHistory := pojo.CashHistory{
		UserId:      user.ID,
		AwardUni:    order.OrderNo,
		Amount:      order.Amount,
		StartAmount: startAmount,
		EndAmount:   endAmount,
		CashMark:    "充值到账",
		CashDesc:    fmt.Sprintf("订单号:%s", order.OrderNo),
		FromUserId:  0,
	}
	if err := tx.Create(&cashHistory).Error; err != nil {
		tx.Rollback()
		return err
	}

	updateUser := map[string]any{
		"balance":         gorm.Expr(fmt.Sprintf("balance + %.3f", order.Amount)),
		"recharge_amount": gorm.Expr("recharge_amount + ?", order.Amount),
	}
	if err := tx.Model(&pojo.TgUser{}).Where("id = ?", user.ID).Updates(updateUser).Error; err != nil {
		tx.Rollback()
		return err
	}

	updateOrder := map[string]any{
		"status":          8,
		"pay_time":        time.Now(),
		"notify_time":     time.Now(),
		"credit_amount":   order.Amount,
		"provider_status": "manual_callback",
	}
	if err := tx.Model(&pojo.RechargeOrder{}).Where("id = ?", order.ID).Updates(updateOrder).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := s.applyRechargeRebate(tx, order, user); err != nil {
		tx.Rollback()
		return err
	}
	if isFirstRecharge {
		if err := s.applyInviteFirstRechargeReward(tx, order, user, time.Now()); err != nil {
			tx.Rollback()
			return err
		}
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}

	user.Balance = endAmount
	return nil
}

func (s *TelegramBotService) applyInviteFirstRechargeReward(tx *gorm.DB, order pojo.RechargeOrder, user pojo.TgUser, now time.Time) error {
	if user.ParentID == nil || *user.ParentID == 0 {
		return nil
	}

	rate := s.getRebateRate("invite_first_recharge_reward")
	if rate <= 0 {
		return nil
	}

	parentID := *user.ParentID
	var parent pojo.TgUser
	if err := tx.Where("id = ?", parentID).First(&parent).Error; err != nil || parent.ID == 0 {
		return nil
	}
	if parent.Status != 1 {
		return nil
	}

	rebateAmount := utils.ToMoney(order.Amount).Multiply(rate / 100).ToDollars()
	if rebateAmount <= 0 {
		return nil
	}

	idempotencyKey := fmt.Sprintf("first_recharge_reward:%s:%d", order.OrderNo, parentID)
	var existing pojo.TgUserRebateRecord
	if err := tx.Where("idempotency_key = ?", idempotencyKey).First(&existing).Error; err == nil && existing.ID > 0 {
		return nil
	}

	currency := strings.TrimSpace(order.Currency)
	if currency == "" {
		currency = "USDT"
	}
	remark := strPtr("first_recharge_reward")
	record := pojo.TgUserRebateRecord{
		TenantId:       &order.TenantId,
		SubUserId:      user.ID,
		ParentUserId:   parentID,
		SourceType:     5,
		SourceOrderId:  order.OrderNo,
		SourceAmount:   order.Amount,
		RebateRate:     rate,
		RebateAmount:   rebateAmount,
		Currency:       currency,
		Status:         1,
		SettledAt:      &now,
		IdempotencyKey: idempotencyKey,
		Remark:         remark,
	}
	if err := tx.Create(&record).Error; err != nil {
		return err
	}

	return tx.Model(&pojo.TgUser{}).
		Where("id = ?", parentID).
		Updates(map[string]any{
			"rebate_amount":       gorm.Expr("rebate_amount + ?", rebateAmount),
			"rebate_total_amount": gorm.Expr("rebate_total_amount + ?", rebateAmount),
		}).Error
}

func (s *TelegramBotService) generateRechargeOrderNo(userID int64) string {
	return fmt.Sprintf("RC%s%06d", time.Now().Format("20060102150405"), rand.IntN(1000000))
}

func strPtr(value string) *string {
	return &value
}

func (s *TelegramBotService) handleRechargeCustomCallback(ctx context.Context, b *bot.Bot, update *models.Update) {
	query := update.CallbackQuery
	if query == nil {
		return
	}

	redisKey := fmt.Sprintf("tg_recharge_custom:%d", query.From.ID)
	_ = utils.RD.Set(ctx, redisKey, time.Now().Unix(), 15*time.Minute).Err()

	_, _ = b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: query.ID,
		Text:            "请发送自定义充值金额（R$）",
		ShowAlert:       true,
	})
}

func (s *TelegramBotService) handleRechargeUsdtCallback(ctx context.Context, b *bot.Bot, update *models.Update) {
	query := update.CallbackQuery
	if query == nil {
		return
	}

	_, _ = b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: query.ID,
		Text:            "已切换到USDT充值",
		ShowAlert:       true,
	})
}

func (s *TelegramBotService) handleRechargeCloseCallback(ctx context.Context, b *bot.Bot, update *models.Update) {
	query := update.CallbackQuery
	if query == nil {
		return
	}

	redisKey := fmt.Sprintf("tg_recharge_custom:%d", query.From.ID)
	_ = utils.RD.Del(ctx, redisKey).Err()

	_, _ = b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: query.ID,
		Text:            "已关闭",
	})
}

// handleGrabCallback 处理抢红包回调
func (s *TelegramBotService) handleGrabCallback(ctx context.Context, b *bot.Bot, update *models.Update) {
	query := update.CallbackQuery
	if query == nil {
		return
	}

	chatID := query.Message.Message.Chat.ID
	userID := int64(query.From.ID)
	callbackData := query.Data

	// 解析 luckyID 和数量（可选）
	luckyIDStr := strings.TrimPrefix(callbackData, "qiang-")
	parts := strings.Split(luckyIDStr, "-")
	if len(parts) == 0 {
		_, _ = b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: query.ID,
			Text:            "无效的红包ID",
			ShowAlert:       true,
		})
		return
	}
	luckyID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		_, _ = b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: query.ID,
			Text:            "无效的红包ID",
			ShowAlert:       true,
		})
		return
	}
	grabIndex := 1
	if len(parts) >= 2 {
		if cnt, parseErr := strconv.Atoi(parts[1]); parseErr == nil && cnt > 0 {
			grabIndex = cnt
		}
	}

	// 使用 Redis 分布式锁，根据红包ID+第几个包加锁，锁键格式：bgu_tg_grab_{luckyID}_{index}
	lockKey := fmt.Sprintf("bgu_tg_grab_%d_%d", luckyID, grabIndex)
	lockTimeout := 5 * time.Second // 锁超时时间5秒

	// 尝试获取锁
	acquired, err := utils.AcquireLock(lockKey, lockTimeout)
	if err != nil {
		log.Printf("获取Redis锁失败: %v", err)
		// Redis 锁获取失败，继续执行（降级处理）
	} else if !acquired {
		_, _ = b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: query.ID,
			Text:            "该红包正在被处理，请稍后再试",
			ShowAlert:       true,
		})
		return
	}
	if acquired {
		defer func() {
			_ = utils.ReleaseLock(lockKey)
		}()
	}

	// 检查红包状态与第几个包
	luckyMoney, grabbedCount, err := services.GetRedPacketStatus(s.DB, luckyID)
	if err != nil {
		_, _ = b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: query.ID,
			Text:            "红包状态获取失败",
			ShowAlert:       true,
		})
		return
	}
	if grabIndex > luckyMoney.Number || grabIndex <= 0 {
		_, _ = b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: query.ID,
			Text:            "红包数量不足",
			ShowAlert:       true,
		})
		return
	}
	if int(grabbedCount) >= luckyMoney.Number {
		_, _ = b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: query.ID,
			Text:            "该红包已全部被领取",
			ShowAlert:       true,
		})
		return
	}
	claimedKey := fmt.Sprintf("bgu_tg_grab_done_%d_%d", luckyID, grabIndex)
	if data := utils.RD.Get(ctx, claimedKey); data != nil && data.Err() == nil && data.Val() != "" {
		_, _ = b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: query.ID,
			Text:            fmt.Sprintf("第%d包已被领取", grabIndex),
			ShowAlert:       true,
		})
		return
	}
	// 处理抢红包
	result, err := s.HandleGrabCallback(chatID, userID, luckyID, grabIndex)
	if err != nil {
		_, _ = b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: query.ID,
			Text:            err.Error(),
			ShowAlert:       true,
		})
		return
	}
	amount, _ := result["amount"].(float64)
	_ = utils.RD.SetEX(ctx, claimedKey, fmt.Sprintf("%.2f", amount), 48*time.Hour).Err()
	isThunder, _ := result["isThunder"].(int)
	loseMoney, _ := result["loseMoney"].(float64)

	// 显示结果
	msg := ""
	if isThunder == 1 {
		msg = fmt.Sprintf("第%d包：金额%.2fU，中雷损失%.2fU", grabIndex, amount, loseMoney)
	} else {
		msg = fmt.Sprintf("第%d包：金额%.2fU，未中雷", grabIndex, amount)
	}
	_, _ = b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: query.ID,
		Text:            msg,
		ShowAlert:       true,
	})

	// 更新消息
	s.updateRedPacketMessage(ctx, b, query.Message.Message, luckyID, result)
}

// handleBalanceCallback 处理余额查询回调
func (s *TelegramBotService) handleBalanceCallback(ctx context.Context, b *bot.Bot, update *models.Update) {
	query := update.CallbackQuery
	if query == nil {
		return
	}

	userID := int64(query.From.ID)
	userName := query.From.FirstName
	if userName == "" {
		userName = query.From.Username
	}

	balanceText, err := s.HandleBalanceCommand(userID)
	if err != nil {
		balanceText = err.Error()
	}

	_, _ = b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: query.ID,
		Text:            fmt.Sprintf("[ %s ]\n-----------------------------\n%s", userName, balanceText),
		ShowAlert:       true,
	})
}

func (s *TelegramBotService) handleBalanceActionCallback(ctx context.Context, b *bot.Bot, update *models.Update) {
	query := update.CallbackQuery
	if query == nil || query.Message.Message == nil {
		return
	}

	chatID := query.Message.Message.Chat.ID
	userID := int64(query.From.ID)
	action := strings.TrimPrefix(query.Data, "balance_")

	switch action {
	case "recharge":
		if err := s.sendRechargePanel(ctx, b, chatID, userID); err != nil {
			_, _ = b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
				CallbackQueryID: query.ID,
				Text:            err.Error(),
				ShowAlert:       true,
			})
			return
		}
		if query.Message.Message.Chat.Type == "group" || query.Message.Message.Chat.Type == "supergroup" {
			_ = s.sendRechargePanel(ctx, b, userID, userID)
		}
	case "team":
		user, err := s.GetTgUserByTelegramID(userID)
		if err != nil {
			_, _ = b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
				CallbackQueryID: query.ID,
				Text:            "未注册用户",
				ShowAlert:       true,
			})
			return
		}
		text, err := s.buildTeamMessage(user)
		if err != nil {
			_, _ = b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
				CallbackQueryID: query.ID,
				Text:            "获取团队信息失败",
				ShowAlert:       true,
			})
			return
		}
		_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   text,
		})
	case "invite":
		user, err := s.GetTgUserByTelegramID(userID)
		if err != nil {
			_, _ = b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
				CallbackQueryID: query.ID,
				Text:            "未注册用户",
				ShowAlert:       true,
			})
			return
		}
		_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   s.buildInviteText(user),
		})
	case "commission":
		user, err := s.GetTgUserByTelegramID(userID)
		if err != nil {
			_, _ = b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
				CallbackQueryID: query.ID,
				Text:            "未注册用户",
				ShowAlert:       true,
			})
			return
		}
		text := s.buildCommissionText(user)
		_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   text,
		})
	case "cash_flow":
		user, err := s.GetTgUserByTelegramID(userID)
		if err != nil {
			_, _ = b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
				CallbackQueryID: query.ID,
				Text:            "未注册用户",
				ShowAlert:       true,
			})
			return
		}
		text, err := s.buildCashFlowText(user.ID)
		if err != nil {
			_, _ = b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
				CallbackQueryID: query.ID,
				Text:            "获取流水失败",
				ShowAlert:       true,
			})
			return
		}
		_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   text,
		})
	default:
		_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   "功能开发中",
		})
	}

	_, _ = b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: query.ID,
		Text:            "操作成功",
	})
}

// handleTodayDataCallback 处理今日报表回调
func (s *TelegramBotService) handleTodayDataCallback(ctx context.Context, b *bot.Bot, update *models.Update) {
	query := update.CallbackQuery
	if query == nil {
		return
	}

	// TODO: 实现今日报表功能
	_, _ = b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: query.ID,
		Text:            "今日报表功能开发中",
		ShowAlert:       true,
	})
}

// handleShareDataCallback 处理推广查询回调
func (s *TelegramBotService) handleShareDataCallback(ctx context.Context, b *bot.Bot, update *models.Update) {
	query := update.CallbackQuery
	if query == nil {
		return
	}

	// TODO: 实现推广查询功能
	_, _ = b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: query.ID,
		Text:            "推广查询功能开发中",
		ShowAlert:       true,
	})
}

func (s *TelegramBotService) sendRechargePanel(ctx context.Context, b *bot.Bot, chatID int64, userID int64) error {
	user, err := s.GetTgUserByTelegramID(userID)
	if err != nil {
		return fmt.Errorf("未注册用户")
	}
	displayName := formatTgUserDisplayName(user)
	text := fmt.Sprintf("用户名: %s\n余额: %.3f R$", displayName, user.Balance)
	_, err = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      chatID,
		Text:        text,
		ReplyMarkup: s.buildRechargeKeyboard(),
	})
	return err
}

func (s *TelegramBotService) buildTeamMessage(user *pojo.TgUser) (string, error) {
	start := startOfToday(time.Now())
	end := start.Add(24 * time.Hour)

	var inviteTotal int64
	s.DB.Model(&pojo.TgUser{}).Where("parent_id = ?", user.ID).Count(&inviteTotal)

	var inviteToday int64
	s.DB.Model(&pojo.TgUser{}).
		Where("parent_id = ?", user.ID).
		Where("created_at >= ? AND created_at < ?", start, end).
		Count(&inviteToday)

	subIDs := s.getDirectSubUserIDs(user.ID)
	var rechargeTotal int64
	var rechargeToday int64
	if len(subIDs) > 0 {
		s.DB.Model(&pojo.RechargeOrder{}).
			Where("status = ?", 8).
			Where("user_id IN ?", subIDs).
			Distinct("user_id").
			Count(&rechargeTotal)

		s.DB.Model(&pojo.RechargeOrder{}).
			Where("status = ?", 8).
			Where("user_id IN ?", subIDs).
			Where("pay_time >= ? AND pay_time < ?", start, end).
			Distinct("user_id").
			Count(&rechargeToday)
	}

	rebateToday, err := s.getRebateTodayAmount(user.ID)
	if err != nil {
		return "", err
	}

	displayName := formatTgUserDisplayName(user)
	text := fmt.Sprintf(`👥 我的团队
用户：%s
邀请用户数：%d
充值用户数：%d
今日邀请用户数：%d
今日充值用户数：%d
累计返水金额：%.3f
可用返水金额：%.3f
今日返水金额：%.3f`, displayName, inviteTotal, rechargeTotal, inviteToday, rechargeToday, user.RebateTotalAmount, user.RebateAmount, rebateToday)
	return text, nil
}

func (s *TelegramBotService) buildInviteText(user *pojo.TgUser) string {
	inviteCode := ""
	if user.InviteCode != nil {
		inviteCode = *user.InviteCode
	}
	inviteBaseURL := utils.GlobalConfig.Telegram.InviteBaseURL
	if inviteBaseURL == "" {
		inviteBaseURL = "https://t.me/goodLuckEveryOne66Bot/?start="
	}
	inviteLink := fmt.Sprintf("%s%s", inviteBaseURL, inviteCode)
	return fmt.Sprintf("邀请链接：%s", inviteLink)
}

func (s *TelegramBotService) buildCommissionText(user *pojo.TgUser) string {
	rebateToday, _ := s.getRebateTodayAmount(user.ID)
	return fmt.Sprintf(`佣金明细
累计返水金额：%.3f
可用返水金额：%.3f
今日返水金额：%.3f`, user.RebateTotalAmount, user.RebateAmount, rebateToday)
}

func (s *TelegramBotService) getRebateTodayAmount(parentUserID int64) (float64, error) {
	start := startOfToday(time.Now())
	end := start.Add(24 * time.Hour)
	var rebateToday float64
	err := s.DB.Model(&pojo.TgUserRebateRecord{}).
		Select("COALESCE(SUM(rebate_amount), 0)").
		Where("parent_user_id = ?", parentUserID).
		Where("settled_at >= ? AND settled_at < ?", start, end).
		Scan(&rebateToday).Error
	if err != nil {
		return 0, err
	}
	return rebateToday, nil
}

func (s *TelegramBotService) getDirectSubUserIDs(parentID int64) []int64 {
	subIDs := make([]int64, 0)
	s.DB.Model(&pojo.TgUser{}).Where("parent_id = ?", parentID).Pluck("id", &subIDs)
	return subIDs
}

func (s *TelegramBotService) getSendPacketImage(chatID int64) string {
	redisKey := fmt.Sprintf("bgu_auth_group_send_packet_image_%d", chatID)
	imageURL, err := utils.RD.Get(context.Background(), redisKey).Result()
	if err == nil && imageURL != "" {
		return imageURL
	}

	var authGroup pojo.AuthGroup
	if err := s.DB.Where("group_id = ?", chatID).First(&authGroup).Error; err == nil {
		if authGroup.SendPacketImage != "" {
			_ = utils.RD.SetEX(context.Background(), redisKey, authGroup.SendPacketImage, utils.GetRandomRangeSecond(20*60, 40*60)).Err()
			return authGroup.SendPacketImage
		}
	}

	defaultURL := "https://th.bing.com/th/id/OIP.SwXBL3takNL_f6IRmiFUbgHaIx?w=150&h=180&c=7&r=0&o=7&cb=defcache2&dpr=1.5&pid=1.7&rm=3&defcache=1"
	_ = utils.RD.SetEX(context.Background(), redisKey, defaultURL, utils.GetRandomRangeSecond(20*60, 40*60)).Err()
	return defaultURL
}

// updateRedPacketMessage 更新红包消息
func (s *TelegramBotService) updateRedPacketMessage(ctx context.Context, b *bot.Bot, message *models.Message, luckyID int64, result map[string]interface{}) {
	chatID := message.Chat.ID
	messageID := message.ID

	// 获取红包状态
	luckyMoney, grabbedCount, err := services.GetRedPacketStatus(s.DB, luckyID)
	if err != nil {
		log.Printf("获取红包状态失败: %v", err)
		return
	}

	// 获取群组信息
	groupInfo, _ := s.GetGroupInfo(chatID)

	// 检查是否全部领取完成
	if int(grabbedCount) >= luckyMoney.Number {
		// 获取领取历史
		historyList, err := repository.GetLuckyHistoryByLuckyId(s.DB, luckyID)
		if err != nil {
			log.Printf("获取领取历史失败: %v", err)
			return
		}

		// 格式化完成消息
		completeText := FormatRedPacketCompleteMessage(luckyMoney, historyList)

		// 构建键盘（只有其他按钮，没有抢红包按钮）
		inlineKeyboard := s.buildInlineKeyboard(luckyID, luckyMoney.Number, int(grabbedCount), luckyMoney.Amount, luckyMoney.Thunder, groupInfo)

		// 尝试编辑消息
		_, err = b.EditMessageCaption(ctx, &bot.EditMessageCaptionParams{
			ChatID:      chatID,
			MessageID:   messageID,
			Caption:     completeText,
			ParseMode:   models.ParseModeHTML,
			ReplyMarkup: inlineKeyboard,
		})
		if err != nil {
			log.Printf("编辑消息失败: %v", err)
		}
	} else {
		// 更新抢红包按钮
		inlineKeyboard := s.buildInlineKeyboard(luckyID, luckyMoney.Number, int(grabbedCount), luckyMoney.Amount, luckyMoney.Thunder, groupInfo)
		messageText := FormatRedPacketMessage(luckyMoney.SenderName, luckyMoney.Amount)

		// 尝试编辑消息文本
		_, err = b.EditMessageCaption(ctx, &bot.EditMessageCaptionParams{
			ChatID:      chatID,
			MessageID:   messageID,
			Caption:     messageText,
			ParseMode:   models.ParseModeHTML,
			ReplyMarkup: inlineKeyboard,
		})
		if err != nil {
			// 如果编辑文本失败，尝试只编辑键盘
			_, _ = b.EditMessageReplyMarkup(ctx, &bot.EditMessageReplyMarkupParams{
				ChatID:      chatID,
				MessageID:   messageID,
				ReplyMarkup: inlineKeyboard,
			})
		}
	}
}

// buildInlineKeyboard 构建内联键盘
func (s *TelegramBotService) buildInlineKeyboard(luckyID int64, number int, grabbedCount int, amount float64, thunder int, groupInfo *pojo.AuthGroup) *models.InlineKeyboardMarkup {
	var rows [][]models.InlineKeyboardButton

	// 抢红包按钮
	if grabbedCount < number {
		buttonText := fmt.Sprintf("🧧抢红包[%d/%d]总%.0fU 💥雷%d", number, grabbedCount, amount, thunder)
		rows = append(rows, []models.InlineKeyboardButton{
			{
				Text:         buttonText,
				CallbackData: fmt.Sprintf("qiang-%d", luckyID),
			},
		})
		// 快捷抢指定数量
		//remain := number - grabbedCount
		quickCounts := []int{1, 2, 3, 4, 5, 6}
		var quickRow []models.InlineKeyboardButton
		for _, c := range quickCounts {
			if c <= 0 || c > number {
				continue
			}
			buttonText := fmt.Sprintf("第%d包", c)
			claimedKey := fmt.Sprintf("bgu_tg_grab_done_%d_%d", luckyID, c)
			if data := utils.RD.Get(context.Background(), claimedKey); data != nil && data.Err() == nil && data.Val() != "" {
				buttonText = fmt.Sprintf("第%d包✅%sU", c, data.Val())
			}
			quickRow = append(quickRow, models.InlineKeyboardButton{
				Text:         buttonText,
				CallbackData: fmt.Sprintf("qiang-%d-%d", luckyID, c),
			})
			if len(quickRow) == 3 {
				rows = append(rows, quickRow)
				quickRow = nil
			}
		}
		if len(quickRow) > 0 {
			rows = append(rows, quickRow)
		}
	}

	// 其他按钮行
	var row []models.InlineKeyboardButton
	if groupInfo != nil {
		if groupInfo.ServiceURL != "" {
			row = append(row, models.InlineKeyboardButton{
				Text: "客服",
				URL:  groupInfo.ServiceURL,
			})
		}
		if groupInfo.RechargeURL != "" {
			row = append(row, models.InlineKeyboardButton{
				Text: "充值",
				URL:  groupInfo.RechargeURL,
			})
		}
		if groupInfo.ChannelURL != "" {
			row = append(row, models.InlineKeyboardButton{
				Text: "玩法",
				URL:  groupInfo.ChannelURL,
			})
		}
	}
	row = append(row, models.InlineKeyboardButton{
		Text:         "余额",
		CallbackData: "balance",
	})
	if len(row) > 0 {
		rows = append(rows, row)
	}

	// 第二行
	rows = append(rows, []models.InlineKeyboardButton{
		{
			Text:         "推广查询",
			CallbackData: "share_data",
		},
		{
			Text:         "今日报表",
			CallbackData: "today_data",
		},
	})

	return &models.InlineKeyboardMarkup{
		InlineKeyboard: rows,
	}
}

func (s *TelegramBotService) buildBalanceKeyboard() *models.InlineKeyboardMarkup {
	inviteBaseURL := utils.GlobalConfig.Telegram.InviteBaseURL
	if inviteBaseURL == "" {
		inviteBaseURL = "https://t.me/goodLuckEveryOne66Bot"
	}
	rows := [][]models.InlineKeyboardButton{
		{
			{Text: "充值", URL: inviteBaseURL},
			{Text: "提现", CallbackData: "balance_withdraw"},
			{Text: "提现账户", CallbackData: "balance_withdraw_account"},
			{Text: "幸运转盘", CallbackData: "balance_lucky"},
		},
		{
			{Text: "我的团队", CallbackData: "balance_team"},
			{Text: "邀请好友", CallbackData: "balance_invite"},
			{Text: "佣金明细", CallbackData: "balance_commission"},
			{Text: "流水明细", CallbackData: "balance_cash_flow"},
		},
		{
			{Text: "语言", CallbackData: "balance_language"},
			{Text: "游戏规则", CallbackData: "balance_rules"},
		},
	}

	return &models.InlineKeyboardMarkup{InlineKeyboard: rows}
}

func (s *TelegramBotService) buildCashFlowText(userID int64) (string, error) {
	now := time.Now()
	todayStart := startOfToday(now)
	yesterdayStart := todayStart.AddDate(0, 0, -1)
	tomorrowStart := todayStart.AddDate(0, 0, 1)

	todayTotal, err := s.sumCashHistoryAbsAmount(userID, &todayStart, &tomorrowStart)
	if err != nil {
		return "", err
	}

	yesterdayTotal, err := s.sumCashHistoryAbsAmount(userID, &yesterdayStart, &todayStart)
	if err != nil {
		return "", err
	}

	allTotal, err := s.sumCashHistoryAbsAmount(userID, nil, nil)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("流水明细\n今日流水：%.2f U\n昨日流水：%.2f U\n总流水：%.2f U", todayTotal, yesterdayTotal, allTotal), nil
}

func (s *TelegramBotService) sumCashHistoryAbsAmount(userID int64, start *time.Time, end *time.Time) (float64, error) {
	query := s.DB.Model(&pojo.CashHistory{}).Where("user_id = ?", userID)
	if start != nil {
		query = query.Where("created_at >= ?", *start)
	}
	if end != nil {
		query = query.Where("created_at < ?", *end)
	}

	var total float64
	if err := query.Select("COALESCE(SUM(ABS(amount)), 0)").Scan(&total).Error; err != nil {
		return 0, err
	}

	return total, nil
}

// ParseRedPacketCommand 解析红包命令
// 支持格式：
// - 发10-1 或 10-1 (金额-雷)
// - 发10-3-1 或 10-3-1 (金额-数量-雷)
func ParseRedPacketCommand(text string) (amount float64, number *int, thunder int, ok bool) {
	// 格式1: 发10-3-1 或 10-3-1 (金额-数量-雷)
	pattern1 := regexp.MustCompile(`(?i)(?:发[包]*)?(\d+)[-/](\d+)[-/](\d+)`)
	matches1 := pattern1.FindStringSubmatch(text)
	if len(matches1) == 4 {
		amount, _ = strconv.ParseFloat(matches1[1], 64)
		num, _ := strconv.Atoi(matches1[2])
		thunder, _ := strconv.Atoi(matches1[3])
		if amount >= 5 && thunder >= 0 && thunder <= 9 {
			return amount, &num, thunder, true
		}
	}

	// 格式2: 发10-1 或 10-1 (金额-雷)
	pattern2 := regexp.MustCompile(`(?i)(?:发[包]*)?(\d+)[-/]([0-9])`)
	matches2 := pattern2.FindStringSubmatch(text)
	if len(matches2) == 3 {
		amount, _ = strconv.ParseFloat(matches2[1], 64)
		thunder, _ = strconv.Atoi(matches2[2])
		if amount >= 5 && thunder >= 0 && thunder <= 9 {
			return amount, nil, thunder, true
		}
	}

	return 0, nil, 0, false
}

// HandleRedPacketCommand 处理发送红包命令
func (s *TelegramBotService) HandleRedPacketCommand(chatID int64, userID int64, userName string, text string) (map[string]interface{}, error) {
	amount, number, thunder, ok := ParseRedPacketCommand(text)
	if !ok {
		return nil, fmt.Errorf("指令格式错误，正确格式：发10-1 或 发10-3-1")
	}

	// 使用 Redis 分布式锁，根据群ID加锁，锁键格式：bgu_tg_send_{chatID}
	lockKey := fmt.Sprintf("bgu_tg_send_%d", chatID)
	lockTimeout := 5 * time.Second // 锁超时时间5秒

	// 尝试获取锁
	acquired, err := utils.AcquireLock(lockKey, lockTimeout)
	if err != nil {
		log.Printf("获取Redis锁失败: %v", err)
		// Redis 锁获取失败，继续执行（降级处理）
	} else if !acquired {
		// 锁已被其他进程持有，返回错误提示
		return nil, fmt.Errorf("群组内正在处理其他红包操作，请稍后再试")
	}

	// 确保在函数返回时释放锁
	defer func() {
		if acquired {
			_ = utils.ReleaseLock(lockKey)
		}
	}()

	// 获取或创建用户
	user, err := s.GetOrCreateTgUserByTelegramID(userID, userName, chatID)
	if err != nil {
		return nil, fmt.Errorf("用户未注册，请先注册: %v", err)
	}

	// 构建发送红包请求
	req := pojo.LuckyMoneySend{
		Amount:  amount,
		Thunder: thunder,
		Number:  number,
		ChatID:  chatID,
	}

	// 调用服务层发送红包（使用系统用户ID）
	luckyMoney, err := services.SendRedPacket(s.DB, user.ID, userName, req, s.TablePrefix)
	if err != nil {
		return nil, err
	}

	// 构建返回结果
	result := map[string]interface{}{
		"luckyId":    luckyMoney.ID,
		"amount":     luckyMoney.Amount,
		"number":     luckyMoney.Number,
		"thunder":    luckyMoney.Thunder,
		"senderName": luckyMoney.SenderName,
	}

	return result, nil
}

// HandleGrabCallback 处理抢红包回调
func (s *TelegramBotService) HandleGrabCallback(chatID int64, userID int64, luckyID int64, grabIndex int) (map[string]interface{}, error) {
	// 获取或创建用户
	user, err := s.GetOrCreateTgUserByTelegramID(userID, "", chatID)
	if err != nil {
		return nil, fmt.Errorf("用户未注册，请先注册: %v", err)
	}

	// 调用服务层抢红包（使用系统用户ID）
	result, err := services.GrabRedPacket(s.DB, luckyID, user.ID, s.TablePrefix, grabIndex)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// HandleBalanceCommand 处理余额查询命令
func (s *TelegramBotService) HandleBalanceCommand(userID int64) (string, error) {
	user, err := s.GetTgUserByTelegramID(userID)
	if err != nil {
		return "", fmt.Errorf("用户未注册")
	}
	displayName := formatTgUserDisplayName(user)
	inviteCode := ""
	if user.InviteCode != nil {
		inviteCode = *user.InviteCode
	}
	inviteBaseURL := utils.GlobalConfig.Telegram.InviteBaseURL
	if inviteBaseURL == "" {
		inviteBaseURL = "https://t.me/goodLuckEveryOne66Bot/?start="
	}
	inviteLink := fmt.Sprintf("%s%s", inviteBaseURL, inviteCode)
	return fmt.Sprintf("用户：%s\n余额：%.3f U\n邀请链接：%s", displayName, user.Balance, inviteLink), nil
}

// HandleRegisterCommand 处理注册命令
func (s *TelegramBotService) HandleRegisterCommand(chatID int64, userID int64, userName string, userUsername string) (string, error) {
	_ = chatID
	var tgUser pojo.TgUser
	err := s.DB.Where("tg_id = ?", userID).First(&tgUser).Error
	if err == nil && tgUser.ID > 0 {
		statusText := formatTgUserStatus(tgUser.Status)
		return fmt.Sprintf(`✅ 您已注册！

👤 用户ID：<code>%d</code>
💰 余额：<code>%.3f U</code>
📊 状态：<code>%s</code>
📅 注册时间：<code>%s</code>`,
			tgUser.ID,
			tgUser.Balance,
			statusText,
			tgUser.CreatedAt.Format("2006-01-02 15:04:05")), nil
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return "", fmt.Errorf("注册失败: %v", err)
	}

	parentID := s.getParentIDByInviteCode(userID)
	inviteCode, err := s.generateInviteCode()
	if err != nil {
		return "", fmt.Errorf("注册失败: %v", err)
	}

	displayName := userName
	if displayName == "" {
		displayName = userUsername
	}
	if displayName == "" {
		displayName = fmt.Sprintf("User_%d", userID)
	}
	usernamePtr := strPtr(userUsername)
	if userUsername == "" {
		usernamePtr = nil
	}
	firstNamePtr := strPtr(displayName)
	if displayName == "" {
		firstNamePtr = nil
	}

	registerGiftAmount := s.getRegisterGiftAmount()
	newUser := pojo.TgUser{
		Username:   usernamePtr,
		FirstName:  firstNamePtr,
		TgID:       userID,
		Balance:    registerGiftAmount,
		GiftAmount: registerGiftAmount,
		GiftTotal:  registerGiftAmount,
		Status:     1,
		ParentID:   parentID,
		InviteCode: &inviteCode,
	}
	if err = s.DB.Create(&newUser).Error; err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") || strings.Contains(err.Error(), "1062") {
			var retryUser pojo.TgUser
			retryErr := s.DB.Where("tg_id = ?", userID).First(&retryUser).Error
			if retryErr == nil && retryUser.ID > 0 {
				statusText := formatTgUserStatus(retryUser.Status)
				return fmt.Sprintf(`✅ 您已注册！

👤 用户ID：<code>%d</code>
💰 余额：<code>%.3f U</code>
📊 状态：<code>%s</code>
📅 注册时间：<code>%s</code>`,
					retryUser.ID,
					retryUser.Balance,
					statusText,
					retryUser.CreatedAt.Format("2006-01-02 15:04:05")), nil
			}
		}
		return "", fmt.Errorf("注册失败: %v", err)
	}

	return fmt.Sprintf(`🎉 注册成功！

👤 用户ID：<code>%d</code>
💰 余额：<code>%.3f U</code>
📊 状态：<code>正常</code>
🔑 邀请码：<code>%s</code>

现在您可以开始使用红包功能了！`, newUser.ID, newUser.Balance, inviteCode), nil
}

// HandleHelpCommand 处理帮助命令
func (s *TelegramBotService) HandleHelpCommand() string {
	return `注册：/register
发红包：发10-1 或 10-1
发红包（指定数量）：发10-3-1 或 10-3-1
查余额：查 或 1 或 余额
获取ID：获取群信息`
}

// HandleGroupInfoCommand 处理群信息命令
func (s *TelegramBotService) HandleGroupInfoCommand(chatID int64, userID int64) string {
	return fmt.Sprintf("群ID：<code>%d</code>\n用户ID：<code>%d</code>", chatID, userID)
}

// CheckGroupAuth 检查群组是否授权
func (s *TelegramBotService) CheckGroupAuth(chatID int64) bool {
	var authGroup pojo.AuthGroup
	err := s.DB.Where("group_id = ? AND status = ?", chatID, 1).First(&authGroup).Error
	return err == nil && authGroup.ID > 0
}

// GetGroupInfo 获取群组信息
func (s *TelegramBotService) GetGroupInfo(chatID int64) (*pojo.AuthGroup, error) {
	var authGroup pojo.AuthGroup
	err := s.DB.Where("group_id = ?", chatID).First(&authGroup).Error
	if err != nil {
		return nil, err
	}
	return &authGroup, nil
}

// FormatRedPacketMessage 格式化红包消息
func FormatRedPacketMessage(senderName string, amount float64) string {
	formattedName := utils.FormatName(senderName, 8)
	return fmt.Sprintf("[ <code>%s</code> ]发了个 %.0f U红包，快来抢！", formattedName, amount)
}

// FormatRedPacketCompleteMessage 格式化红包完成消息
func FormatRedPacketCompleteMessage(luckyMoney *pojo.LuckyMoney, historyList []pojo.LuckyHistory) string {
	var details strings.Builder
	loseMoneyTotal := 0.0

	for i, history := range historyList {
		if history.IsThunder == 1 {
			details.WriteString(fmt.Sprintf("%d.[💣] <code>%.2f</code> U <code>%s</code>\n",
				i+1, history.Amount, utils.FormatName(history.FirstName, 8)))
			loseMoneyTotal += history.LoseMoney
		} else {
			details.WriteString(fmt.Sprintf("%d.[💵] <code>%.2f</code> U <code>%s</code>\n",
				i+1, history.Amount, utils.FormatName(history.FirstName, 8)))
		}
	}

	profit := loseMoneyTotal - luckyMoney.Amount
	profitTxt := fmt.Sprintf("%+.2f", profit)
	if profit >= 0 {
		profitTxt = "+" + profitTxt
	}

	return fmt.Sprintf(`[ <code>%s</code> ]的红包已被领完！
🧧红包金额：%.0f U
🛎红包倍数：%.2f
💥中雷数字：%d
--------领取详情--------
%s
<pre>💹 中雷盈利： %.2f</pre>
<pre>💹 发包成本：-%.2f</pre>
<pre>💹 包主实收：%s</pre>`,
		utils.FormatName(luckyMoney.SenderName, 8),
		luckyMoney.Amount,
		luckyMoney.LoseRate,
		luckyMoney.Thunder,
		details.String(),
		loseMoneyTotal,
		luckyMoney.Amount,
		profitTxt)
}

// GetTelegramUserIDFromUsername 从用户名获取 Telegram User ID
// 假设用户名格式为 tg_123456789
func GetTelegramUserIDFromUsername(username string) (int64, error) {
	if !strings.HasPrefix(username, "tg_") {
		return 0, fmt.Errorf("无效的用户名格式")
	}
	userIDStr := strings.TrimPrefix(username, "tg_")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		return 0, err
	}
	return userID, nil
}

// GetOrCreateTgUserByTelegramID 根据 Telegram ID 获取或创建Telegram用户
func (s *TelegramBotService) GetOrCreateTgUserByTelegramID(telegramUserID int64, userName string, chatID int64) (*pojo.TgUser, error) {
	_ = chatID
	// 使用 Redis 分布式锁，锁键格式：bgu_tg_reg_{telegramUserID}
	lockKey := fmt.Sprintf("bgu_tg_reg_%d", telegramUserID)
	lockTimeout := 10 * time.Second // 锁超时时间10秒

	// 尝试获取锁
	acquired, err := utils.AcquireLock(lockKey, lockTimeout)
	if err != nil {
		log.Printf("获取Redis锁失败: %v", err)
		// Redis 锁获取失败，继续执行（降级处理）
	} else if !acquired {
		// 锁已被其他进程持有，等待一小段时间后重试
		time.Sleep(100 * time.Millisecond)
		// 再次尝试获取锁
		acquired, err = utils.AcquireLock(lockKey, lockTimeout)
		if err != nil || !acquired {
			return nil, fmt.Errorf("获取用户注册锁失败，请稍后重试")
		}
	}

	// 确保在函数返回时释放锁
	defer func() {
		if acquired {
			_ = utils.ReleaseLock(lockKey)
		}
	}()

	var existingUser pojo.TgUser
	err = s.DB.Where("tg_id = ?", telegramUserID).First(&existingUser).Error
	if err == nil && existingUser.ID > 0 {
		return &existingUser, nil
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	inviteCode, err := s.generateInviteCode()
	if err != nil {
		return nil, err
	}

	displayName := userName
	if displayName == "" {
		displayName = fmt.Sprintf("User_%d", telegramUserID)
	}
	registerGiftAmount := s.getRegisterGiftAmount()
	newUser := pojo.TgUser{
		FirstName:  strPtr(displayName),
		TgID:       telegramUserID,
		Balance:    registerGiftAmount,
		GiftAmount: registerGiftAmount,
		GiftTotal:  registerGiftAmount,
		Status:     1,
		InviteCode: &inviteCode,
	}

	if err := s.DB.Create(&newUser).Error; err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") || strings.Contains(err.Error(), "1062") {
			var retryUser pojo.TgUser
			retryErr := s.DB.Where("tg_id = ?", telegramUserID).First(&retryUser).Error
			if retryErr == nil && retryUser.ID > 0 {
				return &retryUser, nil
			}
		}
		return nil, fmt.Errorf("创建用户失败: %v", err)
	}

	return &newUser, nil
}

// GetTgUserByTelegramID 根据 Telegram ID 获取用户
func (s *TelegramBotService) GetTgUserByTelegramID(telegramUserID int64) (*pojo.TgUser, error) {
	var user pojo.TgUser
	err := s.DB.Where("tg_id = ?", telegramUserID).First(&user).Error
	if err != nil {
		return nil, err
	}
	if user.ID == 0 {
		return nil, fmt.Errorf("用户不存在")
	}
	return &user, nil
}

func formatTgUserDisplayName(user *pojo.TgUser) string {
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

func formatTgUserStatus(status int8) string {
	switch status {
	case 1:
		return "正常"
	case 0:
		return "已禁用"
	case -1:
		return "已删除"
	default:
		return "未知"
	}

}

func parseStartInviteCode(text string) string {
	parts := strings.Fields(text)
	if len(parts) < 2 {
		return ""
	}
	return parts[1]
}

func startOfToday(now time.Time) time.Time {
	year, month, day := now.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, now.Location())
}

func (s *TelegramBotService) getParentIDByInviteCode(userID int64) *int64 {
	ctx := context.Background()
	redisKey := fmt.Sprintf("tg_start_invite:%d", userID)
	inviteCode, err := utils.RD.Get(ctx, redisKey).Result()
	if err != nil || inviteCode == "" {
		return nil
	}
	_ = utils.RD.Del(ctx, redisKey).Err()

	var parent pojo.TgUser
	if err := s.DB.Where("invite_code = ?", inviteCode).First(&parent).Error; err != nil {
		return nil
	}
	if parent.ID == 0 {
		return nil
	}
	return &parent.ID
}

func (s *TelegramBotService) generateInviteCode() (string, error) {
	for i := 0; i < 10; i++ {
		code := fmt.Sprintf("%06d", rand.IntN(1000000))
		var count int64
		if err := s.DB.Model(&pojo.TgUser{}).Where("invite_code = ?", code).Count(&count).Error; err != nil {
			return "", err
		}
		if count == 0 {
			return code, nil
		}
	}
	return fmt.Sprintf("%06d", rand.IntN(1000000)), nil
}

func (s *TelegramBotService) applyRechargeRebate(tx *gorm.DB, order pojo.RechargeOrder, user pojo.TgUser) error {
	firstRate := s.getRebateRate("first_level_rebate")
	secondRate := s.getRebateRate("two_level_rebate")
	if (firstRate <= 0 && secondRate <= 0) || user.ParentID == nil || *user.ParentID == 0 {
		return nil
	}

	parent1ID := *user.ParentID
	var parent1 pojo.TgUser
	if err := tx.Where("id = ?", parent1ID).First(&parent1).Error; err != nil || parent1.ID == 0 {
		return nil
	}

	now := time.Now()
	if err := s.createRebateRecord(tx, 1, order, user.ID, parent1.ID, firstRate, now); err != nil {
		return err
	}

	if secondRate <= 0 || parent1.ParentID == nil || *parent1.ParentID == 0 {
		return nil
	}

	parent2ID := *parent1.ParentID
	var parent2 pojo.TgUser
	if err := tx.Where("id = ?", parent2ID).First(&parent2).Error; err != nil || parent2.ID == 0 {
		return nil
	}

	return s.createRebateRecord(tx, 2, order, user.ID, parent2.ID, secondRate, now)
}

func (s *TelegramBotService) getRebateRate(key string) float64 {
	defaultValue := "0"
	val := utils.GetStringCache(s.TablePrefix, key, &defaultValue)
	if val == nil || *val == "" {
		return 0
	}
	rate, err := strconv.ParseFloat(*val, 64)
	if err != nil {
		return 0
	}
	return rate
}

func (s *TelegramBotService) getRegisterGiftAmount() float64 {
	defaultValue := "0"
	val := utils.GetStringCache(s.TablePrefix, "register_gift_amount", &defaultValue)
	if val == nil || *val == "" {
		return 0
	}

	amount, err := strconv.ParseFloat(strings.TrimSpace(*val), 64)
	if err != nil || amount < 0 {
		return 0
	}

	return amount
}

func (s *TelegramBotService) createRebateRecord(tx *gorm.DB, level int, order pojo.RechargeOrder, subUserID int64, parentUserID int64, rate float64, now time.Time) error {
	if rate <= 0 || parentUserID == 0 {
		return nil
	}

	sourceType := 2
	idempotencyKey := fmt.Sprintf("%d:%s:%d", sourceType, order.OrderNo, parentUserID)
	var existing pojo.TgUserRebateRecord
	if err := tx.Where("idempotency_key = ?", idempotencyKey).First(&existing).Error; err == nil && existing.ID > 0 {
		return nil
	}

	rebateAmount := utils.ToMoney(order.Amount).Multiply(rate / 100).ToDollars()
	if rebateAmount <= 0 {
		return nil
	}

	record := pojo.TgUserRebateRecord{
		TenantId:       &order.TenantId,
		SubUserId:      subUserID,
		ParentUserId:   parentUserID,
		SourceType:     sourceType,
		SourceOrderId:  order.OrderNo,
		SourceAmount:   order.Amount,
		RebateRate:     rate,
		RebateAmount:   rebateAmount,
		Currency:       order.Currency,
		Status:         1,
		SettledAt:      &now,
		IdempotencyKey: idempotencyKey,
		Remark:         strPtr(fmt.Sprintf("level_%d_rebate", level)),
	}

	if err := tx.Create(&record).Error; err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") || strings.Contains(err.Error(), "1062") {
			return nil
		}
		return err
	}

	return tx.Model(&pojo.TgUser{}).
		Where("id = ?", parentUserID).
		Updates(map[string]any{
			"rebate_amount":       gorm.Expr("rebate_amount + ?", rebateAmount),
			"rebate_total_amount": gorm.Expr("rebate_total_amount + ?", rebateAmount),
		}).Error
}
