package repository

import (
	"BaseGoUni/core/pojo"
	"errors"
	"fmt"
	"math/rand/v2"
	"strings"
	"time"

	"github.com/jinzhu/copier"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	RechargeGameUnlockEnabledConfigKey = "recharge_game_unlock_enabled"
	rechargeGameUnlockEnabledDefault   = true
	AppGameFakeOnlineMin               = 80
	AppGameFakeOnlineMax               = 200
)

// GetRechargeGameUnlockEnabled controls whether users must unlock games by recharging or transferring commission.
func GetRechargeGameUnlockEnabled(db *gorm.DB) bool {
	if db == nil {
		return rechargeGameUnlockEnabledDefault
	}

	var config pojo.SysConfig
	db.Where("config_key = ?", RechargeGameUnlockEnabledConfigKey).First(&config)
	if config.ID == 0 {
		_ = db.Clauses(clause.OnConflict{DoNothing: true}).Create(&pojo.SysConfig{
			ConfigKey:   RechargeGameUnlockEnabledConfigKey,
			ConfigValue: "1",
			ConfigDesc:  "充值解锁游戏开关：1=需充值满50或佣金转入，0=关闭充值限制",
		}).Error
		return rechargeGameUnlockEnabledDefault
	}

	enabled, ok := parseRechargeGameUnlockEnabled(config.ConfigValue)
	if !ok {
		return rechargeGameUnlockEnabledDefault
	}
	return enabled
}

func parseRechargeGameUnlockEnabled(raw string) (bool, bool) {
	switch strings.TrimSpace(raw) {
	case "1":
		return true, true
	case "0":
		return false, true
	default:
		return false, false
	}
}

func GetAppGames(db *gorm.DB, search pojo.AppGameSearch) (result pojo.AppGameResp) {
	var list []pojo.AppGame
	query := buildAppGameSearchQuery(db.Model(&pojo.AppGame{}), search)

	query.Count(&result.Total)
	query.Order("sort asc, show_index asc, game_id desc").
		Limit(search.PageSize).
		Offset(search.PageSize * search.CurrentPage).
		Find(&list)

	result.List = list
	result.PageSize = search.PageSize
	result.CurrentPage = search.CurrentPage
	return result
}

func GetAppGameHomeList(db *gorm.DB, search pojo.AppGameSearch) (result pojo.AppGameHomeListResp) {
	var list []pojo.AppGame
	query := buildAppGameSearchQuery(db.Model(&pojo.AppGame{}), search).
		Where("COALESCE(disabled_flag, 0) = 0")

	query.Count(&result.Total)
	query.Order("sort asc, show_index asc, game_id desc").
		Limit(search.PageSize).
		Offset(search.PageSize * search.CurrentPage).
		Find(&list)

	result.List = make([]pojo.AppGameHomeItem, 0, len(list))
	for _, item := range list {
		result.List = append(result.List, buildAppGameHomeItem(item))
	}
	result.PageSize = search.PageSize
	result.CurrentPage = search.CurrentPage
	return result
}

func GetAppGameListByCategoryCode(db *gorm.DB, search pojo.AppGameSearch) pojo.AppGameHomeListResp {
	search.CategoryCode = strings.ToLower(strings.TrimSpace(search.CategoryCode))
	search.ThirdGameCategory = ""
	return GetAppGameHomeList(db, search)
}

func GetAppGameCategoryList(db *gorm.DB, search pojo.AppGameSearch) pojo.AppGameCategoryListResp {
	categories := GetAppGameThirdCategories(db, search.CategoryCode).List
	if strings.TrimSpace(search.ThirdGameCategory) == "" && len(categories) > 0 {
		search.ThirdGameCategory = categories[0]
	}

	listResp := GetAppGameHomeList(db, search)
	return pojo.AppGameCategoryListResp{
		BasePageResponse:    listResp.BasePageResponse,
		ThirdGameCategories: categories,
		ThirdGameCategory:   strings.TrimSpace(search.ThirdGameCategory),
	}
}

