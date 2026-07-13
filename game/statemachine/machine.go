package statemachine

import (
	"epam/task/pkg/errs"
)

type State string

type Event string

const (
	StateTurnX State = "TURN_X"
	StateTurnO State = "TURN_O"
	StateWonX  State = "WON_X"
	StateWonO  State = "WON_O"
	StateDraw  State = "DRAW"
)

const (
	EventMoveX    Event = "MOVE_X"
	EventMoveO    Event = "MOVE_O"
	EventWinX     Event = "WIN_X"
	EventWinO     Event = "WIN_O"
	EventDraw     Event = "DRAW"
	EventForfeitX Event = "FORFEIT_X"
	EventForfeitO Event = "FORFEIT_O"
)

var transitions = map[State]map[Event]State{
	StateTurnX: {
		EventMoveX:    StateTurnO,
		EventWinX:     StateWonX,
		EventDraw:     StateDraw,
		EventForfeitX: StateWonO,
		EventForfeitO: StateWonX,
	},
	StateTurnO: {
		EventMoveO:    StateTurnX,
		EventWinO:     StateWonO,
		EventDraw:     StateDraw,
		EventForfeitX: StateWonO,
		EventForfeitO: StateWonX,
	},
	StateWonX: {},
	StateWonO: {},
	StateDraw: {},
}

func Initial() State {
	return StateTurnX
}

func IsValid(state State) bool {
	_, ok := transitions[state]
	return ok
}

func IsTerminal(state State) bool {
	outgoing, ok := transitions[state]
	return ok && len(outgoing) == 0
}

func Next(from State, event Event) (State, error) {
	outgoing, ok := transitions[from]
	if !ok {
		return from, errs.Newf(errs.CodeInvalidInput, "unknown state %q", from)
	}
	next, ok := outgoing[event]
	if !ok {
		return from, errs.Newf(errs.CodeInvalidTransition, "event %q not allowed in state %q", event, from)
	}
	return next, nil
}

type Machine struct {
	current State
}

func NewGameMachine(initial State) (*Machine, error) {
	if !IsValid(initial) {
		return nil, errs.Newf(errs.CodeInvalidInput, "unknown state %q", initial)
	}
	return &Machine{current: initial}, nil
}

func (m *Machine) Current() State {
	return m.current
}

func (m *Machine) CanFire(event Event) bool {
	_, ok := transitions[m.current][event]
	return ok
}

func (m *Machine) Fire(event Event) (State, error) {
	next, err := Next(m.current, event)
	if err != nil {
		return m.current, err
	}
	m.current = next
	return next, nil
}
