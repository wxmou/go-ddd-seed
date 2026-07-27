// Package docs contains swagger doc-only types referenced in API annotations.
// These types are NEVER used at runtime; they exist solely to produce
// accurate OpenAPI schemas for error responses and edge cases.
package docs

// APIError represents the unified error response body.
// The actual runtime type is utils.Response, but swaggo cannot
// represent the generic {code, message, data} pattern, so this
// type documents the error shape explicitly.
type APIError struct {
	Code    int    `json:"code" example:"2"`
	Message string `json:"message" example:"请求参数错误"`
}