package actions

import (
	"github.com/spf13/cobra"
)

func loginCommand(newClient ClientFactory) *cobra.Command {
	return &cobra.Command{
		Use:   "login",
		Short: "login with user and password and store fresh tokens",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
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
		},
	}
}
