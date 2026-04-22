package repository

import (
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/utils"
	"encoding/json"
	"errors"
	"github.com/jinzhu/copier"
	"gorm.io/gorm"
)

func GetRoleMenuIds(db *gorm.DB, roleId string) (result []int64, err error) {
	var role pojo.SysRole
	db.Where("id = ?", roleId).First(&role)
	if role.ID == 0 {
		return nil, errors.New("permission_not_found")
	}
	var menus []pojo.SysMenu
	if role.Code == "admin" {
		db.Find(&menus)
	} else {
		var menuIds []int64
		_ = json.Unmarshal([]byte(role.MenuIdStr), &menuIds)
		db.Where("id in ?", menuIds).Find(&menus)
	}
	result = make([]int64, 0)
	for _, menu := range menus {
		result = append(result, menu.ID)
	}
	return result, err
}

func GetRoleMenus(db *gorm.DB, roleSearch pojo.RoleSearch) (result []pojo.BackMenu, err error) {
	var menus []pojo.SysMenu
	db.Find(&menus)
	result = make([]pojo.BackMenu, 0)
	for _, menu := range menus {
		_ = json.Unmarshal([]byte(menu.MetaStr), &menu.Meta)
		var tempRoleBack pojo.BackMenu
		_ = copier.Copy(&tempRoleBack, &menu)
		_ = copier.Copy(&tempRoleBack, &menu.Meta)
		_ = copier.Copy(&tempRoleBack, &menu.Meta.Transition)
		tempRoleBack.NameCode = tempRoleBack.Meta.Title
		result = append(result, tempRoleBack)
	}
	return result, err
}

func GetRoleIds(db *gorm.DB, hostInfo pojo.HostInfo, userId int64) (result []pojo.RoleBack, err error) {
	searchUser := utils.GetTempUser(hostInfo.TablePrefix, userId)
	if searchUser.ID == 0 {
		return result, errors.New("user_not_found")
	}
	_ = json.Unmarshal([]byte(searchUser.RoleStr), &searchUser.Roles)
	var roles []pojo.SysRole
	db.Where("code in ?", searchUser.Roles).Find(&roles)
	result = make([]pojo.RoleBack, 0)
	for _, role := range roles {
		var tempRoleBack pojo.RoleBack
		_ = copier.Copy(&tempRoleBack, &role)
		_ = json.Unmarshal([]byte(role.MenuIdStr), &tempRoleBack.MenuIds)
		result = append(result, tempRoleBack)
	}
	return result, err
}

func DelRole(db *gorm.DB, currentUser pojo.SysUser, id string) (result string, err error) {
	var dbRole pojo.SysRole
	db.Where("id = ?", id).Find(&dbRole)
	if dbRole.ID == 0 {
		return result, errors.New("record_not_found_delete")
	}
	if dbRole.Code == "admin" {
		return result, errors.New("admin_role_cannot_delete")
	}
	if currentUser.UserType != 1 {
		_ = json.Unmarshal([]byte(currentUser.RoleStr), &currentUser.Roles)
		var roles []pojo.SysRole
		db.Where("id in ?", currentUser.Roles).Find(&roles)
		haveRole := currentUser.Username == "admin"
		if !haveRole {
			for _, role := range roles {
				var menuIds []string
				_ = json.Unmarshal([]byte(role.MenuIdStr), &menuIds)
				if utils.InStrings(menuIds, id) {
					haveRole = true
					break
				}
			}
		}
		if !haveRole {
			return result, errors.New("permission_delete_denied")
		}
	}
	db.Delete(&dbRole)
	return "success", nil
}

func SetRole(db *gorm.DB, roleSet pojo.RoleSet) (result string, err error) {
	var dbRole pojo.SysRole
	if roleSet.ID == 0 {
		db.Where("code = ?", roleSet.Code).
			First(&dbRole)
		if dbRole.ID != 0 {
			return "fail", errors.New("permission_code_not_unique")
		}
		_ = copier.Copy(&dbRole, &roleSet)
		menuStr, _ := json.Marshal(roleSet.MenuIds)
		dbRole.MenuIdStr = string(menuStr)
		err = db.Create(&dbRole).Error
	} else {
		db.Where("id = ?", roleSet.ID).
			First(&dbRole)
		if dbRole.ID == 0 {
			return "fail", errors.New("permission_not_found_update")
		}
		_ = copier.Copy(&dbRole, roleSet)
		menuStr, _ := json.Marshal(roleSet.MenuIds)
		dbRole.MenuIdStr = string(menuStr)
		err = db.Save(&dbRole).Error
	}
	return "success", err
}

func GetRoles(db *gorm.DB, roleSearch pojo.RoleSearch) (result pojo.RoleResp) {
	var roles []pojo.SysRole
	if roleSearch.Code != "" {
		db = db.Where("code like ?", "%"+roleSearch.Code+"%")
	}
	if roleSearch.Name != "" {
		db = db.Where("name like ?", "%"+roleSearch.Name+"%")
	}
	db = db.Table(pojo.RoleTableName)
	db.Count(&result.Total)
	db.Find(&roles)
	for _, role := range roles {
		var tempRoleBack pojo.RoleBack
		_ = copier.Copy(&tempRoleBack, &role)
		_ = json.Unmarshal([]byte(role.MenuIdStr), &tempRoleBack.MenuIds)
		result.List = append(result.List, tempRoleBack)
	}
	result.PageSize = 100
	result.CurrentPage = 0
	return result
}
