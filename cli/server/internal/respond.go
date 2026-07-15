package internal

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

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

func writeErr(c *gin.Context, err error) {
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

	c.JSON(status, api.ErrorResponse{
		Code:    string(code),
		Message: message,
	})
}
