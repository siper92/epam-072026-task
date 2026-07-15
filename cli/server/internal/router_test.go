package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"ticTacSolved/task/cli/server/internal/handlers"
	"ticTacSolved/task/game/data/gen"
	"ticTacSolved/task/game/service"
	"ticTacSolved/task/pkg/api"
	"ticTacSolved/task/pkg/errs"
)

type stubGameService struct {
	game        gen.Game
	games       []gen.Game
	token       string
	claims      service.GameToken
	leaders     []gen.Stat
	err         error
	validateErr error
}

var _ service.GameService = (*stubGameService)(nil)

func (s *stubGameService) CreateGame(
	context.Context, string, bool,
) (gen.Game, string, error) {
	return s.game, s.token, s.err
}

func (s *stubGameService) JoinGame(
	context.Context, string, string, string,
) (gen.Game, string, error) {
	return s.game, s.token, s.err
}

func (s *stubGameService) WaitingGames(context.Context) ([]gen.Game, error) {
	return s.games, s.err
}

func (s *stubGameService) GetGame(context.Context, string) (gen.Game, error) {
	return s.game, s.err
}

func (s *stubGameService) MakeMove(
	context.Context, string, int, int,
) (gen.Game, error) {
	return s.game, s.err
}

func (s *stubGameService) ValidateGameToken(
	context.Context, string,
) (service.GameToken, error) {
	if s.validateErr != nil {
		return service.GameToken{}, s.validateErr
	}
	return s.claims, nil
}

func (s *stubGameService) ValidateJoinCode(gen.Game, string) error {
	return s.err
}

func (s *stubGameService) Leaders(
	context.Context, int64,
) ([]gen.Stat, error) {
	return s.leaders, s.err
}

func (s *stubGameService) Watch(
	context.Context, string,
) (<-chan gen.Game, func(), error) {
	if s.err != nil {
		return nil, nil, s.err
	}
	updates := make(chan gen.Game)
	close(updates)
	return updates, func() {}, nil
}

type stubQueue struct {
	game  gen.Game
	token string
	err   error
}

var _ service.QueueService = (*stubQueue)(nil)

func (s *stubQueue) Join(context.Context, string) (gen.Game, string, error) {
	return s.game, s.token, s.err
}

type stubTokens struct {
	playerID string
	result   handlers.LoginResult
	err      error
}

var _ Tokens = (*stubTokens)(nil)

func (s *stubTokens) Login(
	context.Context, string, string, int64, int64,
) (handlers.LoginResult, error) {
	return s.result, s.err
}

func (s *stubTokens) Refresh(
	context.Context, string, int64,
) (api.Token, error) {
	return s.result.Session, s.err
}

func (s *stubTokens) ValidateSession(
	context.Context, string,
) (string, error) {
	if s.err != nil {
		return "", s.err
	}
	return s.playerID, nil
}

func serveRouter(
	svc *stubGameService,
	queue *stubQueue,
	tokens Tokens,
	req *http.Request,
) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	router := NewRouter(svc, queue, tokens)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func authedRequest(method string, path string, body any) *http.Request {
	var reader *bytes.Reader
	if body == nil {
		reader = bytes.NewReader(nil)
	} else {
		raw, _ := json.Marshal(body)
		reader = bytes.NewReader(raw)
	}
	req := httptest.NewRequest(method, path, reader)
	req.Header.Set(api.HeaderAuthorization, api.BearerPrefix+"session")
	return req
}

func decodeErrorResponse(t *testing.T, rec *httptest.ResponseRecorder) api.ErrorResponse {
	t.Helper()
	var resp api.ErrorResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}
	return resp
}

func TestRouterRequiresSession(t *testing.T) {
	cases := []struct {
		name   string
		method string
		path   string
	}{
		{name: "list games", method: http.MethodGet, path: api.PathGames},
		{name: "create game", method: http.MethodPost, path: api.PathGames},
		{name: "get game", method: http.MethodGet, path: api.GamePath("g1")},
		{name: "join game", method: http.MethodPost, path: api.JoinPath("g1")},
		{name: "move", method: http.MethodPost, path: api.MovePath("g1")},
		{name: "queue", method: http.MethodPost, path: PathQueue},
		{name: "leaderboard", method: http.MethodGet, path: PathLeaderboard},
		{name: "watch", method: http.MethodGet, path: api.GamePath("g1") + "/watch"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			rec := serveRouter(&stubGameService{}, &stubQueue{}, &stubTokens{}, req)

			if rec.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
			}
			resp := decodeErrorResponse(t, rec)
			if resp.Code != string(errs.CodeInvalidToken) {
				t.Fatalf("code = %q, want %q", resp.Code, errs.CodeInvalidToken)
			}
		})
	}
}

