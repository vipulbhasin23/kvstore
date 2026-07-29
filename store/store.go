package store

import "errors"

type Store struct {
	data map[string]string
}

var ErrKeyNotFound = errors.New("key not found")

func NewStore() *Store {
	return &Store{data: make(map[string]string)}
}

func (s *Store) Get(key string) (string, error) {
	value, ok := s.data[key]

	if !ok {
		return "", ErrKeyNotFound
	}

	return value, nil
}

func (s *Store) Set(key, value string) {
	s.data[key] = value
}

func (s *Store) Delete(key string) {
	delete(s.data, key)
}
