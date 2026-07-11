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

type transitionKey struct {
	from  State
	event Event
}

var gameTransitions = map[transitionKey]State{
	{StateTurnX, EventMoveX}:    StateTurnO,
	{StateTurnO, EventMoveO}:    StateTurnX,
	{StateTurnX, EventWinX}:     StateWonX,
	{StateTurnO, EventWinO}:     StateWonO,
	{StateTurnX, EventDraw}:     StateDraw,
	{StateTurnO, EventDraw}:     StateDraw,
	{StateTurnX, EventForfeitX}: StateWonO,
	{StateTurnX, EventForfeitO}: StateWonX,
	{StateTurnO, EventForfeitX}: StateWonO,
	{StateTurnO, EventForfeitO}: StateWonX,
}

var validStates = map[State]bool{
	StateTurnX: true,
	StateTurnO: true,
	StateWonX:  true,
	StateWonO:  true,
	StateDraw:  true,
}

type Machine struct {
	current     State
	transitions map[transitionKey]State
}

func NewGameMachine(initial State) (*Machine, error) {
	if !validStates[initial] {
		return nil, errs.Newf(errs.CodeInvalidInput, "unknown state %q", initial)
	}
	return &Machine{current: initial, transitions: gameTransitions}, nil
}

func (m *Machine) Current() State {
	return m.current
}

func (m *Machine) CanFire(event Event) bool {
	_, ok := m.transitions[transitionKey{m.current, event}]
	return ok
}

func (m *Machine) Fire(event Event) (State, error) {
	next, ok := m.transitions[transitionKey{m.current, event}]
	if !ok {
		return m.current, errs.Newf(errs.CodeInvalidTransition, "event %q not allowed in state %q", event, m.current)
	}
	m.current = next
	return next, nil
}

func IsTerminal(state State) bool {
	return state == StateWonX || state == StateWonO || state == StateDraw
}
