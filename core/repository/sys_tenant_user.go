package repository

import (
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/utils"
	"context"
	"encoding/json"
	"errors"
	"github.com/jinzhu/copier"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"time"
)

// GetSysTenantUsers 租户用户列表（分页）
func GetSysTenantUsers(db *gorm.DB, search pojo.SysTenantUserSearch) (result pojo.SysTenantUserResp) {
	var users []pojo.SysTenantUser
	query := db.Model(&pojo.SysTenantUser{})

	if search.TenantId > 0 {
		query = query.Where("tenant_id = ?", search.TenantId)
	}
	if search.Username != "" {
		query = query.Where("username like ?", "%"+search.Username+"%")
	}
	if search.Email != "" {
		query = query.Where("email = ?", search.Email)
	}
	if search.Mobile != "" {
		query = query.Where("mobile = ?", search.Mobile)
	}
	if search.RoleCode != "" {
		query = query.Where("role_code = ?", search.RoleCode)
	}
	if search.IsOwner != nil {
		query = query.Where("is_owner = ?", *search.IsOwner)
	}
	if search.Status != nil {
		query = query.Where("status = ?", *search.Status)
	}
	if search.Require2fa != nil {
		query = query.Where("require_2fa = ?", *search.Require2fa)
	}

	query.Count(&result.Total)
	query = query.Order("id desc").Limit(search.PageSize).Offset(search.PageSize * search.CurrentPage)
	query.Find(&users)

	for _, user := range users {
		var temp pojo.SysTenantUserBack
		_ = copier.Copy(&temp, &user)
		result.List = append(result.List, temp)
	}

	result.PageSize = search.PageSize
	result.CurrentPage = search.CurrentPage
	return result
}

// SetSysTenantUser 创建或更新租户用户
func SetSysTenantUser(db *gorm.DB, req pojo.SysTenantUserSet) (result pojo.SysTenantUserBack, err error) {
	var dbUser pojo.SysTenantUser
	if req.ID > 0 {
		db.Where("id = ?", req.ID).First(&dbUser)
		if dbUser.ID == 0 {
			return result, errors.New("更新的数据不存在")
		}
		_ = copier.Copy(&dbUser, &req)
		err = db.Save(&dbUser).Error
	} else {
		_ = copier.Copy(&dbUser, &req)
		err = db.Create(&dbUser).Error
	}
	if err != nil {
		return result, err
	}
	_ = copier.Copy(&result, &dbUser)
	return result, nil
}

// DelSysTenantUser 删除租户用户
func DelSysTenantUser(db *gorm.DB, id int64) (result string, err error) {
	var dbUser pojo.SysTenantUser
	db.Where("id = ?", id).First(&dbUser)
	if dbUser.ID == 0 {
		return result, errors.New("删除的数据不存在")
	}
	err = db.Delete(&dbUser).Error
	if err != nil {
		return result, err
	}
	return "success", nil
}

// GetSysTenantUserById 根据ID获取租户用户
func GetSysTenantUserById(db *gorm.DB, id int64) (result pojo.SysTenantUserBack, err error) {
	var dbUser pojo.SysTenantUser
	db.Where("id = ?", id).First(&dbUser)
	if dbUser.ID == 0 {
		return result, errors.New("数据不存在")
	}
	_ = copier.Copy(&result, &dbUser)
	return result, nil
}

// SysTenantUserLogin 租户用户登录
func SysTenantUserLogin(db *gorm.DB, hostInfo pojo.HostInfo, req pojo.SysTenantUserLogin, onlineUser pojo.OnlineUser) (result pojo.SysTenantUserLoginBack, err error) {
	var users []pojo.SysTenantUser
	db.Where("username = ?", req.Username).Order("id desc").Limit(2).Find(&users)
	if len(users) == 0 {
		return result, errors.New("user login error")
	}
	if len(users) > 1 {
		return result, errors.New("用户名重复，请联系管理员")
	}
	dbUser := users[0]

	var dbTenant pojo.SysTenant
	db.Where("id = ?", dbUser.TenantId).First(&dbTenant)
	if dbTenant.ID == 0 || dbTenant.Status != 1 {
		return result, errors.New("User account disabled")
	}
	if dbUser.Status != 1 {
		return result, errors.New("User account disabled")
	}
	if dbUser.LockedUntil != nil && dbUser.LockedUntil.After(time.Now()) {
		return result, errors.New("账号已锁定")
	}

	if err = bcrypt.CompareHashAndPassword([]byte(dbUser.PasswordHash), []byte(req.Password)); err != nil {
		_ = db.Model(&pojo.SysTenantUser{}).Where("id = ?", dbUser.ID).Update("login_fail_count", gorm.Expr("login_fail_count + 1")).Error
		return result, errors.New("user login error")
	}

	now := time.Now()
	ip := onlineUser.Ip
	ua := onlineUser.Browser
	_ = db.Model(&pojo.SysTenantUser{}).Where("id = ?", dbUser.ID).Updates(map[string]any{
		"last_login_at":    now,
		"last_login_ip":    ip,
		"last_login_ua":    ua,
		"login_fail_count": 0,
	}).Error

	result = pojo.SysTenantUserLoginBack{
		UserId:       dbUser.ID,
		TenantId:     dbUser.TenantId,
		Username:     dbUser.Username,
		RoleCode:     dbUser.RoleCode,
		IsOwner:      dbUser.IsOwner,
		UserType:     4,
		PasswordAlgo: dbUser.PasswordAlgo,
	}
	token, err := utils.GetJwtToken(hostInfo.AccessSecret, hostInfo.AccessExpire, dbUser.Username, dbUser.ID, result.UserType, hostInfo.HostName)
	if err != nil {
		return result, err
	}
	result.AccessToken = token

	key := utils.KeyRdTenantOnline + utils.MD5(token)
	onlineUser.UserId = dbUser.ID
	onlineUser.Key = key
	userJSON, _ := json.Marshal(onlineUser)
	utils.RD.SetEX(context.Background(), key, string(userJSON), time.Duration(hostInfo.AccessExpire)*time.Second)
	return result, nil
}
