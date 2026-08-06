# kvstore

A key-value store built from scratch in Go - get/set/delete, durability via a
write-ahead log, and crash recovery.

## Status

- [x] Project scaffolding (go.mod, main.go)
- [x] Basic get/set/delete (in-memory)
- [x] Write-Ahead Log (WAL) for durability
- [x] Crash recovery from WAL
- [x] Concurrency support
- [x] Tests

## Running

```bash
git clone https://github.com/vipulbhasin23/kvstore.git
cd kvstore
go run main.go
```

## Design notes

See [DECISIONS.md](./DECISIONS.md) for tradeoffs and reasoning as they come up.
