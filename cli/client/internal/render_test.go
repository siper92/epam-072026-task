package internal

import (
	"strings"
	"testing"

	"ticTacSolved/task/pkg/api"
)

func TestRenderBoard(t *testing.T) {
	cases := []struct {
		name  string
		board string
		want  string
	}{
		{
			name:  "empty board",
			board: "_________",
			want: " _ | _ | _ \n" +
				"---+---+---\n" +
				" _ | _ | _ \n" +
				"---+---+---\n" +
				" _ | _ | _ \n",
		},
		{
			name:  "board with moves",
			board: "X_O_X____",
			want: " X | _ | O \n" +
				"---+---+---\n" +
				" _ | X | _ \n" +
				"---+---+---\n" +
				" _ | _ | _ \n",
		},
		{
			name:  "malformed board is passed through",
			board: "XO",
			want:  "XO\n",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := RenderBoard(tc.board); got != tc.want {
				t.Fatalf("RenderBoard(%q) =\n%q\nwant\n%q", tc.board, got, tc.want)
			}
		})
	}
}

func TestRenderGame(t *testing.T) {
	cases := []struct {
		name        string
		game        api.GameResponse
		wantLines   []string
		absentLines []string
	}{
		{
			name: "private game with players",
			game: api.GameResponse{
				ID:      "g1",
				Code:    "join-code",
				Status:  "PlayerX_TURN",
				PlayerX: "px",
				PlayerO: "po",
				Board:   "_________",
			},
			wantLines: []string{
				"game:   g1",
				"code:   join-code",
				"status: PlayerX_TURN",
				"X: px  O: po",
			},
		},
		{
			name: "public game without players hides optional lines",
			game: api.GameResponse{
				ID:     "g2",
				Status: "WAITING_FOR_PLAYERS",
				Board:  "_________",
			},
			wantLines:   []string{"game:   g2", "status: WAITING_FOR_PLAYERS"},
			absentLines: []string{"code:", "X: "},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := RenderGame(tc.game)
			for _, line := range tc.wantLines {
				if !strings.Contains(got, line) {
					t.Fatalf("RenderGame() = %q, missing %q", got, line)
				}
			}
			for _, line := range tc.absentLines {
				if strings.Contains(got, line) {
					t.Fatalf("RenderGame() = %q, must not contain %q", got, line)
				}
			}
		})
	}
}

func TestRenderGames(t *testing.T) {
	cases := []struct {
		name  string
		games []api.GameResponse
		want  string
	}{
		{
			name: "no games",
			want: "no games waiting for players\n",
		},
		{
			name: "one game per line",
			games: []api.GameResponse{
				{ID: "g1", Status: "WAITING_FOR_PLAYERS", IsPublic: true},
				{ID: "g2", Status: "WAITING_FOR_PLAYERS", IsPublic: false},
			},
			want: "g1  status=WAITING_FOR_PLAYERS  public=true\n" +
				"g2  status=WAITING_FOR_PLAYERS  public=false\n",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := RenderGames(tc.games); got != tc.want {
				t.Fatalf("RenderGames() = %q, want %q", got, tc.want)
			}
		})
	}
}
