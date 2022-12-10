package store_test

import (
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/xpy123993/shorten/store"
)

func TestCoverage(t *testing.T) {
	generator := store.YukiS0Generator{
		Prime:      31,
		MaxEntries: 16,
		Base:       rand.Uint64(),
	}
	table := make(map[string]uint64)
	for i := uint64(0); i < generator.MaxEntries; i++ {
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
