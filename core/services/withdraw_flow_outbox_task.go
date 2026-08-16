package services

import (
	"BaseGoUni/core/repository"
	"BaseGoUni/core/utils"
	"log"
	"strings"
	"sync"
	"time"
)

const (
	withdrawFlowOutboxBatchSize    = 100
	withdrawFlowOutboxPollInterval = 2 * time.Second
	withdrawFlowOutboxLockTTL      = 20 * time.Second
)

var withdrawFlowOutboxWorkerOnce sync.Once

func StartWithdrawFlowOutboxWorkerDefaultTable() {
	withdrawFlowOutboxWorkerOnce.Do(func() {
		go func() {
			ticker := time.NewTicker(withdrawFlowOutboxPollInterval)
			defer ticker.Stop()
			for {
				sweepWithdrawFlowOutboxDefaultTableOnce()
				<-ticker.C
			}
		}()
	})
}

func sweepWithdrawFlowOutboxDefaultTableOnce() {
	prefix := strings.TrimSpace(utils.CsConfig.DefaultHost.TablePrefix)
	if prefix == "" {
		return
	}
	lockKey := "withdraw_flow_outbox_default:" + prefix
	locked, err := utils.AcquireLock(lockKey, withdrawFlowOutboxLockTTL)
	if err != nil || !locked {
		return
	}
	defer utils.ReleaseLock(lockKey)

	_, err = repository.ProcessWithdrawFlowOutboxBatch(utils.NewPrefixDb(prefix), withdrawFlowOutboxBatchSize)
	if err != nil {
		log.Printf("[withdraw-flow-outbox] process default table failed prefix=%s err=%v", prefix, err)
	}
}
