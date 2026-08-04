package store

import (
	"errors"
	"os"
	"path/filepath"
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
