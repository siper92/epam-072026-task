package internal

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"ticTacSolved/task/pkg/api"
	"ticTacSolved/task/pkg/errs"
)

const playerIDKey = "playerID"

func PlayerID(c *gin.Context) string {
	return c.GetString(playerIDKey)
}

func Logging() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		log.Printf(
			"%s %s %d %s",
			c.Request.Method,
			c.Request.URL.Path,
			c.Writer.Status(),
			time.Since(start),
		)
	}
}

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf(
					"panic on %s %s: %v",
					c.Request.Method,
					c.Request.URL.Path,
					rec,
				)
				c.AbortWithStatusJSON(
					http.StatusInternalServerError,
					api.ErrorResponse{
						Code:    codeInternal,
						Message: "internal server error",
					},
				)
			}
		}()

		c.Next()

		if err := c.Errors.Last(); err != nil {
			writeErr(c, err.Err)
		}
	}
}

func RequireSession(sessions SessionValidator) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := bearerToken(c.Request)
		if err != nil {
			c.Abort()
			_ = c.Error(err)
			return
		}
		playerID, err := sessions.ValidateSession(c.Request.Context(), token)
		if err != nil {
			c.Abort()
			_ = c.Error(err)
			return
		}
		c.Set(playerIDKey, playerID)
		c.Next()
	}
}

func bearerToken(r *http.Request) (string, error) {
	header := r.Header.Get(api.HeaderAuthorization)
	token, found := strings.CutPrefix(header, api.BearerPrefix)
	if !found || token == "" {
		return "", errs.New(errs.CodeInvalidToken, "missing bearer token")
	}
	return token, nil
}
