package commandHandler

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/go-ddd-seed/go-ddd-seed/internal/application/command"
	domainRepo "github.com/go-ddd-seed/go-ddd-seed/internal/domain/port/repo"
	"github.com/go-ddd-seed/go-ddd-seed/internal/domain/model/kv_config"
)

// KvConfigResult 键值配置操作结果（应用层 DTO，控制器只依赖此类型）
type KvConfigResult struct {
	ID          string    `json:"id"`
	Key         string    `json:"key"`
	Value       string    `json:"value"`
	Description string    `json:"description"`
	Status      int       `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// KvConfigCommandHandler 键值配置命令处理器
type KvConfigCommandHandler struct {
	repo domainRepo.KvConfigRepository
}

// NewKvConfigCommandHandler 创建键值配置命令处理器
func NewKvConfigCommandHandler(repo domainRepo.KvConfigRepository) *KvConfigCommandHandler {
	return &KvConfigCommandHandler{repo: repo}
}

// Create 创建配置
func (h *KvConfigCommandHandler) Create(ctx context.Context, cmd *command.CreateKvConfigCommand) (*KvConfigResult, error) {
	// 检查 key 唯一性（通过写仓储加载领域对象）
	existing, _ := h.repo.FindByKey(ctx, cmd.Key)
	if existing != nil {
		return nil, ErrKvConfigKeyDuplicate
	}

	config, err := kv_config.NewKvConfig(uuid.New().String(), cmd.Key, cmd.Value, cmd.Description)
	if err != nil {
		return nil, err
	}

	if err := h.repo.Save(ctx, config); err != nil {
		return nil, err
	}
	return toKvConfigResult(config), nil
}

// Update 更新配置
func (h *KvConfigCommandHandler) Update(ctx context.Context, cmd *command.UpdateKvConfigCommand) (*KvConfigResult, error) {
	// 通过写仓储加载完整聚合，以正确支持领域事件
	config, err := h.repo.FindByID(ctx, cmd.ID)
	if err != nil {
		return nil, err
	}

	config.Update(cmd.Value, cmd.Description)
	if err := h.repo.Save(ctx, config); err != nil {
		return nil, err
	}
	return toKvConfigResult(config), nil
}

// toKvConfigResult 聚合转应用层 DTO
func toKvConfigResult(config *kv_config.KvConfig) *KvConfigResult {
	return &KvConfigResult{
		ID:          config.ID,
		Key:         config.Key,
		Value:       config.Value,
		Description: config.Description,
		Status:      config.Status,
		CreatedAt:   config.CreatedAt,
		UpdatedAt:   config.UpdatedAt,
	}
}

// Delete 删除配置
func (h *KvConfigCommandHandler) Delete(ctx context.Context, cmd *command.DeleteKvConfigCommand) error {
	// 先检查是否存在（通过写仓储加载领域对象）
	if _, err := h.repo.FindByID(ctx, cmd.ID); err != nil {
		return err
	}
	return h.repo.Delete(ctx, cmd.ID)
}