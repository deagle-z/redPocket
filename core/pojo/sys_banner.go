package pojo

import "time"

// SysBanner 轮播图配置表
type SysBanner struct {
	BaseModel
	TenantId   int64      `json:"tenantId" gorm:"column:tenant_id;type:bigint unsigned;not null;index:idx_tenant_position;comment:租户ID"`
	BannerName string     `json:"bannerName" gorm:"column:banner_name;type:varchar(128);not null;comment:轮播图名称"`
	Position   string     `json:"position" gorm:"column:position;type:varchar(64);not null;index:idx_tenant_position;comment:位置标识：首页首页=home 首页弹窗=popup 活动页=activity"`
	Platform   string     `json:"platform" gorm:"column:platform;type:varchar(32);not null;default:all;comment:平台：web/app/h5/all"`
	ImageURL   string     `json:"imageUrl" gorm:"column:image_url;type:varchar(512);not null;comment:图片URL"`
	ThumbURL   *string    `json:"thumbUrl" gorm:"column:thumb_url;type:varchar(512);comment:缩略图"`
	JumpType   string     `json:"jumpType" gorm:"column:jump_type;type:varchar(32);not null;default:none;comment:跳转类型：none/url/internal/product/activity"`
	JumpValue  *string    `json:"jumpValue" gorm:"column:jump_value;type:varchar(512);comment:跳转值（URL/ID/路由）"`
	Sort       int        `json:"sort" gorm:"column:sort;type:int;not null;default:0;comment:排序（越小越靠前）"`
	Status     int8       `json:"status" gorm:"column:status;type:tinyint;not null;default:1;index:idx_status_time;comment:状态：1启用 0禁用"`
	StartTime  *time.Time `json:"startTime" gorm:"column:start_time;type:datetime;index:idx_status_time;comment:开始时间"`
	EndTime    *time.Time `json:"endTime" gorm:"column:end_time;type:datetime;index:idx_status_time;comment:结束时间"`
	ClickCount int64      `json:"clickCount" gorm:"column:click_count;type:bigint;not null;default:0;comment:点击次数"`
	ShowCount  int64      `json:"showCount" gorm:"column:show_count;type:bigint;not null;default:0;comment:展示次数"`
	Remark     *string    `json:"remark" gorm:"column:remark;type:varchar(255);comment:备注"`
}

type SysBannerSearch struct {
	PageInfo
	TenantId   int64  `json:"tenantId"`
	BannerName string `json:"bannerName"`
	Position   string `json:"position"`
	Platform   string `json:"platform"`
	JumpType   string `json:"jumpType"`
	Status     *int8  `json:"status"`
}

type SysBannerSet struct {
	ID         int64   `json:"id"`
	TenantId   int64   `json:"tenantId"`
	BannerName string  `json:"bannerName"`
	Position   string  `json:"position"`
	Platform   string  `json:"platform"`
	ImageURL   string  `json:"imageUrl"`
	ThumbURL   *string `json:"thumbUrl"`
	JumpType   string  `json:"jumpType"`
	JumpValue  *string `json:"jumpValue"`
	Sort       int     `json:"sort"`
	Status     int8    `json:"status"`
	StartTime  *int64  `json:"startTime"`
	EndTime    *int64  `json:"endTime"`
	Remark     *string `json:"remark"`
}

type SysBannerBack struct {
	ID         int64      `json:"id"`
	CreatedAt  time.Time  `json:"createdAt"`
	UpdatedAt  time.Time  `json:"updatedAt"`
	TenantId   int64      `json:"tenantId"`
	BannerName string     `json:"bannerName"`
	Position   string     `json:"position"`
	Platform   string     `json:"platform"`
	ImageURL   string     `json:"imageUrl"`
	ThumbURL   *string    `json:"thumbUrl"`
	JumpType   string     `json:"jumpType"`
	JumpValue  *string    `json:"jumpValue"`
	Sort       int        `json:"sort"`
	Status     int8       `json:"status"`
	StartTime  *time.Time `json:"startTime"`
	EndTime    *time.Time `json:"endTime"`
	ClickCount int64      `json:"clickCount"`
	ShowCount  int64      `json:"showCount"`
	Remark     *string    `json:"remark"`
}

type SysBannerResp struct {
	BasePageResponse[SysBannerBack]
}

// SysBannerAppReq App端获取轮播图请求
type SysBannerAppReq struct {
	Platform string `json:"platform"` // web/app/h5，为空则返回all+该平台
}

// SysBannerGroupedResp App端轮播图按position分组返回
type SysBannerGroupedResp map[string][]SysBannerBack

var SysBannerTableName = "sys_banner"

func (SysBanner) TableName() string {
	return SysBannerTableName
}
