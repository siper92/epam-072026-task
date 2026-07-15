package internal

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"ticTacSolved/task/game/service"
	"ticTacSolved/task/pkg/api"
)

func NewRouter(games service.GameService, tokens Tokens) http.Handler {
	gin.SetMode(gin.ReleaseMode)

	h := &handlers{games: games, tokens: tokens}

	router := gin.New()
	router.Use(Logging(), ErrorHandler())

	router.POST(api.PathLogin, h.login)
	router.POST(api.PathRefresh, h.refresh)

	authed := router.Group("", RequireSession(tokens))
	authed.GET(api.PathGames, h.listGames)
	authed.POST(api.PathGames, h.createGame)
	authed.GET(api.PathGames+"/:id", h.getGame)
	authed.POST(api.PathGames+"/:id/join", h.joinGame)
	authed.POST(api.PathGames+"/:id/move", h.moveGame)

	return router
}
