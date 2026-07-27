package controller

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-ddd-seed/go-ddd-seed/internal/application/port/repo"
	"github.com/go-ddd-seed/go-ddd-seed/internal/application/command"
	"github.com/go-ddd-seed/go-ddd-seed/internal/application/commandHandler"
	"github.com/go-ddd-seed/go-ddd-seed/internal/application/queryService"
	"github.com/go-ddd-seed/go-ddd-seed/internal/trigger/http/req"
	"github.com/go-ddd-seed/go-ddd-seed/internal/trigger/http/resp"
	"github.com/go-ddd-seed/go-ddd-seed/pkg/utils"
	"github.com/go-ddd-seed/go-ddd-seed/pkg/validator"

	_ "github.com/go-ddd-seed/go-ddd-seed/docs" // for swaggo type references in annotations
)

// RbacController RBAC 权限管理控制器
type RbacController struct {
	handler *commandHandler.RbacCommandHandler
	roleSvc *queryService.RoleQueryService
	permSvc *queryService.PermissionQueryService
	userSvc *queryService.UserQueryService
}

// NewRbacController 创建 RBAC 控制器
func NewRbacController(
	handler *commandHandler.RbacCommandHandler,
	roleSvc *queryService.RoleQueryService,
	permSvc *queryService.PermissionQueryService,
	userSvc *queryService.UserQueryService,
) *RbacController {
	return &RbacController{
		handler: handler,
		roleSvc: roleSvc,
		permSvc: permSvc,
		userSvc: userSvc,
	}
}

// RegisterRoutes 注册路由（需要在 Auth 中间件之后注册）
func (c *RbacController) RegisterRoutes(rg *gin.RouterGroup) {
	// 角色管理
	roles := rg.Group("/roles")
	roles.POST("", c.CreateRole)
	roles.PUT("/:id", c.UpdateRole)
	roles.DELETE("/:id", c.DeleteRole)
	roles.PUT("/:id/enable", c.EnableRole)
	roles.PUT("/:id/disable", c.DisableRole)
	roles.GET("/:id", c.GetRoleByID)
	roles.GET("", c.ListRoles)

	// 角色-权限绑定
	roles.POST("/:id/permissions", c.AssignPermissions)
	roles.DELETE("/:id/permissions/:permId", c.RemovePermission)

	// 权限管理
	perms := rg.Group("/permissions")
	perms.POST("", c.CreatePermission)
	perms.PUT("/:id", c.UpdatePermission)
	perms.DELETE("/:id", c.DeletePermission)
	perms.GET("", c.ListPermissions)

	// 用户-角色绑定
	users := rg.Group("/users")
	users.POST("/:id/roles", c.AssignRole)
	users.DELETE("/:id/roles/:roleId", c.RemoveRole)
}

// ----- 角色管理 -----

// CreateRole 创建角色
// @Summary      创建角色
// @Description  创建新的角色定义
// @Tags         权限管理-角色
// @Accept       json
// @Produce      json
// @Param        request  body  req.CreateRoleReq  true  "创建角色请求参数"
// @Success      200  {object}  resp.RoleResp  "创建成功"
// @Failure      400  {object}  docs.APIError  "请求参数错误"
// @Security     ApiKeyAuth
// @Router       /roles [post]
func (c *RbacController) CreateRole(ctx *gin.Context) {
	var createReq req.CreateRoleReq
	if err := ctx.ShouldBindJSON(&createReq); err != nil {
		utils.FailWithMsg(ctx, http.StatusBadRequest, 2, err.Error())
		return
	}
	if err := validator.ValidateRequest(createReq); err != nil {
		utils.ValidationFail(ctx, err)
		return
	}

	cmd := &command.CreateRoleCommand{
		Name:        createReq.Name,
		Code:        createReq.Code,
		Description: createReq.Description,
	}

	result, err := c.handler.CreateRole(ctx.Request.Context(), cmd)
	if err != nil {
		utils.Error(ctx, err)
		return
	}
	utils.Success(ctx, toRoleResp(result))
}

