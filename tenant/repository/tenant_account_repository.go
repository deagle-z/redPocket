package repository

import (
	"BaseGoUni/core/pojo"
	"errors"
	"strings"

	"github.com/pquerna/otp/totp"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const tenantTwofaIssuer = "RedTenantAdmin"

// ChangeTenantUserPassword 修改当前租户用户密码（校验旧密码）。
func ChangeTenantUserPassword(db *gorm.DB, tenantID int64, userID int64, oldPwd string, newPwd string) error {
	oldPwd = strings.TrimSpace(oldPwd)
	newPwd = strings.TrimSpace(newPwd)
	if len(newPwd) < 6 || len(newPwd) > 64 {
		return errors.New("password_length_6_64")
	}
	var user pojo.SysTenantUser
	db.Where("id = ? AND tenant_id = ?", userID, tenantID).First(&user)
	if user.ID == 0 {
		return errors.New("user_not_found")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPwd)); err != nil {
		return errors.New("old_password_incorrect")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(newPwd), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return db.Model(&pojo.SysTenantUser{}).Where("id = ?", user.ID).Updates(map[string]any{
		"password_hash": string(hash),
		"password_algo": "bcrypt",
	}).Error
}

// GetTenantUserTwofaStatus 返回当前用户是否已绑定 Google 验证码。
func GetTenantUserTwofaStatus(db *gorm.DB, tenantID int64, userID int64) (bool, error) {
	var user pojo.SysTenantUser
	db.Where("id = ? AND tenant_id = ?", userID, tenantID).First(&user)
	if user.ID == 0 {
		return false, errors.New("user_not_found")
	}
	bound := user.Require2fa && user.TwofaSecret != nil && strings.TrimSpace(*user.TwofaSecret) != ""
	return bound, nil
}

// GenerateTenantUserTwofa 生成一个新的 TOTP 密钥与 otpauth 链接（未保存，绑定时再写入）。
func GenerateTenantUserTwofa(db *gorm.DB, tenantID int64, userID int64) (secret string, otpauthURL string, err error) {
	var user pojo.SysTenantUser
	db.Where("id = ? AND tenant_id = ?", userID, tenantID).First(&user)
	if user.ID == 0 {
		return "", "", errors.New("user_not_found")
	}
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      tenantTwofaIssuer,
		AccountName: user.Username,
	})
	if err != nil {
		return "", "", err
	}
	return key.Secret(), key.URL(), nil
}

// BindTenantUserTwofa 校验动态码后绑定 Google 验证码（开启 require_2fa）。
func BindTenantUserTwofa(db *gorm.DB, tenantID int64, userID int64, secret string, code string) error {
	secret = strings.TrimSpace(secret)
	code = strings.TrimSpace(code)
	if secret == "" || code == "" {
		return errors.New("invalid_params")
	}
	var user pojo.SysTenantUser
	db.Where("id = ? AND tenant_id = ?", userID, tenantID).First(&user)
	if user.ID == 0 {
		return errors.New("user_not_found")
	}
	if !totp.Validate(code, secret) {
		return errors.New("2fa_code_incorrect")
	}
	return db.Model(&pojo.SysTenantUser{}).Where("id = ?", user.ID).Updates(map[string]any{
		"twofa_secret": secret,
		"require_2fa":  true,
	}).Error
}

// UnbindTenantUserTwofa 校验当前动态码后解绑 Google 验证码。
func UnbindTenantUserTwofa(db *gorm.DB, tenantID int64, userID int64, code string) error {
	code = strings.TrimSpace(code)
	var user pojo.SysTenantUser
	db.Where("id = ? AND tenant_id = ?", userID, tenantID).First(&user)
	if user.ID == 0 {
		return errors.New("user_not_found")
	}
	if user.TwofaSecret == nil || strings.TrimSpace(*user.TwofaSecret) == "" {
		// 未绑定，直接视为成功
		return db.Model(&pojo.SysTenantUser{}).Where("id = ?", user.ID).Update("require_2fa", false).Error
	}
	if !totp.Validate(code, strings.TrimSpace(*user.TwofaSecret)) {
		return errors.New("2fa_code_incorrect")
	}
	return db.Model(&pojo.SysTenantUser{}).Where("id = ?", user.ID).Updates(map[string]any{
		"twofa_secret": nil,
		"require_2fa":  false,
	}).Error
}
