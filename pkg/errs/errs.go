package errs

import (
	"errors"
	"fmt"
)

type Code string

const (
	CodeInvalidInput      Code = "INVALID_INPUT"
	CodeGameNotFound      Code = "GAME_NOT_FOUND"
	CodeUnknownPlayer     Code = "UNKNOWN_PLAYER"
	CodeOutOfTurn         Code = "OUT_OF_TURN"
	CodeOutOfBounds       Code = "OUT_OF_BOUNDS"
	CodeCellOccupied      Code = "CELL_OCCUPIED"
	CodeGameFinished      Code = "GAME_FINISHED"
	CodeInvalidAction     Code = "INVALID_ACTION"
	CodeInvalidTransition Code = "INVALID_TRANSITION"
	CodeStorageFailure    Code = "STORAGE_FAILURE"
)

type Error struct {
	Code    Code
	Message string
	Err     error
}

func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func (e *Error) Unwrap() error {
	return e.Err
}

func New(code Code, message string) *Error {
	return &Error{Code: code, Message: message}
}

func Newf(code Code, format string, args ...any) *Error {
	return &Error{Code: code, Message: fmt.Sprintf(format, args...)}
}

func Wrap(code Code, message string, err error) *Error {
	return &Error{Code: code, Message: message, Err: err}
}

func CodeOf(err error) Code {
	var e *Error
	if errors.As(err, &e) {
		return e.Code
	}
	return ""
}

func HasCode(err error, code Code) bool {
	return CodeOf(err) == code
}
