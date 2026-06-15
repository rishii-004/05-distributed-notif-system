# 05-distributed-notif-system

A distributed notification system built in Go.

## Project Structure

```
.
├── cmd/server/          # Entry point
├── internal/
│   ├── handler/         # HTTP handlers
│   ├── service/         # Business logic
│   ├── model/           # Domain types
│   ├── repository/      # Data access interface
│   └── broker/          # Message broker interface
└── pkg/config/          # Configuration
```

## Quick Start

```bash
go run ./cmd/server
```
