package repository

import (
	"BaseGoUni/core/pay"
	"BaseGoUni/core/pojo"
	"fmt"
	"log"
	"strings"
	"time"

	"gorm.io/gorm"
)

const (
	WithdrawPayoutQueryTimeout = 24 * time.Hour
	withdrawPayoutQueryLimit   = 200
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
		action, err := ProcessTimedOutProcessingWithdrawPayoutOrder(db, order)
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
		return action, nil
	}
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
