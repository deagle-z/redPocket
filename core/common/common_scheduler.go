package common

import (
	"BaseGoUni/core/services"
	"BaseGoUni/core/utils"
	"github.com/robfig/cron/v3"
	"log"
	"time"
)

func InitScheduler() {
	// 初始化定时任务
	c := cron.New(cron.WithSeconds())
	// 是否运行调度器的检查
	if !utils.CsConfig.RunScheduler {
		c.Start()
		return
	}

	// 公共的任务添加函数，减少重复代码
	addScheduledTask := func(schedule string, lockKey string, lockDuration time.Duration, task func(), logMessage string) {
		_, err := c.AddFunc(schedule, func() {
			lock, lockErr := utils.AcquireLock(lockKey, lockDuration)
			if lockErr != nil {
				log.Printf("任务获取锁异常: %s, error: %v", logMessage, lockErr)
				return
			}
			if !lock {
				return
			}
			defer func() {
				if releaseErr := utils.ReleaseLock(lockKey); releaseErr != nil {
					log.Printf("任务释放锁异常: %s, error: %v", logMessage, releaseErr)
				}
			}()
			task()
		})
		if err != nil {
			log.Printf("添加任务失败: %s, error: %v", logMessage, err)
		} else {
			log.Printf("添加任务成功: %s", logMessage)
		}
	}

	addScheduledTask("*/10 * * * * *", "host_info_get", 1*time.Minute, func() {
		utils.FlushTempHostInfo()
	}, "")
	addScheduledTask("0 * * * * *", "lucky_expire_sweep", 2*time.Minute, func() {
		services.SweepExpiredLuckyPacketsAllHosts()
		services.SweepExpiredTrialLuckyPacketsAllHosts()
	}, "扫描过期红包")
	addScheduledTask("30 * * * * *", "trial_lucky_ensure", 2*time.Minute, func() {
		services.EnsureMinActiveTrialLuckyPacketsAllHosts()
	}, "补齐试玩机器人红包")
	addScheduledTask("0 * * * * *", "withdraw_payout_query_all_hosts", 2*time.Minute, func() {
		services.SweepTimedOutProcessingWithdrawPayoutOrdersAllHosts()
	}, "处理超过5小时的代付订单")
	addScheduledTask("0 0 * * * *", "ggr_transaction_cleanup_all_hosts", 30*time.Minute, func() {
		services.CleanupExpiredGGRTransactionsAllHosts()
	}, "清理24小时前的GGR交易幂等记录")
	services.StartBotLotteryTask()
	services.StartUsdtRechargeScanTask()
	c.Start()
	log.Println("Scheduler started successfully")
}