func TestRouterLogin(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tokens := &stubTokens{result: handlers.LoginResult{
			PlayerID: "p1",
			Session:  api.Token{Value: "s", ExpiresAt: 1},
			Refresh:  api.Token{Value: "r", ExpiresAt: 2},
		}}
		req := httptest.NewRequest(
			http.MethodPost,
			api.PathLogin,
			strings.NewReader(`{"user":"u","password":"p"}`),
		)
		rec := serveRouter(&stubGameService{}, &stubQueue{}, tokens, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
		}
		var resp api.LoginResponse
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if resp.PlayerID != "p1" || resp.Session.Value != "s" || resp.Refresh.Value != "r" {
			t.Fatalf("unexpected response: %+v", resp)
		}
	})

	t.Run("invalid input", func(t *testing.T) {
		tokens := &stubTokens{err: errs.New(errs.CodeInvalidInput, "missing")}
		req := httptest.NewRequest(http.MethodPost, api.PathLogin, nil)
		rec := serveRouter(&stubGameService{}, &stubQueue{}, tokens, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}
	})
}

func TestRouterCreateGame(t *testing.T) {
	cases := []struct {
		name     string
		game     gen.Game
		wantCode string
	}{
		{
			name: "public game hides the code",
			game: gen.Game{ID: "g1", IsPublic: true, Board: "_________"},
		},
		{
			name:     "private game returns the join code",
			game:     gen.Game{ID: "g1", Code: "abcd", Board: "_________"},
			wantCode: "abcd",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := &stubGameService{game: tc.game, token: "game-token"}
			tokens := &stubTokens{playerID: "p1"}
			req := authedRequest(http.MethodPost, api.PathGames, api.CreateGameRequest{})
			rec := serveRouter(svc, &stubQueue{}, tokens, req)

			if rec.Code != http.StatusCreated {
				t.Fatalf("status = %d, want %d", rec.Code, http.StatusCreated)
			}
			var resp api.GameResponse
			if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}
			if resp.ID != "g1" || resp.GameToken != "game-token" {
				t.Fatalf("unexpected response: %+v", resp)
			}
			if resp.Code != tc.wantCode {
				t.Fatalf("code = %q, want %q", resp.Code, tc.wantCode)
			}
		})
	}
}

