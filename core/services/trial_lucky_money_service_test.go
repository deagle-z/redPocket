package services

import (
	"BaseGoUni/core/pojo"
	"testing"
)

func TestCanTrialUserGrabSenderAllowsOwnUserPacket(t *testing.T) {
	if !canTrialUserGrabSender(pojo.TrialActorUser, 10, 10) {
		t.Fatalf("canTrialUserGrabSender() = false, want true")
	}
}

func TestCanTrialBotGrabSenderRejectsOwnBotPacket(t *testing.T) {
	if canTrialBotGrabSender(pojo.TrialActorBot, 10, 10) {
		t.Fatalf("canTrialBotGrabSender() = true, want false")
	}
}
