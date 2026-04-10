package pojo

import "time"

// SysBanner 租户Banner主表
type SysBanner struct {
	BaseModel
	TenantId    int64      `json:"tenantId" gorm:"column:tenant_id;type:bigint unsigned;not null;uniqueIndex:uk_tenant_banner_code;index:idx_tenant_position_platform;index:idx_tenant_status_deleted_sort;index:idx_tenant_time_range;index:idx_tenant_display_type;index:idx_tenant_jump_type;comment:租户ID"`
	BannerName  string     `json:"bannerName" gorm:"column:banner_name;type:varchar(128);not null;comment:Banner内部管理名称"`
	BannerCode  *string    `json:"bannerCode" gorm:"column:banner_code;type:varchar(64);uniqueIndex:uk_tenant_banner_code;comment:Banner编码，便于程序识别"`
	Position    string     `json:"position" gorm:"column:position;type:varchar(64);not null;index:idx_tenant_position_platform;comment:位置标识：home/popup/activity/member_center 等"`
	Platform    string     `json:"platform" gorm:"column:platform;type:varchar(32);not null;default:all;index:idx_tenant_position_platform;comment:投放平台：web/app/h5/pc/all"`
	BannerType  string     `json:"bannerType" gorm:"column:banner_type;type:varchar(32);not null;default:image;comment:Banner类型：image=图片 video=视频 popup=弹窗"`
	JumpType    string     `json:"jumpType" gorm:"column:jump_type;type:varchar(32);not null;default:none;index:idx_tenant_jump_type;comment:跳转类型：none/url/internal/product/activity/category"`
	DisplayType string     `json:"displayType" gorm:"column:display_type;type:varchar(32);not null;default:banner;index:idx_tenant_display_type;comment:展示类型：banner=轮播 popup=弹窗 float=悬浮"`
	OpenMode    string     `json:"openMode" gorm:"column:open_mode;type:varchar(32);not null;default:current;comment:打开方式：current=当前页 new_window=新窗口 mini_app=小程序"`
	Sort        int        `json:"sort" gorm:"column:sort;type:int;not null;default:0;index:idx_tenant_status_deleted_sort;comment:排序值，越小越靠前"`
	Status      int8       `json:"status" gorm:"column:status;type:tinyint;not null;default:1;index:idx_tenant_status_deleted_sort;comment:状态：1启用 0禁用"`
	IsDeleted   int8       `json:"isDeleted" gorm:"column:is_deleted;type:tinyint;not null;default:0;index:idx_tenant_status_deleted_sort;comment:是否删除：0否 1是"`
	StartTime   *time.Time `json:"startTime" gorm:"column:start_time;type:datetime;index:idx_tenant_time_range;comment:生效开始时间"`
	EndTime     *time.Time `json:"endTime" gorm:"column:end_time;type:datetime;index:idx_tenant_time_range;comment:生效结束时间"`
	ShowCount   int64      `json:"showCount" gorm:"column:show_count;type:bigint unsigned;not null;default:0;comment:展示次数"`
	ClickCount  int64      `json:"clickCount" gorm:"column:click_count;type:bigint unsigned;not null;default:0;comment:点击次数"`
	Version     int        `json:"version" gorm:"column:version;type:int;not null;default:1;comment:版本号，便于并发控制"`
	Remark      *string    `json:"remark" gorm:"column:remark;type:varchar(255);comment:备注"`
	CreatedBy   *int64     `json:"createdBy" gorm:"column:created_by;type:bigint unsigned;comment:创建人"`
	UpdatedBy   *int64     `json:"updatedBy" gorm:"column:updated_by;type:bigint unsigned;comment:更新人"`
}

type SysBannerSearch struct {
	PageInfo
	TenantId     int64  `json:"tenantId"`
	BannerName   string `json:"bannerName"`
	BannerCode   string `json:"bannerCode"`
	Position     string `json:"position"`
	Platform     string `json:"platform"`
	BannerType   string `json:"bannerType"`
	JumpType     string `json:"jumpType"`
	DisplayType  string `json:"displayType"`
	LanguageCode string `json:"languageCode"`
	CountryCode  string `json:"countryCode"`
	Status       *int8  `json:"status"`
	StartTime    *int64 `json:"startTime"`
	EndTime      *int64 `json:"endTime"`
}

