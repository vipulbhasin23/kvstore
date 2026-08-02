package store

import "errors"

type Store struct {
	data map[string]string
	wal  *WAL
}

var ErrKeyNotFound = errors.New("key not found")

func NewStore(walPath string) (*Store, error) {
	wal, err := NewWAL(walPath)
	if err != nil {
		return nil, err
	}
	return &Store{
		data: make(map[string]string),
		wal:  wal,
	}, nil
}

func (s *Store) Get(key string) (string, error) {
	value, ok := s.data[key]

	if !ok {
		return "", ErrKeyNotFound
	}

	return value, nil
}

func (s *Store) Set(key, value string) error {
	if err := s.wal.AppendSet(key, value); err != nil {
		return err
	}
	s.data[key] = value
	return nil
}

func (s *Store) Delete(key string) error {
	if err := s.wal.AppendDelete(key); err != nil {
		return err
	}
	delete(s.data, key)
	return nil
}

func (s *Store) Close() error {
	return s.wal.Close()
}
