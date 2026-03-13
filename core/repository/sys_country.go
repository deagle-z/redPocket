package repository

import (
	"BaseGoUni/core/pojo"
	"errors"
	"github.com/jinzhu/copier"
	"gorm.io/gorm"
)

// GetSysCountries 国家列表（分页）
func GetSysCountries(db *gorm.DB, search pojo.SysCountrySearch) (result pojo.SysCountryResp) {
	var list []pojo.SysCountry
	query := db.Model(&pojo.SysCountry{})

	if search.CountryCode != "" {
		query = query.Where("country_code = ?", search.CountryCode)
	}
	if search.CountryName != "" {
		query = query.Where("country_name_cn LIKE ? OR country_name_en LIKE ?", "%"+search.CountryName+"%", "%"+search.CountryName+"%")
	}
	if search.CurrencyCode != "" {
		query = query.Where("currency_code = ?", search.CurrencyCode)
	}
	if search.Status != nil {
		query = query.Where("status = ?", *search.Status)
	}

	query.Count(&result.Total)
	query = query.Order("sort asc, id desc").Limit(search.PageSize).Offset(search.PageSize * search.CurrentPage)
	query.Find(&list)

	for _, item := range list {
		var temp pojo.SysCountryBack
		_ = copier.Copy(&temp, &item)
		result.List = append(result.List, temp)
	}

	result.PageSize = search.PageSize
	result.CurrentPage = search.CurrentPage
	return result
}

// SetSysCountry 创建或更新国家
func SetSysCountry(db *gorm.DB, req pojo.SysCountrySet) (result pojo.SysCountryBack, err error) {
	var entity pojo.SysCountry
	if req.ID > 0 {
		db.Where("id = ?", req.ID).First(&entity)
		if entity.ID == 0 {
			return result, errors.New("更新的数据不存在")
		}
		_ = copier.Copy(&entity, &req)
		err = db.Save(&entity).Error
	} else {
		_ = copier.Copy(&entity, &req)
		err = db.Create(&entity).Error
	}
	if err != nil {
		return result, err
	}
	_ = copier.Copy(&result, &entity)
	return result, nil
}

// DelSysCountry 删除国家
func DelSysCountry(db *gorm.DB, id int64) (result string, err error) {
	var entity pojo.SysCountry
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

// GetSysCountryById 根据ID获取国家
func GetSysCountryById(db *gorm.DB, id int64) (result pojo.SysCountryBack, err error) {
	var entity pojo.SysCountry
	db.Where("id = ?", id).First(&entity)
	if entity.ID == 0 {
		return result, errors.New("数据不存在")
	}
	_ = copier.Copy(&result, &entity)
	return result, nil
}
