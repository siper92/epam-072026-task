package internal

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"ticTacSolved/task/pkg/api"
	"ticTacSolved/task/pkg/errs"
	"ticTacSolved/task/pkg/session"
)

type fakeGameClient struct {
	data    session.Data
	games   []api.GameResponse
	game    api.GameResponse
	leaders []LeaderEntry
	err     error
	calls   []string
	ids     []string
}

var _ GameClient = (*fakeGameClient)(nil)

func (f *fakeGameClient) Session() (session.Data, error) {
	return f.data, f.err
}

func (f *fakeGameClient) Login(context.Context) (session.Data, error) {
	f.calls = append(f.calls, "login")
	return f.data, f.err
}

func (f *fakeGameClient) Refresh(context.Context) (session.Data, error) {
	f.calls = append(f.calls, "refresh")
	return f.data, f.err
}

func (f *fakeGameClient) WaitingGames(context.Context) ([]api.GameResponse, error) {
	f.calls = append(f.calls, "list")
	return f.games, f.err
}

func (f *fakeGameClient) CreateGame(
	_ context.Context,
	public bool,
) (api.GameResponse, error) {
	f.calls = append(f.calls, "create")
	game := f.game
	game.IsPublic = public
	return game, f.err
}

func (f *fakeGameClient) QueueJoin(context.Context) (api.GameResponse, error) {
	f.calls = append(f.calls, "queue")
	return f.game, f.err
}

func (f *fakeGameClient) Leaderboard(
	context.Context,
	int64,
) ([]LeaderEntry, error) {
	f.calls = append(f.calls, "leaders")
	return f.leaders, f.err
}

func (f *fakeGameClient) Watch(
	_ context.Context,
	id string,
) (<-chan api.GameResponse, error) {
	f.calls = append(f.calls, "watch")
	f.ids = append(f.ids, id)
	if f.err != nil {
		return nil, f.err
	}
	updates := make(chan api.GameResponse, 1)
	updates <- f.game
	close(updates)
	return updates, nil
}

func (f *fakeGameClient) JoinGame(
	_ context.Context,
	id string,
	_ string,
) (api.GameResponse, error) {
	f.calls = append(f.calls, "join")
	f.ids = append(f.ids, id)
	return f.game, f.err
}

func (f *fakeGameClient) GetGame(
	_ context.Context,
	id string,
) (api.GameResponse, error) {
	f.calls = append(f.calls, "show")
	f.ids = append(f.ids, id)
	return f.game, f.err
}

func (f *fakeGameClient) Move(
	_ context.Context,
	id string,
	_ int,
	_ int,
) (api.GameResponse, error) {
	f.calls = append(f.calls, "move")
	f.ids = append(f.ids, id)
	return f.game, f.err
}

func newTestCmd() (*cobra.Command, *bytes.Buffer) {
	cmd := &cobra.Command{}
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	return cmd, buf
}

