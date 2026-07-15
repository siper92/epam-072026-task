package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/spf13/cobra"

	"ticTacSolved/task/game/state_machine"
	"ticTacSolved/task/pkg/api"
	"ticTacSolved/task/pkg/errs"
	"ticTacSolved/task/pkg/session"
)

const (
	StatusWaiting    = "waiting"
	StatusInProgress = "in_progress"
	StatusFinished   = "finished"
)

type Printer struct {
	json bool
}

func NewPrinter(output string) Printer {
	return Printer{json: output == OutputJSON}
}

type loginView struct {
	PlayerID     string `json:"player_id"`
	SessionToken string `json:"session_token"`
	RefreshToken string `json:"refresh_token"`
}

type createView struct {
	ID        string `json:"id"`
	GameToken string `json:"game_token"`
	JoinCode  string `json:"join_code,omitempty"`
}

type joinView struct {
	ID        string `json:"id"`
	GameToken string `json:"game_token"`
}

type gameView struct {
	ID     string       `json:"id"`
	Status string       `json:"status"`
	Board  [3][3]string `json:"board"`
	Next   string       `json:"next,omitempty"`
	Winner *string      `json:"winner"`
}

type gamesView struct {
	Games []gameView `json:"games"`
}

type leadersView struct {
	Leaders []LeaderEntry `json:"leaders"`
}

func (p Printer) Login(cmd *cobra.Command, data session.Data) {
	if p.json {
		p.write(cmd, loginView{
			PlayerID:     data.PlayerID,
			SessionToken: data.Session.Value,
			RefreshToken: data.Refresh.Value,
		})
		return
	}
	cmd.Printf("logged in as %s\n", data.PlayerID)
	cmd.Printf("session valid until %s\n", formatUnix(data.Session.ExpiresAt))
}

func (p Printer) Refreshed(cmd *cobra.Command, data session.Data) {
	if p.json {
		p.write(cmd, loginView{
			PlayerID:     data.PlayerID,
			SessionToken: data.Session.Value,
			RefreshToken: data.Refresh.Value,
		})
		return
	}
	cmd.Printf(
		"session refreshed, valid until %s\n",
		formatUnix(data.Session.ExpiresAt),
	)
}

func (p Printer) Created(cmd *cobra.Command, game api.GameResponse) {
	if p.json {
		p.write(cmd, createView{
			ID:        game.ID,
			GameToken: game.GameToken,
			JoinCode:  game.Code,
		})
		return
	}
	cmd.Print(RenderGame(game))
}

func (p Printer) Joined(cmd *cobra.Command, game api.GameResponse) {
	if p.json {
		p.write(cmd, joinView{ID: game.ID, GameToken: game.GameToken})
		return
	}
	cmd.Print(RenderGame(game))
}

func (p Printer) Game(cmd *cobra.Command, game api.GameResponse) {
	if p.json {
		p.write(cmd, toGameView(game))
		return
	}
	cmd.Print(RenderGame(game))
}

func (p Printer) Games(cmd *cobra.Command, games []api.GameResponse) {
	if p.json {
		view := gamesView{Games: make([]gameView, 0, len(games))}
		for _, game := range games {
			view.Games = append(view.Games, toGameView(game))
		}
		p.write(cmd, view)
		return
	}
	cmd.Print(RenderGames(games))
}

func (p Printer) Leaders(cmd *cobra.Command, leaders []LeaderEntry) {
	if p.json {
		view := leadersView{Leaders: leaders}
		if view.Leaders == nil {
			view.Leaders = make([]LeaderEntry, 0)
		}
		p.write(cmd, view)
		return
	}
	cmd.Print(RenderLeaders(leaders))
}

func (p Printer) write(cmd *cobra.Command, v any) {
	_ = json.NewEncoder(cmd.OutOrStdout()).Encode(v)
}

func PrintError(w io.Writer, output string, err error) {
	if output != OutputJSON {
		fmt.Fprintf(w, "error: %v\n", err)
		return
	}
	code := errs.CodeOf(err)
	if code == "" {
		code = errs.CodeInvalidInput
	}
	message := err.Error()
	var typed *errs.Error
	if errors.As(err, &typed) {
		message = typed.Message
	}
	_ = json.NewEncoder(w).Encode(api.ErrorResponse{
		Code:    string(code),
		Message: message,
	})
}

func toGameView(game api.GameResponse) gameView {
	return gameView{
		ID:     game.ID,
		Status: viewStatus(game.Status),
		Board:  viewBoard(game.Board),
		Next:   viewNext(game.Status),
		Winner: viewWinner(game.Status),
	}
}

func viewStatus(status string) string {
	switch state_machine.GameStatus(status) {
	case state_machine.StatusWaitingForPlayers:
		return StatusWaiting
	case state_machine.StatusGameOverDraw,
		state_machine.StatusGameOverPlayerXWin,
		state_machine.StatusGameOverPlayerOWin:
		return StatusFinished
	}
	return StatusInProgress
}

func viewNext(status string) string {
	switch state_machine.GameStatus(status) {
	case state_machine.StatusPlayerXTurn:
		return "x"
	case state_machine.StatusPlayerOTurn:
		return "o"
	}
	return ""
}

func viewWinner(status string) *string {
	var mark string
	switch state_machine.GameStatus(status) {
	case state_machine.StatusGameOverPlayerXWin:
		mark = "x"
	case state_machine.StatusGameOverPlayerOWin:
		mark = "o"
	default:
		return nil
	}
	return &mark
}

func viewBoard(board string) [3][3]string {
	var out [3][3]string
	cells := []rune(board)
	if len(cells) != 9 {
		return out
	}
	for i, r := range cells {
		cell := ""
		switch r {
		case 'X', 'x':
			cell = "x"
		case 'O', 'o':
			cell = "o"
		}
		out[i/3][i%3] = cell
	}
	return out
}

func formatUnix(unix int64) string {
	return time.Unix(unix, 0).Format(time.RFC3339)
}
