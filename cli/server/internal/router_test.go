package internal_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"ticTacSolved/task/cli/server/internal"
	"ticTacSolved/task/game/auth"
	"ticTacSolved/task/game/data"
	"ticTacSolved/task/game/service"
	"ticTacSolved/task/game/state_machine"
	"ticTacSolved/task/pkg/api"
	"ticTacSolved/task/pkg/config"
)

func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	t.Setenv("JWT_SECRET", "test-secret")
	config.LoadEnv()

	store := data.NewMemoryStore()
	authService := auth.NewService(store)
	games := service.NewGameService(store, store, authService)
	tokens := internal.NewTokens(authService, store)

	ts := httptest.NewServer(internal.NewRouter(games, tokens))
	t.Cleanup(ts.Close)
	return ts
}

type call struct {
	method    string
	path      string
	session   string
	gameToken string
	body      any
}

func do(t *testing.T, ts *httptest.Server, c call, out any) (int, api.ErrorResponse) {
	t.Helper()

	var buf bytes.Buffer
	if c.body != nil {
		if err := json.NewEncoder(&buf).Encode(c.body); err != nil {
			t.Fatalf("failed to encode body: %v", err)
		}
	}

	req, err := http.NewRequest(c.method, ts.URL+c.path, &buf)
	if err != nil {
		t.Fatalf("failed to build request: %v", err)
	}
	if c.session != "" {
		req.Header.Set(api.HeaderAuthorization, api.BearerPrefix+c.session)
	}
	if c.gameToken != "" {
		req.Header.Set(api.HeaderGameToken, c.gameToken)
	}

	resp, err := ts.Client().Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		var errResp api.ErrorResponse
		if err = json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			t.Fatalf("failed to decode error response: %v", err)
		}
		return resp.StatusCode, errResp
	}

	if out != nil {
		if err = json.NewDecoder(resp.Body).Decode(out); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
	}
	return resp.StatusCode, api.ErrorResponse{}
}

func login(t *testing.T, ts *httptest.Server, user string) api.LoginResponse {
	t.Helper()
	var resp api.LoginResponse
	status, errResp := do(t, ts, call{
		method: http.MethodPost,
		path:   api.PathLogin,
		body:   api.LoginRequest{User: user, Password: "pw-" + user},
	}, &resp)
	if status != http.StatusOK {
		t.Fatalf("login status = %d, error = %+v", status, errResp)
	}
	return resp
}

func TestLoginValidation(t *testing.T) {
	ts := newTestServer(t)

	cases := []struct {
		name       string
		body       any
		wantStatus int
		wantCode   string
	}{
		{
			name:       "missing user",
			body:       api.LoginRequest{Password: "pw"},
			wantStatus: http.StatusBadRequest,
			wantCode:   "INVALID_INPUT",
		},
		{
			name:       "missing password",
			body:       api.LoginRequest{User: "alice"},
			wantStatus: http.StatusBadRequest,
			wantCode:   "INVALID_INPUT",
		},
		{
			name:       "invalid json",
			body:       "not-an-object",
			wantStatus: http.StatusBadRequest,
			wantCode:   "INVALID_INPUT",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			status, errResp := do(t, ts, call{
				method: http.MethodPost,
				path:   api.PathLogin,
				body:   tc.body,
			}, nil)
			if status != tc.wantStatus {
				t.Fatalf("status = %d, want %d", status, tc.wantStatus)
			}
			if errResp.Code != tc.wantCode {
				t.Fatalf("code = %q, want %q", errResp.Code, tc.wantCode)
			}
		})
	}
}

func TestLoginClampsSessionTTL(t *testing.T) {
	ts := newTestServer(t)

	var resp api.LoginResponse
	status, errResp := do(t, ts, call{
		method: http.MethodPost,
		path:   api.PathLogin,
		body: api.LoginRequest{
			User:              "clamp-user",
			Password:          "pw",
			SessionTTLSeconds: 7200,
		},
	}, &resp)
	if status != http.StatusOK {
		t.Fatalf("login status = %d, error = %+v", status, errResp)
	}

	limit := time.Now().Add(time.Hour + time.Minute).Unix()
	if resp.Session.ExpiresAt > limit {
		t.Fatalf(
			"session expires at %d, want at most %d",
			resp.Session.ExpiresAt, limit,
		)
	}
}

func TestAuthRequired(t *testing.T) {
	ts := newTestServer(t)
	creds := login(t, ts, "auth-user")

	cases := []struct {
		name    string
		session string
	}{
		{name: "no token"},
		{name: "garbage token", session: "garbage"},
		{name: "refresh token as session", session: creds.Refresh.Value},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			status, errResp := do(t, ts, call{
				method:  http.MethodGet,
				path:    api.PathGames,
				session: tc.session,
			}, nil)
			if status != http.StatusUnauthorized {
				t.Fatalf("status = %d, want %d", status, http.StatusUnauthorized)
			}
			if errResp.Code != "INVALID_TOKEN" {
				t.Fatalf("code = %q, want INVALID_TOKEN", errResp.Code)
			}
		})
	}
}

