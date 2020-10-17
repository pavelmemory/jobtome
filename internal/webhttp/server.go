package webhttp

import (
	"context"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/pavelmemory/jobtome/internal/logging"
)

func NewServer(hl http.Handler) *http.Server {
	return &http.Server{
		Handler: hl,
	}
}

func StartServer(l net.Listener, srv *http.Server) error {
	err := srv.Serve(l)
	if err == http.ErrServerClosed {
		return nil
	}

	return err
}

func StopServer(logger logging.Logger, srv *http.Server, gracePeriod time.Duration) error {
	logger.Info("server is stopping serving")

	ctx, cancel := context.WithTimeout(context.Background(), gracePeriod)
	defer cancel()

	err := srv.Shutdown(ctx)
	switch err {
	case nil, http.ErrServerClosed:
		logger.Info("server stopped normally")
		err = nil
	case context.DeadlineExceeded:
		logger.WithString("grace_period", gracePeriod.String()).Info("server forced to stop")
		if cErr := srv.Close(); cErr != nil && cErr != http.ErrServerClosed {
			logger.WithError(cErr).Error("server forced close")
			err = cErr
		}
	default:
		logger.WithError(err).Error("shutdown server")
	}

	return err
}

func Serve(ctx context.Context, logger logging.Logger, srv *http.Server, port int) error {
	lis, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		logger.WithError(err).Error("listener instantiation")
		return err
	}

	srv.Addr = lis.Addr().String()

	logger = logger.WithString("addr", lis.Addr().String())
	logger.Info("server is starting listening")

	startErr := make(chan error)
	go func() { startErr <- StartServer(lis, srv) }()

	select {
	case err := <-startErr:
		logger.WithError(err).Error("server start")
		return err
	case <-ctx.Done():
		return StopServer(logger, srv, time.Minute)
	}
}
