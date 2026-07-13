package gen

import (
	"ticTacSolved/task/game/auth"
)

const (
	ClaimPlayerID   = auth.ClaimPlayerID
	ClaimPlayerName = "player_name"
	ClaimGameID     = "game_id"
	ClaimGameCode   = "game_code"
	ClaimMark       = "mark"
)

type PlayerModel interface {
	auth.Tokenizable
	GetID() string
	GetName() string
}

type GameModel interface {
	auth.Tokenizable
	GetID() string
	GetCode() string
	GetIsPublic() bool
	GetBoard() string
	GetStatus() string
	GetPlayerX() string
	GetPlayerO() string
}

var (
	_ PlayerModel = Player{}
	_ GameModel   = Game{}
)

func (p Player) GetID() string   { return p.ID }
func (p Player) GetName() string { return p.Name }

func (p Player) TokenData() map[string]string {
	return map[string]string{
		ClaimPlayerID:   p.ID,
		ClaimPlayerName: p.Name,
	}
}

func (g Game) GetID() string      { return g.ID }
func (g Game) GetCode() string    { return g.Code }
func (g Game) GetIsPublic() bool  { return g.IsPublic }
func (g Game) GetBoard() string   { return g.Board }
func (g Game) GetStatus() string  { return g.Status }
func (g Game) GetPlayerX() string { return g.PlayerX }
func (g Game) GetPlayerO() string { return g.PlayerO }

func (g Game) TokenData() map[string]string {
	return map[string]string{
		ClaimGameID:   g.ID,
		ClaimGameCode: g.Code,
	}
}
