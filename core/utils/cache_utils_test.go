package utils

import (
	"os"
	"strings"
	"testing"
)

func TestGetStringCacheDoesNotLogCacheHits(t *testing.T) {
	source, err := os.ReadFile("cache_utils.go")
	if err != nil {
		t.Fatalf("read cache_utils.go: %v", err)
	}

	if strings.Contains(string(source), "[config-cache] string cache hit") {
		t.Fatal("GetStringCache should not log Redis cache hits")
	}
}
