package auth

import (
	"context"
	"time"

	"ticTacSolved/task/pkg/errs"
)

var (
	DefaultTokenTTL = 24 * time.Hour
	GameTokenTTL    = 12 * time.Hour
)

type TokenStore interface {
	SaveToken(ctx context.Context, token string, expiresAt int64) error
	GetTokenExpiry(ctx context.Context, token string) (int64, error)
}

type Service interface {
	Issue(ctx context.Context, source Tokenizable, ttl time.Duration) (string, error)
	Validate(ctx context.Context, token string) (map[string]string, error)
}

type service struct {
	store TokenStore
	now   func() time.Time
}

var _ Service = (*service)(nil)

func NewService(store TokenStore) Service {
	return &service{store: store, now: time.Now}
}

func (s *service) Issue(ctx context.Context, source Tokenizable, ttl time.Duration) (string, error) {
	if ttl <= 0 {
		ttl = DefaultTokenTTL
	}
	token, err := ToToken(source)
	if err != nil {
		return "", err
	}
	if err = s.store.SaveToken(ctx, token, s.now().Add(ttl).Unix()); err != nil {
		return "", err
	}
	return token, nil
}

func (s *service) Validate(ctx context.Context, token string) (map[string]string, error) {
	expiresAt, err := s.store.GetTokenExpiry(ctx, token)
	if err != nil {
		return nil, err
	}
	return ValidateToken(token, expiresAt, s.now())
}

func ValidateToken(token string, expiresAt int64, now time.Time) (map[string]string, error) {
	if now.Unix() >= expiresAt {
		return nil, errs.New(errs.CodeInvalidToken, "token expired")
	}
	return GetClaims(token)
}
