# Development Roadmap: Distributed Notification System

**Strategy:** Start with local Go primitives (channels + goroutines), prove the logic works, then swap out each component for real distributed infrastructure. This avoids "infrastructure paralysis."

---

## Phase 1: The Local Core (In-Memory Foundation)

**Goal:** Build the Gin API and background processing using only native Go. No external dependencies.

### Step 1.1 — Setup the Gin Ingestion Endpoint

- Create `POST /api/v1/events` endpoint.
- Define event payload: `EventID`, `EventType`, `UserID`, `Timestamp`, `Payload json.RawMessage`.
- Parse and validate incoming JSON.

### Step 1.2 — The In-Memory Queue

- Initialize a buffered Go channel: `var EventQueue = make(chan Event, 100)`.
- On valid event, push to channel. Return `202 Accepted` immediately.
- No blocking the client.

### Step 1.3 — Native Goroutine Workers

- Spin up 2–3 background goroutines with `for event := range EventQueue`.
- Each worker logs `"Worker processed event: {EventType}"` to prove async execution.

**Files to create/modify:**
- `internal/handler/event.go` — Gin handler
- `internal/model/event.go` — Event struct
- `internal/queue/queue.go` — Channel-based queue
- `internal/worker/worker.go` — Goroutine workers
- `cmd/server/main.go` — Wire everything

**Dependencies:** None (pure Go)

---

## Phase 2: Decoupling with a Message Broker

**Goal:** Replace the Go channel with RabbitMQ so the API scales independently from workers.

### Step 2.1 — Infrastructure Spin-up

- Create `docker-compose.yml` with RabbitMQ service.
- `make infra-up` / `make infra-down` for lifecycle.

### Step 2.2 — Refactor the Ingestion API (Publisher)

- Remove the Go channel.
- Integrate `github.com/rabbitmq/amqp091-go`.
- Gin handler publishes event JSON to a RabbitMQ exchange.

### Step 2.3 — Extract the Workers (Consumers)

- Move worker loop into a separate `cmd/worker/main.go`.
- Worker connects to RabbitMQ and consumes from the queue.
- API and worker are now independent processes.

**Files to create/modify:**
- `docker-compose.yml`
- `cmd/worker/main.go` — new consumer entry point
- `internal/broker/rabbitmq.go` — RabbitMQ client
- `internal/handler/event.go` — swap channel for broker
- `Makefile` — infra commands

**Dependencies:** `amqp091-go`, Docker

---

## Phase 3: Implementing the Worker Services

**Goal:** Build real business logic inside the worker.

### Step 3.1 — The Notification Worker

- Consumer routine filters for critical events (e.g., `"order_placed"`).
- Mock notification helper: formatted console output simulating email/SMS.
- `time.Sleep(500ms)` to mimic network latency.

### Step 3.2 — The Analytics Aggregator Worker

- Second consumer routine intercepts *all* events.
- Add Redis to `docker-compose.yml`.
- Use `go-redis` to update real-time metrics:
  - `PFADD` (HyperLogLog) for daily active users.
  - `INCRBY` for feature usage counts.

**Files to create/modify:**
- `internal/worker/notifier.go` — notification logic
- `internal/worker/analytics.go` — Redis aggregator
- `internal/store/redis.go` — Redis client
- `docker-compose.yml` — add Redis

**Dependencies:** `go-redis`, Redis

---

## Phase 4: Persistence & System Protection

**Goal:** Long-term storage + rate limiting.

### Step 4.1 — The Event Store

- Add PostgreSQL (or MongoDB) to `docker-compose.yml`.
- Create an "Archiver" worker routine that stores every raw event into the DB.
- Permanent, un-aggregated history for audit/analytics.

### Step 4.2 — Token-Bucket Rate Limiting

- Gin middleware using Redis for token-bucket algorithm.
- Check user IP/ID for available tokens.
- Return `429 Too Many Requests` when exhausted.

**Files to create/modify:**
- `internal/store/postgres.go` — DB client
- `internal/worker/archiver.go` — event persistence
- `internal/middleware/ratelimit.go` — rate limiter
- `internal/migrations/` — schema files
- `docker-compose.yml` — add PostgreSQL

**Dependencies:** `pgx` (or `mongo-driver`), Redis

---

## Phase 5: Scaling & Chaos Testing

**Goal:** Prove the distributed architecture works under stress.

### Step 5.1 — Horizontal Scaling Test

- Run 3 worker instances in separate terminals.
- Flood API with 1,000 rapid events.
- Verify RabbitMQ distributes load evenly across all instances.

### Step 5.2 — The Crash Test

- Flood API again. Midway, `Ctrl+C` one worker.
- Verify zero event loss — remaining workers pick up the slack.

---

## Dependency Graph

```
 Phase 1          Phase 2          Phase 3          Phase 4          Phase 5
┌─────────┐     ┌──────────┐     ┌──────────┐     ┌──────────┐     ┌─────────┐
│  Local  │ ──► │ RabbitMQ │ ──► │  Redis   │ ──► │PostgreSQL│ ──► │  Chaos  │
│ Channel │     │  Broker  │     │Analytics │     │ Archive  │     │  Tests  │
└─────────┘     └──────────┘     └──────────┘     └──────────┘     └─────────┘
     │               │               │               │
     ▼               ▼               ▼               ▼
  Gin API        Decoupled        Workers         Rate Limiter
                 Publisher/       with Real
                 Consumer         Business
                                  Logic
```

## Quick Reference

| Phase | What You Get | External Deps |
|-------|-------------|---------------|
| 1 | Working async API with in-memory queue | None |
| 2 | Decoupled API + Worker via RabbitMQ | Docker, RabbitMQ |
| 3 | Notifications + Real-time Analytics | Docker, Redis |
| 4 | Event Archive + Rate Limiting | Docker, PostgreSQL |
| 5 | Proof that it scales horizontally | — |
