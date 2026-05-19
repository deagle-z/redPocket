package pojo

import (
	"time"
)

// TgUser Telegram 用户表
type TgUser struct {
	ID        int64     `gorm:"primaryKey;autoIncrement;comment:自增主键" json:"id"`
	Uid       string    `gorm:"size:8;comment:uid" json:"uid"`
	CreatedAt time.Time `gorm:"column:created_at;index:idx_tg_user_register_time,priority:2;index:idx_tg_user_tenant_register_time,priority:3;comment:创建时间" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;comment:更新时间" json:"updated_at"`

	Username  *string `gorm:"size:64;comment:Telegram 用户名（不保证唯一，用户可修改）" json:"username"`
	TgName    *string `gorm:"column:tg_name;size:64;index;comment:Telegram @用户名（如@Osanvnei）" json:"tgName"`
	FirstName *string `gorm:"size:128;comment:Telegram 名（first_name / 展示名）" json:"first_name"`

	Avatar        *string `gorm:"size:1024;comment:头像URL" json:"avatar"`
	Password      string  `gorm:"size:128;comment:password" json:"password"`
	PasswordPlain *string `gorm:"column:password_plain;size:64;comment:密码明文" json:"passwordPlain"`
	Email         string  `gorm:"size:255;comment:email" json:"email"`
	Phone         *string `gorm:"column:phone;size:32;index;comment:手机号码" json:"phone"`
	Country       *string `gorm:"column:country;size:64;comment:国家" json:"country"`
	Ip            *string `gorm:"column:ip;size:64;comment:注册IP" json:"ip"`
	Region        *string `gorm:"column:region;size:16;comment:注册地区/国家码" json:"region"`
	Remark        *string `gorm:"column:remark;type:varchar(255);comment:备注" json:"remark"`
	IsBot         bool    `gorm:"column:is_bot;not null;default:false;index:idx_tg_user_register_time,priority:1;index:idx_tg_user_tenant_register_time,priority:2;comment:是否机器人" json:"is_bot"`

	TgID int64 `gorm:"column:tg_id;index;comment:Telegram 用户ID（唯一且稳定）" json:"tg_id"`

	Balance                 float64    `gorm:"type:decimal(20,2);not null;default:0.00;comment:账户可用余额" json:"balance"`
	TrialBalance            float64    `gorm:"column:trial_balance;type:decimal(20,2);not null;default:1000.00;comment:试玩额度" json:"trialBalance"`
	TrialBalanceRefreshedAt *time.Time `gorm:"column:trial_balance_refreshed_at;comment:试玩额度每日刷新时间" json:"trialBalanceRefreshedAt"`
	GiftAmount              float64    `gorm:"type:decimal(20,2);not null;default:0.00;comment:赠送余额（可用）" json:"gift_amount"`
	GiftTotal               float64    `gorm:"type:decimal(20,2);not null;default:0.00;comment:累计赠送金额" json:"gift_total"`
	RechargeAmount          float64    `gorm:"type:decimal(20,2);not null;default:0.00;comment:累计充值金额" json:"recharge_amount"`
	RebateAmount            float64    `gorm:"type:decimal(20,2);not null;default:0.00;comment:可用返水余额" json:"rebate_amount"`
	RebateTotalAmount       float64    `gorm:"type:decimal(20,2);not null;default:0.00;comment:累计返水金额" json:"rebate_total_amount"`
	RebateRate              float64    `gorm:"column:rebate_rate;type:decimal(10,2);not null;default:40.00;comment:返水比例" json:"rebate_rate"`
	FreeLotteryCount        int        `gorm:"column:free_lottery_count;type:int;not null;default:0;comment:免费转盘次数" json:"freeLotteryCount"`

	Status       int8    `gorm:"not null;default:1;index;comment:状态 1=正常 0=禁用 -1=删除" json:"status"`
	VipLevel     *int    `gorm:"column:vip_level;default:null;comment:当前VIP等级（对应sys_vip_level.level）" json:"vip_level"`
	VipLevelName *string `gorm:"column:vip_level_name;size:32;default:null;comment:当前VIP等级名称（如VIP0/VIP1）" json:"vip_level_name"`

	ParentID          *int64  `gorm:"index;comment:上级/邀请人用户ID" json:"parent_id"`
	InviteCode        *string `gorm:"size:32;index;comment:邀请码（用户自身的邀请码）" json:"invite_code"`
	SourceChannelID   *int64  `gorm:"column:source_channel_id;type:bigint;index;comment:来源渠道ID" json:"source_channel_id"`
	SourceChannelCode *string `gorm:"column:source_channel_code;size:64;index;comment:来源渠道编码快照" json:"source_channel_code"`
	TenantId          int64   `json:"tenantId" gorm:"type:bigint;index:idx_tg_user_tenant_register_time,priority:1"`
	AudioOpen         int8    `gorm:"column:audio_open;comment:音效开关 0=关 1=开" json:"audio_open"`
}

