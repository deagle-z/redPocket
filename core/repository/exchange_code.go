package repository

import (
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/utils"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"regexp"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var exchangeCodePattern = regexp.MustCompile(`^\d{6}$`)

func GetExchangeCodes(db *gorm.DB, search pojo.ExchangeCodeSearch) (pojo.ExchangeCodePage, error) {
	var result pojo.ExchangeCodePage
	query := db.Model(&pojo.ExchangeCode{})

	if code := strings.TrimSpace(search.Code); code != "" {
		query = query.Where("code LIKE ?", "%"+code+"%")
	}
	if search.Status != nil {
		query = query.Where("status = ?", *search.Status)
	} else {
		query = query.Where("status <> ?", -1)
	}

	if err := query.Count(&result.Total).Error; err != nil {
		return result, err
	}

	var list []pojo.ExchangeCode
	if err := query.Order("id desc").
		Limit(search.PageSize).
		Offset(search.PageSize * search.CurrentPage).
		Find(&list).Error; err != nil {
		return result, err
	}

	result.List = make([]pojo.ExchangeCodeBack, 0, len(list))
	for _, item := range list {
		result.List = append(result.List, exchangeCodeToBack(item))
	}
	result.PageSize = search.PageSize
	result.CurrentPage = search.CurrentPage
	return result, nil
}

func GetExchangeCodeByID(db *gorm.DB, id int64) (pojo.ExchangeCodeBack, error) {
	var entity pojo.ExchangeCode
	if err := db.Where("id = ? AND status <> ?", id, -1).First(&entity).Error; err != nil {
		return pojo.ExchangeCodeBack{}, errors.New("record_not_found")
	}
	return exchangeCodeToBack(entity), nil
}

func SetExchangeCode(db *gorm.DB, currentUser pojo.SysUser, req pojo.ExchangeCodeSet) (pojo.ExchangeCodeSetResult, error) {
	req.Code = normalizeExchangeCode(req.Code)
	req.Remark = normalizeExchangeRemark(req.Remark)

	if req.ID > 0 {
		back, err := updateExchangeCode(db, req)
		if err != nil {
			return pojo.ExchangeCodeSetResult{}, err
		}
		return pojo.ExchangeCodeSetResult{ExchangeCodeBack: back}, nil
	}
	return createExchangeCode(db, currentUser, req)
}

func DelExchangeCode(db *gorm.DB, id int64) error {
	var entity pojo.ExchangeCode
	if err := db.Where("id = ? AND status <> ?", id, -1).First(&entity).Error; err != nil {
		return errors.New("record_not_found_delete")
	}
	return db.Model(&pojo.ExchangeCode{}).Where("id = ?", id).Update("status", -1).Error
}

// BatchDelExchangeCode 批量软删除兑换码，返回实际删除数量。
func BatchDelExchangeCode(db *gorm.DB, ids []int64) (int64, error) {
	if len(ids) == 0 {
		return 0, errors.New("invalid_params")
	}
	result := db.Model(&pojo.ExchangeCode{}).
		Where("id IN ? AND status <> ?", ids, -1).
		Update("status", -1)
	if result.Error != nil {
		return 0, result.Error
	}
	return result.RowsAffected, nil
}

func RedeemExchangeCode(db *gorm.DB, tenantID int64, userID int64, code string) (pojo.ExchangeCodeRedeemBack, error) {
	code = normalizeExchangeCode(code)
	if !exchangeCodePattern.MatchString(code) {
		return pojo.ExchangeCodeRedeemBack{}, errors.New("exchange_code_format_error")
	}
	if userID <= 0 {
		return pojo.ExchangeCodeRedeemBack{}, errors.New("user_not_found")
	}

	lockKey := fmt.Sprintf("exchange_code_redeem:%d:%s", userID, code)
	if utils.RD != nil {
		acquired, lockErr := utils.AcquireLock(lockKey, 5*time.Second)
		if lockErr != nil || !acquired {
			return pojo.ExchangeCodeRedeemBack{}, errors.New("operation_too_frequent")
		}
		defer func() {
			_ = utils.ReleaseLock(lockKey)
		}()
	}

	var result pojo.ExchangeCodeRedeemBack
	now := time.Now()
	err := db.Transaction(func(tx *gorm.DB) error {
		var exchange pojo.ExchangeCode
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("code = ? AND status <> ?", code, -1).
			First(&exchange).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("exchange_code_not_found")
			}
			return err
		}
		if exchange.Status != 1 {
			return errors.New("exchange_code_disabled")
		}
		if exchange.RedeemCount >= exchange.MaxRedeemCount {
			return errors.New("exchange_code_redeemed_out")
		}

		var user pojo.TgUser
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ?", userID).
			First(&user).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("user_not_found")
			}
			return err
		}
		if user.Status != 1 {
			return errors.New("user_disabled_contact_admin")
		}
		if tenantID > 0 && user.TenantId != tenantID {
			return errors.New("user_not_found")
		}

		redeemed, err := hasUserRedeemedExchangeCode(tx, exchange.ID, userID)
		if err != nil {
			return err
		}
		if redeemed {
			return errors.New("exchange_code_already_redeemed")
		}

		// 同一批次：一个用户只能兑换一次
		if exchange.BatchNo != "" {
			batchRedeemed, err := hasUserRedeemedExchangeBatch(tx, exchange.BatchNo, userID)
			if err != nil {
				return err
			}
			if batchRedeemed {
				return errors.New("exchange_code_batch_already_redeemed")
			}
		}

		amount := utils.Truncate2(exchange.Amount)
		if amount <= 0 {
			return errors.New("exchange_code_amount_invalid")
		}
		beforeBalance := utils.Truncate2(user.Balance)
		afterBalance := utils.Truncate2(beforeBalance + amount)

		// 仅入账余额；提现限制走 v2 流水批次（不使用 v1 的 gift_amount/限提余额机制）
		if err := tx.Model(&pojo.TgUser{}).Where("id = ?", userID).Updates(map[string]any{
			"balance": gorm.Expr("balance + ?", amount),
		}).Error; err != nil {
			return err
		}

		history := pojo.CashHistory{
			UserId:          userID,
			AwardUni:        buildExchangeCodeAwardUni(exchange.ID, userID),
			Amount:          amount,
			StartAmount:     beforeBalance,
			EndAmount:       afterBalance,
			CashMark:        "兑换码赠送",
			CashDesc:        fmt.Sprintf("兑换码%s赠送%.2f金币", exchange.Code, amount),
			Type:            pojo.CashHistoryTypeExchangeCodeGift,
			IsGift:          1,
			FromUserId:      0,
			SourceChannelID: user.SourceChannelID,
		}
		if err := tx.Create(&history).Error; err != nil {
			return err
		}

		record := pojo.ExchangeCodeRedeem{
			CodeID:          exchange.ID,
			Code:            exchange.Code,
			BatchNo:         exchange.BatchNo,
			UserID:          userID,
			TenantID:        tenantID,
			Amount:          amount,
			BeforeBalance:   beforeBalance,
			AfterBalance:    afterBalance,
			CashHistoryID:   history.ID,
			SourceChannelID: user.SourceChannelID,
		}
		if err := tx.Create(&record).Error; err != nil {
			return err
		}

		// v2 提现流水批次：兑换所得需完成 amount × 赠送倍数 的流水后方可提现
		if err := EnsureWithdrawFlowBatchForGift(
			tx, user,
			pojo.WithdrawFlowBatchSourceExchangeCode,
			record.ID,
			fmt.Sprintf("exchange_code_%d", record.ID),
			pojo.WithdrawFlowBatchSourceExchangeCode,
			amount,
		); err != nil {
			return err
		}

		update := tx.Model(&pojo.ExchangeCode{}).
			Where("id = ? AND redeem_count < max_redeem_count", exchange.ID).
			UpdateColumn("redeem_count", gorm.Expr("redeem_count + ?", 1))
		if update.Error != nil {
			return update.Error
		}
		if update.RowsAffected == 0 {
			return errors.New("exchange_code_redeemed_out")
		}

		result = pojo.ExchangeCodeRedeemBack{
			Code:       exchange.Code,
			Amount:     amount,
			Balance:    afterBalance,
			RedeemedAt: now,
		}
		return nil
	})
	return result, err
}

