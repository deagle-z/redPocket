package api

import (
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/repository"
	"BaseGoUni/core/utils"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func TgAuthLogin(ctx *gin.Context) {
	var req pojo.TgAuthLoginReq
	if err := ctx.BindJSON(&req); err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	if req.ID <= 0 || req.AuthDate <= 0 || req.Hash == "" {
		utils.ErrorBack(ctx, "参数格式错误")
		return
	}

	tempHostInfo := ctx.MustGet("hostInfo").(pojo.HostInfo)
	db := ctx.MustGet("db").(*gorm.DB)
	onlineUser := pojo.OnlineUser{
		Username:  req.Username,
		Browser:   ctx.GetHeader("User-Agent"),
		Ip:        utils.GetIPAddress(ctx),
		LoginTime: time.Now(),
	}
	data, err := repository.TgAuthLogin(db, tempHostInfo, req, onlineUser)
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, data)
}
