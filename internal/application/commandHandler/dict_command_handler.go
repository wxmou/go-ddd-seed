package commandHandler

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/go-ddd-seed/go-ddd-seed/internal/application/port/repo"
	"github.com/go-ddd-seed/go-ddd-seed/internal/application/command"
	domainRepo "github.com/go-ddd-seed/go-ddd-seed/internal/domain/port/repo"
	"github.com/go-ddd-seed/go-ddd-seed/internal/domain/model/dict_type"
)

// 应用层状态常量（镜像领域层值，避免 DTO 直接引用领域常量）
const (
	appDictStatusEnabled  = 1
	appDictStatusDisabled = 0
)

// DictTypeResult 字典类型操作结果（应用层 DTO）
type DictTypeResult struct {
	ID          string            `json:"id"`
	Code        string            `json:"code"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Status      int               `json:"status"`
	Entries     []*DictEntryResult `json:"entries,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// IsEnabled 是否启用
func (r *DictTypeResult) IsEnabled() bool {
	return r.Status == appDictStatusEnabled
}

// DictEntryResult 字典条目操作结果
type DictEntryResult struct {
	ID        string    `json:"id"`
	Label     string    `json:"label"`
	Value     string    `json:"value"`
	SortOrder int       `json:"sort_order"`
	Status    int       `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// IsEnabled 是否启用
func (r *DictEntryResult) IsEnabled() bool {
	return r.Status == appDictStatusEnabled
}

// DictCommandHandler 字典命令处理器
// 处理字典类型和字典条目的所有写操作
type DictCommandHandler struct {
	dictRepo  domainRepo.DictTypeRepository
	cacheRepo repo.DictCacheRepository
}

// NewDictCommandHandler 创建字典命令处理器
func NewDictCommandHandler(
	dictRepo domainRepo.DictTypeRepository,
	cacheRepo repo.DictCacheRepository,
) *DictCommandHandler {
	return &DictCommandHandler{
		dictRepo:  dictRepo,
		cacheRepo: cacheRepo,
	}
}

// ----- 字典类型操作 -----

// CreateType 创建字典类型
func (h *DictCommandHandler) CreateType(ctx context.Context, cmd *command.CreateDictTypeCommand) (*DictTypeResult, error) {
	// 检查 code 唯一性
	existing, _ := h.dictRepo.FindByCode(ctx, cmd.Code)
	if existing != nil {
		return nil, ErrDictTypeCodeDuplicate
	}

	dictType, err := dict_type.NewDictType(uuid.New().String(), cmd.Code, cmd.Name, cmd.Description)
	if err != nil {
		return nil, err
	}

	if err := h.dictRepo.Save(ctx, dictType); err != nil {
		return nil, err
	}
	return toDictTypeResult(dictType), nil
}

// UpdateType 更新字典类型
func (h *DictCommandHandler) UpdateType(ctx context.Context, cmd *command.UpdateDictTypeCommand) (*DictTypeResult, error) {
	// 加载聚合（含条目）
	dictType, err := h.dictRepo.FindByID(ctx, cmd.ID)
	if err != nil {
		return nil, err
	}

	dictType.Update(cmd.Name, cmd.Description)

	if err := h.dictRepo.Save(ctx, dictType); err != nil {
		return nil, err
	}
	return toDictTypeResult(dictType), nil
}

// DeleteType 删除字典类型（级联删除条目）
func (h *DictCommandHandler) DeleteType(ctx context.Context, cmd *command.DeleteDictTypeCommand) error {
	// 加载完整聚合（获取 code 用于清除缓存）
	dictType, err := h.dictRepo.FindByID(ctx, cmd.ID)
	if err != nil {
		return err
	}

	if err := h.dictRepo.Delete(ctx, cmd.ID); err != nil {
		return err
	}

	// 清除缓存
	_ = h.cacheRepo.DeleteByCode(ctx, dictType.Code)
	return nil
}

// EnableType 启用类型
func (h *DictCommandHandler) EnableType(ctx context.Context, cmd *command.EnableDictTypeCommand) (*DictTypeResult, error) {
	dictType, err := h.dictRepo.FindByID(ctx, cmd.ID)
	if err != nil {
		return nil, err
	}

	if err := dictType.Enable(); err != nil {
		return nil, err
	}

	if err := h.dictRepo.Save(ctx, dictType); err != nil {
		return nil, err
	}
	return toDictTypeResult(dictType), nil
}

// DisableType 禁用类型
func (h *DictCommandHandler) DisableType(ctx context.Context, cmd *command.DisableDictTypeCommand) (*DictTypeResult, error) {
	dictType, err := h.dictRepo.FindByID(ctx, cmd.ID)
	if err != nil {
		return nil, err
	}

	if err := dictType.Disable(); err != nil {
		return nil, err
	}

	if err := h.dictRepo.Save(ctx, dictType); err != nil {
		return nil, err
	}
	return toDictTypeResult(dictType), nil
}

// ----- 字典条目操作 -----

// AddEntry 添加字典条目
func (h *DictCommandHandler) AddEntry(ctx context.Context, cmd *command.AddDictEntryCommand) (*DictTypeResult, error) {
	dictType, err := h.dictRepo.FindByID(ctx, cmd.TypeID)
	if err != nil {
		return nil, err
	}

	entry := &dict_type.DictEntry{
		ID:        uuid.New().String(),
		Label:     cmd.Label,
		Value:     cmd.Value,
		SortOrder: cmd.SortOrder,
		Status:    dict_type.DictEntryStatusEnabled,
	}

	if err := dictType.AddEntry(entry); err != nil {
		return nil, err
	}

	if err := h.dictRepo.Save(ctx, dictType); err != nil {
		return nil, err
	}

	// 刷新缓存
	h.refreshCache(ctx, dictType)
	return toDictTypeResult(dictType), nil
}

// UpdateEntry 更新字典条目
func (h *DictCommandHandler) UpdateEntry(ctx context.Context, cmd *command.UpdateDictEntryCommand) (*DictTypeResult, error) {
	// 通过写仓储找到条目所属的类型 ID
	typeID, err := h.dictRepo.FindEntryTypeID(ctx, cmd.ID)
	if err != nil {
		return nil, err
	}

	dictType, err := h.dictRepo.FindByID(ctx, typeID)
	if err != nil {
		return nil, err
	}

	if err := dictType.UpdateEntry(cmd.ID, cmd.Label, cmd.Value, cmd.SortOrder); err != nil {
		return nil, err
	}

	if err := h.dictRepo.Save(ctx, dictType); err != nil {
		return nil, err
	}

	h.refreshCache(ctx, dictType)
	return toDictTypeResult(dictType), nil
}

// RemoveEntry 移除字典条目
func (h *DictCommandHandler) RemoveEntry(ctx context.Context, cmd *command.RemoveDictEntryCommand) (*DictTypeResult, error) {
	typeID, err := h.dictRepo.FindEntryTypeID(ctx, cmd.ID)
	if err != nil {
		return nil, err
	}

	dictType, err := h.dictRepo.FindByID(ctx, typeID)
	if err != nil {
		return nil, err
	}

	dictType.RemoveEntry(cmd.ID)

	if err := h.dictRepo.Save(ctx, dictType); err != nil {
		return nil, err
	}

	h.refreshCache(ctx, dictType)
	return toDictTypeResult(dictType), nil
}

// EnableEntry 启用条目
func (h *DictCommandHandler) EnableEntry(ctx context.Context, cmd *command.EnableDictEntryCommand) (*DictTypeResult, error) {
	typeID, err := h.dictRepo.FindEntryTypeID(ctx, cmd.ID)
	if err != nil {
		return nil, err
	}

	dictType, err := h.dictRepo.FindByID(ctx, typeID)
	if err != nil {
		return nil, err
	}

	if err := dictType.EnableEntry(cmd.ID); err != nil {
		return nil, err
	}

	if err := h.dictRepo.Save(ctx, dictType); err != nil {
		return nil, err
	}

	h.refreshCache(ctx, dictType)
	return toDictTypeResult(dictType), nil
}

// DisableEntry 禁用条目
func (h *DictCommandHandler) DisableEntry(ctx context.Context, cmd *command.DisableDictEntryCommand) (*DictTypeResult, error) {
	typeID, err := h.dictRepo.FindEntryTypeID(ctx, cmd.ID)
	if err != nil {
		return nil, err
	}

	dictType, err := h.dictRepo.FindByID(ctx, typeID)
	if err != nil {
		return nil, err
	}

	if err := dictType.DisableEntry(cmd.ID); err != nil {
		return nil, err
	}

	if err := h.dictRepo.Save(ctx, dictType); err != nil {
		return nil, err
	}

	h.refreshCache(ctx, dictType)
	return toDictTypeResult(dictType), nil
}

// refreshCache 刷新指定类型的缓存
func (h *DictCommandHandler) refreshCache(ctx context.Context, dictType *dict_type.DictType) {
	enabled := dictType.GetEnabledEntries()
	entries := make([]*repo.DictEntryEntry, 0, len(enabled))
	for _, e := range enabled {
		entries = append(entries, &repo.DictEntryEntry{
			Label: e.Label,
			Value: e.Value,
		})
	}
	_ = h.cacheRepo.RefreshByCode(ctx, dictType.Code, entries)
}

// toDictTypeResult 聚合转应用层 DTO
func toDictTypeResult(dictType *dict_type.DictType) *DictTypeResult {
	r := &DictTypeResult{
		ID:          dictType.ID,
		Code:        dictType.Code,
		Name:        dictType.Name,
		Description: dictType.Description,
		Status:      dictType.Status,
		CreatedAt:   dictType.CreatedAt,
		UpdatedAt:   dictType.UpdatedAt,
	}
	if len(dictType.Entries) > 0 {
		r.Entries = make([]*DictEntryResult, 0, len(dictType.Entries))
		for _, e := range dictType.Entries {
			r.Entries = append(r.Entries, &DictEntryResult{
				ID:        e.ID,
				Label:     e.Label,
				Value:     e.Value,
				SortOrder: e.SortOrder,
				Status:    e.Status,
				CreatedAt: e.CreatedAt,
				UpdatedAt: e.UpdatedAt,
			})
		}
	}
	return r
}

// 字典命令处理器错误
