package main

import (
	"context"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"github.com/phoenix-industries/omega-server/internal/authservice"
	"github.com/phoenix-industries/omega-server/internal/buildinfo"
	"github.com/phoenix-industries/omega-server/internal/filesservice/v1"
	"github.com/phoenix-industries/omega-server/pkg/auth"
	"github.com/phoenix-industries/omega-server/pkg/database"
	"github.com/phoenix-industries/omega-server/pkg/httputil"
	"github.com/phoenix-industries/omega-server/pkg/kernel"
)

const defaultPort = ":5000"

var port = flag.String("port", "", "")

func main() {
	flag.Parse()

	envfile := ".env"
	if buildinfo.DevMode() {
		envfile = ".env.dev"
	}
	if err := godotenv.Load(envfile); err != nil {
		slog.Error("failed to load environment variables", "error", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	go func() {
		<-ctx.Done() // wait for first signal
		forceCtx, forceStop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
		defer forceStop()
		<-forceCtx.Done()
		slog.Warn("second shutdown signal received - forcing immediate exit")
		os.Exit(130)
	}()

	defer func() {
		if r := recover(); r != nil {
			slog.Error("application panic", "panic", r)
			os.Exit(1)
		}
	}()

	done := make(chan struct{})
	go func() {
		defer close(done)
		if err := run(ctx); err != nil {
			slog.Error(err.Error())
			os.Exit(1)
		}
	}()

	select {
	case <-done:
		return
	case <-ctx.Done():
		slog.Info("shutdown signal received, waiting for server to exit gracefully")
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer shutdownCancel()

	select {
	case <-done:
		slog.Info("successfully shutdown server gracefully")
	case <-shutdownCtx.Done():
		slog.Info("shutdown timeout reached, forcing server to exit")
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	if *port == "" {
		*port = os.Getenv("PORT")
		if *port == "" {
			*port = defaultPort
		}
	}
	if (*port)[0] != ':' {
		*port = ":" + *port
	}

	logger := slog.Default()
	logger.Info(
		"build info",
		"build_tag", buildinfo.BuildTag,
		"go_version", buildinfo.GoVersion,
		"system_tag", buildinfo.SystemTag,
	)
	if buildinfo.DevMode() {
		logger.Info("dev mode enabled")
	}

	dbConfig, err := database.ConfigFromEnv()
	if err != nil {
		return err
	}
	if buildinfo.DevMode() {
		logger.Info("database: connecting", "connection_string", dbConfig.ConnectionString())
	}

	db, err := database.ConnectWithLogger(ctx, dbConfig, logger)
	if err != nil {
		return err
	}
	defer db.Close()

	if err := db.Migrate(ctx); err != nil {
		return err
	}

	auth, err := auth.NewFromEnv()
	if err != nil {
		return err
	}

	router := httputil.NewRouter(logger.WithGroup("http.router"))
	router.Use(httputil.RecoveryMiddleware(logger))
	router.Use(httputil.CorsMiddleware())
	router.Use(httputil.LoggingMiddleware(logger))
	router.HandleFuncNative("/", httputil.NotFoundHandler(logger))

	k := kernel.NewKernel(kernel.NewEnv(router, db, auth, logger.WithGroup("kernel")))
	services := []kernel.Service{
		authservice.New(),
		filesservice.New(filesservice.ConfigDefault()),
	}
	if err := k.Register(ctx, services...); err != nil {
		return err
	}
	if err := k.Start(ctx); err != nil {
		return err
	}
	defer k.Stop(ctx)

	srv := &http.Server{
		Addr:         *port,
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	return httputil.RunServer(ctx, srv, logger)
}
