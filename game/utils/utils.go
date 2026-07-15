package utils

import (
	"strconv"
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
