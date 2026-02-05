package pojo

import (
	"time"
)

type PkManager struct {
	BaseModel
	ApkPackage string `yaml:"apkPackage" json:"apkPackage" gorm:"type:varchar(255);not null;"` // APK 包名
	Url        string `yaml:"url" json:"url" gorm:"type:varchar(2048);not null;"`              // WebView 展示的 URL
	Name       string `yaml:"name" json:"name" gorm:"type:varchar(2048);not null;"`            // 名称
	IsActive   int8   `yaml:"isActive" json:"isActive" gorm:"type:tinyint(1);default:1;"`      // 是否启用
}

type PkManagerSearch struct {
	PageInfo
	ApkPackage string `json:"apkPackage"` // APK 包名
	Name       string `json:"name"`       // APK 包名
	IsActive   *int8  `json:"isActive"`   // 是否启用
}

type PkManagerSet struct {
	ID         int64  `json:"id"`                                     // ID
	ApkPackage string `json:"apkPackage" binding:"required,max=255"`  // APK 包名，必填，最长255个字符
	Name       string `json:"name" binding:"required,max=255"`        // APK 包名，必填，最长255个字符
	Url        string `json:"url" binding:"required,max=2048"`        // WebView 展示的 URL，必填，最长2048个字符
	IsActive   int8   `json:"isActive" binding:"omitempty,oneof=0 1"` // 是否启用，可选，值为0或1
}

type PkManagerBack struct {
	ID         int64     `json:"id"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
	ApkPackage string    `json:"apkPackage"` // APK 包名
	Name       string    `json:"name"`       // APK 包名
	Url        string    `json:"url"`        // WebView 展示的 URL
	IsActive   int8      `json:"isActive"`   // 是否启用
}

type PkManagerResp struct {
	BasePageResponse[PkManagerBack]
}

var PkManagerTableName = "pk_manager"

func (PkManager) TableName() string {
	return PkManagerTableName
}
