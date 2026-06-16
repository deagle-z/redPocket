package pojo

const (
	InviteRewardStageLegacy   = "legacy"
	InviteRewardStageRecharge = "recharge"
	InviteRewardStageBet      = "bet"
)

// TgUserInviteRewardLog 邀请返佣发放记录（受益人=上级；同一下级按充值/投注达标阶段分别发放）
type TgUserInviteRewardLog struct {
	BaseModel
	TenantID   int64   `json:"tenantId" gorm:"column:tenant_id;type:bigint;not null;default:0;index;comment:租户ID"`
	UserID     int64   `json:"userId" gorm:"column:user_id;type:bigint;not null;uniqueIndex:idx_invite_reward_user_sub_stage,priority:1;comment:受益人(上级)ID"`
	SubUserID  int64   `json:"subUserId" gorm:"column:sub_user_id;type:bigint;not null;default:0;uniqueIndex:idx_invite_reward_user_sub_stage,priority:2;comment:触发的下级ID"`
	Stage      string  `json:"stage" gorm:"column:stage;type:varchar(32);not null;default:'legacy';uniqueIndex:idx_invite_reward_user_sub_stage,priority:3;comment:奖励阶段 legacy旧规则 recharge充值 bet投注达标"`
	ValidUsers int     `json:"validUsers" gorm:"column:valid_users;type:int;not null;default:0;comment:发放时上级对应阶段人数(档位)"`
	Amount     float64 `json:"amount" gorm:"column:amount;type:decimal(20,2);not null;default:0;comment:发放金额"`
}

func (TgUserInviteRewardLog) TableName() string {
	return "tg_user_invite_reward_log"
}
