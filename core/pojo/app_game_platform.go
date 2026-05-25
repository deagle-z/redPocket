package pojo

import "time"

type AppGamePlatform struct {
	PlatformCode string     `json:"platformCode" gorm:"column:platform_code;type:varchar(255);primaryKey"`
	AppID        *string    `json:"appId" gorm:"column:app_id;type:varchar(255)"`
	Currency     *string    `json:"currency" gorm:"column:currency;type:varchar(255)"`
	Secret       *string    `json:"secret" gorm:"column:secret;type:varchar(255)"`
	HandlingFee  *float64   `json:"handlingFee" gorm:"column:handling_fee;type:decimal(20,2)"`
	Sort         *int       `json:"sort" gorm:"column:sort;type:int"`
	DisabledFlag *int       `json:"disabledFlag" gorm:"column:disabled_flag;type:int"`
	DeletedFlag  *int       `json:"deletedFlag" gorm:"column:deleted_flag;type:int"`
	Remark       *string    `json:"remark" gorm:"column:remark;type:varchar(255)"`
	UpdateTime   *time.Time `json:"updateTime" gorm:"column:update_time"`
	CreateTime   *time.Time `json:"createTime" gorm:"column:create_time"`
}

type AppGamePlatformSearch struct {
	PageInfo
	PlatformCode string `json:"platformCode"`
	AppID        string `json:"appId"`
	Currency     string `json:"currency"`
	DisabledFlag *int   `json:"disabledFlag"`
	DeletedFlag  *int   `json:"deletedFlag"`
}

type AppGamePlatformSet struct {
	PlatformCode string   `json:"platformCode"`
	AppID        *string  `json:"appId"`
	Currency     *string  `json:"currency"`
	Secret       *string  `json:"secret"`
	HandlingFee  *float64 `json:"handlingFee"`
	Sort         *int     `json:"sort"`
	DisabledFlag *int     `json:"disabledFlag"`
	DeletedFlag  *int     `json:"deletedFlag"`
	Remark       *string  `json:"remark"`
}

type AppGamePlatformResp struct {
	BasePageResponse[AppGamePlatform]
}

var AppGamePlatformTableName = "app_game_platform"

func (AppGamePlatform) TableName() string {
	return AppGamePlatformTableName
}
