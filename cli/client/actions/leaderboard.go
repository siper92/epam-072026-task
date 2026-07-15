package actions

import (
	"github.com/spf13/cobra"
)

func leaderboardCommand(newClient ClientFactory, newPrinter PrinterFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "leaderboard",
		Short: "list the best players by recorded results",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			c, err := newClient()
			if err != nil {
				return err
			}
			limit, err := cmd.Flags().GetInt64("limit")
			if err != nil {
				return err
			}
			leaders, err := c.Leaderboard(cmd.Context(), limit)
			if err != nil {
				return err
			}
			newPrinter().Leaders(cmd, leaders)
			return nil
		},
	}
	cmd.Flags().Int64("limit", 10, "maximum number of players to list")
	return cmd
}
