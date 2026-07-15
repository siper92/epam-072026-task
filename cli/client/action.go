package client

import (
	"strconv"
	"time"

	"github.com/spf13/cobra"

	"ticTacSolved/task/cli/client/internal"
	"ticTacSolved/task/pkg/errs"
)

var actionCmd = &cobra.Command{
	Use:   "action",
	Short: "perform a single client action against the server",
}

var loginActionCmd = &cobra.Command{
	Use:   "login",
	Short: "login with user and password and store fresh tokens",
	Args:  cobra.NoArgs,
	RunE:  runLoginAction,
}

var refreshActionCmd = &cobra.Command{
	Use:   "refresh",
	Short: "get a new session token using the stored refresh token",
	Args:  cobra.NoArgs,
	RunE:  runRefreshAction,
}

var listActionCmd = &cobra.Command{
	Use:   "list",
	Short: "list public games waiting for players",
	Args:  cobra.NoArgs,
	RunE:  runListAction,
}

var createActionCmd = &cobra.Command{
	Use:   "create",
	Short: "create a new game and store its game token",
	Args:  cobra.NoArgs,
	RunE:  runCreateAction,
}

var joinActionCmd = &cobra.Command{
	Use:   "join <game-id>",
	Short: "join a game and store its game token",
	Args:  cobra.ExactArgs(1),
	RunE:  runJoinAction,
}

var showActionCmd = &cobra.Command{
	Use:   "show [game-id]",
	Short: "show a game, defaults to the stored current game",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runShowAction,
}

var moveActionCmd = &cobra.Command{
	Use:   "move [game-id] <row> <col>",
	Short: "make a move, game id defaults to the stored current game",
	Args:  cobra.RangeArgs(2, 3),
	RunE:  runMoveAction,
}

func init() {
	createActionCmd.Flags().Bool("private", false, "create a private game with a join code")
	joinActionCmd.Flags().String("code", "", "join code for a private game")

	actionCmd.AddCommand(
		loginActionCmd,
		refreshActionCmd,
		listActionCmd,
		createActionCmd,
		joinActionCmd,
		showActionCmd,
		moveActionCmd,
	)
	rootCmd.AddCommand(actionCmd)
}

func runLoginAction(cmd *cobra.Command, _ []string) error {
	c, err := newClient()
	if err != nil {
		return err
	}
	data, err := c.Login(cmd.Context())
	if err != nil {
		return err
	}
	cmd.Printf("logged in as %s\n", data.PlayerID)
	cmd.Printf("session valid until %s\n", formatUnix(data.Session.ExpiresAt))
	return nil
}

func runRefreshAction(cmd *cobra.Command, _ []string) error {
	c, err := newClient()
	if err != nil {
		return err
	}
	data, err := c.Refresh(cmd.Context())
	if err != nil {
		return err
	}
	cmd.Printf("session refreshed, valid until %s\n", formatUnix(data.Session.ExpiresAt))
	return nil
}

func runListAction(cmd *cobra.Command, _ []string) error {
	c, err := newClient()
	if err != nil {
		return err
	}
	games, err := c.WaitingGames(cmd.Context())
	if err != nil {
		return err
	}
	cmd.Print(internal.RenderGames(games))
	return nil
}

func runCreateAction(cmd *cobra.Command, _ []string) error {
	c, err := newClient()
	if err != nil {
		return err
	}
	private, err := cmd.Flags().GetBool("private")
	if err != nil {
		return err
	}
	game, err := c.CreateGame(cmd.Context(), !private)
	if err != nil {
		return err
	}
	cmd.Print(internal.RenderGame(game))
	return nil
}

func runJoinAction(cmd *cobra.Command, args []string) error {
	c, err := newClient()
	if err != nil {
		return err
	}
	code, err := cmd.Flags().GetString("code")
	if err != nil {
		return err
	}
	game, err := c.JoinGame(cmd.Context(), args[0], code)
	if err != nil {
		return err
	}
	cmd.Print(internal.RenderGame(game))
	return nil
}

func runShowAction(cmd *cobra.Command, args []string) error {
	c, err := newClient()
	if err != nil {
		return err
	}
	id := ""
	if len(args) == 1 {
		id = args[0]
	} else {
		data, err := c.Session()
		if err != nil {
			return err
		}
		id = data.GameID
	}
	if id == "" {
		return errs.New(errs.CodeInvalidInput, "game id is required")
	}
	game, err := c.GetGame(cmd.Context(), id)
	if err != nil {
		return err
	}
	cmd.Print(internal.RenderGame(game))
	return nil
}

func runMoveAction(cmd *cobra.Command, args []string) error {
	c, err := newClient()
	if err != nil {
		return err
	}
	id := ""
	cellArgs := args
	if len(args) == 3 {
		id = args[0]
		cellArgs = args[1:]
	}
	row, col, err := parseCell(cellArgs[0], cellArgs[1])
	if err != nil {
		return err
	}
	game, err := c.Move(cmd.Context(), id, row, col)
	if err != nil {
		return err
	}
	cmd.Print(internal.RenderGame(game))
	return nil
}

func parseCell(rowArg string, colArg string) (int, int, error) {
	row, err := strconv.Atoi(rowArg)
	if err != nil {
		return 0, 0, errs.Newf(errs.CodeInvalidInput, "invalid row %q", rowArg)
	}
	col, err := strconv.Atoi(colArg)
	if err != nil {
		return 0, 0, errs.Newf(errs.CodeInvalidInput, "invalid col %q", colArg)
	}
	return row, col, nil
}

func formatUnix(unix int64) string {
	return time.Unix(unix, 0).Format(time.RFC3339)
}
