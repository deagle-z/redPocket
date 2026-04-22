package repository

import (
	"BaseGoUni/core/pojo"
	"errors"
	"strings"
	"time"

	"github.com/jinzhu/copier"
	"gorm.io/gorm"
)

// GetSysBanners Banner列表（分页）
func GetSysBanners(db *gorm.DB, search pojo.SysBannerSearch) (result pojo.SysBannerResp) {
	var list []pojo.SysBanner
	query := db.Model(&pojo.SysBanner{}).Where("is_deleted = ?", 0)

	if search.TenantId > 0 {
		query = query.Where("tenant_id = ?", search.TenantId)
	}
	if search.BannerName != "" {
		query = query.Where("banner_name LIKE ?", "%"+strings.TrimSpace(search.BannerName)+"%")
	}
	if search.BannerCode != "" {
		query = query.Where("banner_code LIKE ?", "%"+strings.TrimSpace(search.BannerCode)+"%")
	}
	if search.Position != "" {
		query = query.Where("position = ?", strings.TrimSpace(search.Position))
	}
	if search.Platform != "" {
		query = query.Where("platform = ?", strings.TrimSpace(search.Platform))
	}
	if search.BannerType != "" {
		query = query.Where("banner_type = ?", strings.TrimSpace(search.BannerType))
	}
	if search.JumpType != "" {
		query = query.Where("jump_type = ?", strings.TrimSpace(search.JumpType))
	}
	if search.DisplayType != "" {
		query = query.Where("display_type = ?", strings.TrimSpace(search.DisplayType))
	}
	if search.Status != nil {
		query = query.Where("status = ?", *search.Status)
	}
	if search.LanguageCode != "" {
		query = query.Where("id IN (?)",
			db.Model(&pojo.SysBannerI18n{}).
				Select("banner_id").
				Where("is_deleted = ?", 0).
				Where("language_code = ?", strings.TrimSpace(search.LanguageCode)),
		)
	}
	if search.CountryCode != "" {
		query = query.Where("id IN (?)",
			db.Model(&pojo.SysBannerCountryRel{}).
				Select("banner_id").
				Where("country_code = ?", strings.TrimSpace(search.CountryCode)),
		)
	}

	query.Count(&result.Total)
	query = query.Order("sort asc, id desc").Limit(search.PageSize).Offset(search.PageSize * search.CurrentPage)
	query.Find(&list)

	result.List = buildSysBannerBackList(db, list)
	result.PageSize = search.PageSize
	result.CurrentPage = search.CurrentPage
	return result
}

// SetSysBanner 创建或更新Banner
func SetSysBanner(db *gorm.DB, req pojo.SysBannerSet) (result pojo.SysBannerBack, err error) {
	err = db.Transaction(func(tx *gorm.DB) error {
		var entity pojo.SysBanner
		if req.ID > 0 {
			tx.Where("id = ?", req.ID).First(&entity)
			if entity.ID == 0 {
				return errors.New("record_not_found_update")
			}
		}

		_ = copier.Copy(&entity, &req)
		entity.BannerName = strings.TrimSpace(entity.BannerName)
		entity.Position = strings.TrimSpace(entity.Position)
		entity.Platform = strings.TrimSpace(entity.Platform)
		entity.BannerType = strings.TrimSpace(entity.BannerType)
		entity.JumpType = strings.TrimSpace(entity.JumpType)
		entity.DisplayType = strings.TrimSpace(entity.DisplayType)
		entity.OpenMode = strings.TrimSpace(entity.OpenMode)
		entity.StartTime = unixToTimePtr(req.StartTime)
		entity.EndTime = unixToTimePtr(req.EndTime)
		entity.IsDeleted = 0

		if req.ID > 0 {
			if err = tx.Save(&entity).Error; err != nil {
				return err
			}
		} else {
			if err = tx.Create(&entity).Error; err != nil {
				return err
			}
		}

		if err = tx.Where("banner_id = ?", entity.ID).Delete(&pojo.SysBannerI18n{}).Error; err != nil {
			return err
		}
		if err = tx.Where("banner_id = ?", entity.ID).Delete(&pojo.SysBannerCountryRel{}).Error; err != nil {
			return err
		}

		for _, item := range req.I18nList {
			i18n := pojo.SysBannerI18n{}
			_ = copier.Copy(&i18n, &item)
			i18n.TenantId = entity.TenantId
			i18n.BannerId = entity.ID
			i18n.LanguageCode = strings.TrimSpace(i18n.LanguageCode)
			if i18n.CountryCode != nil {
				trimmed := strings.TrimSpace(*i18n.CountryCode)
				if trimmed == "" {
					i18n.CountryCode = nil
				} else {
					i18n.CountryCode = &trimmed
				}
			}
			i18n.ImageURL = strings.TrimSpace(i18n.ImageURL)
			if i18n.ImageURL == "" {
				continue
			}
			if err = tx.Create(&i18n).Error; err != nil {
				return err
			}
		}

		for _, item := range req.CountryList {
			rel := pojo.SysBannerCountryRel{}
			_ = copier.Copy(&rel, &item)
			rel.TenantId = entity.TenantId
			rel.BannerId = entity.ID
			rel.CountryCode = strings.TrimSpace(rel.CountryCode)
			if rel.CountryCode == "" {
				continue
			}
			if err = tx.Create(&rel).Error; err != nil {
				return err
			}
		}

		resultList := buildSysBannerBackList(tx, []pojo.SysBanner{entity})
		if len(resultList) == 0 {
			return errors.New("service_busy_retry")
		}
		result = resultList[0]
		return nil
	})
	return result, err
}