const exchangeCodeGenerateCountMax = 500

func createExchangeCode(db *gorm.DB, currentUser pojo.SysUser, req pojo.ExchangeCodeSet) (pojo.ExchangeCodeSetResult, error) {
	amount := utils.Truncate2(req.Amount)
	if amount <= 0 {
		return pojo.ExchangeCodeSetResult{}, errors.New("exchange_code_amount_invalid")
	}
	if req.MaxRedeemCount <= 0 {
		return pojo.ExchangeCodeSetResult{}, errors.New("exchange_code_max_count_invalid")
	}
	status := int8(1)
	if req.Status != nil {
		if !isEditableExchangeCodeStatus(*req.Status) {
			return pojo.ExchangeCodeSetResult{}, errors.New("invalid_status")
		}
		status = *req.Status
	}

	// 指定了固定兑换码：只能创建一个，不归入批次
	if req.Code != "" {
		if !exchangeCodePattern.MatchString(req.Code) {
			return pojo.ExchangeCodeSetResult{}, errors.New("exchange_code_format_error")
		}
		back, err := createSingleExchangeCode(db, currentUser, req.Code, amount, req.MaxRedeemCount, status, req.Remark, "")
		if err != nil {
			return pojo.ExchangeCodeSetResult{}, err
		}
		return pojo.ExchangeCodeSetResult{ExchangeCodeBack: back, Codes: []string{back.Code}}, nil
	}

	// 随机生成：支持批量，同一次生成共用一个批次号（同批次一个用户仅可兑换一次）
	count := req.GenerateCount
	if count <= 0 {
		count = 1
	}
	if count > exchangeCodeGenerateCountMax {
		return pojo.ExchangeCodeSetResult{}, errors.New("exchange_code_generate_count_invalid")
	}

	batchNo := generateExchangeBatchNo()
	var first pojo.ExchangeCodeBack
	codes := make([]string, 0, count)
	for i := 0; i < count; i++ {
		code, err := generateExchangeCode(db)
		if err != nil {
			return pojo.ExchangeCodeSetResult{}, err
		}
		back, err := createSingleExchangeCode(db, currentUser, code, amount, req.MaxRedeemCount, status, req.Remark, batchNo)
		if err != nil {
			return pojo.ExchangeCodeSetResult{}, err
		}
		if i == 0 {
			first = back
		}
		codes = append(codes, back.Code)
	}
	return pojo.ExchangeCodeSetResult{ExchangeCodeBack: first, Codes: codes}, nil
}

