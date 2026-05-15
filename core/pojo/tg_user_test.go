package pojo

import (
	"reflect"
	"strings"
	"testing"
)

func TestTgUserIncludesTgNameColumn(t *testing.T) {
	field, ok := reflect.TypeOf(TgUser{}).FieldByName("TgName")
	if !ok {
		t.Fatal("TgUser should include TgName")
	}

	gormTag := field.Tag.Get("gorm")
	if !strings.Contains(gormTag, "column:tg_name") {
		t.Fatalf("TgName gorm tag should use column:tg_name, got %q", gormTag)
	}
	if got := field.Tag.Get("json"); got != "tgName" {
		t.Fatalf("TgName json tag = %q, want %q", got, "tgName")
	}
}

func TestTgChannelMemberTableName(t *testing.T) {
	if got := (TgChannelMember{}).TableName(); got != "tg_channel_member" {
		t.Fatalf("TgChannelMember table name = %q, want %q", got, "tg_channel_member")
	}
}
