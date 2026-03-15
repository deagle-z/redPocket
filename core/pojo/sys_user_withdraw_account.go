package pojo

import "time"

type SysUserWithdrawAccount struct {
	BaseModel
	TenantID    int64           `json:"tenantId" gorm:"column:tenant_id;type:bigint unsigned;not null;index:idx_user;comment:租户ID"`
	UserID      int64           `json:"userId" gorm:"column:user_id;type:bigint unsigned;not null;index:idx_user;comment:用户ID"`
	CountryCode string          `json:"countryCode" gorm:"column:country_code;type:varchar(8);not null;comment:国家code"`
	AccountData string          `json:"accountData" gorm:"column:account_data;type:json;not null;comment:账户信息JSON"`
	IsDefault   int8            `json:"isDefault" gorm:"column:is_default;type:tinyint;not null;default:0;comment:是否默认账户"`
	Status      int8            `json:"status" gorm:"column:status;type:tinyint;not null;default:1;comment:状态 1正常 0禁用"`
	Remark      *string         `json:"remark" gorm:"column:remark;type:varchar(255);comment:备注"`
}

type SysUserWithdrawAccountSearch struct {
	PageInfo
	UserID      int64  `json:"userId"`
	CountryCode string `json:"countryCode"`
	IsDefault   *int8  `json:"isDefault"`
	Status      *int8  `json:"status"`
}

type SysUserWithdrawAccountSet struct {
	ID          int64   `json:"id"`
	UserID      int64   `json:"userId"`   // 管理员创建时指定用户
	CountryCode string  `json:"countryCode"`
	AccountData string  `json:"accountData"`
	IsDefault   int8    `json:"isDefault"`
	Status      int8    `json:"status"`
	Remark      *string `json:"remark"`
}

type SysUserWithdrawAccountBack struct {
	ID          int64     `json:"id"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	TenantID    int64     `json:"tenantId"`
	UserID      int64     `json:"userId"`
	CountryCode string    `json:"countryCode"`
	AccountData string    `json:"accountData"`
	IsDefault   int8      `json:"isDefault"`
	Status      int8      `json:"status"`
	Remark      *string   `json:"remark"`
}

type SysUserWithdrawAccountResp struct {
	BasePageResponse[SysUserWithdrawAccountBack]
}

var SysUserWithdrawAccountTableName = "sys_user_withdraw_account"

func (SysUserWithdrawAccount) TableName() string {
	return SysUserWithdrawAccountTableName
}
