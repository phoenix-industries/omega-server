package filesservice

import (
	"context"
	"log/slog"

	"github.com/phoenix-industries/omega-server/pkg/auth"
	"github.com/phoenix-industries/omega-server/pkg/database"
	"github.com/phoenix-industries/omega-server/pkg/kernel"
)

var _ kernel.Service = (*Service)(nil)

type Service struct {
	db     *database.Database
	auth   *auth.Auth
	logger *slog.Logger
	config *Config
}

func New(config *Config) *Service {
	return &Service{config: config}
}

func (s *Service) Name() string {
	return "files/v1"
}

func (s *Service) Register(ctx context.Context, env *kernel.Env) error {
	name := s.Name()
	s.db = env.Database()
	s.auth = env.Auth()
	s.logger = env.Logger().WithGroup(name)

	r := env.Router().Group("/" + name)

	r.HandleFunc("POST /upload", s.HandleUpload)
	r.HandleFuncNative("GET /download/{id}", s.HandleDownload)

	return nil
}

func (s *Service) Start(ctx context.Context) error {
	return nil
}

func (s *Service) Stop(ctx context.Context) error {
	return nil
}
