package repository

import (
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/utils"
	"errors"
	"fmt"
	"github.com/jinzhu/copier"
	"gorm.io/gorm"
	"math/rand/v2"
	"strings"
	"time"

	"gorm.io/gorm/clause"
)

// GetTgUsers Telegram用户列表（分页）
func GetTgUsers(db *gorm.DB, search pojo.TgUserSearch) (result pojo.TgUserAdminResp) {
	var users []pojo.TgUser
	query := db.Model(&pojo.TgUser{})

	if search.TgID > 0 {
		query = query.Where("tg_id = ?", search.TgID)
	}
	if uid := strings.TrimSpace(search.Uid); uid != "" {
		query = query.Where("uid = ?", uid)
	}
	if search.Username != "" {
		query = query.Where("username like ?", "%"+search.Username+"%")
	}
	if search.FirstName != "" {
		query = query.Where("first_name like ?", "%"+search.FirstName+"%")
	}
	if search.Phone != "" {
		query = query.Where("phone like ?", "%"+search.Phone+"%")
	}
	if search.Country != "" {
		query = query.Where("country = ?", search.Country)
	}
	if search.Ip != "" {
		query = query.Where("ip = ?", strings.TrimSpace(search.Ip))
	}
	if search.Region != "" {
		query = query.Where("region = ?", strings.TrimSpace(strings.ToUpper(search.Region)))
	}
	if search.IsBot != nil {
		query = query.Where("is_bot = ?", *search.IsBot)
	}
	if search.Status != nil {
		query = query.Where("status = ?", *search.Status)
	}
	if search.ParentID != nil {
		query = query.Where("parent_id = ?", *search.ParentID)
	}
	if search.ParentUid != "" {
		query = query.Where("parent_id IN (?)", db.Model(&pojo.TgUser{}).
			Select("id").
			Where("uid = ?", strings.TrimSpace(search.ParentUid)))
	}
	if search.InviteCode != "" {
		query = query.Where("invite_code = ?", search.InviteCode)
	}

	query.Count(&result.Total)
	query = query.Order("id desc").Limit(search.PageSize).Offset(search.PageSize * search.CurrentPage)
	query.Find(&users)

	for _, user := range users {
		var temp pojo.TgUserAdminBack
		_ = copier.Copy(&temp, &user)
		result.List = append(result.List, temp)
	}
	fillAdminTgUserParentUIDs(db, result.List)
	fillAdminTgUserTenantNames(db, result.List)

	result.PageSize = search.PageSize
	result.CurrentPage = search.CurrentPage
	return result
}

// SetTgUser 创建或更新Telegram用户
func SetTgUser(db *gorm.DB, req pojo.TgUserSet) (result pojo.TgUserAdminBack, err error) {
	var dbUser pojo.TgUser
	if req.ID > 0 {
		db.Where("id = ?", req.ID).First(&dbUser)
		if dbUser.ID == 0 {
			return result, errors.New("record_not_found_update")
		}
		updates := buildTgUserUpdateMap(req)
		if len(updates) > 0 {
			err = db.Model(&pojo.TgUser{}).Where("id = ?", dbUser.ID).Updates(updates).Error
		}
		if err == nil {
			db.Where("id = ?", dbUser.ID).First(&dbUser)
		}
	} else {
		_ = copier.Copy(&dbUser, &req)
		if req.RebateRate != nil {
			dbUser.RebateRate = utils.Truncate2(*req.RebateRate)
		} else {
			dbUser.RebateRate = getDefaultInviteLuckyRebateRate(db)
		}
		if req.FreeLotteryCount != nil {
			dbUser.FreeLotteryCount = *req.FreeLotteryCount
		}
		err = db.Create(&dbUser).Error
	}
	if err != nil {
		return result, err
	}
	_ = copier.Copy(&result, &dbUser)
	return result, nil
}

func buildTgUserUpdateMap(req pojo.TgUserSet) map[string]any {
	updates := map[string]any{
		"username":            req.Username,
		"first_name":          req.FirstName,
		"avatar":              req.Avatar,
		"phone":               req.Phone,
		"country":             req.Country,
		"ip":                  req.Ip,
		"region":              req.Region,
		"remark":              req.Remark,
		"is_bot":              req.IsBot,
		"tg_id":               req.TgID,
		"status":              req.Status,
		"parent_id":           req.ParentID,
		"invite_code":         req.InviteCode,
		"source_channel_id":   req.SourceChannelID,
		"source_channel_code": req.SourceChannelCode,
		"tenant_id":           req.TenantId,
	}
	if req.RebateRate != nil {
		updates["rebate_rate"] = utils.Truncate2(*req.RebateRate)
	}
	if req.FreeLotteryCount != nil {
		updates["free_lottery_count"] = *req.FreeLotteryCount
	}
	return updates
}

