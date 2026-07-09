package api

import (
	"BaseGoUni/core/game"
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/repository"
	"BaseGoUni/core/utils"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const hgGameAssetDomain = "https://hgapi.com"
const appGameLaunchMinimumRechargeAmount = 50.0

// GetAppGames godoc
//
//	@Summary		获取本地游戏列表
//	@Tags			游戏
//	@Accept			json
//	@Produce		json
//	@Param			data body		pojo.AppGameSearch	true	"查询条件"
//	@Success		200	{object}		pojo.AppGameResp
//	@Router			/api/v1/admin/appGame/list [post]
func GetAppGames(ctx *gin.Context) {
	var search pojo.AppGameSearch
	search.SetPageDefaults()
	if err := ctx.ShouldBindJSON(&search); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result := repository.GetAppGames(db, search)
	utils.SuccessObjBack(ctx, result)
}

// GetAppGameThirdCategories godoc
//
//	@Summary		按分类获取游戏厂商列表
//	@Tags			游戏
//	@Accept			json
//	@Produce		json
//	@Param			data body		pojo.AppGameThirdCategorySearch	true	"查询条件"
//	@Success		200	{object}		pojo.AppGameThirdCategoryResp
//	@Router			/api/v1/app/appGame/thirdCategories [post]
func GetAppGameThirdCategories(ctx *gin.Context) {
	var search pojo.AppGameThirdCategorySearch
	if err := ctx.ShouldBindJSON(&search); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result := repository.GetAppGameThirdCategories(db, search.CategoryCode)
	utils.SuccessObjBack(ctx, result)
}

// GetAppGameListApp godoc
//
//	@Summary		App端分页查询游戏列表
//	@Tags			游戏
//	@Accept			json
//	@Produce		json
//	@Param			data body		pojo.AppGameSearch	true	"查询条件"
//	@Success		200	{object}		pojo.AppGameHomeListResp
//	@Router			/api/v1/app/appGame/list [post]
func GetAppGameListApp(ctx *gin.Context) {
	var search pojo.AppGameSearch
	search.SetPageDefaults()
	if err := ctx.ShouldBindJSON(&search); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result := repository.GetAppGameHomeList(db, search)
	utils.SuccessObjBack(ctx, result)
}

// GetAppGameListByCategoryCodeApp godoc
//
//	@Summary		App端按分类查询游戏列表
//	@Tags			游戏
//	@Accept			json
//	@Produce		json
//	@Param			data body		pojo.AppGameSearch	true	"查询条件"
//	@Success		200	{object}		pojo.AppGameHomeListResp
//	@Router			/api/v1/app/appGame/gameList [post]
func GetAppGameListByCategoryCodeApp(ctx *gin.Context) {
	var search pojo.AppGameSearch
	search.SetPageDefaults()
	if err := ctx.ShouldBindJSON(&search); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result := repository.GetAppGameListByCategoryCode(db, search)
	utils.SuccessObjBack(ctx, result)
}

// GetAppGameCategoryListApp godoc
//
//	@Summary		App端按分类和厂商分页查询游戏列表
//	@Tags			游戏
//	@Accept			json
//	@Produce		json
//	@Param			data body		pojo.AppGameSearch	true	"查询条件"
//	@Success		200	{object}		pojo.AppGameCategoryListResp
//	@Router			/api/v1/app/appGame/categoryList [post]
func GetAppGameCategoryListApp(ctx *gin.Context) {
	var search pojo.AppGameSearch
	search.SetPageDefaults()
	if err := ctx.ShouldBindJSON(&search); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result := repository.GetAppGameCategoryList(db, search)
	utils.SuccessObjBack(ctx, result)
}

// SetAppGame godoc
//
//	@Summary		创建或更新本地游戏
//	@Tags			游戏
//	@Accept			json
//	@Produce		json
//	@Param			data body		pojo.AppGameSet	true	"游戏信息"
//	@Success		200	{object}		pojo.AppGame
//	@Router			/api/v1/admin/appGame [post]
func SetAppGame(ctx *gin.Context) {
	var req pojo.AppGameSet
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result, err := repository.SetAppGame(db, req)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}

// GetAppHomeGames godoc
//
//	@Summary		App端获取首页游戏列表
//	@Tags			游戏
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}		pojo.AppGameHomeResp
//	@Router			/api/v1/app/appGame/home [get]
func GetAppHomeGames(ctx *gin.Context) {
	db := ctx.MustGet("db").(*gorm.DB)
	result := repository.GetHomeAppGamesGrouped(db)
	utils.SuccessObjBack(ctx, result)
}

