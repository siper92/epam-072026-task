package client

import (
	"errors"
	"io/fs"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"ticTacSolved/task/cli/client/internal"
)

var cfg = viper.New()

var rootCmd = &cobra.Command{
	Use:   "ttt-client",
	Short: "tic tac toe game client",
	Long: "tic tac toe game client\n\n" +
		"type file runs one shot action commands and stores tokens on disk,\n" +
		"type cli starts an interactive play loop with the same token storage",
	SilenceUsage: true,
	PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
		return loadDotEnv()
	},
	RunE: runRoot,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	flags := rootCmd.PersistentFlags()
	flags.String(internal.KeyServer, "http://127.0.0.1:8080", "server base url")
	flags.String(internal.KeyUser, "guest", "user name used for login")
	flags.String(internal.KeyPassword, "guest", "password used for login")
	flags.String(internal.KeyType, internal.TypeFile, "client type: cli or file")
	flags.String(internal.KeyToken, "", "preset session token, skips login and refresh")
	flags.Int64(internal.KeyTokenTTL, 86400, "requested refresh token ttl in seconds")
	flags.Int64(internal.KeySessionTTL, 900, "requested session token ttl in seconds")
	flags.String(
		internal.KeySessionFile,
		internal.DefaultSessionFile(),
		"path of the stored session data",
	)

	cfg.SetEnvPrefix("TTT")
	cfg.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	cfg.AutomaticEnv()

	keys := []string{
		internal.KeyServer,
		internal.KeyUser,
		internal.KeyPassword,
		internal.KeyType,
		internal.KeyToken,
		internal.KeyTokenTTL,
		internal.KeySessionTTL,
		internal.KeySessionFile,
	}
	for _, key := range keys {
		mustBindFlag(key)
	}
}

func mustBindFlag(key string) {
	if err := cfg.BindPFlag(key, rootCmd.PersistentFlags().Lookup(key)); err != nil {
		panic(err)
	}
}

func loadDotEnv() error {
	dot := viper.New()
	dot.SetConfigFile(".env")
	dot.SetConfigType("env")
	if err := dot.ReadInConfig(); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		return err
	}
	for _, key := range dot.AllKeys() {
		cfg.SetDefault(dotEnvKey(key), dot.Get(key))
	}
	return nil
}

func dotEnvKey(key string) string {
	key = strings.TrimPrefix(strings.ToLower(key), "ttt_")
	return strings.ReplaceAll(key, "_", "-")
}

func runRoot(cmd *cobra.Command, _ []string) error {
	conf, err := internal.NewConfig(cfg)
	if err != nil {
		return err
	}
	if conf.Type == internal.TypeCLI {
		return runInteractive(cmd, conf)
	}
	return cmd.Help()
}

func newClient() (*internal.Client, error) {
	conf, err := internal.NewConfig(cfg)
	if err != nil {
		return nil, err
	}
	return newClientFor(conf), nil
}

func newClientFor(conf internal.Config) *internal.Client {
	return internal.NewClient(conf, internal.NewSessionStore(conf))
}
