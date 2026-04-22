package services

import (
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/repository"
	"BaseGoUni/core/utils"
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
)

func handleRechargeFirstGiftInstallmentTask(ctx context.Context, task *asynq.Task) error {
	var payload pojo.RechargeFirstGiftInstallmentPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return err
	}
	if payload.TablePrefix == "" || payload.OrderNo == "" || payload.InstallmentIndex <= 0 || payload.GiftAmount <= 0 {
		return nil
	}

	db := utils.NewPrefixDb(payload.TablePrefix)
	if db == nil {
		return fmt.Errorf("db not ready for prefix=%s", payload.TablePrefix)
	}
	return repository.ApplyFirstRechargeGiftInstallmentByOrderNo(
		db,
		payload.OrderNo,
		payload.InstallmentIndex,
		payload.GiftAmount,
		payload.TotalRate,
		payload.Ratio,
		payload.RatioBase,
	)
}
