package repository

import (
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/utils"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/jinzhu/copier"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const (
	tenantServiceLinksCacheByTenant = "bgu_tenant_service_links:tenant:%d"
	tenantServiceLinksCacheByHost   = "bgu_tenant_service_links:host:%s"
)

// GetSysTenants 租户列表（分页）
func GetSysTenants(db *gorm.DB, search pojo.SysTenantSearch) (result pojo.SysTenantResp) {
	var tenants []pojo.SysTenant
	query := db.Model(&pojo.SysTenant{})

	if search.TenantCode != "" {
		query = query.Where("tenant_code = ?", search.TenantCode)
	}
	if search.TenantName != "" {
		query = query.Where("tenant_name like ?", "%"+search.TenantName+"%")
	}
	if search.TenantType != nil {
		query = query.Where("tenant_type = ?", *search.TenantType)
	}
	if search.Status != nil {
		query = query.Where("status = ?", *search.Status)
	}
	if search.OwnerUserId != nil {
		query = query.Where("owner_user_id = ?", *search.OwnerUserId)
	}
	if search.PlanCode != "" {
		query = query.Where("plan_code = ?", search.PlanCode)
	}
	if search.BindDomain != "" {
		query = query.Where("bind_domain = ?", search.BindDomain)
	}

	query.Count(&result.Total)
	query = query.Order("id desc").Limit(search.PageSize).Offset(search.PageSize * search.CurrentPage)
	query.Find(&tenants)

	for _, tenant := range tenants {
		var temp pojo.SysTenantBack
		_ = copier.Copy(&temp, &tenant)
		result.List = append(result.List, temp)
	}

	result.PageSize = search.PageSize
	result.CurrentPage = search.CurrentPage
	return result
}

// SetSysTenant 创建或更新租户
func SetSysTenant(db *gorm.DB, req pojo.SysTenantSet) (result pojo.SysTenantBack, err error) {
	req.BindDomain, err = normalizeBindDomain(req.BindDomain)
	if err != nil {
		return result, err
	}
	if err = ensureTenantBindDomainUnique(db, req.BindDomain, req.ID); err != nil {
		return result, err
	}

	var dbTenant pojo.SysTenant
	if req.ID > 0 {
		db.Where("id = ?", req.ID).First(&dbTenant)
		if dbTenant.ID == 0 {
			return result, errors.New("record_not_found_update")
		}
		oldBindDomain := dbTenant.BindDomain
		_ = copier.Copy(&dbTenant, &req)
		if req.BindDomain == nil {
			dbTenant.BindDomain = oldBindDomain
		}
		err = db.Save(&dbTenant).Error
		if err == nil {
			ClearTenantServiceLinksCache(dbTenant.ID)
		}
	} else {
		loginAccount := strings.TrimSpace(req.LoginAccount)
		if loginAccount == "" {
			loginAccount = strings.TrimSpace(req.TenantCode)
		}
		loginPassword := req.LoginPassword
		if loginPassword == "" {
			loginPassword = req.Password
		}
		if loginAccount == "" || len(loginAccount) > 64 {
			return result, errors.New("tenant_login_account_required")
		}
		if len(loginPassword) < 6 || len(loginPassword) > 64 {
			return result, errors.New("tenant_password_length_6_64")
		}

		err = db.Transaction(func(tx *gorm.DB) error {
			var userCount int64
			if countErr := tx.Model(&pojo.SysTenantUser{}).Where("username = ?", loginAccount).Count(&userCount).Error; countErr != nil {
				return countErr
			}
			if userCount > 0 {
				return errors.New("username_duplicate_contact_admin")
			}

			_ = copier.Copy(&dbTenant, &req)
			if createErr := tx.Create(&dbTenant).Error; createErr != nil {
				return createErr
			}
			passwordHash, hashErr := bcrypt.GenerateFromPassword([]byte(loginPassword), bcrypt.DefaultCost)
			if hashErr != nil {
				return hashErr
			}
			tenantUser := pojo.SysTenantUser{
				TenantId:     dbTenant.ID,
				Username:     loginAccount,
				PasswordHash: string(passwordHash),
				PasswordAlgo: "bcrypt",
				BindDomain:   dbTenant.BindDomain,
				RoleCode:     "owner",
				IsOwner:      true,
				Status:       1,
			}
			if createUserErr := tx.Create(&tenantUser).Error; createUserErr != nil {
				return createUserErr
			}
			dbTenant.OwnerUserId = &tenantUser.ID
			if updateErr := tx.Model(&pojo.SysTenant{}).Where("id = ?", dbTenant.ID).Update("owner_user_id", tenantUser.ID).Error; updateErr != nil {
				return updateErr
			}
			if req.BindDomain != nil {
				if nginxErr := createNginxBindDomainConfig(*req.BindDomain); nginxErr != nil {
					return nginxErr
				}
			}
			return nil
		})
	}
	if err != nil {
		return result, err
	}
	_ = copier.Copy(&result, &dbTenant)
	return result, nil
}

func ensureTenantBindDomainUnique(db *gorm.DB, bindDomain *string, excludeID int64) error {
	if bindDomain == nil {
		return nil
	}

	var count int64
	query := db.Model(&pojo.SysTenant{}).Where("bind_domain = ?", *bindDomain)
	if excludeID > 0 {
		query = query.Where("id <> ?", excludeID)
	}
	if err := query.Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return errors.New("bind_domain_duplicate")
	}
	return nil
}

