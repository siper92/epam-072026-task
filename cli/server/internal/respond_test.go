package internal

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"ticTacSolved/task/pkg/api"
	"ticTacSolved/task/pkg/errs"
)

func TestWriteErr(t *testing.T) {
	cases := []struct {
		name       string
		err        error
		wantStatus int
		wantCode   string
		wantMsg    string
	}{
		{
			name:       "invalid input maps to bad request",
			err:        errs.New(errs.CodeInvalidInput, "bad input"),
			wantStatus: http.StatusBadRequest,
			wantCode:   "INVALID_INPUT",
			wantMsg:    "bad input",
		},
		{
			name:       "out of bounds maps to bad request",
			err:        errs.New(errs.CodeOutOfBounds, "cell outside board"),
			wantStatus: http.StatusBadRequest,
			wantCode:   "OUT_OF_BOUNDS",
			wantMsg:    "cell outside board",
		},
		{
			name:       "invalid token maps to unauthorized",
			err:        errs.New(errs.CodeInvalidToken, "bad token"),
			wantStatus: http.StatusUnauthorized,
			wantCode:   "INVALID_TOKEN",
			wantMsg:    "bad token",
		},
		{
			name:       "not found maps to not found",
			err:        errs.New(errs.CodeNotFound, "missing"),
			wantStatus: http.StatusNotFound,
			wantCode:   "NOT_FOUND",
			wantMsg:    "missing",
		},
		{
			name:       "game not found maps to not found",
			err:        errs.New(errs.CodeGameNotFound, "no such game"),
			wantStatus: http.StatusNotFound,
			wantCode:   "GAME_NOT_FOUND",
			wantMsg:    "no such game",
		},
		{
			name:       "cell occupied maps to conflict",
			err:        errs.New(errs.CodeCellOccupied, "cell taken"),
			wantStatus: http.StatusConflict,
			wantCode:   "CELL_OCCUPIED",
			wantMsg:    "cell taken",
		},
		{
			name:       "invalid transition maps to conflict",
			err:        errs.New(errs.CodeInvalidTransition, "not your turn"),
			wantStatus: http.StatusConflict,
			wantCode:   "INVALID_TRANSITION",
			wantMsg:    "not your turn",
		},
		{
			name:       "game finished maps to conflict",
			err:        errs.New(errs.CodeGameFinished, "game over"),
			wantStatus: http.StatusConflict,
			wantCode:   "GAME_FINISHED",
			wantMsg:    "game over",
		},
		{
			name:       "wrapped error hides the cause",
			err:        errs.Wrap(errs.CodeInvalidInput, "bad body", errors.New("boom")),
			wantStatus: http.StatusBadRequest,
			wantCode:   "INVALID_INPUT",
			wantMsg:    "bad body",
		},
		{
			name:       "unmapped code keeps the code with internal status",
			err:        errs.New(errs.CodeStorageFailure, "db down"),
			wantStatus: http.StatusInternalServerError,
			wantCode:   "STORAGE_FAILURE",
			wantMsg:    "db down",
		},
		{
			name:       "plain error maps to internal",
			err:        errors.New("boom"),
			wantStatus: http.StatusInternalServerError,
			wantCode:   codeInternal,
			wantMsg:    "boom",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			writeErr(rec, tc.err)

			if rec.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d", rec.Code, tc.wantStatus)
			}
			var resp api.ErrorResponse
			if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
				t.Fatalf("failed to decode error response: %v", err)
			}
			if resp.Code != tc.wantCode {
				t.Fatalf("code = %q, want %q", resp.Code, tc.wantCode)
			}
			if resp.Message != tc.wantMsg {
				t.Fatalf("message = %q, want %q", resp.Message, tc.wantMsg)
			}
		})
	}
}
