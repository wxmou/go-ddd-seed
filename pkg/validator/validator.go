package validator

import (
	"reflect"
	"strings"
	"sync"

	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	zh_translations "github.com/go-playground/validator/v10/translations/zh"
)

var (
	// validate 是全局单例校验器
	validate *validator.Validate
	// trans 是中文翻译器
	trans ut.Translator
	// once 确保校验器只初始化一次
	once sync.Once
)

// ValidateError 统一校验错误结构
type ValidateError struct {
	Code    int          `json:"code"`
	Message string       `json:"message"`
	Fields  []FieldError `json:"fields,omitempty"`
}

// FieldError 字段级错误
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// Error 实现 error 接口
func (e *ValidateError) Error() string {
	if len(e.Fields) == 0 {
		return e.Message
	}
	var sb strings.Builder
	sb.WriteString(e.Message)
	for _, f := range e.Fields {
		sb.WriteString("; ")
		sb.WriteString(f.Field)
		sb.WriteString(": ")
		sb.WriteString(f.Message)
	}
	return sb.String()
}

// getValidator 获取或初始化全局校验器实例（线程安全）
func getValidator() *validator.Validate {
	once.Do(func() {
		validate = validator.New()

		// 注册 JSON tag 作为字段名
		validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
			name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
			if name == "" || name == "-" {
				return fld.Name
			}
			return name
		})

		// 注册自定义校验器
		registerCustomValidators(validate)

		// 设置中文翻译器
		zhLocale := zh.New()
		uni := ut.New(zhLocale, zhLocale)
		trans, _ = uni.GetTranslator("zh")

		// 注册默认中文翻译
		_ = zh_translations.RegisterDefaultTranslations(validate, trans)

		// 注册自定义翻译
		_ = registerCustomTranslations(validate)
	})
	return validate
}

// ValidateRequest 校验请求 DTO，统一返回 ValidateError
func ValidateRequest(req interface{}) *ValidateError {
	v := getValidator()
	err := v.Struct(req)
	if err != nil {
		return formatError(err)
	}
	return nil
}

// ValidateVar 校验单个变量
func ValidateVar(value interface{}, tag string) *ValidateError {
	v := getValidator()
	err := v.Var(value, tag)
	if err != nil {
		return formatError(err)
	}
	return nil
}

// formatError 将 validator.ValidationErrors 转换为统一格式
func formatError(err error) *ValidateError {
	// 创建基本错误
	ve := &ValidateError{
		Code:    2, // 对应 errors.ErrBadRequest.Code
		Message: "请求参数校验失败",
	}

	// 转换为 validator 错误
	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		// 非 validator 错误，直接返回通用信息
		ve.Fields = []FieldError{
			{
				Field:   "",
				Message: err.Error(),
			},
		}
		return ve
	}

	// 提取每个字段的错误信息
	ve.Fields = make([]FieldError, 0, len(validationErrors))
	for _, fe := range validationErrors {
		fieldName := fe.Field() // 由 RegisterTagNameFunc 从 json tag 提取
		fe := fe

		var msg string
		// 优先使用已注册的中文翻译
		if trans != nil {
			t, err := trans.T(fe.Tag(), fieldName)
			if err == nil {
				msg = t
			} else {
				msg = defaultErrorMessage(&fe)
			}
		} else {
			msg = defaultErrorMessage(&fe)
		}

		ve.Fields = append(ve.Fields, FieldError{
			Field:   fieldName,
			Message: msg,
		})
	}

	return ve
}

// defaultErrorMessage 生成默认错误消息（翻译不可用时兜底）
func defaultErrorMessage(fe *validator.FieldError) string {
	field := (*fe).Field()
	tag := (*fe).Tag()
	param := (*fe).Param()

	switch tag {
	case "required":
		return field + "不能为空"
	case "min":
		return field + "长度不能小于" + param
	case "max":
		return field + "长度不能大于" + param
	case "oneof":
		return field + "取值不在允许范围内: " + param
	case "email":
		return field + "邮箱格式不正确"
	case "len":
		return field + "长度必须等于" + param
	case "gte":
		return field + "不能小于" + param
	case "gt":
		return field + "必须大于" + param
	case "lte":
		return field + "不能大于" + param
	case "lt":
		return field + "必须小于" + param
	case "numeric":
		return field + "必须是数字"
	case "alphanum":
		return field + "只能包含字母和数字"
	case "ascii":
		return field + "只能包含ASCII字符"
	case "phone":
		return field + "手机号格式不正确"
	case "id_card":
		return field + "身份证号格式不正确"
	case "enum":
		if param != "" {
			return field + "取值不在允许范围内: [" + param + "]"
		}
		return field + "取值不在允许范围内"
	case "password":
		return field + "密码必须包含字母和数字，且长度不少于6位"
	default:
		return field + "校验不通过"
	}
}