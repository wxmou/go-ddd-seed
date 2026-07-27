package command

// CreateKvConfigCommand 创建键值配置命令
type CreateKvConfigCommand struct {
	Key         string `json:"key"`
	Value       string `json:"value"`
	Description string `json:"description"`
}

// UpdateKvConfigCommand 更新键值配置命令
type UpdateKvConfigCommand struct {
	ID          string // from URL param
	Value       string `json:"value"`
	Description string `json:"description"`
}

// DeleteKvConfigCommand 删除键值配置命令
type DeleteKvConfigCommand struct {
	ID string // from URL param
}