// DelTgUser 删除Telegram用户
func DelTgUser(db *gorm.DB, id int64) (result string, err error) {
	var dbUser pojo.TgUser
	db.Where("id = ?", id).First(&dbUser)
	if dbUser.ID == 0 {
		return result, errors.New("record_not_found_delete")
	}
	err = db.Delete(&dbUser).Error
	if err != nil {
		return result, err
	}
	return "success", nil
}

// GetTgUserById 根据ID获取Telegram用户
func GetTgUserById(db *gorm.DB, id int64) (result pojo.TgUserAdminBack, err error) {
	var dbUser pojo.TgUser
	db.Where("id = ?", id).First(&dbUser)
	if dbUser.ID == 0 {
		return result, errors.New("record_not_found")
	}
	_ = copier.Copy(&result, &dbUser)
	return result, nil
}

// SetTgUserStatus 更新Telegram用户状态
func SetTgUserStatus(db *gorm.DB, id int64, status int8) (result pojo.TgUserAdminBack, err error) {
	var dbUser pojo.TgUser
	db.Where("id = ?", id).First(&dbUser)
	if dbUser.ID == 0 {
		return result, errors.New("record_not_found")
	}
	err = db.Model(&dbUser).Update("status", status).Error
	if err != nil {
		return result, err
	}
	_ = copier.Copy(&result, &dbUser)
	result.Status = status
	return result, nil
}

func SetTgUserRebateRate(db *gorm.DB, id int64, rebateRate float64) (result pojo.TgUserAdminBack, err error) {
	var dbUser pojo.TgUser
	db.Where("id = ?", id).First(&dbUser)
	if dbUser.ID == 0 {
		return result, errors.New("record_not_found")
	}
	rebateRate = utils.Truncate2(rebateRate)
	err = db.Model(&dbUser).Update("rebate_rate", rebateRate).Error
	if err != nil {
		return result, err
	}
	_ = copier.Copy(&result, &dbUser)
	result.RebateRate = rebateRate
	return result, nil
}

func AddTgUserRebateAmount(db *gorm.DB, id int64, amount float64) (result pojo.TgUserAdminBack, err error) {
	amount = utils.Truncate2(amount)
	if amount <= 0 {
		return result, errors.New("invalid_amount")
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		var dbUser pojo.TgUser
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", id).First(&dbUser).Error; err != nil {
			return errors.New("record_not_found")
		}
		startAmount := utils.Truncate2(dbUser.RebateAmount)
		endAmount := utils.Truncate2(startAmount + amount)
		if err := tx.Model(&pojo.TgUser{}).Where("id = ?", dbUser.ID).Updates(map[string]any{
			"rebate_amount":       gorm.Expr("rebate_amount + ?", amount),
			"rebate_total_amount": gorm.Expr("rebate_total_amount + ?", amount),
		}).Error; err != nil {
			return err
		}
		if err := tx.Create(&pojo.CashHistory{
			UserId:          dbUser.ID,
			AwardUni:        fmt.Sprintf("admin_manual_rebate_%d_%d", dbUser.ID, time.Now().UnixNano()),
			Amount:          amount,
			StartAmount:     startAmount,
			EndAmount:       endAmount,
			CashMark:        "后台加佣金",
			CashDesc:        fmt.Sprintf("后台手工增加佣金%.2f", amount),
			Type:            pojo.CashHistoryTypeAdminManualRebate,
			IsGift:          0,
			FromUserId:      0,
			SourceChannelID: dbUser.SourceChannelID,
		}).Error; err != nil {
			return err
		}
		if err := tx.Where("id = ?", dbUser.ID).First(&dbUser).Error; err != nil {
			return err
		}
		_ = copier.Copy(&result, &dbUser)
		return nil
	})
	return result, err
}

