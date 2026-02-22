package api

import (
	"BaseGoUni/core/pojo"
	"BaseGoUni/core/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"strconv"
)

func getTenantID(ctx *gin.Context) (tenantID int64, ok bool) {
	hostInfo := ctx.MustGet("hostInfo").(pojo.HostInfo)
	userId := ctx.MustGet("userId").(int64)
	user := utils.GetTempTenantUser(hostInfo.TablePrefix, userId)
	if user.ID == 0 || user.Status != 1 {
		utils.UnauthorizedBack(ctx, "token_error")
		return 0, false
	}
	return user.TenantId, true
}

func getDB(ctx *gin.Context) *gorm.DB {
	return ctx.MustGet("db").(*gorm.DB)
}

func parsePathID(ctx *gin.Context) (int64, bool) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		utils.ErrorBack(ctx, "参数格式错误")
		return 0, false
	}
	return id, true
}
