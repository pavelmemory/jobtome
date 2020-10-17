package main

import (
	"context"
	"os"
	"os/signal"

	"go.uber.org/zap/zapcore"

	"github.com/pavelmemory/jobtome/internal"
	"github.com/pavelmemory/jobtome/internal/config"
	"github.com/pavelmemory/jobtome/internal/logging"
	shortenserv "github.com/pavelmemory/jobtome/internal/shorten"
	"github.com/pavelmemory/jobtome/internal/storage"
	"github.com/pavelmemory/jobtome/internal/storage/migrations"
	shortenrepo "github.com/pavelmemory/jobtome/internal/storage/shorten"
	"github.com/pavelmemory/jobtome/internal/webhttp"
)

func run([]string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ctx = interrupt(ctx)

	settings, err := config.NewEnvSettings(os.Getenv("ENV_PREFIX"))
	if err != nil {
		logger := logging.NewZapLogger(zapcore.InfoLevel.String())
		logger.WithError(err).Error("settings initialization")
		logger.Sync()
		return err
	}

	logger := logging.NewZapLogger(settings.LogLevel())
	defer logger.Sync()

	logger.WithString("version", internal.Version).
		WithString("commit_sha", internal.CommitSHA).
		WithString("build_timestamp", internal.BuildTimestamp).
		Info("executable build info")

	if err := migrations.Up(settings.StorageFilePath()); err != nil {
		logger.WithError(err).Error("sqlite3 database migration")
		return err
	}

	sqlLite, err := storage.NewSQLLite(settings.StorageFilePath())
	if err != nil {
		logger.WithError(err).Error("sqlite3 connection establishment")
		return err
	}
	defer sqlLite.Close()

	shortenService := shortenserv.NewService(sqlLite, shortenrepo.Repo{})

	select {
	case err := <-runAPI(ctx, logger, shortenService, settings.HTTPPort()):
		return err
	case err := <-runResolver(ctx, logger, shortenService):
		return err
	}
}

func runAPI(ctx context.Context, logger logging.Logger, shorter webhttp.ShortenService, port int) <-chan error {
	shortenHandler := webhttp.NewShortenHandler(shorter)
	router := webhttp.NewRouter(logger)
	shortenHandler.Register(router)
	infoHandler := webhttp.InfoHandler{}
	infoHandler.Register(router)

	srv := webhttp.NewServer(router)
	errChan := make(chan error)
	go func() {
		errChan <- webhttp.Serve(ctx, logger, srv, port)
	}()
	return errChan
}

func runResolver(ctx context.Context, logger logging.Logger, resolver webhttp.Resolver) <-chan error {
	resolverHandler := webhttp.NewResolverHandler(resolver)
	router := webhttp.NewRouter(logger)
	resolverHandler.Register(router)
	srv := webhttp.NewServer(router)
	errChan := make(chan error)
	go func() {
		errChan <- webhttp.Serve(ctx, logger, srv, 80)
	}()
	return errChan
}

// interrupt listens for SIGINT and cancels context.
func interrupt(ctx context.Context) context.Context {
	cctx, cancel := context.WithCancel(ctx)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	go func() {
		<-c
		cancel()
	}()

	return cctx
}
