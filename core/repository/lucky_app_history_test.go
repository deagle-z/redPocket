package repository

import (
	"BaseGoUni/core/pojo"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestLuckyAppHistoryBackExposesGameMode(t *testing.T) {
	field, ok := reflect.TypeOf(pojo.LuckyAppHistoryBack{}).FieldByName("GameMode")
	if !ok {
		t.Fatalf("LuckyAppHistoryBack missing GameMode")
	}
	if got := field.Tag.Get("json"); got != "gameMode" {
		t.Fatalf("GameMode json tag = %q, want gameMode", got)
	}
	if got := field.Tag.Get("gorm"); !strings.Contains(got, "column:game_mode") {
		t.Fatalf("GameMode gorm tag = %q, want column:game_mode", got)
	}
}

func TestLuckyAppHistoryUnionSelectsGameMode(t *testing.T) {
	src, err := os.ReadFile("lucky_history.go")
	if err != nil {
		t.Fatalf("read lucky_history.go: %v", err)
	}
	content := string(src)
	if count := strings.Count(content, "m.game_mode AS game_mode"); count != 2 {
		t.Fatalf("m.game_mode AS game_mode count = %d, want 2", count)
	}
	if !strings.Contains(content, "t.game_mode") {
		t.Fatalf("list query missing t.game_mode")
	}
}
