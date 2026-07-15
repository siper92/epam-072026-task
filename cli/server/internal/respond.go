package internal

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"ticTacSolved/task/pkg/api"
	"ticTacSolved/task/pkg/errs"
)

const codeInternal = "INTERNAL_ERROR"

var statusByCode = map[errs.Code]int{
	errs.CodeInvalidInput:      http.StatusBadRequest,
	errs.CodeOutOfBounds:       http.StatusBadRequest,
	errs.CodeInvalidToken:      http.StatusUnauthorized,
	errs.CodeNotFound:          http.StatusNotFound,
	errs.CodeGameNotFound:      http.StatusNotFound,
	errs.CodeCellOccupied:      http.StatusConflict,
	errs.CodeInvalidTransition: http.StatusConflict,
	errs.CodeGameFinished:      http.StatusConflict,
	errs.CodeOutOfTurn:         http.StatusConflict,
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("failed to encode response: %v", err)
	}
}

func writeErr(w http.ResponseWriter, err error) {
	code := errs.CodeOf(err)
	status, known := statusByCode[code]
	if !known {
		status = http.StatusInternalServerError
	}
	if code == "" {
		code = codeInternal
	}

	message := err.Error()
	var typed *errs.Error
	if errors.As(err, &typed) {
		message = typed.Message
	}

	writeJSON(w, status, api.ErrorResponse{
		Code:    string(code),
		Message: message,
	})
}
