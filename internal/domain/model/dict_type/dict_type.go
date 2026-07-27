package dict_type

import (
	"time"

	"github.com/go-ddd-seed/go-ddd-seed/internal/domain"
	"github.com/go-ddd-seed/go-ddd-seed/internal/domain/model"
)

// DictTypeStatus 字典类型状态
const (
	DictTypeStatusDisabled = 0
	DictTypeStatusEnabled  = 1
)

// DictType 字典类型聚合根
// 包裹字典类型信息和其下的字典条目列表（DictEntry 实体）
type DictType struct {
	model.AggregateRoot
	ID          string
	Code        string
	Name        string
	Description string
	Status      int
	Entries     []*DictEntry
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// NewDictType 创建新的字典类型
func NewDictType(id, code, name, description string) (*DictType, error) {
	if code == "" {
		return nil, ErrDictTypeCodeEmpty
	}
	if name == "" {
		return nil, ErrDictTypeNameEmpty
	}
	now := time.Now()
	return &DictType{
		ID:          id,
		Code:        code,
		Name:        name,
		Description: description,
		Status:      DictTypeStatusEnabled,
		Entries:     make([]*DictEntry, 0),
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// Update 更新字典类型基本信息
func (d *DictType) Update(name, description string) {
	d.Name = name
	d.Description = description
	d.UpdatedAt = time.Now()
}

// Enable 启用类型
func (d *DictType) Enable() error {
	if d.Status == DictTypeStatusEnabled {
		return ErrDictTypeAlreadyEnabled
	}
	d.Status = DictTypeStatusEnabled
	d.UpdatedAt = time.Now()
	return nil
}

// Disable 禁用类型
func (d *DictType) Disable() error {
	if d.Status == DictTypeStatusDisabled {
		return ErrDictTypeAlreadyDisabled
	}
	d.Status = DictTypeStatusDisabled
	d.UpdatedAt = time.Now()
	return nil
}

// IsEnabled 是否启用
func (d *DictType) IsEnabled() bool {
	return d.Status == DictTypeStatusEnabled
}

// AddEntry 添加字典条目
func (d *DictType) AddEntry(entry *DictEntry) error {
	// 幂等性：同 ID 条目已存在则跳过
	for _, e := range d.Entries {
		if e.ID == entry.ID {
			return nil
		}
	}
	// 同类型下 value 不可重复
	for _, e := range d.Entries {
		if e.Value == entry.Value {
			return ErrDictEntryValueDuplicate
		}
	}
	d.Entries = append(d.Entries, entry)
	d.UpdatedAt = time.Now()
	return nil
}

// RemoveEntry 移除字典条目
func (d *DictType) RemoveEntry(entryID string) {
	for i, e := range d.Entries {
		if e.ID == entryID {
			d.Entries = append(d.Entries[:i], d.Entries[i+1:]...)
			d.UpdatedAt = time.Now()
			return
		}
	}
}

// UpdateEntry 更新字典条目
func (d *DictType) UpdateEntry(entryID, label, value string, sortOrder int) error {
	for _, e := range d.Entries {
		if e.ID == entryID {
			// 同类型下 value 不可重复（排除自身）
			for _, other := range d.Entries {
				if other.ID != entryID && other.Value == value {
					return ErrDictEntryValueDuplicate
				}
			}
			e.Label = label
			e.Value = value
			e.SortOrder = sortOrder
			e.UpdatedAt = time.Now()
			d.UpdatedAt = time.Now()
			return nil
		}
	}
	return ErrDictEntryNotFound
}

// EnableEntry 启用条目
func (d *DictType) EnableEntry(entryID string) error {
	for _, e := range d.Entries {
		if e.ID == entryID {
			if e.IsEnabled() {
				return ErrDictEntryAlreadyEnabled
			}
			e.Enable()
			d.UpdatedAt = time.Now()
			return nil
		}
	}
	return ErrDictEntryNotFound
}

// DisableEntry 禁用条目
func (d *DictType) DisableEntry(entryID string) error {
	for _, e := range d.Entries {
		if e.ID == entryID {
			if !e.IsEnabled() {
				return ErrDictEntryAlreadyDisabled
			}
			e.Disable()
			d.UpdatedAt = time.Now()
			return nil
		}
	}
	return ErrDictEntryNotFound
}

// GetEnabledEntries 获取已启用的条目（按排序）
func (d *DictType) GetEnabledEntries() []*DictEntry {
	result := make([]*DictEntry, 0)
	for _, e := range d.Entries {
		if e.IsEnabled() {
			result = append(result, e)
		}
	}
	return result
}

// GetSortedEntries 获取所有条目（按排序）
func (d *DictType) GetSortedEntries() []*DictEntry {
	return d.Entries
}

// DictType 聚合错误码范围: 2200-2299
var (
	ErrDictTypeCodeEmpty       = domain.NewDomainError(2200, "dict type code cannot be empty")
	ErrDictTypeNameEmpty       = domain.NewDomainError(2201, "dict type name cannot be empty")
	ErrDictTypeAlreadyEnabled  = domain.NewDomainError(2202, "dict type is already enabled")
	ErrDictTypeAlreadyDisabled = domain.NewDomainError(2203, "dict type is already disabled")
	ErrDictEntryValueDuplicate = domain.NewDomainError(2210, "dict entry value already exists in this type")
	ErrDictEntryNotFound       = domain.NewDomainError(2211, "dict entry not found")
	ErrDictEntryAlreadyEnabled = domain.NewDomainError(2212, "dict entry is already enabled")
	ErrDictEntryAlreadyDisabled = domain.NewDomainError(2213, "dict entry is already disabled")
)