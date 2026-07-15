package utils

import (
	"testing"

	"ticTacSolved/task/pkg/errs"
)

func TestParseCell(t *testing.T) {
	cases := []struct {
		name    string
		rowArg  string
		colArg  string
		wantRow int
		wantCol int
		wantErr errs.Code
	}{
		{name: "valid cell", rowArg: "1", colArg: "2", wantRow: 1, wantCol: 2},
		{name: "zero cell", rowArg: "0", colArg: "0"},
		{name: "negative values parse", rowArg: "-1", colArg: "3", wantRow: -1, wantCol: 3},
		{name: "invalid row", rowArg: "a", colArg: "0", wantErr: errs.CodeInvalidInput},
		{name: "invalid col", rowArg: "0", colArg: "b", wantErr: errs.CodeInvalidInput},
		{name: "empty args", rowArg: "", colArg: "", wantErr: errs.CodeInvalidInput},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			row, col, err := ParseCell(tc.rowArg, tc.colArg)
			if tc.wantErr != "" {
				if !errs.HasCode(err, tc.wantErr) {
					t.Fatalf("ParseCell() error = %v, want code %s", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseCell() failed: %v", err)
			}
			if row != tc.wantRow || col != tc.wantCol {
				t.Fatalf(
					"ParseCell() = (%d, %d), want (%d, %d)",
					row, col, tc.wantRow, tc.wantCol,
				)
			}
		})
	}
}
