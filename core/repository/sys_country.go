package repository

import (
	"BaseGoUni/core/pojo"
	"encoding/json"
	"errors"
	"github.com/jinzhu/copier"
	"gorm.io/gorm"
	"log"
	"strings"
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
	if req.WithdrawFields == "" {
		req.WithdrawFields = "[]"
	}
	if req.RechargeFields == "" {
		req.RechargeFields = "[]"
	}
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

// GetAppCountries App端获取所有可用国家。
// detectedCode 为 IP 解析出的国家码，若命中则将该国家置于列表首位；否则按 sort 默认排序。
func GetAppCountries(db *gorm.DB, detectedCode string) []pojo.AppCountryItem {
	var list []pojo.SysCountry
	db.Where("status = 1").Order("sort asc, id asc").Find(&list)

	result := make([]pojo.AppCountryItem, 0, len(list))
	var detected *pojo.AppCountryItem

	for _, item := range list {
		ci := pojo.AppCountryItem{
			ID:             item.ID,
			CountryCode:    item.CountryCode,
			CountryNameEn:  item.CountryNameEn,
			CountryNameCn:  item.CountryNameCn,
			CurrencyCode:   item.CurrencyCode,
			CurrencySymbol: item.CurrencySymbol,
			Sort:           item.Sort,
		}
		if detectedCode != "" && strings.EqualFold(item.CountryCode, detectedCode) {
			detected = &ci
		} else {
			result = append(result, ci)
		}
	}

	if detected != nil {
		return append([]pojo.AppCountryItem{*detected}, result...)
	}
	return result
}

// GetCountryRechargeInfo App端获取国家充值信息（充值字段 + 通道 + 支付方式）
func GetCountryRechargeInfo(db *gorm.DB, countryCode string) (result pojo.AppCountryRechargeInfo, err error) {
	result.Channels = []pojo.AppRechargeChannelItem{}

	// 1. 验证国家存在且启用
	var country pojo.SysCountry
	db.Where("country_code = ? AND status = 1", countryCode).First(&country)
	if country.ID == 0 {
		return result, errors.New("国家不存在或已禁用")
	}

	// 2. 解析充值字段配置
	if country.RechargeFields != nil && *country.RechargeFields != "" {
		var fields interface{}
		if jsonErr := json.Unmarshal([]byte(*country.RechargeFields), &fields); jsonErr == nil {
			result.RechargeFields = fields
		} else {
			result.RechargeFields = []interface{}{}
		}
	} else {
		result.RechargeFields = []interface{}{}
	}

	// 3. 获取适用该国家的充值通道（country_code 匹配或为空=全局通道）
	var channels []pojo.SysPayChannel
	db.Where(
		"channel_type IN ? AND (country_code = ? OR country_code IS NULL OR country_code = '') AND status = 1 AND deleted_at = 0",
		[]string{"deposit", "both"}, countryCode,
	).Order("sort asc, id asc").Find(&channels)

	// 4. 为每个通道加载支付方式
	for _, ch := range channels {
		item := pojo.AppRechargeChannelItem{
			ID:           ch.ID,
			ChannelCode:  ch.ChannelCode,
			ChannelName:  ch.ChannelName,
			ProviderType: ch.ProviderType,
			Icon:         ch.Icon,
			Sort:         ch.Sort,
			Methods:      []pojo.AppPayMethodItem{},
		}

		var bindings []pojo.SysPayChannelMethod
		db.Where("channel_id = ?", ch.ID).Find(&bindings)
		if len(bindings) > 0 {
			methodIds := make([]int64, 0, len(bindings))
			for _, b := range bindings {
				methodIds = append(methodIds, b.MethodID)
			}
			var methods []pojo.SysPayMethod
			db.Where("id IN ? AND status = 1 AND deleted_at = 0", methodIds).Order("sort asc").Find(&methods)
			for _, m := range methods {
				item.Methods = append(item.Methods, pojo.AppPayMethodItem{
					ID:         m.ID,
					MethodCode: m.MethodCode,
					MethodName: m.MethodName,
					Icon:       m.Icon,
					Sort:       m.Sort,
				})
			}
		}

		result.Channels = append(result.Channels, item)
	}

	return result, nil
}

// parseFieldsJSON 将国家字段 JSON 字符串解析为 interface{}，失败返回空数组
func parseFieldsJSON(raw *string) interface{} {
	if raw == nil || *raw == "" {
		return []interface{}{}
	}
	var out interface{}
	if err := json.Unmarshal([]byte(*raw), &out); err != nil {
		return []interface{}{}
	}
	return out
}

// GetCountryWithdrawFields App端获取国家提现字段配置
func GetCountryWithdrawFields(db *gorm.DB, countryCode string) (interface{}, error) {
	var country pojo.SysCountry
	db.Where("country_code = ? AND status = 1", countryCode).First(&country)
	if country.ID == 0 {
		return nil, errors.New("国家不存在或已禁用")
	}
	return parseFieldsJSON(country.WithdrawFields), nil
}

// GetCountryRechargeFields App端获取国家充值字段配置
func GetCountryRechargeFields(db *gorm.DB, countryCode string) (interface{}, error) {
	var country pojo.SysCountry
	db.Debug().Where("country_code = ? AND status = 1", countryCode).First(&country)
	if country.ID == 0 {
		return nil, errors.New("国家不存在或已禁用")
	}
	log.Printf("RechargeFields: %v", country)
	return parseFieldsJSON(country.RechargeFields), nil
}
