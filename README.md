# GoRedis

A Redis-compatible in-memory data store built from scratch in Go. Implements the RESP (Redis Serialization Protocol) wire protocol, meaning any standard Redis client can connect to it directly.

Built as a learning project to understand how Redis works under the hood — from TCP connections and protocol parsing to concurrent data access and expiration algorithms.

---

## Features

- **RESP protocol** — parses and encodes the full Redis wire protocol
- **Concurrent connections** — each client handled in its own goroutine
- **RWMutex locking** — read/write separation for safe concurrent access
- **Dual encoding** — values stored as `StringEncoding` or `IntEncoding` internally, matching Redis object encoding
- **TTL support** — per-key expiration with millisecond precision
- **Lazy expiration** — expired keys are evicted on access
- **Active expiration engine** — background cleanup runs 10 times/sec, modelled after Redis 6's expiration algorithm
- **Min-heap tracking** — keys expiring within 30 seconds are tracked in a min-heap for fast eviction

---

## Supported Commands

| Command | Syntax | Description |
|---|---|---|
| `PING` | `PING [message]` | Returns PONG or echoes the message |
| `ECHO` | `ECHO message` | Returns the message |
| `SET` | `SET key value [EX seconds]` | Set a key with optional TTL |
| `GET` | `GET key` | Get the value of a key |
| `DEL` | `DEL key [key ...]` | Delete one or more keys |
| `INCR` | `INCR key` | Increment an integer value atomically |
| `TTL` | `TTL key` | Get remaining time-to-live in seconds |

---

## Getting Started

**Requirements:** Go 1.21+

```bash
git clone https://github.com/blvckbill/redis-from-scratch
cd redis-from-scratch
go run ./cmd/server
```

Server starts on port `6369`. Connect with any Redis client:

```bash
redis-cli -p 6369 PING
redis-cli -p 6369 SET name "GoRedis"
redis-cli -p 6369 GET name
redis-cli -p 6369 SET counter 0 EX 60
redis-cli -p 6369 INCR counter
redis-cli -p 6369 TTL counter
```

---

## Project Structure

```
.
├── cmd/
│   └── server/         # Entrypoint
├── internal/
│   ├── protocol/       # RESP parser and encoder
│   │   └── resp.go
│   ├── store/          # In-memory data store
│   │   └── store.go
│   └── server/         # TCP server and command handlers
│       ├── server.go
│       └── handlers.go
```

---

## How It Works

### RESP Protocol

Every Redis client communicates using RESP. A command like `SET name GoRedis` is sent over the wire as:

```
*3\r\n$3\r\nSET\r\n$4\r\nname\r\n$7\r\nGoRedis\r\n
```

The parser reads from the TCP stream byte by byte, reconstructs the command, and hands it to the router. The encoder converts the result back into RESP before writing to the connection.

### Expiration

Expiration is handled two ways:

**Lazy** — on every `GET`, `TTL`, `INCR`, and `DEL`, the key's expiry is checked. If it's past, the key is deleted right then and the caller gets a miss.

**Active** — a background goroutine runs every 100ms and does two things each cycle:

1. Sweeps the min-heap for keys that were flagged as expiring soon and deletes any that have now passed
2. Samples the expiration dictionary randomly and deletes expired keys it finds

If more than 25% of sampled keys are expired, it loops immediately instead of waiting for the next tick. This is the same adaptive behaviour Redis uses to handle high expiry load without burning CPU unnecessarily. Each cycle is capped at 25ms.

### Concurrency

All store operations acquire the appropriate lock before touching shared state:

- `GET` → `RLock` (multiple readers allowed)
- `SET`, `DEL`, `INCR` → `Lock` (exclusive write)
- `INCR` holds the lock for the full read-modify-write cycle, making it atomic

---

## What's Next

- [ ] AOF persistence
- [ ] RDB snapshots
- [ ] `EXISTS`, `KEYS`, `DBSIZE` commands
- [ ] `PX` option for SET (millisecond TTL)
- [ ] Pub/Sub
- [ ] Benchmark suite

---

## Motivation

Most Redis deep-dives stop at "here's how a hash map works." This project goes further — implementing the actual wire protocol, the dual-dictionary expiration model, and the active cleanup algorithm described in the Redis 6 internals documentation.