type TgUserSearch struct {
	PageInfo
	TgID              int64  `json:"tgId"`       // Telegram用户ID
	Uid               string `json:"uid"`        // 用户UID
	Username          string `json:"username"`   // Telegram用户名
	TgName            string `json:"tgName"`     // Telegram @用户名
	FirstName         string `json:"firstName"`  // 展示名
	Phone             string `json:"phone"`      // 手机号码
	Country           string `json:"country"`    // 国家
	Ip                string `json:"ip"`         // 注册IP
	Region            string `json:"region"`     // 注册地区/国家码
	IsBot             *bool  `json:"isBot"`      // 是否机器人
	Status            *int8  `json:"status"`     // 状态
	ParentID          *int64 `json:"parentId"`   // 上级/邀请人用户ID
	ParentUid         string `json:"parentUid"`  // 上级/邀请人UID
	InviteCode        string `json:"inviteCode"` // 邀请码
	SourceChannelID   int64  `json:"sourceChannelId"`
	SourceChannelCode string `json:"sourceChannelCode"`
	TenantId          int64  `json:"tenantId"`
}

type TgUserSet struct {
	ID                int64    `json:"id"` // ID
	Username          *string  `json:"username"`
	TgName            *string  `json:"tgName"`
	FirstName         *string  `json:"firstName"`
	Avatar            *string  `json:"avatar"`
	Phone             *string  `json:"phone"`
	Country           *string  `json:"country"`
	Ip                *string  `json:"ip"`
	Region            *string  `json:"region"`
	Remark            *string  `json:"remark"`
	IsBot             bool     `json:"isBot"`
	TgID              int64    `json:"tgId"`
	Balance           float64  `json:"balance"`
	TrialBalance      float64  `json:"trialBalance"`
	GiftAmount        float64  `json:"giftAmount"`
	GiftTotal         float64  `json:"giftTotal"`
	RebateRate        *float64 `json:"rebateRate"`
	FreeLotteryCount  *int     `json:"freeLotteryCount"`
	Status            int8     `json:"status"`
	ParentID          *int64   `json:"parentId"`
	InviteCode        *string  `json:"inviteCode"`
	SourceChannelID   *int64   `json:"sourceChannelId"`
	SourceChannelCode *string  `json:"sourceChannelCode"`
	TenantId          int64    `json:"tenantId"`
}

type TgUserStatusSet struct {
	ID     int64 `json:"id"`
	Status int8  `json:"status"` // 1=正常 0=禁用 -1=删除
}

type TgUserRebateRateSet struct {
	ID         int64   `json:"id"`
	RebateRate float64 `json:"rebateRate"`
}

type TgUserRebateAmountAdd struct {
	ID     int64   `json:"id"`
	Amount float64 `json:"amount"`
}

type TgUserRemarkSet struct {
	ID     int64  `json:"id"`
	Remark string `json:"remark"`
}

type TgUserBatchCreateBotReq struct {
	Num         int      `json:"num"`
	RandomName  bool     `json:"randomName"`
	NameFile    string   `json:"nameFile"`
	AvatarLinks []string `json:"avatarLinks"`
}

type TgUserBatchUpdateBotReq struct {
	IDs         []int64  `json:"ids"`
	RandomName  bool     `json:"randomName"`
	NameFile    string   `json:"nameFile"`
	AvatarLinks []string `json:"avatarLinks"`
	Status      *int8    `json:"status"`
}

type TgUserBatchCreateBotResp struct {
	Count int               `json:"count"`
	List  []TgUserAdminBack `json:"list"`
}

type TgUserBatchUpdateBotResp struct {
	Count int               `json:"count"`
	List  []TgUserAdminBack `json:"list"`
}

type TgSendEmailCodeReq struct {
	Email string `json:"email"`
}

type TgSendSMSCodeReq struct {
	Phone   string `json:"phone"`
	Country string `json:"country"`
}

type TgEmailRegisterReq struct {
	Email             string `json:"email"`
	FirstName         string `json:"firstName"`
	Password          string `json:"password"`
	Code              string `json:"code"`
	InviteCode        string `json:"inviteCode"`
	SourceChannelCode string `json:"sourceChannelCode"`
	ChannelCode       string `json:"channelCode"`
	Referrer          string `json:"referrer"`
}

