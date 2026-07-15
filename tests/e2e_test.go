package tests

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"ticTacSolved/task/cli/server/app"
	"ticTacSolved/task/game/data"
	"ticTacSolved/task/pkg/api"
	"ticTacSolved/task/pkg/config"
)

const (
	pathQueue       = "/api/queue"
	pathLeaderboard = "/api/leaderboard"

	codeOutOfTurn    = "OUT_OF_TURN"
	codeCellOccupied = "CELL_OCCUPIED"
	codeOutOfBounds  = "OUT_OF_BOUNDS"
	codeGameFinished = "GAME_FINISHED"
	codeInvalidToken = "INVALID_TOKEN"

	statusWaiting = "WAITING_FOR_PLAYERS"
	statusXTurn   = "PlayerX_TURN"
	statusXWin    = "GAME_OVER_PlayerX_WIN"
)

func TestMain(m *testing.M) {
	os.Setenv("JWT_SECRET", "e2e-secret")
	config.LoadEnv()

	os.Exit(m.Run())
}

type client struct {
	t        *testing.T
	baseURL  string
	session  string
	playerID string
}

func newServer(t *testing.T) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(app.NewHandler(data.NewMemoryStore()))
	t.Cleanup(srv.Close)
	return srv
}

func login(
	t *testing.T,
	srv *httptest.Server,
	user string,
	password string,
) *client {
	t.Helper()
	c := &client{t: t, baseURL: srv.URL}
	status, body := c.do(
		http.MethodPost,
		api.PathLogin,
		"",
		api.LoginRequest{User: user, Password: password},
	)
	if status != http.StatusOK {
		t.Fatalf("login status = %d, body %s", status, body)
	}
	var resp api.LoginResponse
	mustDecode(t, body, &resp)
	if resp.PlayerID == "" || resp.Session.Value == "" || resp.Refresh.Value == "" {
		t.Fatalf("incomplete login response: %+v", resp)
	}
	c.session = resp.Session.Value
	c.playerID = resp.PlayerID
	return c
}

func (c *client) do(
	method string,
	path string,
	gameToken string,
	body any,
) (int, []byte) {
	c.t.Helper()
	var reader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			c.t.Fatalf("failed to encode request: %v", err)
		}
		reader = bytes.NewReader(raw)
	}
	req, err := http.NewRequest(method, c.baseURL+path, reader)
	if err != nil {
		c.t.Fatalf("failed to build request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if c.session != "" {
		req.Header.Set(api.HeaderAuthorization, api.BearerPrefix+c.session)
	}
	if gameToken != "" {
		req.Header.Set(api.HeaderGameToken, gameToken)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		c.t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		c.t.Fatalf("failed to read response: %v", err)
	}
	return resp.StatusCode, raw
}

func (c *client) createGame(public bool) api.GameResponse {
	c.t.Helper()
	status, body := c.do(
		http.MethodPost,
		api.PathGames,
		"",
		api.CreateGameRequest{Public: public},
	)
	if status != http.StatusCreated {
		c.t.Fatalf("create status = %d, body %s", status, body)
	}
	var game api.GameResponse
	mustDecode(c.t, body, &game)
	if game.ID == "" || game.GameToken == "" {
		c.t.Fatalf("incomplete create response: %+v", game)
	}
	return game
}

func (c *client) joinGame(id string, code string) (int, []byte) {
	c.t.Helper()
	return c.do(
		http.MethodPost,
		api.JoinPath(id),
		"",
		api.JoinGameRequest{Code: code},
	)
}

func (c *client) move(
	id string,
	gameToken string,
	row int,
	col int,
) (int, []byte) {
	c.t.Helper()
	return c.do(
		http.MethodPost,
		api.MovePath(id),
		gameToken,
		api.MoveRequest{Row: row, Col: col},
	)
}

func (c *client) listGames() api.GamesResponse {
	c.t.Helper()
	status, body := c.do(http.MethodGet, api.PathGames, "", nil)
	if status != http.StatusOK {
		c.t.Fatalf("list status = %d, body %s", status, body)
	}
	var resp api.GamesResponse
	mustDecode(c.t, body, &resp)
	return resp
}

func mustDecode(t *testing.T, raw []byte, dst any) {
	t.Helper()
	if err := json.Unmarshal(raw, dst); err != nil {
		t.Fatalf("failed to decode %s: %v", raw, err)
	}
}

func errorCode(t *testing.T, raw []byte) string {
	t.Helper()
	var resp api.ErrorResponse
	mustDecode(t, raw, &resp)
	if resp.Code == "" {
		t.Fatalf("missing error code in %s", raw)
	}
	return resp.Code
}

func assertMoveError(
	t *testing.T,
	c *client,
	gameID string,
	gameToken string,
	row int,
	col int,
	wantCode string,
) {
	t.Helper()
	status, body := c.move(gameID, gameToken, row, col)
	if status < 400 {
		t.Fatalf("move status = %d, want an error, body %s", status, body)
	}
	if code := errorCode(t, body); code != wantCode {
		t.Fatalf("error code = %q, want %q", code, wantCode)
	}
}

