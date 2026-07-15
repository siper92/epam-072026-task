package internal

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"ticTacSolved/task/pkg/api"
	"ticTacSolved/task/pkg/errs"
)

type stubValidator struct {
	playerID string
	err      error
}

var _ SessionValidator = (*stubValidator)(nil)

func (s *stubValidator) ValidateSession(context.Context, string) (string, error) {
	return s.playerID, s.err
}

func TestRequireSession(t *testing.T) {
	cases := []struct {
		name       string
		header     string
		validator  *stubValidator
		wantStatus int
		wantPlayer string
	}{
		{
			name:       "missing header",
			validator:  &stubValidator{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "not a bearer header",
			header:     "Basic abc",
			validator:  &stubValidator{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "empty bearer token",
			header:     strings.TrimSpace(api.BearerPrefix),
			validator:  &stubValidator{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:   "rejected token",
			header: api.BearerPrefix + "bad",
			validator: &stubValidator{
				err: errs.New(errs.CodeInvalidToken, "bad token"),
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "valid token exposes the player id",
			header:     api.BearerPrefix + "good",
			validator:  &stubValidator{playerID: "p1"},
			wantStatus: http.StatusOK,
			wantPlayer: "p1",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gotPlayer := ""
			next := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
				gotPlayer = PlayerID(r.Context())
			})
			handler := RequireSession(tc.validator)(next)

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tc.header != "" {
				req.Header.Set(api.HeaderAuthorization, tc.header)
			}
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d", rec.Code, tc.wantStatus)
			}
			if gotPlayer != tc.wantPlayer {
				t.Fatalf("player id = %q, want %q", gotPlayer, tc.wantPlayer)
			}
		})
	}
}

func TestChainAppliesMiddlewareInOrder(t *testing.T) {
	var order []string
	mw := func(name string) Middleware {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				order = append(order, name)
				next.ServeHTTP(w, r)
			})
		}
	}
	handler := Chain(
		http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
			order = append(order, "handler")
		}),
		mw("first"),
		mw("second"),
	)

	handler.ServeHTTP(
		httptest.NewRecorder(),
		httptest.NewRequest(http.MethodGet, "/", nil),
	)

	want := "first,second,handler"
	if got := strings.Join(order, ","); got != want {
		t.Fatalf("call order = %q, want %q", got, want)
	}
}

func TestRecoverTurnsPanicIntoInternalError(t *testing.T) {
	handler := Recover(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("boom")
	}))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
	var resp api.ErrorResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}
	if resp.Code != codeInternal {
		t.Fatalf("code = %q, want %q", resp.Code, codeInternal)
	}
}

func TestPlayerIDWithoutSession(t *testing.T) {
	if got := PlayerID(context.Background()); got != "" {
		t.Fatalf("PlayerID() = %q, want empty", got)
	}
}
