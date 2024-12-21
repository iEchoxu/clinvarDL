package cdlerror

import (
	"reflect"
)

type EditSettingError struct {
	baseErrorHandler
}

func NewEditSettingError(errorMessage string) *EditSettingError {
	return &EditSettingError{
		baseErrorHandler{
			ErrorMessage: errorMessage,
			CustomErrors: make(map[reflect.Type]error, 50),
		},
	}
}

func (e *EditSettingError) AddError(err error) *EditSettingError {
	customErrors, standardErrors := e.baseErrorHandler.addErrors(err)
	e.baseErrorHandler.CustomErrors = customErrors
	e.baseErrorHandler.StandardErrors = standardErrors

	return e
}
func (e *EditSettingError) HandleError(err error) bool {
	return e.baseErrorHandler.handleErrors(err)
}
