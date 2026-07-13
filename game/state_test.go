package game

import (
	"testing"

	"epam/task/pkg/errs"
)

func TestGridEncodeParseRoundtrip(t *testing.T) {
	cases := []struct {
		name    string
		rows    [GridSize]string
		encoded string
	}{
		{"empty", [GridSize]string{"___", "___", "___"}, "_________"},
		{"mixed", [GridSize]string{"XO_", "_X_", "O_X"}, "XO__X_O_X"},
		{"full", [GridSize]string{"XOX", "OOX", "XXO"}, "XOXOOXXXO"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			grid := gridFromRows(tc.rows)
			encoded := grid.Encode()
			if encoded != tc.encoded {
				t.Errorf("expected encoding %q, got %q", tc.encoded, encoded)
			}
			parsed, err := ParseGrid(encoded)
			if err != nil {
				t.Fatalf("ParseGrid failed: %v", err)
			}
			if parsed != grid {
				t.Errorf("roundtrip mismatch: %v != %v", parsed, grid)
			}
		})
	}
}

func TestParseGridInvalid(t *testing.T) {
	cases := []struct {
		name  string
		input string
	}{
		{"empty", ""},
		{"too short", "XO_"},
		{"too long", "XO__X_O_X_"},
		{"bad mark", "XO__Z_O_X"},
		{"lowercase mark", "xo__x_o_x"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := ParseGrid(tc.input); !errs.HasCode(err, errs.CodeInvalidInput) {
				t.Errorf("expected code %q, got %v", errs.CodeInvalidInput, err)
			}
		})
	}
}
