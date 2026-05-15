package repository

import (
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/utils"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"math/rand/v2"
	"strconv"
	"strings"
	"time"

	"github.com/jinzhu/copier"
	"gorm.io/gorm"
)

func TgAuthLogin(db *gorm.DB, hostInfo pojo.HostInfo, req pojo.TgAuthLoginReq, onlineUser pojo.OnlineUser, region string) (result pojo.TgAuthLoginBack, err error) {
	if err = utils.VerifyTelegramLoginWidget(utils.GlobalConfig.Telegram.BotToken, req, time.Now()); err != nil {
		return result, err
	}

	var dbUser pojo.TgUser
	err = db.Transaction(func(tx *gorm.DB) error {
		queryErr := tx.Where("tg_id = ?", req.ID).First(&dbUser).Error
		if queryErr != nil {
			if !errors.Is(queryErr, gorm.ErrRecordNotFound) {
				return queryErr
			}
			newUser, createErr := createTgUserFromAuth(tx, req, onlineUser.Ip, region)
			if createErr != nil {
				return createErr
			}
			dbUser = newUser
		}

		if dbUser.Status != 1 {
			return errors.New("user_disabled_contact_admin")
		}

		updates := map[string]any{}
		username := nullableString(req.Username)
		tgName := formatTelegramAtName(req.Username)
		firstName := nullableString(req.FirstName)
		avatar := nullableString(req.PhotoURL)
		if !stringPtrEquals(dbUser.Username, username) {
			updates["username"] = username
		}
		if !stringPtrEquals(dbUser.TgName, tgName) {
			updates["tg_name"] = tgName
		}
		if !stringPtrEquals(dbUser.FirstName, firstName) {
			updates["first_name"] = firstName
		}
		if !stringPtrEquals(dbUser.Avatar, avatar) {
			updates["avatar"] = avatar
		}
		if len(updates) > 0 {
			if err := tx.Model(&pojo.TgUser{}).Where("id = ?", dbUser.ID).Updates(updates).Error; err != nil {
				return err
			}
			dbUser.Username = username
			dbUser.TgName = tgName
			dbUser.FirstName = firstName
			dbUser.Avatar = avatar
		}

		claimUsername := fmt.Sprintf("tg_%d", dbUser.TgID)
		token, tokenErr := utils.GetAppJwtToken(hostInfo.AccessSecret, hostInfo.AccessExpire, claimUsername, dbUser.ID, hostInfo.HostName, dbUser.TenantId)
		if tokenErr != nil {
			return tokenErr
		}

		key := utils.KeyRdTgOnline + utils.MD5(token)
		onlineUser.UserId = dbUser.ID
		onlineUser.Username = claimUsername
		onlineUser.Key = key
		userJSON, _ := json.Marshal(onlineUser)
		_ = utils.RD.SetEX(context.Background(), key, string(userJSON), time.Duration(hostInfo.AccessExpire)*time.Second).Err()
		utils.TouchTgOnlineUser(dbUser.TenantId, dbUser.ID)

		var tgUserBack pojo.TgUserBack
		_ = copier.Copy(&tgUserBack, &dbUser)
		result = pojo.TgAuthLoginBack{
			AccessToken: token,
			UserType:    5,
			ExpiresIn:   hostInfo.AccessExpire,
			TgUser:      tgUserBack,
		}
		return nil
	})
	return result, err
}