func generateExchangeBatchNo() string {
	return fmt.Sprintf("B%d%s", time.Now().UnixNano(), strings.ToUpper(utils.RandomString(4)))
}

func createSingleExchangeCode(db *gorm.DB, currentUser pojo.SysUser, code string, amount float64, maxRedeemCount int, status int8, remark string, batchNo string) (pojo.ExchangeCodeBack, error) {
	var existing int64
	if err := db.Model(&pojo.ExchangeCode{}).Where("code = ?", code).Count(&existing).Error; err != nil {
		return pojo.ExchangeCodeBack{}, err
	}
	if existing > 0 {
		return pojo.ExchangeCodeBack{}, errors.New("exchange_code_exists")
	}

	entity := pojo.ExchangeCode{
		Code:           code,
		Amount:         amount,
		MaxRedeemCount: maxRedeemCount,
		BatchNo:        batchNo,
		Status:         status,
		Remark:         remark,
		CreatedBy:      currentUser.ID,
	}
	if err := db.Create(&entity).Error; err != nil {
		return pojo.ExchangeCodeBack{}, err
	}
	return exchangeCodeToBack(entity), nil
}

func updateExchangeCode(db *gorm.DB, req pojo.ExchangeCodeSet) (pojo.ExchangeCodeBack, error) {
	var entity pojo.ExchangeCode
	if err := db.Where("id = ? AND status <> ?", req.ID, -1).First(&entity).Error; err != nil {
		return pojo.ExchangeCodeBack{}, errors.New("record_not_found_update")
	}

	updates := map[string]any{
		"remark": req.Remark,
	}
	if req.Status != nil {
		if !isEditableExchangeCodeStatus(*req.Status) {
			return pojo.ExchangeCodeBack{}, errors.New("invalid_status")
		}
		updates["status"] = *req.Status
	}
	if err := db.Model(&pojo.ExchangeCode{}).Where("id = ?", entity.ID).Updates(updates).Error; err != nil {
		return pojo.ExchangeCodeBack{}, err
	}
	if err := db.Where("id = ?", entity.ID).First(&entity).Error; err != nil {
		return pojo.ExchangeCodeBack{}, err
	}
	return exchangeCodeToBack(entity), nil
}

