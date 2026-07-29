# Decisions

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
