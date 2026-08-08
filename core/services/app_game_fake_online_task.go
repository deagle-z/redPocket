package services

import (
	"BaseGoUni/core/repository"
	"BaseGoUni/core/utils"
	"log"
	"time"
)

func RefreshAppGameFakeOnlineCountsAllTenants() {
	startTime := time.Now()
	prefixCount := 0
	updatedCount := int64(0)
	failedCount := 0

	utils.LimitPrefix(func(prefix string) {
		prefixCount++
		db := utils.NewPrefixDb(prefix)
		if db == nil {
			failedCount++
			log.Printf("[app-game] refresh fake online count failed: prefix=%s error=database_unavailable", prefix)
			return
		}

		rowsAffected, err := repository.RefreshAppGameFakeOnlineCounts(db)
		if err != nil {
			failedCount++
			log.Printf("[app-game] refresh fake online count failed: prefix=%s error=%v", prefix, err)
			return
		}
		updatedCount += rowsAffected
	})

	log.Printf(
		"[app-game] fake online refresh finished: prefixes=%d updated=%d failed=%d cost=%.2fs",
		prefixCount,
		updatedCount,
		failedCount,
		time.Since(startTime).Seconds(),
	)
}