func generateExchangeCode(db *gorm.DB) (string, error) {
	for i := 0; i < 50; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(1000000))
		if err != nil {
			return "", err
		}
		code := fmt.Sprintf("%06d", n.Int64())
		var count int64
		if err := db.Model(&pojo.ExchangeCode{}).Where("code = ?", code).Count(&count).Error; err != nil {
			return "", err
		}
		if count == 0 {
			return code, nil
		}
	}
	return "", errors.New("exchange_code_generate_failed")
}

func hasUserRedeemedExchangeCode(tx *gorm.DB, codeID int64, userID int64) (bool, error) {
	var count int64
	err := tx.Model(&pojo.ExchangeCodeRedeem{}).
		Where("code_id = ? AND user_id = ?", codeID, userID).
		Count(&count).Error
	return count > 0, err
}

func hasUserRedeemedExchangeBatch(tx *gorm.DB, batchNo string, userID int64) (bool, error) {
	if batchNo == "" {
		return false, nil
	}
	var count int64
	err := tx.Model(&pojo.ExchangeCodeRedeem{}).
		Where("batch_no = ? AND user_id = ?", batchNo, userID).
		Count(&count).Error
	return count > 0, err
}

func exchangeCodeToBack(entity pojo.ExchangeCode) pojo.ExchangeCodeBack {
	return pojo.ExchangeCodeBack{
		ID:             entity.ID,
		CreatedAt:      entity.CreatedAt,
		UpdatedAt:      entity.UpdatedAt,
		Code:           entity.Code,
		Amount:         utils.Truncate2(entity.Amount),
		MaxRedeemCount: entity.MaxRedeemCount,
		RedeemCount:    entity.RedeemCount,
		BatchNo:        entity.BatchNo,
		Status:         entity.Status,
		Remark:         entity.Remark,
		CreatedBy:      entity.CreatedBy,
	}
}

func normalizeExchangeCode(code string) string {
	return strings.TrimSpace(code)
}

func normalizeExchangeRemark(remark string) string {
	remark = strings.TrimSpace(remark)
	runes := []rune(remark)
	if len(runes) > 255 {
		return string(runes[:255])
	}
	return remark
}

func isEditableExchangeCodeStatus(status int8) bool {
	return status == 0 || status == 1
}

func buildExchangeCodeAwardUni(codeID int64, userID int64) string {
	return fmt.Sprintf("exchange_code_%d_%d", codeID, userID)
}
