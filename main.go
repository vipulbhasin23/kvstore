package main

import (
	"errors"
	"fmt"

	"github.com/vipulbhasin23/kvstore/store"
)

func main() {
	s := store.NewStore()

	s.Set("foo", "bar")
	s.Set("hello", "world")

	if v, err := s.Get("foo"); !errors.Is(err, store.ErrKeyNotFound) {
		fmt.Println("foo =", v)
	} else {
		fmt.Println("Error retrieving foo:", err)
	}

	s.Delete("foo")
	if _, err := s.Get("foo"); errors.Is(err, store.ErrKeyNotFound) {
		fmt.Println("Successfully deleted foo")
	}

}
