package services

import (
	"log"
	"strings"
	"time"

	"BaseGoUni/core/pojo"
	"BaseGoUni/core/utils"

	"gorm.io/gorm"
)

const GGRTransactionRetention = 24 * time.Hour

// CleanupExpiredGGRTransactionsAllHosts removes GGR idempotency records whose
// server receipt time is older than the configured 24-hour retention window.
func CleanupExpiredGGRTransactionsAllHosts() {
	startedAt := time.Now()
	cutoff := GGRTransactionCleanupCutoff(startedAt)

	var hostInfos []pojo.HostInfo
	if err := utils.Db.Model(&pojo.HostInfo{}).
		Where("enabled = ?", true).
		Where("table_prefix <> ''").
		Find(&hostInfos).Error; err != nil {
		log.Printf("[ggr_cleanup] load host infos failed err=%v", err)
		return
	}

	seenPrefixes := make(map[string]struct{}, len(hostInfos))
	var totalDeleted int64
	failed := 0
	for _, hostInfo := range hostInfos {
		prefix := strings.TrimSpace(hostInfo.TablePrefix)
		if prefix == "" {
			continue
		}
		if _, exists := seenPrefixes[prefix]; exists {
			continue
		}
		seenPrefixes[prefix] = struct{}{}

		db := utils.NewPrefixDb(prefix)
		if db == nil {
			failed++
			log.Printf("[ggr_cleanup] skipped prefix=%s: db not ready", prefix)
			continue
		}
		deleted, err := DeleteGGRTransactionsBefore(db, cutoff)
		if err != nil {
			failed++
			log.Printf("[ggr_cleanup] delete failed prefix=%s cutoff=%s err=%v", prefix, cutoff.Format(time.RFC3339), err)
			continue
		}
		totalDeleted += deleted
	}

	log.Printf(
		"[ggr_cleanup] finished prefixes=%d deleted=%d failed=%d cutoff=%s cost=%.2fs",
		len(seenPrefixes),
		totalDeleted,
		failed,
		cutoff.Format(time.RFC3339),
		time.Since(startedAt).Seconds(),
	)
}

func GGRTransactionCleanupCutoff(now time.Time) time.Time {
	return now.Add(-GGRTransactionRetention)
}

func DeleteGGRTransactionsBefore(db *gorm.DB, cutoff time.Time) (int64, error) {
	if db == nil {
		return 0, gorm.ErrInvalidDB
	}
	result := db.Where("received_at < ?", cutoff).Delete(&pojo.GGRTransaction{})
	return result.RowsAffected, result.Error
}
