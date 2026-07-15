package session

import (
	"os"
	"path/filepath"
	"testing"

	"ticTacSolved/task/pkg/errs"
)

func TestTokenValid(t *testing.T) {
	cases := []struct {
		name  string
		token Token
		now   int64
		want  bool
	}{
		{name: "valid", token: Token{Value: "t", ExpiresAt: 100}, now: 50, want: true},
		{name: "expired", token: Token{Value: "t", ExpiresAt: 100}, now: 100, want: false},
		{name: "empty value", token: Token{ExpiresAt: 100}, now: 50, want: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.token.Valid(tc.now); got != tc.want {
				t.Fatalf("Valid() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestStoreRoundTrip(t *testing.T) {
	stores := []struct {
		name  string
		store Store
	}{
		{name: "memory", store: NewMemoryStore()},
		{name: "file", store: NewFileStore(filepath.Join(t.TempDir(), "dir", "session.json"))},
	}
	data := Data{
		ServerURL: "http://localhost:8080",
		PlayerID:  "p1",
		Session:   Token{Value: "session", ExpiresAt: 100},
		Refresh:   Token{Value: "refresh", ExpiresAt: 200},
		GameID:    "g1",
		GameToken: "game-token",
	}
	for _, tc := range stores {
		t.Run(tc.name, func(t *testing.T) {
			empty, err := tc.store.Load()
			if err != nil {
				t.Fatalf("Load() on empty store failed: %v", err)
			}
			if empty != (Data{}) {
				t.Fatalf("Load() on empty store = %+v, want zero data", empty)
			}

			if err = tc.store.Save(data); err != nil {
				t.Fatalf("Save() failed: %v", err)
			}
			loaded, err := tc.store.Load()
			if err != nil {
				t.Fatalf("Load() failed: %v", err)
			}
			if loaded != data {
				t.Fatalf("Load() = %+v, want %+v", loaded, data)
			}

			if err = tc.store.Clear(); err != nil {
				t.Fatalf("Clear() failed: %v", err)
			}
			cleared, err := tc.store.Load()
			if err != nil {
				t.Fatalf("Load() after Clear() failed: %v", err)
			}
			if cleared != (Data{}) {
				t.Fatalf("Load() after Clear() = %+v, want zero data", cleared)
			}
		})
	}
}

func TestFileStoreLoadInvalidJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "session.json")
	if err := os.WriteFile(path, []byte("{not json"), 0o600); err != nil {
		t.Fatalf("failed to seed file: %v", err)
	}

	_, err := NewFileStore(path).Load()
	if !errs.HasCode(err, errs.CodeStorageFailure) {
		t.Fatalf("Load() error = %v, want code %s", err, errs.CodeStorageFailure)
	}
}
