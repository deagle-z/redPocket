package repository

import (
	"BaseGoUni/core/pojo"
	"errors"
	"github.com/jinzhu/copier"
	"gorm.io/gorm"
)

func GetAuthGroups(db *gorm.DB, tenantID int64, search pojo.AuthGroupSearch) (result pojo.AuthGroupResp) {
	var authGroups []pojo.AuthGroup
	query := db.Model(&pojo.AuthGroup{}).Where("tenant_id = ?", tenantID)
	if search.GroupID > 0 {
		query = query.Where("group_id = ?", search.GroupID)
	}
	if search.Status != nil {
		query = query.Where("status = ?", *search.Status)
	}
	query.Count(&result.Total)
	query = query.Order("id desc").Limit(search.PageSize).Offset(search.PageSize * search.CurrentPage)
	query.Find(&authGroups)
	for _, group := range authGroups {
		var tempGroup pojo.AuthGroupBack
		_ = copier.Copy(&tempGroup, &group)
		result.List = append(result.List, tempGroup)
	}
	result.PageSize = search.PageSize
	result.CurrentPage = search.CurrentPage
	return result
}

func SetAuthGroup(db *gorm.DB, tenantID int64, req pojo.AuthGroup) (result pojo.AuthGroupBack, err error) {
	req.TenantId = tenantID
	var dbData pojo.AuthGroup
	if req.ID > 0 {
		db.Where("id = ? and tenant_id = ?", req.ID, tenantID).First(&dbData)
		if dbData.ID == 0 {
			return result, errors.New("更新的数据不存在")
		}
		_ = copier.Copy(&dbData, &req)
		err = db.Save(&dbData).Error
	} else {
		_ = copier.Copy(&dbData, &req)
		err = db.Create(&dbData).Error
	}
	if err != nil {
		return result, err
	}
	_ = copier.Copy(&result, &dbData)
	return result, nil
}

func DelAuthGroup(db *gorm.DB, tenantID int64, id int64) (groupID int64, result string, err error) {
	var dbData pojo.AuthGroup
	db.Where("id = ? and tenant_id = ?", id, tenantID).First(&dbData)
	if dbData.ID == 0 {
		return groupID, result, errors.New("删除的数据不存在")
	}
	groupID = dbData.GroupID
	if err = db.Delete(&dbData).Error; err != nil {
		return groupID, result, err
	}
	return groupID, "success", nil
}
