package controller

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-ddd-seed/go-ddd-seed/internal/application/command"
	"github.com/go-ddd-seed/go-ddd-seed/internal/application/commandHandler"
	"github.com/go-ddd-seed/go-ddd-seed/internal/application/queryService"
	"github.com/go-ddd-seed/go-ddd-seed/internal/trigger/http/req"
	"github.com/go-ddd-seed/go-ddd-seed/internal/trigger/http/resp"
	"github.com/go-ddd-seed/go-ddd-seed/pkg/utils"
	"github.com/go-ddd-seed/go-ddd-seed/pkg/validator"
)

// KvConfigController 键值配置控制器
// 写操作 -> commandHandler（经过领域层）
// 读操作 -> queryService（直接查询，CQRS 读写分离）
type KvConfigController struct {
	handler      *commandHandler.KvConfigCommandHandler
	queryService *queryService.KvConfigQueryService
}

// NewKvConfigController 创建键值配置控制器
func NewKvConfigController(
	handler *commandHandler.KvConfigCommandHandler,
	queryService *queryService.KvConfigQueryService,
) *KvConfigController {
	return &KvConfigController{handler: handler, queryService: queryService}
}

// RegisterRoutes 注册路由
func (c *KvConfigController) RegisterRoutes(rg *gin.RouterGroup) {
	kv := rg.Group("/infra/kv-configs")
	kv.POST("", c.Create)
	kv.PUT("/:id", c.Update)
	kv.DELETE("/:id", c.Delete)
	kv.GET("/:id", c.GetByID)
	kv.GET("/key/:key", c.GetByKey)
	kv.GET("", c.List)
}

// Create 创建配置
// @Summary      创建键值配置
// @Description  创建一条新的键值配置记录
// @Tags         基础设施-配置管理
// @Accept       json
// @Produce      json
// @Param        request  body  req.KvConfigCreateReq  true  "创建配置请求参数"
// @Success      200  {object}  resp.KvConfigResp  "创建成功"
// @Failure      400  {object}  docs.APIError      "请求参数错误"
// @Security     ApiKeyAuth
// @Router       /infra/kv-configs [post]
func (c *KvConfigController) Create(ctx *gin.Context) {
	var createReq req.KvConfigCreateReq
	if err := ctx.ShouldBindJSON(&createReq); err != nil {
		utils.FailWithMsg(ctx, http.StatusBadRequest, 2, err.Error())
		return
	}
	if err := validator.ValidateRequest(createReq); err != nil {
		utils.ValidationFail(ctx, err)
		return
	}

	cmd := &command.CreateKvConfigCommand{
		Key:         createReq.Key,
		Value:       createReq.Value,
		Description: createReq.Description,
	}

	result, err := c.handler.Create(ctx.Request.Context(), cmd)
	if err != nil {
		utils.Error(ctx, err)
		return
	}
	utils.Success(ctx, toKvConfigResp(result))
}

// Update 更新配置
// @Summary      更新键值配置
// @Description  更新指定键值配置的值和描述
// @Tags         基础设施-配置管理
// @Accept       json
// @Produce      json
// @Param        id       path  string                true  "配置ID"
// @Param        request  body  req.KvConfigUpdateReq  true  "更新配置请求参数"
// @Success      200  {object}  resp.KvConfigResp  "更新成功"
// @Failure      400  {object}  docs.APIError      "请求参数错误"
// @Failure      404  {object}  docs.APIError      "配置不存在"
// @Security     ApiKeyAuth
// @Router       /infra/kv-configs/{id} [put]
func (c *KvConfigController) Update(ctx *gin.Context) {
	id := ctx.Param("id")

	var updateReq req.KvConfigUpdateReq
	if err := ctx.ShouldBindJSON(&updateReq); err != nil {
		utils.FailWithMsg(ctx, http.StatusBadRequest, 2, err.Error())
		return
	}
	if err := validator.ValidateRequest(updateReq); err != nil {
		utils.ValidationFail(ctx, err)
		return
	}

	cmd := &command.UpdateKvConfigCommand{
		ID:          id,
		Value:       updateReq.Value,
		Description: updateReq.Description,
	}

	result, err := c.handler.Update(ctx.Request.Context(), cmd)
	if err != nil {
		utils.Error(ctx, err)
		return
	}
	utils.Success(ctx, toKvConfigResp(result))
}

