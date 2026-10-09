package httputil

import (
	"errors"
	"fmt"
	"net/http"
)

var (
	ErrBadRequest     = NewStatusError(nil, "bad request", http.StatusBadRequest)
	ErrUnauthorized   = NewStatusError(nil, "unauthorized", http.StatusUnauthorized)
	ErrInvalidBody    = NewStatusError(nil, "invalid request body", http.StatusBadRequest)
	ErrInternalServer = NewStatusError(nil, "internal server error", http.StatusInternalServerError)
	ErrNotFound       = NewStatusError(nil, "not found", http.StatusNotFound)
)

type StatusError struct {
	msg  string
	err  error
	code int
}

func NewStatusError(err error, msg string, code int) *StatusError {
	return &StatusError{
		err:  err,
		msg:  msg,
		code: code,
	}
}

func StatusErrorFromCode(code int) *StatusError {
	return NewStatusError(nil, http.StatusText(code), code)
}

func (e *StatusError) StatusCode() int {
	return e.code
}

func (e *StatusError) Error() string {
	if e.msg == "" {
		return http.StatusText(e.code)
	}
	return e.msg
}

func (e *StatusError) Unwrap() error {
	return e.err
}

func (e *StatusError) Wrap() error {
	if e.err == nil {
		return errors.New(e.msg)
	}
	return fmt.Errorf("%s: %w", e.msg, e.err)
}

func (e *StatusError) String() string {
	if e.err == nil {
		return fmt.Sprintf("[%d] %s", e.code, e.msg)
	}
	return fmt.Sprintf("[%d] %s: %s", e.code, e.msg, e.err.Error())
}

func (e *StatusError) Response() *Response {
	return NewResponse(e.code, nil, e)
}
