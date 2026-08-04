package store

import (
	"os"
	"path/filepath"
	"reflect"
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
		t.Errorf("wal: expected: %q, got: %q", expected, string(content))
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
		t.Errorf("wal: expected: %q, got: %q", expected, string(content))
	}
}

func TestParseLineSet(t *testing.T) {
	entry, err := parseLine("SET foo bar\n")
	if err != nil {
		t.Errorf("parseLine: expected SET, got: %v", err)
	}
	expected := Entry{Op: OpSet, Key: "foo", Value: "bar"}
	if !reflect.DeepEqual(entry, expected) {
		t.Errorf("parseLine: expected: %+v, got: %+v", expected, entry)
	}
}

func TestParseLineDelete(t *testing.T) {
	entry, err := parseLine("DELETE foo\n")
	if err != nil {
		t.Errorf("parseLine: expected to parse DELETE, got: %v", err)
	}
	expected := Entry{Op: OpDelete, Key: "foo"}
	if !reflect.DeepEqual(entry, expected) {
		t.Errorf("parseLine: expected: %+v, got: %+v", expected, entry)
	}
}

func TestParseLineSetMissingValue(t *testing.T) {
	_, err := parseLine("SET foo\n")
	if err == nil {
		t.Errorf("parseLine: expected error on missing value for SET, got nil")
	}
}

func TestParseLineDeleteMissingKey(t *testing.T) {
	_, err := parseLine("DELETE\n")
	if err == nil {
		t.Errorf("parseLine: expected error on missing key for DELETE, got nil")
	}
}

func TestParseLineUnknownCommand(t *testing.T) {
	_, err := parseLine("FOO bar\n")
	if err == nil {
		t.Errorf("parseLine(%q): expected error, got nil", "FOO bar\n")
	}
}

func TestReplayEmptyWAL(t *testing.T) {
	walPath := filepath.Join(t.TempDir(), "test.wal")
	wal, err := NewWAL(walPath)
	mustT(t, err)

	result, err := wal.Replay()
	mustT(t, err)

	if result.Truncated {
		t.Errorf("replay: expected truncated: false, got: true")
	}
	var expected []Entry
	if !reflect.DeepEqual(expected, result.Entries) {
		t.Errorf("replay: expected: %+v, got: %+v", expected, result.Entries)
	}
}

func TestReplayAppliesEntriesInOrder(t *testing.T) {
	walPath := filepath.Join(t.TempDir(), "test.wal")
	wal, err := NewWAL(walPath)
	mustT(t, err)

	mustT(t, wal.AppendSet("foo", "bar"))
	mustT(t, wal.AppendSet("baz", "qu"))
	mustT(t, wal.AppendDelete("foo"))

	mustT(t, wal.Close())

	wal, err = NewWAL(walPath)
	mustT(t, err)

	result, err := wal.Replay()
	mustT(t, err)

	if result.Truncated {
		t.Errorf("replay: expected truncated: false, got: true")
	}

	expected := []Entry{
		{Op: OpSet, Key: "foo", Value: "bar"},
		{Op: OpSet, Key: "baz", Value: "qu"},
		{Op: OpDelete, Key: "foo"},
	}

	if !reflect.DeepEqual(expected, result.Entries) {
		t.Errorf("replay: expected: %+v, got: %+v", expected, result.Entries)
	}
}

func TestReplayTruncatedTail(t *testing.T) {
	walPath := filepath.Join(t.TempDir(), "test.wal")
	mustT(t, os.WriteFile(walPath, []byte("SET foo bar\nSET baz qu"), 0644))

	wal, err := NewWAL(walPath)
	mustT(t, err)

	result, err := wal.Replay()
	mustT(t, err)

	if !result.Truncated {
		t.Fatalf("expected truncated: true, got: false")
	}
	expected := []Entry{
		{Op: OpSet, Key: "foo", Value: "bar"},
	}

	if !reflect.DeepEqual(expected, result.Entries) {
		t.Errorf("replay: expected: %+v, got: %+v", expected, result.Entries)
	}
}

func TestReplayCleanTrailingNewLine(t *testing.T) {
	walPath := filepath.Join(t.TempDir(), "test.wal")
	mustT(t, os.WriteFile(walPath, []byte("SET foo bar\nSET baz qu\n"), 0644))

	wal, err := NewWAL(walPath)
	mustT(t, err)

	result, err := wal.Replay()
	mustT(t, err)

	if result.Truncated {
		t.Fatalf("replay: expected truncated: false, got: true")
	}

	expected := []Entry{
		{Op: OpSet, Key: "foo", Value: "bar"},
		{Op: OpSet, Key: "baz", Value: "qu"},
	}

	if !reflect.DeepEqual(expected, result.Entries) {
		t.Errorf("replay: expected: %+v, got %+v", expected, result.Entries)
	}
}