func TestRefreshFlow(t *testing.T) {
	ts := newTestServer(t)
	creds := login(t, ts, "refresh-user")

	var refreshed api.RefreshResponse
	status, errResp := do(t, ts, call{
		method: http.MethodPost,
		path:   api.PathRefresh,
		body:   api.RefreshRequest{RefreshToken: creds.Refresh.Value},
	}, &refreshed)
	if status != http.StatusOK {
		t.Fatalf("refresh status = %d, error = %+v", status, errResp)
	}
	if refreshed.Session.Value == "" {
		t.Fatal("refresh returned an empty session token")
	}

	status, errResp = do(t, ts, call{
		method:  http.MethodGet,
		path:    api.PathGames,
		session: refreshed.Session.Value,
	}, nil)
	if status != http.StatusOK {
		t.Fatalf("list with refreshed session = %d, error = %+v", status, errResp)
	}

	cases := []struct {
		name       string
		token      string
		wantStatus int
		wantCode   string
	}{
		{
			name:       "missing token",
			wantStatus: http.StatusBadRequest,
			wantCode:   "INVALID_INPUT",
		},
		{
			name:       "session token as refresh",
			token:      refreshed.Session.Value,
			wantStatus: http.StatusUnauthorized,
			wantCode:   "INVALID_TOKEN",
		},
		{
			name:       "garbage token",
			token:      "garbage",
			wantStatus: http.StatusUnauthorized,
			wantCode:   "INVALID_TOKEN",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			status, errResp := do(t, ts, call{
				method: http.MethodPost,
				path:   api.PathRefresh,
				body:   api.RefreshRequest{RefreshToken: tc.token},
			}, nil)
			if status != tc.wantStatus {
				t.Fatalf("status = %d, want %d", status, tc.wantStatus)
			}
			if errResp.Code != tc.wantCode {
				t.Fatalf("code = %q, want %q", errResp.Code, tc.wantCode)
			}
		})
	}
}