// LaunchAppGame godoc
//
//	@Summary		App端获取游戏登录URL
//	@Tags			游戏
//	@Accept			json
//	@Produce		json
//	@Param			data body		pojo.AppGameLaunchReq	true	"启动参数"
//	@Success		200	{object}		pojo.AppGameLaunchResp
//	@Router			/api/v1/app/appGame/launch [post]
func LaunchAppGame(ctx *gin.Context) {
	var req pojo.AppGameLaunchReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	language := strings.TrimSpace(req.Language)
	if language == "" {
		language = "en"
	}

	userID := ctx.MustGet("userId").(int64)
	db := ctx.MustGet("db").(*gorm.DB)
	var tgUser pojo.TgUser
	if err := db.Select("id, uid, status, recharge_amount, tg_name, first_name, password_plain").Where("id = ? AND status <> ?", userID, int8(-1)).First(&tgUser).Error; err != nil {
		utils.ErrorBack(ctx, "player not found")
		return
	}
	if tgUser.Status != 1 {
		utils.ErrorBack(ctx, "player disabled")
		return
	}
	rebateTransferred := false
	if tgUser.RechargeAmount < appGameLaunchMinimumRechargeAmount {
		rebateTransferred = repository.GetUserRechargeCount(db, userID).RebateTransferred
	}
	if !canLaunchAppGame(tgUser.RechargeAmount, rebateTransferred) {
		utils.ErrorBack(ctx, "game_recharge_required")
		return
	}

	appGame, err := repository.GetEnabledAppGameByID(db, req.GameID)
	if err != nil {
		utils.ErrorBack(ctx, "game not found")
		return
	}
	thirdGameID := strings.TrimSpace(appGameStringValue(appGame.ThirdGameID))
	if thirdGameID == "" {
		utils.ErrorBack(ctx, "third game id is empty")
		return
	}

	platformCode := appGamePlatformCode(appGame.PlatformCode)
	if shouldLaunchWithHGClient(platformCode) {
		launchHGAppGame(ctx, tgUser, thirdGameID, language)
		return
	}
	if shouldLaunchWithGSCClient(platformCode) {
		launchGSCAppGame(ctx, tgUser)
		return
	}

	utils.ErrorBack(ctx, "game_launch_not_supported")
}

// LaunchGSCSportGame godoc
//
//	@Summary		App端直接启动 GSC 体育
//	@Tags			游戏
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}		pojo.AppGameLaunchResp
//	@Router			/api/v1/app/gsc/launch [post]
func LaunchGSCSportGame(ctx *gin.Context) {
	userID := ctx.MustGet("userId").(int64)
	db := ctx.MustGet("db").(*gorm.DB)
	var tgUser pojo.TgUser
	if err := db.Select("id, uid, status, recharge_amount, tg_name, first_name, password_plain").Where("id = ? AND status <> ?", userID, int8(-1)).First(&tgUser).Error; err != nil {
		utils.ErrorBack(ctx, "player not found")
		return
	}
	if tgUser.Status != 1 {
		utils.ErrorBack(ctx, "player disabled")
		return
	}
	rebateTransferred := false
	if tgUser.RechargeAmount < appGameLaunchMinimumRechargeAmount {
		rebateTransferred = repository.GetUserRechargeCount(db, userID).RebateTransferred
	}
	if !canLaunchAppGame(tgUser.RechargeAmount, rebateTransferred) {
		utils.ErrorBack(ctx, "game_recharge_required")
		return
	}

	launchGSCAppGame(ctx, tgUser)
}

