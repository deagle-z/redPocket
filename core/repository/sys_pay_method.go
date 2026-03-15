package repository

import (
	"BaseGoUni/core/pojo"
	"errors"
	"github.com/jinzhu/copier"
	"gorm.io/gorm"
	"time"
)

// GetSysPayMethods 支付方式列表（分页）
func GetSysPayMethods(db *gorm.DB, search pojo.SysPayMethodSearch) (result pojo.SysPayMethodResp) {
	var list []pojo.SysPayMethod
	query := db.Model(&pojo.SysPayMethod{}).Where("deleted_at = 0")

	if search.MethodCode != "" {
		query = query.Where("method_code LIKE ?", "%"+search.MethodCode+"%")
	}
	if search.MethodName != "" {
		query = query.Where("method_name LIKE ?", "%"+search.MethodName+"%")
	}
	if search.Status != nil {
		query = query.Where("status = ?", *search.Status)
	}

	query.Count(&result.Total)
	query = query.Order("sort asc, id desc").Limit(search.PageSize).Offset(search.PageSize * search.CurrentPage)
	query.Find(&list)

	for _, item := range list {
		var temp pojo.SysPayMethodBack
		_ = copier.Copy(&temp, &item)
		result.List = append(result.List, temp)
	}

	result.PageSize = search.PageSize
	result.CurrentPage = search.CurrentPage
	return result
}

// SetSysPayMethod 创建或更新支付方式
func SetSysPayMethod(db *gorm.DB, req pojo.SysPayMethodSet) (result pojo.SysPayMethodBack, err error) {
	var entity pojo.SysPayMethod
	if req.ID > 0 {
		db.Where("id = ? AND deleted_at = 0", req.ID).First(&entity)
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

// DelSysPayMethod 软删除支付方式
func DelSysPayMethod(db *gorm.DB, id int64) (result string, err error) {
	var entity pojo.SysPayMethod
	db.Where("id = ? AND deleted_at = 0", id).First(&entity)
	if entity.ID == 0 {
		return result, errors.New("删除的数据不存在")
	}
	err = db.Model(&entity).Update("deleted_at", time.Now().Unix()).Error
	if err != nil {
		return result, err
	}
	return "success", nil
}

// GetSysPayMethodById 根据ID获取支付方式
func GetSysPayMethodById(db *gorm.DB, id int64) (result pojo.SysPayMethodBack, err error) {
	var entity pojo.SysPayMethod
	db.Where("id = ? AND deleted_at = 0", id).First(&entity)
	if entity.ID == 0 {
		return result, errors.New("数据不存在")
	}
	_ = copier.Copy(&result, &entity)
	return result, nil
}
