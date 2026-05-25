package pojo

import "time"

type AppGame struct {
	GameID            int64      `json:"gameId" gorm:"column:game_id;type:bigint;primaryKey;autoIncrement"`
	GameName          *string    `json:"gameName" gorm:"column:game_name;type:varchar(255)"`
	CategoryCode      *string    `json:"categoryCode" gorm:"column:category_code;type:varchar(255)"`
	ShowIndex         *int       `json:"showIndex" gorm:"column:show_index;type:int"`
	Hot               *int       `json:"hot" gorm:"column:hot;type:int"`
	HomeShow          *int       `json:"homeShow" gorm:"column:home_show;type:int;default:0;comment:是否首页展示 0否 1是"`
	Type              *int       `json:"type" gorm:"column:type;type:int"`
	ParentID          *int64     `json:"parentId" gorm:"column:parent_id;type:bigint"`
	PlatformCode      *string    `json:"platformCode" gorm:"column:platform_code;type:varchar(255);index"`
	ThirdGameID       *string    `json:"thirdGameId" gorm:"column:third_game_id;type:varchar(255);index"`
	ThirdGameName     *string    `json:"thirdGameName" gorm:"column:third_game_name;type:varchar(255)"`
	ThirdGameCategory *string    `json:"thirdGameCategory" gorm:"column:third_game_category;type:varchar(255)"`
	HorizontalImage   *string    `json:"horizontalImage" gorm:"column:horizontal_image;type:varchar(255)"`
	GameIcon          *string    `json:"gameIcon" gorm:"column:game_icon;type:varchar(255)"`
	TypeIcon          *string    `json:"typeIcon" gorm:"column:type_icon;type:varchar(255)"`
	TypeActiveIcon    *string    `json:"typeActiveIcon" gorm:"column:type_active_icon;type:varchar(255)"`
	TenantID          *int64     `json:"tenantId" gorm:"column:tenant_id;type:bigint;index"`
	Sort              *int       `json:"sort" gorm:"column:sort;type:int"`
	DisabledFlag      *int       `json:"disabledFlag" gorm:"column:disabled_flag;type:int"`
	DeletedFlag       *int       `json:"deletedFlag" gorm:"column:deleted_flag;type:int"`
	Remark            *string    `json:"remark" gorm:"column:remark;type:varchar(255)"`
	UpdateTime        *time.Time `json:"updateTime" gorm:"column:update_time"`
	CreateTime        *time.Time `json:"createTime" gorm:"column:create_time"`
}

type AppGameSearch struct {
	PageInfo
	GameID            int64  `json:"gameId"`
	GameName          string `json:"gameName"`
	CategoryCode      string `json:"categoryCode"`
	Type              *int   `json:"type"`
	ParentID          *int64 `json:"parentId"`
	PlatformCode      string `json:"platformCode"`
	ThirdGameID       string `json:"thirdGameId"`
	ThirdGameName     string `json:"thirdGameName"`
	ThirdGameCategory string `json:"thirdGameCategory"`
	TenantID          *int64 `json:"tenantId"`
	Hot               *int   `json:"hot"`
	HomeShow          *int   `json:"homeShow"`
	DisabledFlag      *int   `json:"disabledFlag"`
	DeletedFlag       *int   `json:"deletedFlag"`
}

type AppGameSet struct {
	GameID            int64   `json:"gameId"`
	GameName          *string `json:"gameName"`
	CategoryCode      *string `json:"categoryCode"`
	ShowIndex         *int    `json:"showIndex"`
	Hot               *int    `json:"hot"`
	HomeShow          *int    `json:"homeShow"`
	Type              *int    `json:"type"`
	ParentID          *int64  `json:"parentId"`
	PlatformCode      *string `json:"platformCode"`
	ThirdGameID       *string `json:"thirdGameId"`
	ThirdGameName     *string `json:"thirdGameName"`
	ThirdGameCategory *string `json:"thirdGameCategory"`
	HorizontalImage   *string `json:"horizontalImage"`
	GameIcon          *string `json:"gameIcon"`
	TypeIcon          *string `json:"typeIcon"`
	TypeActiveIcon    *string `json:"typeActiveIcon"`
	TenantID          *int64  `json:"tenantId"`
	Sort              *int    `json:"sort"`
	DisabledFlag      *int    `json:"disabledFlag"`
	DeletedFlag       *int    `json:"deletedFlag"`
	Remark            *string `json:"remark"`
}

type AppGameResp struct {
	BasePageResponse[AppGame]
}

type AppGameHomeItem struct {
	GameID          int64  `json:"gameId"`
	GameName        string `json:"gameName"`
	CategoryCode    string `json:"categoryCode"`
	Type            *int   `json:"type"`
	Manufacturer    string `json:"manufacturer"`
	GameIcon        string `json:"gameIcon"`
	HorizontalImage string `json:"horizontalImage"`
	Sort            int    `json:"sort"`
	ShowIndex       int    `json:"showIndex"`
}

type AppGameHomeResp map[string][]AppGameHomeItem

type AppGameSyncReq struct {
	Language     string `json:"language"`
	PlatformCode string `json:"platformCode"`
}

type AppGameLaunchReq struct {
	GameID   int64  `json:"gameId" binding:"required"`
	Language string `json:"language"`
}

type AppGameLaunchResp struct {
	URL string `json:"url"`
}

type AppGameSyncResp struct {
	Total   int `json:"total"`
	Created int `json:"created"`
	Updated int `json:"updated"`
	Skipped int `json:"skipped"`
}

var AppGameTableName = "app_game"

func (AppGame) TableName() string {
	return AppGameTableName
}