// UpdateRole 更新角色
// @Summary      更新角色
// @Description  更新指定角色的名称和描述
// @Tags         权限管理-角色
// @Accept       json
// @Produce      json
// @Param        id       path  string            true  "角色ID"
// @Param        request  body  req.UpdateRoleReq  true  "更新角色请求参数"
// @Success      200  {object}  resp.RoleResp  "更新成功"
// @Failure      400  {object}  docs.APIError  "请求参数错误"
// @Failure      404  {object}  docs.APIError  "角色不存在"
// @Security     ApiKeyAuth
// @Router       /roles/{id} [put]
func (c *RbacController) UpdateRole(ctx *gin.Context) {
	id := ctx.Param("id")

	var updateReq req.UpdateRoleReq
	if err := ctx.ShouldBindJSON(&updateReq); err != nil {
		utils.FailWithMsg(ctx, http.StatusBadRequest, 2, err.Error())
		return
	}
	if err := validator.ValidateRequest(updateReq); err != nil {
		utils.ValidationFail(ctx, err)
		return
	}

	cmd := &command.UpdateRoleCommand{
		ID:          id,
		Name:        updateReq.Name,
		Description: updateReq.Description,
	}

	result, err := c.handler.UpdateRole(ctx.Request.Context(), cmd)
	if err != nil {
		utils.Error(ctx, err)
		return
	}
	utils.Success(ctx, toRoleResp(result))
}

// DeleteRole 删除角色
// @Summary      删除角色
// @Description  删除指定的角色
// @Tags         权限管理-角色
// @Produce      json
// @Param        id  path  string  true  "角色ID"
// @Success      200  {object}  docs.APIError  "删除成功"
// @Failure      404  {object}  docs.APIError  "角色不存在"
// @Security     ApiKeyAuth
// @Router       /roles/{id} [delete]
func (c *RbacController) DeleteRole(ctx *gin.Context) {
	id := ctx.Param("id")

	if err := c.handler.DeleteRole(ctx.Request.Context(), &command.DeleteRoleCommand{ID: id}); err != nil {
		utils.Error(ctx, err)
		return
	}
	utils.Success(ctx, nil)
}

// EnableRole 启用角色
// @Summary      启用角色
// @Description  启用指定的角色
// @Tags         权限管理-角色
// @Produce      json
// @Param        id  path  string  true  "角色ID"
// @Success      200  {object}  resp.RoleResp  "启用成功"
// @Failure      404  {object}  docs.APIError  "角色不存在"
// @Security     ApiKeyAuth
// @Router       /roles/{id}/enable [put]
func (c *RbacController) EnableRole(ctx *gin.Context) {
	id := ctx.Param("id")

	result, err := c.handler.EnableRole(ctx.Request.Context(), &command.EnableRoleCommand{ID: id})
	if err != nil {
		utils.Error(ctx, err)
		return
	}
	utils.Success(ctx, result)
}

// DisableRole 禁用角色
// @Summary      禁用角色
// @Description  禁用指定的角色
// @Tags         权限管理-角色
// @Produce      json
// @Param        id  path  string  true  "角色ID"
// @Success      200  {object}  resp.RoleResp  "禁用成功"
// @Failure      404  {object}  docs.APIError  "角色不存在"
// @Security     ApiKeyAuth
// @Router       /roles/{id}/disable [put]
func (c *RbacController) DisableRole(ctx *gin.Context) {
	id := ctx.Param("id")

	result, err := c.handler.DisableRole(ctx.Request.Context(), &command.DisableRoleCommand{ID: id})
	if err != nil {
		utils.Error(ctx, err)
		return
	}
	utils.Success(ctx, result)
}

// GetRoleByID 按 ID 查询角色
// @Summary      按ID查询角色
// @Description  根据ID获取角色详情
// @Tags         权限管理-角色
// @Produce      json
// @Param        id  path  string  true  "角色ID"
// @Success      200  {object}  resp.RoleResp  "角色详情"
// @Failure      404  {object}  docs.APIError  "角色不存在"
// @Security     ApiKeyAuth
// @Router       /roles/{id} [get]
func (c *RbacController) GetRoleByID(ctx *gin.Context) {
	id := ctx.Param("id")

	result, err := c.roleSvc.GetByID(ctx.Request.Context(), id)
	if err != nil {
		utils.Error(ctx, err)
		return
	}
	utils.Success(ctx, result)
}

// ListRoles 角色列表
// @Summary      角色列表
// @Description  查询所有角色列表
// @Tags         权限管理-角色
// @Produce      json
// @Success      200  {array}   resp.RoleResp  "角色列表"
// @Security     ApiKeyAuth
// @Router       /roles [get]
func (c *RbacController) ListRoles(ctx *gin.Context) {
	result, err := c.roleSvc.List(ctx.Request.Context())
	if err != nil {
		utils.Error(ctx, err)
		return
	}
	utils.Success(ctx, result)
}

// ----- 权限管理 -----

