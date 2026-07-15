package internal

import (
	"context"
	"time"

	"ticTacSolved/task/pkg/api"
	"ticTacSolved/task/pkg/errs"
	"ticTacSolved/task/pkg/session"
)

type authRequests interface {
	loginRequest(ctx context.Context) (api.LoginResponse, error)
	refreshRequest(ctx context.Context, refreshToken string) (api.RefreshResponse, error)
}

type TokenManager struct {
	cfg   Config
	store session.Store
	raw   authRequests
	now   func() time.Time
}

func NewTokenManager(
	cfg Config,
	store session.Store,
	raw authRequests,
) *TokenManager {
	return &TokenManager{cfg: cfg, store: store, raw: raw, now: time.Now}
}

func (m *TokenManager) Data() (session.Data, error) {
	return m.store.Load()
}

func (m *TokenManager) SessionToken(ctx context.Context) (string, error) {
	if m.cfg.Token != "" {
		return m.presetToken()
	}

	data, err := m.store.Load()
	if err != nil {
		return "", err
	}

	now := m.now().Unix()
	if data.Session.Valid(now) {
		return data.Session.Value, nil
	}
	if data.Refresh.Valid(now) {
		refreshed, err := m.Refresh(ctx)
		if err != nil {
			return "", err
		}
		return refreshed.Session.Value, nil
	}

	loggedIn, err := m.Login(ctx)
	if err != nil {
		return "", err
	}
	return loggedIn.Session.Value, nil
}

func (m *TokenManager) presetToken() (string, error) {
	data, err := m.store.Load()
	if err != nil {
		return "", err
	}
	if data.Session.Value == m.cfg.Token {
		return m.cfg.Token, nil
	}

	data.ServerURL = m.cfg.ServerURL
	data.Session = session.Token{
		Value:     m.cfg.Token,
		ExpiresAt: m.expiresIn(m.cfg.SessionTTL),
	}
	if err = m.store.Save(data); err != nil {
		return "", err
	}
	return m.cfg.Token, nil
}

func (m *TokenManager) Login(ctx context.Context) (session.Data, error) {
	if m.cfg.User == "" || m.cfg.Password == "" {
		return session.Data{}, errs.New(
			errs.CodeInvalidInput,
			"login required: user and password must be set",
		)
	}

	resp, err := m.raw.loginRequest(ctx)
	if err != nil {
		return session.Data{}, err
	}

	data, err := m.store.Load()
	if err != nil {
		return session.Data{}, err
	}
	data.ServerURL = m.cfg.ServerURL
	data.PlayerID = resp.PlayerID
	data.Session = session.Token(resp.Session)
	data.Refresh = session.Token(resp.Refresh)
	if err = m.store.Save(data); err != nil {
		return session.Data{}, err
	}
	return data, nil
}

func (m *TokenManager) Refresh(ctx context.Context) (session.Data, error) {
	data, err := m.store.Load()
	if err != nil {
		return session.Data{}, err
	}
	if !data.Refresh.Valid(m.now().Unix()) {
		return session.Data{}, errs.New(
			errs.CodeInvalidToken,
			"no valid refresh token, login required",
		)
	}

	resp, err := m.raw.refreshRequest(ctx, data.Refresh.Value)
	if err != nil {
		return session.Data{}, err
	}
	data.Session = session.Token(resp.Session)
	if err = m.store.Save(data); err != nil {
		return session.Data{}, err
	}
	return data, nil
}

func (m *TokenManager) SaveGame(gameID string, gameToken string) error {
	data, err := m.store.Load()
	if err != nil {
		return err
	}
	data.GameID = gameID
	if gameToken != "" {
		data.GameToken = gameToken
	}
	return m.store.Save(data)
}

func (m *TokenManager) expiresIn(ttlSeconds int64) int64 {
	return m.now().Add(time.Duration(ttlSeconds) * time.Second).Unix()
}
