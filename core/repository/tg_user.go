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
)

// GetTgUsers Telegram用户列表（分页）
func GetTgUsers(db *gorm.DB, search pojo.TgUserSearch) (result pojo.TgUserAdminResp) {
	var users []pojo.TgUser
	query := db.Model(&pojo.TgUser{})

	if search.TgID > 0 {
		query = query.Where("tg_id = ?", search.TgID)
	}
	if search.Username != "" {
		query = query.Where("username like ?", "%"+search.Username+"%")
	}
	if search.FirstName != "" {
		query = query.Where("first_name like ?", "%"+search.FirstName+"%")
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
			return result, errors.New("更新的数据不存在")
		}
		_ = copier.Copy(&dbUser, &req)
		err = db.Save(&dbUser).Error
	} else {
		_ = copier.Copy(&dbUser, &req)
		err = db.Create(&dbUser).Error
	}
	if err != nil {
		return result, err
	}
	_ = copier.Copy(&result, &dbUser)
	return result, nil
}

// DelTgUser 删除Telegram用户
func DelTgUser(db *gorm.DB, id int64) (result string, err error) {
	var dbUser pojo.TgUser
	db.Where("id = ?", id).First(&dbUser)
	if dbUser.ID == 0 {
		return result, errors.New("删除的数据不存在")
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
		return result, errors.New("数据不存在")
	}
	_ = copier.Copy(&result, &dbUser)
	return result, nil
}

// SetTgUserStatus 更新Telegram用户状态
func SetTgUserStatus(db *gorm.DB, id int64, status int8) (result pojo.TgUserAdminBack, err error) {
	var dbUser pojo.TgUser
	db.Where("id = ?", id).First(&dbUser)
	if dbUser.ID == 0 {
		return result, errors.New("数据不存在")
	}
	err = db.Model(&dbUser).Update("status", status).Error
	if err != nil {
		return result, err
	}
	_ = copier.Copy(&result, &dbUser)
	result.Status = status
	return result, nil
}

func BatchCreateBotTgUsers(db *gorm.DB, req pojo.TgUserBatchCreateBotReq) (result pojo.TgUserBatchCreateBotResp, err error) {
	names := parseBotNames(req.NameFile)
	if !req.RandomName && len(names) == 0 {
		return result, errors.New("名称文件不能为空")
	}
	if req.Num <= 0 {
		return result, errors.New("num 必须大于 0")
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
