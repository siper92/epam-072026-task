package internal

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"ticTacSolved/task/pkg/api"
	"ticTacSolved/task/pkg/errs"
	"ticTacSolved/task/pkg/session"
)

const (
	pathQueue       = "/api/queue"
	pathLeaderboard = "/api/leaderboard"
)

type LeaderEntry struct {
	PlayerID string `json:"player_id"`
	Wins     int64  `json:"wins"`
	Losses   int64  `json:"losses"`
	Draws    int64  `json:"draws"`
}

type leaderboardResponse struct {
	Leaders []LeaderEntry `json:"leaders"`
}

type Client struct {
	cfg     Config
	baseURL string
	http    *http.Client
	stream  *http.Client
	tokens  *TokenManager
}

func NewClient(cfg Config, store session.Store) *Client {
	c := &Client{
		cfg:     cfg,
		baseURL: strings.TrimRight(cfg.ServerURL, "/"),
		http:    &http.Client{Timeout: 30 * time.Second},
		stream:  &http.Client{},
	}
	c.tokens = NewTokenManager(cfg, store, c)
	return c
}

func (c *Client) Session() (session.Data, error) {
	return c.tokens.Data()
}

func (c *Client) Login(ctx context.Context) (session.Data, error) {
	return c.tokens.Login(ctx)
}

func (c *Client) Refresh(ctx context.Context) (session.Data, error) {
	return c.tokens.Refresh(ctx)
}

func (c *Client) WaitingGames(ctx context.Context) ([]api.GameResponse, error) {
	token, err := c.tokens.SessionToken(ctx)
	if err != nil {
		return nil, err
	}

	var out api.GamesResponse
	err = c.doJSON(ctx, http.MethodGet, api.PathGames, token, "", nil, &out)
	if err != nil {
		return nil, err
	}
	return out.Games, nil
}

func (c *Client) CreateGame(ctx context.Context, public bool) (api.GameResponse, error) {
	token, err := c.tokens.SessionToken(ctx)
	if err != nil {
		return api.GameResponse{}, err
	}

	var game api.GameResponse
	in := api.CreateGameRequest{Public: public}
	err = c.doJSON(ctx, http.MethodPost, api.PathGames, token, "", in, &game)
	if err != nil {
		return api.GameResponse{}, err
	}
	if err = c.tokens.SaveGame(game.ID, game.GameToken); err != nil {
		return api.GameResponse{}, err
	}
	return game, nil
}

func (c *Client) JoinGame(
	ctx context.Context,
	id string,
	code string,
) (api.GameResponse, error) {
	token, err := c.tokens.SessionToken(ctx)
	if err != nil {
		return api.GameResponse{}, err
	}

	var game api.GameResponse
	in := api.JoinGameRequest{Code: code}
	err = c.doJSON(ctx, http.MethodPost, api.JoinPath(id), token, "", in, &game)
	if err != nil {
		return api.GameResponse{}, err
	}
	if err = c.tokens.SaveGame(game.ID, game.GameToken); err != nil {
		return api.GameResponse{}, err
	}
	return game, nil
}

func (c *Client) QueueJoin(ctx context.Context) (api.GameResponse, error) {
	token, err := c.tokens.SessionToken(ctx)
	if err != nil {
		return api.GameResponse{}, err
	}

	var game api.GameResponse
	err = c.doJSON(ctx, http.MethodPost, pathQueue, token, "", nil, &game)
	if err != nil {
		return api.GameResponse{}, err
	}
	if err = c.tokens.SaveGame(game.ID, game.GameToken); err != nil {
		return api.GameResponse{}, err
	}
	return game, nil
}

func (c *Client) GetGame(ctx context.Context, id string) (api.GameResponse, error) {
	token, err := c.tokens.SessionToken(ctx)
	if err != nil {
		return api.GameResponse{}, err
	}

	var game api.GameResponse
	err = c.doJSON(ctx, http.MethodGet, api.GamePath(id), token, "", nil, &game)
	if err != nil {
		return api.GameResponse{}, err
	}
	return game, nil
}

func (c *Client) Move(
	ctx context.Context,
	id string,
	row int,
	col int,
) (api.GameResponse, error) {
	token, err := c.tokens.SessionToken(ctx)
	if err != nil {
		return api.GameResponse{}, err
	}
	data, err := c.tokens.Data()
	if err != nil {
		return api.GameResponse{}, err
	}
	if id == "" {
		id = data.GameID
	}
	if id == "" {
		return api.GameResponse{}, errs.New(
			errs.CodeInvalidInput,
			"game id is required, join or create a game first",
		)
	}
	gameToken := c.cfg.GameToken
	if gameToken == "" {
		gameToken = data.GameToken
	}

	var game api.GameResponse
	in := api.MoveRequest{Row: row, Col: col}
	err = c.doJSON(
		ctx,
		http.MethodPost,
		api.MovePath(id),
		token,
		gameToken,
		in,
		&game,
	)
	if err != nil {
		return api.GameResponse{}, err
	}
	return game, nil
}

