package controller

import (
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/go-ddd-seed/go-ddd-seed/internal/application/command"
	"github.com/go-ddd-seed/go-ddd-seed/internal/application/commandHandler"
	appRepo "github.com/go-ddd-seed/go-ddd-seed/internal/application/port/repo"
	appApi "github.com/go-ddd-seed/go-ddd-seed/internal/application/port/thirdPartyApi"
	"github.com/go-ddd-seed/go-ddd-seed/internal/application/queryService"
	"github.com/go-ddd-seed/go-ddd-seed/internal/trigger/http/resp"
	"github.com/go-ddd-seed/go-ddd-seed/pkg/utils"
)

// FileController 文件控制器
// 写操作 -> commandHandler（经过领域层）
// 读操作 -> queryService（直接查询，CQRS 读写分离）
type FileController struct {
	handler      *commandHandler.FileCommandHandler
	queryService *queryService.FileRecordQueryService
	fileStore    appApi.FileStorage
}

// NewFileController 创建文件控制器
func NewFileController(
	handler *commandHandler.FileCommandHandler,
	queryService *queryService.FileRecordQueryService,
	fileStore appApi.FileStorage,
) *FileController {
	return &FileController{
		handler:      handler,
		queryService: queryService,
		fileStore:    fileStore,
	}
}

// RegisterRoutes 注册路由
func (c *FileController) RegisterRoutes(rg *gin.RouterGroup) {
	files := rg.Group("/files")
	files.POST("/upload", c.Upload)
	files.POST("/upload/batch", c.BatchUpload)
	files.GET("/:id/download", c.Download)
	files.GET("/:id", c.GetFileInfo)
	files.GET("", c.ListFiles)
	files.DELETE("/:id", c.Delete)
}

// Upload 上传文件
// @Summary      上传文件
// @Description  上传单个文件，支持关联业务对象
// @Tags         文件管理
// @Accept       multipart/form-data
// @Produce      json
// @Param        file        formData  file    true   "上传文件"
// @Param        attach_type formData  string  false  "关联业务类型"
// @Param        attach_id   formData  string  false  "关联业务ID"
// @Success      200  {object}  utils.Response{data=resp.FileUploadResp}  "上传成功"
// @Failure      400  {object}  docs.APIError  "请求参数错误"
// @Security     ApiKeyAuth
// @Router       /files/upload [post]
func (c *FileController) Upload(ctx *gin.Context) {
	file, err := ctx.FormFile("file")
	if err != nil {
		utils.FailWithMsg(ctx, http.StatusBadRequest, 2, "请选择上传文件")
		return
	}

	attachType := ctx.PostForm("attach_type")
	attachID := ctx.PostForm("attach_id")

	// 获取上传者 ID（从 JWT token 中提取）
	uploaderID := ctx.GetString("user_id")
	if uploaderID == "" {
		uploaderID = "unknown"
	}

	// 打开文件
	src, err := file.Open()
	if err != nil {
		utils.FailWithMsg(ctx, http.StatusInternalServerError, 7, "打开文件失败")
		return
	}
	defer src.Close()

	// 构造命令
	cmd := &command.UploadFileCommand{
		FileName:   file.Filename,
		FileSize:   file.Size,
		MIMEType:   file.Header.Get("Content-Type"),
		Content:    src,
		AttachType: attachType,
		AttachID:   attachID,
		UploaderID: uploaderID,
	}

	result, err := c.handler.Upload(ctx.Request.Context(), cmd)
	if err != nil {
		utils.Error(ctx, err)
		return
	}

	utils.Success(ctx, &resp.FileUploadResp{
		ID:           result.ID,
		FileName:     result.FileName,
		Size:         result.Size,
		MIMEType:     result.MIMEType,
		AccessURL:    result.AccessURL,
		ThumbnailURL: result.ThumbnailURL,
	})
}

// BatchUpload 批量上传
// @Summary      批量上传文件
// @Description  批量上传多个文件
// @Tags         文件管理
// @Accept       multipart/form-data
// @Produce      json
// @Param        files       formData  []file  true   "上传文件列表"
// @Param        attach_type formData  string  false  "关联业务类型"
// @Param        attach_id   formData  string  false  "关联业务ID"
// @Success      200  {object}  utils.Response{data=[]resp.FileUploadResp}  "上传成功"
// @Failure      400  {object}  docs.APIError  "请求参数错误"
// @Security     ApiKeyAuth
// @Router       /files/upload/batch [post]
func (c *FileController) BatchUpload(ctx *gin.Context) {
	form, err := ctx.MultipartForm()
	if err != nil {
		utils.FailWithMsg(ctx, http.StatusBadRequest, 2, "解析表单数据失败")
		return
	}

	files := form.File["files"]
	if len(files) == 0 {
		utils.FailWithMsg(ctx, http.StatusBadRequest, 2, "请选择上传文件")
		return
	}

	attachType := ctx.PostForm("attach_type")
	attachID := ctx.PostForm("attach_id")
	uploaderID := ctx.GetString("user_id")
	if uploaderID == "" {
		uploaderID = "unknown"
	}

	var results []resp.FileUploadResp
	for _, file := range files {
		src, err := file.Open()
		if err != nil {
			continue
		}

		cmd := &command.UploadFileCommand{
			FileName:   file.Filename,
			FileSize:   file.Size,
			MIMEType:   file.Header.Get("Content-Type"),
			Content:    src,
			AttachType: attachType,
			AttachID:   attachID,
			UploaderID: uploaderID,
		}

		result, err := c.handler.Upload(ctx.Request.Context(), cmd)
		src.Close()
		if err != nil {
			continue
		}

		results = append(results, resp.FileUploadResp{
			ID:        result.ID,
			FileName:  result.FileName,
			Size:      result.Size,
			MIMEType:  result.MIMEType,
			AccessURL: result.AccessURL,
		})
	}

	utils.Success(ctx, results)
}

