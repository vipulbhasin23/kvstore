package store

import (
	"errors"
	"log"
	"sync"
)

// Store is an in-memory key-value store backed by a write-ahead log for
// durability. All writes are persisted to the WAL before being applied
// to the in-memory map.
type Store struct {
	mu   sync.RWMutex
	data map[string]string
	wal  *WAL
}

// ErrKeyNotFound is returned by Get when the requested key doesn't exist.
var ErrKeyNotFound = errors.New("key not found")

// NewStore opens (creating if necessary) the WAL at walPath, replays it
// to rebuild in-memory state, and returns a ready-to-use Store. If the
// WAL's trailing entry was truncated, that's logged but not treated as
// an error.
func NewStore(walPath string) (*Store, error) {
	wal, err := NewWAL(walPath)
	if err != nil {
		return nil, err
	}

	result, err := wal.Replay()
	if err != nil {
		_ = wal.Close() // best-effort cleanup: Replay's error is what the caller needs
		return nil, err
	}
	if result.Truncated {
		log.Printf("wal: discarded incomplete trailing entry during replay")
	}
	data := make(map[string]string)

	for _, e := range result.Entries {
		switch e.Op {
		case OpSet:
			data[e.Key] = e.Value
		case OpDelete:
			delete(data, e.Key)
		}
	}

	return &Store{
		data: data,
		wal:  wal,
	}, nil
}

// Get returns the value for key, or ErrKeyNotFound if it doesn't exist.
func (s *Store) Get(key string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	value, ok := s.data[key]

	if !ok {
		return "", ErrKeyNotFound
	}

	return value, nil
}

// Set writes key/value to the WAL and, on success, updates the
// in-memory map.
func (s *Store) Set(key, value string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.wal.AppendSet(key, value); err != nil {
		return err
	}
	s.data[key] = value
	return nil
}

// Delete removes key, first recording the deletion in the WAL and then,
// on success, removing it from the in-memory map.
func (s *Store) Delete(key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.wal.AppendDelete(key); err != nil {
		return err
	}
	delete(s.data, key)
	return nil
}

// Close closes the underlying WAL.
func (s *Store) Close() error {
	return s.wal.Close()
}