// TgEmailLogin 邮箱密码登录
func TgEmailLogin(db *gorm.DB, hostInfo pojo.HostInfo, req pojo.TgEmailLoginReq, onlineUser pojo.OnlineUser) (result pojo.TgAuthLoginBack, err error) {
	email := strings.TrimSpace(strings.ToLower(req.Email))
	if !utils.IsEmail(email) {
		return result, errors.New("email_format_error")
	}
	if strings.TrimSpace(req.Password) == "" {
		return result, errors.New("password_required")
	}

	var users []pojo.TgUser
	db.Where("email = ?", email).Order("id desc").Limit(2).Find(&users)
	if len(users) == 0 {
		return result, errors.New("account_or_password_incorrect")
	}
	if len(users) > 1 {
		return result, errors.New("email_duplicate_contact_admin")
	}
	dbUser := users[0]
	if dbUser.Status != 1 {
		return result, errors.New("user_disabled_contact_admin")
	}
	if dbUser.Password == "" {
		return result, errors.New("account_or_password_incorrect")
	}
	if err = bcrypt.CompareHashAndPassword([]byte(dbUser.Password), []byte(req.Password)); err != nil {
		return result, errors.New("account_or_password_incorrect")
	}

	claimUsername := email
	if dbUser.Username != nil && strings.TrimSpace(*dbUser.Username) != "" {
		claimUsername = strings.TrimSpace(*dbUser.Username)
	}
	token, err := utils.GetAppJwtToken(hostInfo.AccessSecret, hostInfo.AccessExpire, claimUsername, dbUser.ID, hostInfo.HostName, dbUser.TenantId)
	if err != nil {
		return result, err
	}
	key := utils.KeyRdTgOnline + utils.MD5(token)
	onlineUser.UserId = dbUser.ID
	onlineUser.Username = claimUsername
	onlineUser.Key = key
	userJSON, _ := json.Marshal(onlineUser)
	_ = utils.RD.SetEX(context.Background(), key, string(userJSON), time.Duration(hostInfo.AccessExpire)*time.Second).Err()
	utils.TouchTgOnlineUser(dbUser.TenantId, dbUser.ID)

	var tgUserBack pojo.TgUserBack
	_ = copier.Copy(&tgUserBack, &dbUser)
	result = pojo.TgAuthLoginBack{
		AccessToken: token,
		UserType:    5,
		ExpiresIn:   hostInfo.AccessExpire,
		TgUser:      tgUserBack,
	}
	return result, nil
}

// TgPhoneLogin 手机号密码登录
func TgPhoneLogin(db *gorm.DB, hostInfo pojo.HostInfo, req pojo.TgPhoneLoginReq, onlineUser pojo.OnlineUser) (result pojo.TgAuthLoginBack, err error) {
	phone := strings.TrimSpace(req.Phone)
	if !utils.IsPhone(phone) {
		return result, errors.New("phone_format_error")
	}
	if strings.TrimSpace(req.Password) == "" {
		return result, errors.New("password_required")
	}

	query := db.Where("phone = ?", phone)
	var users []pojo.TgUser
	query.Order("id desc").Limit(2).Find(&users)
	if len(users) == 0 {
		return result, errors.New("account_or_password_incorrect")
	}
	if len(users) > 1 {
		return result, errors.New("phone_duplicate_contact_admin")
	}
	dbUser := users[0]
	if dbUser.Status != 1 {
		return result, errors.New("user_disabled_contact_admin")
	}
	if dbUser.Password == "" {
		return result, errors.New("account_or_password_incorrect")
	}
	if err = bcrypt.CompareHashAndPassword([]byte(dbUser.Password), []byte(req.Password)); err != nil {
		return result, errors.New("account_or_password_incorrect")
	}

	claimUsername := phone
	if dbUser.Username != nil && strings.TrimSpace(*dbUser.Username) != "" {
		claimUsername = strings.TrimSpace(*dbUser.Username)
	}
	token, err := utils.GetAppJwtToken(hostInfo.AccessSecret, hostInfo.AccessExpire, claimUsername, dbUser.ID, hostInfo.HostName, dbUser.TenantId)
	if err != nil {
		return result, err
	}
	key := utils.KeyRdTgOnline + utils.MD5(token)
	onlineUser.UserId = dbUser.ID
	onlineUser.Username = claimUsername
	onlineUser.Key = key
	userJSON, _ := json.Marshal(onlineUser)
	_ = utils.RD.SetEX(context.Background(), key, string(userJSON), time.Duration(hostInfo.AccessExpire)*time.Second).Err()
	utils.TouchTgOnlineUser(dbUser.TenantId, dbUser.ID)

	var tgUserBack pojo.TgUserBack
	_ = copier.Copy(&tgUserBack, &dbUser)
	result = pojo.TgAuthLoginBack{
		AccessToken: token,
		UserType:    5,
		ExpiresIn:   hostInfo.AccessExpire,
		TgUser:      tgUserBack,
	}
	return result, nil
}

