package api

import (
	"BaseGoUni/core/game"
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/repository"
	"BaseGoUni/core/utils"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const hgGameAssetDomain = "https://hgapi.com"

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
	if err := db.Select("id, uid, status").Where("id = ? AND status <> ?", userID, int8(-1)).First(&tgUser).Error; err != nil {
		utils.ErrorBack(ctx, "player not found")
		return
	}
	if tgUser.Status != 1 {
		utils.ErrorBack(ctx, "player disabled")
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
	platformCode := strings.TrimSpace(req.PlatformCode)
	if platformCode == "" {
		platformCode = "hg"
	}

	resp, err := game.NewClient().GameList(language)
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

	games := resp.Data.Games()
	result := pojo.AppGameSyncResp{Total: len(games)}
	db := ctx.MustGet("db").(*gorm.DB)
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
				result.Updated++
			}
		}
		return nil
	})
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}

	utils.SuccessObjBack(ctx, result)
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