func playXWin(
	t *testing.T,
	alice *client,
	bob *client,
	gameID string,
	tokenX string,
	tokenO string,
) api.GameResponse {
	t.Helper()
	moves := []struct {
		c     *client
		token string
		row   int
		col   int
	}{
		{alice, tokenX, 0, 0},
		{bob, tokenO, 1, 0},
		{alice, tokenX, 0, 1},
		{bob, tokenO, 1, 1},
		{alice, tokenX, 0, 2},
	}
	var last api.GameResponse
	for i, m := range moves {
		status, body := m.c.move(gameID, m.token, m.row, m.col)
		if status != http.StatusOK {
			t.Fatalf("move %d status = %d, body %s", i, status, body)
		}
		mustDecode(t, body, &last)
	}
	if last.Status != statusXWin {
		t.Fatalf("status = %q, want %q", last.Status, statusXWin)
	}
	return last
}

func TestLoginCreatesUnknownUsers(t *testing.T) {
	srv := newServer(t)

	first := login(t, srv, "alice", "secret")
	second := login(t, srv, "alice", "secret")
	if first.playerID != second.playerID {
		t.Fatalf(
			"same credentials produced different players: %q vs %q",
			first.playerID, second.playerID,
		)
	}

	c := &client{t: t, baseURL: srv.URL}
	status, body := c.do(
		http.MethodPost,
		api.PathLogin,
		"",
		api.LoginRequest{User: "alice"},
	)
	if status != http.StatusBadRequest {
		t.Fatalf("login without password status = %d, body %s", status, body)
	}
}

func TestSessionIsRequired(t *testing.T) {
	srv := newServer(t)
	c := &client{t: t, baseURL: srv.URL}

	status, body := c.do(http.MethodGet, api.PathGames, "", nil)
	if status != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", status, http.StatusUnauthorized)
	}
	if code := errorCode(t, body); code != codeInvalidToken {
		t.Fatalf("error code = %q, want %q", code, codeInvalidToken)
	}

	c.session = "not-a-token"
	status, body = c.do(http.MethodGet, api.PathGames, "", nil)
	if status != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", status, http.StatusUnauthorized)
	}
	if code := errorCode(t, body); code != codeInvalidToken {
		t.Fatalf("error code = %q, want %q", code, codeInvalidToken)
	}
}

func TestLobbyFlow(t *testing.T) {
	srv := newServer(t)
	alice := login(t, srv, "alice", "secret")
	bob := login(t, srv, "bob", "secret")

	game := alice.createGame(true)
	if game.Status != statusWaiting || game.Board != "_________" {
		t.Fatalf("unexpected new game: %+v", game)
	}
	if game.Code != "" {
		t.Fatalf("public game should have no join code, got %q", game.Code)
	}

	listed := bob.listGames()
	if len(listed.Games) != 1 || listed.Games[0].ID != game.ID {
		t.Fatalf("expected game %q in lobby, got %+v", game.ID, listed.Games)
	}

	status, body := bob.joinGame(game.ID, "")
	if status != http.StatusOK {
		t.Fatalf("join status = %d, body %s", status, body)
	}
	var joined api.GameResponse
	mustDecode(t, body, &joined)
	if joined.GameToken == "" || joined.Status != statusXTurn {
		t.Fatalf("unexpected joined game: %+v", joined)
	}

	if games := bob.listGames(); len(games.Games) != 0 {
		t.Fatalf("joined game still listed: %+v", games.Games)
	}

	status, body = alice.do(http.MethodGet, api.GamePath(game.ID), "", nil)
	if status != http.StatusOK {
		t.Fatalf("get game status = %d, body %s", status, body)
	}
	var state api.GameResponse
	mustDecode(t, body, &state)
	if state.ID != game.ID || state.Status != statusXTurn {
		t.Fatalf("unexpected game state: %+v", state)
	}
}

func TestPrivateGameNeedsJoinCode(t *testing.T) {
	srv := newServer(t)
	alice := login(t, srv, "alice", "secret")
	bob := login(t, srv, "bob", "secret")

	game := alice.createGame(false)
	if game.Code == "" {
		t.Fatal("private game should return a join code to the creator")
	}

	if listed := bob.listGames(); len(listed.Games) != 0 {
		t.Fatalf("private game should not be listed, got %+v", listed.Games)
	}

	status, body := bob.joinGame(game.ID, "wrong-code")
	if status < 400 {
		t.Fatalf("join with wrong code status = %d, body %s", status, body)
	}
	if code := errorCode(t, body); code == "" {
		t.Fatalf("expected a stable error code, body %s", body)
	}

	status, body = bob.joinGame(game.ID, game.Code)
	if status != http.StatusOK {
		t.Fatalf("join with code status = %d, body %s", status, body)
	}
}