func createTgUserFromAuth(tx *gorm.DB, req pojo.TgAuthLoginReq, ip string, region string) (pojo.TgUser, error) {
	displayName := strings.TrimSpace(req.FirstName)
	if displayName == "" {
		displayName = fmt.Sprintf("User_%d", req.ID)
	}
	username := nullableString(req.Username)
	tgName := formatTelegramAtName(req.Username)
	firstName := nullableString(displayName)
	avatar := nullableString(req.PhotoURL)
	sourceChannelCode := FirstSourceChannelCode(req.SourceChannelCode, req.ChannelCode)
	sourceChannel, err := ResolveSourceChannelByCode(tx, 0, sourceChannelCode)
	if err != nil {
		return pojo.TgUser{}, err
	}
	defaultRebateRate := getDefaultInviteLuckyRebateRate(tx)

	for i := 0; i < 5; i++ {
		inviteCode, err := generateInviteCode(tx)
		if err != nil {
			return pojo.TgUser{}, err
		}
		uid, err := generateUniqueUID(tx)
		if err != nil {
			return pojo.TgUser{}, err
		}
		newUser := pojo.TgUser{
			Uid:               uid,
			Username:          username,
			TgName:            tgName,
			FirstName:         firstName,
			Avatar:            avatar,
			Ip:                nullableString(ip),
			Region:            nullableString(strings.ToUpper(strings.TrimSpace(region))),
			TgID:              req.ID,
			TrialBalance:      10000,
			Status:            1,
			InviteCode:        &inviteCode,
			SourceChannelID:   nil,
			SourceChannelCode: nil,
			TenantId:          0,
			RebateRate:        defaultRebateRate,
		}
		if sourceChannel != nil {
			newUser.SourceChannelID = &sourceChannel.ID
			newUser.SourceChannelCode = &sourceChannel.ChannelCode
		}
		if err := tx.Create(&newUser).Error; err != nil {
			if strings.Contains(err.Error(), "Duplicate entry") || strings.Contains(err.Error(), "1062") {
				var exist pojo.TgUser
				if findErr := tx.Where("tg_id = ?", req.ID).First(&exist).Error; findErr == nil && exist.ID > 0 {
					return exist, nil
				}
				continue
			}
			return pojo.TgUser{}, err
		}
		return newUser, nil
	}
	return pojo.TgUser{}, errors.New("create_tg_user_failed")
}

func generateInviteCode(db *gorm.DB) (string, error) {
	for i := 0; i < 10; i++ {
		code := fmt.Sprintf("%06d", rand.IntN(1000000))
		var count int64
		if err := db.Model(&pojo.TgUser{}).Where("invite_code = ?", code).Count(&count).Error; err != nil {
			return "", err
		}
		if count == 0 {
			return code, nil
		}
	}
	return fmt.Sprintf("%06d", rand.IntN(1000000)), nil
}

func getDefaultInviteLuckyRebateRate(db *gorm.DB) float64 {
	const defaultRate = 40.0
	var config pojo.SysConfig
	if err := db.Where("config_key = ?", "invite_lucky_rebate_rate").First(&config).Error; err != nil || strings.TrimSpace(config.ConfigValue) == "" {
		return defaultRate
	}
	rate, err := strconv.ParseFloat(strings.TrimSpace(config.ConfigValue), 64)
	if err != nil || rate < 0 {
		return defaultRate
	}
	return rate
}

func GetRegisterFreeLotteryCount(db *gorm.DB) int {
	const defaultCount int = 1
	var config pojo.SysConfig
	if err := db.Where("config_key = ?", "register_free_lottery_count").First(&config).Error; err != nil || strings.TrimSpace(config.ConfigValue) == "" {
		return defaultCount
	}
	count, err := strconv.Atoi(strings.TrimSpace(config.ConfigValue))
	if err != nil || count < 0 {
		return defaultCount
	}
	return count
}

func nullableString(v string) *string {
	v = strings.TrimSpace(v)
	if v == "" {
		return nil
	}
	return &v
}

func formatTelegramAtName(username string) *string {
	username = strings.TrimSpace(username)
	username = strings.TrimPrefix(username, "@")
	if username == "" {
		return nil
	}
	value := "@" + truncateRunes(username, 63)
	return &value
}