// CreatePermission 创建权限
// @Summary      创建权限
// @Description  创建新的权限定义
// @Tags         权限管理-权限
// @Accept       json
// @Produce      json
// @Param        request  body  req.CreatePermissionReq  true  "创建权限请求参数"
// @Success      200  {object}  resp.PermissionResp  "创建成功"
// @Failure      400  {object}  docs.APIError        "请求参数错误"
// @Security     ApiKeyAuth
// @Router       /permissions [post]
func (c *RbacController) CreatePermission(ctx *gin.Context) {
	var createReq req.CreatePermissionReq
	if err := ctx.ShouldBindJSON(&createReq); err != nil {
		utils.FailWithMsg(ctx, http.StatusBadRequest, 2, err.Error())
		return
	}
	if err := validator.ValidateRequest(createReq); err != nil {
		utils.ValidationFail(ctx, err)
		return
	}

	cmd := &command.CreatePermissionCommand{
		Name:        createReq.Name,
		Code:        createReq.Code,
		Description: createReq.Description,
	}

	result, err := c.handler.CreatePermission(ctx.Request.Context(), cmd)
	if err != nil {
		utils.Error(ctx, err)
		return
	}
	utils.Success(ctx, toPermissionResp(result))
}

// UpdatePermission 更新权限
// @Summary      更新权限
// @Description  更新指定权限的名称和描述
// @Tags         权限管理-权限
// @Accept       json
// @Produce      json
// @Param        id       path  string                 true  "权限ID"
// @Param        request  body  req.UpdatePermissionReq  true  "更新权限请求参数"
// @Success      200  {object}  resp.PermissionResp  "更新成功"
// @Failure      400  {object}  docs.APIError        "请求参数错误"
// @Failure      404  {object}  docs.APIError        "权限不存在"
// @Security     ApiKeyAuth
// @Router       /permissions/{id} [put]
func (c *RbacController) UpdatePermission(ctx *gin.Context) {
	id := ctx.Param("id")

	var updateReq req.UpdatePermissionReq
	if err := ctx.ShouldBindJSON(&updateReq); err != nil {
		utils.FailWithMsg(ctx, http.StatusBadRequest, 2, err.Error())
		return
	}
	if err := validator.ValidateRequest(updateReq); err != nil {
		utils.ValidationFail(ctx, err)
		return
	}

	cmd := &command.UpdatePermissionCommand{
		ID:          id,
		Name:        updateReq.Name,
		Description: updateReq.Description,
	}

	result, err := c.handler.UpdatePermission(ctx.Request.Context(), cmd)
	if err != nil {
		utils.Error(ctx, err)
		return
	}
	utils.Success(ctx, toPermissionResp(result))
}

// DeletePermission 删除权限
// @Summary      删除权限
// @Description  删除指定的权限
// @Tags         权限管理-权限
// @Produce      json
// @Param        id  path  string  true  "权限ID"
// @Success      200  {object}  docs.APIError  "删除成功"
// @Failure      404  {object}  docs.APIError  "权限不存在"
// @Security     ApiKeyAuth
// @Router       /permissions/{id} [delete]
func (c *RbacController) DeletePermission(ctx *gin.Context) {
	id := ctx.Param("id")

	if err := c.handler.DeletePermission(ctx.Request.Context(), &command.DeletePermissionCommand{ID: id}); err != nil {
		utils.Error(ctx, err)
		return
	}
	utils.Success(ctx, nil)
}

// ListPermissions 权限列表
// @Summary      权限列表
// @Description  查询所有权限列表
// @Tags         权限管理-权限
// @Produce      json
// @Success      200  {array}   resp.PermissionResp  "权限列表"
// @Security     ApiKeyAuth
// @Router       /permissions [get]
func (c *RbacController) ListPermissions(ctx *gin.Context) {
	result, err := c.permSvc.List(ctx.Request.Context())
	if err != nil {
		utils.Error(ctx, err)
		return
	}
	utils.Success(ctx, result)
}

// ----- 角色-权限绑定 -----

// AssignPermissions 为角色批量分配权限
// @Summary      为角色分配权限
// @Description  为指定角色批量分配权限
// @Tags         权限管理-角色
// @Accept       json
// @Produce      json
// @Param        id       path  string                  true  "角色ID"
// @Param        request  body  req.AssignPermissionReq  true  "分配权限请求参数"
// @Success      200  {object}  docs.APIError  "分配成功"
// @Failure      400  {object}  docs.APIError  "请求参数错误"
// @Failure      404  {object}  docs.APIError  "角色或权限不存在"
// @Security     ApiKeyAuth
// @Router       /roles/{id}/permissions [post]
func (c *RbacController) AssignPermissions(ctx *gin.Context) {
	roleID := ctx.Param("id")

	var assignReq req.AssignPermissionReq
	if err := ctx.ShouldBindJSON(&assignReq); err != nil {
		utils.FailWithMsg(ctx, http.StatusBadRequest, 2, err.Error())
		return
	}
	if err := validator.ValidateRequest(assignReq); err != nil {
		utils.ValidationFail(ctx, err)
		return
	}

	result, err := c.handler.AssignPermissionToRole(ctx.Request.Context(), &command.AssignRolePermissionCommand{
		RoleID:        roleID,
		PermissionIDs: assignReq.PermissionIDs,
	})
	if err != nil {
		utils.Error(ctx, err)
		return
	}
	utils.Success(ctx, result)
}

