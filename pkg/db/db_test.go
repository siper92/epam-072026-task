package db

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"reflect"
	"strings"
	"testing"
)

type capturedQuery struct {
	query string
	args  []driver.Value
}

type fakeConn struct {
	columns  []string
	rows     [][]driver.Value
	captured []capturedQuery
}

func (c *fakeConn) Prepare(query string) (driver.Stmt, error) {
	return &fakeStmt{conn: c, query: query}, nil
}

func (c *fakeConn) Close() error {
	return nil
}

func (c *fakeConn) Begin() (driver.Tx, error) {
	return nil, errors.ErrUnsupported
}

type fakeStmt struct {
	conn  *fakeConn
	query string
}

func (s *fakeStmt) Close() error {
	return nil
}

func (s *fakeStmt) NumInput() int {
	return -1
}

func (s *fakeStmt) Exec(args []driver.Value) (driver.Result, error) {
	s.conn.captured = append(s.conn.captured, capturedQuery{query: s.query, args: args})
	return driver.RowsAffected(1), nil
}

func (s *fakeStmt) Query(args []driver.Value) (driver.Rows, error) {
	s.conn.captured = append(s.conn.captured, capturedQuery{query: s.query, args: args})
	return &fakeRows{columns: s.conn.columns, rows: s.conn.rows}, nil
}

type fakeRows struct {
	columns []string
	rows    [][]driver.Value
	pos     int
}

func (r *fakeRows) Columns() []string {
	return r.columns
}

func (r *fakeRows) Close() error {
	return nil
}

func (r *fakeRows) Next(dest []driver.Value) error {
	if r.pos >= len(r.rows) {
		return io.EOF
	}
	copy(dest, r.rows[r.pos])
	r.pos++
	return nil
}

type fakeDriver struct{}

func (fakeDriver) Open(string) (driver.Conn, error) {
	return nil, errors.ErrUnsupported
}

type fakeConnector struct {
	conn *fakeConn
}

func (c fakeConnector) Connect(context.Context) (driver.Conn, error) {
	return c.conn, nil
}

func (c fakeConnector) Driver() driver.Driver {
	return fakeDriver{}
}

func newFakeQueries(t *testing.T, columns []string, rows [][]driver.Value) (*Queries, *fakeConn) {
	t.Helper()
	conn := &fakeConn{columns: columns, rows: rows}
	sqldb := sql.OpenDB(fakeConnector{conn: conn})
	t.Cleanup(func() { sqldb.Close() })
	return New(sqldb), conn
}

var userColumns = []string{"id", "username", "password_hash", "role", "created_at"}

var gameColumns = []string{"id", "player_x", "player_o", "grid", "status", "winner_id", "move_count"}

func lastCaptured(t *testing.T, conn *fakeConn) capturedQuery {
	t.Helper()
	if len(conn.captured) == 0 {
		t.Fatal("expected a query to be executed")
	}
	return conn.captured[len(conn.captured)-1]
}

func TestCreateUser(t *testing.T) {
	q, conn := newFakeQueries(t, userColumns, [][]driver.Value{
		{"u1", "alice", "hash", "player", "2026-07-11T00:00:00Z"},
	})
	user, err := q.CreateUser(context.Background(), CreateUserParams{
		ID:           "u1",
		Username:     "alice",
		PasswordHash: "hash",
		Role:         "player",
	})
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}
	want := User{ID: "u1", Username: "alice", PasswordHash: "hash", Role: "player", CreatedAt: "2026-07-11T00:00:00Z"}
	if user != want {
		t.Errorf("unexpected user:\n got  %+v\n want %+v", user, want)
	}
	captured := lastCaptured(t, conn)
	if !strings.Contains(captured.query, "INSERT INTO users") {
		t.Errorf("unexpected query: %q", captured.query)
	}
	wantArgs := []driver.Value{"u1", "alice", "hash", "player"}
	if !reflect.DeepEqual(captured.args, wantArgs) {
		t.Errorf("unexpected args:\n got  %v\n want %v", captured.args, wantArgs)
	}
}

