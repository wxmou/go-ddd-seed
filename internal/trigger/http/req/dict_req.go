package req

// DictTypeCreateReq 创建字典类型请求
type DictTypeCreateReq struct {
	Code        string `json:"code" binding:"required,min=1,max=100,alphanum"`
	Name        string `json:"name" binding:"required,min=1,max=200"`
	Description string `json:"description" binding:"max=500"`
}

// DictTypeUpdateReq 更新字典类型请求
type DictTypeUpdateReq struct {
	Name        string `json:"name" binding:"required,min=1,max=200"`
	Description string `json:"description" binding:"max=500"`
}

// DictTypeListReq 查询字典类型列表请求
type DictTypeListReq struct {
	Page     int `form:"page" json:"page" binding:"min=1"`
	PageSize int `form:"page_size" json:"page_size" binding:"min=1,max=100"`
	Status   int `form:"status" json:"status" binding:"oneof=-1 0 1"`
}

// DictEntryAddReq 添加字典条目请求
type DictEntryAddReq struct {
	TypeID    string `json:"type_id" binding:"required"`
	Label     string `json:"label" binding:"required,min=1,max=200"`
	Value     string `json:"value" binding:"required,min=1,max=200"`
	SortOrder int    `json:"sort_order" binding:"min=0"`
}

// DictEntryUpdateReq 更新字典条目请求
type DictEntryUpdateReq struct {
	Label     string `json:"label" binding:"required,min=1,max=200"`
	Value     string `json:"value" binding:"required,min=1,max=200"`
	SortOrder int    `json:"sort_order" binding:"min=0"`
}
