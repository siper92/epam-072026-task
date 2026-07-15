package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"ticTacSolved/task/pkg/api"
	"ticTacSolved/task/pkg/errs"
)

func (h *Handlers) Refresh(c *gin.Context) {
	var req api.RefreshRequest
	if err := decodeBody(c, &req); err != nil {
		_ = c.Error(err)
		return
	}

	if req.RefreshToken == "" {
		_ = c.Error(errs.New(errs.CodeInvalidInput, "refresh token is required"))
		return
	}

	session, err := h.auth.Refresh(
		c.Request.Context(),
		req.RefreshToken,
		req.SessionTTLSeconds,
	)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.RefreshResponse{Session: session})
}
