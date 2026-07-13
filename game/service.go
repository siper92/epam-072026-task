package game

import (
	"sync"

	"epam/task/pkg/errs"
	"epam/task/pkg/util"
)

type ActionType string

const (
	ActionMove    ActionType = "MOVE"
	ActionForfeit ActionType = "FORFEIT"
)

type Action struct {
	PlayerID string
	Type     ActionType
	Row      int
	Col      int
}

type GameService interface {
	NewGame(playerXID, playerOID string) (GameState, error)
	GameAction(gameID string, action Action) (GameState, error)
	GetState(gameID string) (GameState, error)
}

type Service struct {
	mu    sync.Mutex
	store Store
}

var _ GameService = (*Service)(nil)

func NewService(store Store) *Service {
	return &Service{store: store}
}

func (s *Service) NewGame(playerXID, playerOID string) (GameState, error) {
	if playerXID == "" || playerOID == "" {
		return GameState{}, errs.New(errs.CodeInvalidInput, "both player IDs are required")
	}
	if playerXID == playerOID {
		return GameState{}, errs.New(errs.CodeInvalidInput, "players must be distinct")
	}
	state := GameState{
		ID:      util.NewID(),
		PlayerX: playerXID,
		PlayerO: playerOID,
		Grid:    NewGrid(),
		Status:  state_machine.Initial(),
	}
	if err := s.save(state); err != nil {
		return GameState{}, err
	}
	return state, nil
}

func (s *Service) GameAction(gameID string, action Action) (GameState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	state, err := s.load(gameID)
	if err != nil {
		return GameState{}, err
	}
	if state.Finished() {
		return GameState{}, errs.Newf(errs.CodeGameFinished, "game %q is already finished", gameID)
	}
	mark, err := playerMark(&state, action.PlayerID)
	if err != nil {
		return GameState{}, err
	}
	event, err := applyAction(&state, mark, action)
	if err != nil {
		return GameState{}, err
	}
	next, err := state_machine.Next(state.Status, event)
	if err != nil {
		return GameState{}, err
	}
	state.Status = next
	switch state.Status {
	case state_machine.StateWonX:
		state.WinnerID = state.PlayerX
	case state_machine.StateWonO:
		state.WinnerID = state.PlayerO
	}
	if err := s.save(state); err != nil {
		return GameState{}, err
	}
	return state, nil
}

func (s *Service) GetState(gameID string) (GameState, error) {
	return s.load(gameID)
}

func applyAction(state *GameState, mark Mark, action Action) (state_machine.Event, error) {
	switch action.Type {
	case ActionMove:
		if err := validateMove(state, mark, action.Row, action.Col); err != nil {
			return "", err
		}
		applyMove(state, mark, action.Row, action.Col)
		return moveEvent(state.Grid, mark), nil
	case ActionForfeit:
		return forfeitEvent(mark), nil
	default:
		return "", errs.Newf(errs.CodeInvalidAction, "unsupported action type %q", action.Type)
	}
}

func moveEvent(grid Grid, mark Mark) state_machine.Event {
	switch {
	case hasWin(grid, mark):
		if mark == MarkO {
			return state_machine.EventWinO
		}
		return state_machine.EventWinX
	case isFull(grid):
		return state_machine.EventDraw
	case mark == MarkO:
		return state_machine.EventMoveO
	default:
		return state_machine.EventMoveX
	}
}

func forfeitEvent(mark Mark) state_machine.Event {
	if mark == MarkO {
		return state_machine.EventForfeitO
	}
	return state_machine.EventForfeitX
}

func (s *Service) load(gameID string) (GameState, error) {
	if gameID == "" {
		return GameState{}, errs.New(errs.CodeInvalidInput, "game ID is required")
	}
	state, err := s.store.Load(gameID)
	if err != nil {
		if errs.CodeOf(err) != "" {
			return GameState{}, err
		}
		return GameState{}, errs.Wrap(errs.CodeStorageFailure, "failed to load game", err)
	}
	if !state_machine.IsValid(state.Status) {
		return GameState{}, errs.Newf(errs.CodeStorageFailure, "game %q has corrupt status %q", gameID, state.Status)
	}
	return state, nil
}

func (s *Service) save(state GameState) error {
	if err := s.store.Save(state); err != nil {
		if errs.CodeOf(err) != "" {
			return err
		}
		return errs.Wrap(errs.CodeStorageFailure, "failed to save game", err)
	}
	return nil
}
