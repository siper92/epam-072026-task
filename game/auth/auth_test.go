package auth

import (
	"os"
	"strings"
	"testing"

	"ticTacSolved/task/pkg/config"
	"ticTacSolved/task/pkg/errs"
)

func TestMain(m *testing.M) {
	os.Setenv("JWT_SECRET", "test-secret")
	config.LoadEnv()

	os.Exit(m.Run())
}

func TestRoundTrip(t *testing.T) {
	in := map[string]string{"sub": "player-1", "role": "black"}
	token, err := MapToToken(in)
	if err != nil {
		t.Fatalf("MapToToken: %v", err)
	}

	out, err := GetClaims(token)
	if err != nil {
		t.Fatalf("Claims: %v", err)
	}

	if len(out) != len(in) || out["sub"] != "player-1" || out["role"] != "black" {
		t.Fatalf("claims mismatch: %v", out)
	}
}

func TestTamperedTokenRejected(t *testing.T) {
	token, err := MapToToken(map[string]string{"sub": "player-1"})
	if err != nil {
		t.Fatalf("MapToToken: %v", err)
	}

	parts := strings.Split(token, ".")
	tampered := parts[0] + "." + parts[1] + "." + strings.Repeat("A", len(parts[2]))
	if _, err := GetClaims(tampered); !errs.HasCode(err, errs.CodeInvalidToken) {
		t.Fatalf("expected CodeInvalidToken, got %v", err)
	}

	if _, err := GetClaims("not-a-token"); !errs.HasCode(err, errs.CodeInvalidToken) {
		t.Fatalf("expected CodeInvalidToken, got %v", err)
	}
}

func TestGetClaim(t *testing.T) {
	token, err := MapToToken(map[string]string{"sub": "player-1"})
	if err != nil {
		t.Fatalf("MapToToken: %v", err)
	}

	value, err := GetClaim(token, "sub")
	if err != nil {
		t.Fatalf("GetClaim: %v", err)
	}

	if value != "player-1" {
		t.Fatalf("unexpected claim value: %q", value)
	}

	if _, err := GetClaim(token, "missing"); !errs.HasCode(err, errs.CodeInvalidToken) {
		t.Fatalf("expected CodeInvalidToken, got %v", err)
	}
}

type testUser struct {
	ID string
}

func (u testUser) TokenData() map[string]string {
	return map[string]string{"sub": u.ID}
}

func TestToToken(t *testing.T) {
	token, err := ToToken(testUser{ID: "player-2"})
	if err != nil {
		t.Fatalf("ToToken: %v", err)
	}

	value, err := GetClaim(token, "sub")
	if err != nil {
		t.Fatalf("GetClaim: %v", err)
	}

	if value != "player-2" {
		t.Fatalf("unexpected claim value: %q", value)
	}
}