func TestRunInteractiveAction(t *testing.T) {
	cases := []struct {
		name        string
		command     string
		args        []string
		current     string
		fake        *fakeGameClient
		wantErr     errs.Code
		wantCalls   []string
		wantCurrent string
		wantOutput  string
	}{
		{
			name:    "list renders waiting games",
			command: "list",
			fake: &fakeGameClient{games: []api.GameResponse{
				{ID: "g1", Status: "WAITING_FOR_PLAYERS", IsPublic: true},
			}},
			wantCalls:  []string{"list"},
			wantOutput: "g1",
		},
		{
			name:        "create sets the current game",
			command:     "create",
			fake:        &fakeGameClient{game: api.GameResponse{ID: "g-new"}},
			wantCalls:   []string{"create"},
			wantCurrent: "g-new",
			wantOutput:  "game:   g-new",
		},
		{
			name:    "join requires a game id",
			command: "join",
			fake:    &fakeGameClient{},
			wantErr: errs.CodeInvalidInput,
		},
		{
			name:        "join sets the current game",
			command:     "join",
			args:        []string{"g2", "code"},
			fake:        &fakeGameClient{game: api.GameResponse{ID: "g2"}},
			wantCalls:   []string{"join"},
			wantCurrent: "g2",
		},
		{
			name:        "queue sets the current game",
			command:     "queue",
			fake:        &fakeGameClient{game: api.GameResponse{ID: "g-q"}},
			wantCalls:   []string{"queue"},
			wantCurrent: "g-q",
		},
		{
			name:    "show without a current game",
			command: "show",
			fake:    &fakeGameClient{},
			wantErr: errs.CodeInvalidInput,
		},
		{
			name:        "show uses the current game",
			command:     "show",
			current:     "g3",
			fake:        &fakeGameClient{game: api.GameResponse{ID: "g3"}},
			wantCalls:   []string{"show"},
			wantCurrent: "g3",
		},
		{
			name:    "move rejects a single numeric arg",
			command: "move",
			args:    []string{"1"},
			fake:    &fakeGameClient{},
			wantErr: errs.CodeOutOfBounds,
		},
		{
			name:        "move accepts a cell name",
			command:     "move",
			args:        []string{"b2"},
			current:     "g4",
			fake:        &fakeGameClient{game: api.GameResponse{ID: "g4"}},
			wantCalls:   []string{"move"},
			wantCurrent: "g4",
		},
		{
			name:    "move rejects a malformed cell",
			command: "move",
			args:    []string{"a", "b"},
			fake:    &fakeGameClient{},
			wantErr: errs.CodeOutOfBounds,
		},
		{
			name:        "move updates the current game",
			command:     "move",
			args:        []string{"0", "1"},
			current:     "g4",
			fake:        &fakeGameClient{game: api.GameResponse{ID: "g4"}},
			wantCalls:   []string{"move"},
			wantCurrent: "g4",
		},
		{
			name:    "watch streams until the game finishes",
			command: "watch",
			current: "g5",
			fake: &fakeGameClient{game: api.GameResponse{
				ID:     "g5",
				Status: "GAME_OVER_PlayerX_WIN",
			}},
			wantCalls:   []string{"watch"},
			wantCurrent: "g5",
			wantOutput:  "GAME_OVER_PlayerX_WIN",
		},
		{
			name:    "watch requires a game",
			command: "watch",
			fake:    &fakeGameClient{},
			wantErr: errs.CodeInvalidInput,
		},
		{
			name:    "unknown command",
			command: "dance",
			fake:    &fakeGameClient{},
			wantErr: errs.CodeInvalidAction,
		},
		{
			name:    "client error is propagated",
			command: "list",
			fake:    &fakeGameClient{err: errs.New(errs.CodeInvalidToken, "denied")},
			wantErr: errs.CodeInvalidToken,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cmd, out := newTestCmd()
			current := tc.current

			err := runInteractiveAction(
				context.Background(),
				cmd,
				tc.fake,
				&current,
				tc.command,
				tc.args,
			)
			if tc.wantErr != "" {
				if !errs.HasCode(err, tc.wantErr) {
					t.Fatalf("error = %v, want code %s", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("runInteractiveAction() failed: %v", err)
			}
			if len(tc.wantCalls) > 0 &&
				strings.Join(tc.fake.calls, ",") != strings.Join(tc.wantCalls, ",") {
				t.Fatalf("calls = %v, want %v", tc.fake.calls, tc.wantCalls)
			}
			if current != tc.wantCurrent {
				t.Fatalf("current = %q, want %q", current, tc.wantCurrent)
			}
			if tc.wantOutput != "" && !strings.Contains(out.String(), tc.wantOutput) {
				t.Fatalf("output = %q, missing %q", out.String(), tc.wantOutput)
			}
		})
	}
}

func TestHandleLine(t *testing.T) {
	cases := []struct {
		name       string
		line       string
		wantQuit   bool
		wantOutput string
	}{
		{name: "quit exits", line: "quit", wantQuit: true, wantOutput: "bye"},
		{name: "exit exits", line: "exit", wantQuit: true},
		{name: "help prints commands", line: "help", wantOutput: "move <row> <col>"},
		{name: "empty line is ignored", line: "   "},
		{name: "action error is printed", line: "dance", wantOutput: "error:"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cmd, out := newTestCmd()
			current := ""

			quit := handleLine(
				context.Background(),
				cmd,
				&fakeGameClient{},
				&current,
				tc.line,
			)
			if quit != tc.wantQuit {
				t.Fatalf("handleLine(%q) quit = %v, want %v", tc.line, quit, tc.wantQuit)
			}
			if tc.wantOutput != "" && !strings.Contains(out.String(), tc.wantOutput) {
				t.Fatalf("output = %q, missing %q", out.String(), tc.wantOutput)
			}
		})
	}
}
