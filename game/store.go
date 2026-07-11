package game

import (
	"sync"

	"epam/task/pkg/errs"
)

type Store interface {
	Save(state GameState) error
	Load(gameID string) (GameState, error)
}

type MemoryStore struct {
	mu    sync.RWMutex
	games map[string]GameState
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{games: make(map[string]GameState)}
}

func (s *MemoryStore) Save(state GameState) error {
	if state.ID == "" {
		return errs.New(errs.CodeInvalidInput, "game state must have an ID")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.games[state.ID] = state
	return nil
}

func (s *MemoryStore) Load(gameID string) (GameState, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	state, ok := s.games[gameID]
	if !ok {
		return GameState{}, errs.Newf(errs.CodeGameNotFound, "game %q not found", gameID)
	}
	return state, nil
}
