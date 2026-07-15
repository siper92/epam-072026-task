package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"ticTacSolved/task/pkg/api"
)

func (h *Handlers) Login(c *gin.Context) {
	var req api.LoginRequest
	if err := decodeBody(c, &req); err != nil {
		_ = c.Error(err)
		return
	}

	result, err := h.auth.Login(
		c.Request.Context(),
		req.User,
		req.Password,
		req.SessionTTLSeconds,
		req.RefreshTTLSeconds,
	)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.LoginResponse{
		PlayerID: result.PlayerID,
		Session:  result.Session,
		Refresh:  result.Refresh,
	})
}
