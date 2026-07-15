package cmd

import (
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"ticTacSolved/task/cli/server/internal"
	"ticTacSolved/task/game/data"
)

const (
	keyServerHost = "server.host"
	keyServerPort = "server.port"
	keyDBPath     = "db.path"
	keyDBDriver   = "db.driver"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the HTTP server for the game service",
	RunE:  runServe,
}

func init() {
	serveCmd.Flags().String("host", "127.0.0.1", "host the HTTP server binds to")
	serveCmd.Flags().Int("port", 8080, "port the HTTP server listens on")
	serveCmd.Flags().String("db", "", "sqlite database path, empty means in-memory store")

	viper.SetEnvPrefix("GAME")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
	viper.SetDefault(keyDBDriver, "sqlite")

	mustBindFlag(keyServerHost, serveCmd, "host")
	mustBindFlag(keyServerPort, serveCmd, "port")
	mustBindFlag(keyDBPath, serveCmd, "db")

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

	cfg := internal.ServerConfig{
		Host: viper.GetString(keyServerHost),
		Port: viper.GetInt(keyServerPort),
	}
	if dbPath := viper.GetString(keyDBPath); dbPath != "" {
		store, db, err := data.OpenSQLStore(
			ctx,
			viper.GetString(keyDBDriver),
			dbPath,
		)
		if err != nil {
			return err
		}
		defer db.Close()
		cfg.Store = store
	}

	srv := internal.NewServer(cfg)
	cmd.Printf("http server listening on %s\n", srv.Addr())

	return srv.Run(ctx)
}
