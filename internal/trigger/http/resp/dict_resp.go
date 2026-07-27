package resp

// DictTypeResp 字典类型响应
type DictTypeResp struct {
	ID          string           `json:"id"`
	Code        string           `json:"code"`
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Status      int              `json:"status"`
	Entries     []*DictEntryResp `json:"entries,omitempty"`
	CreatedAt   string           `json:"created_at"`
	UpdatedAt   string           `json:"updated_at"`
}

// DictEntryResp 字典条目响应
type DictEntryResp struct {
	ID        string `json:"id"`
	TypeID    string `json:"type_id"`
	Label     string `json:"label"`
	Value     string `json:"value"`
	SortOrder int    `json:"sort_order"`
	Status    int    `json:"status"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}