// DelSysBanner 删除Banner
func DelSysBanner(db *gorm.DB, id int64) (result string, err error) {
	var entity pojo.SysBanner
	db.Where("id = ?", id).First(&entity)
	if entity.ID == 0 {
		return result, errors.New("record_not_found_delete")
	}
	err = db.Transaction(func(tx *gorm.DB) error {
		if err = tx.Model(&pojo.SysBanner{}).Where("id = ?", id).Update("is_deleted", 1).Error; err != nil {
			return err
		}
		if err = tx.Model(&pojo.SysBannerI18n{}).Where("banner_id = ?", id).Update("is_deleted", 1).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return result, err
	}
	return "success", nil
}

// GetSysBannerById 根据ID获取Banner
func GetSysBannerById(db *gorm.DB, id int64) (result pojo.SysBannerBack, err error) {
	var entity pojo.SysBanner
	db.Where("id = ?", id).Where("is_deleted = ?", 0).First(&entity)
	if entity.ID == 0 {
		return result, errors.New("record_not_found")
	}
	list := buildSysBannerBackList(db, []pojo.SysBanner{entity})
	if len(list) == 0 {
		return result, errors.New("record_not_found")
	}
	return list[0], nil
}

// GetSysBannersGroupedByPosition App端获取有效Banner并按position分组
func GetSysBannersGroupedByPosition(db *gorm.DB, req pojo.SysBannerAppReq) pojo.SysBannerGroupedResp {
	var list []pojo.SysBanner
	now := time.Now()
	query := db.Model(&pojo.SysBanner{}).
		Where("status = ?", 1).
		Where("is_deleted = ?", 0).
		Where("(start_time IS NULL OR start_time <= ?)", now).
		Where("(end_time IS NULL OR end_time >= ?)", now).
		Order("sort asc, id desc")

	if req.Platform != "" {
		query = query.Where("platform IN ?", []string{"all", req.Platform})
	}
	if req.Position != "" {
		query = query.Where("position = ?", strings.TrimSpace(req.Position))
	}
	if req.CountryCode != "" {
		query = query.Where("id IN (?)",
			db.Model(&pojo.SysBannerCountryRel{}).
				Select("banner_id").
				Where("status = ?", 1).
				Where("country_code = ?", strings.TrimSpace(req.CountryCode)),
		)
	}

	query.Find(&list)

	i18nMap, _ := loadBannerI18nMap(db, collectBannerIDs(list), true)
	result := pojo.SysBannerGroupedResp{}
	lang := strings.TrimSpace(req.Lang)
	if lang == "" {
		lang = strings.TrimSpace(req.LegacyLanguageCode)
	}
	for _, item := range list {
		selected := pickAppI18n(i18nMap[item.ID], lang, req.CountryCode)
		if selected == nil {
			continue
		}
		temp := pojo.SysBannerAppItem{
			ID:           item.ID,
			TenantId:     item.TenantId,
			BannerName:   item.BannerName,
			BannerCode:   item.BannerCode,
			Position:     item.Position,
			Platform:     item.Platform,
			BannerType:   item.BannerType,
			JumpType:     item.JumpType,
			DisplayType:  item.DisplayType,
			OpenMode:     item.OpenMode,
			Sort:         item.Sort,
			Status:       item.Status,
			StartTime:    item.StartTime,
			EndTime:      item.EndTime,
			LanguageCode: selected.LanguageCode,
			CountryCode:  selected.CountryCode,
			Title:        selected.Title,
			SubTitle:     selected.SubTitle,
			Description:  selected.Description,
			ButtonText:   selected.ButtonText,
			ImageURL:     selected.ImageURL,
			ThumbURL:     selected.ThumbURL,
			BgImageURL:   selected.BgImageURL,
			IconURL:      selected.IconURL,
			VideoURL:     selected.VideoURL,
			JumpValue:    selected.JumpValue,
			TextColor:    selected.TextColor,
			ButtonColor:  selected.ButtonColor,
			BgColor:      selected.BgColor,
		}
		result[item.Position] = append(result[item.Position], temp)
	}
	return result
}

func buildSysBannerBackList(db *gorm.DB, banners []pojo.SysBanner) []pojo.SysBannerBack {
	result := make([]pojo.SysBannerBack, 0, len(banners))
	if len(banners) == 0 {
		return result
	}

	ids := collectBannerIDs(banners)
	i18nMap, _ := loadBannerI18nMap(db, ids, false)
	countryMap := loadBannerCountryRelMap(db, ids)

	for _, item := range banners {
		var temp pojo.SysBannerBack
		_ = copier.Copy(&temp, &item)
		temp.I18nList = i18nMap[item.ID]
		temp.CountryList = countryMap[item.ID]
		result = append(result, temp)
	}
	return result
}

func collectBannerIDs(list []pojo.SysBanner) []int64 {
	ids := make([]int64, 0, len(list))
	for _, item := range list {
		ids = append(ids, item.ID)
	}
	return ids
}

func loadBannerI18nMap(db *gorm.DB, bannerIDs []int64, enabledOnly bool) (map[int64][]pojo.SysBannerI18nBack, []pojo.SysBannerI18n) {
	result := map[int64][]pojo.SysBannerI18nBack{}
	if len(bannerIDs) == 0 {
		return result, nil
	}

	var list []pojo.SysBannerI18n
	query := db.Model(&pojo.SysBannerI18n{}).Where("banner_id IN ?", bannerIDs).Where("is_deleted = ?", 0)
	if enabledOnly {
		query = query.Where("status = ?", 1)
	}
	query.Order("id asc").Find(&list)

	for _, item := range list {
		var temp pojo.SysBannerI18nBack
		_ = copier.Copy(&temp, &item)
		result[item.BannerId] = append(result[item.BannerId], temp)
	}
	return result, list
}

func loadBannerCountryRelMap(db *gorm.DB, bannerIDs []int64) map[int64][]pojo.SysBannerCountryRelBack {
	result := map[int64][]pojo.SysBannerCountryRelBack{}
	if len(bannerIDs) == 0 {
		return result
	}

	var list []pojo.SysBannerCountryRel
	db.Model(&pojo.SysBannerCountryRel{}).
		Where("banner_id IN ?", bannerIDs).
		Order("id asc").
		Find(&list)

	for _, item := range list {
		var temp pojo.SysBannerCountryRelBack
		_ = copier.Copy(&temp, &item)
		result[item.BannerId] = append(result[item.BannerId], temp)
	}
	return result
}

func pickAppI18n(list []pojo.SysBannerI18nBack, languageCode, countryCode string) *pojo.SysBannerI18nBack {
	if len(list) == 0 {
		return nil
	}
	lang := strings.TrimSpace(languageCode)
	country := strings.TrimSpace(countryCode)

	if lang != "" && country != "" {
		for i := range list {
			if list[i].LanguageCode == lang && list[i].CountryCode != nil && *list[i].CountryCode == country {
				return &list[i]
			}
		}
	}
	if lang != "" {
		for i := range list {
			if list[i].LanguageCode == lang && list[i].CountryCode == nil {
				return &list[i]
			}
		}
		for i := range list {
			if list[i].LanguageCode == lang {
				return &list[i]
			}
		}
	}
	if country != "" {
		for i := range list {
			if list[i].LanguageCode == "all" && list[i].CountryCode != nil && *list[i].CountryCode == country {
				return &list[i]
			}
		}
	}
	for i := range list {
		if list[i].LanguageCode == "all" && list[i].CountryCode == nil {
			return &list[i]
		}
	}
	for i := range list {
		if list[i].LanguageCode == "all" {
			return &list[i]
		}
	}
	if country != "" {
		for i := range list {
			if list[i].CountryCode != nil && *list[i].CountryCode == country {
				return &list[i]
			}
		}
	}
	return &list[0]
}

func unixToTimePtr(ts *int64) *time.Time {
	if ts == nil || *ts <= 0 {
		return nil
	}
	t := time.Unix(*ts, 0)
	return &t
}
