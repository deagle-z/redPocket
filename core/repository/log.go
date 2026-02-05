package repository

import (
	"BaseGoUni/core/pojo"
	"gorm.io/gorm"
)

func GetManageLogs(db *gorm.DB, currentUser pojo.SysUser, searchData pojo.ManageLogSearch) (result pojo.ManageLogResp, err error) {
	query := db.Model(pojo.ManageLog{})
	if searchData.Username != "" {
		query.Where("username = ?", searchData.Username)
	}
	result.PageSize = searchData.PageSize
	result.CurrentPage = searchData.CurrentPage
	query.Count(&result.Total)
	err = query.Limit(result.PageSize).
		Offset(result.PageSize * result.CurrentPage).
		Order("id desc").
		Find(&result.List).Error
	return result, err
}
