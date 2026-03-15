package pojo

import "time"

type SysPayMethod struct {
	BaseModel
	MethodCode string  `json:"methodCode" gorm:"column:method_code;type:varchar(64);uniqueIndex;not null;comment:支付方式编码，如 upi / pix / bank_card / usdt_trc20"`
	MethodName string  `json:"methodName" gorm:"column:method_name;type:varchar(128);not null;comment:支付方式名称"`
	Icon       *string `json:"icon" gorm:"column:icon;type:varchar(255);comment:支付方式图标URL"`
	Sort       int     `json:"sort" gorm:"column:sort;type:int;not null;default:0;comment:排序，越小越靠前"`
	Status     int8    `json:"status" gorm:"column:status;type:tinyint;not null;default:1;comment:状态 1启用 0禁用"`
	Remark     *string `json:"remark" gorm:"column:remark;type:varchar(255);comment:备注"`
	DeletedAt  int64   `json:"deletedAt" gorm:"column:deleted_at;type:bigint;not null;default:0;comment:软删除时间戳，0=未删除"`
}

type SysPayMethodSearch struct {
	PageInfo
	MethodCode string `json:"methodCode"`
	MethodName string `json:"methodName"`
	Status     *int8  `json:"status"`
}

type SysPayMethodSet struct {
	ID         int64   `json:"id"`
	MethodCode string  `json:"methodCode"`
	MethodName string  `json:"methodName"`
	Icon       *string `json:"icon"`
	Sort       int     `json:"sort"`
	Status     int8    `json:"status"`
	Remark     *string `json:"remark"`
}

type SysPayMethodBack struct {
	ID         int64     `json:"id"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
	MethodCode string    `json:"methodCode"`
	MethodName string    `json:"methodName"`
	Icon       *string   `json:"icon"`
	Sort       int       `json:"sort"`
	Status     int8      `json:"status"`
	Remark     *string   `json:"remark"`
}

type SysPayMethodResp struct {
	BasePageResponse[SysPayMethodBack]
}

var SysPayMethodTableName = "sys_pay_method"

func (SysPayMethod) TableName() string {
	return SysPayMethodTableName
}
