package repository

import (
	"BaseGoUni/core/pojo"
	"errors"
	"github.com/jinzhu/copier"
	"gorm.io/gorm"
)

// GetSysCustomFields 自定义字段列表（分页）
func GetSysCustomFields(db *gorm.DB, search pojo.SysCustomFieldSearch) (result pojo.SysCustomFieldResp) {
	var list []pojo.SysCustomField
	query := db.Model(&pojo.SysCustomField{})

	if search.FieldKey != "" {
		query = query.Where("field_key = ?", search.FieldKey)
	}
	if search.FieldLabel != "" {
		query = query.Where("field_label LIKE ?", "%"+search.FieldLabel+"%")
	}
	if search.FieldType != "" {
		query = query.Where("field_type = ?", search.FieldType)
	}
	if search.DataType != "" {
		query = query.Where("data_type = ?", search.DataType)
	}
	if search.Status != nil {
		query = query.Where("status = ?", *search.Status)
	}

	query.Count(&result.Total)
	query = query.Order("id desc").Limit(search.PageSize).Offset(search.PageSize * search.CurrentPage)
	query.Find(&list)

	for _, item := range list {
		var temp pojo.SysCustomFieldBack
		_ = copier.Copy(&temp, &item)
		result.List = append(result.List, temp)
	}

	result.PageSize = search.PageSize
	result.CurrentPage = search.CurrentPage
	return result
}

// SetSysCustomField 创建或更新自定义字段
func SetSysCustomField(db *gorm.DB, req pojo.SysCustomFieldSet) (result pojo.SysCustomFieldBack, err error) {
	var entity pojo.SysCustomField
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

// DelSysCustomField 删除自定义字段
func DelSysCustomField(db *gorm.DB, id int64) (result string, err error) {
	var entity pojo.SysCustomField
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

// GetSysCustomFieldById 根据ID获取自定义字段
func GetSysCustomFieldById(db *gorm.DB, id int64) (result pojo.SysCustomFieldBack, err error) {
	var entity pojo.SysCustomField
	db.Where("id = ?", id).First(&entity)
	if entity.ID == 0 {
		return result, errors.New("数据不存在")
	}
	_ = copier.Copy(&result, &entity)
	return result, nil
}
