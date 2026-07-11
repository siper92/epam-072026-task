package game

import (
	"sync"

	"epam/task/game/statemachine"
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
		Status:  statemachine.StateTurnX,
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
	machine, err := statemachine.NewGameMachine(state.Status)
	if err != nil {
		return GameState{}, err
	}

	switch action.Type {
	case ActionMove:
		if err := s.processMove(&state, machine, mark, action.Row, action.Col); err != nil {
			return GameState{}, err
		}
	case ActionForfeit:
		if err := s.processForfeit(machine, mark); err != nil {
			return GameState{}, err
		}
	default:
		return GameState{}, errs.Newf(errs.CodeInvalidAction, "unsupported action type %q", action.Type)
	}

	state.Status = machine.Current()
	switch state.Status {
	case statemachine.StateWonX:
		state.WinnerID = state.PlayerX
	case statemachine.StateWonO:
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

func (s *Service) processMove(state *GameState, machine *statemachine.Machine, mark Mark, row, col int) error {
	if err := validateMove(state, mark, row, col); err != nil {
		return err
	}
	applyMove(state, mark, row, col)
	event := moveEvent(state.Grid, mark)
	_, err := machine.Fire(event)
	return err
}

func (s *Service) processForfeit(machine *statemachine.Machine, mark Mark) error {
	event := statemachine.EventForfeitX
	if mark == MarkO {
		event = statemachine.EventForfeitO
	}
	_, err := machine.Fire(event)
	return err
}

func moveEvent(grid Grid, mark Mark) statemachine.Event {
	switch {
	case hasWin(grid, mark) && mark == MarkX:
		return statemachine.EventWinX
	case hasWin(grid, mark) && mark == MarkO:
		return statemachine.EventWinO
	case isFull(grid):
		return statemachine.EventDraw
	case mark == MarkX:
		return statemachine.EventMoveX
	default:
		return statemachine.EventMoveO
	}
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