func truncateRunes(v string, max int) string {
	if max <= 0 {
		return ""
	}
	runes := []rune(strings.TrimSpace(v))
	if len(runes) <= max {
		return string(runes)
	}
	return string(runes[:max])
}

func stringPtrEquals(a, b *string) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

// SendTgEmailCode 发送邮箱验证码，非dev按IP限流1分钟1次；dev返回验证码。
func SendTgEmailCode(email string, ip string, isDev bool) (string, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	if !utils.IsEmail(email) {
		return "", errors.New("email_format_error")
	}

	if !isDev {
		limitKey := fmt.Sprintf("bgu_tg_email_code_limit_%s", ip)
		ok, err := utils.RD.SetNX(context.Background(), limitKey, "1", time.Minute).Result()
		if err != nil {
			return "", errors.New("service_busy_retry")
		}
		if !ok {
			return "", errors.New("request_too_frequent_retry")
		}
	}

	code := fmt.Sprintf("%06d", rand.IntN(1000000))
	codeKey := fmt.Sprintf("bgu_tg_email_code_%s", email)
	if err := utils.RD.SetEX(context.Background(), codeKey, code, 10*time.Minute).Err(); err != nil {
		return "", errors.New("service_busy_retry")
	}

	if !isDev {
		if err := sendEmailCodeByThirdAPI(email, code); err != nil {
			_ = utils.RD.Del(context.Background(), codeKey).Err()
			return "", errors.New("verify_code_send_failed")
		}
	}
	return code, nil
}

// SendTgSMSCode 发送短信验证码，非dev按IP限流1分钟1次；dev返回验证码。
func SendTgSMSCode(phone string, country string, ip string, isDev bool) (string, error) {
	phone = strings.TrimSpace(phone)
	country = strings.TrimSpace(strings.ToUpper(country))
	if !utils.IsPhone(phone) {
		return "", errors.New("phone_format_error")
	}
	sendPhone := utils.NormalizeSMSPhone(country, phone)

	if !isDev {
		limitKey := fmt.Sprintf("bgu_tg_sms_code_limit_%s", ip)
		ok, err := utils.RD.SetNX(context.Background(), limitKey, "1", time.Minute).Result()
		if err != nil {
			return "", errors.New("service_busy_retry")
		}
		if !ok {
			return "", errors.New("request_too_frequent_retry")
		}
	}

	code := fmt.Sprintf("%06d", rand.IntN(1000000))
	codeKey := fmt.Sprintf("bgu_tg_sms_code_%s", buildTgPhoneCacheKey(country, phone))
	if err := utils.RD.SetEX(context.Background(), codeKey, code, 10*time.Minute).Err(); err != nil {
		return "", errors.New("service_busy_retry")
	}

	//if !isDev {
	client := utils.NewITNioSMSClient()
	content := fmt.Sprintf("Your verification code is %s. It is valid for 10 minutes.", code)
	if _, err := client.SendSMS(sendPhone, content, "", ""); err != nil {
		_ = utils.RD.Del(context.Background(), codeKey).Err()
		return "", errors.New("verify_code_send_failed")
	}
	//}
	return code, nil
}

