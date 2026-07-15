package internal

import (
	"bufio"
	"context"
	"errors"
	"io"
	"os"
	"os/signal"
	"strings"
	"ticTacSolved/task/game/utils"
	"time"

	"github.com/spf13/cobra"

	"ticTacSolved/task/pkg/errs"
)

const interactiveHelp = `commands:
  list                 show public games waiting for players
  create [private]     create a game and enter it
  join <id> [code]     join a game, code needed for private games
  show                 show the current game
  move <row> <col>     make a move in the current game
  help                 show this help
  quit                 exit
`

func RunInteractive(cmd *cobra.Command, conf Config) error {
	c := NewClient(conf, NewSessionStore(conf))

	ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt)
	defer stop()

	lines := make(chan string)
	go readLines(cmd.InOrStdin(), lines)

	current := currentGameID(c)
	timeout := time.Duration(conf.SessionTTL) * time.Second
	cmd.Println("interactive tic tac toe client, type help for commands")
	for {
		cmd.Print("ttt> ")
		line, ok, timedOut := nextLine(ctx, lines, timeout)
		if timedOut {
			cmd.Println("\nsession ttl reached, exiting")
			return nil
		}
		if !ok {
			cmd.Println("\nbye")
			return nil
		}
		if quit := handleLine(ctx, cmd, c, &current, line); quit {
			return nil
		}
	}
}

func readLines(r io.Reader, lines chan<- string) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		lines <- scanner.Text()
	}
	close(lines)
}

func nextLine(
	ctx context.Context,
	lines <-chan string,
	timeout time.Duration,
) (line string, ok bool, timedOut bool) {
	readCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	select {
	case <-readCtx.Done():
		timedOut = errors.Is(readCtx.Err(), context.DeadlineExceeded)
		return "", false, timedOut
	case line, ok = <-lines:
		return line, ok, false
	}
}

func handleLine(
	ctx context.Context,
	cmd *cobra.Command,
	c GameClient,
	current *string,
	line string,
) bool {
	fields := strings.Fields(line)
	if len(fields) == 0 {
		return false
	}

	name, args := fields[0], fields[1:]
	switch name {
	case "quit", "exit":
		cmd.Println("bye")
		return true
	case "help":
		cmd.Print(interactiveHelp)
	default:
		if err := runInteractiveAction(ctx, cmd, c, current, name, args); err != nil {
			cmd.Printf("error: %v\n", err)
		}
	}
	return false
}

func runInteractiveAction(
	ctx context.Context,
	cmd *cobra.Command,
	c GameClient,
	current *string,
	name string,
	args []string,
) error {
	switch name {
	case "list":
		games, err := c.WaitingGames(ctx)
		if err != nil {
			return err
		}
		cmd.Print(RenderGames(games))
		return nil
	case "create":
		public := len(args) == 0 || args[0] != "private"
		game, err := c.CreateGame(ctx, public)
		if err != nil {
			return err
		}
		*current = game.ID
		cmd.Print(RenderGame(game))
		return nil
	case "join":
		if len(args) < 1 {
			return errs.New(errs.CodeInvalidInput, "usage: join <id> [code]")
		}
		code := ""
		if len(args) > 1 {
			code = args[1]
		}
		game, err := c.JoinGame(ctx, args[0], code)
		if err != nil {
			return err
		}
		*current = game.ID
		cmd.Print(RenderGame(game))
		return nil
	case "show":
		id := *current
		if len(args) > 0 {
			id = args[0]
		}
		if id == "" {
			return errs.New(errs.CodeInvalidInput, "no current game, join or create first")
		}
		game, err := c.GetGame(ctx, id)
		if err != nil {
			return err
		}
		cmd.Print(RenderGame(game))
		return nil
	case "move":
		if len(args) != 2 {
			return errs.New(errs.CodeInvalidInput, "usage: move <row> <col>")
		}
		row, col, err := utils.ParseCell(args[0], args[1])
		if err != nil {
			return err
		}
		game, err := c.Move(ctx, *current, row, col)
		if err != nil {
			return err
		}
		*current = game.ID
		cmd.Print(RenderGame(game))
		return nil
	}
	return errs.Newf(errs.CodeInvalidAction, "unknown command %q, type help", name)
}

func currentGameID(c GameClient) string {
	data, err := c.Session()
	if err != nil {
		return ""
	}
	return data.GameID
}
