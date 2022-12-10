package store

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"sync"
)

// Store is a thread-safe implementation of url shorten service.
type Store struct {
	Version   string          `json:"version"`
	Generator YukiS0Generator `json:"generator"`

	mu           sync.RWMutex
	Table        map[string]string `json:"table"`
	reverseTable map[string]string
}

// NewStore creates a Store on given `filePath`.
// It returns an error if `filePath` is not writable.
func NewStore(filePath string) (*Store, error) {
	store := Store{
		Version: YukiS0Version,
		Generator: YukiS0Generator{
			Prime:      yukis0Prime,
			MaxEntries: yukiS0MaxEntries,
			Base:       rand.Uint64(),
		},
		Table:        make(map[string]string),
		reverseTable: make(map[string]string),
	}
	if err := store.DumpToDisk(filePath); err != nil {
		return nil, err
	}
	return &store, nil
}

// OpenOrCreate creates / loads an object on given `filePath`.
func OpenOrCreate(filePath string) (*Store, error) {
	fp, err := os.Open(filePath)
	if os.IsNotExist(err) {
		return NewStore(filePath)
	}
	if err != nil {
		return nil, err
	}
	defer fp.Close()
	store := Store{}
	if err := json.NewDecoder(fp).Decode(&store); err != nil {
		return nil, err
	}
	if store.Version != YukiS0Version {
		return nil, fmt.Errorf("unsupported generator version: %s", store.Version)
	}
	store.reverseTable = make(map[string]string)
	for key, val := range store.Table {
		store.reverseTable[val] = key
	}
	return &store, nil
}

// DumpToDisk dumps the current state into the disk.
func (store *Store) DumpToDisk(filePath string) error {
	store.mu.RLock()
	defer store.mu.RUnlock()

	data, err := json.Marshal(store)
	if err != nil {
		return err
	}
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return err
	}
	return nil
}

// AddLink generates a short word for given `url`.
func (store *Store) AddLink(url string) (string, error) {
	store.mu.Lock()
	defer store.mu.Unlock()

	if shorten, ok := store.reverseTable[url]; ok {
		return shorten, nil
	}

	shorten, err := store.Generator.Generate(uint64(len(store.Table)))
	if err != nil {
		return "", err
	}
	if _, ok := store.Table[shorten]; ok {
		return "", fmt.Errorf("URL collision")
	}
	store.Table[shorten] = url
	store.reverseTable[url] = shorten
	return shorten, nil
}

// Query returns the corresponding url from the given word.
// Returns an error if the word does not exist.
func (store *Store) Query(word string) (string, error) {
	store.mu.RLock()
	url, exist := store.Table[word]
	store.mu.RUnlock()
	if !exist {
		return "", fmt.Errorf("url does not exist")
	}
	return url, nil
}
