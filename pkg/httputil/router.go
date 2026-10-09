package httputil

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"
)

type HandlerFunc func(http.ResponseWriter, *http.Request) *Response

type Router struct {
	mux         *http.ServeMux
	logger      *slog.Logger
	prefix      string
	middlewares []Middleware
}

func NewRouter(logger *slog.Logger) *Router {
	if logger == nil {
		logger = slog.Default().WithGroup("http")
	}
	r := &Router{
		mux:         http.NewServeMux(),
		prefix:      "",
		logger:      logger,
		middlewares: []Middleware{},
	}
	return r
}

func (r *Router) Group(prefix string) *Router {
	sub := &Router{
		mux:         r.mux,
		logger:      r.logger,
		prefix:      r.prefix + prefix,
		middlewares: make([]Middleware, len(r.middlewares)),
	}
	copy(sub.middlewares, r.middlewares)
	return sub
}

func (r *Router) Use(middlewares ...Middleware) {
	r.middlewares = append(r.middlewares, middlewares...)
}

func (r *Router) Handle(pattern string, handler http.Handler) {
	method, path, found := pattern, "", false
	if i := strings.IndexAny(pattern, " \t"); i >= 0 {
		method, path, found = pattern[:i], strings.TrimLeft(pattern[i+1:], " \t"), true
	}
	if !found {
		path = method
		method = ""
	}
	if method != "" {
		method += " "
	}
	r.mux.Handle(method+r.prefix+path, handler)
}

func (r *Router) applyMiddlewares(handler http.Handler) http.Handler {
	return ChainMiddlewares(r.middlewares...)(handler)
}

func (r *Router) HandleFuncNative(pattern string, handler http.HandlerFunc) {
	h := r.applyMiddlewares(handler)
	r.Handle(pattern, h)
}

func (r *Router) HandleFunc(pattern string, handler HandlerFunc) {
	r.Handle(pattern, r.wrapHandler(handler))
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}

func (r *Router) wrapHandler(h HandlerFunc) http.Handler {
	var handler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		res := h(w, req)
		if res == nil {
			// no news is good news
			res = NewResponseOK(http.StatusNoContent, nil)
		}
		if !res.OK && res.StatusCode >= http.StatusInternalServerError {
			oerr := "<none>"
			if err := res.EmbededError(); err != nil {
				var statusErr *StatusError
				if errors.As(err, &statusErr) {
					oerr = statusErr.String()
				} else {
					oerr = err.Error()
				}
			}
			r.logger.ErrorContext(
				req.Context(),
				"failed to handle request",
				"path", req.URL.Path,
				"method", req.Method,
				"status_code", res.StatusCode,
				"error", res.Error,
				"original_error", oerr,
			)
		}
		if err := res.WriteJSON(w); err != nil {
			r.logger.ErrorContext(
				req.Context(),
				"failed to encode response",
				"path", req.URL.Path,
				"method", req.Method,
				"error", err,
				"response_error", res.Error,
				"original_error", res.EmbededError(),
			)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
	})
	return r.applyMiddlewares(handler)
}
