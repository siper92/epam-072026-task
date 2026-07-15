package actions

import (
	"github.com/spf13/cobra"

	"ticTacSolved/task/cli/client/internal"
)

type ClientFactory func() (internal.GameClient, error)

type PrinterFactory func() internal.Printer

func Command(newClient ClientFactory, newPrinter PrinterFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "action",
		Short: "perform a single client action against the server",
	}

	cmd.AddCommand(
		loginCommand(newClient, newPrinter),
		refreshCommand(newClient, newPrinter),
		listCommand(newClient, newPrinter),
		createCommand(newClient, newPrinter),
		joinCommand(newClient, newPrinter),
		queueCommand(newClient, newPrinter),
		leaderboardCommand(newClient, newPrinter),
		showCommand(newClient, newPrinter),
		stateCommand(newClient, newPrinter),
		moveCommand(newClient, newPrinter),
	)

	return cmd
}
