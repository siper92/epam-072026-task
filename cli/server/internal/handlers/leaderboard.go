package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"ticTacSolved/task/pkg/errs"
)

const defaultLeaderboardLimit = 10

type LeaderEntry struct {
	PlayerID string `json:"player_id"`
	Wins     int64  `json:"wins"`
	Losses   int64  `json:"losses"`
	Draws    int64  `json:"draws"`
}

type LeaderboardResponse struct {
	Leaders []LeaderEntry `json:"leaders"`
}

func (h *Handlers) Leaderboard(c *gin.Context) {
	limit, err := leaderboardLimit(c.Query("limit"))
	if err != nil {
		_ = c.Error(err)
		return
	}

	leaders, err := h.games.Leaders(c.Request.Context(), limit)
	if err != nil {
		_ = c.Error(err)
		return
	}

	resp := LeaderboardResponse{Leaders: make([]LeaderEntry, 0, len(leaders))}
	for _, leader := range leaders {
		resp.Leaders = append(resp.Leaders, LeaderEntry{
			PlayerID: leader.PlayerID,
			Wins:     leader.Wins,
			Losses:   leader.Losses,
			Draws:    leader.Draws,
		})
	}

	c.JSON(http.StatusOK, resp)
}

func leaderboardLimit(raw string) (int64, error) {
	if raw == "" {
		return defaultLeaderboardLimit, nil
	}
	limit, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || limit <= 0 {
		return 0, errs.Newf(errs.CodeInvalidInput, "invalid limit %q", raw)
	}
	return limit, nil
}
