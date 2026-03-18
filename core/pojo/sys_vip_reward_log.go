package pojo

const (
	VipRewardTypeUpgrade  int8 = 1 // 升级奖励
	VipRewardTypeKeep     int8 = 2 // 保级奖励
	VipRewardTypeManual   int8 = 3 // 手动补发

	VipRewardStatusPending int8 = 0 // 待发放
	VipRewardStatusDone    int8 = 1 // 已发放
	VipRewardStatusFail    int8 = -1 // 失败
)

type SysVipRewardLog struct {
	BaseModel
	TenantID    int64   `json:"tenantId"    gorm:"column:tenant_id;not null;comment:租户ID"`
	UserID      int64   `json:"userId"      gorm:"column:user_id;not null;comment:用户ID"`
	VipLevel    int     `json:"vipLevel"    gorm:"column:vip_level;not null;comment:达到的VIP等级"`
	LevelName   string  `json:"levelName"   gorm:"column:level_name;size:32;not null;comment:等级名称"`
	RewardType  int8    `json:"rewardType"  gorm:"column:reward_type;type:tinyint;not null;default:1;comment:奖励类型：1升级 2保级 3手动补发"`
	BonusAmount float64 `json:"bonusAmount" gorm:"column:bonus_amount;type:decimal(18,2);not null;default:0;comment:发放金额"`
	Status      int8    `json:"status"      gorm:"column:status;type:tinyint;not null;default:1;comment:状态：1已发放 0待发放 -1失败"`
	Remark      *string `json:"remark"      gorm:"column:remark;size:255;comment:备注"`
}

func (SysVipRewardLog) TableName() string {
	return "sys_vip_reward_log"
}

type SysVipRewardLogSearch struct {
	PageInfo
	TenantID   int64  `json:"tenantId"`
	UserID     int64  `json:"userId"`
	VipLevel   *int   `json:"vipLevel"`
	RewardType *int8  `json:"rewardType"`
	Status     *int8  `json:"status"`
}

type SysVipRewardLogBack struct {
	ID          int64   `json:"id"`
	TenantID    int64   `json:"tenantId"`
	UserID      int64   `json:"userId"`
	VipLevel    int     `json:"vipLevel"`
	LevelName   string  `json:"levelName"`
	RewardType  int8    `json:"rewardType"`
	BonusAmount float64 `json:"bonusAmount"`
	Status      int8    `json:"status"`
	Remark      *string `json:"remark"`
	CreatedAt   string  `json:"createdAt"`
	UpdatedAt   string  `json:"updatedAt"`
}

type SysVipRewardLogResp struct {
	BasePageResponse[SysVipRewardLogBack]
}