func launchHGAppGame(ctx *gin.Context, tgUser pojo.TgUser, thirdGameID string, language string) {
	resp, err := game.NewClient().GameLaunch(tgUser.Uid, thirdGameID, language)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	if resp.Code != game.GameCodeSuccess {
		msg := strings.TrimSpace(resp.Error)
		if msg == "" {
			msg = game.ErrorMessage(resp.Code)
		}
		utils.ErrorBack(ctx, fmt.Sprintf("game api error code=%d error=%s", resp.Code, msg))
		return
	}

	utils.SuccessObjBack(ctx, pojo.AppGameLaunchResp{URL: resp.Data.URL})
}

func launchGSCAppGame(ctx *gin.Context, tgUser pojo.TgUser) {
	resp, err := game.NewGSCClient().LaunchGame(game.GSCLaunchGameInput{
		MemberAccount:    strings.TrimSpace(tgUser.Uid),
		Nickname:         appGameGSCNickname(tgUser),
		IP:               utils.GetIPAddress(ctx),
		OperatorLobbyURL: appGameGSCOperatorLobbyURL(ctx),
	})
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	if resp.Code != 0 && resp.Code != http.StatusOK {
		msg := strings.TrimSpace(resp.Message)
		if msg == "" {
			msg = "gsc launch failed"
		}
		utils.ErrorBack(ctx, fmt.Sprintf("gsc launch api error code=%d message=%s", resp.Code, msg))
		return
	}
	if strings.TrimSpace(resp.URL) == "" && strings.TrimSpace(resp.Content) == "" {
		utils.ErrorBack(ctx, "gsc launch response empty")
		return
	}

	utils.SuccessObjBack(ctx, pojo.AppGameLaunchResp{
		URL:     resp.URL,
		Content: resp.Content,
	})
}

func appGameGSCNickname(user pojo.TgUser) string {
	if value := appGameStringValue(user.FirstName); value != "" {
		return value
	}
	if value := appGameStringValue(user.TgName); value != "" {
		return value
	}
	return strings.TrimSpace(user.Uid)
}

func appGameGSCOperatorLobbyURL(ctx *gin.Context) string {
	host := strings.TrimSpace(ctx.GetHeader("X-Forwarded-Host"))
	if host == "" {
		host = strings.TrimSpace(ctx.Request.Host)
	}
	if host == "" {
		return ""
	}
	scheme := strings.TrimSpace(ctx.GetHeader("X-Forwarded-Proto"))
	if scheme == "" {
		if ctx.Request.TLS != nil {
			scheme = "https"
		} else {
			scheme = "http"
		}
	}
	return scheme + "://" + host
}

func appGamePlatformCode(platformCode *string) string {
	return strings.ToLower(strings.TrimSpace(appGameStringValue(platformCode)))
}

func shouldLaunchWithGSCClient(platformCode string) bool {
	return strings.EqualFold(strings.TrimSpace(platformCode), "gsc")
}

func canLaunchAppGame(rechargeAmount float64, rebateTransferred bool) bool {
	return rechargeAmountAtLeastMinimum(rechargeAmount) || rebateTransferred
}

func rechargeAmountAtLeastMinimum(rechargeAmount float64) bool {
	return rechargeAmount >= appGameLaunchMinimumRechargeAmount
}

func shouldLaunchWithHGClient(platformCode string) bool {
	platformCode = strings.ToLower(strings.TrimSpace(platformCode))
	return platformCode == "" || platformCode == "hg"
}

