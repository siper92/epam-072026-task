package cmd

import (
	"context"
	"database/sql"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"ticTacSolved/task/cli/server/internal"
	"ticTacSolved/task/game/data"
	"ticTacSolved/task/pkg/errs"
)

const (
	keyServerHost = "server.host"
	keyServerPort = "server.port"
	keyDBStorage  = "db.storage"
	keyDBPath     = "db.path"
	keyDBDriver   = "db.driver"

	storageMemory = "memory"
	storageSQLite = "sqlite"

	defaultDBPath = "./_local/db.sqlite3"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the HTTP server for the game service",
	RunE:  runServe,
}

func init() {
	serveCmd.Flags().String("host", "127.0.0.1", "host the HTTP server binds to")
	serveCmd.Flags().Int("port", 8080, "port the HTTP server listens on")
	serveCmd.Flags().String(
		"storage",
		storageMemory,
		"storage backend, allowed: memory, sqlite",
	)
	serveCmd.Flags().String(
		"db",
		defaultDBPath,
		"sqlite database file path, used when storage is sqlite",
	)

	viper.SetEnvPrefix("GAME")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
	viper.SetDefault(keyDBDriver, "sqlite")

	mustBindFlag(keyServerHost, serveCmd, "host")
	mustBindFlag(keyServerPort, serveCmd, "port")
	mustBindFlag(keyDBStorage, serveCmd, "storage")
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

	switch storage := viper.GetString(keyDBStorage); storage {
	case storageMemory:
	case storageSQLite:
		store, db, err := openSQLiteStore(ctx)
		if err != nil {
			return err
		}
		defer db.Close()
		cfg.Store = store
	default:
		return errs.Newf(
			errs.CodeInvalidInput,
			"unknown storage %q, allowed: %s, %s",
			storage,
			storageMemory,
			storageSQLite,
		)
	}

	srv := internal.NewServer(cfg)
	cmd.Printf("http server listening on %s\n", srv.Addr())

	return srv.Run(ctx)
}

func openSQLiteStore(ctx context.Context) (data.Store, *sql.DB, error) {
	dbPath := viper.GetString(keyDBPath)
	if dbPath == "" {
		dbPath = defaultDBPath
	}
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		return nil, nil, errs.Wrap(
			errs.CodeInvalidInput,
			"failed to create database directory",
			err,
		)
	}
	return data.OpenSQLStore(ctx, viper.GetString(keyDBDriver), dbPath)
}
