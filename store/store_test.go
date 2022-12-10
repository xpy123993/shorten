package store_test

import (
	"os"
	"path"
	"testing"

	"github.com/xpy123993/shorten/store"
)

func createTestStore(t *testing.T) (*store.Store, string) {
	filePath := path.Join(t.TempDir(), "test.json")
	urlStore, err := store.OpenOrCreate(filePath)
	if err != nil {
		t.Fatalf("cannot create store: %v", err)
	}
	return urlStore, filePath
}

func TestCreateFromNonExistingFile(t *testing.T) {
	_, filePath := createTestStore(t)
	stat, err := os.Stat(filePath)
	if err != nil || stat.Size() == 0 {
		t.Fatalf("cannot dump an empty store into the disk")
	}
}

func TestDumpAndOpen(t *testing.T) {
	urlStore, filePath := createTestStore(t)

	word, err := urlStore.AddLink("http://test-link")
	if err != nil {
		t.Fatal(err)
	}
	if err := urlStore.DumpToDisk(filePath); err != nil {
		t.Fatal(err)
	}
	urlStore, err = store.OpenOrCreate(filePath)
	if err != nil {
		t.Fatal(err)
	}
	result, err := urlStore.Query(word)
	if err != nil {
		t.Fatal(err)
	}
	if result != "http://test-link" {
		t.Fatalf("expecting %v and %v to be the same", result, "http://test-link")
	}
	word2, err := urlStore.AddLink("http://test-link")
	if err != nil {
		t.Fatal(err)
	}
	if word != word2 {
		t.Errorf("expect %v and %v to be the same", word, word2)
	}
}

func TestQueryNotFound(t *testing.T) {
	urlStore, _ := createTestStore(t)
	if link, err := urlStore.Query("link"); err == nil {
		t.Fatalf("expect to return an error, got data: %s", link)
	}
}

func TestQuery(t *testing.T) {
	urlStore, _ := createTestStore(t)
	word, err := urlStore.AddLink("http://test-link")
	if err != nil {
		t.Fatal(err)
	}
	if len(word) == 0 {
		t.Fatalf("empty word returned from the store")
	}
	result, err := urlStore.Query(word)
	if err != nil {
		t.Fatal(err)
	}
	if result != "http://test-link" {
		t.Fatalf("result mismatched: %v vs %v", result, "http://test-link")
	}
}