// SyncAppGames godoc
//
//	@Summary		同步第三方游戏列表到本地游戏表
//	@Tags			游戏
//	@Accept			json
//	@Produce		json
//	@Param			data body		pojo.AppGameSyncReq	false	"同步参数"
//	@Success		200	{object}		pojo.AppGameSyncResp
//	@Router			/api/v1/admin/appGame/sync [post]
func SyncAppGames(ctx *gin.Context) {
	var req pojo.AppGameSyncReq
	_ = ctx.ShouldBindJSON(&req)

	language := strings.TrimSpace(req.Language)
	if language == "" {
		language = "en"
	}
	platformCode := strings.ToLower(strings.TrimSpace(req.PlatformCode))
	if platformCode == "" {
		platformCode = "hg"
	}

	db := ctx.MustGet("db").(*gorm.DB)
	var (
		result pojo.AppGameSyncResp
		err    error
	)
	switch platformCode {
	case "hg":
		result, err = syncHGAppGames(db, platformCode, language)
	case "gsc":
		result, err = syncGSCAppGames(db, platformCode)
	default:
		utils.ErrorBack(ctx, "unsupported_game_platform")
		return
	}
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}

	utils.SuccessObjBack(ctx, result)
}

func syncHGAppGames(db *gorm.DB, platformCode string, language string) (pojo.AppGameSyncResp, error) {
	resp, err := game.NewClient().GameList(language)
	if err != nil {
		return pojo.AppGameSyncResp{}, err
	}
	if resp.Code != game.GameCodeSuccess {
		msg := strings.TrimSpace(resp.Error)
		if msg == "" {
			msg = game.ErrorMessage(resp.Code)
		}
		return pojo.AppGameSyncResp{}, fmt.Errorf("game api error code=%d error=%s", resp.Code, msg)
	}

	games := resp.Data.Games()
	result := pojo.AppGameSyncResp{Total: len(games)}
	err = db.Transaction(func(tx *gorm.DB) error {
		for i, item := range games {
			if strings.TrimSpace(item.ID) == "" && strings.TrimSpace(item.GameID) == "" {
				result.Skipped++
				continue
			}
			created, err := repository.UpsertAppGameByThirdID(tx, buildAppGameSetFromGameInfo(platformCode, i+1, item))
			if err != nil {
				return err
			}
			if created {
				result.Created++
			} else {
				result.Skipped++
			}
		}
		return nil
	})
	if err != nil {
		return result, err
	}

	return result, nil
}

func syncGSCAppGames(db *gorm.DB, platformCode string) (pojo.AppGameSyncResp, error) {
	cfg := game.GetGSCConfig()
	games, err := game.NewGSCClientWithConfig(cfg).GameList()
	if err != nil {
		return pojo.AppGameSyncResp{}, err
	}

	result := pojo.AppGameSyncResp{Total: len(games)}
	seenThirdGameIDs := make(map[string]struct{})
	err = db.Transaction(func(tx *gorm.DB) error {
		for i, item := range games {
			thirdGameID := strings.TrimSpace(item.GameCode)
			if thirdGameID == "" {
				result.Skipped++
				continue
			}
			if !shouldSyncGSCProviderGame(item, cfg.SupportCurrency) {
				result.Skipped++
				continue
			}
			thirdGameIDKey := strings.ToLower(thirdGameID)
			if _, exists := seenThirdGameIDs[thirdGameIDKey]; exists {
				result.Skipped++
				continue
			}
			seenThirdGameIDs[thirdGameIDKey] = struct{}{}

			created, err := repository.UpsertAppGameByThirdID(tx, buildAppGameSetFromGSCProviderGame(platformCode, i+1, item, cfg.CategoryMap))
			if err != nil {
				return err
			}
			if created {
				result.Created++
			} else {
				result.Skipped++
			}
		}
		return nil
	})
	if err != nil {
		return result, err
	}
	return result, nil
}

