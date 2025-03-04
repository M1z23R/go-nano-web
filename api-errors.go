package gonanoweb

type ApiError struct {
	StatusCode int    // HTTP status code
	Message    string // User-facing error message
	err        error  // Original error (not exposed to client)
}

func (e ApiError) WithError(err error) ApiError {
	e.err = err
	return e
}

func (e ApiError) Error() string {
	if e.err != nil {
		return e.err.Error()
	}
	return e.Message
}

func (e ApiError) Unwrap() error {
	return e.err
}
