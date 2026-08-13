package repository

import (
	"BaseGoUni/core/pay"
	"BaseGoUni/core/pojo"
	"fmt"
	"log"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	WithdrawPayoutQueryTimeout = 5 * time.Hour
	withdrawPayoutQueryLimit   = 200
)

const (
	withdrawPayoutTimeoutFailCode = "PAYOUT_TIMEOUT_5H"
	withdrawPayoutTimeoutFailMsg  = "代付发起超过5小时仍未完成，订单已失败并退款"
)

type withdrawPayoutQueryAction int

const (
	withdrawPayoutQueryActionProcessing withdrawPayoutQueryAction = iota
	withdrawPayoutQueryActionSuccess
	withdrawPayoutQueryActionFailed
)

func SweepTimedOutProcessingWithdrawPayoutOrders(db *gorm.DB) (matched int, processed int, failed int) {
	return SweepTimedOutProcessingWithdrawPayoutOrdersBefore(db, time.Now().Add(-WithdrawPayoutQueryTimeout), withdrawPayoutQueryLimit)
}

func SweepTimedOutProcessingWithdrawPayoutOrdersBefore(db *gorm.DB, cutoff time.Time, limit int) (matched int, processed int, failed int) {
	if db == nil {
		return 0, 0, 0
	}
	if limit <= 0 {
		limit = withdrawPayoutQueryLimit
	}

	var orders []pojo.WithdrawOrderBr
	if err := db.Model(&pojo.WithdrawOrderBr{}).
		Where("status = ?", 2).
		Where("updated_at <= ?", cutoff).
		Order("updated_at asc").
		Limit(limit).
		Find(&orders).Error; err != nil {
		log.Printf("[withdraw_payout_query] query timed out orders failed err=%v", err)
		return 0, 0, 0
	}

	matched = len(orders)
	for _, order := range orders {
		action, err := processTimedOutProcessingWithdrawPayoutOrder(db, order, cutoff)
		if err != nil {
			failed++
			log.Printf("[withdraw_payout_query] process failed orderID=%d orderNo=%s err=%v", order.ID, order.OrderNo, err)
			continue
		}
		if action != withdrawPayoutQueryActionProcessing {
			processed++
		}
	}
	return matched, processed, failed
}

func ProcessTimedOutProcessingWithdrawPayoutOrder(db *gorm.DB, order pojo.WithdrawOrderBr) (withdrawPayoutQueryAction, error) {
	return processTimedOutProcessingWithdrawPayoutOrder(db, order, time.Now().Add(-WithdrawPayoutQueryTimeout))
}

func processTimedOutProcessingWithdrawPayoutOrder(db *gorm.DB, order pojo.WithdrawOrderBr, cutoff time.Time) (withdrawPayoutQueryAction, error) {
	providerCode := withdrawPayoutProviderCode(order)
	if providerCode == "" {
		return withdrawPayoutQueryActionFailed, failWithdrawPayoutQueryOrder(db, order, WithdrawPayoutCallback{
			ResultCode: "PAYOUT_QUERY_PROVIDER_REQUIRED",
			ResultMsg:  "代付查单失败：订单缺少provider",
			Failed:     true,
		})
	}

	queryProvider := pay.GetPayoutQuery(providerCode)
	if queryProvider == nil {
		return withdrawPayoutQueryActionFailed, failWithdrawPayoutQueryOrder(db, order, WithdrawPayoutCallback{
			ResultCode: "PAYOUT_QUERY_PROVIDER_UNSUPPORTED",
			ResultMsg:  fmt.Sprintf("代付查单失败：provider不支持查单 %s", providerCode),
			Failed:     true,
		})
	}

	resp, err := queryProvider.QueryPayoutOrder(withdrawPayoutQueryRequest(order))
	if err != nil {
		return withdrawPayoutQueryActionFailed, failWithdrawPayoutQueryOrder(db, order, WithdrawPayoutCallback{
			ProviderPayoutNo: ptrValue(order.ProviderPayoutNo),
			ResultCode:       "PAYOUT_QUERY_NOT_FOUND",
			ResultMsg:        withdrawFailMsgFromError(err),
			Failed:           true,
		})
	}

	action := withdrawPayoutQueryDecision(resp)
	switch action {
	case withdrawPayoutQueryActionSuccess:
		return action, ProcessWithdrawOrderPayoutCallback(db, WithdrawPayoutCallback{
			LocalOrderNo:     withdrawPayoutLocalOrderNo(order),
			ProviderPayoutNo: resp.ProviderOrderNo,
			ProviderStatus:   fmt.Sprintf("%d", resp.Status),
			ProviderAmount:   resp.Amount,
			ProviderPayTime:  resp.PayTime,
			Success:          true,
			Failed:           false,
			ResultCode:       resp.SubCode,
			ResultMsg:        resp.SubMsg,
		})
	case withdrawPayoutQueryActionFailed:
		return action, failWithdrawPayoutQueryOrder(db, order, WithdrawPayoutCallback{
			ProviderPayoutNo: resp.ProviderOrderNo,
			ProviderStatus:   fmt.Sprintf("%d", resp.Status),
			ProviderAmount:   resp.Amount,
			ProviderPayTime:  resp.PayTime,
			ResultCode:       firstWithdrawValue(resp.SubCode, fmt.Sprintf("PAYOUT_STATUS_%d", resp.Status)),
			ResultMsg:        firstWithdrawValue(resp.SubMsg, fmt.Sprintf("三方代付查询失败状态：%d", resp.Status)),
			Failed:           true,
		})
	default:
		timedOut, err := failTimedOutProcessingWithdrawPayoutOrder(db, order, cutoff, WithdrawPayoutCallback{
			ProviderPayoutNo: resp.ProviderOrderNo,
			ProviderStatus:   fmt.Sprintf("%d", resp.Status),
			ResultCode:       withdrawPayoutTimeoutFailCode,
			ResultMsg:        withdrawPayoutTimeoutFailMsg,
			Failed:           true,
		})
		if err != nil {
			return withdrawPayoutQueryActionFailed, err
		}
		if timedOut {
			return withdrawPayoutQueryActionFailed, nil
		}
		return withdrawPayoutQueryActionProcessing, nil
	}
}

