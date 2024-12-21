package errors

type EmptyResultError struct {
	msg string
}

func NewEmptyResultError(msg string) *EmptyResultError {
	return &EmptyResultError{
		msg: msg,
	}
}

func (e *EmptyResultError) Error() string {
	return e.msg
}

func (e *EmptyResultError) ShouldRetry() bool {
	return true
}
