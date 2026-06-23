package services

import (
	"BaseGoUni/core/repository"
	"BaseGoUni/core/utils"
	"log"
	"time"
)

func SweepTimedOutProcessingWithdrawPayoutOrdersDefaultTable() {
	startAt := time.Now()
	matched, processed, failed := repository.SweepTimedOutProcessingWithdrawPayoutOrders(utils.Db)
	log.Printf("[withdraw_payout_query] sweep default table finished matched=%d processed=%d failed=%d cost=%.2fs",
		matched, processed, failed, time.Since(startAt).Seconds())
}
