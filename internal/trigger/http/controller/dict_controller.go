package controller

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-ddd-seed/go-ddd-seed/internal/application/command"
	"github.com/go-ddd-seed/go-ddd-seed/internal/application/commandHandler"
	"github.com/go-ddd-seed/go-ddd-seed/internal/application/queryService"
	"github.com/go-ddd-seed/go-ddd-seed/internal/application/port/repo"
	"github.com/go-ddd-seed/go-ddd-seed/internal/trigger/http/req"
	"github.com/go-ddd-seed/go-ddd-seed/internal/trigger/http/resp"
	"github.com/go-ddd-seed/go-ddd-seed/pkg/utils"
	"github.com/go-ddd-seed/go-ddd-seed/pkg/validator"
)

// DictController 字典控制器
// 写操作 -> commandHandler（经过领域层）
// 读操作 -> queryService（直接查询，CQRS 读写分离）
// 前端枚举获取 -> queryService（优先缓存）
type DictController struct {
	handler      *commandHandler.DictCommandHandler
	queryService *queryService.DictQueryService
}

// NewDictController 创建字典控制器
func NewDictController(
	handler *commandHandler.DictCommandHandler,
	queryService *queryService.DictQueryService,
) *DictController {
	return &DictController{handler: handler, queryService: queryService}
}

// WarmUp 缓存预热（启动时调用）
func (c *DictController) WarmUp(ctx context.Context) error {
	return c.queryService.WarmUpCache(ctx)
}

// RegisterRoutes 注册路由
func (c *DictController) RegisterRoutes(rg *gin.RouterGroup) {
	// 字典类型路由
	types := rg.Group("/infra/dict-types")
	types.POST("", c.CreateType)
	types.PUT("/:id", c.UpdateType)
	types.DELETE("/:id", c.DeleteType)
	types.PUT("/:id/enable", c.EnableType)
	types.PUT("/:id/disable", c.DisableType)
	types.GET("/:id", c.GetTypeByID)
	types.GET("/code/:code", c.GetTypeByCode)
	types.GET("", c.ListTypes)

	// 字典条目路由
	entries := rg.Group("/infra/dict-entries")
	entries.POST("", c.AddEntry)
	entries.PUT("/:id", c.UpdateEntry)
	entries.DELETE("/:id", c.RemoveEntry)
	entries.PUT("/:id/enable", c.EnableEntry)
	entries.PUT("/:id/disable", c.DisableEntry)
	entries.GET("/:id", c.GetEntryByID)
	entries.GET("", c.ListEntries)

	// 前端枚举值获取（含缓存）
	dicts := rg.Group("/infra/dicts")
	dicts.GET("/entries/:code", c.GetEntriesByCode)
}

// ----- 字典类型 -----

// CreateType 创建字典类型
// @Summary      创建字典类型
// @Description  创建新的字典类型定义
// @Tags         基础设施-字典管理
// @Accept       json
// @Produce      json
// @Param        request  body  req.DictTypeCreateReq  true  "创建字典类型请求参数"
// @Success      200  {object}  resp.DictTypeResp  "创建成功"
// @Failure      400  {object}  docs.APIError      "请求参数错误"
// @Security     ApiKeyAuth
// @Router       /infra/dict-types [post]
func (c *DictController) CreateType(ctx *gin.Context) {
	var createReq req.DictTypeCreateReq
	if err := ctx.ShouldBindJSON(&createReq); err != nil {
		utils.FailWithMsg(ctx, http.StatusBadRequest, 2, err.Error())
		return
	}
	if err := validator.ValidateRequest(createReq); err != nil {
		utils.ValidationFail(ctx, err)
		return
	}

	cmd := &command.CreateDictTypeCommand{
		Code:        createReq.Code,
		Name:        createReq.Name,
		Description: createReq.Description,
	}

	result, err := c.handler.CreateType(ctx.Request.Context(), cmd)
	if err != nil {
		utils.Error(ctx, err)
		return
	}
	utils.Success(ctx, toDictTypeResp(result))
}

