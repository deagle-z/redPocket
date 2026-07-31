package repository

import (
	"BaseGoUni/core/pojo"
	"errors"
	"math/big"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/plugin/dbresolver"
)

var errCryptoManualSupplementCompletedConflict = errors.New("manual_supplement_completed_conflict")

func GetCryptoRechargeExceptions(db *gorm.DB, search pojo.CryptoRechargeExceptionSearch) (pojo.BasePageResponse[pojo.UsdtRechargeTx], error) {
	if search.CurrentPage < 0 {
		search.CurrentPage = 0
	}
	if search.PageSize <= 0 {
		search.PageSize = 10
	} else if search.PageSize > 100 {
		search.PageSize = 100
	}
	result := pojo.BasePageResponse[pojo.UsdtRechargeTx]{
		PageSize:    search.PageSize,
		CurrentPage: search.CurrentPage,
		List:        make([]pojo.UsdtRechargeTx, 0),
	}
	query := db.Clauses(dbresolver.Write).Model(&pojo.UsdtRechargeTx{}).
		Where("review_status = ?", pojo.CryptoRechargeTxReviewManualRequired)
	if network := strings.ToUpper(strings.TrimSpace(search.Network)); network != "" {
		query = query.Where("network = ?", network)
	}
	if token := strings.ToUpper(strings.TrimSpace(search.Token)); token != "" {
		query = query.Where("token = ?", token)
	}
	if txID := strings.ToLower(strings.TrimSpace(search.TxID)); txID != "" {
		query = query.Where("tx_id = ?", txID)
	}
	if code := strings.ToUpper(strings.TrimSpace(search.ExceptionCode)); code != "" {
		query = query.Where("exception_code = ?", code)
	}
	if err := query.Count(&result.Total).Error; err != nil {
		return result, err
	}
	if err := query.Order("id desc").
		Limit(search.PageSize).
		Offset(search.PageSize * search.CurrentPage).
		Find(&result.List).Error; err != nil {
		return result, err
	}
	return result, nil
}

