package store

import (
	"errors"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"
)

func lookup(t *testing.T, s *Store, key string) (string, bool) {
	t.Helper()
	v, err := s.Get(key)
	if errors.Is(err, ErrKeyNotFound) {
		return "", false
	}
	if err != nil {
		t.Fatal(err)
	}
	return v, true
}

// For each pair of operations: pause the first after its WAL append, start the
// second, and assert the second cannot complete until the first is done. That
// is what keeps WAL order == map order. Then compare live vs replayed state.
func TestStore_LockSpansWALAndMap(t *testing.T) {
	set := func(v string) func(*Store) error {
		return func(s *Store) error { return s.Set("k", v) }
	}
	del := func(s *Store) error { return s.Delete("k") }

	cases := []struct {
		name          string
		seed          bool
		first, second func(*Store) error
	}{
		{"set then set", false, set("A"), set("B")},
		{"set then delete", false, set("A"), del},
		{"delete then set", true, del, set("B")},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "wal.log")
			s, err := NewStore(path)
			if err != nil {
				t.Fatal(err)
			}
			if tc.seed {
				if err := s.Set("k", "seed"); err != nil {
					t.Fatal(err)
				}
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
				if err := tc.first(s); err != nil {
					t.Error(err)
				}
			}()
			select {
			case <-reached: // first op has written to the WAL and is paused
			case <-time.After(5 * time.Second):
				t.Fatal("first op never reached afterWAL (failed before the WAL append?)")
			}

			bDone := make(chan struct{})
			go func() {
				defer close(bDone)
				if err := tc.second(s); err != nil {
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
				t.Error("second op completed while first was paused between WAL append and map update: lock does not span both")
			}

			liveV, liveOK := lookup(t, s, "k")
			s.Close()
			s2, err := NewStore(path)
			if err != nil {
				t.Fatal(err)
			}
			defer s2.Close()
			replayV, replayOK := lookup(t, s2, "k")
			if liveV != replayV || liveOK != replayOK {
				t.Errorf("live=(%q,%v) replayed=(%q,%v)", liveV, liveOK, replayV, replayOK)
			}
		})
	}
}
