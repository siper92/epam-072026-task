package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"strings"
	"ticTacSolved/task/game/data/gen"

	"ticTacSolved/task/game/auth"
	"ticTacSolved/task/game/data"
	"ticTacSolved/task/game/state_machine"
	"ticTacSolved/task/pkg/errs"
)

const (
	emptyBoard = "_________"
	markX      = "X"
	markO      = "O"
)

type Lobby interface {
	CreateGame(ctx context.Context, playerID string, public bool) (gen.Game, string, error)
	JoinGame(ctx context.Context, gameID string, playerID string, code string) (gen.Game, string, error)
	WaitingGames(ctx context.Context) ([]gen.Game, error)
}

type GamePlay interface {
	GetGame(ctx context.Context, gameID string) (gen.Game, error)
	MakeMove(ctx context.Context, token string, row int, col int) (gen.Game, error)
}

type Stats interface {
	Leaders(ctx context.Context, limit int64) ([]gen.Stat, error)
}

type Watcher interface {
	Watch(ctx context.Context, gameID string) (<-chan gen.Game, func(), error)
}

type GameService interface {
	Lobby
	GamePlay
	Validator
	Stats
	Watcher
}

type gameService struct {
	games  data.GameStore
	lobby  data.LobbyStore
	stats  data.StatsStore
	tokens auth.Service
	watch  *broadcaster
}

var _ GameService = (*gameService)(nil)

func NewGameService(
	games data.GameStore,
	lobby data.LobbyStore,
	stats data.StatsStore,
	tokens auth.Service,
) GameService {
	return &gameService{
		games:  games,
		lobby:  lobby,
		stats:  stats,
		tokens: tokens,
		watch:  newBroadcaster(),
	}
}

func (s *gameService) CreateGame(ctx context.Context, playerID string, public bool) (gen.Game, string, error) {
	if playerID == "" {
		return gen.Game{}, "", errs.New(errs.CodeInvalidInput, "player id is required")
	}

	game := gen.Game{
		ID:       randomHex(16),
		IsPublic: public,
		Board:    emptyBoard,
		Status:   string(state_machine.StatusWaitingForPlayers),
		PlayerX:  playerID,
	}
	if !public {
		game.Code = randomHex(4)
	}

	if err := s.games.CreateGame(ctx, game); err != nil {
		return gen.Game{}, "", err
	}

	token, err := s.issueGameToken(ctx, game, playerID, markX)
	if err != nil {
		return gen.Game{}, "", err
	}

	return game, token, nil
}

func (s *gameService) JoinGame(ctx context.Context, gameID string, playerID string, code string) (gen.Game, string, error) {
	if playerID == "" {
		return gen.Game{}, "", errs.New(errs.CodeInvalidInput, "player id is required")
	}
	game, err := s.games.GetGame(ctx, gameID)
	if err != nil {
		return gen.Game{}, "", err
	}
	if game.Status != string(state_machine.StatusWaitingForPlayers) {
		return gen.Game{}, "", errs.Newf(errs.CodeInvalidTransition, "game %q is not waiting for players", gameID)
	}
	if playerID == game.PlayerX {
		return gen.Game{}, "", errs.New(errs.CodeInvalidInput, "player already in the game")
	}
	if err = s.ValidateJoinCode(game, code); err != nil {
		return gen.Game{}, "", err
	}
	game.PlayerO = playerID
	game.Status = string(state_machine.StatusPlayerXTurn)
	if err = s.games.SetPlayerO(ctx, game.ID, playerID, game.Status); err != nil {
		return gen.Game{}, "", err
	}
	token, err := s.issueGameToken(ctx, game, playerID, markO)
	if err != nil {
		return gen.Game{}, "", err
	}
	s.watch.publish(game)
	return game, token, nil
}

func (s *gameService) WaitingGames(ctx context.Context) ([]gen.Game, error) {
	return s.lobby.ListWaitingGames(ctx, string(state_machine.StatusWaitingForPlayers))
}

func (s *gameService) GetGame(ctx context.Context, gameID string) (gen.Game, error) {
	return s.games.GetGame(ctx, gameID)
}

func (s *gameService) MakeMove(ctx context.Context, token string, row int, col int) (gen.Game, error) {
	gameToken, err := s.ValidateGameToken(ctx, token)
	if err != nil {
		return gen.Game{}, err
	}

	game, err := s.games.GetGame(ctx, gameToken.GameID)
	if err != nil {
		return gen.Game{}, err
	}
	if game.Code != gameToken.Code {
		return gen.Game{}, errs.New(errs.CodeInvalidToken, "token does not match the game code")
	}
	if game.Status == string(state_machine.StatusWaitingForPlayers) {
		return gen.Game{}, errs.New(errs.CodeInvalidTransition, "game has not started yet")
	}

	machine, err := state_machine.NewStateMachine(game.Board)
	if err != nil {
		return gen.Game{}, err
	}

	state := machine.GetCurrentState()
	if state.CurrentPlayer != "" && state.CurrentPlayer != gameToken.Mark {
		return gen.Game{}, errs.Newf(errs.CodeOutOfTurn, "it is not player %q turn", gameToken.Mark)
	}

	if err = machine.ProcessMove(row, col); err != nil {
		return gen.Game{}, err
	}

	state = machine.GetCurrentState()
	game.Board = encodeBoard(state.Board)
	game.Status = string(state.State)
	if err = s.games.UpdateGameState(ctx, game.ID, game.Board, game.Status); err != nil {
		return gen.Game{}, err
	}
	if err = s.recordFinish(ctx, game); err != nil {
		return gen.Game{}, err
	}
	s.watch.publish(game)

	return game, nil
}

func (s *gameService) Watch(
	ctx context.Context,
	gameID string,
) (<-chan gen.Game, func(), error) {
	if _, err := s.games.GetGame(ctx, gameID); err != nil {
		return nil, nil, err
	}
	updates, cancel := s.watch.subscribe(gameID)
	return updates, cancel, nil
}

func (s *gameService) Leaders(ctx context.Context, limit int64) ([]gen.Stat, error) {
	if limit <= 0 {
		limit = 10
	}
	return s.stats.ListLeaders(ctx, limit)
}

func (s *gameService) recordFinish(ctx context.Context, game gen.Game) error {
	switch state_machine.GameStatus(game.Status) {
	case state_machine.StatusGameOverPlayerXWin:
		return s.stats.RecordResult(ctx, game.PlayerX, game.PlayerO, false)
	case state_machine.StatusGameOverPlayerOWin:
		return s.stats.RecordResult(ctx, game.PlayerO, game.PlayerX, false)
	case state_machine.StatusGameOverDraw:
		return s.stats.RecordResult(ctx, game.PlayerX, game.PlayerO, true)
	}
	return nil
}

func (s *gameService) issueGameToken(ctx context.Context, game gen.Game, playerID string, mark string) (string, error) {
	claims := game.TokenData()
	claims[gen.ClaimPlayerID] = playerID
	claims[gen.ClaimMark] = mark
	return s.tokens.Issue(ctx, mapClaims(claims), auth.GameTokenTTL)
}

type mapClaims map[string]string

func (c mapClaims) TokenData() map[string]string { return c }

func encodeBoard(board [3][3]string) string {
	var b strings.Builder
	for _, row := range board {
		for _, cell := range row {
			b.WriteString(cell)
		}
	}
	return b.String()
}

func randomHex(n int) string {
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		panic(err)
	}
	return hex.EncodeToString(buf)
}