type SysBannerSet struct {
	ID          int64                    `json:"id"`
	TenantId    int64                    `json:"tenantId"`
	BannerName  string                   `json:"bannerName"`
	BannerCode  *string                  `json:"bannerCode"`
	Position    string                   `json:"position"`
	Platform    string                   `json:"platform"`
	BannerType  string                   `json:"bannerType"`
	JumpType    string                   `json:"jumpType"`
	DisplayType string                   `json:"displayType"`
	OpenMode    string                   `json:"openMode"`
	Sort        int                      `json:"sort"`
	Status      int8                     `json:"status"`
	StartTime   *int64                   `json:"startTime"`
	EndTime     *int64                   `json:"endTime"`
	Version     int                      `json:"version"`
	Remark      *string                  `json:"remark"`
	I18nList    []SysBannerI18nSet       `json:"i18nList"`
	CountryList []SysBannerCountryRelSet `json:"countryList"`
}

type SysBannerBack struct {
	ID          int64                     `json:"id"`
	CreatedAt   time.Time                 `json:"createdAt"`
	UpdatedAt   time.Time                 `json:"updatedAt"`
	TenantId    int64                     `json:"tenantId"`
	BannerName  string                    `json:"bannerName"`
	BannerCode  *string                   `json:"bannerCode"`
	Position    string                    `json:"position"`
	Platform    string                    `json:"platform"`
	BannerType  string                    `json:"bannerType"`
	JumpType    string                    `json:"jumpType"`
	DisplayType string                    `json:"displayType"`
	OpenMode    string                    `json:"openMode"`
	Sort        int                       `json:"sort"`
	Status      int8                      `json:"status"`
	IsDeleted   int8                      `json:"isDeleted"`
	StartTime   *time.Time                `json:"startTime"`
	EndTime     *time.Time                `json:"endTime"`
	ClickCount  int64                     `json:"clickCount"`
	ShowCount   int64                     `json:"showCount"`
	Version     int                       `json:"version"`
	Remark      *string                   `json:"remark"`
	CreatedBy   *int64                    `json:"createdBy"`
	UpdatedBy   *int64                    `json:"updatedBy"`
	I18nList    []SysBannerI18nBack       `json:"i18nList"`
	CountryList []SysBannerCountryRelBack `json:"countryList"`
}

type SysBannerResp struct {
	BasePageResponse[SysBannerBack]
}

// SysBannerI18n 租户Banner多语言表
type SysBannerI18n struct {
	BaseModel
	TenantId     int64   `json:"tenantId" gorm:"column:tenant_id;type:bigint unsigned;not null;index:idx_tenant_banner;index:idx_tenant_lang_country_status;index:idx_tenant_deleted;comment:租户ID"`
	BannerId     int64   `json:"bannerId" gorm:"column:banner_id;type:bigint unsigned;not null;uniqueIndex:uk_banner_lang_country;index:idx_tenant_banner;comment:Banner主表ID"`
	LanguageCode string  `json:"languageCode" gorm:"column:language_code;type:varchar(16);not null;uniqueIndex:uk_banner_lang_country;index:idx_tenant_lang_country_status;comment:语言代码：zh-CN/en-US/th-TH/vi-VN"`
	CountryCode  *string `json:"countryCode" gorm:"column:country_code;type:varchar(16);uniqueIndex:uk_banner_lang_country;index:idx_tenant_lang_country_status;comment:国家代码：CN/TH/VN/ID，NULL表示该语言通用"`
	Title        *string `json:"title" gorm:"column:title;type:varchar(255);comment:主标题"`
	SubTitle     *string `json:"subTitle" gorm:"column:sub_title;type:varchar(255);comment:副标题"`
	Description  *string `json:"description" gorm:"column:description;type:text;comment:描述文案"`
	ButtonText   *string `json:"buttonText" gorm:"column:button_text;type:varchar(64);comment:按钮文案"`
	ImageURL     string  `json:"imageUrl" gorm:"column:image_url;type:varchar(512);not null;comment:主图URL"`
	ThumbURL     *string `json:"thumbUrl" gorm:"column:thumb_url;type:varchar(512);comment:缩略图URL"`
	BgImageURL   *string `json:"bgImageUrl" gorm:"column:bg_image_url;type:varchar(512);comment:背景图URL"`
	IconURL      *string `json:"iconUrl" gorm:"column:icon_url;type:varchar(512);comment:图标URL"`
	VideoURL     *string `json:"videoUrl" gorm:"column:video_url;type:varchar(512);comment:视频URL"`
	JumpValue    *string `json:"jumpValue" gorm:"column:jump_value;type:varchar(512);comment:跳转值：URL/路由/业务ID"`
	TextColor    *string `json:"textColor" gorm:"column:text_color;type:varchar(32);comment:文字颜色"`
	ButtonColor  *string `json:"buttonColor" gorm:"column:button_color;type:varchar(32);comment:按钮颜色"`
	BgColor      *string `json:"bgColor" gorm:"column:bg_color;type:varchar(32);comment:背景色"`
	Status       int8    `json:"status" gorm:"column:status;type:tinyint;not null;default:1;index:idx_tenant_lang_country_status;comment:状态：1启用 0禁用"`
	IsDeleted    int8    `json:"isDeleted" gorm:"column:is_deleted;type:tinyint;not null;default:0;index:idx_tenant_deleted;comment:是否删除：0否 1是"`
	Remark       *string `json:"remark" gorm:"column:remark;type:varchar(255);comment:备注"`
	CreatedBy    *int64  `json:"createdBy" gorm:"column:created_by;type:bigint unsigned;comment:创建人"`
	UpdatedBy    *int64  `json:"updatedBy" gorm:"column:updated_by;type:bigint unsigned;comment:更新人"`
}

