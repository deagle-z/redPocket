package repository

import (
	"BaseGoUni/core/pojo"
	"errors"
	"github.com/jinzhu/copier"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
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
	} else {
		err = db.Transaction(func(tx *gorm.DB) error {
			_ = copier.Copy(&dbTenant, &req)
			if createErr := tx.Create(&dbTenant).Error; createErr != nil {
				return createErr
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
