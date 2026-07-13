package util

import (
	"crypto/rand"
	"encoding/hex"

	"ticTacSolved/task/pkg/errs"
)

const gridSize = 3

func NewID() string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		panic(err)
	}
	return hex.EncodeToString(buf)
}

func EncodeGrid(cells [gridSize][gridSize]rune) string {
	encoded := make([]rune, 0, gridSize*gridSize)
	for row := 0; row < gridSize; row++ {
		for col := 0; col < gridSize; col++ {
			encoded = append(encoded, cells[row][col])
		}
	}
	return string(encoded)
}

func DecodeGrid(encoded string) ([gridSize][gridSize]rune, error) {
	var cells [gridSize][gridSize]rune
	runes := []rune(encoded)
	if len(runes) != gridSize*gridSize {
		return cells, errs.Newf(errs.CodeInvalidInput, "encoded grid must have %d cells, got %d", gridSize*gridSize, len(runes))
	}
	for i, r := range runes {
		cells[i/gridSize][i%gridSize] = r
	}
	return cells, nil
}

func CloneMap[K comparable, V any](src map[K]V) map[K]V {
	dst := make(map[K]V, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}