func buildAppGameSearchQuery(query *gorm.DB, search pojo.AppGameSearch) *gorm.DB {

	if search.DeletedFlag == nil {
		query = query.Where("COALESCE(deleted_flag, 0) = 0")
	} else {
		query = query.Where("deleted_flag = ?", *search.DeletedFlag)
	}
	if search.GameID > 0 {
		query = query.Where("game_id = ?", search.GameID)
	}
	if gameName := strings.TrimSpace(search.GameName); gameName != "" {
		query = query.Where("game_name LIKE ?", "%"+gameName+"%")
	}
	if categoryCode := strings.TrimSpace(search.CategoryCode); categoryCode != "" {
		query = query.Where("category_code = ?", categoryCode)
	}
	if search.Type != nil {
		query = query.Where("type = ?", *search.Type)
	}
	if search.ParentID != nil {
		query = query.Where("parent_id = ?", *search.ParentID)
	}
	if platformCode := strings.TrimSpace(search.PlatformCode); platformCode != "" {
		query = query.Where("platform_code = ?", platformCode)
	}
	if thirdGameID := strings.TrimSpace(search.ThirdGameID); thirdGameID != "" {
		query = query.Where("third_game_id = ?", thirdGameID)
	}
	if thirdGameName := strings.TrimSpace(search.ThirdGameName); thirdGameName != "" {
		query = query.Where("third_game_name LIKE ?", "%"+thirdGameName+"%")
	}
	if thirdGameCategory := strings.TrimSpace(search.ThirdGameCategory); thirdGameCategory != "" {
		query = query.Where("third_game_category = ?", thirdGameCategory)
	}
	if search.TenantID != nil {
		query = query.Where("tenant_id = ?", *search.TenantID)
	}
	if search.Hot != nil {
		query = query.Where("hot = ?", *search.Hot)
	}
	if search.HomeShow != nil {
		query = query.Where("home_show = ?", *search.HomeShow)
	}
	if search.DisabledFlag != nil {
		query = query.Where("disabled_flag = ?", *search.DisabledFlag)
	}
	return query
}

func GetAppGameThirdCategories(db *gorm.DB, categoryCode string) pojo.AppGameThirdCategoryResp {
	var list []string
	query := db.Model(&pojo.AppGame{}).
		Where("COALESCE(deleted_flag, 0) = 0").
		Where("COALESCE(disabled_flag, 0) = 0").
		Where("third_game_category IS NOT NULL").
		Where("TRIM(third_game_category) <> ''")
	if categoryCode = strings.TrimSpace(categoryCode); categoryCode != "" {
		query = query.Where("category_code = ?", categoryCode)
	}
	query.Distinct("third_game_category").
		Order("third_game_category asc").
		Pluck("third_game_category", &list)
	return pojo.AppGameThirdCategoryResp{List: list}
}

func GetAppGameByID(db *gorm.DB, gameID int64) (pojo.AppGame, error) {
	var entity pojo.AppGame
	err := db.Where("game_id = ? AND COALESCE(deleted_flag, 0) = 0", gameID).First(&entity).Error
	if err != nil {
		return pojo.AppGame{}, err
	}
	return entity, nil
}

func GetEnabledAppGameByID(db *gorm.DB, gameID int64) (pojo.AppGame, error) {
	var entity pojo.AppGame
	err := db.Where("game_id = ? AND COALESCE(deleted_flag, 0) = 0 AND COALESCE(disabled_flag, 0) = 0", gameID).First(&entity).Error
	if err != nil {
		return pojo.AppGame{}, err
	}
	return entity, nil
}

func GetHomeAppGamesGrouped(db *gorm.DB) pojo.AppGameHomeResp {
	var games []pojo.AppGame
	var hotGames []pojo.AppGame
	db.Model(&pojo.AppGame{}).
		Where("COALESCE(deleted_flag, 0) = 0").
		Where("COALESCE(disabled_flag, 0) = 0").
		Where("COALESCE(home_show, 0) = 1").
		Order("category_code asc, type asc, sort asc, show_index asc, game_id desc").
		Find(&games)
	db.Model(&pojo.AppGame{}).
		Where("COALESCE(deleted_flag, 0) = 0").
		Where("COALESCE(disabled_flag, 0) = 0").
		Where("COALESCE(hot, 0) = 1").
		Order("sort asc, show_index asc, game_id desc").
		Find(&hotGames)

	result := pojo.AppGameHomeResp{
		"hot": make([]pojo.AppGameHomeItem, 0),
	}
	for _, item := range hotGames {
		result["hot"] = append(result["hot"], buildAppGameHomeItem(item))
	}
	for _, item := range games {
		homeItem := buildAppGameHomeItem(item)
		categoryCode := strings.TrimSpace(valueString(item.CategoryCode))
		if categoryCode == "" && item.Type != nil {
			categoryCode = fmt.Sprintf("type_%d", *item.Type)
		}
		if categoryCode == "" {
			categoryCode = "default"
		}
		if _, exists := result[categoryCode]; !exists {
			result[categoryCode] = make([]pojo.AppGameHomeItem, 0)
		}
		result[categoryCode] = append(result[categoryCode], homeItem)
	}
	return result
}

func buildAppGameHomeItem(item pojo.AppGame) pojo.AppGameHomeItem {
	return pojo.AppGameHomeItem{
		GameID:          item.GameID,
		GameName:        valueString(item.GameName),
		CategoryCode:    valueString(item.CategoryCode),
		Type:            item.Type,
		Manufacturer:    valueString(item.ThirdGameCategory),
		GameIcon:        valueString(item.GameIcon),
		HorizontalImage: valueString(item.HorizontalImage),
		Sort:            valueInt(item.Sort),
		ShowIndex:       valueInt(item.ShowIndex),
		FakeOnlineCount: item.FakeOnlineCount,
	}
}

