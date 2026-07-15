package session

import "sync"

type MemoryStore struct {
	mu   sync.RWMutex
	data Data
}

var _ Store = (*MemoryStore)(nil)

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{}
}

func (s *MemoryStore) Load() (Data, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.data, nil
}

func (s *MemoryStore) Save(data Data) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data = data
	return nil
}

func (s *MemoryStore) Clear() error {
	return s.Save(Data{})
}
