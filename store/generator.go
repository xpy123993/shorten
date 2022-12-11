package store

import (
	"encoding/base64"
	"fmt"
	"math/big"
)

const (
	// YukiS0Version is a reserved keyword to identiy generator version.
	YukiS0Version          = "yuki-s0"
	yukiS0MaxEntries int64 = 1 << 24
)

// YukiS0Generator is a shorten link generator.
type YukiS0Generator struct {
	Prime      big.Int `json:"generator-prime"`
	MaxEntries int64   `json:"max-entries"`
	Base       int64   `json:"base"`
}

// Generate generates a shorten word for the given index.
// If `index` is out of supported range, an error will be returned.
// The returend string will be guaranteed to be uniquely identified an index.
func (generator *YukiS0Generator) Generate(index int64) (string, error) {
	if index >= generator.MaxEntries {
		return "", fmt.Errorf("full capacity: expecting a number at most %v, got %v", generator.MaxEntries, index)
	}
	z := big.Int{}
	z.Mul(&generator.Prime, big.NewInt(generator.Base+index)).Mod(&z, big.NewInt(generator.MaxEntries))
	result := z.Int64() & (yukiS0MaxEntries - 1)
	return base64.RawURLEncoding.EncodeToString([]byte{
		byte(result >> 16), byte(result >> 8), byte(result),
	}), nil
}
