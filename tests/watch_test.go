package tests

import (
	"bufio"
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"ticTacSolved/task/pkg/api"
)

func TestWatchStreamsGameUpdates(t *testing.T) {
	srv := newServer(t)
	alice := login(t, srv, "alice", "secret")
	bob := login(t, srv, "bob", "secret")

	game := alice.createGame(true)
	status, body := bob.joinGame(game.ID, "")
	if status != http.StatusOK {
		t.Fatalf("join status = %d, body %s", status, body)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		srv.URL+api.GamePath(game.ID)+"/watch",
		nil,
	)
	if err != nil {
		t.Fatalf("failed to build request: %v", err)
	}
	req.Header.Set(api.HeaderAuthorization, api.BearerPrefix+bob.session)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("watch request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("watch status = %d", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); ct != "text/event-stream" {
		t.Fatalf("content type = %q, want text/event-stream", ct)
	}

	scanner := bufio.NewScanner(resp.Body)
	initial := nextEvent(t, scanner)
	if initial.ID != game.ID || initial.Status != statusXTurn {
		t.Fatalf("unexpected initial event: %+v", initial)
	}

	status, body = alice.move(game.ID, game.GameToken, 0, 0)
	if status != http.StatusOK {
		t.Fatalf("move status = %d, body %s", status, body)
	}

	update := nextEvent(t, scanner)
	if update.Board != "X________" {
		t.Fatalf("event board = %q, want X________", update.Board)
	}
}

func nextEvent(t *testing.T, scanner *bufio.Scanner) api.GameResponse {
	t.Helper()
	for scanner.Scan() {
		payload, found := strings.CutPrefix(scanner.Text(), "data: ")
		if !found {
			continue
		}
		var game api.GameResponse
		if err := json.Unmarshal([]byte(payload), &game); err != nil {
			t.Fatalf("failed to decode event %q: %v", payload, err)
		}
		return game
	}
	t.Fatalf("stream ended before an event: %v", scanner.Err())
	return api.GameResponse{}
}
