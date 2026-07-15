package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handlers) QueueJoin(c *gin.Context) {
	game, token, err := h.queue.Join(c.Request.Context(), PlayerID(c))
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, toGameResponse(game, token, false))
}
