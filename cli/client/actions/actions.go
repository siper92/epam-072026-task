package actions

import (
	"time"

	"github.com/spf13/cobra"

	"ticTacSolved/task/cli/client/internal"
)

type ClientFactory func() (internal.GameClient, error)

func Command(newClient ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "action",
		Short: "perform a single client action against the server",
	}

	cmd.AddCommand(
		loginCommand(newClient),
		refreshCommand(newClient),
		listCommand(newClient),
		createCommand(newClient),
		joinCommand(newClient),
		showCommand(newClient),
		moveCommand(newClient),
	)

	return cmd
}

func formatUnix(unix int64) string {
	return time.Unix(unix, 0).Format(time.RFC3339)
}
