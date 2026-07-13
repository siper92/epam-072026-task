package state_machine

import (
	"strings"
	"ticTacSolved/task/pkg/errs"
)

type GameStatus string

const (
	StatusWaitingForPlayers  GameStatus = "WAITING_FOR_PLAYERS"
	StatusPlayerXTurn        GameStatus = "PlayerX_TURN"
	StatusPlayerXLeft        GameStatus = "PlayerX_LEFT"
	StatusPlayerOTurn        GameStatus = "PlayerO_TURN"
	StatusPlayerOLeft        GameStatus = "PlayerO_LEFT"
	StatusGameOverDraw       GameStatus = "GAME_OVER_DRAW"
	StatusGameOverPlayerXWin GameStatus = "GAME_OVER_PlayerX_WIN"
	StatusGameOverPlayerOWin GameStatus = "GAME_OVER_PlayerO_WIN"
)

const (
	cellEmpty   = "_"
	cellPlayerX = "X"
	cellPlayerO = "O"
)

type GameState struct {
	Board         [3][3]string
	CurrentPlayer string
	MoveCount     int
	State         GameStatus
}

type StateMachine interface {
	ProcessMove(row int, col int) error
	GetCurrentState() GameState
	PlayerLeft(player string) error
}

func parseGameState(encoded string) (GameState, error) {
	encoded = strings.Trim(strings.ToUpper(encoded), " \n\t")
	if len(encoded) != 9 {
		return GameState{}, errs.Newf(
			errs.CodeInvalidInput,
			"encoded board must be 9 characters, got %d",
			len(encoded),
		)
	}

	var state GameState
	countX, countO := 0, 0
	for i, r := range encoded {
		cell := string(r)
		switch cell {
		case cellPlayerX:
			countX++
		case cellPlayerO:
			countO++
		case cellEmpty:
		default:
			return GameState{}, errs.Newf(
				errs.CodeInvalidInput,
				"invalid cell %q in encoded board", cell,
			)
		}
		state.Board[i/3][i%3] = cell
	}

	if countX != countO && countX != countO+1 {
		return GameState{}, errs.Newf(
			errs.CodeInvalidInput,
			"impossible mark counts: %d X vs %d O", countX, countO,
		)
	}

	state.MoveCount = countX + countO
	winX := boardWinner(state.Board, cellPlayerX)
	winO := boardWinner(state.Board, cellPlayerO)
	switch {
	case winX && winO:
		return GameState{}, errs.New(errs.CodeInvalidInput, "both players cannot win")
	case winX:
		if countX == countO {
			return GameState{}, errs.New(errs.CodeInvalidInput, "X cannot win with equal mark counts")
		}
		state.State = StatusGameOverPlayerXWin
	case winO:
		if countX == countO+1 {
			return GameState{}, errs.New(errs.CodeInvalidInput, "O cannot win when X has moved last")
		}
		state.State = StatusGameOverPlayerOWin
	case boardFull(state.Board):
		state.State = StatusGameOverDraw
	case countX == countO:
		state.State = StatusPlayerXTurn
		state.CurrentPlayer = cellPlayerX
	default:
		state.State = StatusPlayerOTurn
		state.CurrentPlayer = cellPlayerO
	}

	return state, nil
}

type stateMachine struct {
	state GameState
}

var _ StateMachine = (*stateMachine)(nil)

func newStateMachine(state GameState) StateMachine {
	return &stateMachine{state: state}
}

func NewStateMachine(data string) (StateMachine, error) {
	machine, err := parseGameState(data)
	if err != nil {
		return nil, err
	}

	return newStateMachine(machine), nil
}

func (m *stateMachine) ProcessMove(row int, col int) error {
	switch m.state.State {
	case StatusWaitingForPlayers:
		return errs.New(errs.CodeInvalidTransition, "game has not started yet")
	case StatusPlayerXTurn, StatusPlayerOTurn:
	default:
		return errs.Newf(errs.CodeGameFinished, "game is over in state %q", m.state.State)
	}
	if row < 0 || row > 2 || col < 0 || col > 2 {
		return errs.Newf(errs.CodeOutOfBounds, "cell (%d,%d) is outside the 3x3 board", row, col)
	}
	if m.state.Board[row][col] != cellEmpty {
		return errs.Newf(errs.CodeCellOccupied, "cell (%d,%d) is already occupied", row, col)
	}
	mark := m.state.CurrentPlayer
	m.state.Board[row][col] = mark
	m.state.MoveCount++
	switch {
	case boardWinner(m.state.Board, mark):
		if mark == cellPlayerX {
			m.state.State = StatusGameOverPlayerXWin
		} else {
			m.state.State = StatusGameOverPlayerOWin
		}
		m.state.CurrentPlayer = ""
	case boardFull(m.state.Board):
		m.state.State = StatusGameOverDraw
		m.state.CurrentPlayer = ""
	case mark == cellPlayerX:
		m.state.State = StatusPlayerOTurn
		m.state.CurrentPlayer = cellPlayerO
	default:
		m.state.State = StatusPlayerXTurn
		m.state.CurrentPlayer = cellPlayerX
	}
	return nil
}

func (m *stateMachine) PlayerLeft(player string) error {
	if player != cellPlayerX && player != cellPlayerO {
		return errs.Newf(errs.CodeInvalidInput, "unknown player %q", player)
	}
	switch m.state.State {
	case StatusGameOverDraw, StatusGameOverPlayerXWin, StatusGameOverPlayerOWin, StatusPlayerXLeft, StatusPlayerOLeft:
		return errs.Newf(errs.CodeGameFinished, "game is over in state %q", m.state.State)
	}
	if player == cellPlayerX {
		m.state.State = StatusGameOverPlayerOWin
	} else {
		m.state.State = StatusGameOverPlayerXWin
	}
	m.state.CurrentPlayer = ""
	return nil
}

func (m *stateMachine) GetCurrentState() GameState {
	return m.state
}

func boardWinner(board [3][3]string, mark string) bool {
	for i := 0; i < 3; i++ {
		if board[i][0] == mark && board[i][1] == mark && board[i][2] == mark {
			return true
		}
		if board[0][i] == mark && board[1][i] == mark && board[2][i] == mark {
			return true
		}
	}
	if board[0][0] == mark && board[1][1] == mark && board[2][2] == mark {
		return true
	}
	return board[0][2] == mark && board[1][1] == mark && board[2][0] == mark
}

func boardFull(board [3][3]string) bool {
	for _, row := range board {
		for _, cell := range row {
			if cell == cellEmpty {
				return false
			}
		}
	}
	return true
}
