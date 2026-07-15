package internal

import (
	"net/http"

	"ticTacSolved/task/game/service"
	"ticTacSolved/task/pkg/api"
)

func NewRouter(games service.GameService, tokens Tokens) http.Handler {
	h := &handlers{games: games, tokens: tokens}
	requireSession := RequireSession(tokens)
	session := func(handler http.HandlerFunc) http.Handler {
		return Chain(handler, requireSession)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST "+api.PathLogin, h.login)
	mux.HandleFunc("POST "+api.PathRefresh, h.refresh)

	mux.Handle("GET "+api.PathGames, session(h.listGames))
	mux.Handle("POST "+api.PathGames, session(h.createGame))
	mux.Handle("GET "+api.PathGames+"/{id}", session(h.getGame))
	mux.Handle("POST "+api.PathGames+"/{id}/join", session(h.joinGame))
	mux.Handle("POST "+api.PathGames+"/{id}/move", session(h.moveGame))

	return Chain(mux, Recover, Logging)
}
