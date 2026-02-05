package utils

import (
	"BaseGoUni/core/pojo"
	"encoding/json"
	"time"
)

func NotifyManageLog(logStr string, prefix string) {
	var logInfo pojo.ManageLog
	_ = json.Unmarshal([]byte(logStr), &logInfo)
	if logInfo.Ip == "" {
		return
	}
	db := NewPrefixDb(prefix)
	db.Create(&logInfo)
}

func CleanManageLog(data string, prefix string) {
	now := time.Now()
	endTime := now.AddDate(0, 0, -14)
	db := NewPrefixDb(prefix)
	db.Where("created_at < ?", endTime).Delete(&pojo.ManageLog{})
}
