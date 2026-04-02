package repository

import (
	"BaseGoUni/core/pojo"
	"errors"
	"strings"
	"time"

	"github.com/jinzhu/copier"
	"gorm.io/gorm"
)

// GetSysBanners 轮播图列表（分页）
func GetSysBanners(db *gorm.DB, search pojo.SysBannerSearch) (result pojo.SysBannerResp) {
	var list []pojo.SysBanner
	query := db.Model(&pojo.SysBanner{})

	if search.TenantId > 0 {
		query = query.Where("tenant_id = ?", search.TenantId)
	}
	if search.BannerName != "" {
		query = query.Where("banner_name LIKE ?", "%"+strings.TrimSpace(search.BannerName)+"%")
	}
	if search.Position != "" {
		query = query.Where("position = ?", strings.TrimSpace(search.Position))
	}
	if search.Platform != "" {
		query = query.Where("platform = ?", strings.TrimSpace(search.Platform))
	}
	if search.JumpType != "" {
		query = query.Where("jump_type = ?", strings.TrimSpace(search.JumpType))
	}
	if search.Status != nil {
		query = query.Where("status = ?", *search.Status)
	}

	query.Count(&result.Total)
	query = query.Order("sort asc, id desc").Limit(search.PageSize).Offset(search.PageSize * search.CurrentPage)
	query.Find(&list)

	for _, item := range list {
		var temp pojo.SysBannerBack
		_ = copier.Copy(&temp, &item)
		result.List = append(result.List, temp)
	}

	result.PageSize = search.PageSize
	result.CurrentPage = search.CurrentPage
	return result
}

// SetSysBanner 创建或更新轮播图
func SetSysBanner(db *gorm.DB, req pojo.SysBannerSet) (result pojo.SysBannerBack, err error) {
	var entity pojo.SysBanner
	if req.ID > 0 {
		db.Where("id = ?", req.ID).First(&entity)
		if entity.ID == 0 {
			return result, errors.New("更新的数据不存在")
		}
	}

	_ = copier.Copy(&entity, &req)
	entity.BannerName = strings.TrimSpace(entity.BannerName)
	entity.Position = strings.TrimSpace(entity.Position)
	entity.Platform = strings.TrimSpace(entity.Platform)
	entity.ImageURL = strings.TrimSpace(entity.ImageURL)
	entity.JumpType = strings.TrimSpace(entity.JumpType)
	entity.StartTime = unixToTimePtr(req.StartTime)
	entity.EndTime = unixToTimePtr(req.EndTime)

	if req.ID > 0 {
		err = db.Save(&entity).Error
	} else {
		err = db.Create(&entity).Error
	}
	if err != nil {
		return result, err
	}

	_ = copier.Copy(&result, &entity)
	return result, nil
}

// DelSysBanner 删除轮播图
func DelSysBanner(db *gorm.DB, id int64) (result string, err error) {
	var entity pojo.SysBanner
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

// GetSysBannerById 根据ID获取轮播图
func GetSysBannerById(db *gorm.DB, id int64) (result pojo.SysBannerBack, err error) {
	var entity pojo.SysBanner
	db.Where("id = ?", id).First(&entity)
	if entity.ID == 0 {
		return result, errors.New("数据不存在")
	}
	_ = copier.Copy(&result, &entity)
	return result, nil
}

// GetSysBannersGroupedByPosition App端获取有效轮播图并按position分组
// platform: 请求平台(app/h5/web)；返回 platform=all 及匹配平台的启用记录，按 sort asc 排列
func GetSysBannersGroupedByPosition(db *gorm.DB, platform string) pojo.SysBannerGroupedResp {
	var list []pojo.SysBanner
	now := time.Now()
	query := db.Model(&pojo.SysBanner{}).
		Where("status = ?", 1).
		Where("(start_time IS NULL OR start_time <= ?)", now).
		Where("(end_time IS NULL OR end_time >= ?)", now).
		Order("sort asc, id desc")

	if platform != "" {
		query = query.Where("platform IN ?", []string{"all", platform})
	}

	query.Find(&list)

	result := pojo.SysBannerGroupedResp{}
	for _, item := range list {
		var temp pojo.SysBannerBack
		_ = copier.Copy(&temp, &item)
		result[item.Position] = append(result[item.Position], temp)
	}
	return result
}

func unixToTimePtr(ts *int64) *time.Time {
	if ts == nil || *ts <= 0 {
		return nil
	}
	t := time.Unix(*ts, 0)
	return &t
}