func (c *Client) Watch(
	ctx context.Context,
	id string,
) (<-chan api.GameResponse, error) {
	if id == "" {
		return nil, errs.New(errs.CodeInvalidInput, "game id is required")
	}
	token, err := c.tokens.SessionToken(ctx)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		c.baseURL+api.GamePath(id)+"/watch",
		nil,
	)
	if err != nil {
		return nil, errs.Wrap(errs.CodeInvalidInput, "failed to build request", err)
	}
	req.Header.Set(api.HeaderAuthorization, api.BearerPrefix+token)
	req.Header.Set("Accept", "text/event-stream")

	resp, err := c.stream.Do(req)
	if err != nil {
		return nil, errs.Wrap(errs.CodeInvalidAction, "request to server failed", err)
	}
	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		return nil, decodeError(resp.StatusCode, raw)
	}

	updates := make(chan api.GameResponse)
	go func() {
		defer close(updates)
		defer resp.Body.Close()
		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			payload, found := strings.CutPrefix(scanner.Text(), "data: ")
			if !found {
				continue
			}
			var game api.GameResponse
			if err := json.Unmarshal([]byte(payload), &game); err != nil {
				return
			}
			select {
			case updates <- game:
			case <-ctx.Done():
				return
			}
		}
	}()
	return updates, nil
}

func (c *Client) Leaderboard(ctx context.Context, limit int64) ([]LeaderEntry, error) {
	token, err := c.tokens.SessionToken(ctx)
	if err != nil {
		return nil, err
	}

	path := pathLeaderboard
	if limit > 0 {
		path += "?limit=" + strconv.FormatInt(limit, 10)
	}
	var out leaderboardResponse
	if err = c.doJSON(ctx, http.MethodGet, path, token, "", nil, &out); err != nil {
		return nil, err
	}
	return out.Leaders, nil
}

func (c *Client) loginRequest(ctx context.Context) (api.LoginResponse, error) {
	in := api.LoginRequest{
		User:              c.cfg.User,
		Password:          c.cfg.Password,
		SessionTTLSeconds: c.cfg.SessionTTL,
		RefreshTTLSeconds: c.cfg.TokenTTL,
	}
	var out api.LoginResponse
	err := c.doJSON(ctx, http.MethodPost, api.PathLogin, "", "", in, &out)
	if err != nil {
		return api.LoginResponse{}, err
	}
	return out, nil
}

func (c *Client) refreshRequest(
	ctx context.Context,
	refreshToken string,
) (api.RefreshResponse, error) {
	in := api.RefreshRequest{
		RefreshToken:      refreshToken,
		SessionTTLSeconds: c.cfg.SessionTTL,
	}
	var out api.RefreshResponse
	err := c.doJSON(ctx, http.MethodPost, api.PathRefresh, "", "", in, &out)
	if err != nil {
		return api.RefreshResponse{}, err
	}
	return out, nil
}

func (c *Client) doJSON(
	ctx context.Context,
	method string,
	path string,
	sessionToken string,
	gameToken string,
	in any,
	out any,
) error {
	var body io.Reader
	if in != nil {
		raw, err := json.Marshal(in)
		if err != nil {
			return errs.Wrap(errs.CodeInvalidInput, "failed to encode request", err)
		}
		body = bytes.NewReader(raw)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
	if err != nil {
		return errs.Wrap(errs.CodeInvalidInput, "failed to build request", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if sessionToken != "" {
		req.Header.Set(api.HeaderAuthorization, api.BearerPrefix+sessionToken)
	}
	if gameToken != "" {
		req.Header.Set(api.HeaderGameToken, gameToken)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return errs.Wrap(errs.CodeInvalidAction, "request to server failed", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return errs.Wrap(errs.CodeInvalidAction, "failed to read response", err)
	}
	if resp.StatusCode < http.StatusOK ||
		resp.StatusCode >= http.StatusMultipleChoices {
		return decodeError(resp.StatusCode, raw)
	}
	if out == nil {
		return nil
	}
	if err = json.Unmarshal(raw, out); err != nil {
		return errs.Wrap(errs.CodeInvalidAction, "failed to decode response", err)
	}
	return nil
}

func decodeError(status int, raw []byte) error {
	var envelope api.ErrorResponse
	if err := json.Unmarshal(raw, &envelope); err == nil && envelope.Code != "" {
		return envelope.Err()
	}
	return errs.Newf(codeForStatus(status), "server returned status %d", status)
}

func codeForStatus(status int) errs.Code {
	switch status {
	case http.StatusBadRequest:
		return errs.CodeInvalidInput
	case http.StatusUnauthorized, http.StatusForbidden:
		return errs.CodeInvalidToken
	case http.StatusNotFound:
		return errs.CodeNotFound
	case http.StatusConflict:
		return errs.CodeInvalidTransition
	}
	return errs.CodeInvalidAction
}
