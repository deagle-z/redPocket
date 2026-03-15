package repository

import (
	"BaseGoUni/core/pojo"
	"errors"
	"github.com/jinzhu/copier"
	"gorm.io/gorm"
)

// GetSysUserWithdrawAccounts 提现账户列表（分页，管理员用）
func GetSysUserWithdrawAccounts(db *gorm.DB, search pojo.SysUserWithdrawAccountSearch) (result pojo.SysUserWithdrawAccountResp) {
	var list []pojo.SysUserWithdrawAccount
	query := db.Model(&pojo.SysUserWithdrawAccount{})

	if search.UserID > 0 {
		query = query.Where("user_id = ?", search.UserID)
	}
	if search.CountryCode != "" {
		query = query.Where("country_code = ?", search.CountryCode)
	}
	if search.IsDefault != nil {
		query = query.Where("is_default = ?", *search.IsDefault)
	}
	if search.Status != nil {
		query = query.Where("status = ?", *search.Status)
	}

	query.Count(&result.Total)
	query = query.Order("id desc").Limit(search.PageSize).Offset(search.PageSize * search.CurrentPage)
	query.Find(&list)

	for _, item := range list {
		var temp pojo.SysUserWithdrawAccountBack
		_ = copier.Copy(&temp, &item)
		result.List = append(result.List, temp)
	}

	result.PageSize = search.PageSize
	result.CurrentPage = search.CurrentPage
	return result
}

// GetSysUserWithdrawAccountById 根据ID获取提现账户（管理员用）
func GetSysUserWithdrawAccountById(db *gorm.DB, id int64) (result pojo.SysUserWithdrawAccountBack, err error) {
	var entity pojo.SysUserWithdrawAccount
	db.Where("id = ?", id).First(&entity)
	if entity.ID == 0 {
		return result, errors.New("数据不存在")
	}
	_ = copier.Copy(&result, &entity)
	return result, nil
}

// AdminSetSysUserWithdrawAccount 创建或更新提现账户（管理员用）
func AdminSetSysUserWithdrawAccount(db *gorm.DB, req pojo.SysUserWithdrawAccountSet, tenantID, userID int64) (result pojo.SysUserWithdrawAccountBack, err error) {
	var entity pojo.SysUserWithdrawAccount
	if req.ID > 0 {
		db.Where("id = ?", req.ID).First(&entity)
		if entity.ID == 0 {
			return result, errors.New("更新的数据不存在")
		}
		_ = copier.Copy(&entity, &req)
		err = db.Save(&entity).Error
	} else {
		_ = copier.Copy(&entity, &req)
		entity.TenantID = tenantID
		entity.UserID = userID
		err = db.Create(&entity).Error
	}
	if err != nil {
		return result, err
	}
	// 若设为默认，清除同用户其他默认
	if entity.IsDefault == 1 {
		db.Model(&pojo.SysUserWithdrawAccount{}).
			Where("user_id = ? AND id != ?", entity.UserID, entity.ID).
			Update("is_default", 0)
	}
	_ = copier.Copy(&result, &entity)
	return result, nil
}

// AdminDelSysUserWithdrawAccount 删除提现账户（管理员用）
func AdminDelSysUserWithdrawAccount(db *gorm.DB, id int64) (result string, err error) {
	var entity pojo.SysUserWithdrawAccount
	db.Where("id = ?", id).First(&entity)
	if entity.ID == 0 {
		return result, errors.New("删除的数据不存在")
	}
	err = db.Delete(&entity).Error
	if err != nil {
		return result, err
	}
	return "success", nil
}

// GetAppWithdrawAccounts App端获取当前用户的提现账户列表
func GetAppWithdrawAccounts(db *gorm.DB, userID int64) []pojo.SysUserWithdrawAccountBack {
	var list []pojo.SysUserWithdrawAccount
	db.Where("user_id = ? AND status = 1", userID).Order("is_default desc, id desc").Find(&list)

	result := make([]pojo.SysUserWithdrawAccountBack, 0, len(list))
	for _, item := range list {
		var temp pojo.SysUserWithdrawAccountBack
		_ = copier.Copy(&temp, &item)
		result = append(result, temp)
	}
	return result
}

// AppAddWithdrawAccount App端新增提现账户
func AppAddWithdrawAccount(db *gorm.DB, req pojo.SysUserWithdrawAccountSet, tenantID, userID int64) (result pojo.SysUserWithdrawAccountBack, err error) {
	var entity pojo.SysUserWithdrawAccount
	_ = copier.Copy(&entity, &req)
	entity.TenantID = tenantID
	entity.UserID = userID
	// 若是第一个账户，自动设为默认
	var count int64
	db.Model(&pojo.SysUserWithdrawAccount{}).Where("user_id = ? AND status = 1", userID).Count(&count)
	if count == 0 {
		entity.IsDefault = 1
	}
	err = db.Create(&entity).Error
	if err != nil {
		return result, err
	}
	if entity.IsDefault == 1 {
		db.Model(&pojo.SysUserWithdrawAccount{}).
			Where("user_id = ? AND id != ?", userID, entity.ID).
			Update("is_default", 0)
	}
	_ = copier.Copy(&result, &entity)
	return result, nil
}

// AppUpdateWithdrawAccount App端修改自己的提现账户信息
func AppUpdateWithdrawAccount(db *gorm.DB, id, userID int64, req pojo.SysUserWithdrawAccountSet) (result pojo.SysUserWithdrawAccountBack, err error) {
	var entity pojo.SysUserWithdrawAccount
	db.Where("id = ? AND user_id = ? AND status = 1", id, userID).First(&entity)
	if entity.ID == 0 {
		return result, errors.New("账户不存在")
	}
	entity.AccountData = req.AccountData
	if req.Remark != nil {
		entity.Remark = req.Remark
	}
	err = db.Save(&entity).Error
	if err != nil {
		return result, err
	}
	_ = copier.Copy(&result, &entity)
	return result, nil
}

// AppDelWithdrawAccount App端删除自己的提现账户
func AppDelWithdrawAccount(db *gorm.DB, id, userID int64) (result string, err error) {
	var entity pojo.SysUserWithdrawAccount
	db.Where("id = ? AND user_id = ?", id, userID).First(&entity)
	if entity.ID == 0 {
		return result, errors.New("账户不存在")
	}
	err = db.Delete(&entity).Error
	if err != nil {
		return result, err
	}
	// 若删除的是默认账户，将最新一条设为默认
	if entity.IsDefault == 1 {
		var next pojo.SysUserWithdrawAccount
		db.Where("user_id = ? AND status = 1", userID).Order("id desc").First(&next)
		if next.ID > 0 {
			db.Model(&next).Update("is_default", 1)
		}
	}
	return "success", nil
}

// AppSetDefaultWithdrawAccount App端设置默认提现账户
func AppSetDefaultWithdrawAccount(db *gorm.DB, id, userID int64) (result string, err error) {
	var entity pojo.SysUserWithdrawAccount
	db.Where("id = ? AND user_id = ? AND status = 1", id, userID).First(&entity)
	if entity.ID == 0 {
		return result, errors.New("账户不存在")
	}
	// 清除旧默认，设置新默认
	err = db.Transaction(func(tx *gorm.DB) error {
		if e := tx.Model(&pojo.SysUserWithdrawAccount{}).
			Where("user_id = ?", userID).
			Update("is_default", 0).Error; e != nil {
			return e
		}
		return tx.Model(&entity).Update("is_default", 1).Error
	})
	if err != nil {
		return result, err
	}
	return "success", nil
}
