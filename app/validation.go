package app

//对请求的参数进行验证
import (
	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"gopkg.in/go-playground/validator.v9"
	zh_translations "gopkg.in/go-playground/validator.v9/translations/zh"
	"regexp"
	"strings"
)

//安装绑定验证
func validate(s interface{}) (bool, string) {
	Validate := validator.New()
	zh_ch := zh.New()
	uni := ut.New(zh_ch)
	trans, _ := uni.GetTranslator("zh")
	zh_translations.RegisterDefaultTranslations(Validate, trans)
	errs := Validate.Struct(s)
	if errs != nil {
		var msg string
		for _, err := range errs.(validator.ValidationErrors) {
			msg += err.Translate(trans) + "\n"
		}
		return false, msg
	}
	return true, ""
}

//登陆参数验证
type loginValidtor struct {
	Username string `form:"username"  validate:"lte=20,required"`
	Password string `form:"password"  validate:"gte=6,lte=16,required,printascii"`
}

func (lv *loginValidtor) isOk() (bool, string) {
	if strings.ContainsAny(lv.Username, " \n\t\r") {
		return false, "Username 不能包含空字符"
	}
	if strings.ContainsAny(lv.Password, " \n\t\r") {
		return false, "Password 不能包含空字符"
	}
	return validate(lv)
}

type registerValidtor struct {
	Username string `form:"username"  validate:"lte=20,required"`
	Password string `form:"password"  validate:"gte=6,lte=16,required,printascii"`
	Email    string `form:"email"  validate:"email,required"`
	School   string `form:"school" validate:"lte=20,required"`
	Desc     string `form:"desc" validate:"lte=255"`
}

func (rv *registerValidtor) isOk() (bool, string) {
	if strings.ContainsAny(rv.Username, " \n\t\r") {
		return false, "Username 不能包含空字符"
	}
	if strings.ContainsAny(rv.Password, " \n\t\r") {
		return false, "Password 不能包含空字符"
	}
	return validate(rv)
}

type updateValidtor struct {
	Username    string `form:"username"  validate:"lte=20"`
	OldPassword string `form:"old_password"  validate:"lte=16,printascii"`
	NewPassword string `form:"new_password"  validate:"lte=16,printascii"`
	Email       string `form:"email"`
	School      string `form:"school" validate:"lte=20"`
	Desc        string `form:"description" validate:"lte=255"`
}

func (uv *updateValidtor) isOk() (bool, string) {
	if uv.Username != "" && strings.ContainsAny(uv.Username, " \n\t\r") {
		return false, "用户名不能包含空字符"
	}
	if uv.NewPassword != "" {
		if strings.ContainsAny(uv.NewPassword, " \n\t\r") {
			return false, "密码不能包含空字符"
		}
		if len(uv.NewPassword) < 6 {
			return false, "密码长度至少是6"
		}
	}
	if uv.Email != "" {
		pattern := `\w+([-+.]\w+)*@\w+([-.]\w+)*\.\w+([-.]\w+)*` //匹配电子邮箱
		reg := regexp.MustCompile(pattern)
		if !reg.MatchString(uv.Email) {
			return false, "邮箱不合法"
		}
	}
	return validate(uv)
}
