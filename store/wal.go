package store

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

type WAL struct {
	file *os.File
}

type Op int

const (
	OpSet Op = iota
	OpDelete
)

type Entry struct {
	Op    Op
	Key   string
	Value string
}
type ReplayResult struct {
	Entries   []Entry
	Truncated bool
}

func NewWAL(path string) (*WAL, error) {
	file, err := os.OpenFile(path, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}

	return &WAL{file: file}, nil
}

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

func (w *WAL) AppendSet(key, value string) error {
	line := fmt.Sprintf("SET %s %s\n", key, value)
	return w.writeAndSync(line)
}

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

func (w *WAL) Close() error {
	return w.file.Close()
}
