package api

import (
	"BaseGoUni/core/utils"
	"github.com/gin-gonic/gin"
	"strings"
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
		utils.ErrorBack(ctx, "file_required")
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

const (
	appUploadMaxSize = 5 * 1024 * 1024 // 5 MB
)

// AppUpload app端上传文件（R2）
func AppUpload(ctx *gin.Context) {
	file, err := ctx.FormFile("file")
	if err != nil {
		utils.ErrorBack(ctx, "file_required")
		return
	}
	if file.Size > appUploadMaxSize {
		utils.ErrorBack(ctx, "file_too_large_5mb")
		return
	}
	contentType := file.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		utils.ErrorBack(ctx, "image_only")
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
