package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"ticTacSolved/task/pkg/api"
)

func (h *Handlers) CreateGame(c *gin.Context) {
	var req api.CreateGameRequest
	if err := decodeBody(c, &req); err != nil {
		_ = c.Error(err)
		return
	}

	game, token, err := h.games.CreateGame(
		c.Request.Context(),
		PlayerID(c),
		req.Public,
	)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, toGameResponse(game, token, true))
}
