package Errors

type CustomError struct {
	msg  string
	code int
}

func (c CustomError) Error() string {
	return c.msg
}
func (c CustomError) ErrorCode() int {
	return c.code
}

func New(code int, message string) *CustomError {
	return &CustomError{
		msg:  message,
		code: code,
	}
}
