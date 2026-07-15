package handlers

import "github.com/gin-gonic/gin"

const playerIDKey = "playerID"

func PlayerID(c *gin.Context) string {
	return c.GetString(playerIDKey)
}

func SetPlayerID(c *gin.Context, playerID string) {
	c.Set(playerIDKey, playerID)
}
