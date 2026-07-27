package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	thirdPartyApi "github.com/go-ddd-seed/go-ddd-seed/internal/application/port/thirdPartyApi"
)

// 操作类型常量
const (
	ActionCreate = "create"
	ActionUpdate = "update"
	ActionDelete = "delete"
)

// getActionFromMethod 根据 HTTP 方法推断操作类型
func getActionFromMethod(method string) string {
	switch method {
	case http.MethodPost:
		return ActionCreate
	case http.MethodPut:
		return ActionUpdate
	case http.MethodDelete:
		return ActionDelete
	default:
		return ""
	}
}

// 业务模块路径 → 目标类型映射表
// 约定：URL 路径的第二段（/api/v1/{module}/...）作为目标类型
// 特殊映射可通过此表覆盖
var pathTargetMap = map[string]string{
	"roles":        "role",
	"permissions":  "permission",
	"users":        "user",
	"dict-types":   "dict_type",
	"dict-entries": "dict_entry",
	"kv-configs":   "kv_config",
}

// inferTargetType 从 URL 路径推断目标类型
// 例如：/api/v1/roles → role, /api/v1/dict-types → dict_type
func inferTargetType(path string) string {
	// 移除 API 前缀 /api/v1/
	trimmed := strings.TrimPrefix(path, "/api/v1/")
	// 取第一段
	parts := strings.SplitN(trimmed, "/", 2)
	if len(parts) == 0 {
		return ""
	}
	module := parts[0]
	if mapped, ok := pathTargetMap[module]; ok {
		return mapped
	}
	return module
}

// bodyCacheWriter 包装 ResponseWriter，同时缓冲响应体
type bodyCacheWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *bodyCacheWriter) Write(data []byte) (int, error) {
	w.body.Write(data)
	return w.ResponseWriter.Write(data)
}

// BodyCache 请求体缓存中间件
// 必须在 Audit 中间件之前注册
func BodyCache() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodPost ||
			c.Request.Method == http.MethodPut ||
			c.Request.Method == http.MethodDelete {
			body, _ := io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
			c.Set("audit_log_body", string(body))
		}
		c.Next()
	}
}

// AuditLogger 审计日志中间件
// 记录所有写请求（POST/PUT/DELETE）的审计信息
// 需要先注册 Auth 和 BodyCache 中间件
func AuditLogger(repo thirdPartyApi.AuditLogRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		if method != http.MethodPost && method != http.MethodPut && method != http.MethodDelete {
			c.Next()
			return
		}

		action := getActionFromMethod(method)
		if action == "" {
			c.Next()
			return
		}

		// 从 Context 获取用户信息（由 Auth 中间件注入）
		userID, _ := GetCurrentUserID(c)
		userName, _ := GetCurrentUsername(c)
		traceID := GetTraceID(c)

		// 获取请求体
		requestBody, _ := c.Get("audit_log_body")
		reqBodyStr, _ := requestBody.(string)

		// 推断目标类型
		targetType := inferTargetType(c.Request.URL.Path)
		// 获取目标 ID（对于 PUT/DELETE，从 URL 参数提取）
		targetID := c.Param("id")

		// 包装 ResponseWriter 以捕获响应体
		writer := &bodyCacheWriter{
			ResponseWriter: c.Writer,
			body:           &bytes.Buffer{},
		}
		c.Writer = writer

		c.Next()

		// 响应体捕获（仅对 POST 有用，因为 POST 的创建 ID 在响应中）
		responseBody := strings.TrimSpace(writer.body.String())
		// 尝试从响应体中提取 ID（POST 创建场景）
		if action == ActionCreate && targetID == "" {
			targetID = extractIDFromResponse(responseBody)
		}

		// 构造审计日志
		logEntry := &thirdPartyApi.AuditLogDTO{
			OperatorID:   userID,
			OperatorName: userName,
			Action:       action,
			TargetType:   targetType,
			TargetID:     targetID,
			RequestBody:  truncateString(reqBodyStr, 4096),
			ResponseBody: truncateString(responseBody, 2048),
			ClientIP:     c.ClientIP(),
			UserAgent:    c.Request.UserAgent(),
			TraceID:      traceID,
		}

		// 异步写入（不阻塞请求）
		go func() {
			_ = repo.Save(c.Request.Context(), logEntry)
		}()
	}
}

// extractIDFromResponse 尝试从响应 JSON 中提取 data.id
// 统一响应格式：{"code":0,"message":"ok","data":{"id":"xxx",...}}
func extractIDFromResponse(body string) string {
	if body == "" {
		return ""
	}
	var resp struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal([]byte(body), &resp); err != nil {
		return ""
	}
	return resp.Data.ID
}

// truncateString 截断长字符串
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen]
}
