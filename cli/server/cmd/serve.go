package cmd

import (
	"os"
	"os/signal"
	"strings"
	"syscall"
	"ticTacSolved/task/cli/server/internal"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	keyServerHost = "server.host"
	keyServerPort = "server.port"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the HTTP server for the game service",
	RunE:  runServe,
}

func init() {
	serveCmd.Flags().String("host", "127.0.0.1", "host the HTTP server binds to")
	serveCmd.Flags().Int("port", 8080, "port the HTTP server listens on")

	viper.SetEnvPrefix("GAME")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	mustBindFlag(keyServerHost, serveCmd, "host")
	mustBindFlag(keyServerPort, serveCmd, "port")

	rootCmd.AddCommand(serveCmd)
}

func mustBindFlag(key string, cmd *cobra.Command, flag string) {
	if err := viper.BindPFlag(key, cmd.Flags().Lookup(flag)); err != nil {
		panic(err)
	}
}

func runServe(cmd *cobra.Command, _ []string) error {
	ctx, stop := signal.NotifyContext(
		cmd.Context(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	srv := internal.NewServer(
		viper.GetString(keyServerHost),
		viper.GetInt(keyServerPort),
	)
	cmd.Printf("http server listening on %s\n", srv.Addr())

	return srv.Run(ctx)
}
