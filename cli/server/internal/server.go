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

func NewServer(host string, port int) *Server {
	store := data.NewMemoryStore()
	authService := auth.NewService(store)
	games := service.NewGameService(store, store, authService)
	tokens := NewTokens(authService, store)

	addr := fmt.Sprintf("%s:%d", host, port)
	return &Server{
		addr: addr,
		http: &http.Server{
			Addr:              addr,
			Handler:           NewRouter(games, tokens),
			ReadHeaderTimeout: 5 * time.Second,
		},
	}
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