func TestRouterMoveGame(t *testing.T) {
	claims := service.GameToken{GameID: "g1", PlayerID: "p1", Mark: "X"}
	cases := []struct {
		name       string
		gameToken  string
		svc        *stubGameService
		playerID   string
		wantStatus int
		wantCode   errs.Code
	}{
		{
			name:       "missing game token",
			svc:        &stubGameService{claims: claims},
			playerID:   "p1",
			wantStatus: http.StatusUnauthorized,
			wantCode:   errs.CodeInvalidToken,
		},
		{
			name:      "token for another game",
			gameToken: "t",
			svc: &stubGameService{
				claims: service.GameToken{GameID: "other", PlayerID: "p1", Mark: "X"},
			},
			playerID:   "p1",
			wantStatus: http.StatusUnauthorized,
			wantCode:   errs.CodeInvalidToken,
		},
		{
			name:       "token for another player",
			gameToken:  "t",
			svc:        &stubGameService{claims: claims},
			playerID:   "p2",
			wantStatus: http.StatusUnauthorized,
			wantCode:   errs.CodeInvalidToken,
		},
		{
			name:      "out of turn is a conflict",
			gameToken: "t",
			svc: &stubGameService{
				claims: claims,
				err:    errs.New(errs.CodeOutOfTurn, "not your turn"),
			},
			playerID:   "p1",
			wantStatus: http.StatusConflict,
			wantCode:   errs.CodeOutOfTurn,
		},
		{
			name:      "occupied cell is a conflict",
			gameToken: "t",
			svc: &stubGameService{
				claims: claims,
				err:    errs.New(errs.CodeCellOccupied, "occupied"),
			},
			playerID:   "p1",
			wantStatus: http.StatusConflict,
			wantCode:   errs.CodeCellOccupied,
		},
		{
			name:      "out of bounds is a bad request",
			gameToken: "t",
			svc: &stubGameService{
				claims: claims,
				err:    errs.New(errs.CodeOutOfBounds, "outside"),
			},
			playerID:   "p1",
			wantStatus: http.StatusBadRequest,
			wantCode:   errs.CodeOutOfBounds,
		},
		{
			name:      "finished game is a conflict",
			gameToken: "t",
			svc: &stubGameService{
				claims: claims,
				err:    errs.New(errs.CodeGameFinished, "over"),
			},
			playerID:   "p1",
			wantStatus: http.StatusConflict,
			wantCode:   errs.CodeGameFinished,
		},
		{
			name:      "successful move",
			gameToken: "t",
			svc: &stubGameService{
				claims: claims,
				game:   gen.Game{ID: "g1", Board: "X________"},
			},
			playerID:   "p1",
			wantStatus: http.StatusOK,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tokens := &stubTokens{playerID: tc.playerID}
			req := authedRequest(
				http.MethodPost,
				api.MovePath("g1"),
				api.MoveRequest{Row: 0, Col: 0},
			)
			if tc.gameToken != "" {
				req.Header.Set(api.HeaderGameToken, tc.gameToken)
			}
			rec := serveRouter(tc.svc, &stubQueue{}, tokens, req)

			if rec.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d", rec.Code, tc.wantStatus)
			}
			if tc.wantCode == "" {
				return
			}
			resp := decodeErrorResponse(t, rec)
			if resp.Code != string(tc.wantCode) {
				t.Fatalf("code = %q, want %q", resp.Code, tc.wantCode)
			}
		})
	}
}

func TestRouterQueue(t *testing.T) {
	queue := &stubQueue{
		game:  gen.Game{ID: "g1", IsPublic: true, Board: "_________"},
		token: "queue-token",
	}
	req := authedRequest(http.MethodPost, PathQueue, nil)
	rec := serveRouter(&stubGameService{}, queue, &stubTokens{playerID: "p1"}, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	var resp api.GameResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.ID != "g1" || resp.GameToken != "queue-token" {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestRouterLeaderboard(t *testing.T) {
	t.Run("returns entries", func(t *testing.T) {
		svc := &stubGameService{leaders: []gen.Stat{
			{PlayerID: "p1", Wins: 2},
			{PlayerID: "p2", Losses: 2},
		}}
		req := authedRequest(http.MethodGet, PathLeaderboard, nil)
		rec := serveRouter(svc, &stubQueue{}, &stubTokens{playerID: "p1"}, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
		}
		var resp handlers.LeaderboardResponse
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if len(resp.Leaders) != 2 || resp.Leaders[0].PlayerID != "p1" {
			t.Fatalf("unexpected response: %+v", resp)
		}
	})

	t.Run("rejects a bad limit", func(t *testing.T) {
		req := authedRequest(http.MethodGet, PathLeaderboard+"?limit=abc", nil)
		rec := serveRouter(
			&stubGameService{},
			&stubQueue{},
			&stubTokens{playerID: "p1"},
			req,
		)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}
	})
}

func TestRouterWatch(t *testing.T) {
	svc := &stubGameService{game: gen.Game{
		ID:     "g1",
		Board:  "XXX_OO___",
		Status: "GAME_OVER_PlayerX_WIN",
	}}
	req := authedRequest(http.MethodGet, api.GamePath("g1")+"/watch", nil)
	rec := serveRouter(svc, &stubQueue{}, &stubTokens{playerID: "p1"}, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if got := rec.Header().Get("Content-Type"); got != "text/event-stream" {
		t.Fatalf("content type = %q, want text/event-stream", got)
	}
	body := rec.Body.String()
	if !strings.HasPrefix(body, "data: ") {
		t.Fatalf("body = %q, want an SSE data event", body)
	}
	var event api.GameResponse
	payload := strings.TrimPrefix(strings.Split(body, "\n")[0], "data: ")
	if err := json.Unmarshal([]byte(payload), &event); err != nil {
		t.Fatalf("failed to decode event: %v", err)
	}
	if event.ID != "g1" || event.Status != "GAME_OVER_PlayerX_WIN" {
		t.Fatalf("unexpected event: %+v", event)
	}
}