// RegisterTgByEmail 邮箱注册。
func RegisterTgByEmail(db *gorm.DB, email string, firstName string, password string, code string, sourceChannelCode string, tenantID int64, inviteCode string, ip string, region string) (pojo.TgUser, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	firstName = strings.TrimSpace(firstName)
	code = strings.TrimSpace(code)
	if !utils.IsEmail(email) {
		return pojo.TgUser{}, errors.New("email_format_error")
	}
	if len([]rune(firstName)) > 128 {
		return pojo.TgUser{}, errors.New("first_name_too_long")
	}
	if len(password) < 6 || len(password) > 64 {
		return pojo.TgUser{}, errors.New("password_length_6_64")
	}
	if len(code) != 6 {
		return pojo.TgUser{}, errors.New("code_format_error")
	}

	codeKey := fmt.Sprintf("bgu_tg_email_code_%s", email)
	cacheCode, err := utils.RD.Get(context.Background(), codeKey).Result()
	if err != nil {
		return pojo.TgUser{}, errors.New("code_expired")
	}
	if cacheCode != code {
		return pojo.TgUser{}, errors.New("code_incorrect")
	}

	var exist pojo.TgUser
	if err = db.Where("email = ?", email).First(&exist).Error; err == nil && exist.ID > 0 {
		return pojo.TgUser{}, errors.New("email_registered")
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return pojo.TgUser{}, errors.New("service_busy_retry")
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return pojo.TgUser{}, errors.New("service_busy_retry")
	}

	displayName := firstName
	if displayName == "" {
		displayName = email
		if idx := strings.Index(displayName, "@"); idx > 0 {
			displayName = displayName[:idx]
		}
		displayName = strings.TrimSpace(displayName)
		if displayName == "" {
			displayName = fmt.Sprintf("User_%06d", rand.IntN(1000000))
		}
	}
	displayName = truncateRunes(displayName, 128)
	username := truncateRunes(displayName, 64)

	var newUser pojo.TgUser
	err = db.Transaction(func(tx *gorm.DB) error {
		parentID, parentTenantID, err := resolveRegisterInviteParent(tx, tenantID, inviteCode)
		if err != nil {
			return err
		}
		if tenantID == 0 && parentTenantID > 0 {
			tenantID = parentTenantID
		}
		sourceChannel, err := ResolveSourceChannelByCode(tx, tenantID, sourceChannelCode)
		if err != nil {
			return err
		}
		defaultRebateRate := getDefaultInviteLuckyRebateRate(tx)
		for i := 0; i < 5; i++ {
			//tgID := time.Now().UnixNano()/1e3 + int64(rand.IntN(1000))
			ownInviteCode := fmt.Sprintf("%06d", rand.IntN(1000000))
			uid, uidErr := generateUniqueUID(tx)
			if uidErr != nil {
				return uidErr
			}
			randomAvatar := fmt.Sprintf("https://pub-bd25d6a357314ec1823d725e93570e3d.r2.dev/game/avatar%d.png", rand.IntN(9)+1)
			user := pojo.TgUser{
				Uid:               uid,
				Username:          &username,
				FirstName:         &displayName,
				Avatar:            &randomAvatar,
				Password:          string(passwordHash),
				PasswordPlain:     nullableString(password),
				Email:             email,
				Ip:                nullableString(ip),
				Region:            nullableString(strings.ToUpper(strings.TrimSpace(region))),
				TrialBalance:      10000,
				Status:            1,
				ParentID:          parentID,
				InviteCode:        &ownInviteCode,
				SourceChannelID:   nil,
				SourceChannelCode: nil,
				TenantId:          tenantID,
				RebateRate:        defaultRebateRate,
			}
			if sourceChannel != nil {
				user.SourceChannelID = &sourceChannel.ID
				user.SourceChannelCode = &sourceChannel.ChannelCode
			}
			if createErr := tx.Create(&user).Error; createErr != nil {
				if strings.Contains(createErr.Error(), "Duplicate entry") || strings.Contains(createErr.Error(), "1062") {
					continue
				}
				return createErr
			}
			newUser = user
			return nil
		}
		return errors.New("register_failed_retry")
	})
	if err != nil {
		return pojo.TgUser{}, err
	}

	_ = utils.RD.Del(context.Background(), codeKey).Err()
	return newUser, nil
}