func TestGetUserByUsername(t *testing.T) {
	q, conn := newFakeQueries(t, userColumns, [][]driver.Value{
		{"u2", "bob", "hash2", "admin", "2026-07-11T00:00:00Z"},
	})
	user, err := q.GetUserByUsername(context.Background(), "bob")
	if err != nil {
		t.Fatalf("GetUserByUsername failed: %v", err)
	}
	if user.ID != "u2" || user.Role != "admin" {
		t.Errorf("unexpected user: %+v", user)
	}
	captured := lastCaptured(t, conn)
	if !strings.Contains(captured.query, "WHERE username = ?") {
		t.Errorf("unexpected query: %q", captured.query)
	}
	if !reflect.DeepEqual(captured.args, []driver.Value{"bob"}) {
		t.Errorf("unexpected args: %v", captured.args)
	}
}

func TestGetUserByIDNotFound(t *testing.T) {
	q, _ := newFakeQueries(t, userColumns, nil)
	_, err := q.GetUserByID(context.Background(), "missing")
	if !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("expected sql.ErrNoRows, got %v", err)
	}
}

func TestUpsertGame(t *testing.T) {
	q, conn := newFakeQueries(t, gameColumns, nil)
	err := q.UpsertGame(context.Background(), UpsertGameParams{
		ID:        "g1",
		PlayerX:   "u1",
		PlayerO:   "u2",
		Grid:      "_X_______",
		Status:    "TURN_O",
		WinnerID:  "",
		MoveCount: 1,
	})
	if err != nil {
		t.Fatalf("UpsertGame failed: %v", err)
	}
	captured := lastCaptured(t, conn)
	if !strings.Contains(captured.query, "INSERT INTO games") || !strings.Contains(captured.query, "ON CONFLICT (id)") {
		t.Errorf("unexpected query: %q", captured.query)
	}
	wantArgs := []driver.Value{"g1", "u1", "u2", "_X_______", "TURN_O", "", int64(1)}
	if !reflect.DeepEqual(captured.args, wantArgs) {
		t.Errorf("unexpected args:\n got  %v\n want %v", captured.args, wantArgs)
	}
}

func TestGetGame(t *testing.T) {
	q, conn := newFakeQueries(t, gameColumns, [][]driver.Value{
		{"g1", "u1", "u2", "__X__O___", "TURN_X", "", int64(2)},
	})
	game, err := q.GetGame(context.Background(), "g1")
	if err != nil {
		t.Fatalf("GetGame failed: %v", err)
	}
	want := Game{ID: "g1", PlayerX: "u1", PlayerO: "u2", Grid: "__X__O___", Status: "TURN_X", WinnerID: "", MoveCount: 2}
	if game != want {
		t.Errorf("unexpected game:\n got  %+v\n want %+v", game, want)
	}
	captured := lastCaptured(t, conn)
	if !reflect.DeepEqual(captured.args, []driver.Value{"g1"}) {
		t.Errorf("unexpected args: %v", captured.args)
	}
}

func TestGetGameNotFound(t *testing.T) {
	q, _ := newFakeQueries(t, gameColumns, nil)
	_, err := q.GetGame(context.Background(), "missing")
	if !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("expected sql.ErrNoRows, got %v", err)
	}
}

func TestListGamesByPlayer(t *testing.T) {
	q, conn := newFakeQueries(t, gameColumns, [][]driver.Value{
		{"g1", "u1", "u2", "_________", "TURN_X", "", int64(0)},
		{"g2", "u3", "u1", "X________", "TURN_O", "", int64(1)},
	})
	games, err := q.ListGamesByPlayer(context.Background(), ListGamesByPlayerParams{PlayerX: "u1", PlayerO: "u1"})
	if err != nil {
		t.Fatalf("ListGamesByPlayer failed: %v", err)
	}
	if len(games) != 2 {
		t.Fatalf("expected 2 games, got %d", len(games))
	}
	if games[0].ID != "g1" || games[1].ID != "g2" {
		t.Errorf("unexpected games: %+v", games)
	}
	captured := lastCaptured(t, conn)
	if !reflect.DeepEqual(captured.args, []driver.Value{"u1", "u1"}) {
		t.Errorf("unexpected args: %v", captured.args)
	}
}
