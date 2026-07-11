package errs

import (
	"errors"
	"fmt"
	"testing"
)

func TestErrorMessage(t *testing.T) {
	cases := []struct {
		name    string
		code    Code
		message string
		want    string
	}{
		{"game not found", CodeGameNotFound, "game missing", "[GAME_NOT_FOUND] game missing"},
		{"out of turn", CodeOutOfTurn, "wait your turn", "[" + string(CodeOutOfTurn) + "] wait your turn"},
		{"empty message", CodeInvalidInput, "", "[" + string(CodeInvalidInput) + "] "},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := New(tc.code, tc.message)
			if err.Error() != tc.want {
				t.Errorf("expected %q, got %q", tc.want, err.Error())
			}
		})
	}
}

func TestNewf(t *testing.T) {
	err := Newf(CodeOutOfBounds, "cell (%d,%d)", 4, 5)
	if err.Message != "cell (4,5)" {
		t.Errorf("unexpected message %q", err.Message)
	}
	if CodeOf(err) != CodeOutOfBounds {
		t.Errorf("expected code %q, got %q", CodeOutOfBounds, CodeOf(err))
	}
	plain := Newf(CodeInvalidInput, "no args")
	if plain.Message != "no args" {
		t.Errorf("unexpected message %q", plain.Message)
	}
	mixed := Newf(CodeUnknownPlayer, "player %q in game %s", "eve", "g42")
	if mixed.Message != `player "eve" in game g42` {
		t.Errorf("unexpected message %q", mixed.Message)
	}
}

func TestWrapAndUnwrap(t *testing.T) {
	cause := errors.New("disk full")
	err := Wrap(CodeStorageFailure, "save failed", cause)
	if !errors.Is(err, cause) {
		t.Error("expected errors.Is to find the cause")
	}
	if CodeOf(err) != CodeStorageFailure {
		t.Errorf("expected code %q, got %q", CodeStorageFailure, CodeOf(err))
	}
	wrapped := fmt.Errorf("outer: %w", err)
	if CodeOf(wrapped) != CodeStorageFailure {
		t.Errorf("expected CodeOf to see through wrapping, got %q", CodeOf(wrapped))
	}
	deeplyWrapped := fmt.Errorf("outermost: %w", fmt.Errorf("outer: %w", err))
	if CodeOf(deeplyWrapped) != CodeStorageFailure {
		t.Errorf("expected CodeOf to see through nested wrapping, got %q", CodeOf(deeplyWrapped))
	}
}

func TestWrapCodedCause(t *testing.T) {
	inner := New(CodeGameNotFound, "game missing")
	outer := Wrap(CodeStorageFailure, "load failed", inner)
	if CodeOf(outer) != CodeStorageFailure {
		t.Errorf("expected outer code %q, got %q", CodeStorageFailure, CodeOf(outer))
	}
	if !errors.Is(outer, inner) {
		t.Error("expected errors.Is to find the coded cause")
	}
}

func TestCodeOfNonCodedError(t *testing.T) {
	if CodeOf(errors.New("plain")) != "" {
		t.Error("expected empty code for non-coded error")
	}
	if CodeOf(nil) != "" {
		t.Error("expected empty code for nil error")
	}
	if CodeOf(fmt.Errorf("outer: %w", errors.New("inner"))) != "" {
		t.Error("expected empty code for wrapped non-coded error")
	}
}

func TestHasCode(t *testing.T) {
	err := New(CodeOutOfTurn, "wait")
	if !HasCode(err, CodeOutOfTurn) {
		t.Error("expected HasCode to match")
	}
	if HasCode(err, CodeCellOccupied) {
		t.Error("did not expect HasCode to match a different code")
	}
	if HasCode(nil, CodeOutOfTurn) {
		t.Error("did not expect HasCode to match nil error")
	}
	if HasCode(errors.New("plain"), CodeOutOfTurn) {
		t.Error("did not expect HasCode to match non-coded error")
	}
	wrapped := fmt.Errorf("outer: %w", err)
	if !HasCode(wrapped, CodeOutOfTurn) {
		t.Error("expected HasCode to see through wrapping")
	}
}
