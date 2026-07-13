package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"strings"

	"ticTacSolved/task/pkg/config"
	"ticTacSolved/task/pkg/errs"
)

const headerJSON = `{"alg":"HS256","typ":"JWT"}`

type Tokenizable interface {
	TokenData() map[string]string
}

func secret() ([]byte, error) {
	s := config.GetEnv("JWT_SECRET")
	if s == "" {
		return nil, errs.New(errs.CodeInvalidInput, "JWT_SECRET is not set")
	}
	return []byte(s), nil
}

func sign(data string, key []byte) string {
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(data))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func MapToToken(claims map[string]string) (string, error) {
	key, err := secret()
	if err != nil {
		return "", err
	}
	payload, err := json.Marshal(claims)
	if err != nil {
		return "", errs.Wrap(errs.CodeInvalidInput, "failed to encode claims", err)
	}
	header := base64.RawURLEncoding.EncodeToString([]byte(headerJSON))
	body := base64.RawURLEncoding.EncodeToString(payload)
	signingInput := header + "." + body
	return signingInput + "." + sign(signingInput, key), nil
}

func ToToken(t Tokenizable) (string, error) {
	return MapToToken(t.TokenData())
}

func GetClaims(token string) (map[string]string, error) {
	key, err := secret()
	if err != nil {
		return nil, err
	}

	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, errs.New(errs.CodeInvalidToken, "malformed token")
	}

	expected := sign(parts[0]+"."+parts[1], key)
	if !hmac.Equal([]byte(expected), []byte(parts[2])) {
		return nil, errs.New(errs.CodeInvalidToken, "invalid signature")
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, errs.Wrap(errs.CodeInvalidToken, "invalid payload encoding", err)
	}

	claims := map[string]string{}
	if err = json.Unmarshal(payload, &claims); err != nil {
		return nil, errs.Wrap(errs.CodeInvalidToken, "invalid payload", err)
	}

	return claims, nil
}

func GetClaim(token string, key string) (string, error) {
	claims, err := GetClaims(token)
	if err != nil {
		return "", err
	}

	value, ok := claims[key]
	if !ok {
		return "", errs.Newf(errs.CodeInvalidToken, "claim %q not found", key)
	}

	return value, nil
}
