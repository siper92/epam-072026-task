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
		{name: "invalid row", rowArg: "a", colArg: "0", wantErr: errs.CodeOutOfBounds},
		{name: "invalid col", rowArg: "0", colArg: "b", wantErr: errs.CodeOutOfBounds},
		{name: "empty args", rowArg: "", colArg: "", wantErr: errs.CodeOutOfBounds},
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

func TestParseCellName(t *testing.T) {
	cases := []struct {
		name    string
		cell    string
		wantRow int
		wantCol int
		wantErr errs.Code
	}{
		{name: "top left", cell: "a1", wantRow: 0, wantCol: 0},
		{name: "center", cell: "b2", wantRow: 1, wantCol: 1},
		{name: "bottom right", cell: "c3", wantRow: 2, wantCol: 2},
		{name: "upper case", cell: "B3", wantRow: 1, wantCol: 2},
		{name: "padded", cell: " a2 ", wantRow: 0, wantCol: 1},
		{name: "row out of range", cell: "d1", wantErr: errs.CodeOutOfBounds},
		{name: "col out of range", cell: "a4", wantErr: errs.CodeOutOfBounds},
		{name: "zero col", cell: "a0", wantErr: errs.CodeOutOfBounds},
		{name: "too long", cell: "a12", wantErr: errs.CodeOutOfBounds},
		{name: "numeric only", cell: "11", wantErr: errs.CodeOutOfBounds},
		{name: "empty", cell: "", wantErr: errs.CodeOutOfBounds},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			row, col, err := ParseCellName(tc.cell)
			if tc.wantErr != "" {
				if !errs.HasCode(err, tc.wantErr) {
					t.Fatalf("ParseCellName() error = %v, want code %s", err, tc.wantErr)
				}
				if IsCellName(tc.cell) {
					t.Fatalf("IsCellName(%q) = true, want false", tc.cell)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseCellName() failed: %v", err)
			}
			if row != tc.wantRow || col != tc.wantCol {
				t.Fatalf(
					"ParseCellName() = (%d, %d), want (%d, %d)",
					row, col, tc.wantRow, tc.wantCol,
				)
			}
			if !IsCellName(tc.cell) {
				t.Fatalf("IsCellName(%q) = false, want true", tc.cell)
			}
		})
	}
}
