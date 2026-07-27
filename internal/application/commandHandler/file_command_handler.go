package commandHandler

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"path/filepath"

	"github.com/google/uuid"

	appApi "github.com/go-ddd-seed/go-ddd-seed/internal/application/port/thirdPartyApi"
	"github.com/go-ddd-seed/go-ddd-seed/internal/application/command"
	domainRepo "github.com/go-ddd-seed/go-ddd-seed/internal/domain/port/repo"
	"github.com/go-ddd-seed/go-ddd-seed/internal/domain/model/file_record"
)

// FileCommandHandler 文件命令处理器
type FileCommandHandler struct {
	repo      domainRepo.FileRecordRepository
	fileStore appApi.FileStorage
	channel   string // 当前存储渠道标识
}

// NewFileCommandHandler 创建文件命令处理器
func NewFileCommandHandler(
	repo domainRepo.FileRecordRepository,
	fileStore appApi.FileStorage,
	channel string,
) *FileCommandHandler {
	return &FileCommandHandler{
		repo:      repo,
		fileStore: fileStore,
		channel:   channel,
	}
}

// FileUploadResult 文件上传结果（应用层 DTO）
type FileUploadResult struct {
	ID           string `json:"id"`
	FileName     string `json:"file_name"`
	Size         int64  `json:"size"`
	MIMEType     string `json:"mime_type"`
	AccessURL    string `json:"access_url"`
	ThumbnailURL string `json:"thumbnail_url,omitempty"`
}

// Upload 上传文件
func (h *FileCommandHandler) Upload(ctx context.Context, cmd *command.UploadFileCommand) (*FileUploadResult, error) {
	// 1. 校验文件大小
	if cmd.FileSize <= 0 {
		return nil, ErrFileSizeInvalid
	}

	// 2. 计算 MD5 校验
	// 先读取内容到 buffer 以便计算 MD5 和后续上传
	var contentBytes []byte
	var err error

	// 尝试从 io.Reader 读取全部内容
	if seeker, ok := cmd.Content.(io.ReadSeeker); ok {
		contentBytes, err = io.ReadAll(seeker)
		if err != nil {
			return nil, fmt.Errorf("读取文件内容失败: %w", err)
		}
		// 重置读取位置以便后续上传
		seeker.Seek(0, io.SeekStart)
		cmd.Content = seeker
	} else {
		contentBytes, err = io.ReadAll(cmd.Content)
		if err != nil {
			return nil, fmt.Errorf("读取文件内容失败: %w", err)
		}
		// 无法重置，使用 bytes.Reader
		cmd.Content = &bytesReader{data: contentBytes}
	}

	md5Hash := computeMD5(contentBytes)

	// 3. 生成存储路径
	storagePath := generateStoragePath(cmd.AttachType, cmd.AttachID, cmd.FileName)

	// 4. 上传文件到存储后端
	if err := h.fileStore.Upload(ctx, cmd.Content, storagePath); err != nil {
		return nil, fmt.Errorf("上传文件到存储后端失败: %w", err)
	}

	// 5. 创建 FileRecord 聚合根
	id := uuid.New().String()
	fr, err := file_record.NewFileRecord(
		id,
		cmd.FileName,
		storagePath,
		cmd.MIMEType,
		h.channel,
		md5Hash,
		cmd.UploaderID,
		cmd.FileSize,
	)
	if err != nil {
		return nil, err
	}

	// 6. 关联业务对象（可选）
	if cmd.AttachType != "" && cmd.AttachID != "" {
		if err := fr.Associate(cmd.AttachType, cmd.AttachID); err != nil {
			return nil, err
		}
	}

	// 7. 保存到数据库
	if err := h.repo.Save(ctx, fr); err != nil {
		return nil, fmt.Errorf("保存文件记录失败: %w", err)
	}

	// 8. 生成访问 URL
	accessURL, _ := h.fileStore.GetURL(ctx, storagePath, 0)

	return &FileUploadResult{
		ID:        fr.ID,
		FileName:  fr.FileName,
		Size:      fr.Size,
		MIMEType:  fr.MIMEType,
		AccessURL: accessURL,
	}, nil
}

// Delete 删除文件（软删除）
func (h *FileCommandHandler) Delete(ctx context.Context, cmd *command.DeleteFileCommand) error {
	// 通过写仓储加载完整聚合
	fr, err := h.repo.FindByID(ctx, cmd.ID)
	if err != nil {
		return err
	}

	// 调用聚合根业务方法
	if err := fr.Delete(); err != nil {
		return err
	}

	// 保存（发布领域事件）
	return h.repo.Save(ctx, fr)
}

// generateStoragePath 生成存储路径
// 格式：{attach_type}/{attach_id}/{uuid}_{filename} 或 default/{uuid}_{filename}
func generateStoragePath(attachType, attachID, fileName string) string {
	prefix := uuid.New().String()
	ext := filepath.Ext(fileName)

	if attachType != "" && attachID != "" {
		return filepath.Join(attachType, attachID, prefix+ext)
	}
	return filepath.Join("default", prefix+ext)
}

// computeMD5 计算 MD5 校验
func computeMD5(data []byte) string {
	hash := md5.Sum(data)
	return hex.EncodeToString(hash[:])
}

// bytesReader 将 []byte 包装为 io.Reader
type bytesReader struct {
	data []byte
	pos  int
}

func (r *bytesReader) Read(p []byte) (n int, err error) {
	if r.pos >= len(r.data) {
		return 0, io.EOF
	}
	n = copy(p, r.data[r.pos:])
	r.pos += n
	return n, nil
}

// 文件命令处理器错误
var (
	ErrFileSizeInvalid = fmt.Errorf("文件大小无效")
)