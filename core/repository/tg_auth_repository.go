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
	"strings"
	"time"

	"github.com/jinzhu/copier"
	"gorm.io/gorm"
)

func TgAuthLogin(db *gorm.DB, hostInfo pojo.HostInfo, req pojo.TgAuthLoginReq, onlineUser pojo.OnlineUser) (result pojo.TgAuthLoginBack, err error) {
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
			newUser, createErr := createTgUserFromAuth(tx, req)
			if createErr != nil {
				return createErr
			}
			dbUser = newUser
		}

		if dbUser.Status != 1 {
			return errors.New("User account disabled")
		}

		updates := map[string]any{}
		username := nullableString(req.Username)
		firstName := nullableString(req.FirstName)
		avatar := nullableString(req.PhotoURL)
		if !stringPtrEquals(dbUser.Username, username) {
			updates["username"] = username
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
		return result, errors.New("邮箱格式错误")
	}
	if strings.TrimSpace(req.Password) == "" {
		return result, errors.New("密码不能为空")
	}

	var users []pojo.TgUser
	db.Where("email = ?", email).Order("id desc").Limit(2).Find(&users)
	if len(users) == 0 {
		return result, errors.New("账号或密码错误")
	}
	if len(users) > 1 {
		return result, errors.New("邮箱重复，请联系管理员")
	}
	dbUser := users[0]
	if dbUser.Status != 1 {
		return result, errors.New("User account disabled")
	}
	if dbUser.Password == "" {
		return result, errors.New("账号或密码错误")
	}
	if err = bcrypt.CompareHashAndPassword([]byte(dbUser.Password), []byte(req.Password)); err != nil {
		return result, errors.New("账号或密码错误")
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
	country := strings.TrimSpace(strings.ToUpper(req.Country))
	if !utils.IsPhone(phone) {
		return result, errors.New("手机号格式错误")
	}
	if strings.TrimSpace(req.Password) == "" {
		return result, errors.New("密码不能为空")
	}

	query := db.Where("phone = ?", phone)
	if country != "" {
		query = query.Where("country = ?", country)
	}
	var users []pojo.TgUser
	query.Order("id desc").Limit(2).Find(&users)
	if len(users) == 0 {
		return result, errors.New("账号或密码错误")
	}
	if len(users) > 1 {
		return result, errors.New("手机号重复，请联系管理员")
	}
	dbUser := users[0]
	if dbUser.Status != 1 {
		return result, errors.New("User account disabled")
	}
	if dbUser.Password == "" {
		return result, errors.New("账号或密码错误")
	}
	if err = bcrypt.CompareHashAndPassword([]byte(dbUser.Password), []byte(req.Password)); err != nil {
		return result, errors.New("账号或密码错误")
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

func createTgUserFromAuth(tx *gorm.DB, req pojo.TgAuthLoginReq) (pojo.TgUser, error) {
	displayName := strings.TrimSpace(req.FirstName)
	if displayName == "" {
		displayName = fmt.Sprintf("User_%d", req.ID)
	}
	username := nullableString(req.Username)
	firstName := nullableString(displayName)
	avatar := nullableString(req.PhotoURL)
	sourceChannel, err := ResolveSourceChannelByCode(tx, 0, req.SourceChannelCode)
	if err != nil {
		return pojo.TgUser{}, err
	}

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
			FirstName:         firstName,
			Avatar:            avatar,
			TgID:              req.ID,
			Status:            1,
			InviteCode:        &inviteCode,
			SourceChannelID:   nil,
			SourceChannelCode: nil,
			TenantId:          0,
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
	return pojo.TgUser{}, errors.New("create tg user failed")
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

func nullableString(v string) *string {
	v = strings.TrimSpace(v)
	if v == "" {
		return nil
	}
	return &v
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
		return "", errors.New("邮箱格式错误")
	}

	if !isDev {
		limitKey := fmt.Sprintf("bgu_tg_email_code_limit_%s", ip)
		ok, err := utils.RD.SetNX(context.Background(), limitKey, "1", time.Minute).Result()
		if err != nil {
			return "", errors.New("服务异常，请稍后重试")
		}
		if !ok {
			return "", errors.New("请求过于频繁，请1分钟后重试")
		}
	}

	code := fmt.Sprintf("%06d", rand.IntN(1000000))
	codeKey := fmt.Sprintf("bgu_tg_email_code_%s", email)
	if err := utils.RD.SetEX(context.Background(), codeKey, code, 10*time.Minute).Err(); err != nil {
		return "", errors.New("服务异常，请稍后重试")
	}

	if !isDev {
		if err := sendEmailCodeByThirdAPI(email, code); err != nil {
			_ = utils.RD.Del(context.Background(), codeKey).Err()
			return "", errors.New("验证码发送失败")
		}
	}
	return code, nil
}

// SendTgSMSCode 发送短信验证码，非dev按IP限流1分钟1次；dev返回验证码。
func SendTgSMSCode(phone string, country string, ip string, isDev bool) (string, error) {
	phone = strings.TrimSpace(phone)
	country = strings.TrimSpace(strings.ToUpper(country))
	if !utils.IsPhone(phone) {
		return "", errors.New("手机号格式错误")
	}
	sendPhone := utils.NormalizeSMSPhone(country, phone)

	if !isDev {
		limitKey := fmt.Sprintf("bgu_tg_sms_code_limit_%s", ip)
		ok, err := utils.RD.SetNX(context.Background(), limitKey, "1", time.Minute).Result()
		if err != nil {
			return "", errors.New("服务异常，请稍后重试")
		}
		if !ok {
			return "", errors.New("请求过于频繁，请1分钟后重试")
		}
	}

	code := fmt.Sprintf("%06d", rand.IntN(1000000))
	codeKey := fmt.Sprintf("bgu_tg_sms_code_%s", buildTgPhoneCacheKey(country, phone))
	if err := utils.RD.SetEX(context.Background(), codeKey, code, 10*time.Minute).Err(); err != nil {
		return "", errors.New("服务异常，请稍后重试")
	}

	//if !isDev {
	client := utils.NewITNioSMSClient()
	content := fmt.Sprintf("Your verification code is %s. It is valid for 10 minutes.", code)
	if _, err := client.SendSMS(sendPhone, content, "", ""); err != nil {
		_ = utils.RD.Del(context.Background(), codeKey).Err()
		return "", errors.New("验证码发送失败")
	}
	//}
	return code, nil
}

// RegisterTgByEmail 邮箱注册。
func RegisterTgByEmail(db *gorm.DB, email string, password string, code string, sourceChannelCode string) (pojo.TgUser, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	code = strings.TrimSpace(code)
	if !utils.IsEmail(email) {
		return pojo.TgUser{}, errors.New("邮箱格式错误")
	}
	if len(password) < 6 || len(password) > 64 {
		return pojo.TgUser{}, errors.New("密码长度需在6-64位")
	}
	if len(code) != 6 {
		return pojo.TgUser{}, errors.New("验证码格式错误")
	}

	codeKey := fmt.Sprintf("bgu_tg_email_code_%s", email)
	cacheCode, err := utils.RD.Get(context.Background(), codeKey).Result()
	if err != nil {
		return pojo.TgUser{}, errors.New("验证码已失效")
	}
	if cacheCode != code {
		return pojo.TgUser{}, errors.New("验证码错误")
	}

	var exist pojo.TgUser
	if err = db.Where("email = ?", email).First(&exist).Error; err == nil && exist.ID > 0 {
		return pojo.TgUser{}, errors.New("邮箱已注册")
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return pojo.TgUser{}, errors.New("服务异常，请稍后重试")
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return pojo.TgUser{}, errors.New("服务异常，请稍后重试")
	}

	displayName := email
	if idx := strings.Index(displayName, "@"); idx > 0 {
		displayName = displayName[:idx]
	}
	displayName = strings.TrimSpace(displayName)
	if displayName == "" {
		displayName = fmt.Sprintf("User_%06d", rand.IntN(1000000))
	}
	username := displayName

	var newUser pojo.TgUser
	err = db.Transaction(func(tx *gorm.DB) error {
		sourceChannel, err := ResolveSourceChannelByCode(tx, 0, sourceChannelCode)
		if err != nil {
			return err
		}
		for i := 0; i < 5; i++ {
			//tgID := time.Now().UnixNano()/1e3 + int64(rand.IntN(1000))
			inviteCode := fmt.Sprintf("%06d", rand.IntN(1000000))
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
				Email:             email,
				Status:            1,
				InviteCode:        &inviteCode,
				SourceChannelID:   nil,
				SourceChannelCode: nil,
				TenantId:          0,
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
		return errors.New("注册失败，请重试")
	})
	if err != nil {
		return pojo.TgUser{}, err
	}

	_ = utils.RD.Del(context.Background(), codeKey).Err()
	return newUser, nil
}

// RegisterTgByPhone 手机号注册。
func RegisterTgByPhone(db *gorm.DB, phone string, country string, password string, code string, sourceChannelCode string) (pojo.TgUser, error) {
	phone = strings.TrimSpace(phone)
	country = strings.TrimSpace(strings.ToUpper(country))
	code = strings.TrimSpace(code)
	if !utils.IsPhone(phone) {
		return pojo.TgUser{}, errors.New("手机号格式错误")
	}
	if len(password) < 6 || len(password) > 64 {
		return pojo.TgUser{}, errors.New("密码长度需在6-64位")
	}
	if len(code) != 6 {
		return pojo.TgUser{}, errors.New("验证码格式错误")
	}

	codeKey := fmt.Sprintf("bgu_tg_sms_code_%s", buildTgPhoneCacheKey(country, phone))
	cacheCode, err := utils.RD.Get(context.Background(), codeKey).Result()
	if err != nil {
		return pojo.TgUser{}, errors.New("验证码已失效")
	}
	if cacheCode != code {
		return pojo.TgUser{}, errors.New("验证码错误")
	}

	query := db.Where("phone = ?", phone)
	if country != "" {
		query = query.Where("country = ?", country)
	}
	var exist pojo.TgUser
	if err = query.First(&exist).Error; err == nil && exist.ID > 0 {
		return pojo.TgUser{}, errors.New("手机号已注册")
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return pojo.TgUser{}, errors.New("服务异常，请稍后重试")
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return pojo.TgUser{}, errors.New("服务异常，请稍后重试")
	}

	displayName := phone
	if len(displayName) > 4 {
		displayName = displayName[len(displayName)-4:]
	}
	displayName = fmt.Sprintf("User_%s", displayName)
	username := displayName

	var newUser pojo.TgUser
	err = db.Transaction(func(tx *gorm.DB) error {
		sourceChannel, err := ResolveSourceChannelByCode(tx, 0, sourceChannelCode)
		if err != nil {
			return err
		}
		for i := 0; i < 5; i++ {
			inviteCode := fmt.Sprintf("%06d", rand.IntN(1000000))
			uid, uidErr := generateUniqueUID(tx)
			if uidErr != nil {
				return uidErr
			}
			newUser = pojo.TgUser{
				Uid:               uid,
				Username:          &username,
				FirstName:         &displayName,
				Password:          string(passwordHash),
				Phone:             &phone,
				Country:           nullableString(country),
				Status:            1,
				InviteCode:        &inviteCode,
				SourceChannelID:   nil,
				SourceChannelCode: nil,
				TenantId:          0,
			}
			if sourceChannel != nil {
				newUser.SourceChannelID = &sourceChannel.ID
				newUser.SourceChannelCode = &sourceChannel.ChannelCode
			}
			if err := tx.Create(&newUser).Error; err != nil {
				if strings.Contains(err.Error(), "Duplicate entry") || strings.Contains(err.Error(), "1062") {
					var exist pojo.TgUser
					q := tx.Where("phone = ?", phone)
					if country != "" {
						q = q.Where("country = ?", country)
					}
					if findErr := q.First(&exist).Error; findErr == nil && exist.ID > 0 {
						return errors.New("手机号已注册")
					}
					continue
				}
				return err
			}
			return nil
		}
		return errors.New("手机号注册失败")
	})
	if err != nil {
		return pojo.TgUser{}, err
	}

	_ = utils.RD.Del(context.Background(), codeKey).Err()
	return newUser, nil
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
		return errors.New("邮箱格式错误")
	}
	if len(code) != 6 {
		return errors.New("验证码格式错误")
	}
	if len(newPassword) < 6 || len(newPassword) > 64 {
		return errors.New("密码长度需在6-64位")
	}

	limitKey := fmt.Sprintf("bgu_tg_forgot_pwd_limit_%s", ip)
	ok, err := utils.RD.SetNX(context.Background(), limitKey, "1", time.Minute).Result()
	if err != nil {
		return errors.New("服务异常，请稍后重试")
	}
	if !ok {
		return errors.New("请求过于频繁，请1分钟后重试")
	}

	codeKey := fmt.Sprintf("bgu_tg_email_code_%s", email)
	cacheCode, err := utils.RD.Get(context.Background(), codeKey).Result()
	if err != nil {
		return errors.New("验证码已失效")
	}
	if cacheCode != code {
		return errors.New("验证码错误")
	}

	var users []pojo.TgUser
	db.Where("email = ?", email).Order("id desc").Limit(2).Find(&users)
	if len(users) == 0 {
		return errors.New("账号不存在")
	}
	if len(users) > 1 {
		return errors.New("邮箱重复，请联系管理员")
	}
	dbUser := users[0]
	if dbUser.Status != 1 {
		return errors.New("User account disabled")
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("服务异常，请稍后重试")
	}
	if err = db.Model(&pojo.TgUser{}).Where("id = ?", dbUser.ID).Update("password", string(passwordHash)).Error; err != nil {
		return errors.New("服务异常，请稍后重试")
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
		return errors.New("手机号格式错误")
	}
	if len(code) != 6 {
		return errors.New("验证码格式错误")
	}
	if len(newPassword) < 6 || len(newPassword) > 64 {
		return errors.New("密码长度需在6-64位")
	}

	limitKey := fmt.Sprintf("bgu_tg_forgot_pwd_phone_limit_%s", ip)
	ok, err := utils.RD.SetNX(context.Background(), limitKey, "1", time.Minute).Result()
	if err != nil {
		return errors.New("服务异常，请稍后重试")
	}
	if !ok {
		return errors.New("请求过于频繁，请1分钟后重试")
	}

	codeKey := fmt.Sprintf("bgu_tg_sms_code_%s", buildTgPhoneCacheKey(country, phone))
	cacheCode, err := utils.RD.Get(context.Background(), codeKey).Result()
	if err != nil {
		return errors.New("验证码已失效")
	}
	if cacheCode != code {
		return errors.New("验证码错误")
	}

	query := db.Where("phone = ?", phone)
	if country != "" {
		query = query.Where("country = ?", country)
	}
	var users []pojo.TgUser
	query.Order("id desc").Limit(2).Find(&users)
	if len(users) == 0 {
		return errors.New("账号不存在")
	}
	if len(users) > 1 {
		return errors.New("手机号重复，请联系管理员")
	}
	dbUser := users[0]
	if dbUser.Status != 1 {
		return errors.New("User account disabled")
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("服务异常，请稍后重试")
	}
	if err = db.Model(&pojo.TgUser{}).Where("id = ?", dbUser.ID).Update("password", string(passwordHash)).Error; err != nil {
		return errors.New("服务异常，请稍后重试")
	}
	_ = utils.RD.Del(context.Background(), codeKey).Err()
	return nil
}

// GetCurrentTgUserInfo 获取当前TG用户信息。
func GetCurrentTgUserInfo(db *gorm.DB, accessSecret string, token string) (pojo.TgCurrentUserInfo, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return pojo.TgCurrentUserInfo{}, errors.New("Authorization header is missing")
	}
	userID, userType, _, _, err := utils.ParseToken(accessSecret, token)
	if err != nil || userID <= 0 {
		return pojo.TgCurrentUserInfo{}, errors.New("token is invalid")
	}
	if userType != 5 {
		return pojo.TgCurrentUserInfo{}, errors.New("not support api")
	}

	key := utils.KeyRdTgOnline + utils.MD5(token)
	data := utils.RD.Get(context.Background(), key)
	if data == nil || data.Err() != nil || data.Val() == "" {
		return pojo.TgCurrentUserInfo{}, errors.New("token is passed")
	}

	var user pojo.TgUser
	if err = db.Where("id = ?", userID).First(&user).Error; err != nil {
		return pojo.TgCurrentUserInfo{}, errors.New("用户不存在")
	}
	if user.Status != 1 {
		return pojo.TgCurrentUserInfo{}, errors.New("User account disabled")
	}

	return pojo.TgCurrentUserInfo{
		Avatar:       user.Avatar,
		Balance:      user.Balance,
		Uid:          user.Uid,
		Username:     user.Username,
		TgID:         user.TgID,
		GiftAmount:   user.GiftAmount,
		RebateAmount: user.RebateAmount,
		Email:        user.Email,
		Phone:        user.Phone,
		Country:      user.Country,
		VipLevel:     user.VipLevel,
		VipLevelName: user.VipLevelName,
		AudioOpen:    user.AudioOpen,
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
	return "", errors.New("生成uid失败，请重试")
}