// RegisterTgByPhone 手机号注册。
func RegisterTgByPhone(db *gorm.DB, phone string, country string, firstName string, password string, sourceChannelCode string, tenantID int64, inviteCode string, ip string, region string) (pojo.TgUser, error) {
	phone = utils.NormalizePhoneDigits(phone)
	country = utils.InferCountryByPhone("+"+phone, country)
	firstName = strings.TrimSpace(firstName)
	if !utils.IsPhone(phone) {
		return pojo.TgUser{}, errors.New("phone_format_error")
	}
	if !utils.HasSupportedRegisterPhoneDialCode(phone) {
		return pojo.TgUser{}, errors.New("phone_country_code_required")
	}
	if len([]rune(firstName)) > 128 {
		return pojo.TgUser{}, errors.New("first_name_too_long")
	}
	if len(password) < 6 || len(password) > 64 {
		return pojo.TgUser{}, errors.New("password_length_6_64")
	}

	var exist pojo.TgUser
	if err := db.Where("phone = ? AND status <> ?", phone, -1).First(&exist).Error; err == nil && exist.ID > 0 {
		return pojo.TgUser{}, errors.New("phone_registered")
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return pojo.TgUser{}, errors.New("service_busy_retry")
	}

	displayName := firstName
	if displayName == "" {
		displayName = phone
		if len(displayName) > 4 {
			displayName = displayName[len(displayName)-4:]
		}
		displayName = fmt.Sprintf("User_%s", displayName)
	}
	displayName = truncateRunes(displayName, 128)
	username := truncateRunes(displayName, 64)

	var newUser pojo.TgUser
	err = db.Transaction(func(tx *gorm.DB) error {
		parentID, parentTenantID, err := resolveRegisterInviteParent(tx, tenantID, inviteCode)
		if err != nil {
			return err
		}
		if tenantID == 0 && parentTenantID > 0 {
			tenantID = parentTenantID
		}
		sourceChannel, err := ResolveSourceChannelByCode(tx, tenantID, sourceChannelCode)
		if err != nil {
			return err
		}
		defaultRebateRate := getDefaultInviteLuckyRebateRate(tx)
		for i := 0; i < 5; i++ {
			ownInviteCode := fmt.Sprintf("%06d", rand.IntN(1000000))
			uid, uidErr := generateUniqueUID(tx)
			if uidErr != nil {
				return uidErr
			}
			newUser = pojo.TgUser{
				Uid:               uid,
				Username:          &username,
				FirstName:         &displayName,
				Password:          string(passwordHash),
				PasswordPlain:     nullableString(password),
				Phone:             &phone,
				Country:           nullableString(country),
				Ip:                nullableString(ip),
				Region:            nullableString(strings.ToUpper(strings.TrimSpace(region))),
				TrialBalance:      10000,
				Status:            1,
				ParentID:          parentID,
				InviteCode:        &ownInviteCode,
				SourceChannelID:   nil,
				SourceChannelCode: nil,
				TenantId:          tenantID,
				RebateRate:        defaultRebateRate,
			}
			if sourceChannel != nil {
				newUser.SourceChannelID = &sourceChannel.ID
				newUser.SourceChannelCode = &sourceChannel.ChannelCode
			}
			if err := tx.Create(&newUser).Error; err != nil {
				if strings.Contains(err.Error(), "Duplicate entry") || strings.Contains(err.Error(), "1062") {
					var exist pojo.TgUser
					if findErr := tx.Where("phone = ? AND status <> ?", phone, -1).First(&exist).Error; findErr == nil && exist.ID > 0 {
						return errors.New("phone_registered")
					}
					continue
				}
				return err
			}
			return nil
		}
		return errors.New("phone_register_failed")
	})
	if err != nil {
		return pojo.TgUser{}, err
	}
	return newUser, nil
}

func resolveRegisterInviteParent(db *gorm.DB, tenantID int64, inviteCode string) (*int64, int64, error) {
	inviteCode = strings.TrimSpace(inviteCode)
	if inviteCode == "" {
		return nil, 0, nil
	}

	var parent pojo.TgUser
	query := db.Where("invite_code = ? AND status = ?", inviteCode, 1)
	if tenantID > 0 {
		query = query.Where("tenant_id = ?", tenantID)
	}
	if err := query.First(&parent).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, 0, nil
		}
		return nil, 0, err
	}
	return &parent.ID, parent.TenantId, nil
}

func buildTgPhoneCacheKey(country string, phone string) string {
	country = strings.TrimSpace(strings.ToUpper(country))
	phone = strings.TrimSpace(phone)
	if country == "" {
		return phone
	}
	return fmt.Sprintf("%s_%s", country, phone)
}

