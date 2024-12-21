package cdlerror

import (
	"reflect"
)

type SetEmailOrAPIError struct {
	baseErrorHandler
}

func NewSetEmailOrAPIError(errorMessage string) *SetEmailOrAPIError {
	return &SetEmailOrAPIError{
		baseErrorHandler{
			ErrorMessage: errorMessage,
			CustomErrors: make(map[reflect.Type]error, 50),
		},
	}
}

func (s *SetEmailOrAPIError) AddError(err error) *SetEmailOrAPIError {
	customErrors, standardErrors := s.baseErrorHandler.addErrors(err)
	s.baseErrorHandler.CustomErrors = customErrors
	s.baseErrorHandler.StandardErrors = standardErrors

	return s
}
func (s *SetEmailOrAPIError) HandleError(err error) bool {
	return s.baseErrorHandler.handleErrors(err)
}
