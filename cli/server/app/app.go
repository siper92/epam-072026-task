package app

import (
	"net/http"

	"ticTacSolved/task/cli/server/internal"
	"ticTacSolved/task/game/data"
)

func NewHandler(store data.Store) http.Handler {
	return internal.BuildHandler(store)
}
