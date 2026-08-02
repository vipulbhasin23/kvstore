# Decisions

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
