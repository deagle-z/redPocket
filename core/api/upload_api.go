package api

import (
	"BaseGoUni/core/utils"
	"github.com/gin-gonic/gin"
)

// AdminUpload godoc
//
//	@Summary		上传文件（R2）
//	@Tags			文件上传
//	@Accept			multipart/form-data
//	@Produce		json
//	@Param			file formData file true "上传文件"
//	@Success		200	{object}	map[string]string
//	@Router			/api/v1/admin/upload [post]
func AdminUpload(ctx *gin.Context) {
	file, err := ctx.FormFile("file")
	if err != nil {
		utils.ErrorBack(ctx, "file is required")
		return
	}
	url, err := utils.UploadR2(ctx, file, "upload")
	if err != nil {
		utils.ErrorBack(ctx, err.Error())
		return
	}
	utils.SuccessObjBack(ctx, map[string]string{
		"url": url,
	})
}
