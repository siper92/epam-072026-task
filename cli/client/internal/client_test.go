package internal

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"ticTacSolved/task/pkg/api"
	"ticTacSolved/task/pkg/errs"
	"ticTacSolved/task/pkg/session"
)

type stubServer struct {
	loginCalls    int
	refreshCalls  int
	lastAuth      string
	lastGameToken string
	rejectGames   bool
	plainErrors   bool
}

func (s *stubServer) handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST "+api.PathLogin, s.handleLogin)
	mux.HandleFunc("POST "+api.PathRefresh, s.handleRefresh)
	mux.HandleFunc("GET "+api.PathGames, s.handleList)
	mux.HandleFunc("POST "+api.PathGames, s.handleCreate)
	mux.HandleFunc("POST /api/games/{id}/move", s.handleMove)
	return mux
}

func (s *stubServer) handleLogin(w http.ResponseWriter, _ *http.Request) {
	s.loginCalls++
	writeStubJSON(w, http.StatusOK, api.LoginResponse{
		PlayerID: "p1",
		Session:  api.Token{Value: "session-login", ExpiresAt: futureUnix()},
		Refresh:  api.Token{Value: "refresh-login", ExpiresAt: futureUnix()},
	})
}

func (s *stubServer) handleRefresh(w http.ResponseWriter, r *http.Request) {
	s.refreshCalls++
	var req api.RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeStubJSON(w, http.StatusBadRequest, api.ErrorResponse{
			Code: string(errs.CodeInvalidInput), Message: "bad request",
		})
		return
	}
	if req.RefreshToken != "refresh-valid" {
		writeStubJSON(w, http.StatusUnauthorized, api.ErrorResponse{
			Code: string(errs.CodeInvalidToken), Message: "bad refresh token",
		})
		return
	}
	writeStubJSON(w, http.StatusOK, api.RefreshResponse{
		Session: api.Token{Value: "session-refreshed", ExpiresAt: futureUnix()},
	})
}

func (s *stubServer) handleList(w http.ResponseWriter, r *http.Request) {
	s.lastAuth = r.Header.Get(api.HeaderAuthorization)
	if s.rejectGames {
		if s.plainErrors {
			http.Error(w, "denied", http.StatusUnauthorized)
			return
		}
		writeStubJSON(w, http.StatusUnauthorized, api.ErrorResponse{
			Code: string(errs.CodeInvalidToken), Message: "invalid session",
		})
		return
	}
	writeStubJSON(w, http.StatusOK, api.GamesResponse{
		Games: []api.GameResponse{{ID: "g1", Status: "WAITING_FOR_PLAYERS", IsPublic: true}},
	})
}

func (s *stubServer) handleCreate(w http.ResponseWriter, r *http.Request) {
	s.lastAuth = r.Header.Get(api.HeaderAuthorization)
	writeStubJSON(w, http.StatusOK, api.GameResponse{
		ID:        "g-created",
		Board:     "_________",
		Status:    "WAITING_FOR_PLAYERS",
		Code:      "join-code",
		GameToken: "game-token-created",
	})
}

func (s *stubServer) handleMove(w http.ResponseWriter, r *http.Request) {
	s.lastAuth = r.Header.Get(api.HeaderAuthorization)
	s.lastGameToken = r.Header.Get(api.HeaderGameToken)
	writeStubJSON(w, http.StatusOK, api.GameResponse{
		ID:     r.PathValue("id"),
		Board:  "X________",
		Status: "PlayerO_TURN",
	})
}

func writeStubJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func futureUnix() int64 {
	return time.Now().Add(time.Hour).Unix()
}

func pastUnix() int64 {
	return time.Now().Add(-time.Hour).Unix()
}

func testConfig(url string) Config {
	return Config{
		ServerURL:  url,
		User:       "alice",
		Password:   "secret",
		Type:       TypeFile,
		TokenTTL:   3600,
		SessionTTL: 900,
	}
}