func SetTgUserRemark(db *gorm.DB, id int64, remark string) (result pojo.TgUserAdminBack, err error) {
	var dbUser pojo.TgUser
	db.Where("id = ?", id).First(&dbUser)
	if dbUser.ID == 0 {
		return result, errors.New("record_not_found")
	}
	remark = strings.TrimSpace(remark)
	if len([]rune(remark)) > 255 {
		return result, errors.New("remark_too_long")
	}
	var remarkPtr *string
	if remark != "" {
		remarkPtr = &remark
	}
	err = db.Model(&dbUser).Update("remark", remarkPtr).Error
	if err != nil {
		return result, err
	}
	_ = copier.Copy(&result, &dbUser)
	result.Remark = remarkPtr
	return result, nil
}

func BatchCreateBotTgUsers(db *gorm.DB, req pojo.TgUserBatchCreateBotReq) (result pojo.TgUserBatchCreateBotResp, err error) {
	names := parseBotNames(req.NameFile)
	if !req.RandomName && len(names) == 0 {
		return result, errors.New("name_file_required")
	}
	if req.Num <= 0 {
		return result, errors.New("num_must_gt_zero")
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		baseTgID, tgIDErr := nextBotBaseTgID(tx)
		if tgIDErr != nil {
			return tgIDErr
		}

		result.List = make([]pojo.TgUserAdminBack, 0, req.Num)
		for i := 0; i < req.Num; i++ {
			inviteCode, inviteErr := generateInviteCode(tx)
			if inviteErr != nil {
				return inviteErr
			}
			uid, uidErr := generateUniqueUID(tx)
			if uidErr != nil {
				return uidErr
			}

			displayName := buildBotDisplayName(req.RandomName, names, i)
			username := buildBotUsername(displayName, uid, i)
			avatar := pickBotAvatar(req.AvatarLinks, i)
			tgID := baseTgID + int64(i)

			user := pojo.TgUser{
				Uid:        uid,
				Username:   &username,
				FirstName:  &displayName,
				Avatar:     avatar,
				IsBot:      true,
				TgID:       tgID,
				Balance:    999999,
				GiftAmount: 0,
				GiftTotal:  0,
				RebateRate: getDefaultInviteLuckyRebateRate(tx),
				Status:     1,
				InviteCode: &inviteCode,
				TenantId:   0,
			}
			if err := tx.Create(&user).Error; err != nil {
				return err
			}

			var temp pojo.TgUserAdminBack
			_ = copier.Copy(&temp, &user)
			result.List = append(result.List, temp)
		}
		result.Count = len(result.List)
		return nil
	})

	return result, err
}

func BatchUpdateBotTgUsers(db *gorm.DB, req pojo.TgUserBatchUpdateBotReq) (result pojo.TgUserBatchUpdateBotResp, err error) {
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
	if !shouldUpdateName && !shouldUpdateAvatar && !shouldUpdateStatus {
		return result, errors.New("no_fields_to_update")
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		var bots []pojo.TgUser
		if err := tx.Where("id IN ? AND is_bot = ?", ids, true).Find(&bots).Error; err != nil {
			return err
		}
		if len(bots) != len(ids) {
			return errors.New("bot_user_not_found")
		}

		botMap := make(map[int64]pojo.TgUser, len(bots))
		for _, bot := range bots {
			botMap[bot.ID] = bot
		}

		result.List = make([]pojo.TgUserAdminBack, 0, len(ids))
		for index, id := range ids {
			bot := botMap[id]
			updates := map[string]any{}

			if shouldUpdateName {
				displayName := buildBotDisplayName(req.RandomName, names, index)
				username := buildBotUsername(displayName, bot.Uid, index)
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

			if len(updates) == 0 {
				continue
			}
			if err := tx.Model(&pojo.TgUser{}).Where("id = ? AND is_bot = ?", id, true).Updates(updates).Error; err != nil {
				return err
			}

			var temp pojo.TgUserAdminBack
			_ = copier.Copy(&temp, &bot)
			result.List = append(result.List, temp)
		}
		result.Count = len(result.List)
		return nil
	})

	return result, err
}

func parseBotNames(nameFile string) []string {
	lines := strings.Split(strings.ReplaceAll(nameFile, "\r\n", "\n"), "\n")
	result := make([]string, 0, len(lines))
	for _, line := range lines {
		name := strings.TrimSpace(line)
		if name == "" {
			continue
		}
		result = append(result, name)
	}
	return result
}

func buildBotDisplayName(randomName bool, names []string, index int) string {
	if !randomName && len(names) > 0 {
		return names[index%len(names)]
	}
	return buildHumanLikeName()
}

