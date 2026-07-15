package internal

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	"ticTacSolved/task/pkg/api"
	"ticTacSolved/task/pkg/errs"
)

type Middleware func(http.Handler) http.Handler

func Chain(handler http.Handler, mws ...Middleware) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		handler = mws[i](handler)
	}
	return handler
}

type ctxKey int

const playerIDKey ctxKey = iota

func PlayerID(ctx context.Context) string {
	id, _ := ctx.Value(playerIDKey).(string)
	return id
}

func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("panic on %s %s: %v", r.Method, r.URL.Path, rec)
				writeJSON(w, http.StatusInternalServerError, api.ErrorResponse{
					Code:    codeInternal,
					Message: "internal server error",
				})
			}
		}()
		next.ServeHTTP(w, r)
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		start := time.Now()
		next.ServeHTTP(rec, r)
		log.Printf(
			"%s %s %d %s",
			r.Method, r.URL.Path, rec.status, time.Since(start),
		)
	})
}

func RequireSession(sessions SessionValidator) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, err := bearerToken(r)
			if err != nil {
				writeErr(w, err)
				return
			}
			playerID, err := sessions.ValidateSession(r.Context(), token)
			if err != nil {
				writeErr(w, err)
				return
			}
			ctx := context.WithValue(r.Context(), playerIDKey, playerID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func bearerToken(r *http.Request) (string, error) {
	header := r.Header.Get(api.HeaderAuthorization)
	token, found := strings.CutPrefix(header, api.BearerPrefix)
	if !found || token == "" {
		return "", errs.New(errs.CodeInvalidToken, "missing bearer token")
	}
	return token, nil
}