// UpdateType 更新字典类型
// @Summary      更新字典类型
// @Description  更新指定字典类型的名称和描述
// @Tags         基础设施-字典管理
// @Accept       json
// @Produce      json
// @Param        id       path  string                true  "字典类型ID"
// @Param        request  body  req.DictTypeUpdateReq  true  "更新字典类型请求参数"
// @Success      200  {object}  resp.DictTypeResp  "更新成功"
// @Failure      400  {object}  docs.APIError      "请求参数错误"
// @Failure      404  {object}  docs.APIError      "字典类型不存在"
// @Security     ApiKeyAuth
// @Router       /infra/dict-types/{id} [put]
func (c *DictController) UpdateType(ctx *gin.Context) {
	id := ctx.Param("id")

	var updateReq req.DictTypeUpdateReq
	if err := ctx.ShouldBindJSON(&updateReq); err != nil {
		utils.FailWithMsg(ctx, http.StatusBadRequest, 2, err.Error())
		return
	}
	if err := validator.ValidateRequest(updateReq); err != nil {
		utils.ValidationFail(ctx, err)
		return
	}

	cmd := &command.UpdateDictTypeCommand{
		ID:          id,
		Name:        updateReq.Name,
		Description: updateReq.Description,
	}

	result, err := c.handler.UpdateType(ctx.Request.Context(), cmd)
	if err != nil {
		utils.Error(ctx, err)
		return
	}
	utils.Success(ctx, toDictTypeResp(result))
}

// DeleteType 删除字典类型
// @Summary      删除字典类型
// @Description  删除指定的字典类型
// @Tags         基础设施-字典管理
// @Produce      json
// @Param        id  path  string  true  "字典类型ID"
// @Success      200  {object}  docs.APIError  "删除成功"
// @Failure      404  {object}  docs.APIError  "字典类型不存在"
// @Security     ApiKeyAuth
// @Router       /infra/dict-types/{id} [delete]
func (c *DictController) DeleteType(ctx *gin.Context) {
	id := ctx.Param("id")
	if err := c.handler.DeleteType(ctx.Request.Context(), &command.DeleteDictTypeCommand{ID: id}); err != nil {
		utils.Error(ctx, err)
		return
	}
	utils.Success(ctx, nil)
}

// EnableType 启用字典类型
// @Summary      启用字典类型
// @Description  启用指定的字典类型
// @Tags         基础设施-字典管理
// @Produce      json
// @Param        id  path  string  true  "字典类型ID"
// @Success      200  {object}  resp.DictTypeResp  "启用成功"
// @Failure      404  {object}  docs.APIError      "字典类型不存在"
// @Security     ApiKeyAuth
// @Router       /infra/dict-types/{id}/enable [put]
func (c *DictController) EnableType(ctx *gin.Context) {
	id := ctx.Param("id")
	result, err := c.handler.EnableType(ctx.Request.Context(), &command.EnableDictTypeCommand{ID: id})
	if err != nil {
		utils.Error(ctx, err)
		return
	}
	utils.Success(ctx, toDictTypeResp(result))
}

// DisableType 禁用字典类型
// @Summary      禁用字典类型
// @Description  禁用指定的字典类型
// @Tags         基础设施-字典管理
// @Produce      json
// @Param        id  path  string  true  "字典类型ID"
// @Success      200  {object}  resp.DictTypeResp  "禁用成功"
// @Failure      404  {object}  docs.APIError      "字典类型不存在"
// @Security     ApiKeyAuth
// @Router       /infra/dict-types/{id}/disable [put]
func (c *DictController) DisableType(ctx *gin.Context) {
	id := ctx.Param("id")
	result, err := c.handler.DisableType(ctx.Request.Context(), &command.DisableDictTypeCommand{ID: id})
	if err != nil {
		utils.Error(ctx, err)
		return
	}
	utils.Success(ctx, toDictTypeResp(result))
}

// GetTypeByID 按 ID 查询类型
// @Summary      按ID查询字典类型
// @Description  根据ID获取字典类型详情
// @Tags         基础设施-字典管理
// @Produce      json
// @Param        id  path  string  true  "字典类型ID"
// @Success      200  {object}  resp.DictTypeResp  "字典类型详情"
// @Failure      404  {object}  docs.APIError      "字典类型不存在"
// @Security     ApiKeyAuth
// @Router       /infra/dict-types/{id} [get]
func (c *DictController) GetTypeByID(ctx *gin.Context) {
	id := ctx.Param("id")
	result, err := c.queryService.GetByID(ctx.Request.Context(), id)
	if err != nil {
		utils.Error(ctx, err)
		return
	}
	utils.Success(ctx, toDictTypeRespFromDTO(result))
}

