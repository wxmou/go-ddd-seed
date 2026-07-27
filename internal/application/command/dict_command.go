package command

// CreateDictTypeCommand 创建字典类型命令
type CreateDictTypeCommand struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// UpdateDictTypeCommand 更新字典类型命令
type UpdateDictTypeCommand struct {
	ID          string // from URL param
	Name        string `json:"name"`
	Description string `json:"description"`
}

// DeleteDictTypeCommand 删除字典类型命令
type DeleteDictTypeCommand struct {
	ID string // from URL param
}

// EnableDictTypeCommand 启用字典类型命令
type EnableDictTypeCommand struct {
	ID string // from URL param
}

// DisableDictTypeCommand 禁用字典类型命令
type DisableDictTypeCommand struct {
	ID string // from URL param
}

// AddDictEntryCommand 添加字典条目命令
type AddDictEntryCommand struct {
	TypeID    string `json:"type_id"` // 所属类型 ID
	Label     string `json:"label"`
	Value     string `json:"value"`
	SortOrder int    `json:"sort_order"`
}

// UpdateDictEntryCommand 更新字典条目命令
type UpdateDictEntryCommand struct {
	ID        string // from URL param
	Label     string `json:"label"`
	Value     string `json:"value"`
	SortOrder int    `json:"sort_order"`
}

// RemoveDictEntryCommand 移除字典条目命令
type RemoveDictEntryCommand struct {
	ID string // from URL param
}

// EnableDictEntryCommand 启用字典条目命令
type EnableDictEntryCommand struct {
	ID string // from URL param
}

// DisableDictEntryCommand 禁用字典条目命令
type DisableDictEntryCommand struct {
	ID string // from URL param
}