func newTestClient(
	t *testing.T,
	stub *stubServer,
	mutate func(*Config),
	seed session.Data,
) (*Client, session.Store) {
	t.Helper()
	srv := httptest.NewServer(stub.handler())
	t.Cleanup(srv.Close)

	conf := testConfig(srv.URL)
	if mutate != nil {
		mutate(&conf)
	}
	store := session.NewMemoryStore()
	if err := store.Save(seed); err != nil {
		t.Fatalf("failed to seed store: %v", err)
	}
	return NewClient(conf, store), store
}

func TestSessionToken(t *testing.T) {
	cases := []struct {
		name        string
		seed        session.Data
		mutate      func(*Config)
		wantToken   string
		wantErr     errs.Code
		wantLogin   int
		wantRefresh int
	}{
		{
			name:      "preset token flag is used directly",
			mutate:    func(c *Config) { c.Token = "preset-token" },
			wantToken: "preset-token",
		},
		{
			name: "valid stored session is reused",
			seed: session.Data{
				Session: session.Token{Value: "session-stored", ExpiresAt: futureUnix()},
			},
			wantToken: "session-stored",
		},
		{
			name: "expired session with valid refresh refreshes without login",
			seed: session.Data{
				Session: session.Token{Value: "session-old", ExpiresAt: pastUnix()},
				Refresh: session.Token{Value: "refresh-valid", ExpiresAt: futureUnix()},
			},
			wantToken:   "session-refreshed",
			wantRefresh: 1,
		},
		{
			name: "both expired logs in",
			seed: session.Data{
				Session: session.Token{Value: "session-old", ExpiresAt: pastUnix()},
				Refresh: session.Token{Value: "refresh-old", ExpiresAt: pastUnix()},
			},
			wantToken: "session-login",
			wantLogin: 1,
		},
		{
			name:    "login required without credentials",
			mutate:  func(c *Config) { c.User = "" },
			wantErr: errs.CodeInvalidInput,
		},
		{
			name: "rejected refresh token surfaces token error",
			seed: session.Data{
				Refresh: session.Token{Value: "refresh-rejected", ExpiresAt: futureUnix()},
			},
			wantErr:     errs.CodeInvalidToken,
			wantRefresh: 1,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			stub := &stubServer{}
			c, _ := newTestClient(t, stub, tc.mutate, tc.seed)

			token, err := c.tokens.SessionToken(context.Background())
			if tc.wantErr != "" {
				if !errs.HasCode(err, tc.wantErr) {
					t.Fatalf("SessionToken() error = %v, want code %s", err, tc.wantErr)
				}
			} else {
				if err != nil {
					t.Fatalf("SessionToken() failed: %v", err)
				}
				if token != tc.wantToken {
					t.Fatalf("SessionToken() = %q, want %q", token, tc.wantToken)
				}
			}
			if stub.loginCalls != tc.wantLogin {
				t.Fatalf("login calls = %d, want %d", stub.loginCalls, tc.wantLogin)
			}
			if stub.refreshCalls != tc.wantRefresh {
				t.Fatalf("refresh calls = %d, want %d", stub.refreshCalls, tc.wantRefresh)
			}
		})
	}
}

func TestLoginStoresSessionData(t *testing.T) {
	stub := &stubServer{}
	c, store := newTestClient(t, stub, nil, session.Data{})

	data, err := c.Login(context.Background())
	if err != nil {
		t.Fatalf("Login() failed: %v", err)
	}
	if data.PlayerID != "p1" {
		t.Fatalf("PlayerID = %q, want %q", data.PlayerID, "p1")
	}

	stored, err := store.Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}
	if stored.Session.Value != "session-login" {
		t.Fatalf("stored session = %q, want %q", stored.Session.Value, "session-login")
	}
	if stored.Refresh.Value != "refresh-login" {
		t.Fatalf("stored refresh = %q, want %q", stored.Refresh.Value, "refresh-login")
	}
}

func TestRefreshWithoutTokenRequiresLogin(t *testing.T) {
	stub := &stubServer{}
	c, _ := newTestClient(t, stub, nil, session.Data{})

	_, err := c.Refresh(context.Background())
	if !errs.HasCode(err, errs.CodeInvalidToken) {
		t.Fatalf("Refresh() error = %v, want code %s", err, errs.CodeInvalidToken)
	}
	if stub.refreshCalls != 0 {
		t.Fatalf("refresh calls = %d, want 0", stub.refreshCalls)
	}
}

