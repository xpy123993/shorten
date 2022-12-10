package store

import (
	"encoding/base64"
	"fmt"
)

const (
	// YukiS0Version is a reserved keyword to identiy generator version.
	YukiS0Version = "yuki-s0"

	yukis0Prime      uint64 = 26778299
	yukiS0MaxEntries uint64 = 1 << 24
)

// YukiS0Generator is a shorten link generator.
type YukiS0Generator struct {
	Prime      uint64 `json:"generator-prime"`
	MaxEntries uint64 `json:"max-entries"`
	Base       uint64 `json:"base"`
}

// Generate generates a shorten word for the given index.
// If `index` is out of supported range, an error will be returned.
// The returend string will be guaranteed to be uniquely identified an index.
func (generator *YukiS0Generator) Generate(index uint64) (string, error) {
	if index >= generator.MaxEntries {
		return "", fmt.Errorf("full capacity: expecting a number at most %v, got %v", generator.MaxEntries, index)
	}
	result := (generator.Prime * (index + generator.Base)) & (1<<24 - 1)
	return base64.RawURLEncoding.EncodeToString([]byte{
		byte(result >> 16), byte(result >> 8), byte(result),
	}), nil
}
