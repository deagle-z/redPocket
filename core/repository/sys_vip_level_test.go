package repository

import (
	"BaseGoUni/core/pojo"
	"testing"
)

func TestBuildVipUpgradeCashHistoryUsesVipRewardType(t *testing.T) {
	item := pojo.SysVipRewardLog{
		BaseModel:   pojo.BaseModel{ID: 42},
		LevelName:   "VIP1",
		BonusAmount: 8,
	}

	history := buildVipUpgradeCashHistory(1001, item, 100)

	if history.AwardUni != "vip_reward_42" {
		t.Fatalf("AwardUni = %q, want vip_reward_42", history.AwardUni)
	}
	if history.Type != pojo.CashHistoryTypeVipUpgradeReward {
		t.Fatalf("Type = %d, want %d", history.Type, pojo.CashHistoryTypeVipUpgradeReward)
	}
	if history.IsGift != 1 {
		t.Fatalf("IsGift = %d, want 1", history.IsGift)
	}
	if history.StartAmount != 100 || history.EndAmount != 108 {
		t.Fatalf("balance range = %.2f -> %.2f, want 100.00 -> 108.00", history.StartAmount, history.EndAmount)
	}
}
