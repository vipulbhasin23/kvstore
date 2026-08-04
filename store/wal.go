package store

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

// WAL is an append-only, fsync'd write-ahead log used to persist SET and
// DELETE operations before they're applied to the in-memory store.
type WAL struct {
	file *os.File
}

// Op identifies the kind of operation a WAL entry represents.
type Op int

const (
	// OpSet represents a SET operation.
	OpSet Op = iota
	// OpDelete represents a DELETE operation.
	OpDelete
)

// Entry represents a single parsed WAL operation.
// Value is unused for OpDelete.
type Entry struct {
	Op    Op
	Key   string
	Value string
}

// ReplayResult is the outcome of a WAL replay: the entries successfully
// parsed, and whether the final line was truncated and discarded.
type ReplayResult struct {
	Entries   []Entry
	Truncated bool
}

// NewWAL opens (creating if necessary) the WAL file at path for both
// appending and replay, using a single read/write handle for the WAL's
// lifetime.
func NewWAL(path string) (*WAL, error) {
	file, err := os.OpenFile(path, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}

	return &WAL{file: file}, nil
}

// Replay reads the WAL from the start and reconstructs the sequence of
// entries it contains. If the file ends without a fully written final
// line, that trailing line is discarded and ReplayResult.Truncated is
// set to true; corruption elsewhere in the file is treated as a fatal
// error.
func (w *WAL) Replay() (ReplayResult, error) {
	if _, err := w.file.Seek(0, io.SeekStart); err != nil {
		return ReplayResult{}, err
	}

	reader := bufio.NewReader(w.file)

	var entries []Entry
	truncated := false
	lineNum := 0

	for {
		rawLine, err := reader.ReadString('\n')
		lineNum++
		if err == nil {
			// well-formed, terminated line - parse it, treat parse failure as fatal
			entry, err := parseLine(rawLine)
			if err != nil {
				return ReplayResult{}, fmt.Errorf("wal: line %d: %w", lineNum, err)
			}
			entries = append(entries, entry)
		} else if errors.Is(err, io.EOF) {
			if len(rawLine) > 0 {
				// partial trailing data, no newline - truncated tail, discard, set Truncated
				truncated = true
			}
			break
		} else {
			return ReplayResult{}, err // genuine I/O error
		}
	}
	return ReplayResult{Entries: entries, Truncated: truncated}, nil
}

// parseLine parses a single WAL line into an Entry.
func parseLine(line string) (Entry, error) {
	line = strings.TrimSuffix(line, "\n")
	fields := strings.SplitN(line, " ", 2)
	var entry Entry
	switch fields[0] {
	case "SET":
		args := strings.SplitN(fields[1], " ", 2)
		if len(args) != 2 {
			return Entry{}, errors.New("no value present for SET")
		}
		entry = Entry{
			Op:    OpSet,
			Key:   args[0],
			Value: args[1],
		}
	case "DELETE":
		if len(fields) != 2 {
			return Entry{}, errors.New("no key present for DELETE")
		}
		entry = Entry{
			Op:  OpDelete,
			Key: fields[1],
		}
	default:
		return Entry{}, fmt.Errorf("unrecognized command %q", fields[0])
	}
	return entry, nil
}

// AppendSet writes a SET operation for key/value to the WAL and syncs it
// to disk before returning.
func (w *WAL) AppendSet(key, value string) error {
	line := fmt.Sprintf("SET %s %s\n", key, value)
	return w.writeAndSync(line)
}

// AppendDelete writes a DELETE operation for key to the WAL and syncs it
// to disk before returning.
func (w *WAL) AppendDelete(key string) error {
	line := fmt.Sprintf("DELETE %s\n", key)
	return w.writeAndSync(line)
}

func (w *WAL) writeAndSync(line string) error {
	if _, err := w.file.WriteString(line); err != nil {
		return err
	}
	return w.file.Sync()
}

// Close closes the underlying WAL file.
func (w *WAL) Close() error {
	return w.file.Close()
}
