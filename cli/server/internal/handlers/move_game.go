package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"ticTacSolved/task/pkg/api"
	"ticTacSolved/task/pkg/errs"
)

func (h *Handlers) MoveGame(c *gin.Context) {
	gameToken := c.GetHeader(api.HeaderGameToken)
	if gameToken == "" {
		_ = c.Error(errs.New(errs.CodeInvalidToken, "missing game token"))
		return
	}

	claims, err := h.games.ValidateGameToken(c.Request.Context(), gameToken)
	if err != nil {
		_ = c.Error(err)
		return
	}
	if claims.GameID != c.Param("id") {
		_ = c.Error(errs.New(
			errs.CodeInvalidToken,
			"game token does not match the game",
		))
		return
	}
	if claims.PlayerID != PlayerID(c) {
		_ = c.Error(errs.New(
			errs.CodeInvalidToken,
			"game token does not belong to the player",
		))
		return
	}

	var req api.MoveRequest
	if err = decodeBody(c, &req); err != nil {
		_ = c.Error(err)
		return
	}

	game, err := h.games.MakeMove(c.Request.Context(), gameToken, req.Row, req.Col)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, toGameResponse(game, "", false))
}
