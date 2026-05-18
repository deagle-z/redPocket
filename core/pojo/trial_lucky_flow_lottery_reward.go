package pojo

type TrialLuckyFlowLotteryReward struct {
	BaseModel
	TenantId          int64   `json:"tenantId" gorm:"type:bigint;not null;index:idx_trial_lucky_flow_lottery_reward_once,unique,priority:1"`
	UserID            int64   `json:"userId" gorm:"type:bigint;not null;index:idx_trial_lucky_flow_lottery_reward_once,unique,priority:2"`
	ThresholdAmount   float64 `json:"thresholdAmount" gorm:"type:decimal(20,2);not null;default:0.00"`
	RewardCount       int     `json:"rewardCount" gorm:"type:int;not null;default:0"`
	TotalFlowSnapshot float64 `json:"totalFlowSnapshot" gorm:"type:decimal(20,2);not null;default:0.00"`
	SourceLuckyID     int64   `json:"sourceLuckyId" gorm:"type:bigint;not null;default:0;index"`
	SourceHistoryID   int64   `json:"sourceHistoryId" gorm:"type:bigint;not null;default:0;index"`
}

var TrialLuckyFlowLotteryRewardTableName = "trial_lucky_flow_lottery_reward"

func (TrialLuckyFlowLotteryReward) TableName() string {
	return TrialLuckyFlowLotteryRewardTableName
}

type TrialLuckyFlowLotteryRewardProgress struct {
	Enabled              bool    `json:"enabled"`
	ThresholdAmount      float64 `json:"thresholdAmount"`
	RewardCount          int     `json:"rewardCount"`
	TotalFlow            float64 `json:"totalFlow"`
	RemainingFlow        float64 `json:"remainingFlow"`
	ProgressPercent      float64 `json:"progressPercent"`
	Rewarded             bool    `json:"rewarded"`
	CanReward            bool    `json:"canReward"`
	AvailableRewardCount int     `json:"availableRewardCount"`
	Drawn                bool    `json:"drawn"`
	FreeLotteryCount     int     `json:"freeLotteryCount"`
}
