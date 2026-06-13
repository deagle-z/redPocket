package pojo

// TgUserInviteRewardLog 邀请返佣阶梯奖励发放记录（受益人=上级，按档位发放，每档每人仅一次）
type TgUserInviteRewardLog struct {
	BaseModel
	TenantID   int64   `json:"tenantId" gorm:"column:tenant_id;type:bigint;not null;default:0;index;comment:租户ID"`
	UserID     int64   `json:"userId" gorm:"column:user_id;type:bigint;not null;uniqueIndex:idx_invite_reward_user_tier,priority:1;comment:受益人(上级)ID"`
	TierLevel  int     `json:"tierLevel" gorm:"column:tier_level;type:int;not null;uniqueIndex:idx_invite_reward_user_tier,priority:2;comment:档位序号"`
	Threshold  int     `json:"threshold" gorm:"column:threshold;type:int;not null;default:0;comment:触发该档的有效用户数阈值"`
	ValidUsers int     `json:"validUsers" gorm:"column:valid_users;type:int;not null;default:0;comment:发放时有效用户数快照"`
	Amount     float64 `json:"amount" gorm:"column:amount;type:decimal(20,2);not null;default:0;comment:发放金额"`
}

func (TgUserInviteRewardLog) TableName() string {
	return "tg_user_invite_reward_log"
}
