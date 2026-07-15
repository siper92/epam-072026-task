package internal

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"ticTacSolved/task/cli/server/internal/handlers"
	"ticTacSolved/task/game/service"
	"ticTacSolved/task/pkg/api"
)

func NewRouter(games service.GameService, tokens Tokens) http.Handler {
	gin.SetMode(gin.ReleaseMode)

	h := handlers.New(games, tokens)

	router := gin.New()
	router.Use(Logging(), ErrorHandler())

	router.POST(api.PathLogin, h.Login)
	router.POST(api.PathRefresh, h.Refresh)

	authed := router.Group("", RequireSession(tokens))
	authed.GET(api.PathGames, h.ListGames)
	authed.POST(api.PathGames, h.CreateGame)
	authed.GET(api.PathGames+"/:id", h.GetGame)
	authed.POST(api.PathGames+"/:id/join", h.JoinGame)
	authed.POST(api.PathGames+"/:id/move", h.MoveGame)

	return router
}