// DelSysTenant 删除租户
func DelSysTenant(db *gorm.DB, id int64) (result string, err error) {
	var dbTenant pojo.SysTenant
	db.Where("id = ?", id).First(&dbTenant)
	if dbTenant.ID == 0 {
		return result, errors.New("record_not_found_delete")
	}
	err = db.Delete(&dbTenant).Error
	if err != nil {
		return result, err
	}
	ClearTenantServiceLinksCache(dbTenant.ID)
	return "success", nil
}

// GetSysTenantById 根据ID获取租户
func GetSysTenantById(db *gorm.DB, id int64) (result pojo.SysTenantBack, err error) {
	var dbTenant pojo.SysTenant
	db.Where("id = ?", id).First(&dbTenant)
	if dbTenant.ID == 0 {
		return result, errors.New("record_not_found")
	}
	_ = copier.Copy(&result, &dbTenant)
	return result, nil
}

func GetCurrentTenantServiceLinks(db *gorm.DB, tenantID int64, host string) (result pojo.SysTenantServiceLinksBack, err error) {
	host = normalizeTenantHost(host)
	if host != "" {
		hostCacheKey := tenantServiceLinksCacheKey(0, host)
		if cached, ok := getTenantServiceLinksCache(hostCacheKey); ok {
			return cached, nil
		}

		candidates := tenantHostCandidates(host)
		if len(candidates) > 0 {
			var hostTenant pojo.SysTenant
			hostErr := db.Model(&pojo.SysTenant{}).
				Where("status = ? AND bind_domain IN ?", 1, candidates).
				First(&hostTenant).Error
			if hostErr == nil && hostTenant.ID > 0 {
				result = pojo.SysTenantServiceLinksBack{
					TgServiceURL: hostTenant.TgServiceURL,
					WsServiceURL: hostTenant.WsServiceURL,
				}
				setTenantServiceLinksCache(hostCacheKey, result)
				return result, nil
			}
		}
	}

	if tenantID <= 0 {
		return result, nil
	}
	tenantCacheKey := tenantServiceLinksCacheKey(tenantID, "")
	if cached, ok := getTenantServiceLinksCache(tenantCacheKey); ok {
		return cached, nil
	}

	var dbTenant pojo.SysTenant
	if err = db.Model(&pojo.SysTenant{}).
		Where("status = ? AND id = ?", 1, tenantID).
		First(&dbTenant).Error; err != nil {
		return result, errors.New("tenant_not_found")
	}
	result = pojo.SysTenantServiceLinksBack{
		TgServiceURL: dbTenant.TgServiceURL,
		WsServiceURL: dbTenant.WsServiceURL,
	}
	setTenantServiceLinksCache(tenantCacheKey, result)
	return result, nil
}

