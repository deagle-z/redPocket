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
			newUser, createErr := createTgUserFromAuth(tx, hostInfo.TablePrefix, req)
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
		token, tokenErr := utils.GetJwtToken(hostInfo.AccessSecret, hostInfo.AccessExpire, claimUsername, dbUser.ID, 5, hostInfo.HostName)
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
	token, err := utils.GetJwtToken(hostInfo.AccessSecret, hostInfo.AccessExpire, claimUsername, dbUser.ID, 5, hostInfo.HostName)
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

func createTgUserFromAuth(tx *gorm.DB, tablePrefix string, req pojo.TgAuthLoginReq) (pojo.TgUser, error) {
	displayName := strings.TrimSpace(req.FirstName)
	if displayName == "" {
		displayName = fmt.Sprintf("User_%d", req.ID)
	}
	registerGiftAmount := getRegisterGiftAmount(tablePrefix)
	username := nullableString(req.Username)
	firstName := nullableString(displayName)
	avatar := nullableString(req.PhotoURL)

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
			Uid:        uid,
			Username:   username,
			FirstName:  firstName,
			Avatar:     avatar,
			TgID:       req.ID,
			Balance:    registerGiftAmount,
			GiftAmount: registerGiftAmount,
			GiftTotal:  registerGiftAmount,
			Status:     1,
			InviteCode: &inviteCode,
			TenantId:   0,
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

func getRegisterGiftAmount(tablePrefix string) float64 {
	defaultValue := "0"
	val := utils.GetStringCache(tablePrefix, "register_gift_amount", &defaultValue)
	if val == nil || *val == "" {
		return 0
	}
	amount, err := strconv.ParseFloat(strings.TrimSpace(*val), 64)
	if err != nil || amount < 0 {
		return 0
	}
	return amount
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

// RegisterTgByEmail 邮箱注册。
func RegisterTgByEmail(db *gorm.DB, email string, password string, code string) (pojo.TgUser, error) {
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
		for i := 0; i < 5; i++ {
			//tgID := time.Now().UnixNano()/1e3 + int64(rand.IntN(1000))
			inviteCode := fmt.Sprintf("%06d", rand.IntN(1000000))
			uid, uidErr := generateUniqueUID(tx)
			if uidErr != nil {
				return uidErr
			}
			user := pojo.TgUser{
				Uid:        uid,
				Username:   &username,
				FirstName:  &displayName,
				Password:   string(passwordHash),
				Email:      email,
				Status:     1,
				InviteCode: &inviteCode,
				TenantId:   0,
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
