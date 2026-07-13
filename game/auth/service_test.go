package auth

import (
	"context"
	"strings"
	"testing"
	"time"

	"epam/task/pkg/errs"
)

type fakeTokenStore struct {
	tokens map[string]int64
}

func newFakeTokenStore() *fakeTokenStore {
	return &fakeTokenStore{tokens: map[string]int64{}}
}

func (s *fakeTokenStore) SaveToken(_ context.Context, token string, expiresAt int64) error {
	s.tokens[token] = expiresAt
	return nil
}

func (s *fakeTokenStore) GetTokenExpiry(_ context.Context, token string) (int64, error) {
	expiresAt, ok := s.tokens[token]
	if !ok {
		return 0, errs.New(errs.CodeInvalidToken, "unknown token")
	}
	return expiresAt, nil
}

func TestValidateToken(t *testing.T) {
	token, err := MapToToken(map[string]string{"sub": "player-1"})
	if err != nil {
		t.Fatalf("MapToToken: %v", err)
	}
	now := time.Now()

	claims, err := ValidateToken(token, now.Add(time.Hour).Unix(), now)
	if err != nil {
		t.Fatalf("ValidateToken: %v", err)
	}
	if claims["sub"] != "player-1" {
		t.Fatalf("unexpected claims: %v", claims)
	}

	if _, err = ValidateToken(token, now.Add(-time.Hour).Unix(), now); !errs.HasCode(err, errs.CodeInvalidToken) {
		t.Fatalf("expected CodeInvalidToken for expired token, got %v", err)
	}

	parts := strings.Split(token, ".")
	tampered := parts[0] + "." + parts[1] + "." + strings.Repeat("A", len(parts[2]))
	if _, err = ValidateToken(tampered, now.Add(time.Hour).Unix(), now); !errs.HasCode(err, errs.CodeInvalidToken) {
		t.Fatalf("expected CodeInvalidToken for tampered token, got %v", err)
	}
}

func TestServiceIssueAndValidate(t *testing.T) {
	store := newFakeTokenStore()
	svc := NewService(store)
	ctx := context.Background()

	token, err := svc.Issue(ctx, testUser{ID: "player-1"}, 0)
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}

	wantExpiry := time.Now().Add(DefaultTokenTTL).Unix()
	if got := store.tokens[token]; got < wantExpiry-5 || got > wantExpiry+5 {
		t.Fatalf("expected expiry near %d, got %d", wantExpiry, got)
	}

	claims, err := svc.Validate(ctx, token)
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if claims["sub"] != "player-1" {
		t.Fatalf("unexpected claims: %v", claims)
	}
}

func TestServiceValidateUnknownToken(t *testing.T) {
	svc := NewService(newFakeTokenStore())

	token, err := MapToToken(map[string]string{"sub": "player-1"})
	if err != nil {
		t.Fatalf("MapToToken: %v", err)
	}

	if _, err = svc.Validate(context.Background(), token); !errs.HasCode(err, errs.CodeInvalidToken) {
		t.Fatalf("expected CodeInvalidToken for unstored token, got %v", err)
	}
}

func TestServiceValidateExpiredToken(t *testing.T) {
	store := newFakeTokenStore()
	svc := NewService(store).(*service)
	ctx := context.Background()

	token, err := svc.Issue(ctx, testUser{ID: "player-1"}, time.Minute)
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}

	svc.now = func() time.Time { return time.Now().Add(time.Hour) }
	if _, err = svc.Validate(ctx, token); !errs.HasCode(err, errs.CodeInvalidToken) {
		t.Fatalf("expected CodeInvalidToken for expired token, got %v", err)
	}
}