func AdminSupplementCryptoRecharge(
	db *gorm.DB,
	tablePrefix string,
	operatorID int64,
	req pojo.CryptoRechargeManualSupplementReq,
) (pojo.CryptoRechargeManualSupplementBack, error) {
	var result pojo.CryptoRechargeManualSupplementBack
	if db == nil {
		return result, errors.New("db_not_available")
	}
	req.OrderNo = strings.TrimSpace(req.OrderNo)
	req.Reason = strings.TrimSpace(req.Reason)
	if operatorID <= 0 || req.OrderNo == "" {
		return result, errors.New("manual_supplement_fields_required")
	}
	if utf8.RuneCountInString(req.Reason) < 5 || utf8.RuneCountInString(req.Reason) > 500 {
		return result, errors.New("manual_supplement_reason_invalid")
	}
	proofs, err := normalizeCryptoManualProofs(req.Transactions)
	if err != nil {
		return result, err
	}
	if completed, found, completedErr := loadCompletedCryptoManualSupplement(db, req.OrderNo, req.Reason, proofs); completedErr != nil {
		return result, completedErr
	} else if found {
		return completed, nil
	}

	var cryptoOrder pojo.UsdtRechargeOrder
	if err = db.Clauses(dbresolver.Write).Where("order_no = ?", req.OrderNo).First(&cryptoOrder).Error; err != nil {
		return result, err
	}
	providerTradeNo := "MANUAL_CRYPTO:" + proofs[0].TxID
	var supplement pojo.CryptoRechargeManualSupplement
	observedTxIDs := make([]int64, 0, len(proofs))
	now := time.Now()
	err = processRechargeOrderSuccessWithHooks(
		db,
		cryptoOrder.OrderNo,
		providerTradeNo,
		cryptoOrder.PlatformAmount,
		tablePrefix,
		func(tx *gorm.DB, rechargeOrder *pojo.RechargeOrder) error {
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", cryptoOrder.ID).First(&cryptoOrder).Error; err != nil {
				return err
			}
			if cryptoOrder.Status == pojo.UsdtRechargeStatusPaid || rechargeOrder.Status == 1 {
				return errors.New("crypto_order_already_paid")
			}
			if rechargeOrder.Status != 0 && rechargeOrder.Status != 5 {
				return errors.New("order_status_invalid_credit")
			}
			if rechargeOrder.Status == 5 {
				reopened := tx.Model(&pojo.RechargeOrder{}).
					Where("id = ? AND status = ?", rechargeOrder.ID, 5).
					Updates(map[string]any{"status": 0, "provider_status": "MANUAL_REVIEW"})
				if reopened.Error != nil {
					return reopened.Error
				}
				if reopened.RowsAffected != 1 {
					return errors.New("manual_supplement_reopen_failed")
				}
				rechargeOrder.Status = 0
			}

			supplement = pojo.CryptoRechargeManualSupplement{
				OrderNo:         cryptoOrder.OrderNo,
				RechargeOrderID: rechargeOrder.ID,
				UserID:          rechargeOrder.UserId,
				PlatformAmount:  cryptoOrder.PlatformAmount,
				OperatorID:      operatorID,
				Reason:          req.Reason,
				Status:          pojo.CryptoRechargeManualStatusProcessing,
			}
			if err := tx.Create(&supplement).Error; err != nil {
				return err
			}
			for i := range proofs {
				proofs[i].SupplementID = supplement.ID
				var observed pojo.UsdtRechargeTx
				observedErr := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("tx_id = ?", proofs[i].TxID).First(&observed).Error
				if observedErr == nil {
					if observed.MatchedOrderNo != nil && strings.TrimSpace(*observed.MatchedOrderNo) != "" {
						return errors.New("manual_proof_tx_already_matched")
					}
					if observed.Network != proofs[i].SourceNetwork || observed.Token != proofs[i].SourceToken ||
						observed.AmountDecimals != proofs[i].AmountDecimals || observed.AmountAtomic != proofs[i].AmountAtomic {
						return errors.New("manual_proof_conflicts_with_scanned_tx")
					}
					observedTxIDs = append(observedTxIDs, observed.ID)
				} else if !errors.Is(observedErr, gorm.ErrRecordNotFound) {
					return observedErr
				}
				if err := tx.Create(&proofs[i]).Error; err != nil {
					return err
				}
			}
			return nil
		},
		func(tx *gorm.DB, _ *pojo.RechargeOrder) error {
			cryptoUpdates := map[string]any{
				"status":  pojo.UsdtRechargeStatusPaid,
				"paid_at": now,
				"tx_id":   proofs[0].TxID,
			}
			if len(proofs) == 1 && proofs[0].SourceNetwork == cryptoOrder.Network && proofs[0].SourceToken == cryptoOrder.Token {
				paidAtomic := proofs[0].AmountAtomic
				cryptoUpdates["paid_amount_atomic"] = paidAtomic
				if cryptoOrder.Network == pojo.CryptoRechargeNetworkTRC20 && proofs[0].AmountDecimals == pojo.CryptoRechargeDecimalsUSDT {
					if parsed, ok := new(big.Int).SetString(paidAtomic, 10); ok && parsed.IsInt64() {
						cryptoUpdates["paid_amount_micro"] = parsed.Int64()
					}
				}
			}
			cryptoUpdate := tx.Model(&pojo.UsdtRechargeOrder{}).
				Where("id = ? AND status <> ?", cryptoOrder.ID, pojo.UsdtRechargeStatusPaid).
				Updates(cryptoUpdates)
			if cryptoUpdate.Error != nil {
				return cryptoUpdate.Error
			}
			if cryptoUpdate.RowsAffected != 1 {
				return errors.New("manual_supplement_crypto_order_update_failed")
			}
			for _, observedID := range observedTxIDs {
				matchedOrderNo := cryptoOrder.OrderNo
				observedUpdate := tx.Model(&pojo.UsdtRechargeTx{}).
					Where("id = ? AND matched_order_no IS NULL", observedID).
					Updates(map[string]any{
						"matched_order_no": matchedOrderNo,
						"matched_at":       now,
						"review_status":    pojo.CryptoRechargeTxReviewManualSupplemented,
					})
				if observedUpdate.Error != nil {
					return observedUpdate.Error
				}
				if observedUpdate.RowsAffected != 1 {
					return errors.New("manual_proof_tx_update_failed")
				}
			}
			completed := tx.Model(&pojo.CryptoRechargeManualSupplement{}).
				Where("id = ? AND status = ?", supplement.ID, pojo.CryptoRechargeManualStatusProcessing).
				Update("status", pojo.CryptoRechargeManualStatusCompleted)
			if completed.Error != nil {
				return completed.Error
			}
			if completed.RowsAffected != 1 {
				return errors.New("manual_supplement_status_update_failed")
			}
			return nil
		},
	)
	if err != nil {
		completed, found, completedErr := loadCompletedCryptoManualSupplement(db, req.OrderNo, req.Reason, proofs)
		if completedErr != nil {
			if errors.Is(completedErr, errCryptoManualSupplementCompletedConflict) {
				return result, completedErr
			}
			return result, errors.Join(err, completedErr)
		}
		if found {
			return completed, nil
		}
		return result, err
	}
	supplement.Status = pojo.CryptoRechargeManualStatusCompleted
	result.Supplement = supplement
	result.Proofs = proofs
	return result, nil
}

