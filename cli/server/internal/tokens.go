package internal

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"strings"
	"time"

	"ticTacSolved/task/game/auth"
	"ticTacSolved/task/game/data"
	"ticTacSolved/task/game/data/gen"
	"ticTacSolved/task/pkg/api"
	"ticTacSolved/task/pkg/errs"
)

const (
	claimScope   = "scope"
	scopeSession = "session"
	scopeRefresh = "refresh"

	sessionIDPrefix = "session:"

	defaultSessionTTL = 15 * time.Minute
	maxSessionTTL     = time.Hour
	defaultRefreshTTL = 24 * time.Hour
	maxRefreshTTL     = 7 * 24 * time.Hour
)

type LoginResult struct {
	PlayerID string
	Session  api.Token
	Refresh  api.Token
}

type Tokens interface {
	Login(
		ctx context.Context,
		user string,
		password string,
		sessionTTL int64,
		refreshTTL int64,
	) (LoginResult, error)
	Refresh(
		ctx context.Context,
		refreshToken string,
		sessionTTL int64,
	) (api.Token, error)
	ValidateSession(ctx context.Context, token string) (string, error)
}

type tokenService struct {
	sessions auth.Service
	players  data.PlayerStore
	now      func() time.Time
}

var _ Tokens = (*tokenService)(nil)

func NewTokens(sessions auth.Service, players data.PlayerStore) Tokens {
	return &tokenService{
		sessions: sessions,
		players:  players,
		now:      time.Now,
	}
}

func (t *tokenService) Login(
	ctx context.Context,
	user string,
	password string,
	sessionTTL int64,
	refreshTTL int64,
) (LoginResult, error) {
	if user == "" || password == "" {
		return LoginResult{}, errs.New(
			errs.CodeInvalidInput,
			"user and password are required",
		)
	}

	playerID := derivePlayerID(user, password)
	if err := t.ensurePlayer(ctx, playerID, user); err != nil {
		return LoginResult{}, err
	}

	session, err := t.issueSession(ctx, playerID, user, sessionTTL)
	if err != nil {
		return LoginResult{}, err
	}
	refresh, err := t.issueRefresh(playerID, user, refreshTTL)
	if err != nil {
		return LoginResult{}, err
	}

	return LoginResult{
		PlayerID: playerID,
		Session:  session,
		Refresh:  refresh,
	}, nil
}

func (t *tokenService) Refresh(
	ctx context.Context,
	refreshToken string,
	sessionTTL int64,
) (api.Token, error) {
	claims, err := auth.ValidateToken(refreshToken, t.now())
	if err != nil {
		return api.Token{}, err
	}
	if claims[claimScope] != scopeRefresh {
		return api.Token{}, errs.New(errs.CodeInvalidToken, "not a refresh token")
	}

	playerID := claims[auth.ClaimPlayerID]
	if playerID == "" {
		return api.Token{}, errs.New(errs.CodeInvalidToken, "missing player claim")
	}

	return t.issueSession(ctx, playerID, claims[gen.ClaimPlayerName], sessionTTL)
}

func (t *tokenService) ValidateSession(
	ctx context.Context,
	token string,
) (string, error) {
	claims, err := t.sessions.Validate(ctx, token)
	if err != nil {
		return "", err
	}
	if claims[claimScope] != scopeSession {
		return "", errs.New(errs.CodeInvalidToken, "not a session token")
	}

	playerID, found := strings.CutPrefix(
		claims[auth.ClaimPlayerID],
		sessionIDPrefix,
	)
	if !found || playerID == "" {
		return "", errs.New(errs.CodeInvalidToken, "missing session claims")
	}

	return playerID, nil
}

func (t *tokenService) issueSession(
	ctx context.Context,
	playerID string,
	name string,
	ttlSeconds int64,
) (api.Token, error) {
	claims := mapClaims{
		auth.ClaimPlayerID:  sessionIDPrefix + playerID,
		gen.ClaimPlayerName: name,
		claimScope:          scopeSession,
	}
	ttl := clampTTL(ttlSeconds, defaultSessionTTL, maxSessionTTL)
	token, err := t.sessions.Issue(ctx, claims, ttl)
	if err != nil {
		return api.Token{}, err
	}
	return toAPIToken(token)
}

func (t *tokenService) issueRefresh(
	playerID string,
	name string,
	ttlSeconds int64,
) (api.Token, error) {
	ttl := clampTTL(ttlSeconds, defaultRefreshTTL, maxRefreshTTL)
	expiresAt := t.now().Add(ttl).Unix()
	token, err := auth.MapToToken(map[string]string{
		auth.ClaimPlayerID:  playerID,
		gen.ClaimPlayerName: name,
		claimScope:          scopeRefresh,
		auth.ClaimExpiresAt: strconv.FormatInt(expiresAt, 10),
	})
	if err != nil {
		return api.Token{}, err
	}
	return api.Token{Value: token, ExpiresAt: expiresAt}, nil
}

func (t *tokenService) ensurePlayer(
	ctx context.Context,
	playerID string,
	name string,
) error {
	_, err := t.players.GetPlayer(ctx, playerID)
	if err == nil {
		return nil
	}
	if !errs.HasCode(err, errs.CodeNotFound) {
		return err
	}
	return t.players.CreatePlayer(ctx, gen.Player{ID: playerID, Name: name})
}

func derivePlayerID(user string, password string) string {
	sum := sha256.Sum256([]byte(user + ":" + password))
	return hex.EncodeToString(sum[:16])
}

func toAPIToken(token string) (api.Token, error) {
	exp, err := auth.GetClaim(token, auth.ClaimExpiresAt)
	if err != nil {
		return api.Token{}, err
	}
	expiresAt, err := strconv.ParseInt(exp, 10, 64)
	if err != nil {
		return api.Token{}, errs.Wrap(
			errs.CodeInvalidToken,
			"invalid expiration claim",
			err,
		)
	}
	return api.Token{Value: token, ExpiresAt: expiresAt}, nil
}

func clampTTL(
	seconds int64,
	fallback time.Duration,
	limit time.Duration,
) time.Duration {
	if seconds <= 0 {
		return fallback
	}
	ttl := time.Duration(seconds) * time.Second
	if ttl > limit {
		return limit
	}
	return ttl
}

type mapClaims map[string]string

func (c mapClaims) TokenData() map[string]string { return c }
