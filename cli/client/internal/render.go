package internal

import (
	"fmt"
	"strings"

	"ticTacSolved/task/pkg/api"
)

func RenderGame(game api.GameResponse) string {
	var b strings.Builder
	fmt.Fprintf(&b, "game:   %s\n", game.ID)
	if game.Code != "" {
		fmt.Fprintf(&b, "code:   %s\n", game.Code)
	}
	fmt.Fprintf(&b, "status: %s\n", game.Status)
	if game.PlayerX != "" || game.PlayerO != "" {
		fmt.Fprintf(&b, "X: %s  O: %s\n", game.PlayerX, game.PlayerO)
	}
	b.WriteString(RenderBoard(game.Board))
	return b.String()
}

func RenderBoard(board string) string {
	cells := []rune(board)
	if len(cells) != 9 {
		return board + "\n"
	}

	var b strings.Builder
	for row := 0; row < 3; row++ {
		if row > 0 {
			b.WriteString("---+---+---\n")
		}
		for col := 0; col < 3; col++ {
			if col > 0 {
				b.WriteString("|")
			}
			fmt.Fprintf(&b, " %c ", cells[row*3+col])
		}
		b.WriteString("\n")
	}
	return b.String()
}

func RenderLeaders(leaders []LeaderEntry) string {
	if len(leaders) == 0 {
		return "no results recorded yet\n"
	}

	var b strings.Builder
	for i, leader := range leaders {
		fmt.Fprintf(
			&b,
			"%d. %s  wins=%d losses=%d draws=%d\n",
			i+1, leader.PlayerID, leader.Wins, leader.Losses, leader.Draws,
		)
	}
	return b.String()
}

func RenderGames(games []api.GameResponse) string {
	if len(games) == 0 {
		return "no games waiting for players\n"
	}

	var b strings.Builder
	for _, game := range games {
		fmt.Fprintf(&b, "%s  status=%s  public=%v\n", game.ID, game.Status, game.IsPublic)
	}
	return b.String()
}