type SysBannerI18nSet struct {
	ID           int64   `json:"id"`
	TenantId     int64   `json:"tenantId"`
	BannerId     int64   `json:"bannerId"`
	LanguageCode string  `json:"languageCode"`
	CountryCode  *string `json:"countryCode"`
	Title        *string `json:"title"`
	SubTitle     *string `json:"subTitle"`
	Description  *string `json:"description"`
	ButtonText   *string `json:"buttonText"`
	ImageURL     string  `json:"imageUrl"`
	ThumbURL     *string `json:"thumbUrl"`
	BgImageURL   *string `json:"bgImageUrl"`
	IconURL      *string `json:"iconUrl"`
	VideoURL     *string `json:"videoUrl"`
	JumpValue    *string `json:"jumpValue"`
	TextColor    *string `json:"textColor"`
	ButtonColor  *string `json:"buttonColor"`
	BgColor      *string `json:"bgColor"`
	Status       int8    `json:"status"`
	Remark       *string `json:"remark"`
}

type SysBannerI18nBack struct {
	ID           int64     `json:"id"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
	TenantId     int64     `json:"tenantId"`
	BannerId     int64     `json:"bannerId"`
	LanguageCode string    `json:"languageCode"`
	CountryCode  *string   `json:"countryCode"`
	Title        *string   `json:"title"`
	SubTitle     *string   `json:"subTitle"`
	Description  *string   `json:"description"`
	ButtonText   *string   `json:"buttonText"`
	ImageURL     string    `json:"imageUrl"`
	ThumbURL     *string   `json:"thumbUrl"`
	BgImageURL   *string   `json:"bgImageUrl"`
	IconURL      *string   `json:"iconUrl"`
	VideoURL     *string   `json:"videoUrl"`
	JumpValue    *string   `json:"jumpValue"`
	TextColor    *string   `json:"textColor"`
	ButtonColor  *string   `json:"buttonColor"`
	BgColor      *string   `json:"bgColor"`
	Status       int8      `json:"status"`
	IsDeleted    int8      `json:"isDeleted"`
	Remark       *string   `json:"remark"`
	CreatedBy    *int64    `json:"createdBy"`
	UpdatedBy    *int64    `json:"updatedBy"`
}

var SysBannerI18nTableName = "sys_tenant_banner_i18n"

func (SysBannerI18n) TableName() string {
	return SysBannerI18nTableName
}

// SysBannerCountryRel 租户Banner国家投放关系表
type SysBannerCountryRel struct {
	BaseModel
	TenantId    int64   `json:"tenantId" gorm:"column:tenant_id;type:bigint unsigned;not null;index:idx_tenant_banner;index:idx_tenant_country_status;comment:租户ID"`
	BannerId    int64   `json:"bannerId" gorm:"column:banner_id;type:bigint unsigned;not null;uniqueIndex:uk_banner_country;index:idx_tenant_banner;comment:Banner主表ID"`
	CountryCode string  `json:"countryCode" gorm:"column:country_code;type:varchar(16);not null;uniqueIndex:uk_banner_country;index:idx_tenant_country_status;comment:国家代码：CN/TH/VN/ID/US"`
	Status      int8    `json:"status" gorm:"column:status;type:tinyint;not null;default:1;index:idx_tenant_country_status;comment:状态：1启用 0禁用"`
	Remark      *string `json:"remark" gorm:"column:remark;type:varchar(255);comment:备注"`
	CreatedBy   *int64  `json:"createdBy" gorm:"column:created_by;type:bigint unsigned;comment:创建人"`
	UpdatedBy   *int64  `json:"updatedBy" gorm:"column:updated_by;type:bigint unsigned;comment:更新人"`
}

type SysBannerCountryRelSet struct {
	ID          int64   `json:"id"`
	TenantId    int64   `json:"tenantId"`
	BannerId    int64   `json:"bannerId"`
	CountryCode string  `json:"countryCode"`
	Status      int8    `json:"status"`
	Remark      *string `json:"remark"`
}

type SysBannerCountryRelBack struct {
	ID          int64     `json:"id"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	TenantId    int64     `json:"tenantId"`
	BannerId    int64     `json:"bannerId"`
	CountryCode string    `json:"countryCode"`
	Status      int8      `json:"status"`
	Remark      *string   `json:"remark"`
	CreatedBy   *int64    `json:"createdBy"`
	UpdatedBy   *int64    `json:"updatedBy"`
}