// ResetTgPasswordByEmail 忘记密码（邮箱验证码 + 新密码），按IP限流。
func ResetTgPasswordByEmail(db *gorm.DB, email string, code string, newPassword string, ip string) error {
	email = strings.TrimSpace(strings.ToLower(email))
	code = strings.TrimSpace(code)
	newPassword = strings.TrimSpace(newPassword)
	if !utils.IsEmail(email) {
		return errors.New("email_format_error")
	}
	if len(code) != 6 {
		return errors.New("code_format_error")
	}
	if len(newPassword) < 6 || len(newPassword) > 64 {
		return errors.New("password_length_6_64")
	}

	limitKey := fmt.Sprintf("bgu_tg_forgot_pwd_limit_%s", ip)
	ok, err := utils.RD.SetNX(context.Background(), limitKey, "1", time.Minute).Result()
	if err != nil {
		return errors.New("service_busy_retry")
	}
	if !ok {
		return errors.New("request_too_frequent_retry")
	}

	codeKey := fmt.Sprintf("bgu_tg_email_code_%s", email)
	cacheCode, err := utils.RD.Get(context.Background(), codeKey).Result()
	if err != nil {
		return errors.New("code_expired")
	}
	if cacheCode != code {
		return errors.New("code_incorrect")
	}

	var users []pojo.TgUser
	db.Where("email = ?", email).Order("id desc").Limit(2).Find(&users)
	if len(users) == 0 {
		return errors.New("account_not_exist")
	}
	if len(users) > 1 {
		return errors.New("email_duplicate_contact_admin")
	}
	dbUser := users[0]
	if dbUser.Status != 1 {
		return errors.New("user_disabled_contact_admin")
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("service_busy_retry")
	}
	if err = db.Model(&pojo.TgUser{}).Where("id = ?", dbUser.ID).Update("password", string(passwordHash)).Error; err != nil {
		return errors.New("service_busy_retry")
	}
	_ = utils.RD.Del(context.Background(), codeKey).Err()
	return nil
}

// ResetTgPasswordByPhone 忘记密码（短信验证码 + 新密码），按IP限流。
func ResetTgPasswordByPhone(db *gorm.DB, phone string, country string, code string, newPassword string, ip string) error {
	phone = strings.TrimSpace(phone)
	country = strings.TrimSpace(strings.ToUpper(country))
	code = strings.TrimSpace(code)
	newPassword = strings.TrimSpace(newPassword)
	if !utils.IsPhone(phone) {
		return errors.New("phone_format_error")
	}
	if len(code) != 6 {
		return errors.New("code_format_error")
	}
	if len(newPassword) < 6 || len(newPassword) > 64 {
		return errors.New("password_length_6_64")
	}

	limitKey := fmt.Sprintf("bgu_tg_forgot_pwd_phone_limit_%s", ip)
	ok, err := utils.RD.SetNX(context.Background(), limitKey, "1", time.Minute).Result()
	if err != nil {
		return errors.New("service_busy_retry")
	}
	if !ok {
		return errors.New("request_too_frequent_retry")
	}

	codeKey := fmt.Sprintf("bgu_tg_sms_code_%s", buildTgPhoneCacheKey(country, phone))
	cacheCode, err := utils.RD.Get(context.Background(), codeKey).Result()
	if err != nil {
		return errors.New("code_expired")
	}
	if cacheCode != code {
		return errors.New("code_incorrect")
	}

	var users []pojo.TgUser
	db.Where("phone = ? AND status <> ?", phone, -1).Order("id desc").Limit(2).Find(&users)
	if len(users) == 0 {
		return errors.New("account_not_exist")
	}
	if len(users) > 1 {
		return errors.New("phone_duplicate_contact_admin")
	}
	dbUser := users[0]
	if dbUser.Status != 1 {
		return errors.New("user_disabled_contact_admin")
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("service_busy_retry")
	}
	if err = db.Model(&pojo.TgUser{}).Where("id = ?", dbUser.ID).Update("password", string(passwordHash)).Error; err != nil {
		return errors.New("service_busy_retry")
	}
	_ = utils.RD.Del(context.Background(), codeKey).Err()
	return nil
}

