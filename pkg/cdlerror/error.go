package cdlerror

import (
	"fmt"
)

var mapping = make(map[ErrorType]ConfigError, End)

type ConfigError interface {
	HandleError(err error) bool
}

type ErrorType int

func CheckError(errType ErrorType, err error) bool {
	if errType < Start || errType >= End {
		fmt.Printf("error type check failed: %d is not a valid error type\n", errType)
		return false
	}
	return mapping[errType].HandleError(err)
}
