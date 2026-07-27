package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/go-ddd-seed/go-ddd-seed/pkg/utils"
)

// HealthController 健康检查控制器
type HealthController struct{}

// NewHealthController 创建健康检查控制器
func NewHealthController() *HealthController {
	return &HealthController{}
}

// Health 健康检查
// @Summary      健康检查
// @Description  检查服务是否正常运行
// @Tags         系统管理
// @Produce      json
// @Success      200  {object}  map[string]string  "服务正常"
// @Router       /health [get]
func (h *HealthController) Health(c *gin.Context) {
	utils.Success(c, gin.H{
		"status": "ok",
	})
}