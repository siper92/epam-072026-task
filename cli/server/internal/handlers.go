package internal

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"ticTacSolved/task/game/data/gen"
	"ticTacSolved/task/game/service"
	"ticTacSolved/task/pkg/api"
	"ticTacSolved/task/pkg/errs"
)

type handlers struct {
	games  service.GameService
	tokens Tokens
}

func (h *handlers) login(c *gin.Context) {
	var req api.LoginRequest
	if err := decodeBody(c, &req); err != nil {
		_ = c.Error(err)
		return
	}

	result, err := h.tokens.Login(
		c.Request.Context(),
		req.User,
		req.Password,
		req.SessionTTLSeconds,
		req.RefreshTTLSeconds,
	)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.LoginResponse{
		PlayerID: result.PlayerID,
		Session:  result.Session,
		Refresh:  result.Refresh,
	})
}

func (h *handlers) refresh(c *gin.Context) {
	var req api.RefreshRequest
	if err := decodeBody(c, &req); err != nil {
		_ = c.Error(err)
		return
	}

	if req.RefreshToken == "" {
		_ = c.Error(errs.New(errs.CodeInvalidInput, "refresh token is required"))
		return
	}

	session, err := h.tokens.Refresh(
		c.Request.Context(),
		req.RefreshToken,
		req.SessionTTLSeconds,
	)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.RefreshResponse{Session: session})
}

func (h *handlers) listGames(c *gin.Context) {
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

func (h *handlers) createGame(c *gin.Context) {
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

func (h *handlers) getGame(c *gin.Context) {
	game, err := h.games.GetGame(c.Request.Context(), c.Param("id"))
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, toGameResponse(game, "", false))
}

func (h *handlers) joinGame(c *gin.Context) {
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

func (h *handlers) moveGame(c *gin.Context) {
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
