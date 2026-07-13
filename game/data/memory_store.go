package data

import (
	"context"
	"sync"
	"ticTacSolved/task/game/data/gen"

	"ticTacSolved/task/pkg/errs"
)

type MemoryStore struct {
	mu           sync.RWMutex
	games        map[string]gen.Game
	players      map[string]gen.Player
	playerTokens map[string]gen.Token
}

var _ Store = (*MemoryStore)(nil)

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		games:        map[string]gen.Game{},
		players:      map[string]gen.Player{},
		playerTokens: map[string]gen.Token{},
	}
}

func (s *MemoryStore) CreateGame(_ context.Context, game gen.Game) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.games[game.ID]; ok {
		return errs.Newf(errs.CodeInvalidInput, "game %q already exists", game.ID)
	}
	s.games[game.ID] = game
	return nil
}

func (s *MemoryStore) GetGame(_ context.Context, id string) (gen.Game, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	game, ok := s.games[id]
	if !ok {
		return gen.Game{}, errs.Newf(errs.CodeNotFound, "game %q not found", id)
	}
	return game, nil
}

func (s *MemoryStore) UpdateGameState(ctx context.Context, id string, board string, status string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	game, ok := s.games[id]
	if !ok {
		return errs.Newf(errs.CodeNotFound, "game %q not found", id)
	}
	game.Board = board
	game.Status = status
	s.games[id] = game
	return nil
}

func (s *MemoryStore) SetPlayerO(_ context.Context, id string, playerID string, status string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	game, ok := s.games[id]
	if !ok {
		return errs.Newf(errs.CodeNotFound, "game %q not found", id)
	}
	game.PlayerO = playerID
	game.Status = status
	s.games[id] = game
	return nil
}

func (s *MemoryStore) CreatePlayer(_ context.Context, player gen.Player) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.players[player.ID]; ok {
		return errs.Newf(errs.CodeInvalidInput, "player %q already exists", player.ID)
	}
	s.players[player.ID] = player
	return nil
}

func (s *MemoryStore) GetPlayer(_ context.Context, id string) (gen.Player, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	player, ok := s.players[id]
	if !ok {
		return gen.Player{}, errs.Newf(errs.CodeNotFound, "player %q not found", id)
	}
	return player, nil
}

func (s *MemoryStore) ListWaitingGames(_ context.Context, status string) ([]gen.Game, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var games []gen.Game
	for _, game := range s.games {
		if game.IsPublic && game.Status == status {
			games = append(games, game)
		}
	}
	return games, nil
}

func (s *MemoryStore) SaveToken(
	_ context.Context,
	playerID string,
	token string,
	expiresAt int64,
) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.playerTokens[playerID] = gen.Token{
		Token:     token,
		PlayerID:  playerID,
		ExpiresAt: expiresAt,
	}
	return nil
}

func (s *MemoryStore) GetTokenExpiry(_ context.Context, token string) (int64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, stored := range s.playerTokens {
		if stored.Token == token {
			return stored.ExpiresAt, nil
		}
	}
	return 0, errs.New(errs.CodeInvalidToken, "unknown token")
}
