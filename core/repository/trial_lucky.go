package repository

import (
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/utils"
	"errors"
	"fmt"
	"github.com/jinzhu/copier"
	"gorm.io/gorm"
	"strings"
	"time"
)

func GetTrialLuckyMoneyList(db *gorm.DB, search pojo.TrialLuckyMoneySearch) (result pojo.TrialLuckyMoneyResp) {
	var list []pojo.TrialLuckyMoney
	query := db.Model(&pojo.TrialLuckyMoney{})
	if search.SenderID > 0 {
		query = query.Where("sender_id = ?", search.SenderID)
	}
	if search.ChatID != 0 {
		query = query.Where("chat_id = ?", search.ChatID)
	}
	if search.Status != nil {
		query = query.Where("status = ?", *search.Status)
	}
	if search.TenantId > 0 {
		query = query.Where("tenant_id = ?", search.TenantId)
	}

	query.Count(&result.Total)
	query.Order("id desc").Limit(search.PageSize).Offset(search.PageSize * search.CurrentPage).Find(&list)
	for _, item := range list {
		var back pojo.TrialLuckyMoneyBack
		_ = copier.Copy(&back, &item)
		result.List = append(result.List, back)
	}
	result.PageSize = search.PageSize
	result.CurrentPage = search.CurrentPage
	return result
}

func GetTrialCashHistoryList(db *gorm.DB, search pojo.TrialCashHistorySearch) (result pojo.TrialCashHistoryPage) {
	result.PageSize = search.PageSize
	result.CurrentPage = search.CurrentPage
	return result
}

func GetTrialBotUsers(db *gorm.DB, search pojo.TrialBotUserSearch) (result pojo.TrialBotUserResp) {
	var list []pojo.TrialBotUser
	query := db.Model(&pojo.TrialBotUser{})
	if search.Username != "" {
		query = query.Where("username LIKE ?", "%"+strings.TrimSpace(search.Username)+"%")
	}
	if search.FirstName != "" {
		query = query.Where("first_name LIKE ?", "%"+strings.TrimSpace(search.FirstName)+"%")
	}
	if search.Status != nil {
		query = query.Where("status = ?", *search.Status)
	}
	if search.TenantId > 0 {
		query = query.Where("tenant_id = ?", search.TenantId)
	}

	query.Count(&result.Total)
	query.Order("id desc").Limit(search.PageSize).Offset(search.PageSize * search.CurrentPage).Find(&list)
	for _, item := range list {
		var back pojo.TrialBotUserBack
		_ = copier.Copy(&back, &item)
		result.List = append(result.List, back)
	}
	result.PageSize = search.PageSize
	result.CurrentPage = search.CurrentPage
	return result
}

func BatchCreateTrialBotUsers(db *gorm.DB, req pojo.TrialBotBatchCreateReq) (result pojo.TrialBotBatchResp, err error) {
	names := parseBotNames(req.NameFile)
	if !req.RandomName && len(names) == 0 {
		return result, errors.New("name_file_required")
	}
	if req.Num <= 0 {
		return result, errors.New("num_must_gt_zero")
	}
	balance := utils.Truncate2(req.Balance)
	if balance <= 0 {
		balance = 1000000
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		result.List = make([]pojo.TrialBotUserBack, 0, req.Num)
		for i := 0; i < req.Num; i++ {
			displayName := buildBotDisplayName(req.RandomName, names, i)
			username := buildTrialBotUsername(displayName, i)
			avatar := pickBotAvatar(req.AvatarLinks, i)
			bot := pojo.TrialBotUser{
				Username:  &username,
				FirstName: &displayName,
				Avatar:    avatar,
				Balance:   balance,
				Status:    1,
			}
			if err := tx.Create(&bot).Error; err != nil {
				return err
			}
			var back pojo.TrialBotUserBack
			_ = copier.Copy(&back, &bot)
			result.List = append(result.List, back)
		}
		result.Count = len(result.List)
		return nil
	})
	return result, err
}

