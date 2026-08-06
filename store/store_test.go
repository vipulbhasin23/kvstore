package store

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func mustT(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSetAndGet(t *testing.T) {
	walPath := filepath.Join(t.TempDir(), "test.wal")
	s, err := NewStore(walPath)
	mustT(t, err)

	mustT(t, s.Set("foo", "bar"))

	v, err := s.Get("foo")
	mustT(t, err)
	if v != "bar" {
		t.Errorf("get: expected bar, got: %q", v)
	}
}

func TestMissingKey(t *testing.T) {
	walPath := filepath.Join(t.TempDir(), "test.wal")
	s, err := NewStore(walPath)
	mustT(t, err)

	_, err = s.Get("missing")

	if !errors.Is(err, ErrKeyNotFound) {
		t.Fatalf("get: expected: %v, got: %v", ErrKeyNotFound, err)
	}
}

func TestDelete(t *testing.T) {
	walPath := filepath.Join(t.TempDir(), "test.wal")
	s, err := NewStore(walPath)
	mustT(t, err)

	mustT(t, s.Set("foo", "bar"))
	mustT(t, s.Delete("foo"))

	if _, err = s.Get("foo"); !errors.Is(err, ErrKeyNotFound) {
		t.Fatalf("get: expected: %v after delete, got: %v", ErrKeyNotFound, err)
	}
}

func TestStoreRestartReplaysData(t *testing.T) {
	walPath := filepath.Join(t.TempDir(), "test.wal")
	s, err := NewStore(walPath)
	mustT(t, err)

	mustT(t, s.Set("foo", "bar"))
	mustT(t, s.Set("baz", "qu"))
	mustT(t, s.Delete("baz"))
	mustT(t, s.Close())

	s, err = NewStore(walPath)
	mustT(t, err)

	val, err := s.Get("foo")
	mustT(t, err)
	if val != "bar" {
		t.Errorf("replay: expected: bar, got: %q", val)
	}

	_, err = s.Get("baz")
	if !errors.Is(err, ErrKeyNotFound) {
		t.Errorf("get: expected: %v, got: %v", ErrKeyNotFound, err)
	}
}

func TestStoreRestartWithTruncatedWAL(t *testing.T) {
	walPath := filepath.Join(t.TempDir(), "test.wal")
	mustT(t, os.WriteFile(walPath, []byte("SET foo bar\nSET baz qux"), 0644))

	s, err := NewStore(walPath)
	mustT(t, err)

	val, err := s.Get("foo")
	mustT(t, err)

	if val != "bar" {
		t.Errorf("get: expected bar, got: %v", val)
	}

	_, err = s.Get("baz")
	if !errors.Is(err, ErrKeyNotFound) {
		t.Errorf("get: expected: %v, got: %v", ErrKeyNotFound, err)
	}
}

func TestStore_ConcurrentAccess(t *testing.T) {
	walPath := filepath.Join(t.TempDir(), "test.wal")
	s, err := NewStore(walPath)
	mustT(t, err)
	defer s.Close()

	const goroutines = 50
	const opsPerGoroutine = 100

	var wg sync.WaitGroup
	wg.Add(goroutines)
	for g := 0; g < goroutines; g++ {
		go func(id int) {
			defer wg.Done()
			for i := 0; i < opsPerGoroutine; i++ {
				key := fmt.Sprintf("key-%d", i%10)
				switch i % 3 {
				case 0:
					_ = s.Set(key, fmt.Sprintf("g%d-i%d", id, i))
				case 1:
					_, _ = s.Get(key)
				case 2:
					_ = s.Delete(key)
				}
			}
		}(g)
	}
	wg.Wait()
}

func TestStore_ConcurrentWritesReplayConsistent(t *testing.T) {
	goroutines := 50
	iterations := 10

	tmpDir := t.TempDir()

	for i := 0; i < iterations; i++ {
		walPath := filepath.Join(tmpDir, fmt.Sprintf("test-%d.wal", i))
		s, err := NewStore(walPath)
		mustT(t, err)

		var wg sync.WaitGroup
		wg.Add(goroutines)
		for g := 0; g < goroutines; g++ {
			go func(id int) {
				defer wg.Done()
				if err := s.Set("key1", fmt.Sprintf("value-%d", id)); err != nil {
					t.Errorf("Set: %v", err)
				}
			}(g)
		}
		wg.Wait()

		expected, err := s.Get("key1")
		mustT(t, err)
		mustT(t, s.Close())

		reopened, err := NewStore(walPath)
		mustT(t, err)
		got, err := reopened.Get("key1")
		mustT(t, err)
		mustT(t, reopened.Close())

		if got != expected {
			t.Errorf("iteration %d: live: %v, replayed: %v", i, expected, got)
		}
	}
}
