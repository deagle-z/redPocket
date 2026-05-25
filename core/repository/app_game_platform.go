package repository

import (
	"BaseGoUni/core/pojo"
	"errors"
	"strings"
	"time"

	"github.com/jinzhu/copier"
	"gorm.io/gorm"
)

func GetAppGamePlatforms(db *gorm.DB, search pojo.AppGamePlatformSearch) (result pojo.AppGamePlatformResp) {
	var list []pojo.AppGamePlatform
	query := db.Model(&pojo.AppGamePlatform{})

	if search.DeletedFlag == nil {
		query = query.Where("COALESCE(deleted_flag, 0) = 0")
	} else {
		query = query.Where("deleted_flag = ?", *search.DeletedFlag)
	}
	if platformCode := strings.TrimSpace(search.PlatformCode); platformCode != "" {
		query = query.Where("platform_code LIKE ?", "%"+platformCode+"%")
	}
	if appID := strings.TrimSpace(search.AppID); appID != "" {
		query = query.Where("app_id LIKE ?", "%"+appID+"%")
	}
	if currency := strings.TrimSpace(search.Currency); currency != "" {
		query = query.Where("currency = ?", currency)
	}
	if search.DisabledFlag != nil {
		query = query.Where("disabled_flag = ?", *search.DisabledFlag)
	}

	query.Count(&result.Total)
	query.Order("sort asc, platform_code asc").
		Limit(search.PageSize).
		Offset(search.PageSize * search.CurrentPage).
		Find(&list)

	result.List = list
	result.PageSize = search.PageSize
	result.CurrentPage = search.CurrentPage
	return result
}

func GetAppGamePlatformByCode(db *gorm.DB, platformCode string) (pojo.AppGamePlatform, error) {
	var entity pojo.AppGamePlatform
	err := db.Where("platform_code = ? AND COALESCE(deleted_flag, 0) = 0", platformCode).First(&entity).Error
	if err != nil {
		return pojo.AppGamePlatform{}, err
	}
	return entity, nil
}

func SetAppGamePlatform(db *gorm.DB, req pojo.AppGamePlatformSet) (pojo.AppGamePlatform, error) {
	platformCode := strings.TrimSpace(req.PlatformCode)
	if platformCode == "" {
		return pojo.AppGamePlatform{}, errors.New("platform_code_required")
	}

	now := time.Now()
	var entity pojo.AppGamePlatform
	err := db.Where("platform_code = ?", platformCode).First(&entity).Error
	if err == nil {
		_ = copier.Copy(&entity, &req)
		entity.PlatformCode = platformCode
		entity.UpdateTime = &now
		return entity, db.Save(&entity).Error
	}
	if err != gorm.ErrRecordNotFound {
		return pojo.AppGamePlatform{}, err
	}

	_ = copier.Copy(&entity, &req)
	entity.PlatformCode = platformCode
	entity.CreateTime = &now
	entity.UpdateTime = &now
	return entity, db.Create(&entity).Error
}

func DelAppGamePlatform(db *gorm.DB, platformCode string) (string, error) {
	now := time.Now()
	result := db.Model(&pojo.AppGamePlatform{}).
		Where("platform_code = ?", platformCode).
		Updates(map[string]any{
			"deleted_flag": 1,
			"update_time":  now,
		})
	if result.Error != nil {
		return "", result.Error
	}
	if result.RowsAffected == 0 {
		return "", errors.New("record_not_found_delete")
	}
	return "success", nil
}