func normalizeCryptoManualProofs(input []pojo.CryptoRechargeManualProofReq) ([]pojo.CryptoRechargeManualProof, error) {
	if len(input) == 0 || len(input) > 20 {
		return nil, errors.New("manual_supplement_transactions_invalid")
	}
	result := make([]pojo.CryptoRechargeManualProof, 0, len(input))
	seen := make(map[string]struct{}, len(input))
	for _, item := range input {
		network := strings.ToUpper(strings.TrimSpace(item.SourceNetwork))
		token := strings.ToUpper(strings.TrimSpace(item.SourceToken))
		txID := strings.ToLower(strings.TrimSpace(item.TxID))
		amountAtomic := strings.TrimSpace(item.AmountAtomic)
		if network == "" || len(network) > 32 || token == "" || len(token) > 32 || txID == "" || len(txID) > 128 {
			return nil, errors.New("manual_proof_fields_invalid")
		}
		if item.AmountDecimals < 0 || item.AmountDecimals > pojo.CryptoRechargeDecimalsETH {
			return nil, errors.New("manual_proof_decimals_invalid")
		}
		parsedAmount, err := parseCryptoAtomicAmount(amountAtomic)
		if err != nil {
			return nil, errors.New("manual_proof_amount_invalid")
		}
		amountAtomic = parsedAmount.String()
		if len(amountAtomic) > 96 {
			return nil, errors.New("manual_proof_amount_invalid")
		}
		proofKey := txID
		if _, exists := seen[proofKey]; exists {
			return nil, errors.New("manual_proof_duplicate_tx")
		}
		seen[proofKey] = struct{}{}
		result = append(result, pojo.CryptoRechargeManualProof{
			SourceNetwork:  network,
			SourceToken:    token,
			TxID:           txID,
			AmountDecimals: item.AmountDecimals,
			AmountAtomic:   amountAtomic,
		})
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].TxID < result[j].TxID
	})
	return result, nil
}

func loadCompletedCryptoManualSupplement(
	db *gorm.DB,
	orderNo string,
	reason string,
	requestedProofs []pojo.CryptoRechargeManualProof,
) (pojo.CryptoRechargeManualSupplementBack, bool, error) {
	var result pojo.CryptoRechargeManualSupplementBack
	var supplement pojo.CryptoRechargeManualSupplement
	err := db.Clauses(dbresolver.Write).
		Where("order_no = ? AND status = ?", orderNo, pojo.CryptoRechargeManualStatusCompleted).
		First(&supplement).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return result, false, nil
	}
	if err != nil {
		return result, false, err
	}

	proofs := make([]pojo.CryptoRechargeManualProof, 0, len(requestedProofs))
	if err := db.Clauses(dbresolver.Write).
		Where("supplement_id = ?", supplement.ID).
		Order("tx_id ASC").
		Find(&proofs).Error; err != nil {
		return result, true, err
	}
	sort.Slice(proofs, func(i, j int) bool {
		return strings.ToLower(strings.TrimSpace(proofs[i].TxID)) < strings.ToLower(strings.TrimSpace(proofs[j].TxID))
	})
	result.Supplement = supplement
	result.Proofs = proofs
	if !sameCryptoManualSupplementPayload(supplement, proofs, reason, requestedProofs) {
		return result, true, errCryptoManualSupplementCompletedConflict
	}
	return result, true, nil
}

func sameCryptoManualSupplementPayload(
	supplement pojo.CryptoRechargeManualSupplement,
	storedProofs []pojo.CryptoRechargeManualProof,
	reason string,
	requestedProofs []pojo.CryptoRechargeManualProof,
) bool {
	if strings.TrimSpace(supplement.Reason) != reason || len(storedProofs) != len(requestedProofs) {
		return false
	}
	for i := range requestedProofs {
		stored := storedProofs[i]
		requested := requestedProofs[i]
		storedAmount, err := parseCryptoAtomicAmount(stored.AmountAtomic)
		if err != nil ||
			strings.ToUpper(strings.TrimSpace(stored.SourceNetwork)) != requested.SourceNetwork ||
			strings.ToUpper(strings.TrimSpace(stored.SourceToken)) != requested.SourceToken ||
			strings.ToLower(strings.TrimSpace(stored.TxID)) != requested.TxID ||
			stored.AmountDecimals != requested.AmountDecimals ||
			storedAmount.String() != requested.AmountAtomic {
			return false
		}
	}
	return true
}
