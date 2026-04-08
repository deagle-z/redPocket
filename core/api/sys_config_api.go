package api

import (
	"BaseGoUni/core/repository"
	"BaseGoUni/core/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GetAppSysConfig godoc
//
//	@Summary		App端根据key获取系统配置
//	@Tags			系统配置
//	@Produce		json
//	@Param			key path	string	true	"配置key"
//	@Success		200	{object}	pojo.SysConfigBack
//	@Router			/api/v1/app/config/:key [get]
func GetAppSysConfig(ctx *gin.Context) {
	key := ctx.Param("key")
	if key == "" {
		utils.ErrorBack(ctx, "key不能为空")
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	result, err := repository.GetSysConfigByKey(db, key)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, result)
}