func NewAppGameFakeOnlineCount() int {
	return AppGameFakeOnlineMin + rand.IntN(AppGameFakeOnlineMax-AppGameFakeOnlineMin+1)
}

func EnsureAppGameFakeOnlineSchema(db *gorm.DB) error {
	if db == nil {
		return errors.New("app_game_database_required")
	}

	migrator := db.Migrator()
	if !migrator.HasTable(&pojo.AppGame{}) {
		return errors.New("app_game_table_not_found")
	}

	columnAdded := false
	if !migrator.HasColumn(&pojo.AppGame{}, "FakeOnlineCount") {
		if err := migrator.AddColumn(&pojo.AppGame{}, "FakeOnlineCount"); err != nil {
			return err
		}
		columnAdded = true
	}

	query := db.Model(&pojo.AppGame{}).
		Where("COALESCE(deleted_flag, 0) = 0")
	if !columnAdded {
		query = query.Where("fake_online_count < ? OR fake_online_count > ?", AppGameFakeOnlineMin, AppGameFakeOnlineMax)
	}

	return updateAppGameFakeOnlineCounts(query).Error
}

func RefreshAppGameFakeOnlineCounts(db *gorm.DB) (int64, error) {
	if db == nil {
		return 0, errors.New("app_game_database_required")
	}

	result := updateAppGameFakeOnlineCounts(
		db.Model(&pojo.AppGame{}).Where("COALESCE(deleted_flag, 0) = 0"),
	)
	return result.RowsAffected, result.Error
}

func updateAppGameFakeOnlineCounts(query *gorm.DB) *gorm.DB {
	return query.UpdateColumn(
		"fake_online_count",
		gorm.Expr("FLOOR(? + RAND() * ?)", AppGameFakeOnlineMin, AppGameFakeOnlineMax-AppGameFakeOnlineMin+1),
	)
}

func SetAppGame(db *gorm.DB, req pojo.AppGameSet) (pojo.AppGame, error) {
	now := time.Now()
	var entity pojo.AppGame
	if req.GameID > 0 {
		if err := db.Where("game_id = ? AND COALESCE(deleted_flag, 0) = 0", req.GameID).First(&entity).Error; err != nil {
			return pojo.AppGame{}, err
		}
		_ = copier.Copy(&entity, &req)
		entity.UpdateTime = &now
		return entity, db.Save(&entity).Error
	}

	_ = copier.Copy(&entity, &req)
	entity.FakeOnlineCount = NewAppGameFakeOnlineCount()
	entity.CreateTime = &now
	entity.UpdateTime = &now
	return entity, db.Create(&entity).Error
}

func valueString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func valueInt(value *int) int {
	if value == nil {
		return 0
	}
	return *value
}

func UpsertAppGameByThirdID(db *gorm.DB, req pojo.AppGameSet) (created bool, err error) {
	if req.PlatformCode == nil {
		return false, errors.New("platform_code_required")
	}
	if req.ThirdGameID == nil {
		return false, errors.New("third_game_id_required")
	}
	platformCode := strings.TrimSpace(*req.PlatformCode)
	thirdGameID := strings.TrimSpace(*req.ThirdGameID)
	if platformCode == "" {
		return false, errors.New("platform_code_required")
	}
	if thirdGameID == "" {
		return false, errors.New("third_game_id_required")
	}

	now := time.Now()

	var entity pojo.AppGame
	err = db.Where("platform_code = ? AND third_game_id = ?", platformCode, thirdGameID).First(&entity).Error
	if err == nil {
		return false, nil
	}
	if err != gorm.ErrRecordNotFound {
		return false, err
	}

	hot := 0
	homeShow := 0
	disabledFlag := 0
	deletedFlag := 0
	_ = copier.Copy(&entity, &req)
	entity.GameID = 0
	entity.PlatformCode = &platformCode
	entity.ThirdGameID = &thirdGameID
	if entity.Hot == nil {
		entity.Hot = &hot
	}
	if entity.HomeShow == nil {
		entity.HomeShow = &homeShow
	}
	if entity.DisabledFlag == nil {
		entity.DisabledFlag = &disabledFlag
	}
	entity.DeletedFlag = &deletedFlag
	entity.FakeOnlineCount = NewAppGameFakeOnlineCount()
	entity.CreateTime = &now
	entity.UpdateTime = &now
	return true, db.Create(&entity).Error
}

func DelAppGame(db *gorm.DB, gameID int64) (string, error) {
	now := time.Now()
	result := db.Model(&pojo.AppGame{}).
		Where("game_id = ?", gameID).
		Updates(map[string]any{
			"deleted_flag": 1,
			"update_time":  now,
		})
	if result.Error != nil {
		return "", result.Error
	}
	if result.RowsAffected == 0 {
		return "", errors.New("record_not_found_delete")
	}
	return "success", nil
}