func TestMoveRulesAndErrorCodes(t *testing.T) {
	srv := newServer(t)
	alice := login(t, srv, "alice", "secret")
	bob := login(t, srv, "bob", "secret")

	game := alice.createGame(true)
	status, body := bob.joinGame(game.ID, "")
	if status != http.StatusOK {
		t.Fatalf("join status = %d, body %s", status, body)
	}
	var joined api.GameResponse
	mustDecode(t, body, &joined)
	tokenX, tokenO := game.GameToken, joined.GameToken

	assertMoveError(t, bob, game.ID, tokenO, 0, 0, codeOutOfTurn)
	assertMoveError(t, alice, game.ID, "", 0, 0, codeInvalidToken)
	assertMoveError(t, alice, game.ID, "garbage-token", 0, 0, codeInvalidToken)
	assertMoveError(t, bob, game.ID, tokenX, 0, 0, codeInvalidToken)

	status, body = alice.move(game.ID, tokenX, 0, 0)
	if status != http.StatusOK {
		t.Fatalf("move status = %d, body %s", status, body)
	}
	var afterMove api.GameResponse
	mustDecode(t, body, &afterMove)
	if afterMove.Board != "X________" {
		t.Fatalf("board = %q, want X________", afterMove.Board)
	}

	assertMoveError(t, alice, game.ID, tokenX, 1, 1, codeOutOfTurn)
	assertMoveError(t, bob, game.ID, tokenO, 0, 0, codeCellOccupied)
	assertMoveError(t, bob, game.ID, tokenO, 3, 3, codeOutOfBounds)
	assertMoveError(t, bob, game.ID, tokenO, -1, 0, codeOutOfBounds)

	moves := []struct {
		c     *client
		token string
		row   int
		col   int
	}{
		{bob, tokenO, 1, 0},
		{alice, tokenX, 0, 1},
		{bob, tokenO, 1, 1},
		{alice, tokenX, 0, 2},
	}
	var last api.GameResponse
	for i, m := range moves {
		status, body = m.c.move(game.ID, m.token, m.row, m.col)
		if status != http.StatusOK {
			t.Fatalf("move %d status = %d, body %s", i, status, body)
		}
		mustDecode(t, body, &last)
	}
	if last.Status != statusXWin {
		t.Fatalf("status = %q, want %q", last.Status, statusXWin)
	}

	assertMoveError(t, bob, game.ID, tokenO, 2, 2, codeGameFinished)
}

func TestQueuePairsTwoPlayers(t *testing.T) {
	srv := newServer(t)
	alice := login(t, srv, "alice", "secret")
	bob := login(t, srv, "bob", "secret")

	status, body := alice.do(http.MethodPost, pathQueue, "", nil)
	if status != http.StatusOK {
		t.Fatalf("queue status = %d, body %s", status, body)
	}
	var created api.GameResponse
	mustDecode(t, body, &created)
	if created.Status != statusWaiting || created.GameToken == "" {
		t.Fatalf("unexpected queue response: %+v", created)
	}

	status, body = bob.do(http.MethodPost, pathQueue, "", nil)
	if status != http.StatusOK {
		t.Fatalf("queue status = %d, body %s", status, body)
	}
	var paired api.GameResponse
	mustDecode(t, body, &paired)
	if paired.ID != created.ID {
		t.Fatalf("expected pairing with %q, got %q", created.ID, paired.ID)
	}
	if paired.Status != statusXTurn || paired.GameToken == "" {
		t.Fatalf("unexpected pairing response: %+v", paired)
	}

	status, body = alice.move(paired.ID, created.GameToken, 0, 0)
	if status != http.StatusOK {
		t.Fatalf("move with queue token status = %d, body %s", status, body)
	}
}

func TestLeaderboardRecordsResults(t *testing.T) {
	srv := newServer(t)
	alice := login(t, srv, "alice", "secret")
	bob := login(t, srv, "bob", "secret")

	game := alice.createGame(true)
	status, body := bob.joinGame(game.ID, "")
	if status != http.StatusOK {
		t.Fatalf("join status = %d, body %s", status, body)
	}
	var joined api.GameResponse
	mustDecode(t, body, &joined)

	playXWin(t, alice, bob, game.ID, game.GameToken, joined.GameToken)

	status, body = alice.do(http.MethodGet, pathLeaderboard, "", nil)
	if status != http.StatusOK {
		t.Fatalf("leaderboard status = %d, body %s", status, body)
	}
	var resp struct {
		Leaders []struct {
			PlayerID string `json:"player_id"`
			Wins     int64  `json:"wins"`
			Losses   int64  `json:"losses"`
			Draws    int64  `json:"draws"`
		} `json:"leaders"`
	}
	mustDecode(t, body, &resp)
	if len(resp.Leaders) != 2 {
		t.Fatalf("expected 2 leaderboard entries, got %+v", resp.Leaders)
	}
	if resp.Leaders[0].PlayerID != alice.playerID || resp.Leaders[0].Wins != 1 {
		t.Fatalf("unexpected leader: %+v", resp.Leaders[0])
	}
	if resp.Leaders[1].PlayerID != bob.playerID || resp.Leaders[1].Losses != 1 {
		t.Fatalf("unexpected second entry: %+v", resp.Leaders[1])
	}
}