// Delete 删除配置
// @Summary      删除键值配置
// @Description  删除指定键值配置
// @Tags         基础设施-配置管理
// @Produce      json
// @Param        id  path  string  true  "配置ID"
// @Success      200  {object}  docs.APIError  "删除成功"
// @Failure      404  {object}  docs.APIError  "配置不存在"
// @Security     ApiKeyAuth
// @Router       /infra/kv-configs/{id} [delete]
func (c *KvConfigController) Delete(ctx *gin.Context) {
	id := ctx.Param("id")
	if err := c.handler.Delete(ctx.Request.Context(), &command.DeleteKvConfigCommand{ID: id}); err != nil {
		utils.Error(ctx, err)
		return
	}
	utils.Success(ctx, nil)
}

// toKvConfigResp 将应用层 DTO 转换为 HTTP 响应
func toKvConfigResp(dto *commandHandler.KvConfigResult) *resp.KvConfigResp {
	return &resp.KvConfigResp{
		ID:          dto.ID,
		Key:         dto.Key,
		Value:       dto.Value,
		Description: dto.Description,
		Status:      dto.Status,
		CreatedAt:   dto.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   dto.UpdatedAt.Format(time.RFC3339),
	}
}

// GetByID 按 ID 查询
// @Summary      按ID查询配置
// @Description  根据ID获取键值配置详情
// @Tags         基础设施-配置管理
// @Produce      json
// @Param        id  path  string  true  "配置ID"
// @Success      200  {object}  resp.KvConfigResp  "配置详情"
// @Failure      404  {object}  docs.APIError      "配置不存在"
// @Security     ApiKeyAuth
// @Router       /infra/kv-configs/{id} [get]
func (c *KvConfigController) GetByID(ctx *gin.Context) {
	id := ctx.Param("id")
	result, err := c.queryService.GetByID(ctx.Request.Context(), id)
	if err != nil {
		utils.Error(ctx, err)
		return
	}
	utils.Success(ctx, result)
}

// GetByKey 按 Key 查询
// @Summary      按Key查询配置
// @Description  根据配置键名获取键值配置详情
// @Tags         基础设施-配置管理
// @Produce      json
// @Param        key  path  string  true  "配置键名"
// @Success      200  {object}  resp.KvConfigResp  "配置详情"
// @Failure      404  {object}  docs.APIError      "配置不存在"
// @Security     ApiKeyAuth
// @Router       /infra/kv-configs/key/{key} [get]
func (c *KvConfigController) GetByKey(ctx *gin.Context) {
	key := ctx.Param("key")
	result, err := c.queryService.GetByKey(ctx.Request.Context(), key)
	if err != nil {
		utils.Error(ctx, err)
		return
	}
	utils.Success(ctx, result)
}

// List 列表查询
// @Summary      键值配置列表
// @Description  分页查询键值配置列表，支持按状态过滤
// @Tags         基础设施-配置管理
// @Produce      json
// @Param        page      query  int  false  "页码，默认1"  minimum(1)
// @Param        page_size query  int  false  "每页数量，默认20"  maximum(100)
// @Param        status    query  int  false  "状态过滤：-1=全部 0=禁用 1=启用"
// @Success      200  {object}  resp.PaginatedResp{list=[]resp.KvConfigResp}  "配置列表"
// @Failure      400  {object}  docs.APIError  "请求参数错误"
// @Security     ApiKeyAuth
// @Router       /infra/kv-configs [get]
func (c *KvConfigController) List(ctx *gin.Context) {
	var listReq req.KvConfigListReq
	if err := ctx.ShouldBindQuery(&listReq); err != nil {
		utils.FailWithMsg(ctx, http.StatusBadRequest, 2, err.Error())
		return
	}
	if err := validator.ValidateRequest(listReq); err != nil {
		utils.ValidationFail(ctx, err)
		return
	}

	// 参数校验/默认值
	page := listReq.Page
	if page < 1 {
		page = 1
	}
	pageSize := listReq.PageSize
	if pageSize < 1 {
		pageSize = 20
	}
	// status: -1=全部, 0=禁用, 1=启用
	// 当未传入时（零值 0），应视为查询全部(status=-1)，避免误过滤
	status := listReq.Status
	if status == 0 && ctx.Query("status") == "" {
		status = -1 // 默认查询全部
	}

	result, err := c.queryService.List(ctx.Request.Context(), page, pageSize, status)
	if err != nil {
		utils.Error(ctx, err)
		return
	}
	utils.Success(ctx, result)
}