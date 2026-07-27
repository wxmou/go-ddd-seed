package req

import "mime/multipart"

// FileUploadReq 文件上传请求
type FileUploadReq struct {
	File       *multipart.FileHeader `form:"file" binding:"required"` // 上传文件
	AttachType string                `form:"attach_type"`             // 关联业务类型（可选）
	AttachID   string                `form:"attach_id"`               // 关联业务 ID（可选）
}

// FileBatchUploadReq 批量上传请求
type FileBatchUploadReq struct {
	Files      []*multipart.FileHeader `form:"files" binding:"required"` // 上传文件列表
	AttachType string                  `form:"attach_type"`              // 关联业务类型（可选）
	AttachID   string                  `form:"attach_id"`                // 关联业务 ID（可选）
}