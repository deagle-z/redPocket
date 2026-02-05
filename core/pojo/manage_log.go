package pojo

import (
	"time"
)

type ManageLog struct { // 管理员日志
	BaseModel
	Username     string `yaml:"username" json:"username" gorm:"type:varchar(64);index;"` // 操作用户名
	RequestHost  string `yaml:"merchantId" json:"merchantId" gorm:"type:varchar(128);"`  // 请求地址
	RequestBody  string `yaml:"appKey" json:"appKey" gorm:"type:text;"`                  // 请求参数
	ResponseBody string `yaml:"secret" json:"secret" gorm:"type:text;"`                  // 响应参数
	Ip           string `yaml:"payUrl" json:"payUrl" gorm:"type:varchar(64);"`           // 请求ip
}

type ManageLogBack struct {
	ID           int64     `json:"id"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
	Username     string    `json:"username"`   // 操作用户名
	RequestHost  string    `json:"merchantId"` // 请求地址
	RequestBody  string    `json:"appKey"`     // 请求参数
	ResponseBody string    `json:"secret"`     // 响应参数
	Ip           string    `json:"payUrl"`     // 请求ip
}

type ManageLogResp struct {
	BasePageResponse[ManageLogBack]
}

type ManageLogSearch struct {
	PageInfo
	Username string `json:"username"` // 操作用户名
}

var ManageLogTableName = "manage_log"

func (ManageLog) TableName() string {
	return ManageLogTableName
}
