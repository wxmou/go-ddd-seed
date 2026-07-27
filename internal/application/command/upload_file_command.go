package command

import (
	"io"
)

// UploadFileCommand 上传文件命令
type UploadFileCommand struct {
	FileName   string    // 原始文件名
	FileSize   int64     // 文件大小
	MIMEType   string    // MIME 类型
	Content    io.Reader // 文件内容流
	AttachType string    // 关联业务类型（可选）
	AttachID   string    // 关联业务 ID（可选）
	UploaderID string    // 上传者 ID
}

// DeleteFileCommand 删除文件命令
type DeleteFileCommand struct {
	ID string // 文件记录 ID
}