var SysBannerCountryRelTableName = "sys_tenant_banner_country_rel"

func (SysBannerCountryRel) TableName() string {
	return SysBannerCountryRelTableName
}

// SysBannerAppReq App端获取轮播图请求
type SysBannerAppReq struct {
	Platform           string `json:"platform"`               // web/app/h5/pc，为空则返回 all + 该平台
	Position           string `json:"position"`               // home/popup/activity/member_center
	Lang               string `json:"lang"`                   // 当前浏览器语言，如 zh-CN/en-US；未匹配时回退 all
	LegacyLanguageCode string `json:"languageCode,omitempty"` // 兼容旧参数
	CountryCode        string `json:"countryCode"`            // 如 CN/TH/VN/ID
}

type SysBannerAppItem struct {
	ID           int64      `json:"id"`
	TenantId     int64      `json:"tenantId"`
	BannerName   string     `json:"bannerName"`
	BannerCode   *string    `json:"bannerCode"`
	Position     string     `json:"position"`
	Platform     string     `json:"platform"`
	BannerType   string     `json:"bannerType"`
	JumpType     string     `json:"jumpType"`
	DisplayType  string     `json:"displayType"`
	OpenMode     string     `json:"openMode"`
	Sort         int        `json:"sort"`
	Status       int8       `json:"status"`
	StartTime    *time.Time `json:"startTime"`
	EndTime      *time.Time `json:"endTime"`
	LanguageCode string     `json:"languageCode"`
	CountryCode  *string    `json:"countryCode"`
	Title        *string    `json:"title"`
	SubTitle     *string    `json:"subTitle"`
	Description  *string    `json:"description"`
	ButtonText   *string    `json:"buttonText"`
	ImageURL     string     `json:"imageUrl"`
	ThumbURL     *string    `json:"thumbUrl"`
	BgImageURL   *string    `json:"bgImageUrl"`
	IconURL      *string    `json:"iconUrl"`
	VideoURL     *string    `json:"videoUrl"`
	JumpValue    *string    `json:"jumpValue"`
	TextColor    *string    `json:"textColor"`
	ButtonColor  *string    `json:"buttonColor"`
	BgColor      *string    `json:"bgColor"`
}

// SysBannerGroupedResp App端轮播图按position分组返回
type SysBannerGroupedResp map[string][]SysBannerAppItem

var SysBannerTableName = "sys_tenant_banner"

func (SysBanner) TableName() string {
	return SysBannerTableName
}
