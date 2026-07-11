package statemachine

import (
	"testing"

	"epam/task/pkg/errs"
)

func TestNewGameMachineUnknownState(t *testing.T) {
	for _, s := range []State{"LIMBO", "", "turnx and turno"} {
		_, err := NewGameMachine(s)
		if !errs.HasCode(err, errs.CodeInvalidInput) {
			t.Errorf("expected code %q for %q, got %v", errs.CodeInvalidInput, s, err)
		}
	}
}

func TestNewGameMachineValidStates(t *testing.T) {
	for _, s := range []State{StateTurnX, StateTurnO, StateWonX, StateWonO, StateDraw} {
		m, err := NewGameMachine(s)
		if err != nil {
			t.Errorf("NewGameMachine(%q) failed: %v", s, err)
			continue
		}
		if m.Current() != s {
			t.Errorf("expected current state %q, got %q", s, m.Current())
		}
	}
}

func TestValidTransitions(t *testing.T) {
	cases := []struct {
		from  State
		event Event
		to    State
	}{
		{StateTurnX, EventMoveX, StateTurnO},
		{StateTurnO, EventMoveO, StateTurnX},
		{StateTurnX, EventWinX, StateWonX},
		{StateTurnO, EventWinO, StateWonO},
		{StateTurnX, EventDraw, StateDraw},
		{StateTurnO, EventDraw, StateDraw},
		{StateTurnX, EventForfeitX, StateWonO},
		{StateTurnX, EventForfeitO, StateWonX},
		{StateTurnO, EventForfeitX, StateWonO},
		{StateTurnO, EventForfeitO, StateWonX},
	}
	for _, tc := range cases {
		t.Run(string(tc.from)+"+"+string(tc.event), func(t *testing.T) {
			m, err := NewGameMachine(tc.from)
			if err != nil {
				t.Fatalf("NewGameMachine failed: %v", err)
			}
			if !m.CanFire(tc.event) {
				t.Errorf("expected CanFire(%q) in %q", tc.event, tc.from)
			}
			next, err := m.Fire(tc.event)
			if err != nil {
				t.Fatalf("Fire failed: %v", err)
			}
			if next != tc.to || m.Current() != tc.to {
				t.Errorf("expected state %q, got %q", tc.to, next)
			}
		})
	}
}

func TestInvalidTransitions(t *testing.T) {
	type transition struct {
		from  State
		event Event
	}
	cases := []transition{
		{StateTurnX, EventMoveO},
		{StateTurnO, EventMoveX},
		{StateTurnX, EventWinO},
		{StateTurnO, EventWinX},
	}
	events := []Event{EventMoveX, EventMoveO, EventWinX, EventWinO, EventDraw, EventForfeitX, EventForfeitO}
	for _, terminal := range []State{StateWonX, StateWonO, StateDraw} {
		for _, event := range events {
			cases = append(cases, transition{terminal, event})
		}
	}
	for _, tc := range cases {
		t.Run(string(tc.from)+"+"+string(tc.event), func(t *testing.T) {
			m, err := NewGameMachine(tc.from)
			if err != nil {
				t.Fatalf("NewGameMachine failed: %v", err)
			}
			if m.CanFire(tc.event) {
				t.Errorf("did not expect CanFire(%q) in %q", tc.event, tc.from)
			}
			next, err := m.Fire(tc.event)
			if !errs.HasCode(err, errs.CodeInvalidTransition) {
				t.Errorf("expected code %q, got %v", errs.CodeInvalidTransition, err)
			}
			if next != tc.from || m.Current() != tc.from {
				t.Errorf("state must not change on invalid transition, got %q", next)
			}
		})
	}
}

func TestIsTerminal(t *testing.T) {
	for _, s := range []State{StateWonX, StateWonO, StateDraw} {
		if !IsTerminal(s) {
			t.Errorf("expected %q to be terminal", s)
		}
	}
	for _, s := range []State{StateTurnX, StateTurnO} {
		if IsTerminal(s) {
			t.Errorf("did not expect %q to be terminal", s)
		}
	}
}

func containsAll(s string, subs ...string) bool {
	for _, sub := range subs {
		if !contains(s, sub) {
			return false
		}
	}
	return true
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
