package store_test

import (
	"math/big"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/xpy123993/shorten/store"
)

func TestCoverage(t *testing.T) {
	generator := store.YukiS0Generator{
		Prime:      *big.NewInt(31),
		MaxEntries: 1 << 10,
		Base:       rand.Int63(),
	}
	table := make(map[string]int64)
	for i := int64(0); i < generator.MaxEntries; i++ {
		word, err := generator.Generate(i)
		if err != nil {
			t.Fatal(err)
		}
		if _, exist := table[word]; exist {
			t.Fatal("collision")
		}
		table[word] = i
	}
	if len(table) != int(generator.MaxEntries) {
		t.Errorf("expect full coverage %d, got %d", generator.MaxEntries, len(table))
	}
}

func TestMain(m *testing.M) {
	rand.Seed(time.Now().UnixMicro())
	os.Exit(m.Run())
}
