package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-ddd-seed/go-ddd-seed/pkg/logger"
)

// Logger 请求日志中间件（基于 zerolog 结构化日志）
// 记录每个请求的方法、路径、状态码、IP 和耗时
// 自动注入 trace_id（需先注册 Trace 中间件）
// 支持慢请求检测：超过阈值自动记录 Warn 级别
func Logger() gin.HandlerFunc {
	// 从全局默认 Logger 获取
	l := logger.L()

	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// 处理请求
		c.Next()

		latency := time.Since(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()

		if raw != "" {
			path = path + "?" + raw
		}

		// 获取 trace_id
		traceID := GetTraceID(c)

		// 构建日志字段
		fields := map[string]any{
			"http_status":   statusCode,
			"http_method":   method,
			"http_path":     path,
			"client_ip":     clientIP,
			"duration_ms":   latency.Milliseconds(),
			"latency_human": latency.String(),
		}

		if traceID != "" {
			fields["trace_id"] = traceID
		}

		// 获取 body 大小（如有）
		if c.Writer.Size() >= 0 {
			fields["body_size"] = c.Writer.Size()
		}

		// 根据状态码选择日志级别
		msg := "HTTP Request"
		switch {
		case statusCode >= 500:
			l.WithContext(c.Request.Context()).Error(msg, nil, fields)
		case statusCode >= 400:
			l.WithContext(c.Request.Context()).Warn(msg, fields)
		case latency > 500*time.Millisecond:
			// 慢请求（>500ms）提升为 Warn，不管状态码
			fields["slow_request"] = true
			l.WithContext(c.Request.Context()).Warn(msg+" (slow)", fields)
		default:
			l.WithContext(c.Request.Context()).Info(msg, fields)
		}
	}
}