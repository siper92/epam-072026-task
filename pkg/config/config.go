package config

import (
	"os"
	"strings"

	"github.com/spf13/viper"
)

var env = viper.New()

func LoadEnv() {
	loaded := viper.New()
	loaded.AutomaticEnv()

	for _, arg := range os.Args[1:] {
		flag, isFlag := strings.CutPrefix(arg, "--")
		if !isFlag {
			continue
		}
		key, value, hasValue := strings.Cut(flag, "=")
		if !hasValue {
			continue
		}
		loaded.Set(normalizeKey(key), value)
	}

	env = loaded
}

func GetEnv(key string) string {
	return env.GetString(key)
}

func normalizeKey(key string) string {
	return strings.ReplaceAll(key, "-", "_")
}
