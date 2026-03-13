package pojo

import "time"

type SysCountry struct {
	BaseModel
	CountryCode    string  `json:"countryCode" gorm:"column:country_code;type:varchar(8);uniqueIndex;not null;comment:国家编码，如 IN/BR/PH"`
	CountryNameCn  string  `json:"countryNameCn" gorm:"column:country_name_cn;type:varchar(64);not null;comment:国家中文名"`
	CountryNameEn  string  `json:"countryNameEn" gorm:"column:country_name_en;type:varchar(64);not null;comment:国家英文名"`
	CurrencyCode   string  `json:"currencyCode" gorm:"column:currency_code;type:varchar(16);not null;comment:默认币种编码，如 INR/BRL/USD"`
	CurrencySymbol *string `json:"currencySymbol" gorm:"column:currency_symbol;type:varchar(16);comment:货币符号，如 ₹ R$ $"`
	Timezone       *string `json:"timezone" gorm:"column:timezone;type:varchar(64);comment:时区"`
	LanguageCode   *string `json:"languageCode" gorm:"column:language_code;type:varchar(32);comment:语言编码，如 en-IN pt-BR"`
	WithdrawFields *string `json:"withdrawFields" gorm:"column:withdraw_fields;type:json;comment:提现字段配置"`
	RechargeFields *string `json:"rechargeFields" gorm:"column:recharge_fields;type:json;comment:充值字段配置"`
	Sort           int     `json:"sort" gorm:"column:sort;type:int;default:0;comment:排序"`
	Status         int8    `json:"status" gorm:"column:status;type:tinyint;default:1;not null;comment:状态 1启用 0禁用"`
	Remark         *string `json:"remark" gorm:"column:remark;type:varchar(255);comment:备注"`
}

type SysCountrySearch struct {
	PageInfo
	CountryCode  string `json:"countryCode"`
	CountryName  string `json:"countryName"`
	CurrencyCode string `json:"currencyCode"`
	Status       *int8  `json:"status"`
}

type SysCountrySet struct {
	ID             int64   `json:"id"`
	CountryCode    string  `json:"countryCode"`
	CountryNameCn  string  `json:"countryNameCn"`
	CountryNameEn  string  `json:"countryNameEn"`
	CurrencyCode   string  `json:"currencyCode"`
	CurrencySymbol *string `json:"currencySymbol"`
	Timezone       *string `json:"timezone"`
	LanguageCode   *string `json:"languageCode"`
	WithdrawFields *string `json:"withdrawFields"`
	RechargeFields *string `json:"rechargeFields"`
	Sort           int     `json:"sort"`
	Status         int8    `json:"status"`
	Remark         *string `json:"remark"`
}

type SysCountryBack struct {
	ID             int64     `json:"id"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
	CountryCode    string    `json:"countryCode"`
	CountryNameCn  string    `json:"countryNameCn"`
	CountryNameEn  string    `json:"countryNameEn"`
	CurrencyCode   string    `json:"currencyCode"`
	CurrencySymbol *string   `json:"currencySymbol"`
	Timezone       *string   `json:"timezone"`
	LanguageCode   *string   `json:"languageCode"`
	WithdrawFields *string   `json:"withdrawFields"`
	RechargeFields *string   `json:"rechargeFields"`
	Sort           int       `json:"sort"`
	Status         int8      `json:"status"`
	Remark         *string   `json:"remark"`
}

type SysCountryResp struct {
	BasePageResponse[SysCountryBack]
}

var SysCountryTableName = "sys_country"

func (SysCountry) TableName() string {
	return SysCountryTableName
}
