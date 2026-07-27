package repo

import (
	"context"

	"github.com/go-ddd-seed/go-ddd-seed/internal/domain/model/file_record"
)

// FileRecordRepository 文件记录写仓储接口
// 定义在领域层端口，基础设施层实现
type FileRecordRepository interface {
	// Save 保存文件记录（整存整取），自动发布领域事件
	Save(ctx context.Context, record *file_record.FileRecord) error
	// FindByID 按 ID 加载文件记录聚合
	FindByID(ctx context.Context, id string) (*file_record.FileRecord, error)
	// FindByAttach 按业务对象查询关联文件
	FindByAttach(ctx context.Context, attachType, attachID string) ([]*file_record.FileRecord, error)
	// Delete 删除文件记录
	Delete(ctx context.Context, id string) error
}