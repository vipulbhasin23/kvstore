package main

import (
	"errors"
	"fmt"
	"log"

	"github.com/vipulbhasin23/kvstore/store"
)

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	s, err := store.NewStore("store.wal")
	must(err)
	defer func() {
		if err := s.Close(); err != nil {
			log.Println("error closing store:", err)
		}
	}()

	must(s.Set("foo", "bar"))
	must(s.Set("hello", "world"))

	if v, err := s.Get("foo"); !errors.Is(err, store.ErrKeyNotFound) {
		fmt.Println("foo =", v)
	} else {
		fmt.Println("Error retrieving foo:", err)
	}

	must(s.Delete("foo"))
	if _, err := s.Get("foo"); errors.Is(err, store.ErrKeyNotFound) {
		fmt.Println("Successfully deleted foo")
	}

}
