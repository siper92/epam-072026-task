package service

import (
	"context"
	"ticTacSolved/task/game/data/gen"

	"ticTacSolved/task/pkg/errs"
)

type GameToken struct {
	GameID   string
	PlayerID string
	Mark     string
	Code     string
}

type Validator interface {
	ValidateGameToken(ctx context.Context, token string) (GameToken, error)
	ValidateJoinCode(game gen.Game, code string) error
}

func (s *gameService) ValidateGameToken(ctx context.Context, token string) (GameToken, error) {
	claims, err := s.tokens.Validate(ctx, token)
	if err != nil {
		return GameToken{}, err
	}
	gameToken := GameToken{
		GameID:   claims[gen.ClaimGameID],
		PlayerID: claims[gen.ClaimPlayerID],
		Mark:     claims[gen.ClaimMark],
		Code:     claims[gen.ClaimGameCode],
	}
	if gameToken.GameID == "" || gameToken.PlayerID == "" {
		return GameToken{}, errs.New(errs.CodeInvalidToken, "missing game claims")
	}
	if gameToken.Mark != markX && gameToken.Mark != markO {
		return GameToken{}, errs.Newf(errs.CodeInvalidToken, "invalid player mark %q", gameToken.Mark)
	}
	return gameToken, nil
}

func (s *gameService) ValidateJoinCode(game gen.Game, code string) error {
	if game.IsPublic || game.Code == code {
		return nil
	}
	return errs.New(errs.CodeInvalidInput, "invalid game code")
}
