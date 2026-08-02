package store

import (
	"errors"
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
		t.Errorf("expected bar, got: %s", v)
	}
}

func TestMissingKey(t *testing.T) {
	walPath := filepath.Join(t.TempDir(), "test.wal")
	s, err := NewStore(walPath)
	mustT(t, err)

	_, err = s.Get("missing")

	if !errors.Is(err, ErrKeyNotFound) {
		t.Fatalf("expected KeyNotFound, got: %v", err)
	}
}

func TestDelete(t *testing.T) {
	walPath := filepath.Join(t.TempDir(), "test.wal")
	s, err := NewStore(walPath)
	mustT(t, err)

	mustT(t, s.Set("foo", "bar"))
	mustT(t, s.Delete("foo"))

	if _, err = s.Get("foo"); !errors.Is(err, ErrKeyNotFound) {
		t.Fatalf("expected ErrKeyNotFound after delete, got: %v", err)
	}
}