// BindCurrentTgPhone 绑定或换绑当前用户手机号。
func BindCurrentTgPhone(db *gorm.DB, userID int64, phone string, country string, code string) error {
	phone = strings.TrimSpace(phone)
	country = strings.TrimSpace(strings.ToUpper(country))
	country = utils.InferCountryByPhone(phone, country)
	code = strings.TrimSpace(code)
	if userID <= 0 {
		return errors.New("token_invalid")
	}
	if !utils.IsPhone(phone) {
		return errors.New("phone_format_error")
	}
	if len(code) != 6 {
		return errors.New("code_format_error")
	}

	codeKey := fmt.Sprintf("bgu_tg_sms_code_%s", buildTgPhoneCacheKey(country, phone))
	cacheCode, err := utils.RD.Get(context.Background(), codeKey).Result()
	if err != nil {
		return errors.New("code_expired")
	}
	if cacheCode != code {
		return errors.New("code_incorrect")
	}

	var exists pojo.TgUser
	if err = db.Where("phone = ? AND id <> ? AND status <> ?", phone, userID, -1).First(&exists).Error; err == nil && exists.ID > 0 {
		return errors.New("phone_registered")
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.New("service_busy_retry")
	}

	result := db.Model(&pojo.TgUser{}).
		Where("id = ? AND status = ?", userID, 1).
		Updates(map[string]any{
			"phone":   phone,
			"country": nullableString(country),
		})
	if result.Error != nil {
		return errors.New("service_busy_retry")
	}
	if result.RowsAffected == 0 {
		return errors.New("user_not_found")
	}
	_ = utils.RD.Del(context.Background(), codeKey).Err()
	return nil
}

// GetCurrentTgUserInfo 获取当前TG用户信息。
func GetCurrentTgUserInfo(db *gorm.DB, accessSecret string, token string) (pojo.TgCurrentUserInfo, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return pojo.TgCurrentUserInfo{}, errors.New("auth_header_missing")
	}
	userID, userType, _, _, err := utils.ParseToken(accessSecret, token)
	if err != nil || userID <= 0 {
		return pojo.TgCurrentUserInfo{}, errors.New("token_invalid")
	}
	if userType != 5 {
		return pojo.TgCurrentUserInfo{}, errors.New("api_not_supported")
	}

	key := utils.KeyRdTgOnline + utils.MD5(token)
	data := utils.RD.Get(context.Background(), key)
	if data == nil || data.Err() != nil || data.Val() == "" {
		return pojo.TgCurrentUserInfo{}, errors.New("token_pass")
	}

	var user pojo.TgUser
	if err = db.Where("id = ?", userID).First(&user).Error; err != nil {
		return pojo.TgCurrentUserInfo{}, errors.New("user_not_found")
	}
	if user.Status != 1 {
		return pojo.TgCurrentUserInfo{}, errors.New("user_disabled_contact_admin")
	}

	return pojo.TgCurrentUserInfo{
		Avatar:           user.Avatar,
		TenantId:         user.TenantId,
		Balance:          utils.Truncate2(user.Balance),
		TrialBalance:     utils.Truncate2(user.TrialBalance),
		Uid:              user.Uid,
		Username:         user.Username,
		TgName:           user.TgName,
		FirstName:        user.FirstName,
		TgID:             user.TgID,
		GiftAmount:       utils.Truncate2(user.GiftAmount),
		RebateAmount:     utils.Truncate2(user.RebateAmount),
		RebateRate:       utils.Truncate2(user.RebateRate),
		FreeLotteryCount: user.FreeLotteryCount,
		Email:            user.Email,
		Phone:            user.Phone,
		Country:          user.Country,
		VipLevel:         user.VipLevel,
		VipLevelName:     user.VipLevelName,
		AudioOpen:        user.AudioOpen,
	}, nil
}

// sendEmailCodeByThirdAPI 预留三方邮件发送调用。
func sendEmailCodeByThirdAPI(email string, code string) error {
	// TODO: 接入三方邮箱API并发送验证码。
	return nil
}

func generateUniqueUID(db *gorm.DB) (string, error) {
	for i := 0; i < 20; i++ {
		uid := strings.ToUpper(utils.RandomString(8))
		var count int64
		if err := db.Model(&pojo.TgUser{}).Where("uid = ?", uid).Count(&count).Error; err != nil {
			return "", err
		}
		if count == 0 {
			return uid, nil
		}
	}
	return "", errors.New("uid_generate_failed")
}
