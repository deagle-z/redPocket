package api

import (
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/utils"
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
	"gorm.io/gorm"
)

// GetAuthGroups 获取授权群组列表
// @Summary 获取授权群组列表
// @Tags 授权群组管理
// @Accept json
// @Produce json
// @Param data body pojo.AuthGroupSearch true "查询条件"
// @Success 200 {object} pojo.AuthGroupResp
// @Router /api/v1/admin/authGroup/list [post]
func GetAuthGroups(ctx *gin.Context) {
	var search pojo.AuthGroupSearch
	search.SetPageDefaults()
	if err := ctx.ShouldBindJSON(&search); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}

	db := ctx.MustGet("db").(*gorm.DB)

	var authGroups []pojo.AuthGroup
	query := db.Model(&pojo.AuthGroup{})

	if search.GroupID > 0 {
		query = query.Where("group_id = ?", search.GroupID)
	}
	if search.Status != nil {
		query = query.Where("status = ?", *search.Status)
	}

	var total int64
	query.Count(&total)
	query = query.Order("id desc").Limit(search.PageSize).Offset(search.PageSize * search.CurrentPage)
	query.Debug().Find(&authGroups)

	var result pojo.AuthGroupResp
	for _, group := range authGroups {
		var tempGroup pojo.AuthGroupBack
		_ = copier.Copy(&tempGroup, &group)
		result.List = append(result.List, tempGroup)
	}

	result.Total = total
	result.PageSize = search.PageSize
	result.CurrentPage = search.CurrentPage

	utils.SuccessObjBack(ctx, result)
}

// SetAuthGroup 创建或更新授权群组
// @Summary 创建或更新授权群组
// @Tags 授权群组管理
// @Accept json
// @Produce json
// @Param data body pojo.AuthGroup true "群组信息"
// @Success 200 {object} pojo.BaseObjResponse[pojo.AuthGroupBack]
// @Router /api/v1/admin/authGroup [post]
func SetAuthGroup(ctx *gin.Context) {
	var req pojo.AuthGroup
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}

	db := ctx.MustGet("db").(*gorm.DB)

	// 如果是更新操作，先获取旧的 group_id 用于清除缓存
	var oldGroupID int64
	if req.ID > 0 {
		var oldAuthGroup pojo.AuthGroup
		if err := db.Where("id = ?", req.ID).First(&oldAuthGroup).Error; err == nil {
			oldGroupID = oldAuthGroup.GroupID
		}
	}

	if req.ID > 0 {
		// 更新
		if err := db.Save(&req).Error; err != nil {
			utils.ErrorBack(ctx, err.Error())
			return
		}
	} else {
		// 创建
		if err := db.Create(&req).Error; err != nil {
			utils.ErrorBack(ctx, err.Error())
			return
		}
	}

	// 清除相关缓存
	clearAuthGroupCache(req.GroupID)
	// 如果是更新且 group_id 发生变化，也要清除旧 group_id 的缓存
	if req.ID > 0 && oldGroupID > 0 && oldGroupID != req.GroupID {
		clearAuthGroupCache(oldGroupID)
	}

	var result pojo.AuthGroupBack
	_ = copier.Copy(&result, &req)
	utils.SuccessObjBack(ctx, result)
}

// DelAuthGroup 删除授权群组
// @Summary 删除授权群组
// @Tags 授权群组管理
// @Accept json
// @Produce json
// @Param id path int true "群组ID"
// @Success 200 {object} pojo.BaseResponse
// @Router /api/v1/admin/authGroup/:id [delete]
func DelAuthGroup(ctx *gin.Context) {
	idStr := ctx.Param("id")
	var id int64
	_, err := fmt.Sscanf(idStr, "%d", &id)
	if err != nil || id <= 0 {
		utils.ErrorBack(ctx, "无效的群组ID")
		return
	}

	db := ctx.MustGet("db").(*gorm.DB)

	// 删除前先获取 group_id 用于清除缓存
	var authGroup pojo.AuthGroup
	if err := db.Where("id = ?", id).First(&authGroup).Error; err == nil {
		// 清除相关缓存
		clearAuthGroupCache(authGroup.GroupID)
	}

	if err := db.Delete(&pojo.AuthGroup{}, id).Error; err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}

	utils.SuccessBack(ctx, "删除成功")
}

// clearAuthGroupCache 清除授权群组相关的Redis缓存
func clearAuthGroupCache(groupID int64) {
	if groupID <= 0 {
		return
	}
	ctx := context.Background()
	redisKeys := []string{
		fmt.Sprintf("bgu_auth_group_lose_rate_%d", groupID),
		fmt.Sprintf("bgu_auth_group_num_config_%d", groupID),
		fmt.Sprintf("bgu_auth_group_send_commission_%d", groupID),
		fmt.Sprintf("bgu_auth_group_grabbing_commission_%d", groupID),
	}
	for _, key := range redisKeys {
		_ = utils.RD.Del(ctx, key).Err()
	}
}
