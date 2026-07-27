package req

// KvConfigCreateReq 创建键值配置请求
type KvConfigCreateReq struct {
	Key         string `json:"key" binding:"required,min=1,max=255,ascii"`
	Value       string `json:"value" binding:"required"`
	Description string `json:"description" binding:"max=500"`
}

// KvConfigUpdateReq 更新键值配置请求
type KvConfigUpdateReq struct {
	Value       string `json:"value" binding:"required"`
	Description string `json:"description" binding:"max=500"`
}

// KvConfigListReq 查询键值配置列表请求
type KvConfigListReq struct {
	Page     int `form:"page" json:"page" binding:"min=1"`
	PageSize int `form:"page_size" json:"page_size" binding:"min=1,max=100"`
	Status   int `form:"status" json:"status" binding:"oneof=-1 0 1"` // -1=全部, 0=禁用, 1=启用
}