package internal

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"ticTacSolved/task/pkg/api"
	"ticTacSolved/task/pkg/errs"
)

func TestToGameView(t *testing.T) {
	cases := []struct {
		name       string
		game       api.GameResponse
		wantStatus string
		wantNext   string
		wantWinner string
		wantCell   [3]string
	}{
		{
			name:       "waiting game",
			game:       api.GameResponse{ID: "g1", Board: "_________", Status: "WAITING_FOR_PLAYERS"},
			wantStatus: StatusWaiting,
		},
		{
			name:       "x turn",
			game:       api.GameResponse{ID: "g1", Board: "_________", Status: "PlayerX_TURN"},
			wantStatus: StatusInProgress,
			wantNext:   "x",
		},
		{
			name:       "o turn with marks",
			game:       api.GameResponse{ID: "g1", Board: "X________", Status: "PlayerO_TURN"},
			wantStatus: StatusInProgress,
			wantNext:   "o",
			wantCell:   [3]string{"x", "", ""},
		},
		{
			name:       "x win",
			game:       api.GameResponse{ID: "g1", Board: "XXX_OO___", Status: "GAME_OVER_PlayerX_WIN"},
			wantStatus: StatusFinished,
			wantWinner: "x",
			wantCell:   [3]string{"x", "x", "x"},
		},
		{
			name:       "o win",
			game:       api.GameResponse{ID: "g1", Board: "XX_OOOX__", Status: "GAME_OVER_PlayerO_WIN"},
			wantStatus: StatusFinished,
			wantWinner: "o",
			wantCell:   [3]string{"x", "x", ""},
		},
		{
			name:       "draw has no winner",
			game:       api.GameResponse{ID: "g1", Board: "XOXXOOOXX", Status: "GAME_OVER_DRAW"},
			wantStatus: StatusFinished,
			wantCell:   [3]string{"x", "o", "x"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			view := toGameView(tc.game)
			if view.ID != tc.game.ID {
				t.Fatalf("id = %q, want %q", view.ID, tc.game.ID)
			}
			if view.Status != tc.wantStatus {
				t.Fatalf("status = %q, want %q", view.Status, tc.wantStatus)
			}
			if view.Next != tc.wantNext {
				t.Fatalf("next = %q, want %q", view.Next, tc.wantNext)
			}
			winner := ""
			if view.Winner != nil {
				winner = *view.Winner
			}
			if winner != tc.wantWinner {
				t.Fatalf("winner = %q, want %q", winner, tc.wantWinner)
			}
			if view.Board[0] != tc.wantCell {
				t.Fatalf("board row = %v, want %v", view.Board[0], tc.wantCell)
			}
		})
	}
}

func TestGameViewJSONShape(t *testing.T) {
	view := toGameView(api.GameResponse{
		ID:     "g1",
		Board:  "X________",
		Status: "PlayerO_TURN",
	})
	raw, err := json.Marshal(view)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	got := string(raw)
	for _, want := range []string{
		`"id":"g1"`,
		`"status":"in_progress"`,
		`"board":[["x","",""],["","",""],["","",""]]`,
		`"next":"o"`,
		`"winner":null`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("json = %s, missing %s", got, want)
		}
	}
}

func TestPrintError(t *testing.T) {
	cases := []struct {
		name   string
		output string
		err    error
		want   string
	}{
		{
			name:   "json envelope",
			output: OutputJSON,
			err:    errs.New(errs.CodeInvalidToken, "denied"),
			want:   `"message":"denied"`,
		},
		{
			name:   "json envelope keeps the code",
			output: OutputJSON,
			err:    errs.New(errs.CodeInvalidToken, "denied"),
			want:   `"code":"` + string(errs.CodeInvalidToken) + `"`,
		},
		{
			name:   "human error",
			output: OutputHuman,
			err:    errs.New(errs.CodeInvalidToken, "denied"),
			want:   "error:",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			PrintError(buf, tc.output, tc.err)
			if !strings.Contains(buf.String(), tc.want) {
				t.Fatalf("output = %q, missing %q", buf.String(), tc.want)
			}
		})
	}
}
