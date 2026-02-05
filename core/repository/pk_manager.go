package repository

import (
	"BaseGoUni/core/pojo"
	"errors"
	"github.com/jinzhu/copier"
	"gorm.io/gorm"
)

func GetPkManagers(db *gorm.DB, search pojo.PkManagerSearch) (result pojo.PkManagerResp) {
	var pkManagers []pojo.PkManager
	db = db.Model(&pojo.PkManager{})
	
	if search.ApkPackage != "" {
		db = db.Where("apk_package like ?", "%"+search.ApkPackage+"%")
	}
	if search.Name != "" {
		db = db.Where("name like ?", "%"+search.Name+"%")
	}
	if search.IsActive != nil {
		db = db.Where("is_active = ?", *search.IsActive)
	}
	
	db.Count(&result.Total)
	db = db.Order("id desc").Limit(search.PageSize).Offset(search.PageSize * search.CurrentPage)
	db.Find(&pkManagers)
	
	for _, pkManager := range pkManagers {
		var tempPkManagerBack pojo.PkManagerBack
		_ = copier.Copy(&tempPkManagerBack, &pkManager)
		result.List = append(result.List, tempPkManagerBack)
	}
	
	result.PageSize = search.PageSize
	result.CurrentPage = search.CurrentPage
	return result
}

func SetPkManager(db *gorm.DB, pkManagerSet pojo.PkManagerSet) (result pojo.PkManagerBack, err error) {
	var dbPkManager pojo.PkManager
	if pkManagerSet.ID > 0 {
		db.Where("id = ?", pkManagerSet.ID).First(&dbPkManager)
		if dbPkManager.ID == 0 {
			return result, errors.New("更新的数据不存在")
		}
		_ = copier.Copy(&dbPkManager, &pkManagerSet)
		err = db.Save(&dbPkManager).Error
	} else {
		_ = copier.Copy(&dbPkManager, &pkManagerSet)
		err = db.Create(&dbPkManager).Error
	}
	if err != nil {
		return result, err
	}
	_ = copier.Copy(&result, &dbPkManager)
	return result, nil
}

func DelPkManager(db *gorm.DB, id int64) (result string, err error) {
	var dbPkManager pojo.PkManager
	db.Where("id = ?", id).First(&dbPkManager)
	if dbPkManager.ID == 0 {
		return result, errors.New("删除的数据不存在")
	}
	err = db.Delete(&dbPkManager).Error
	if err != nil {
		return result, err
	}
	return "success", nil
}

func GetPkManagerById(db *gorm.DB, id int64) (result pojo.PkManagerBack, err error) {
	var dbPkManager pojo.PkManager
	db.Where("id = ?", id).First(&dbPkManager)
	if dbPkManager.ID == 0 {
		return result, errors.New("数据不存在")
	}
	_ = copier.Copy(&result, &dbPkManager)
	return result, nil
}

func GetPkManagerUrlByName(db *gorm.DB, name string) (url string, err error) {
	var dbPkManager pojo.PkManager
	db.Where("name = ? AND is_active = ?", name, 1).First(&dbPkManager)
	if dbPkManager.ID == 0 {
		return "", errors.New("数据不存在或已停用")
	}
	return dbPkManager.Url, nil
}


