package services

import (
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/repository"
	"BaseGoUni/core/utils"
	"log"
	"strings"
	"time"
)

func SweepTimedOutProcessingWithdrawPayoutOrdersAllHosts() {
	startAt := time.Now()
	var hostInfos []pojo.HostInfo
	if err := utils.Db.Model(&pojo.HostInfo{}).
		Where("enabled = ?", true).
		Where("table_prefix <> ''").
		Find(&hostInfos).Error; err != nil {
		log.Printf("[withdraw_payout_query] sweep host infos failed err=%v", err)
		return
	}

	seenPrefixes := make(map[string]struct{}, len(hostInfos))
	totalMatched := 0
	totalProcessed := 0
	totalFailed := 0
	for _, hostInfo := range hostInfos {
		tablePrefix := strings.TrimSpace(hostInfo.TablePrefix)
		if tablePrefix == "" {
			continue
		}
		if _, ok := seenPrefixes[tablePrefix]; ok {
			continue
		}
		seenPrefixes[tablePrefix] = struct{}{}
		db := utils.NewPrefixDb(tablePrefix)
		if db == nil {
			totalFailed++
			log.Printf("[withdraw_payout_query] sweep skipped prefix=%s: db not ready", tablePrefix)
			continue
		}
		matched, processed, failed := repository.SweepTimedOutProcessingWithdrawPayoutOrders(db)
		totalMatched += matched
		totalProcessed += processed
		totalFailed += failed
	}
	if totalMatched > 0 || totalFailed > 0 {
		log.Printf("[withdraw_payout_query] sweep finished prefixes=%d matched=%d processed=%d failed=%d cost=%.2fs",
			len(seenPrefixes), totalMatched, totalProcessed, totalFailed, time.Since(startAt).Seconds())
	}
}
