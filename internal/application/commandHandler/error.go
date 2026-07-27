package commandHandler

import "github.com/go-ddd-seed/go-ddd-seed/pkg/errors"

var (
	// 应用层错误码独立分段 (5000-5999)，避免与领域层错误码冲突
	// KV 配置错误 (5000-5009)
	ErrKvConfigKeyDuplicate = errors.New(5000, "kv config key already exists")

	// 字典模块错误 (5010-5019)
	ErrDictTypeCodeDuplicate = errors.New(5010, "dict type code already exists")
	ErrDictEntryIDNotFound   = errors.New(5011, "dict entry not found")

	// 认证授权错误 (5020-5029)
	ErrUsernameExists         = errors.New(5020, "用户名已存在")
	ErrRoleCodeDuplicate      = errors.New(5021, "角色编码已存在")
	ErrPermissionCodeDuplicate = errors.New(5022, "权限编码已存在")
)
