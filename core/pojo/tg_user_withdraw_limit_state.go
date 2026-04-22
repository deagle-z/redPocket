package pojo

import "time"

type TgUserWithdrawLimitState struct {
	BaseModel
	UserID                    int64      `json:"userId" gorm:"column:user_id;type:bigint;uniqueIndex"`
	GiftRestrictedBalance     float64    `json:"giftRestrictedBalance" gorm:"column:gift_restricted_balance;type:numeric(20,3);not null;default:0"`
	RechargeRestrictedBalance float64    `json:"rechargeRestrictedBalance" gorm:"column:recharge_restricted_balance;type:numeric(20,3);not null;default:0"`
	GiftFlowConsumed          float64    `json:"giftFlowConsumed" gorm:"column:gift_flow_consumed;type:numeric(20,3);not null;default:0"`
	RechargeFlowConsumed      float64    `json:"rechargeFlowConsumed" gorm:"column:recharge_flow_consumed;type:numeric(20,3);not null;default:0"`
	InitializedAt             *time.Time `json:"initializedAt" gorm:"column:initialized_at;type:datetime(3)"`
}

var TgUserWithdrawLimitStateTableName = "tg_user_withdraw_limit_state"

func (TgUserWithdrawLimitState) TableName() string {
	return TgUserWithdrawLimitStateTableName
}
