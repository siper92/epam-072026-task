package internal

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"ticTacSolved/task/cli/server/internal/handlers"
	"ticTacSolved/task/game/service"
	"ticTacSolved/task/pkg/api"
)

const (
	PathQueue       = "/api/queue"
	PathLeaderboard = "/api/leaderboard"
)

func NewRouter(
	games service.GameService,
	queue service.QueueService,
	tokens Tokens,
) http.Handler {
	gin.SetMode(gin.ReleaseMode)

	h := handlers.New(games, tokens, queue)

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
	authed.GET(api.PathGames+"/:id/watch", h.WatchGame)
	authed.POST(PathQueue, h.QueueJoin)
	authed.GET(PathLeaderboard, h.Leaderboard)

	return router
}