func TestPrivateGameFlow(t *testing.T) {
	ts := newTestServer(t)
	alice := login(t, ts, "alice")
	bob := login(t, ts, "bob")

	var created api.GameResponse
	status, errResp := do(t, ts, call{
		method:  http.MethodPost,
		path:    api.PathGames,
		session: alice.Session.Value,
		body:    api.CreateGameRequest{Public: false},
	}, &created)
	if status != http.StatusCreated {
		t.Fatalf("create status = %d, error = %+v", status, errResp)
	}
	if created.Code == "" || created.GameToken == "" {
		t.Fatalf("create response missing code or game token: %+v", created)
	}
	if created.Status != string(state_machine.StatusWaitingForPlayers) {
		t.Fatalf("create status = %q, want waiting", created.Status)
	}

	t.Run("join with wrong code", func(t *testing.T) {
		status, errResp := do(t, ts, call{
			method:  http.MethodPost,
			path:    api.JoinPath(created.ID),
			session: bob.Session.Value,
			body:    api.JoinGameRequest{Code: "wrong"},
		}, nil)
		if status != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", status, http.StatusBadRequest)
		}
		if errResp.Code != "INVALID_INPUT" {
			t.Fatalf("code = %q, want INVALID_INPUT", errResp.Code)
		}
	})

	var joined api.GameResponse
	status, errResp = do(t, ts, call{
		method:  http.MethodPost,
		path:    api.JoinPath(created.ID),
		session: bob.Session.Value,
		body:    api.JoinGameRequest{Code: created.Code},
	}, &joined)
	if status != http.StatusOK {
		t.Fatalf("join status = %d, error = %+v", status, errResp)
	}
	if joined.GameToken == "" {
		t.Fatal("join response missing game token")
	}
	if joined.Status != string(state_machine.StatusPlayerXTurn) {
		t.Fatalf("join status = %q, want player x turn", joined.Status)
	}

	moveCases := []struct {
		name       string
		session    string
		gameToken  string
		row, col   int
		wantStatus int
		wantCode   string
		wantState  string
	}{
		{
			name:       "move without game token",
			session:    alice.Session.Value,
			wantStatus: http.StatusUnauthorized,
			wantCode:   "INVALID_TOKEN",
		},
		{
			name:      "o moves out of turn",
			session:   bob.Session.Value,
			gameToken: joined.GameToken,
			row:       1, col: 1,
			wantStatus: http.StatusConflict,
			wantCode:   "INVALID_TRANSITION",
		},
		{
			name:      "game token of another player",
			session:   alice.Session.Value,
			gameToken: joined.GameToken,
			row:       0, col: 0,
			wantStatus: http.StatusUnauthorized,
			wantCode:   "INVALID_TOKEN",
		},
		{
			name:      "x 0 0",
			session:   alice.Session.Value,
			gameToken: created.GameToken,
			row:       0, col: 0,
			wantStatus: http.StatusOK,
			wantState:  string(state_machine.StatusPlayerOTurn),
		},
		{
			name:      "o 1 0",
			session:   bob.Session.Value,
			gameToken: joined.GameToken,
			row:       1, col: 0,
			wantStatus: http.StatusOK,
			wantState:  string(state_machine.StatusPlayerXTurn),
		},
		{
			name:      "x on occupied cell",
			session:   alice.Session.Value,
			gameToken: created.GameToken,
			row:       0, col: 0,
			wantStatus: http.StatusConflict,
			wantCode:   "CELL_OCCUPIED",
		},
		{
			name:      "x 0 1",
			session:   alice.Session.Value,
			gameToken: created.GameToken,
			row:       0, col: 1,
			wantStatus: http.StatusOK,
			wantState:  string(state_machine.StatusPlayerOTurn),
		},
		{
			name:      "o 1 1",
			session:   bob.Session.Value,
			gameToken: joined.GameToken,
			row:       1, col: 1,
			wantStatus: http.StatusOK,
			wantState:  string(state_machine.StatusPlayerXTurn),
		},
		{
			name:      "x 0 2 wins",
			session:   alice.Session.Value,
			gameToken: created.GameToken,
			row:       0, col: 2,
			wantStatus: http.StatusOK,
			wantState:  string(state_machine.StatusGameOverPlayerXWin),
		},
		{
			name:      "move after game over",
			session:   bob.Session.Value,
			gameToken: joined.GameToken,
			row:       2, col: 2,
			wantStatus: http.StatusConflict,
			wantCode:   "GAME_FINISHED",
		},
	}
	for _, tc := range moveCases {
		t.Run(tc.name, func(t *testing.T) {
			var game api.GameResponse
			status, errResp := do(t, ts, call{
				method:    http.MethodPost,
				path:      api.MovePath(created.ID),
				session:   tc.session,
				gameToken: tc.gameToken,
				body:      api.MoveRequest{Row: tc.row, Col: tc.col},
			}, &game)
			if status != tc.wantStatus {
				t.Fatalf(
					"status = %d, want %d, error = %+v",
					status, tc.wantStatus, errResp,
				)
			}
			if tc.wantCode != "" && errResp.Code != tc.wantCode {
				t.Fatalf("code = %q, want %q", errResp.Code, tc.wantCode)
			}
			if tc.wantState != "" && game.Status != tc.wantState {
				t.Fatalf("game status = %q, want %q", game.Status, tc.wantState)
			}
		})
	}

	var final api.GameResponse
	status, errResp = do(t, ts, call{
		method:  http.MethodGet,
		path:    api.GamePath(created.ID),
		session: alice.Session.Value,
	}, &final)
	if status != http.StatusOK {
		t.Fatalf("get status = %d, error = %+v", status, errResp)
	}
	if final.Board != "XXXOO____" {
		t.Fatalf("final board = %q, want XXXOO____", final.Board)
	}
	if final.Code != "" {
		t.Fatalf("get game leaked the join code %q", final.Code)
	}
}

func TestPublicGameListing(t *testing.T) {
	ts := newTestServer(t)
	carol := login(t, ts, "carol")
	dave := login(t, ts, "dave")

	var created api.GameResponse
	status, errResp := do(t, ts, call{
		method:  http.MethodPost,
		path:    api.PathGames,
		session: carol.Session.Value,
		body:    api.CreateGameRequest{Public: true},
	}, &created)
	if status != http.StatusCreated {
		t.Fatalf("create status = %d, error = %+v", status, errResp)
	}
	if created.Code != "" {
		t.Fatalf("public game returned a join code %q", created.Code)
	}

	var listed api.GamesResponse
	status, errResp = do(t, ts, call{
		method:  http.MethodGet,
		path:    api.PathGames,
		session: dave.Session.Value,
	}, &listed)
	if status != http.StatusOK {
		t.Fatalf("list status = %d, error = %+v", status, errResp)
	}
	if len(listed.Games) != 1 || listed.Games[0].ID != created.ID {
		t.Fatalf("waiting games = %+v, want the created game", listed.Games)
	}

	status, errResp = do(t, ts, call{
		method:  http.MethodPost,
		path:    api.JoinPath(created.ID),
		session: dave.Session.Value,
		body:    api.JoinGameRequest{},
	}, nil)
	if status != http.StatusOK {
		t.Fatalf("join status = %d, error = %+v", status, errResp)
	}

	status, errResp = do(t, ts, call{
		method:  http.MethodGet,
		path:    api.GamePath("missing-id"),
		session: dave.Session.Value,
	}, nil)
	if status != http.StatusNotFound {
		t.Fatalf("get missing game = %d, want %d", status, http.StatusNotFound)
	}
}
