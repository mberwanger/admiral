package app

type exitError struct {
	err     error
	code    int
	details string
}

//nolint:unused
func wrapErrorWithCode(err error, code int, details string) *exitError {
	return &exitError{
		err:     err,
		code:    code,
		details: details,
	}
}

//nolint:unused
func wrapError(err error, log string) *exitError {
	return wrapErrorWithCode(err, 1, log)
}

func (e *exitError) Error() string {
	return e.err.Error()
}