func buildAppGameSetFromGameInfo(platformCode string, sort int, info game.GameInfo) pojo.AppGameSet {
	thirdGameID := strings.TrimSpace(info.ID)
	if thirdGameID == "" {
		thirdGameID = strings.TrimSpace(info.GameID)
	}
	gameName := strings.TrimSpace(info.Name)
	manufacturer := strings.TrimSpace(info.Manufacturer)
	iconURL := normalizeHGGameAssetURL(info.IconURL)
	portraitIconURL := normalizeHGGameAssetURL(info.PortraitIconURL)
	showIndex := sort
	gameType := info.Type
	categoryCode := appGameCategoryCodeByType(gameType)
	return pojo.AppGameSet{
		GameName:          &gameName,
		CategoryCode:      &categoryCode,
		ShowIndex:         &showIndex,
		Type:              &gameType,
		PlatformCode:      &platformCode,
		ThirdGameID:       &thirdGameID,
		ThirdGameName:     &gameName,
		ThirdGameCategory: &manufacturer,
		HorizontalImage:   &portraitIconURL,
		GameIcon:          &iconURL,
		Sort:              &sort,
	}
}

func buildAppGameSetFromGSCProviderGame(platformCode string, sort int, info game.GSCProviderGame, categoryMap map[string]string) pojo.AppGameSet {
	thirdGameID := strings.TrimSpace(info.GameCode)
	gameName := firstNonEmptyString(info.LangName["0"], info.GameName, thirdGameID)
	gameTypeText := strings.TrimSpace(info.GameType)
	gameIcon := firstNonEmptyString(info.ImageURL, info.LangIcon["0"])
	categoryCode := gscGameCategoryCode(gameTypeText, categoryMap)
	localType := appGameTypeByCategoryCode(categoryCode)
	showIndex := sort
	disabledFlag := 0
	if !isGSCGameActive(info.Status) {
		disabledFlag = 1
	}
	return pojo.AppGameSet{
		GameName:          &gameName,
		CategoryCode:      &categoryCode,
		ShowIndex:         &showIndex,
		Type:              &localType,
		PlatformCode:      &platformCode,
		ThirdGameID:       &thirdGameID,
		ThirdGameName:     &gameName,
		ThirdGameCategory: &gameTypeText,
		HorizontalImage:   &gameIcon,
		GameIcon:          &gameIcon,
		Sort:              &sort,
		DisabledFlag:      &disabledFlag,
	}
}

func shouldSyncGSCProviderGame(info game.GSCProviderGame, supportCurrency string) bool {
	supportCurrency = strings.TrimSpace(supportCurrency)
	if supportCurrency == "" {
		return true
	}
	return strings.EqualFold(strings.TrimSpace(info.SupportCurrency), supportCurrency)
}

func gscGameCategoryCode(gameType string, categoryMap map[string]string) string {
	key := strings.ToUpper(strings.TrimSpace(gameType))
	if categoryMap != nil {
		if categoryCode := strings.TrimSpace(categoryMap[key]); categoryCode != "" {
			return strings.ToLower(categoryCode)
		}
	}
	return "mini"
}

func appGameTypeByCategoryCode(categoryCode string) int {
	switch strings.ToLower(strings.TrimSpace(categoryCode)) {
	case "slots":
		return 0
	case "mini":
		return 1
	case "casino":
		return 2
	case "fishing":
		return 3
	case "lottery":
		return 4
	default:
		return 1
	}
}

func isGSCGameActive(status string) bool {
	status = strings.ToUpper(strings.TrimSpace(status))
	return status == "ACTIVATED" || status == "ACTIVAT"
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func appGameStringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func normalizeHGGameAssetURL(rawURL string) string {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return ""
	}
	lowerURL := strings.ToLower(rawURL)
	if strings.HasPrefix(lowerURL, "http://") || strings.HasPrefix(lowerURL, "https://") {
		return rawURL
	}
	if strings.HasPrefix(rawURL, "/") {
		return hgGameAssetDomain + rawURL
	}
	return hgGameAssetDomain + "/" + rawURL
}

func appGameCategoryCodeByType(gameType int) string {
	switch gameType {
	case 0:
		return "slots"
	case 1:
		return "mini"
	case 2:
		return "casino"
	case 3:
		return "fishing"
	case 4:
		return "lottery"
	default:
		return fmt.Sprintf("type_%d", gameType)
	}
}
