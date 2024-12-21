package errors

type ParametersError struct {
	msg string
}

func NewParametersError(msg string) *ParametersError {
	return &ParametersError{
		msg: msg,
	}
}

func (p *ParametersError) Error() string {
	return p.msg
}