func ClearTenantServiceLinksCache(tenantID int64) {
	if utils.RD == nil {
		return
	}
	ctx := context.Background()
	keys := make([]string, 0, 8)
	if tenantID > 0 {
		keys = append(keys, fmt.Sprintf(tenantServiceLinksCacheByTenant, tenantID))
	}

	iter := utils.RD.Scan(ctx, 0, fmt.Sprintf(tenantServiceLinksCacheByHost, "*"), 100).Iterator()
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	if len(keys) > 0 {
		_ = utils.RD.Del(ctx, keys...).Err()
	}
}

func tenantServiceLinksCacheKey(tenantID int64, host string) string {
	if tenantID > 0 {
		return fmt.Sprintf(tenantServiceLinksCacheByTenant, tenantID)
	}
	host = normalizeTenantHost(host)
	if host == "" {
		return ""
	}
	return fmt.Sprintf(tenantServiceLinksCacheByHost, host)
}

func getTenantServiceLinksCache(key string) (pojo.SysTenantServiceLinksBack, bool) {
	var result pojo.SysTenantServiceLinksBack
	if utils.RD == nil {
		return result, false
	}
	data, err := utils.RD.Get(context.Background(), key).Result()
	if err != nil || data == "" {
		return result, false
	}
	if err = json.Unmarshal([]byte(data), &result); err != nil {
		return result, false
	}
	return result, true
}

func setTenantServiceLinksCache(key string, value pojo.SysTenantServiceLinksBack) {
	if utils.RD == nil {
		return
	}
	data, err := json.Marshal(value)
	if err != nil {
		return
	}
	_ = utils.RD.SetEX(context.Background(), key, data, utils.GetRandomRangeSecond(20*60, 40*60)).Err()
}

func tenantHostCandidates(host string) []string {
	host = normalizeTenantHost(host)
	if host == "" {
		return nil
	}

	candidates := []string{host}
	if strings.HasPrefix(host, "www.") {
		candidates = append(candidates, strings.TrimPrefix(host, "www."))
	}
	parts := strings.Split(host, ".")
	for i := 1; i < len(parts)-1; i++ {
		candidates = append(candidates, "*."+strings.Join(parts[i:], "."))
	}
	return candidates
}

func normalizeTenantHost(host string) string {
	host = strings.ToLower(strings.TrimSpace(host))
	return strings.TrimSuffix(host, ".")
}

// ResetSysTenantPassword 重置租户密码（优先重置租户owner账号）
func ResetSysTenantPassword(db *gorm.DB, req pojo.SysTenantResetPassword) (result string, err error) {
	if req.TenantId <= 0 {
		return result, errors.New("invalid_params")
	}
	if len(req.Password) < 6 || len(req.Password) > 64 {
		return result, errors.New("tenant_password_length_6_64")
	}

	var dbTenant pojo.SysTenant
	db.Where("id = ?", req.TenantId).First(&dbTenant)
	if dbTenant.ID == 0 {
		return result, errors.New("tenant_not_found")
	}

	var dbUser pojo.SysTenantUser
	db.Where("tenant_id = ? and is_owner = ?", req.TenantId, true).Order("id asc").First(&dbUser)
	if dbUser.ID == 0 {
		db.Where("tenant_id = ?", req.TenantId).Order("id asc").First(&dbUser)
	}
	if dbUser.ID == 0 {
		return result, errors.New("tenant_user_missing")
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return result, err
	}
	err = db.Model(&pojo.SysTenantUser{}).Where("id = ?", dbUser.ID).Updates(map[string]any{
		"password_hash":    string(passwordHash),
		"password_algo":    "bcrypt",
		"login_fail_count": 0,
		"locked_until":     nil,
	}).Error
	if err != nil {
		return result, err
	}
	return "success", nil
}
