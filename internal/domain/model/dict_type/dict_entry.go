package dict_type

import "time"

// DictEntryStatus 字典条目状态
const (
	DictEntryStatusDisabled = 0
	DictEntryStatusEnabled  = 1
)

// DictEntry 字典条目实体
// 属于 DictType 聚合内的实体，不独立作为聚合根
type DictEntry struct {
	ID        string
	Label     string
	Value     string
	SortOrder int
	Status    int
	CreatedAt time.Time
	UpdatedAt time.Time
}

// ChangeSort 修改排序
func (e *DictEntry) ChangeSort(sortOrder int) {
	e.SortOrder = sortOrder
	e.UpdatedAt = time.Now()
}

// Enable 启用
func (e *DictEntry) Enable() {
	e.Status = DictEntryStatusEnabled
	e.UpdatedAt = time.Now()
}

// Disable 禁用
func (e *DictEntry) Disable() {
	e.Status = DictEntryStatusDisabled
	e.UpdatedAt = time.Now()
}

// IsEnabled 是否启用
func (e *DictEntry) IsEnabled() bool {
	return e.Status == DictEntryStatusEnabled
}