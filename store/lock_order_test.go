package store

import (
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"
)

// Pauses writer A after its WAL append, starts writer B, and checks that B
// cannot complete until A is done. That is what keeps WAL order == map order.
func TestStore_LockSpansWALAndMap(t *testing.T) {
	path := filepath.Join(t.TempDir(), "wal.log")
	s, err := NewStore(path)
	if err != nil {
		t.Fatal(err)
	}

	reached := make(chan struct{})
	release := make(chan struct{})
	var first atomic.Bool
	afterWAL = func() {
		if first.CompareAndSwap(false, true) {
			close(reached)
			<-release
		}
	}
	t.Cleanup(func() { afterWAL = nil })

	aDone := make(chan struct{})
	go func() {
		defer close(aDone)
		if err := s.Set("k", "A"); err != nil {
			t.Error(err)
		}
	}()
	<-reached // A has written to the WAL and is paused

	bDone := make(chan struct{})
	go func() {
		defer close(bDone)
		if err := s.Set("k", "B"); err != nil {
			t.Error(err)
		}
	}()

	blocked := true
	select {
	case <-bDone:
		blocked = false
	case <-time.After(100 * time.Millisecond):
	}

	close(release)
	<-aDone
	<-bDone

	if !blocked {
		t.Error("B completed while A was paused between WAL append and map update: lock does not span both")
	}

	live, _ := s.Get("k")
	s.Close()
	s2, err := NewStore(path)
	if err != nil {
		t.Fatal(err)
	}
	defer s2.Close()
	replayed, _ := s2.Get("k")
	if live != replayed {
		t.Errorf("live=%q replayed=%q", live, replayed)
	}
}
