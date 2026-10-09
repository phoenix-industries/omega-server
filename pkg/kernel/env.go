package kernel

import (
	"log/slog"

	"github.com/phoenix-industries/omega-server/pkg/auth"
	"github.com/phoenix-industries/omega-server/pkg/database"
	"github.com/phoenix-industries/omega-server/pkg/httputil"
)

type Env struct {
	db     *database.Database
	auth   *auth.Auth
	logger *slog.Logger
	router *httputil.Router
}

func NewEnv(router *httputil.Router, db *database.Database, auth *auth.Auth, logger *slog.Logger) *Env {
	if logger == nil {
		logger = slog.Default().WithGroup("kernel")
	}
	if router == nil {
		router = httputil.NewRouter(logger.WithGroup("http.router"))
	}
	return &Env{
		db:     db,
		auth:   auth,
		logger: logger,
		router: router,
	}
}

func (e *Env) Router() *httputil.Router {
	return e.router
}

func (e *Env) Database() *database.Database {
	return e.db
}

func (e *Env) Auth() *auth.Auth {
	return e.auth
}

func (e *Env) Logger() *slog.Logger {
	return e.logger
}
