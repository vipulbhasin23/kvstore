# Decisions

## Concurrency

1. **Single `sync.RWMutex` on `Store`, held across the WHOLE `Set`/`Delete` operation - the WAL write and the map update together - not just the map update.**
   `Get` takes `RLock`; `Set`/`Delete` take the full `Lock`. This keeps a single
   invariant simple to reason about: WAL write order and map-update order can
   never diverge, because no other goroutine can begin its own WAL write until
   the current writer has released the lock after updating the map. The cost:
   writes are fully serialized, including the fsync, so there's no write
   concurrency at all. Accepted for now, since this project's goal is
   correctness and clarity, not throughput; revisit with a real benchmark
   before relaxing it (see future issue).
2. **`Replay`, called from `NewStore` before the `Store` is returned, takes no lock.**
   Nothing else can access `data` yet at that point, so no synchronization
   is needed.
3. **`TestStore_ConcurrentWritesReplayConsistent` provides only probabilistic coverage of the WAL/map ordering invariant, not a guarantee.**
   Confirmed directly: temporarily splitting the lock (WAL write outside, map
   update inside) still passed this test consistently under natural goroutine
   scheduling, because the WAL write and the lock acquisition happen close
   together in time with little between them. The bug only reproduced
   reliably once an artificial delay was inserted between the two. This is a
   known limitation, not an oversight - a reintroduction of this bug in the
   future isn't guaranteed to be caught by this test alone.

## WAL replay and crash recovery

1. **Single `O_RDWR|O_APPEND` handle, opened once in `NewWAL`**, reused for
   both replay and future appends — avoids a double-open (extra syscalls, a
   race window between close/reopen), and relies on `O_APPEND`'s guarantee
   that every write atomically targets true EOF regardless of where reading
   left the offset.
2. **Truncation detection via `bufio.Reader.ReadString('\n')`, not `bufio.Scanner`.**
   `Scanner` normalizes away whether the final line was
   newline-terminated — exactly the signal needed to detect a crash-truncated
   trailing write, since a crash mid-write (writes fsync before the map updates)
   can only ever corrupt the _last_ line.
3. **A trailing chunk with no terminating newline is discarded unconditionally, without attempting to parse it first.**
   Given the fsync-before-write guarantee (a crash can only ever leave the last
   line incomplete), "doesn't end in `\n`" is sufficient grounds on its own —
   no need to also check whether it happens to parse.
4. **Truncation reported via `ReplayResult{Entries, Truncated bool}`, not a sentinel error.**
   It's expected, designed-for behavior — the entire point of this feature — not
   a failure. Keeps `error` meaning "something genuinely went wrong"
   unconditionally, with no sentinel-checking required by callers.
5. **Corruption beyond the trailing line is out of scope.** A non-final line
   that fails to parse is a hard, fatal error — no attempt to recover from it.
   Real detection would need a WAL format change (e.g. checksums); tracked
   separately in #9.

## Write-ahead log (WAL) for durability

1. **Log format**: Chose a plain-text, one-line-per-entry format over binary or
   JSON. This keeps the log human-readable and is simpler to implement than
   either alternative. Downside: the format can't handle spaces or line breaks
   within keys or values - an accepted tradeoff in favor of simplicity.
2. **Flush strategy**: Flush (sync to disk) after every write, rather than
   batching every N writes or N milliseconds. This is less performant - but
   guarantees no data is ever lost between writes - an acceptable tradeoff given
   this project has no throughput or scalability requirements yet.
3. **Structural choice**: `Store` holds a `*WAL` via composition, rather than
   defining an interface with `WALStore`/`Store` as two implementations.
   Composition is the established Go idiom, and there's no current need for a
   `Store` variant without a WAL - introducing an interface now would be
   premature abstraction. Can revisit if a real second use case appears.
4. **`Close()`**: Added to complete the API — `NewWAL` opens a file, so a
   matching close was needed rather than leaving the handle open indefinitely.
   Follows the standard `io.Closer()` convention(`Close() error`).

## Basic get/set/delete (in-memory)

- Values are `map[string]string` for now - simplest starting point,
  may revisit for `[]byte` once WAL/serialization matters.

- `Store` lives in its own `store/` package rather than in `main.go`, since
  WAL/concurrency logic will need to live alongside this state.

- Pointer receivers (`*Store`) throughout, for consistency and to avoid
  copy issues once a `sync.Mutex` field is added later.

- `Get` returns `(string, error)` with `ErrKeyNotFound`, not comma-ok —
  anticipating other real failure modes (WAL/recovery errors) that will
  need to share the same channel later.

- Not thread-safe yet — deliberate, concurrency is a later roadmap item.