func buildBotUsername(displayName string, uid string, index int) string {
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
	return fmt.Sprintf("%s%d", cleanName, 100+((index+int(uid[0]))%9000))
}

func pickBotAvatar(avatarLinks []string, index int) *string {
	if len(avatarLinks) == 0 {
		return nil
	}
	avatar := strings.TrimSpace(avatarLinks[index%len(avatarLinks)])
	if avatar == "" {
		return nil
	}
	return &avatar
}

func nextBotBaseTgID(db *gorm.DB) (int64, error) {
	var maxTgID int64
	if err := db.Model(&pojo.TgUser{}).Select("COALESCE(MAX(tg_id), 0)").Scan(&maxTgID).Error; err != nil {
		return 0, err
	}

	base := time.Now().UnixMilli()*100 + int64(rand.IntN(100))
	if maxTgID >= base {
		base = maxTgID + 1
	}
	return base, nil
}

func buildHumanLikeName() string {
	firstNames := []string{
		"Lucas", "Mateus", "Gabriel", "Rafael", "Bruno", "Felipe", "Daniel", "Victor",
		"Marcos", "Diego", "Pedro", "Gustavo", "Thiago", "Andre", "Caio", "Enzo",
		"Ana", "Julia", "Mariana", "Beatriz", "Larissa", "Camila", "Amanda", "Sofia",
		"Isabela", "Fernanda", "Leticia", "Bianca", "Helena", "Yasmin", "Clara", "Alice",
	}
	lastNames := []string{
		"Silva", "Santos", "Oliveira", "Souza", "Lima", "Costa", "Pereira", "Rodrigues",
		"Almeida", "Nascimento", "Araujo", "Fernandes", "Carvalho", "Gomes", "Martins", "Rocha",
	}
	firstName := firstNames[rand.IntN(len(firstNames))]
	lastName := lastNames[rand.IntN(len(lastNames))]
	if rand.IntN(100) < 35 {
		return firstName
	}
	return fmt.Sprintf("%s %s", firstName, lastName)
}

func uniqueInt64s(values []int64) []int64 {
	result := make([]int64, 0, len(values))
	seen := make(map[int64]struct{}, len(values))
	for _, value := range values {
		if value <= 0 {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

func fillAdminTgUserParentUIDs(db *gorm.DB, users []pojo.TgUserAdminBack) {
	parentIDs := make([]int64, 0, len(users))
	for _, user := range users {
		if user.ParentID != nil {
			parentIDs = append(parentIDs, *user.ParentID)
		}
	}
	parentIDs = uniqueInt64s(parentIDs)
	if len(parentIDs) == 0 {
		return
	}

	var parents []pojo.TgUser
	_ = db.Model(&pojo.TgUser{}).
		Select("id, uid").
		Where("id IN ?", parentIDs).
		Find(&parents).Error

	parentUIDMap := make(map[int64]string, len(parents))
	for _, parent := range parents {
		parentUIDMap[parent.ID] = parent.Uid
	}

	for i := range users {
		if users[i].ParentID == nil {
			continue
		}
		if uid, ok := parentUIDMap[*users[i].ParentID]; ok && uid != "" {
			users[i].ParentUid = &uid
		}
	}
}

func fillAdminTgUserTenantNames(db *gorm.DB, users []pojo.TgUserAdminBack) {
	tenantIDs := make([]int64, 0, len(users))
	seen := make(map[int64]struct{}, len(users))
	for _, user := range users {
		if user.TenantId <= 0 {
			continue
		}
		if _, ok := seen[user.TenantId]; ok {
			continue
		}
		seen[user.TenantId] = struct{}{}
		tenantIDs = append(tenantIDs, user.TenantId)
	}
	if len(tenantIDs) == 0 {
		return
	}

	var tenants []pojo.SysTenant
	_ = db.Model(&pojo.SysTenant{}).
		Select("id, tenant_name").
		Where("id IN ?", tenantIDs).
		Find(&tenants).Error

	tenantNameMap := make(map[int64]string, len(tenants))
	for _, tenant := range tenants {
		tenantNameMap[tenant.ID] = tenant.TenantName
	}

	for i := range users {
		if name, ok := tenantNameMap[users[i].TenantId]; ok && name != "" {
			users[i].TenantName = &name
		}
	}
}
