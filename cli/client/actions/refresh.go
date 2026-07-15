package actions

import (
	"github.com/spf13/cobra"
)

func refreshCommand(newClient ClientFactory, newPrinter PrinterFactory) *cobra.Command {
	return &cobra.Command{
		Use:   "refresh",
		Short: "get a new session token using the stored refresh token",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			c, err := newClient()
			if err != nil {
				return err
			}

			data, err := c.Refresh(cmd.Context())
			if err != nil {
				return err
			}

			newPrinter().Refreshed(cmd, data)
			return nil
		},
	}
}