// GetTypeByCode 按 Code 查询类型
// @Summary      按编码查询字典类型
// @Description  根据编码获取字典类型详情
// @Tags         基础设施-字典管理
// @Produce      json
// @Param        code  path  string  true  "字典类型编码"
// @Success      200  {object}  resp.DictTypeResp  "字典类型详情"
// @Failure      404  {object}  docs.APIError      "字典类型不存在"
// @Security     ApiKeyAuth
// @Router       /infra/dict-types/code/{code} [get]
func (c *DictController) GetTypeByCode(ctx *gin.Context) {
	code := ctx.Param("code")
	result, err := c.queryService.GetByCode(ctx.Request.Context(), code)
	if err != nil {
		utils.Error(ctx, err)
		return
	}
	utils.Success(ctx, toDictTypeRespFromDTO(result))
}

// ListTypes 类型列表查询
// @Summary      字典类型列表
// @Description  分页查询字典类型列表，支持按状态过滤
// @Tags         基础设施-字典管理
// @Produce      json
// @Param        page      query  int  false  "页码，默认1"  minimum(1)
// @Param        page_size query  int  false  "每页数量，默认20"  maximum(100)
// @Param        status    query  int  false  "状态过滤：-1=全部 0=禁用 1=启用"
// @Success      200  {object}  resp.PaginatedResp{list=[]resp.DictTypeResp}  "字典类型列表"
// @Failure      400  {object}  docs.APIError  "请求参数错误"
// @Security     ApiKeyAuth
// @Router       /infra/dict-types [get]
func (c *DictController) ListTypes(ctx *gin.Context) {
	var listReq req.DictTypeListReq
	if err := ctx.ShouldBindQuery(&listReq); err != nil {
		utils.FailWithMsg(ctx, http.StatusBadRequest, 2, err.Error())
		return
	}
	if err := validator.ValidateRequest(listReq); err != nil {
		utils.ValidationFail(ctx, err)
		return
	}

	page := listReq.Page
	if page < 1 {
		page = 1
	}
	pageSize := listReq.PageSize
	if pageSize < 1 {
		pageSize = 20
	}
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

// ----- 字典条目 -----

// AddEntry 添加字典条目
// @Summary      添加字典条目
// @Description  向指定字典类型添加新的字典条目
// @Tags         基础设施-字典管理
// @Accept       json
// @Produce      json
// @Param        request  body  req.DictEntryAddReq  true  "添加字典条目请求参数"
// @Success      200  {object}  resp.DictTypeResp  "添加成功，返回包含条目的字典类型"
// @Failure      400  {object}  docs.APIError      "请求参数错误"
// @Failure      404  {object}  docs.APIError      "字典类型不存在"
// @Security     ApiKeyAuth
// @Router       /infra/dict-entries [post]
func (c *DictController) AddEntry(ctx *gin.Context) {
	var addReq req.DictEntryAddReq
	if err := ctx.ShouldBindJSON(&addReq); err != nil {
		utils.FailWithMsg(ctx, http.StatusBadRequest, 2, err.Error())
		return
	}
	if err := validator.ValidateRequest(addReq); err != nil {
		utils.ValidationFail(ctx, err)
		return
	}

	cmd := &command.AddDictEntryCommand{
		TypeID:    addReq.TypeID,
		Label:     addReq.Label,
		Value:     addReq.Value,
		SortOrder: addReq.SortOrder,
	}

	result, err := c.handler.AddEntry(ctx.Request.Context(), cmd)
	if err != nil {
		utils.Error(ctx, err)
		return
	}
	utils.Success(ctx, toDictTypeResp(result))
}

// UpdateEntry 更新字典条目
// @Summary      更新字典条目
// @Description  更新指定字典条目的标签、值和排序
// @Tags         基础设施-字典管理
// @Accept       json
// @Produce      json
// @Param        id       path  string                true  "字典条目ID"
// @Param        request  body  req.DictEntryUpdateReq  true  "更新字典条目请求参数"
// @Success      200  {object}  resp.DictTypeResp  "更新成功，返回包含条目的字典类型"
// @Failure      400  {object}  docs.APIError      "请求参数错误"
// @Failure      404  {object}  docs.APIError      "字典条目不存在"
// @Security     ApiKeyAuth
// @Router       /infra/dict-entries/{id} [put]
func (c *DictController) UpdateEntry(ctx *gin.Context) {
	id := ctx.Param("id")

	var updateReq req.DictEntryUpdateReq
	if err := ctx.ShouldBindJSON(&updateReq); err != nil {
		utils.FailWithMsg(ctx, http.StatusBadRequest, 2, err.Error())
		return
	}
	if err := validator.ValidateRequest(updateReq); err != nil {
		utils.ValidationFail(ctx, err)
		return
	}

	cmd := &command.UpdateDictEntryCommand{
		ID:        id,
		Label:     updateReq.Label,
		Value:     updateReq.Value,
		SortOrder: updateReq.SortOrder,
	}

	result, err := c.handler.UpdateEntry(ctx.Request.Context(), cmd)
	if err != nil {
		utils.Error(ctx, err)
		return
	}
	utils.Success(ctx, toDictTypeResp(result))
}

// RemoveEntry 移除字典条目
// @Summary      删除字典条目
// @Description  删除指定的字典条目
// @Tags         基础设施-字典管理
// @Produce      json
// @Param        id  path  string  true  "字典条目ID"
// @Success      200  {object}  resp.DictTypeResp  "删除成功，返回更新后的字典类型"
// @Failure      404  {object}  docs.APIError      "字典条目不存在"
// @Security     ApiKeyAuth
// @Router       /infra/dict-entries/{id} [delete]
func (c *DictController) RemoveEntry(ctx *gin.Context) {
	id := ctx.Param("id")
	result, err := c.handler.RemoveEntry(ctx.Request.Context(), &command.RemoveDictEntryCommand{ID: id})
	if err != nil {
		utils.Error(ctx, err)
		return
	}
	utils.Success(ctx, toDictTypeResp(result))
}

// EnableEntry 启用条目
// @Summary      启用字典条目
// @Description  启用指定的字典条目
// @Tags         基础设施-字典管理
// @Produce      json
// @Param        id  path  string  true  "字典条目ID"
// @Success      200  {object}  resp.DictTypeResp  "启用成功"
// @Failure      404  {object}  docs.APIError      "字典条目不存在"
// @Security     ApiKeyAuth
// @Router       /infra/dict-entries/{id}/enable [put]
func (c *DictController) EnableEntry(ctx *gin.Context) {
	id := ctx.Param("id")
	result, err := c.handler.EnableEntry(ctx.Request.Context(), &command.EnableDictEntryCommand{ID: id})
	if err != nil {
		utils.Error(ctx, err)
		return
	}
	utils.Success(ctx, toDictTypeResp(result))
}

// DisableEntry 禁用条目
// @Summary      禁用字典条目
// @Description  禁用指定的字典条目
// @Tags         基础设施-字典管理
// @Produce      json
// @Param        id  path  string  true  "字典条目ID"
// @Success      200  {object}  resp.DictTypeResp  "禁用成功"
// @Failure      404  {object}  docs.APIError      "字典条目不存在"
// @Security     ApiKeyAuth
// @Router       /infra/dict-entries/{id}/disable [put]
func (c *DictController) DisableEntry(ctx *gin.Context) {
	id := ctx.Param("id")
	result, err := c.handler.DisableEntry(ctx.Request.Context(), &command.DisableDictEntryCommand{ID: id})
	if err != nil {
		utils.Error(ctx, err)
		return
	}
	utils.Success(ctx, toDictTypeResp(result))
}

// GetEntryByID 按 ID 查询条目
// @Summary      按ID查询字典条目
// @Description  根据ID获取字典条目详情
// @Tags         基础设施-字典管理
// @Produce      json
// @Param        id  path  string  true  "字典条目ID"
// @Success      200  {object}  resp.DictEntryResp  "字典条目详情"
// @Failure      404  {object}  docs.APIError       "字典条目不存在"
// @Security     ApiKeyAuth
// @Router       /infra/dict-entries/{id} [get]
func (c *DictController) GetEntryByID(ctx *gin.Context) {
	id := ctx.Param("id")
	result, err := c.queryService.GetEntryByID(ctx.Request.Context(), id)
	if err != nil {
		utils.Error(ctx, err)
		return
	}
	utils.Success(ctx, toDictEntryRespFromDTO(result))
}

// ListEntries 按类型 ID 查询条目列表
// @Summary      查询字典条目列表
// @Description  根据字典类型ID获取所有字典条目
// @Tags         基础设施-字典管理
// @Produce      json
// @Param        type_id  query  string  true  "字典类型ID"
// @Success      200  {array}   resp.DictEntryResp  "字典条目列表"
// @Failure      400  {object}  docs.APIError       "请求参数错误"
// @Security     ApiKeyAuth
// @Router       /infra/dict-entries [get]
func (c *DictController) ListEntries(ctx *gin.Context) {
	typeID := ctx.Query("type_id")
	if typeID == "" {
		utils.FailWithMsg(ctx, http.StatusBadRequest, 2, "type_id is required")
		return
	}

	result, err := c.queryService.GetEntriesByTypeID(ctx.Request.Context(), typeID)
	if err != nil {
		utils.Error(ctx, err)
		return
	}

	entryResp := make([]*resp.DictEntryResp, 0, len(result))
	for _, entry := range result {
		entryResp = append(entryResp, toDictEntryRespFromDTO(entry))
	}
	utils.Success(ctx, entryResp)
}

// GetEntriesByCode 按 typeCode 获取已启用条目
// @Summary      按编码获取可用条目
// @Description  根据字典类型编码获取已启用的字典条目列表（前端枚举值获取，优先使用缓存）
// @Tags         基础设施-字典管理
// @Produce      json
// @Param        code  path  string  true  "字典类型编码"
// @Success      200  {array}   resp.DictEntryResp  "已启用的字典条目列表"
// @Failure      404  {object}  docs.APIError       "字典类型不存在或无可用条目"
// @Security     ApiKeyAuth
// @Router       /infra/dicts/entries/{code} [get]
func (c *DictController) GetEntriesByCode(ctx *gin.Context) {
	code := ctx.Param("code")
	entries, err := c.queryService.GetEntriesByCode(ctx.Request.Context(), code)
	if err != nil {
		utils.Error(ctx, err)
		return
	}
	utils.Success(ctx, entries)
}

// toDictTypeResp 将应用层 DTO 转换为 HTTP 响应
func toDictTypeResp(dto *commandHandler.DictTypeResult) *resp.DictTypeResp {
	r := &resp.DictTypeResp{
		ID:          dto.ID,
		Code:        dto.Code,
		Name:        dto.Name,
		Description: dto.Description,
		Status:      dto.Status,
		CreatedAt:   dto.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   dto.UpdatedAt.Format(time.RFC3339),
	}
	if len(dto.Entries) > 0 {
		r.Entries = make([]*resp.DictEntryResp, 0, len(dto.Entries))
		for _, e := range dto.Entries {
			r.Entries = append(r.Entries, &resp.DictEntryResp{
				ID:        e.ID,
				Label:     e.Label,
				Value:     e.Value,
				SortOrder: e.SortOrder,
				Status:    e.Status,
				CreatedAt: e.CreatedAt.Format(time.RFC3339),
				UpdatedAt: e.UpdatedAt.Format(time.RFC3339),
			})
		}
	}
	return r
}

// toDictTypeRespFromDTO 将读模型 DTO 转换为 HTTP 响应
func toDictTypeRespFromDTO(dto *repo.DictTypeDTO) *resp.DictTypeResp {
	return &resp.DictTypeResp{
		ID:          dto.ID,
		Code:        dto.Code,
		Name:        dto.Name,
		Description: dto.Description,
		Status:      dto.Status,
		CreatedAt:   dto.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   dto.UpdatedAt.Format(time.RFC3339),
	}
}

// toDictEntryRespFromDTO 将读模型 DTO 转换为 HTTP 响应
func toDictEntryRespFromDTO(dto *repo.DictEntryDTO) *resp.DictEntryResp {
	return &resp.DictEntryResp{
		ID:        dto.ID,
		TypeID:    dto.TypeID,
		Label:     dto.Label,
		Value:     dto.Value,
		SortOrder: dto.SortOrder,
		Status:    dto.Status,
		CreatedAt: dto.CreatedAt.Format(time.RFC3339),
		UpdatedAt: dto.UpdatedAt.Format(time.RFC3339),
	}
}
