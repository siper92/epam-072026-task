package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"ticTacSolved/task/pkg/api"
)

func (h *Handlers) JoinGame(c *gin.Context) {
	var req api.JoinGameRequest
	if err := decodeBody(c, &req); err != nil {
		_ = c.Error(err)
		return
	}

	game, token, err := h.games.JoinGame(
		c.Request.Context(),
		c.Param("id"),
		PlayerID(c),
		req.Code,
	)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, toGameResponse(game, token, false))
}
