# kvstore

A key-value store built from scratch in Go - get/set/delete, durability via a
write-ahead log, and crash recovery.

## Status

- [x] Project scaffolding (go.mod, main.go)
- [] Basic get/set/delete (in-memory)
- [] Write-Ahead Log (WAL) for durability
- [] Crash recovery from WAL
- [] Concurrency support
- [] Tests

## Running

\'\'\'bash
git clone https://github.com/vipulbhasin23/kvstore.git
cd kvstore
go run main.go
\'\'\'
