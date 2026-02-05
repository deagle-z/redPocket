package pojo

import "time"

type SysTenantUser struct { // 租户用户表（独立账号密码）
	BaseModel
	TenantId       int64      `json:"tenantId" gorm:"column:tenant_id;type:bigint;not null"`               // 租户ID
	Username       string     `json:"username" gorm:"column:username;type:varchar(64);not null"`           // 登录账号（租户内唯一）
	PasswordHash   string     `json:"passwordHash" gorm:"column:password_hash;type:varchar(255);not null"` // 密码哈希（bcrypt/argon2）
	PasswordAlgo   string     `json:"passwordAlgo" gorm:"column:password_algo;type:varchar(32);default:bcrypt"`
	Email          *string    `json:"email" gorm:"column:email;type:varchar(128)"`                // 邮箱（可选）
	Mobile         *string    `json:"mobile" gorm:"column:mobile;type:varchar(32)"`               // 手机号（可选）
	RoleCode       string     `json:"roleCode" gorm:"column:role_code;type:varchar(64);not null"` // 角色编码（owner/admin/member）
	IsOwner        bool       `json:"isOwner" gorm:"column:is_owner;type:tinyint(1);default:0"`   // 是否租户所有者
	Status         int        `json:"status" gorm:"column:status;type:tinyint;default:1"`         // 状态：0=禁用 1=启用 2=冻结
	LastLoginAt    *time.Time `json:"lastLoginAt" gorm:"column:last_login_at;type:datetime(3)"`   // 最后登录时间
	LastLoginIp    *string    `json:"lastLoginIp" gorm:"column:last_login_ip;type:varchar(64)"`   // 最后登录IP
	LastLoginUa    *string    `json:"lastLoginUa" gorm:"column:last_login_ua;type:varchar(255)"`  // 最后登录UA
	LoginFailCount int        `json:"loginFailCount" gorm:"column:login_fail_count;type:int;default:0"`
	LockedUntil    *time.Time `json:"lockedUntil" gorm:"column:locked_until;type:datetime(3)"` // 锁定截止时间
	Require2fa     bool       `json:"require2fa" gorm:"column:require_2fa;type:tinyint(1);default:0"`
	TwofaSecret    *string    `json:"twofaSecret" gorm:"column:twofa_secret;type:varchar(128)"` // 2FA密钥（TOTP）
	Remark         *string    `json:"remark" gorm:"column:remark;type:varchar(255)"`            // 备注
}

type SysTenantUserSearch struct {
	PageInfo
	TenantId   int64  `json:"tenantId"`
	Username   string `json:"username"`
	Email      string `json:"email"`
	Mobile     string `json:"mobile"`
	RoleCode   string `json:"roleCode"`
	IsOwner    *bool  `json:"isOwner"`
	Status     *int   `json:"status"`
	Require2fa *bool  `json:"require2fa"`
}

type SysTenantUserSet struct {
	ID           int64   `json:"id"` // ID
	TenantId     int64   `json:"tenantId"`
	Username     string  `json:"username"`
	PasswordHash string  `json:"passwordHash"`
	PasswordAlgo string  `json:"passwordAlgo"`
	Email        *string `json:"email"`
	Mobile       *string `json:"mobile"`
	RoleCode     string  `json:"roleCode"`
	IsOwner      bool    `json:"isOwner"`
	Status       int     `json:"status"`
	Require2fa   bool    `json:"require2fa"`
	TwofaSecret  *string `json:"twofaSecret"`
	Remark       *string `json:"remark"`
}

type SysTenantUserBack struct {
	ID             int64      `json:"id"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`
	TenantId       int64      `json:"tenantId"`
	Username       string     `json:"username"`
	PasswordHash   string     `json:"passwordHash"`
	PasswordAlgo   string     `json:"passwordAlgo"`
	Email          *string    `json:"email"`
	Mobile         *string    `json:"mobile"`
	RoleCode       string     `json:"roleCode"`
	IsOwner        bool       `json:"isOwner"`
	Status         int        `json:"status"`
	LastLoginAt    *time.Time `json:"lastLoginAt"`
	LastLoginIp    *string    `json:"lastLoginIp"`
	LastLoginUa    *string    `json:"lastLoginUa"`
	LoginFailCount int        `json:"loginFailCount"`
	LockedUntil    *time.Time `json:"lockedUntil"`
	Require2fa     bool       `json:"require2fa"`
	TwofaSecret    *string    `json:"twofaSecret"`
	Remark         *string    `json:"remark"`
}

type SysTenantUserResp struct {
	BasePageResponse[SysTenantUserBack]
}

var SysTenantUserTableName = "sys_tenant_user"

func (SysTenantUser) TableName() string {
	return SysTenantUserTableName
}
