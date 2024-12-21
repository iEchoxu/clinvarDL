package cdlerror

import (
	"fmt"
	"reflect"

	"github.com/pkg/errors"
)

type baseErrorHandler struct {
	ErrorMessage   string
	CustomErrors   map[reflect.Type]error
	StandardErrors []error
}

// addErrors 将错误添加到对应的错误类型切片中
// 自定义错误类型(即：实现了 error 接口)添加到 CustomErrors 中,返回的类型为: 非 "*errors.fundamental"
// 标准错误类型(即：errors.New 创建)添加到 StandardErrors 切片中,返回的类型为: "*errors.fundamental"
func (b *baseErrorHandler) addErrors(err error) (map[reflect.Type]error, []error) {
	if b.CustomErrors == nil {
		b.CustomErrors = make(map[reflect.Type]error, 50)
	}

	t := reflect.TypeOf(err)

	// 添加自定义类型错误
	if t.String() != "*errors.fundamental" {
		b.CustomErrors[t] = err
		return b.CustomErrors, b.StandardErrors
	}

	// 检查并添加标准错误
	for _, existingErr := range b.StandardErrors {
		if errors.Is(err, existingErr) {
			fmt.Printf("error: \"%s\" already exists, please do not add it again\n", err)
			return b.CustomErrors, b.StandardErrors
		}
	}

	b.StandardErrors = append(b.StandardErrors, err)

	return b.CustomErrors, b.StandardErrors
}

func (b *baseErrorHandler) handleErrors(err error) bool {
	if len(b.StandardErrors) == 0 && len(b.CustomErrors) == 0 {
		fmt.Printf("%T has no error types loaded, please add errors using the AddError method\n", b)
		return false
	}

	// 根据 err 类型动态创建实例，然后匹配是否是自定义错误类型
	for k, v := range b.CustomErrors {
		ptrToKType := reflect.New(k).Interface()
		if ptrToKType != nil {
			var target = ptrToKType
			if errors.As(err, &target) {
				fmt.Printf("%s%s\n", b.ErrorMessage, v.Error())
				return true
			}
		}
	}

	for _, v := range b.StandardErrors {
		if errors.Is(err, v) {
			fmt.Printf("%s: %s\n", b.ErrorMessage, v.Error())
			return true
		}
	}

	return false
}
