package handlers

import (
	"encoding/json"
	"errors"
	"io"

	"github.com/gin-gonic/gin"

	"ticTacSolved/task/game/data/gen"
	"ticTacSolved/task/pkg/api"
	"ticTacSolved/task/pkg/errs"
)

func decodeBody[T any](c *gin.Context, dst *T) error {
	err := json.NewDecoder(c.Request.Body).Decode(dst)
	if err == nil || errors.Is(err, io.EOF) {
		return nil
	}
	return errs.Wrap(errs.CodeInvalidInput, "invalid request body", err)
}

func toGameResponse(
	game gen.Game,
	gameToken string,
	includeCode bool,
) api.GameResponse {
	resp := api.GameResponse{
		ID:        game.ID,
		Board:     game.Board,
		Status:    game.Status,
		PlayerX:   game.PlayerX,
		PlayerO:   game.PlayerO,
		IsPublic:  game.IsPublic,
		GameToken: gameToken,
	}
	if includeCode && !game.IsPublic {
		resp.Code = game.Code
	}
	return resp
}
