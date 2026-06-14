package pojo

// TgUserInviteRewardLog 邀请返佣发放记录（受益人=上级；每个有效下级触发一次，按当时有效用户数所在档位的单价发放）
type TgUserInviteRewardLog struct {
	BaseModel
	TenantID   int64   `json:"tenantId" gorm:"column:tenant_id;type:bigint;not null;default:0;index;comment:租户ID"`
	UserID     int64   `json:"userId" gorm:"column:user_id;type:bigint;not null;uniqueIndex:idx_invite_reward_user_sub,priority:1;comment:受益人(上级)ID"`
	SubUserID  int64   `json:"subUserId" gorm:"column:sub_user_id;type:bigint;not null;default:0;uniqueIndex:idx_invite_reward_user_sub,priority:2;comment:触发的有效下级ID"`
	ValidUsers int     `json:"validUsers" gorm:"column:valid_users;type:int;not null;default:0;comment:发放时上级有效用户数(档位)"`
	Amount     float64 `json:"amount" gorm:"column:amount;type:decimal(20,2);not null;default:0;comment:发放金额"`
}

func (TgUserInviteRewardLog) TableName() string {
	return "tg_user_invite_reward_log"
}
