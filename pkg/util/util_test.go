package util

import (
	"testing"

	"ticTacSolved/task/pkg/errs"
)

func TestNewIDUnique(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 1000; i++ {
		id := NewID()
		if len(id) != 32 {
			t.Fatalf("expected 32-char ID, got %q", id)
		}
		if seen[id] {
			t.Fatalf("duplicate ID %q", id)
		}
		seen[id] = true
	}
}

func TestGridEncodeDecodeRoundtrip(t *testing.T) {
	cases := []struct {
		name    string
		cells   [3][3]rune
		encoded string
	}{
		{
			"mixed",
			[3][3]rune{{'X', 'O', '_'}, {'_', 'X', '_'}, {'O', '_', 'X'}},
			"XO__X_O_X",
		},
		{
			"empty",
			[3][3]rune{{'_', '_', '_'}, {'_', '_', '_'}, {'_', '_', '_'}},
			"_________",
		},
		{
			"full",
			[3][3]rune{{'X', 'O', 'X'}, {'O', 'O', 'X'}, {'X', 'X', 'O'}},
			"XOXOOXXXO",
		},
		{
			"single mark",
			[3][3]rune{{'_', '_', '_'}, {'_', 'X', '_'}, {'_', '_', '_'}},
			"____X____",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			encoded := EncodeGrid(tc.cells)
			if encoded != tc.encoded {
				t.Errorf("unexpected encoding %q, want %q", encoded, tc.encoded)
			}
			decoded, err := DecodeGrid(encoded)
			if err != nil {
				t.Fatalf("DecodeGrid failed: %v", err)
			}
			if decoded != tc.cells {
				t.Errorf("roundtrip mismatch: %v != %v", decoded, tc.cells)
			}
		})
	}
}

func TestDecodeGridInvalidLength(t *testing.T) {
	cases := []struct {
		name  string
		input string
	}{
		{"empty", ""},
		{"too short", "XO"},
		{"one short", "XO__X_O_"},
		{"one long", "XO__X_O_XX"},
		{"way too long", "XO__X_O_XXO__X_O_X"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := DecodeGrid(tc.input)
			if !errs.HasCode(err, errs.CodeInvalidInput) {
				t.Errorf("expected code %q, got %v", errs.CodeInvalidInput, err)
			}
		})
	}
}

func TestCloneMap(t *testing.T) {
	src := map[string]int{"a": 1, "b": 2, "c": 3}
	dst := CloneMap(src)
	if len(dst) != len(src) {
		t.Errorf("expected same length, got %d and %d", len(dst), len(src))
	}
	for k, v := range src {
		if dst[k] != v {
			t.Errorf("expected dst[%q] = %d, got %d", k, v, dst[k])
		}
	}
	dst["a"] = 99
	dst["d"] = 4
	if src["a"] != 1 {
		t.Error("mutating clone must not affect source")
	}
	if _, ok := src["d"]; ok {
		t.Error("adding to clone must not affect source")
	}
	src["b"] = 77
	if dst["b"] != 2 {
		t.Error("mutating source must not affect clone")
	}
}

func TestCloneMapEmpty(t *testing.T) {
	dst := CloneMap(map[string]int{})
	if len(dst) != 0 {
		t.Errorf("expected empty clone, got %d entries", len(dst))
	}
	dst["a"] = 1
	if len(dst) != 1 {
		t.Error("expected clone of empty map to be writable")
	}
}
