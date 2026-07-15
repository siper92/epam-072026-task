package internal

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"ticTacSolved/task/game/auth"
	"ticTacSolved/task/game/data"
	"ticTacSolved/task/game/service"
)

type Runner interface {
	Addr() string
	Run(ctx context.Context) error
}

type Server struct {
	addr string
	http *http.Server
}

var _ Runner = (*Server)(nil)

type ServerConfig struct {
	Host  string
	Port  int
	Store data.Store
}

func NewServer(cfg ServerConfig) *Server {
	store := cfg.Store
	if store == nil {
		store = data.NewMemoryStore()
	}

	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	return &Server{
		addr: addr,
		http: &http.Server{
			Addr:              addr,
			Handler:           BuildHandler(store),
			ReadHeaderTimeout: 5 * time.Second,
		},
	}
}

func BuildHandler(store data.Store) http.Handler {
	authService := auth.NewService(store)
	games := service.NewGameService(store, store, store, authService)
	queue := service.NewQueueService(games)
	tokens := NewTokens(authService, store)
	return NewRouter(games, queue, tokens)
}

func (s *Server) Addr() string {
	return s.addr
}

func (s *Server) Run(ctx context.Context) error {
	errCh := make(chan error, 1)
	go func() {
		errCh <- s.http.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
	}

	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	return s.http.Shutdown(shutdownCtx)
}
