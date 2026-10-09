package httputil

import (
	"encoding/json"
	"errors"
	"net/http"
)

type Response struct {
	StatusCode int    `json:"-"`
	OK         bool   `json:"ok"`
	Data       any    `json:"data,omitempty"`
	err        error  `json:"-"`
	Error      string `json:"error,omitempty"`
}

func NewResponse(statusCode int, data any, err error) *Response {
	if statusCode == 0 {
		switch {
		case err != nil:
			statusCode = http.StatusInternalServerError
		case data == nil:
			statusCode = http.StatusNoContent
		default:
			statusCode = http.StatusOK
		}
	}
	errmsg := ""
	if err != nil {
		if statusCode < http.StatusInternalServerError {
			errmsg = err.Error()
		} else {
			errmsg = http.StatusText(statusCode)
		}
	}
	return &Response{
		StatusCode: statusCode,
		OK:         err == nil,
		Data:       data,
		err:        err,
		Error:      errmsg,
	}
}

func NewResponseOK(statusCode int, data any) *Response {
	return NewResponse(statusCode, data, nil)
}

func NewResponseError(statusCode int, err error) *Response {
	return NewResponse(statusCode, nil, err)
}

func ResponseFromError(err error) *Response {
	if err == nil {
		return NewResponse(http.StatusNoContent, nil, nil)
	}
	var statusErr *StatusError
	if errors.As(err, &statusErr) {
		return NewResponse(statusErr.StatusCode(), nil, statusErr)
	}
	return NewResponse(http.StatusInternalServerError, nil, err)
}

func (r *Response) WriteJSON(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(r.StatusCode)
	if err := json.NewEncoder(w).Encode(r); err != nil {
		return err
	}
	return nil
}

func (r *Response) EmbededError() error {
	if r.err == nil {
		return nil
	}
	return r.err
}
