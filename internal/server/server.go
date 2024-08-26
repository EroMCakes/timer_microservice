package server

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"timer-microservice/internal/config"
)

type Server struct {
	router *chi.Mux
	logger *zap.SugaredLogger
	config *config.Config
}

func NewServer(cfg *config.Config, logger *zap.SugaredLogger) *Server {
	s := &Server{
		router: chi.NewRouter(),
		logger: logger,
		config: cfg,
	}

	s.setupMiddleware()
	return s
}

func (s *Server) Start() error {
	srv := &http.Server{
		Addr:    ":" + s.config.Port,
		Handler: s.router,
	}

	serverCtx, serverStopCtx := context.WithCancel(context.Background())

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-sig

		shutdownCtx, _ := context.WithTimeout(serverCtx, 30*time.Second)

		go func() {
			<-shutdownCtx.Done()
			if shutdownCtx.Err() == context.DeadlineExceeded {
				s.logger.Fatal("graceful shutdown timed out.. forcing exit.")
			}
		}()

		err := srv.Shutdown(shutdownCtx)
		if err != nil {
			s.logger.Fatal(err)
		}
		serverStopCtx()
	}()

	s.logger.Infof("Server is running on port %s", s.config.Port)
	err := srv.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		return err
	}

	<-serverCtx.Done()

	return nil
}
