package pojo

import "time"

type SysTenant struct { // 租户主表
	BaseModel
	TenantCode  string     `json:"tenantCode" gorm:"column:tenant_code;type:varchar(64);uniqueIndex"` // 租户唯一编码（业务侧）
	TenantName  string     `json:"tenantName" gorm:"column:tenant_name;type:varchar(128)"`            // 租户名称
	TenantType  int        `json:"tenantType" gorm:"column:tenant_type;type:tinyint;default:1"`       // 租户类型：1=个人 2=企业 3=代理商 4=内部
	Status      int        `json:"status" gorm:"column:status;type:tinyint;default:1;index"`          // 状态：0=禁用 1=启用 2=冻结
	OwnerUserId *int64     `json:"ownerUserId" gorm:"column:owner_user_id;type:bigint;index"`         // 租户所有者用户ID
	PlanCode    *string    `json:"planCode" gorm:"column:plan_code;type:varchar(64)"`                 // 套餐标识（free/pro/enterprise）
	Timezone    string     `json:"timezone" gorm:"column:timezone;type:varchar(64);default:UTC"`      // 租户时区
	Locale      string     `json:"locale" gorm:"column:locale;type:varchar(32);default:en-US"`        // 默认语言
	Remark      *string    `json:"remark" gorm:"column:remark;type:varchar(255)"`                     // 备注
	DeletedAt   *time.Time `json:"deletedAt" gorm:"column:deleted_at;type:datetime(3)"`               // 软删除时间
}

type SysTenantSearch struct {
	PageInfo
	TenantCode  string `json:"tenantCode"`  // 租户唯一编码
	TenantName  string `json:"tenantName"`  // 租户名称
	TenantType  *int   `json:"tenantType"`  // 租户类型
	Status      *int   `json:"status"`      // 状态
	OwnerUserId *int64 `json:"ownerUserId"` // 租户所有者用户ID
	PlanCode    string `json:"planCode"`    // 套餐标识
}

type SysTenantSet struct {
	ID          int64   `json:"id"` // ID
	TenantCode  string  `json:"tenantCode"`
	TenantName  string  `json:"tenantName"`
	TenantType  int     `json:"tenantType"`
	Status      int     `json:"status"`
	OwnerUserId *int64  `json:"ownerUserId"`
	PlanCode    *string `json:"planCode"`
	Timezone    string  `json:"timezone"`
	Locale      string  `json:"locale"`
	Remark      *string `json:"remark"`
}

type SysTenantBack struct {
	ID          int64      `json:"id"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
	DeletedAt   *time.Time `json:"deletedAt"`
	TenantCode  string     `json:"tenantCode"`
	TenantName  string     `json:"tenantName"`
	TenantType  int        `json:"tenantType"`
	Status      int        `json:"status"`
	OwnerUserId *int64     `json:"ownerUserId"`
	PlanCode    *string    `json:"planCode"`
	Timezone    string     `json:"timezone"`
	Locale      string     `json:"locale"`
	Remark      *string    `json:"remark"`
}

type SysTenantResetPassword struct {
	TenantId int64  `json:"tenantId"`
	Password string `json:"password"`
}

type SysTenantResp struct {
	BasePageResponse[SysTenantBack]
}

var SysTenantTableName = "sys_tenant"

func (SysTenant) TableName() string {
	return SysTenantTableName
}
