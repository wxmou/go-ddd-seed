package controller

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	appRepo "github.com/go-ddd-seed/go-ddd-seed/internal/application/port/repo"
	"github.com/go-ddd-seed/go-ddd-seed/internal/application/queryService"
	"github.com/go-ddd-seed/go-ddd-seed/internal/trigger/http/resp"
	"github.com/go-ddd-seed/go-ddd-seed/pkg/utils"
)

// AuditController 审计日志控制器
type AuditController struct {
	querySvc *queryService.AuditLogQueryService
}

// NewAuditController 创建审计日志控制器
func NewAuditController(querySvc *queryService.AuditLogQueryService) *AuditController {
	return &AuditController{querySvc: querySvc}
}

// RegisterRoutes 注册路由
func (c *AuditController) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/audit-logs", c.List)
}

// List 审计日志列表
// @Summary      审计日志列表
// @Description  分页查询审计日志，支持按时间、操作人、操作类型筛选
// @Tags         审计日志
// @Produce      json
// @Param        operator_id   query  string  false  "操作人ID"
// @Param        action        query  string  false  "操作类型（create/update/delete）"
// @Param        target_type   query  string  false  "目标类型"
// @Param        target_id     query  string  false  "目标ID"
// @Param        start_time    query  string  false  "开始时间（RFC3339）"
// @Param        end_time      query  string  false  "结束时间（RFC3339）"
// @Param        page          query  int     false  "页码（默认1）"
// @Param        page_size     query  int     false  "每页条数（默认20，最大100）"
// @Success      200  {object}  resp.AuditLogListResp  "审计日志列表"
// @Security     ApiKeyAuth
// @Router       /audit-logs [get]
func (c *AuditController) List(ctx *gin.Context) {
	query := appRepo.AuditLogQuery{
		OperatorID: ctx.Query("operator_id"),
		Action:     ctx.Query("action"),
		TargetType: ctx.Query("target_type"),
		TargetID:   ctx.Query("target_id"),
	}

	if st := ctx.Query("start_time"); st != "" {
		t, err := time.Parse(time.RFC3339, st)
		if err == nil {
			query.StartTime = t
		}
	}
	if et := ctx.Query("end_time"); et != "" {
		t, err := time.Parse(time.RFC3339, et)
		if err == nil {
			query.EndTime = t
		}
	}

	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("page_size", "20"))
	if page <= 0 {
		page = 1
	}
	query.Page = page
	query.PageSize = pageSize

	dtos, total, err := c.querySvc.List(ctx.Request.Context(), query)
	if err != nil {
		utils.Error(ctx, err)
		return
	}

	items := make([]resp.AuditLogItem, 0, len(dtos))
	for _, dto := range dtos {
		items = append(items, toAuditLogItem(dto))
	}

	utils.Success(ctx, resp.AuditLogListResp{
		Items: items,
		Total: total,
		Page:  page,
	})
}

// toAuditLogItem 将 AuditLogReadDTO 转换为 AuditLogItem
func toAuditLogItem(dto *appRepo.AuditLogReadDTO) resp.AuditLogItem {
	return resp.AuditLogItem{
		ID:           dto.ID,
		OperatorID:   dto.OperatorID,
		OperatorName: dto.OperatorName,
		Action:       dto.Action,
		TargetType:   dto.TargetType,
		TargetID:     dto.TargetID,
		ClientIP:     dto.ClientIP,
		UserAgent:    dto.UserAgent,
		TraceID:      dto.TraceID,
		CreatedAt:    dto.CreatedAt.Format(time.RFC3339),
	}
}