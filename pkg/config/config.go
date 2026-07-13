package config

import (
	"os"
	"strings"
)

var env = map[string]string{}

func LoadEnv() {
	loaded := make(map[string]string)
	for _, entry := range os.Environ() {
		if key, value, ok := strings.Cut(entry, "="); ok {
			loaded[key] = value
		}
	}

	env = loaded
}

func GetEnv(key string) string {
	return env[key]
}
