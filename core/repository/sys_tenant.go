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
	var dbTenant pojo.SysTenant
	if req.ID > 0 {
		db.Where("id = ?", req.ID).First(&dbTenant)
		if dbTenant.ID == 0 {
			return result, errors.New("更新的数据不存在")
		}
		_ = copier.Copy(&dbTenant, &req)
		err = db.Save(&dbTenant).Error
	} else {
		err = db.Transaction(func(tx *gorm.DB) error {
			_ = copier.Copy(&dbTenant, &req)
			if err := tx.Create(&dbTenant).Error; err != nil {
				return err
			}
			passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.LoginPassword), bcrypt.DefaultCost)
			if err != nil {
				return err
			}
			dbUser := pojo.SysTenantUser{
				TenantId:     dbTenant.ID,
				Username:     req.LoginAccount,
				PasswordHash: string(passwordHash),
				PasswordAlgo: "bcrypt",
				RoleCode:     "owner",
				IsOwner:      true,
				Status:       req.Status,
			}
			return tx.Create(&dbUser).Error
		})
	}
	if err != nil {
		return result, err
	}
	_ = copier.Copy(&result, &dbTenant)
	return result, nil
}

// DelSysTenant 删除租户
func DelSysTenant(db *gorm.DB, id int64) (result string, err error) {
	var dbTenant pojo.SysTenant
	db.Where("id = ?", id).First(&dbTenant)
	if dbTenant.ID == 0 {
		return result, errors.New("删除的数据不存在")
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
		return result, errors.New("数据不存在")
	}
	_ = copier.Copy(&result, &dbTenant)
	return result, nil
}