func failTimedOutProcessingWithdrawPayoutOrder(db *gorm.DB, order pojo.WithdrawOrderBr, cutoff time.Time, cb WithdrawPayoutCallback) (bool, error) {
	if db == nil || order.ID <= 0 {
		return false, nil
	}

	timedOut := false
	err := db.Transaction(func(tx *gorm.DB) error {
		var lockedOrder pojo.WithdrawOrderBr
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ?", order.ID).
			First(&lockedOrder).Error; err != nil {
			return err
		}
		if lockedOrder.Status != 2 || lockedOrder.UpdatedAt.After(cutoff) {
			return nil
		}

		now := time.Now()
		updates := map[string]any{
			"status":       4,
			"fail_code":    strings.TrimSpace(cb.ResultCode),
			"fail_msg":     strings.TrimSpace(cb.ResultMsg),
			"notify_time":  now,
			"notify_count": gorm.Expr("notify_count + 1"),
		}
		if providerStatus := strings.TrimSpace(cb.ProviderStatus); providerStatus != "" {
			updates["provider_status"] = providerStatus
		}
		if providerPayoutNo := strings.TrimSpace(cb.ProviderPayoutNo); providerPayoutNo != "" {
			updates["provider_payout_no"] = providerPayoutNo
		}
		if err := tx.Model(&pojo.WithdrawOrderBr{}).
			Where("id = ?", lockedOrder.ID).
			Updates(updates).Error; err != nil {
			return err
		}
		if err := refundWithdrawAmount(tx, lockedOrder); err != nil {
			return err
		}
		timedOut = true
		return nil
	})
	return timedOut, err
}

func failWithdrawPayoutQueryOrder(db *gorm.DB, order pojo.WithdrawOrderBr, cb WithdrawPayoutCallback) error {
	cb.LocalOrderNo = withdrawPayoutLocalOrderNo(order)
	cb.Success = false
	cb.Failed = true
	return ProcessWithdrawOrderPayoutCallback(db, cb)
}

func withdrawPayoutQueryDecision(resp pay.PayoutQueryResponse) withdrawPayoutQueryAction {
	switch resp.Status {
	case 7:
		return withdrawPayoutQueryActionSuccess
	case 2, 6, 8:
		return withdrawPayoutQueryActionFailed
	case 0, 9:
		return withdrawPayoutQueryActionProcessing
	default:
		return withdrawPayoutQueryActionProcessing
	}
}

func withdrawPayoutProviderCode(order pojo.WithdrawOrderBr) string {
	return strings.TrimSpace(ptrValue(order.Provider))
}

func withdrawPayoutQueryRequest(order pojo.WithdrawOrderBr) pay.PayoutQueryRequest {
	return pay.PayoutQueryRequest{
		OrderNo:         withdrawPayoutLocalOrderNo(order),
		ProviderOrderNo: strings.TrimSpace(ptrValue(order.ProviderPayoutNo)),
	}
}

func withdrawPayoutLocalOrderNo(order pojo.WithdrawOrderBr) string {
	if merchantOrderNo := strings.TrimSpace(ptrValue(order.MerchantOrderNo)); merchantOrderNo != "" {
		return merchantOrderNo
	}
	return strings.TrimSpace(order.OrderNo)
}
