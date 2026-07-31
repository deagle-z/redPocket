package services

import (
	"BaseGoUni/core/utils"
	"sync"
	"time"
)

const cryptoScanLockHeartbeatInterval = time.Minute

type cryptoScanLockHeartbeat struct {
	lock     *utils.OwnedRedisLock
	stopOnce sync.Once
	stop     chan struct{}
	done     chan struct{}
	mu       sync.Mutex
	err      error
}

func startCryptoScanLockHeartbeat(lock *utils.OwnedRedisLock) *cryptoScanLockHeartbeat {
	heartbeat := &cryptoScanLockHeartbeat{
		lock: lock,
		stop: make(chan struct{}),
		done: make(chan struct{}),
	}
	go heartbeat.run()
	return heartbeat
}

func (heartbeat *cryptoScanLockHeartbeat) run() {
	defer close(heartbeat.done)
	ticker := time.NewTicker(cryptoScanLockHeartbeatInterval)
	defer ticker.Stop()
	for {
		select {
		case <-heartbeat.stop:
			return
		case <-ticker.C:
			if err := heartbeat.lock.Renew(); err != nil {
				heartbeat.mu.Lock()
				heartbeat.err = err
				heartbeat.mu.Unlock()
				return
			}
		}
	}
}

func (heartbeat *cryptoScanLockHeartbeat) Stop() {
	if heartbeat == nil {
		return
	}
	heartbeat.stopOnce.Do(func() { close(heartbeat.stop) })
	<-heartbeat.done
}

func (heartbeat *cryptoScanLockHeartbeat) StopAndVerify() error {
	if heartbeat == nil {
		return nil
	}
	heartbeat.Stop()
	heartbeat.mu.Lock()
	err := heartbeat.err
	heartbeat.mu.Unlock()
	if err != nil {
		return err
	}
	return heartbeat.lock.Renew()
}
