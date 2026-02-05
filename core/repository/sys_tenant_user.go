package repository

import (
	"BaseGoUni/core/pojo"
	"errors"
	"github.com/jinzhu/copier"
	"gorm.io/gorm"
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
