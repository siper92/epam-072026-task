package internal

import (
	"os"
	"path/filepath"

	"github.com/spf13/viper"

	"ticTacSolved/task/pkg/errs"
	"ticTacSolved/task/pkg/session"
)

const (
	TypeCLI  = "cli"
	TypeFile = "file"
)

const (
	OutputHuman = "human"
	OutputJSON  = "json"
)

const (
	KeyServer      = "server"
	KeyUser        = "user"
	KeyPassword    = "password"
	KeyType        = "type"
	KeyToken       = "token"
	KeyTokenTTL    = "token-ttl"
	KeySessionTTL  = "session-ttl"
	KeySessionFile = "session-file"
	KeyOutput      = "output"
	KeyGameToken   = "game-token"
)

type Config struct {
	ServerURL   string
	User        string
	Password    string
	Type        string
	Token       string
	TokenTTL    int64
	SessionTTL  int64
	SessionFile string
	Output      string
	GameToken   string
}

func NewConfig(v *viper.Viper) (Config, error) {
	conf := Config{
		ServerURL:   v.GetString(KeyServer),
		User:        v.GetString(KeyUser),
		Password:    v.GetString(KeyPassword),
		Type:        v.GetString(KeyType),
		Token:       v.GetString(KeyToken),
		TokenTTL:    v.GetInt64(KeyTokenTTL),
		SessionTTL:  v.GetInt64(KeySessionTTL),
		SessionFile: v.GetString(KeySessionFile),
		Output:      v.GetString(KeyOutput),
		GameToken:   v.GetString(KeyGameToken),
	}
	if conf.Type != TypeCLI && conf.Type != TypeFile {
		return Config{}, errs.Newf(
			errs.CodeInvalidInput,
			"invalid client type %q, expected cli or file",
			conf.Type,
		)
	}
	if conf.Output == "" {
		conf.Output = OutputHuman
	}
	if conf.Output != OutputHuman && conf.Output != OutputJSON {
		return Config{}, errs.Newf(
			errs.CodeInvalidInput,
			"invalid output %q, expected human or json",
			conf.Output,
		)
	}
	if conf.ServerURL == "" {
		return Config{}, errs.New(errs.CodeInvalidInput, "server url is required")
	}
	if conf.TokenTTL <= 0 || conf.SessionTTL <= 0 {
		return Config{}, errs.New(errs.CodeInvalidInput, "ttl values must be positive")
	}
	return conf, nil
}

func DefaultSessionFile() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".ttt", "session.json")
	}
	return filepath.Join(home, ".ttt", "session.json")
}

func NewSessionStore(conf Config) session.Store {
	if conf.SessionFile == "" {
		return session.NewMemoryStore()
	}
	return session.NewFileStore(conf.SessionFile)
}
