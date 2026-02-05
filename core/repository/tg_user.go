package repository

import (
	"BaseGoUni/core/pojo"
	"errors"
	"github.com/jinzhu/copier"
	"gorm.io/gorm"
)

// GetTgUsers Telegram用户列表（分页）
func GetTgUsers(db *gorm.DB, search pojo.TgUserSearch) (result pojo.TgUserResp) {
	var users []pojo.TgUser
	query := db.Model(&pojo.TgUser{})

	if search.TgID > 0 {
		query = query.Where("tg_id = ?", search.TgID)
	}
	if search.Username != "" {
		query = query.Where("username like ?", "%"+search.Username+"%")
	}
	if search.FirstName != "" {
		query = query.Where("first_name like ?", "%"+search.FirstName+"%")
	}
	if search.Status != nil {
		query = query.Where("status = ?", *search.Status)
	}
	if search.ParentID != nil {
		query = query.Where("parent_id = ?", *search.ParentID)
	}
	if search.InviteCode != "" {
		query = query.Where("invite_code = ?", search.InviteCode)
	}

	query.Count(&result.Total)
	query = query.Order("id desc").Limit(search.PageSize).Offset(search.PageSize * search.CurrentPage)
	query.Find(&users)

	for _, user := range users {
		var temp pojo.TgUserBack
		_ = copier.Copy(&temp, &user)
		result.List = append(result.List, temp)
	}

	result.PageSize = search.PageSize
	result.CurrentPage = search.CurrentPage
	return result
}

// SetTgUser 创建或更新Telegram用户
func SetTgUser(db *gorm.DB, req pojo.TgUserSet) (result pojo.TgUserBack, err error) {
	var dbUser pojo.TgUser
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

// DelTgUser 删除Telegram用户
func DelTgUser(db *gorm.DB, id int64) (result string, err error) {
	var dbUser pojo.TgUser
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

// GetTgUserById 根据ID获取Telegram用户
func GetTgUserById(db *gorm.DB, id int64) (result pojo.TgUserBack, err error) {
	var dbUser pojo.TgUser
	db.Where("id = ?", id).First(&dbUser)
	if dbUser.ID == 0 {
		return result, errors.New("数据不存在")
	}
	_ = copier.Copy(&result, &dbUser)
	return result, nil
}

// SetTgUserStatus 更新Telegram用户状态
func SetTgUserStatus(db *gorm.DB, id int64, status int8) (result pojo.TgUserBack, err error) {
	var dbUser pojo.TgUser
	db.Where("id = ?", id).First(&dbUser)
	if dbUser.ID == 0 {
		return result, errors.New("数据不存在")
	}
	err = db.Model(&dbUser).Update("status", status).Error
	if err != nil {
		return result, err
	}
	_ = copier.Copy(&result, &dbUser)
	result.Status = status
	return result, nil
}
