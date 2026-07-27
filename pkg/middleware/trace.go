package middleware

import (
	"context"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/go-ddd-seed/go-ddd-seed/pkg/logger"
)

// Trace 请求追踪中间件
// 从请求头读取 X-Request-ID，若不存在则自动生成
// 将 trace_id 注入 gin.Context 和 request.Context
func Trace() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取或生成 Trace ID
		traceID := c.GetHeader("X-Request-ID")
		if traceID == "" {
			traceID = uuid.New().String()
		} else {
			// 去除可能的空白字符
			traceID = strings.TrimSpace(traceID)
		}

		// 注入到 gin.Context
		c.Set(string(logger.TraceIDKey), traceID)

		// 注入到 request.Context（后续链路传递）
		ctx := context.WithValue(c.Request.Context(), logger.TraceIDKey, traceID)
		c.Request = c.Request.WithContext(ctx)

		// 设置响应头（透传给客户端）
		c.Header("X-Request-ID", traceID)

		c.Next()
	}
}

// GetTraceID 从 gin.Context 获取 Trace ID
func GetTraceID(c *gin.Context) string {
	v, exists := c.Get(string(logger.TraceIDKey))
	if !exists {
		return ""
	}
	s, ok := v.(string)
	if !ok {
		return ""
	}
	return s
}