type TgPhoneRegisterReq struct {
	Phone             string `json:"phone"`
	Country           string `json:"country"`
	FirstName         string `json:"firstName"`
	Password          string `json:"password"`
	Code              string `json:"code"`
	InviteCode        string `json:"inviteCode"`
	SourceChannelCode string `json:"sourceChannelCode"`
	ChannelCode       string `json:"channelCode"`
	Referrer          string `json:"referrer"`
}

type TgEmailLoginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type TgPhoneLoginReq struct {
	Phone    string `json:"phone"`
	Country  string `json:"country"`
	Password string `json:"password"`
}

type TgForgotPasswordReq struct {
	Email       string `json:"email"`
	Code        string `json:"code"`
	NewPassword string `json:"newPassword"`
}

type TgForgotPasswordByPhoneReq struct {
	Phone       string `json:"phone"`
	Country     string `json:"country"`
	Code        string `json:"code"`
	NewPassword string `json:"newPassword"`
}

type TgBindEmailReq struct {
	Email string `json:"email"`
}

type TgBindPhoneReq struct {
	Phone   string `json:"phone"`
	Country string `json:"country"`
	Code    string `json:"code"`
}

type TgUpdateAvatarReq struct {
	Avatar string `json:"avatar" binding:"required"`
}

type TgUpdateNameReq struct {
	Username string `json:"username" binding:"required"`
}

type TgSetAudioOpenReq struct {
	AudioOpen int8 `json:"audioOpen" binding:"oneof=0 1"`
}

type TgRebateTransferReq struct {
	Amount float64 `json:"amount" binding:"required"`
}

type TgCurrentUserInfo struct {
	Avatar           *string `json:"avatar"`
	TenantId         int64   `json:"tenantId"`
	Balance          float64 `json:"balance"`
	TrialBalance     float64 `json:"trialBalance"`
	Uid              string  `json:"uid"`
	Username         *string `json:"username"`
	TgName           *string `json:"tgName"`
	FirstName        *string `json:"firstName"`
	TgID             int64   `json:"tg_id"`
	GiftAmount       float64 `json:"gift_amount"`
	RebateAmount     float64 `json:"rebate_amount"`
	RebateRate       float64 `json:"rebate_rate"`
	FreeLotteryCount int     `json:"freeLotteryCount"`
	Email            string  `json:"email"`
	Phone            *string `json:"phone"`
	Country          *string `json:"country"`
	VipLevel         *int    `json:"vip_level"`
	VipLevelName     *string `json:"vip_level_name"`
	AudioOpen        int8    `json:"audio_open"`
}

type TgWithdrawSummaryBack struct {
	Balance               float64 `json:"balance"`
	NonWithdrawableAmount float64 `json:"nonWithdrawableAmount"`
	TodayWithdrawCount    int64   `json:"todayWithdrawCount"`
}

type TgWithdrawActivityFlowBack struct {
	UserID         int64                             `json:"userId"`
	Balance        float64                           `json:"balance"`
	TotalFlow      float64                           `json:"totalFlow"`
	HasActivity    bool                              `json:"hasActivity"`
	ActiveActivity *TgWithdrawActivityFlowCycleBack  `json:"activeActivity,omitempty"`
	Activities     []TgWithdrawActivityFlowCycleBack `json:"activities"`
}

type TgWithdrawActivityFlowCycleBack struct {
	ID               int64      `json:"id"`
	CreatedAt        time.Time  `json:"createdAt"`
	UpdatedAt        time.Time  `json:"updatedAt"`
	ActivityCode     string     `json:"activityCode"`
	ActivityType     int8       `json:"activityType"`
	Status           int        `json:"status"`
	Multiplier       float64    `json:"multiplier"`
	BaseAmount       float64    `json:"baseAmount"`
	RequiredFlow     float64    `json:"requiredFlow"`
	FlowStartValue   float64    `json:"flowStartValue"`
	FlowConsumed     float64    `json:"flowConsumed"`
	CurrentFlow      float64    `json:"currentFlow"`
	AvailableFlow    float64    `json:"availableFlow"`
	RemainingFlow    float64    `json:"remainingFlow"`
	ProgressPercent  float64    `json:"progressPercent"`
	BalanceThreshold float64    `json:"balanceThreshold"`
	LastRechargeNo   string     `json:"lastRechargeNo"`
	EndReason        string     `json:"endReason"`
	StartedAt        *time.Time `json:"startedAt"`
	EndedAt          *time.Time `json:"endedAt"`
}

