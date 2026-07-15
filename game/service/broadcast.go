package service

import (
	"sync"

	"ticTacSolved/task/game/data/gen"
)

const watchBuffer = 8

type broadcaster struct {
	mu   sync.Mutex
	subs map[string]map[chan gen.Game]struct{}
}

func newBroadcaster() *broadcaster {
	return &broadcaster{subs: map[string]map[chan gen.Game]struct{}{}}
}

func (b *broadcaster) subscribe(gameID string) (<-chan gen.Game, func()) {
	ch := make(chan gen.Game, watchBuffer)

	b.mu.Lock()
	if b.subs[gameID] == nil {
		b.subs[gameID] = map[chan gen.Game]struct{}{}
	}
	b.subs[gameID][ch] = struct{}{}
	b.mu.Unlock()

	var once sync.Once
	cancel := func() {
		once.Do(func() {
			b.mu.Lock()
			delete(b.subs[gameID], ch)
			if len(b.subs[gameID]) == 0 {
				delete(b.subs, gameID)
			}
			close(ch)
			b.mu.Unlock()
		})
	}
	return ch, cancel
}

func (b *broadcaster) publish(game gen.Game) {
	b.mu.Lock()
	defer b.mu.Unlock()
	for ch := range b.subs[game.ID] {
		select {
		case ch <- game:
		default:
		}
	}
}
