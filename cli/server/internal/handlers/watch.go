package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"ticTacSolved/task/game/data/gen"
	"ticTacSolved/task/game/state_machine"
)

func (h *Handlers) WatchGame(c *gin.Context) {
	game, err := h.games.GetGame(c.Request.Context(), c.Param("id"))
	if err != nil {
		_ = c.Error(err)
		return
	}

	updates, cancel, err := h.games.Watch(c.Request.Context(), game.ID)
	if err != nil {
		_ = c.Error(err)
		return
	}
	defer cancel()

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.WriteHeader(http.StatusOK)

	if !writeGameEvent(c, game) || gameFinished(game) {
		return
	}
	for {
		select {
		case <-c.Request.Context().Done():
			return
		case update, ok := <-updates:
			if !ok {
				return
			}
			if !writeGameEvent(c, update) || gameFinished(update) {
				return
			}
		}
	}
}

func writeGameEvent(c *gin.Context, game gen.Game) bool {
	payload, err := json.Marshal(toGameResponse(game, "", false))
	if err != nil {
		return false
	}
	if _, err = fmt.Fprintf(c.Writer, "data: %s\n\n", payload); err != nil {
		return false
	}
	c.Writer.Flush()
	return true
}

func gameFinished(game gen.Game) bool {
	switch state_machine.GameStatus(game.Status) {
	case state_machine.StatusGameOverDraw,
		state_machine.StatusGameOverPlayerXWin,
		state_machine.StatusGameOverPlayerOWin:
		return true
	}
	return false
}
