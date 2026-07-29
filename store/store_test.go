package store

import (
	"errors"
	"testing"
)

func TestSetAndGet(t *testing.T) {
	s := NewStore()
	s.Set("foo", "bar")

	v, err := s.Get("foo")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if v != "bar" {
		t.Errorf("expected bar, got: %s", v)
	}
}

func TestMissingKey(t *testing.T) {
	s := NewStore()
	_, err := s.Get("missing")

	if !errors.Is(err, ErrKeyNotFound) {
		t.Fatalf("expected KeyNotFound, got: %v", err)
	}
}

func TestDelete(t *testing.T) {
	s := NewStore()

	s.Set("foo", "bar")
	s.Delete("foo")

	if _, err := s.Get("foo"); !errors.Is(err, ErrKeyNotFound) {
		t.Fatalf("expected ErrKeyNotFound after delete, got: %v", err)
	}
}
