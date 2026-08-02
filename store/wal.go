package store

import (
	"fmt"
	"os"
)

type WAL struct {
	file *os.File
}

func NewWAL(path string) (*WAL, error) {
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	return &WAL{file: file}, nil
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