// RemovePermission 移除角色权限
// @Summary      移除角色权限
// @Description  从指定角色中移除特定权限
// @Tags         权限管理-角色
// @Produce      json
// @Param        id      path  string  true  "角色ID"
// @Param        permId  path  string  true  "权限ID"
// @Success      200  {object}  docs.APIError  "移除成功"
// @Failure      404  {object}  docs.APIError  "角色或权限不存在"
// @Security     ApiKeyAuth
// @Router       /roles/{id}/permissions/{permId} [delete]
func (c *RbacController) RemovePermission(ctx *gin.Context) {
	roleID := ctx.Param("id")
	permID := ctx.Param("permId")

	result, err := c.handler.RemovePermissionFromRole(ctx.Request.Context(), &command.RemoveRolePermissionCommand{
		RoleID:       roleID,
		PermissionID: permID,
	})
	if err != nil {
		utils.Error(ctx, err)
		return
	}
	utils.Success(ctx, result)
}

// ----- 用户-角色绑定 -----

// AssignRole 为用户分配角色
// @Summary      为用户分配角色
// @Description  为指定用户分配角色
// @Tags         权限管理-用户角色
// @Accept       json
// @Produce      json
// @Param        id       path  string            true  "用户ID"
// @Param        request  body  req.AssignRoleReq  true  "分配角色请求参数"
// @Success      200  {object}  docs.APIError  "分配成功"
// @Failure      400  {object}  docs.APIError  "请求参数错误"
// @Failure      404  {object}  docs.APIError  "用户或角色不存在"
// @Security     ApiKeyAuth
// @Router       /users/{id}/roles [post]
func (c *RbacController) AssignRole(ctx *gin.Context) {
	userID := ctx.Param("id")

	var assignReq req.AssignRoleReq
	if err := ctx.ShouldBindJSON(&assignReq); err != nil {
		utils.FailWithMsg(ctx, http.StatusBadRequest, 2, err.Error())
		return
	}
	if err := validator.ValidateRequest(assignReq); err != nil {
		utils.ValidationFail(ctx, err)
		return
	}

	if err := c.handler.AssignRoleToUser(ctx.Request.Context(), &command.AssignUserRoleCommand{
		UserID: userID,
		RoleID: assignReq.RoleID,
	}); err != nil {
		utils.Error(ctx, err)
		return
	}
	utils.Success(ctx, nil)
}

// RemoveRole 移除用户角色
// @Summary      移除用户角色
// @Description  从指定用户中移除特定角色
// @Tags         权限管理-用户角色
// @Produce      json
// @Param        id      path  string  true  "用户ID"
// @Param        roleId  path  string  true  "角色ID"
// @Success      200  {object}  docs.APIError  "移除成功"
// @Failure      404  {object}  docs.APIError  "用户或角色不存在"
// @Security     ApiKeyAuth
// @Router       /users/{id}/roles/{roleId} [delete]
func (c *RbacController) RemoveRole(ctx *gin.Context) {
	userID := ctx.Param("id")
	roleID := ctx.Param("roleId")

	if err := c.handler.RemoveRoleFromUser(ctx.Request.Context(), &command.RemoveUserRoleCommand{
		UserID: userID,
		RoleID: roleID,
	}); err != nil {
		utils.Error(ctx, err)
		return
	}
	utils.Success(ctx, nil)
}
// toRoleResp 将应用层 DTO 转换为 HTTP 响应
func toRoleResp(dto *repo.RoleDTO) *resp.RoleResp {
	return &resp.RoleResp{
		ID:          dto.ID,
		Name:        dto.Name,
		Code:        dto.Code,
		Description: dto.Description,
		Status:      dto.Status,
		CreatedAt:   dto.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   dto.UpdatedAt.Format(time.RFC3339),
	}
}

// toPermissionResp 将应用层 DTO 转换为 HTTP 响应
func toPermissionResp(dto *repo.PermissionDTO) *resp.PermissionResp {
	return &resp.PermissionResp{
		ID:          dto.ID,
		Name:        dto.Name,
		Code:        dto.Code,
		Description: dto.Description,
		CreatedAt:   dto.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   dto.UpdatedAt.Format(time.RFC3339),
	}
}
