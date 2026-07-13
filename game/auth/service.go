package auth

import (
	"context"
	"strconv"
	"time"

	"ticTacSolved/task/pkg/errs"
)

const (
	ClaimPlayerID  = "player_id"
	ClaimExpiresAt = "exp"
)

var (
	DefaultTokenTTL = 1 * time.Hour
	GameTokenTTL    = 30 * time.Minute
)

type TokenStore interface {
	SaveToken(ctx context.Context, playerID string, token string, expiresAt int64) error
	GetTokenExpiry(ctx context.Context, token string) (int64, error)
}

type Service interface {
	Issue(ctx context.Context, source Tokenizable, ttl time.Duration) (string, error)
	Validate(ctx context.Context, token string) (map[string]string, error)
}

type service struct {
	store TokenStore
	cache *tokenCache
	now   func() time.Time
}

var _ Service = (*service)(nil)

func NewService(store TokenStore) Service {
	return &service{store: store, cache: liveTokens, now: time.Now}
}

func (s *service) Issue(
	ctx context.Context,
	source Tokenizable,
	ttl time.Duration,
) (string, error) {
	if ttl <= 0 {
		ttl = DefaultTokenTTL
	}
	claims := source.TokenData()
	playerID := claims[ClaimPlayerID]
	if playerID == "" {
		return "", errs.Newf(errs.CodeInvalidInput, "claim %q is required", ClaimPlayerID)
	}
	expiresAt := s.now().Add(ttl).Unix()
	claims[ClaimExpiresAt] = strconv.FormatInt(expiresAt, 10)

	token, err := MapToToken(claims)
	if err != nil {
		return "", err
	}
	if err = s.store.SaveToken(ctx, playerID, token, expiresAt); err != nil {
		return "", err
	}
	s.cache.put(playerID, token)

	return token, nil
}

func (s *service) Validate(ctx context.Context, token string) (map[string]string, error) {
	claims, err := ValidateToken(token, s.now())
	if err != nil {
		return nil, err
	}

	playerID := claims[ClaimPlayerID]
	if active, ok := s.cache.get(playerID); ok {
		if active != token {
			return nil, errs.New(errs.CodeInvalidToken, "token has been replaced")
		}
		return claims, nil
	}

	if _, err = s.store.GetTokenExpiry(ctx, token); err != nil {
		return nil, err
	}
	s.cache.put(playerID, token)

	return claims, nil
}

func ValidateToken(token string, now time.Time) (map[string]string, error) {
	claims, err := GetClaims(token)
	if err != nil {
		return nil, err
	}

	expiresAt, err := strconv.ParseInt(claims[ClaimExpiresAt], 10, 64)
	if err != nil {
		return nil, errs.Wrap(
			errs.CodeInvalidToken,
			"missing or invalid expiration claim",
			err,
		)
	}
	if now.Unix() >= expiresAt {
		return nil, errs.New(errs.CodeInvalidToken, "token expired")
	}

	return claims, nil
}
