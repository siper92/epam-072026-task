package utils

import (
	"strconv"
	"strings"
	"ticTacSolved/task/pkg/errs"
)

func ParseCell(rowArg string, colArg string) (int, int, error) {
	row, err := strconv.Atoi(rowArg)
	if err != nil {
		return 0, 0, errs.Newf(errs.CodeInvalidInput, "invalid row %q", rowArg)
	}
	col, err := strconv.Atoi(colArg)
	if err != nil {
		return 0, 0, errs.Newf(errs.CodeInvalidInput, "invalid col %q", colArg)
	}
	return row, col, nil
}

func ParseCellName(cell string) (int, int, error) {
	name := strings.ToLower(strings.TrimSpace(cell))
	if len(name) != 2 {
		return 0, 0, errs.Newf(errs.CodeOutOfBounds, "cell %q is outside a1..c3", cell)
	}
	row := int(name[0]) - 'a'
	col := int(name[1]) - '1'
	if row < 0 || row > 2 || col < 0 || col > 2 {
		return 0, 0, errs.Newf(errs.CodeOutOfBounds, "cell %q is outside a1..c3", cell)
	}
	return row, col, nil
}

func IsCellName(cell string) bool {
	_, _, err := ParseCellName(cell)
	return err == nil
}