// Download 下载文件
// @Summary      下载文件
// @Description  根据文件记录 ID 下载文件
// @Tags         文件管理
// @Produce      octet-stream
// @Param        id  path  string  true  "文件记录ID"
// @Success      200  {file}   file  "文件内容"
// @Failure      404  {object}  docs.APIError  "文件不存在"
// @Security     ApiKeyAuth
// @Router       /files/{id}/download [get]
func (c *FileController) Download(ctx *gin.Context) {
	id := ctx.Param("id")

	// 通过读仓储获取文件记录
	fileInfo, err := c.queryService.GetByID(ctx.Request.Context(), id)
	if err != nil || fileInfo == nil {
		utils.FailWithMsg(ctx, http.StatusNotFound, 5, "文件不存在")
		return
	}

	// 从存储后端下载
	reader, err := c.fileStore.Download(ctx.Request.Context(), fileInfo.StoragePath)
	if err != nil {
		utils.FailWithMsg(ctx, http.StatusInternalServerError, 7, "下载文件失败")
		return
	}
	defer reader.Close()

	// 设置响应头
	ctx.Header("Content-Type", fileInfo.MIMEType)
	ctx.Header("Content-Disposition", "attachment; filename=\""+fileInfo.FileName+"\"")
	ctx.Status(http.StatusOK)

	// 流式返回
	io.Copy(ctx.Writer, reader)
}

// GetFileInfo 获取文件信息
// @Summary      获取文件信息
// @Description  根据文件记录 ID 获取文件元数据和访问 URL
// @Tags         文件管理
// @Produce      json
// @Param        id  path  string  true  "文件记录ID"
// @Success      200  {object}  utils.Response{data=resp.FileInfoResp}  "文件信息"
// @Failure      404  {object}  docs.APIError  "文件不存在"
// @Security     ApiKeyAuth
// @Router       /files/{id} [get]
func (c *FileController) GetFileInfo(ctx *gin.Context) {
	id := ctx.Param("id")

	dto, err := c.queryService.GetByID(ctx.Request.Context(), id)
	if err != nil || dto == nil {
		utils.FailWithMsg(ctx, http.StatusNotFound, 5, "文件不存在")
		return
	}

	// 生成访问 URL
	accessURL, _ := c.fileStore.GetURL(ctx.Request.Context(), dto.StoragePath, 10*time.Minute)

	// 生成缩略图 URL（如果有）
	var thumbnailURL string
	if dto.ThumbnailPath != "" {
		thumbnailURL, _ = c.fileStore.GetURL(ctx.Request.Context(), dto.ThumbnailPath, 10*time.Minute)
	}

	utils.Success(ctx, &resp.FileInfoResp{
		ID:             dto.ID,
		FileName:       dto.FileName,
		Size:           dto.Size,
		MIMEType:       dto.MIMEType,
		StorageChannel: dto.StorageChannel,
		MD5Hash:        dto.MD5Hash,
		AttachType:     dto.AttachType,
		AttachID:       dto.AttachID,
		UploaderID:     dto.UploaderID,
		ThumbnailURL:   thumbnailURL,
		AccessURL:      accessURL,
		CreatedAt:      dto.CreatedAt.Format(time.DateTime),
	})
}

// ListFiles 文件列表
// @Summary      文件列表
// @Description  按条件分页查询文件记录
// @Tags         文件管理
// @Produce      json
// @Param        attach_type query  string  false  "关联业务类型"
// @Param        attach_id   query  string  false  "关联业务ID"
// @Param        page         query  int     false  "页码"  default(1)
// @Param        page_size    query  int     false  "每页数量"  default(20)
// @Success      200  {object}  utils.Response{data=queryService.PaginatedDTO}  "文件列表"
// @Security     ApiKeyAuth
// @Router       /files [get]
func (c *FileController) ListFiles(ctx *gin.Context) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("page_size", "20"))

	query := appRepo.FileRecordQuery{
		AttachType: ctx.Query("attach_type"),
		AttachID:   ctx.Query("attach_id"),
		Page:       page,
		PageSize:   pageSize,
	}

	result, err := c.queryService.List(ctx.Request.Context(), query)
	if err != nil {
		utils.Error(ctx, err)
		return
	}

	utils.Success(ctx, result)
}

// Delete 删除文件
// @Summary      删除文件
// @Description  软删除文件记录
// @Tags         文件管理
// @Produce      json
// @Param        id  path  string  true  "文件记录ID"
// @Success      200  {object}  utils.Response  "删除成功"
// @Failure      404  {object}  docs.APIError  "文件不存在"
// @Security     ApiKeyAuth
// @Router       /files/{id} [delete]
func (c *FileController) Delete(ctx *gin.Context) {
	id := ctx.Param("id")

	if err := c.handler.Delete(ctx.Request.Context(), &command.DeleteFileCommand{ID: id}); err != nil {
		utils.Error(ctx, err)
		return
	}

	utils.Success(ctx, nil)
}