func BatchUpdateTrialBotUsers(db *gorm.DB, req pojo.TrialBotBatchUpdateReq) (result pojo.TrialBotBatchResp, err error) {
	ids := uniqueInt64s(req.IDs)
	if len(ids) == 0 {
		return result, errors.New("ids_required")
	}
	if req.Status != nil && *req.Status != 1 && *req.Status != 0 && *req.Status != -1 {
		return result, errors.New("invalid_status")
	}
	names := parseBotNames(req.NameFile)
	shouldUpdateName := req.RandomName || len(names) > 0
	shouldUpdateAvatar := len(req.AvatarLinks) > 0
	shouldUpdateStatus := req.Status != nil
	shouldUpdateBalance := req.Balance != nil
	if !shouldUpdateName && !shouldUpdateAvatar && !shouldUpdateStatus && !shouldUpdateBalance {
		return result, errors.New("no_fields_to_update")
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		var bots []pojo.TrialBotUser
		if err := tx.Where("id IN ?", ids).Find(&bots).Error; err != nil {
			return err
		}
		if len(bots) != len(ids) {
			return errors.New("trial_bot_not_found")
		}
		botMap := make(map[int64]pojo.TrialBotUser, len(bots))
		for _, bot := range bots {
			botMap[bot.ID] = bot
		}
		result.List = make([]pojo.TrialBotUserBack, 0, len(ids))
		for index, id := range ids {
			bot := botMap[id]
			updates := map[string]any{}
			if shouldUpdateName {
				displayName := buildBotDisplayName(req.RandomName, names, index)
				username := buildTrialBotUsername(displayName, index)
				updates["first_name"] = displayName
				updates["username"] = username
				bot.FirstName = &displayName
				bot.Username = &username
			}
			if shouldUpdateAvatar {
				avatar := pickBotAvatar(req.AvatarLinks, index)
				if avatar == nil {
					updates["avatar"] = nil
				} else {
					updates["avatar"] = *avatar
				}
				bot.Avatar = avatar
			}
			if shouldUpdateStatus {
				updates["status"] = *req.Status
				bot.Status = *req.Status
			}
			if shouldUpdateBalance {
				balance := utils.Truncate2(*req.Balance)
				updates["balance"] = balance
				bot.Balance = balance
			}
			if err := tx.Model(&pojo.TrialBotUser{}).Where("id = ?", id).Updates(updates).Error; err != nil {
				return err
			}
			var back pojo.TrialBotUserBack
			_ = copier.Copy(&back, &bot)
			result.List = append(result.List, back)
		}
		result.Count = len(result.List)
		return nil
	})
	return result, err
}

func SetTrialBotUserStatus(db *gorm.DB, id int64, status int8) (pojo.TrialBotUserBack, error) {
	var result pojo.TrialBotUserBack
	if id <= 0 {
		return result, errors.New("invalid_params")
	}
	if status != 1 && status != 0 && status != -1 {
		return result, errors.New("invalid_status")
	}
	var bot pojo.TrialBotUser
	if err := db.Where("id = ?", id).First(&bot).Error; err != nil {
		return result, errors.New("trial_bot_not_found")
	}
	if err := db.Model(&bot).Update("status", status).Error; err != nil {
		return result, err
	}
	bot.Status = status
	_ = copier.Copy(&result, &bot)
	return result, nil
}

func DelTrialBotUser(db *gorm.DB, id int64) (string, error) {
	if id <= 0 {
		return "", errors.New("invalid_params")
	}
	var bot pojo.TrialBotUser
	if err := db.Where("id = ?", id).First(&bot).Error; err != nil {
		return "", errors.New("trial_bot_not_found")
	}
	if err := db.Delete(&bot).Error; err != nil {
		return "", err
	}
	return "success", nil
}

func buildTrialBotUsername(displayName string, index int) string {
	name := strings.ToLower(strings.TrimSpace(displayName))
	name = strings.ReplaceAll(name, " ", "_")
	var builder strings.Builder
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' {
			builder.WriteRune(r)
		}
	}
	cleanName := strings.Trim(builder.String(), "_")
	if cleanName == "" {
		cleanName = strings.ToLower(utils.RandomString(6))
	}
	return fmt.Sprintf("%s_trial_%d_%d", cleanName, time.Now().UnixMilli()%100000, index+1)
}