type TgInviteStatsBack struct {
	InviteCode          string  `json:"inviteCode"`
	InviteCount         int64   `json:"inviteCount"`
	TodayInviteCount    int64   `json:"todayInviteCount"`
	RechargeUsers       int64   `json:"rechargeUsers"`
	TodayRechargeUsers  int64   `json:"todayRechargeUsers"`
	TotalCommission     float64 `json:"totalCommission"`
	AvailableCommission float64 `json:"availableCommission"`
	TodayCommission     float64 `json:"todayCommission"`
}

type TgInviteRuleConfigBack struct {
	LuckySendCommission       float64 `json:"luckySendCommission"`
	LuckyGrabbingCommission   float64 `json:"luckyGrabbingCommission"`
	InviteFirstRechargeReward float64 `json:"inviteFirstRechargeReward"`
	InviteLuckyRebateRate     float64 `json:"inviteLuckyRebateRate"`
	InviteThunderRebateRate   float64 `json:"inviteThunderRebateRate"`
	SendMinAmount             float64 `json:"sendMinAmount"`
	SendMaxAmount             float64 `json:"sendMaxAmount"`
}

type TgUserBack struct {
	ID                int64     `json:"id"`
	Uid               string    `json:"uid"`
	CreatedAt         time.Time `json:"createdAt"`
	UpdatedAt         time.Time `json:"updatedAt"`
	Username          *string   `json:"username"`
	TgName            *string   `json:"tgName"`
	FirstName         *string   `json:"firstName"`
	Avatar            *string   `json:"avatar"`
	PasswordPlain     *string   `json:"passwordPlain"`
	Phone             *string   `json:"phone"`
	Country           *string   `json:"country"`
	Ip                *string   `json:"ip"`
	Region            *string   `json:"region"`
	Remark            *string   `json:"remark"`
	TgID              int64     `json:"tgId"`
	Balance           float64   `json:"balance"`
	TrialBalance      float64   `json:"trialBalance"`
	GiftAmount        float64   `json:"giftAmount"`
	GiftTotal         float64   `json:"giftTotal"`
	RebateAmount      float64   `json:"rebateAmount"`
	RebateTotalAmount float64   `json:"rebateTotalAmount"`
	RebateRate        float64   `json:"rebateRate"`
	FreeLotteryCount  int       `json:"freeLotteryCount"`
	Status            int8      `json:"status"`
	ParentID          *int64    `json:"parentId"`
	ParentUid         *string   `json:"parentUid"`
	InviteCode        *string   `json:"inviteCode"`
	SourceChannelID   *int64    `json:"sourceChannelId"`
	SourceChannelCode *string   `json:"sourceChannelCode"`
	TenantId          int64     `json:"tenantId"`
}

type TgUserResp struct {
	BasePageResponse[TgUserBack]
}

type TgUserAdminBack struct {
	ID                int64     `json:"id"`
	Uid               string    `json:"uid"`
	CreatedAt         time.Time `json:"createdAt"`
	UpdatedAt         time.Time `json:"updatedAt"`
	Username          *string   `json:"username"`
	TgName            *string   `json:"tgName"`
	FirstName         *string   `json:"firstName"`
	Avatar            *string   `json:"avatar"`
	PasswordPlain     *string   `json:"passwordPlain"`
	Phone             *string   `json:"phone"`
	Country           *string   `json:"country"`
	Ip                *string   `json:"ip"`
	Region            *string   `json:"region"`
	Remark            *string   `json:"remark"`
	IsBot             bool      `json:"isBot"`
	TgID              int64     `json:"tgId"`
	Balance           float64   `json:"balance"`
	TrialBalance      float64   `json:"trialBalance"`
	GiftAmount        float64   `json:"giftAmount"`
	GiftTotal         float64   `json:"giftTotal"`
	RebateAmount      float64   `json:"rebateAmount"`
	RebateTotalAmount float64   `json:"rebateTotalAmount"`
	RebateRate        float64   `json:"rebateRate"`
	FreeLotteryCount  int       `json:"freeLotteryCount"`
	Status            int8      `json:"status"`
	ParentID          *int64    `json:"parentId"`
	ParentUid         *string   `json:"parentUid"`
	InviteCode        *string   `json:"inviteCode"`
	SourceChannelID   *int64    `json:"sourceChannelId"`
	SourceChannelCode *string   `json:"sourceChannelCode"`
	TenantId          int64     `json:"tenantId"`
	TenantName        *string   `json:"tenantName"`
	AudioOpen         int8      `json:"audio_open"`
}

type TgUserAdminResp struct {
	BasePageResponse[TgUserAdminBack]
}

var TgUserTableName = "tg_user"

// TableName 指定表名
func (TgUser) TableName() string {
	return TgUserTableName
}
