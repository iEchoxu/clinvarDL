package cdlerror

import (
	"reflect"
)

type CreateConfigError struct {
	baseErrorHandler
}

func NewCreateConfigError(errorMessage string) *CreateConfigError {
	return &CreateConfigError{
		baseErrorHandler{
			ErrorMessage: errorMessage,
			CustomErrors: make(map[reflect.Type]error, 50),
		},
	}
}

func (c *CreateConfigError) AddError(err error) *CreateConfigError {
	customErrors, standardErrors := c.baseErrorHandler.addErrors(err)
	c.baseErrorHandler.CustomErrors = customErrors
	c.baseErrorHandler.StandardErrors = standardErrors

	return c
}

func (c *CreateConfigError) HandleError(err error) bool {
	return c.baseErrorHandler.handleErrors(err)
}
