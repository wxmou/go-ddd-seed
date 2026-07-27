package validator

import (
	"fmt"
	"regexp"
	"strings"

	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
)

// 预编译正则
var (
	phoneRegex  = regexp.MustCompile(`^1[3-9]\d{9}$`)
	idCardRegex = regexp.MustCompile(`^[1-9]\d{5}(19|20)\d{2}(0[1-9]|1[0-2])(0[1-9]|[12]\d|3[01])\d{3}[\dXx]$`)
)

// registerCustomValidators 注册所有自定义校验器
func registerCustomValidators(v *validator.Validate) {
	_ = v.RegisterValidation("phone", validatePhone)
	_ = v.RegisterValidation("id_card", validateIDCard)
	_ = v.RegisterValidation("enum", validateEnum)
	_ = v.RegisterValidation("password", validatePassword)
}

// validatePhone 校验中国手机号
func validatePhone(fl validator.FieldLevel) bool {
	return phoneRegex.MatchString(fl.Field().String())
}

// validateIDCard 校验中国18位身份证号
func validateIDCard(fl validator.FieldLevel) bool {
	card := fl.Field().String()
	if !idCardRegex.MatchString(card) {
		return false
	}
	return validateIDCardChecksum(card)
}

// validateIDCardChecksum 校验身份证最后一位校验码
func validateIDCardChecksum(card string) bool {
	weights := []int{7, 9, 10, 5, 8, 4, 2, 1, 6, 3, 7, 9, 10, 5, 8, 4, 2}
	checkCodes := []byte{'1', '0', 'X', '9', '8', '7', '6', '5', '4', '3', '2'}

	sum := 0
	for i := 0; i < 17; i++ {
		sum += int(card[i]-'0') * weights[i]
	}

	expected := checkCodes[sum%11]
	last := card[17]
	if last >= 'a' && last <= 'z' {
		last -= 32
	}
	return last == expected
}

// validateEnum 校验枚举值
// 用法: `validate:"enum=val1,val2,val3"`
func validateEnum(fl validator.FieldLevel) bool {
	param := fl.Param()
	if param == "" {
		return true
	}
	allowed := strings.Split(param, ",")
	value := fl.Field().String()
	for _, a := range allowed {
		if strings.TrimSpace(a) == value {
			return true
		}
	}
	return false
}

// validatePassword 校验密码强度：至少6位，包含字母和数字
func validatePassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()
	if len(password) < 6 {
		return false
	}
	hasLetter := false
	hasDigit := false
	for _, ch := range password {
		if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') {
			hasLetter = true
		} else if ch >= '0' && ch <= '9' {
			hasDigit = true
		}
	}
	return hasLetter && hasDigit
}

// RegisterCustomValidators 公开方法：注册所有自定义校验器
func RegisterCustomValidators(v *validator.Validate) {
	registerCustomValidators(v)
}

// CustomValidatorMessages 返回自定义校验器的中文错误消息模板
// {0} 会被替换为字段名
func CustomValidatorMessages() map[string]string {
	return map[string]string{
		"phone":    "{0}手机号格式不正确",
		"id_card":  "{0}身份证号格式不正确",
		"enum":     "{0}取值不在允许范围内",
		"password": "{0}密码必须包含字母和数字，且长度不少于6位",
	}
}

// registerCustomTranslations 注册自定义校验器的中文翻译
// 需要在 validator.go 中初始化 trans 之后调用
func registerCustomTranslations(v *validator.Validate) error {
	msgs := CustomValidatorMessages()
	for tag, msg := range msgs {
		tagCopy := tag
		msgCopy := msg
		_ = v.RegisterTranslation(tagCopy, trans, func(ut ut.Translator) error {
			return ut.Add(tagCopy, msgCopy, true)
		}, func(ut ut.Translator, fe validator.FieldError) string {
			field := fe.Field()
			if tagCopy == "enum" && fe.Param() != "" {
				return fmt.Sprintf("%s取值不在允许范围内: [%s]", field, fe.Param())
			}
			t, _ := ut.T(tagCopy, field)
			return t
		})
	}
	return nil
}