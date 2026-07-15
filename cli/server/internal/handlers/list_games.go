package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"ticTacSolved/task/pkg/api"
)

func (h *Handlers) ListGames(c *gin.Context) {
	games, err := h.games.WaitingGames(c.Request.Context())
	if err != nil {
		_ = c.Error(err)
		return
	}

	resp := api.GamesResponse{Games: make([]api.GameResponse, 0, len(games))}
	for _, game := range games {
		resp.Games = append(resp.Games, toGameResponse(game, "", false))
	}

	c.JSON(http.StatusOK, resp)
}
