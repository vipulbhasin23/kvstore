package store

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWALAppendAccumulates(t *testing.T) {
	walPath := filepath.Join(t.TempDir(), "test.wal")
	wal, err := NewWAL(walPath)
	mustT(t, err)

	mustT(t, wal.AppendSet("foo", "bar"))
	mustT(t, wal.AppendDelete("foo"))

	content, err := os.ReadFile(walPath)
	mustT(t, err)
	expected := "SET foo bar\nDELETE foo\n"
	if string(content) != expected {
		t.Errorf("expected: %q, got: %q", expected, string(content))
	}
}

func TestWALReopenAppends(t *testing.T) {
	walPath := filepath.Join(t.TempDir(), "test.wal")
	wal, err := NewWAL(walPath)
	mustT(t, err)

	mustT(t, wal.AppendSet("foo", "bar"))
	mustT(t, wal.Close())

	wal, err = NewWAL(walPath)
	mustT(t, err)
	mustT(t, wal.AppendDelete("foo"))

	content, err := os.ReadFile(walPath)
	mustT(t, err)
	expected := "SET foo bar\nDELETE foo\n"
	if string(content) != expected {
		t.Errorf("expected: %q, got: %q", expected, string(content))
	}
}