func TestErrorEnvelopeDecoding(t *testing.T) {
	cases := []struct {
		name        string
		plainErrors bool
		wantCode    errs.Code
	}{
		{name: "json envelope", wantCode: errs.CodeInvalidToken},
		{name: "plain body falls back to status mapping", plainErrors: true, wantCode: errs.CodeInvalidToken},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			stub := &stubServer{rejectGames: true, plainErrors: tc.plainErrors}
			seed := session.Data{
				Session: session.Token{Value: "session-stored", ExpiresAt: futureUnix()},
			}
			c, _ := newTestClient(t, stub, nil, seed)

			_, err := c.WaitingGames(context.Background())
			if !errs.HasCode(err, tc.wantCode) {
				t.Fatalf("WaitingGames() error = %v, want code %s", err, tc.wantCode)
			}
		})
	}
}

func TestCodeForStatus(t *testing.T) {
	cases := []struct {
		status int
		want   errs.Code
	}{
		{status: http.StatusBadRequest, want: errs.CodeInvalidInput},
		{status: http.StatusUnauthorized, want: errs.CodeInvalidToken},
		{status: http.StatusForbidden, want: errs.CodeInvalidToken},
		{status: http.StatusNotFound, want: errs.CodeNotFound},
		{status: http.StatusConflict, want: errs.CodeInvalidTransition},
		{status: http.StatusInternalServerError, want: errs.CodeInvalidAction},
	}
	for _, tc := range cases {
		t.Run(http.StatusText(tc.status), func(t *testing.T) {
			if got := codeForStatus(tc.status); got != tc.want {
				t.Fatalf("codeForStatus(%d) = %s, want %s", tc.status, got, tc.want)
			}
		})
	}
}

func TestMoveSendsSessionAndGameTokenHeaders(t *testing.T) {
	stub := &stubServer{}
	seed := session.Data{
		Session:   session.Token{Value: "session-stored", ExpiresAt: futureUnix()},
		GameID:    "g-stored",
		GameToken: "game-token-stored",
	}
	c, _ := newTestClient(t, stub, nil, seed)

	game, err := c.Move(context.Background(), "", 0, 1)
	if err != nil {
		t.Fatalf("Move() failed: %v", err)
	}
	if game.ID != "g-stored" {
		t.Fatalf("game id = %q, want stored id %q", game.ID, "g-stored")
	}
	if stub.lastAuth != api.BearerPrefix+"session-stored" {
		t.Fatalf("auth header = %q, want bearer session token", stub.lastAuth)
	}
	if stub.lastGameToken != "game-token-stored" {
		t.Fatalf("game token header = %q, want %q", stub.lastGameToken, "game-token-stored")
	}
}

func TestMoveWithoutGameID(t *testing.T) {
	stub := &stubServer{}
	seed := session.Data{
		Session: session.Token{Value: "session-stored", ExpiresAt: futureUnix()},
	}
	c, _ := newTestClient(t, stub, nil, seed)

	_, err := c.Move(context.Background(), "", 0, 0)
	if !errs.HasCode(err, errs.CodeInvalidInput) {
		t.Fatalf("Move() error = %v, want code %s", err, errs.CodeInvalidInput)
	}
}

func TestCreateGamePersistsGameData(t *testing.T) {
	stub := &stubServer{}
	seed := session.Data{
		Session: session.Token{Value: "session-stored", ExpiresAt: futureUnix()},
	}
	c, store := newTestClient(t, stub, nil, seed)

	game, err := c.CreateGame(context.Background(), true)
	if err != nil {
		t.Fatalf("CreateGame() failed: %v", err)
	}
	if game.ID != "g-created" {
		t.Fatalf("game id = %q, want %q", game.ID, "g-created")
	}

	stored, err := store.Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}
	if stored.GameID != "g-created" {
		t.Fatalf("stored game id = %q, want %q", stored.GameID, "g-created")
	}
	if stored.GameToken != "game-token-created" {
		t.Fatalf(
			"stored game token = %q, want %q",
			stored.GameToken,
			"game-token-created",
		)
	}
}
