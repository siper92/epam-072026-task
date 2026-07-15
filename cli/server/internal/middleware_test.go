package internal

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"ticTacSolved/task/cli/server/internal/handlers"
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
	gin.SetMode(gin.TestMode)

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
			router := gin.New()
			router.Use(ErrorHandler())
			router.GET("/", RequireSession(tc.validator), func(c *gin.Context) {
				gotPlayer = handlers.PlayerID(c)
			})

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tc.header != "" {
				req.Header.Set(api.HeaderAuthorization, tc.header)
			}
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			if rec.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d", rec.Code, tc.wantStatus)
			}
			if gotPlayer != tc.wantPlayer {
				t.Fatalf("player id = %q, want %q", gotPlayer, tc.wantPlayer)
			}
		})
	}
}

func TestErrorHandlerWritesAttachedError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(ErrorHandler())
	router.GET("/", func(c *gin.Context) {
		_ = c.Error(errs.New(errs.CodeInvalidInput, "bad input"))
	})

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
	var resp api.ErrorResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}
	if resp.Code != "INVALID_INPUT" {
		t.Fatalf("code = %q, want INVALID_INPUT", resp.Code)
	}
}

func TestErrorHandlerTurnsPanicIntoInternalError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(ErrorHandler())
	router.GET("/", func(c *gin.Context) {
		panic("boom")
	})

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

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
	gin.SetMode(gin.TestMode)

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	if got := handlers.PlayerID(c); got != "" {
		t.Fatalf("PlayerID() = %q, want empty", got)
	}
}
