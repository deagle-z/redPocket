package pojo

type SysVipLevel struct {
	BaseModel
	TenantID            int64    `json:"tenantId" gorm:"column:tenant_id;not null;comment:租户ID"`
	Level               int      `json:"level" gorm:"column:level;not null;comment:等级排序（1=VIP0, 2=VIP1...）"`
	LevelName           string   `json:"levelName" gorm:"column:level_name;type:varchar(32);not null;comment:等级名称（VIP0/VIP1）"`
	AgentTag            *string  `json:"agentTag" gorm:"column:agent_tag;type:varchar(64);comment:代理标签（如：平台默认）"`
	TotalRechargeCount  *int     `json:"totalRechargeCount" gorm:"column:total_recharge_count;comment:总充值次数"`
	TotalRechargeAmount *float64 `json:"totalRechargeAmount" gorm:"column:total_recharge_amount;type:decimal(18,2);comment:总充值金额"`
	TotalValidBet       *float64 `json:"totalValidBet" gorm:"column:total_valid_bet;type:decimal(18,2);comment:总有效投注"`
	MonthRechargeAmount *float64 `json:"monthRechargeAmount" gorm:"column:month_recharge_amount;type:decimal(18,2);comment:当月充值金额"`
	MonthValidBet       *float64 `json:"monthValidBet" gorm:"column:month_valid_bet;type:decimal(18,2);comment:当月有效投注"`
	UpgradeBonusAmount  float64  `json:"upgradeBonusAmount" gorm:"column:upgrade_bonus_amount;type:decimal(18,2);not null;default:0;comment:升级赠送金额"`
	UpgradeType         *int8    `json:"upgradeType" gorm:"column:upgrade_type;type:tinyint;default:1;comment:升级方式：1累计 2当月"`
	KeepLevelCondition  *int8    `json:"keepLevelCondition" gorm:"column:keep_level_condition;type:tinyint;default:0;comment:是否需要保级（0否 1是）"`
	Sort                int      `json:"sort" gorm:"column:sort;not null;default:0;comment:排序"`
	Status              int8     `json:"status" gorm:"column:status;type:tinyint;not null;default:1;comment:状态：1启用 0禁用"`
	Remark              *string  `json:"remark" gorm:"column:remark;type:varchar(255);comment:备注"`
}

func (SysVipLevel) TableName() string {
	return "sys_vip_level"
}

type SysVipLevelSearch struct {
	PageInfo
	TenantID int64 `json:"tenantId"`
	Status   *int8 `json:"status"`
}

type SysVipLevelResp struct {
	BasePageResponse[SysVipLevelBack]
}

type SysVipLevelSet struct {
	ID                  int64    `json:"id"`
	TenantID            int64    `json:"tenantId"`
	Level               int      `json:"level"`
	LevelName           string   `json:"levelName"`
	AgentTag            *string  `json:"agentTag"`
	TotalRechargeCount  *int     `json:"totalRechargeCount"`
	TotalRechargeAmount *float64 `json:"totalRechargeAmount"`
	TotalValidBet       *float64 `json:"totalValidBet"`
	MonthRechargeAmount *float64 `json:"monthRechargeAmount"`
	MonthValidBet       *float64 `json:"monthValidBet"`
	UpgradeBonusAmount  float64  `json:"upgradeBonusAmount"`
	UpgradeType         *int8    `json:"upgradeType"`
	KeepLevelCondition  *int8    `json:"keepLevelCondition"`
	Sort                int      `json:"sort"`
	Status              int8     `json:"status"`
	Remark              *string  `json:"remark"`
}

// AppVipProgressBack app端VIP进度响应
type AppVipProgressBack struct {
	CurrentLevel    *AppVipLevelSimple  `json:"currentLevel"`    // 当前等级，nil表示未达任何等级
	PrevLevel       *AppVipLevelSimple  `json:"prevLevel"`       // 上一等级，nil表示无
	NextLevel       *AppVipLevelSimple  `json:"nextLevel"`       // 下一等级，nil表示已是最高
	Levels          []AppVipLevelSimple `json:"levels"`          // 全部启用等级
	Progress        float64             `json:"progress"`        // 朝下一等级的进度 0-100
	CurrentValue    float64             `json:"currentValue"`    // 当前指标值（累计/当月充值额）
	TargetValue     float64             `json:"targetValue"`     // 下一等级要求值，0表示无下一等级
	NextBonusAmount float64             `json:"nextBonusAmount"` // 下一等级升级奖励金额
}

type AppVipLevelSimple struct {
	Level              int     `json:"level"`
	LevelName          string  `json:"levelName"`
	UpgradeBonusAmount float64 `json:"upgradeBonusAmount"`
}

type SysVipLevelBack struct {
	ID                  int64    `json:"id"`
	TenantID            int64    `json:"tenantId"`
	Level               int      `json:"level"`
	LevelName           string   `json:"levelName"`
	AgentTag            *string  `json:"agentTag"`
	TotalRechargeCount  *int     `json:"totalRechargeCount"`
	TotalRechargeAmount *float64 `json:"totalRechargeAmount"`
	TotalValidBet       *float64 `json:"totalValidBet"`
	MonthRechargeAmount *float64 `json:"monthRechargeAmount"`
	MonthValidBet       *float64 `json:"monthValidBet"`
	UpgradeBonusAmount  float64  `json:"upgradeBonusAmount"`
	UpgradeType         *int8    `json:"upgradeType"`
	KeepLevelCondition  *int8    `json:"keepLevelCondition"`
	Sort                int      `json:"sort"`
	Status              int8     `json:"status"`
	Remark              *string  `json:"remark"`
	CreatedAt           string   `json:"createdAt"`
	UpdatedAt           string   `json:"updatedAt"`
}
