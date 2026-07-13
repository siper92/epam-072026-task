package auth

import (
	"context"
	"strconv"
	"strings"
	"testing"
	"time"

	"ticTacSolved/task/pkg/errs"
)

type storedToken struct {
	token     string
	expiresAt int64
}

type fakeTokenStore struct {
	playerTokens map[string]storedToken
}

func newFakeTokenStore() *fakeTokenStore {
	return &fakeTokenStore{playerTokens: map[string]storedToken{}}
}

func (s *fakeTokenStore) SaveToken(
	_ context.Context,
	playerID string,
	token string,
	expiresAt int64,
) error {
	s.playerTokens[playerID] = storedToken{token: token, expiresAt: expiresAt}
	return nil
}

func (s *fakeTokenStore) GetTokenExpiry(_ context.Context, token string) (int64, error) {
	for _, stored := range s.playerTokens {
		if stored.token == token {
			return stored.expiresAt, nil
		}
	}
	return 0, errs.New(errs.CodeInvalidToken, "unknown token")
}

func newTestService(store TokenStore) *service {
	return &service{store: store, cache: newTokenCache(TokenCacheTTL), now: time.Now}
}

func expiringToken(t *testing.T, playerID string, expiresAt int64) string {
	t.Helper()
	token, err := MapToToken(map[string]string{
		ClaimPlayerID:  playerID,
		ClaimExpiresAt: strconv.FormatInt(expiresAt, 10),
	})
	if err != nil {
		t.Fatalf("MapToToken: %v", err)
	}
	return token
}

func TestValidateToken(t *testing.T) {
	now := time.Now()
	valid := expiringToken(t, "player-1", now.Add(time.Hour).Unix())
	parts := strings.Split(valid, ".")
	tampered := parts[0] + "." + parts[1] + "." + strings.Repeat("A", len(parts[2]))

	noExpiry, err := MapToToken(map[string]string{ClaimPlayerID: "player-1"})
	if err != nil {
		t.Fatalf("MapToToken: %v", err)
	}

	tests := []struct {
		name  string
		token string
		valid bool
	}{
		{"valid token", valid, true},
		{"expired token", expiringToken(t, "player-1", now.Add(-time.Hour).Unix()), false},
		{"missing expiration claim", noExpiry, false},
		{"tampered token", tampered, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			claims, err := ValidateToken(tc.token, now)
			if !tc.valid {
				if !errs.HasCode(err, errs.CodeInvalidToken) {
					t.Fatalf("expected CodeInvalidToken, got %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("ValidateToken: %v", err)
			}

			if claims[ClaimPlayerID] != "player-1" {
				t.Fatalf("unexpected claims: %v", claims)
			}
		})
	}
}

func TestServiceIssueAndValidate(t *testing.T) {
	store := newFakeTokenStore()
	svc := newTestService(store)
	ctx := context.Background()

	token, err := svc.Issue(ctx, testUser{ID: "player-1"}, 0)
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}

	stored := store.playerTokens["player-1"]
	if stored.token != token {
		t.Fatalf("expected token to be persisted for the player")
	}
	wantExpiry := time.Now().Add(DefaultTokenTTL).Unix()
	if stored.expiresAt < wantExpiry-5 || stored.expiresAt > wantExpiry+5 {
		t.Fatalf("expected expiry near %d, got %d", wantExpiry, stored.expiresAt)
	}

	claims, err := svc.Validate(ctx, token)
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if claims[ClaimPlayerID] != "player-1" {
		t.Fatalf("unexpected claims: %v", claims)
	}
	if claims[ClaimExpiresAt] != strconv.FormatInt(stored.expiresAt, 10) {
		t.Fatalf("expected expiration claim in token, got %v", claims)
	}
}

func TestServiceIssueRequiresPlayerID(t *testing.T) {
	svc := newTestService(newFakeTokenStore())

	if _, err := svc.Issue(context.Background(), testUser{}, 0); !errs.HasCode(err, errs.CodeInvalidInput) {
		t.Fatalf("expected CodeInvalidInput, got %v", err)
	}
}

func TestServiceSingleActiveTokenPerPlayer(t *testing.T) {
	store := newFakeTokenStore()
	svc := newTestService(store)
	ctx := context.Background()

	first, err := svc.Issue(ctx, testUser{ID: "player-1"}, time.Hour)
	if err != nil {
		t.Fatalf("Issue first: %v", err)
	}

	svc.now = func() time.Time { return time.Now().Add(time.Minute) }
	second, err := svc.Issue(ctx, testUser{ID: "player-1"}, time.Hour)
	if err != nil {
		t.Fatalf("Issue second: %v", err)
	}
	if first == second {
		t.Fatal("expected a different token on reissue")
	}

	if _, err = svc.Validate(ctx, second); err != nil {
		t.Fatalf("Validate second: %v", err)
	}
	if _, err = svc.Validate(ctx, first); !errs.HasCode(err, errs.CodeInvalidToken) {
		t.Fatalf("expected CodeInvalidToken for replaced token, got %v", err)
	}
}

func TestServiceValidateFallsBackToStore(t *testing.T) {
	store := newFakeTokenStore()
	svc := newTestService(store)
	ctx := context.Background()

	first, err := svc.Issue(ctx, testUser{ID: "player-1"}, time.Hour)
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}

	t.Run("uncached token restored from store", func(t *testing.T) {
		svc.cache = newTokenCache(TokenCacheTTL)
		if _, err := svc.Validate(ctx, first); err != nil {
			t.Fatalf("Validate: %v", err)
		}
		if cached, ok := svc.cache.get("player-1"); !ok || cached != first {
			t.Fatal("expected validated token to be cached again")
		}
	})

	t.Run("replaced token rejected via store", func(t *testing.T) {
		svc.now = func() time.Time { return time.Now().Add(time.Minute) }
		if _, err := svc.Issue(ctx, testUser{ID: "player-1"}, time.Hour); err != nil {
			t.Fatalf("Issue: %v", err)
		}

		svc.cache = newTokenCache(TokenCacheTTL)
		if _, err := svc.Validate(ctx, first); !errs.HasCode(err, errs.CodeInvalidToken) {
			t.Fatalf("expected CodeInvalidToken for replaced token, got %v", err)
		}
	})
}

func TestServiceValidateUnknownToken(t *testing.T) {
	svc := newTestService(newFakeTokenStore())
	token := expiringToken(t, "player-1", time.Now().Add(time.Hour).Unix())

	if _, err := svc.Validate(context.Background(), token); !errs.HasCode(err, errs.CodeInvalidToken) {
		t.Fatalf("expected CodeInvalidToken for unstored token, got %v", err)
	}
}

func TestServiceValidateExpiredToken(t *testing.T) {
	store := newFakeTokenStore()
	svc := newTestService(store)
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

func TestTokenCacheEviction(t *testing.T) {
	cache := newTokenCache(time.Minute)

	cache.put("player-1", "token-1")
	if token, ok := cache.get("player-1"); !ok || token != "token-1" {
		t.Fatalf("expected cached token, got %q (%v)", token, ok)
	}

	cache.now = func() time.Time { return time.Now().Add(2 * time.Minute) }
	if _, ok := cache.get("player-1"); ok {
		t.Fatal("expected stale entry to be ignored")
	}

	cache.evictStale()
	if len(cache.tokens) != 0 {
		t.Fatal("expected stale entries to be removed")
	}
}
