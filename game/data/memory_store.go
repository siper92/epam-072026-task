package data

import (
	"context"
	"sort"
	"sync"
	"ticTacSolved/task/game/data/gen"

	"ticTacSolved/task/pkg/errs"
)

type MemoryStore struct {
	mu           sync.RWMutex
	games        map[string]gen.Game
	players      map[string]gen.Player
	playerTokens map[string]gen.Token
	stats        map[string]gen.Stat
}

var _ Store = (*MemoryStore)(nil)

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		games:        map[string]gen.Game{},
		players:      map[string]gen.Player{},
		playerTokens: map[string]gen.Token{},
		stats:        map[string]gen.Stat{},
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
	delete(s.playerTokens, playerID)
	s.playerTokens[playerID] = gen.Token{
		Token:     token,
		PlayerID:  playerID,
		ExpiresAt: expiresAt,
		Active:    true,
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

func (s *MemoryStore) RecordResult(
	_ context.Context,
	winnerID string,
	loserID string,
	draw bool,
) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if draw {
		s.addStat(winnerID, 0, 0, 1)
		s.addStat(loserID, 0, 0, 1)
		return nil
	}
	s.addStat(winnerID, 1, 0, 0)
	s.addStat(loserID, 0, 1, 0)
	return nil
}

func (s *MemoryStore) addStat(playerID string, wins int64, losses int64, draws int64) {
	stat := s.stats[playerID]
	stat.PlayerID = playerID
	stat.Wins += wins
	stat.Losses += losses
	stat.Draws += draws
	s.stats[playerID] = stat
}

func (s *MemoryStore) ListLeaders(_ context.Context, limit int64) ([]gen.Stat, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	leaders := make([]gen.Stat, 0, len(s.stats))
	for _, stat := range s.stats {
		leaders = append(leaders, stat)
	}
	sort.Slice(leaders, func(i, j int) bool {
		if leaders[i].Wins != leaders[j].Wins {
			return leaders[i].Wins > leaders[j].Wins
		}
		if leaders[i].Draws != leaders[j].Draws {
			return leaders[i].Draws > leaders[j].Draws
		}
		if leaders[i].Losses != leaders[j].Losses {
			return leaders[i].Losses < leaders[j].Losses
		}
		return leaders[i].PlayerID < leaders[j].PlayerID
	})
	if limit > 0 && int64(len(leaders)) > limit {
		leaders = leaders[:limit]
	}
	return leaders, nil
}
