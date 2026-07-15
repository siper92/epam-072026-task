package api

import (
	"ticTacSolved/task/pkg/errs"
)

const (
	HeaderAuthorization = "Authorization"
	HeaderGameToken     = "X-Game-Token"
	BearerPrefix        = "Bearer "
)

const (
	PathLogin   = "/api/login"
	PathRefresh = "/api/refresh"
	PathGames   = "/api/games"
)

func GamePath(id string) string { return PathGames + "/" + id }
func JoinPath(id string) string { return GamePath(id) + "/join" }
func MovePath(id string) string { return GamePath(id) + "/move" }

type Token struct {
	Value     string `json:"value"`
	ExpiresAt int64  `json:"expires_at"`
}

type LoginRequest struct {
	User              string `json:"user"`
	Password          string `json:"password"`
	SessionTTLSeconds int64  `json:"session_ttl_seconds,omitempty"`
	RefreshTTLSeconds int64  `json:"refresh_ttl_seconds,omitempty"`
}

type LoginResponse struct {
	PlayerID string `json:"player_id"`
	Session  Token  `json:"session"`
	Refresh  Token  `json:"refresh"`
}

type RefreshRequest struct {
	RefreshToken      string `json:"refresh_token"`
	SessionTTLSeconds int64  `json:"session_ttl_seconds,omitempty"`
}

type RefreshResponse struct {
	Session Token `json:"session"`
}

type CreateGameRequest struct {
	Public bool `json:"public"`
}

type JoinGameRequest struct {
	Code string `json:"code"`
}

type MoveRequest struct {
	Row int `json:"row"`
	Col int `json:"col"`
}

type GameResponse struct {
	ID        string `json:"id"`
	Board     string `json:"board"`
	Status    string `json:"status"`
	PlayerX   string `json:"player_x"`
	PlayerO   string `json:"player_o"`
	IsPublic  bool   `json:"is_public"`
	Code      string `json:"code,omitempty"`
	GameToken string `json:"game_token,omitempty"`
}

type GamesResponse struct {
	Games []GameResponse `json:"games"`
}

type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e ErrorResponse) Err() error {
	return errs.New(errs.Code(e.Code), e.Message)
}
