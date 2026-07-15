package internal

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"ticTacSolved/task/game/data/gen"
	"ticTacSolved/task/game/service"
	"ticTacSolved/task/pkg/api"
	"ticTacSolved/task/pkg/errs"
)

type handlers struct {
	games  service.GameService
	tokens Tokens
}

func (h *handlers) login(w http.ResponseWriter, r *http.Request) {
	var req api.LoginRequest
	if err := decodeBody(r, &req); err != nil {
		writeErr(w, err)
		return
	}

	result, err := h.tokens.Login(
		r.Context(),
		req.User,
		req.Password,
		req.SessionTTLSeconds,
		req.RefreshTTLSeconds,
	)
	if err != nil {
		writeErr(w, err)
		return
	}

	writeJSON(w, http.StatusOK, api.LoginResponse{
		PlayerID: result.PlayerID,
		Session:  result.Session,
		Refresh:  result.Refresh,
	})
}

func (h *handlers) refresh(w http.ResponseWriter, r *http.Request) {
	var req api.RefreshRequest
	if err := decodeBody(r, &req); err != nil {
		writeErr(w, err)
		return
	}
	if req.RefreshToken == "" {
		writeErr(w, errs.New(errs.CodeInvalidInput, "refresh token is required"))
		return
	}

	session, err := h.tokens.Refresh(
		r.Context(),
		req.RefreshToken,
		req.SessionTTLSeconds,
	)
	if err != nil {
		writeErr(w, err)
		return
	}

	writeJSON(w, http.StatusOK, api.RefreshResponse{Session: session})
}

func (h *handlers) listGames(w http.ResponseWriter, r *http.Request) {
	games, err := h.games.WaitingGames(r.Context())
	if err != nil {
		writeErr(w, err)
		return
	}

	resp := api.GamesResponse{Games: make([]api.GameResponse, 0, len(games))}
	for _, game := range games {
		resp.Games = append(resp.Games, toGameResponse(game, "", false))
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *handlers) createGame(w http.ResponseWriter, r *http.Request) {
	var req api.CreateGameRequest
	if err := decodeBody(r, &req); err != nil {
		writeErr(w, err)
		return
	}

	game, token, err := h.games.CreateGame(
		r.Context(),
		PlayerID(r.Context()),
		req.Public,
	)
	if err != nil {
		writeErr(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, toGameResponse(game, token, true))
}

func (h *handlers) getGame(w http.ResponseWriter, r *http.Request) {
	game, err := h.games.GetGame(r.Context(), r.PathValue("id"))
	if err != nil {
		writeErr(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toGameResponse(game, "", false))
}

func (h *handlers) joinGame(w http.ResponseWriter, r *http.Request) {
	var req api.JoinGameRequest
	if err := decodeBody(r, &req); err != nil {
		writeErr(w, err)
		return
	}

	game, token, err := h.games.JoinGame(
		r.Context(),
		r.PathValue("id"),
		PlayerID(r.Context()),
		req.Code,
	)
	if err != nil {
		writeErr(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toGameResponse(game, token, false))
}

func (h *handlers) moveGame(w http.ResponseWriter, r *http.Request) {
	gameToken := r.Header.Get(api.HeaderGameToken)
	if gameToken == "" {
		writeErr(w, errs.New(errs.CodeInvalidToken, "missing game token"))
		return
	}

	claims, err := h.games.ValidateGameToken(r.Context(), gameToken)
	if err != nil {
		writeErr(w, err)
		return
	}
	if claims.GameID != r.PathValue("id") {
		writeErr(w, errs.New(
			errs.CodeInvalidToken,
			"game token does not match the game",
		))
		return
	}
	if claims.PlayerID != PlayerID(r.Context()) {
		writeErr(w, errs.New(
			errs.CodeInvalidToken,
			"game token does not belong to the player",
		))
		return
	}

	var req api.MoveRequest
	if err = decodeBody(r, &req); err != nil {
		writeErr(w, err)
		return
	}

	game, err := h.games.MakeMove(r.Context(), gameToken, req.Row, req.Col)
	if err != nil {
		writeErr(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toGameResponse(game, "", false))
}

func decodeBody[T any](r *http.Request, dst *T) error {
	err := json.NewDecoder(r.Body).Decode(dst)
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
