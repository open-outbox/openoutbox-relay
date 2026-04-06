# OpenOutbox Relay

**OpenOutbox Relay** is the official **reference implementation** of the [OpenOutbox Specification](https://github.com/open-outbox/openoutbox-spec).

It is a high-performance, reliable implementation of the **Transactional Outbox Pattern**. It ensures `"at-least-once"` delivery of events from your database to message brokers like **Kafka** or **NATS**, solving the distributed transaction problem without complex 2PC protocols.

---

## ✨ Features

* **At-Least-Once Delivery:** Guaranteed event propagation even if the broker or the relay itself crashes.
* **Provider Agnostic:** Built-in support for **Apache Kafka** and **NATS JetStream**.
* **Minimal configuration** Only specify your storage and broker and you're ready to run it.
* **Self-Healing:** Automatically identifies and recovers "stuck" events from crashed workers via a configurable lease expiration policy.
* **Observability:** Integrated with **OpenTelemetry** for distributed tracing and performance monitoring.
* **Smart Retries:** Configurable backoff and retry logic for resilient message publishing

## 📈 Scalability & Performance

The Relay is built for high-throughput environments where performance is critical:

* **High-Concurrency Locking:** Uses `SELECT ... FOR UPDATE SKIP LOCKED` to allow multiple Relay instances to work on the same database table simultaneously without contention or double-processing.
* **Horizontal Scalability:** You can scale the Relay horizontally by simply spinning up more containers; the locking mechanism ensures they naturally load-balance the event workload.
* **Optimized Batching:** Processes events in configurable batches (default 100) to minimize database round-trips and maximize Kafka/NATS throughput.
* **Low Latency:** Written in Go with a focus on non-blocking I/O and minimal memory overhead, ensuring sub-millisecond overhead between the database and the broker.

---

## 🚀 Quick Start

### 1. Database Schema

The Relay expects an `outbox_events` table following the [OpenOutbox Standard](https://github.com/open-outbox/openoutbox-spec). You can find the DDLs for this table in
[schema/postgres/openoutbox.sql](./schema/postgres/openoutbox.sql)

### 2. Run with Docker

```bash
docker run -d \
  --name openoutbox-relay \
  -e STORAGE_TYPE="postgres" \
  -e STORAGE_URL="STORAGE_URL=postgres://postgres:postgres@localhost:5432/postgres" \
  -e PUBLISHER_TYPE="kafka" \
  -e KAFKA_BROKERS="kafka://localhost:9092" \
  openoutbox/relay:latest
```

Or with `docker-compose`:

```yaml
version: '3'

services:
  relay:
    image: openoutbox/relay:latest
    container_name: openoutbox-relay
    environment:
      - STORAGE_TYPE=postgres
      - STORAGE_URL=postgres://postgres:postgres@db:5432/postgres
      - PUBLISHER_TYPE=kafka
      - KAFKA_BROKERS=kafka:9092
    restart: always
```

Apologies for the oversight. Here is the full configuration section in raw Markdown format, ready to be pasted into your README.md.

Markdown
## ⚙️ Configuration

The OpenOutbox Relay is configured entirely via environment variables.

### 1. Core Infrastructure & Connectivity
| Variable | Description | Options / Example | Default |
| :--- | :--- | :--- | :--- |
| `STORAGE_TYPE` | Database engine type | `postgres`, `mysql`,  | `postgres` |
| `PUBLISHER_TYPE` | Target message broker | `nats`, `kafka`, `redis`, `stdout`, `null` | `stdout` |
| `STORAGE_URL` | DB Connection string | `postgres://user:pass@host:5432/db` | — |
| `PUBLISHER_URL` | Broker address | `localhost:9092` or `nats://localhost:4222` | `localhost:9092` |
| `RELAY_ID` | Unique ID for this instance | `relay-worker-01` | `os.Hostname()` |
| `ENVIRONMENT` | Execution mode | `production`, `development` | `production` |

### 2. Engine Tuning (Outbox Polling)

| Variable | Description | Default |
| :--- | :--- | :--- |
| `POLL_INTERVAL` | Frequency of DB polling for new events | `500ms` |
| `BATCH_SIZE` | Max events to process in a single iteration | `100` |
| `LEASE_TIMEOUT` | Duration before a "DELIVERING" event is considered stuck | `3m` |
| `REAP_BATCH_SIZE` | Number of expired leases to reset per cleanup cycle | `100` |
| `SERVER_PORT` | HTTP Server for health checks and metrics | `:9000` |

### 3. Reliability & Backoff
| Variable | Description | Default |
| :--- | :--- | :--- |
| `RETRY_MAX_ATTEMPTS` | Max attempts before an event is marked `DEAD` | `25` |
| `RETRY_BASE_DELAY` | Starting delay for exponential backoff | `1s` |
| `RETRY_MAX_DELAY` | Maximum delay cap for backoff | `24h` |
| `RETRY_JITTER` | Randomness factor (0.15 = 15% jitter) | `0.15` |

### 4. Kafka Specific Tuning
| Variable | Description | Default |
| :--- | :--- | :--- |
| `KAFKA_BATCH_SIZE` | **Critical:** Set to `1` to bypass client-side double-batching | `1` |
| `KAFKA_BATCH_BYTES` | Maximum batch size in bytes | `10485760` |
| `KAFKA_COMPRESSION` | `none`, `gzip`, `snappy`, `lz4`, `zstd` | `snappy` |
| `KAFKA_REQUIRED_ACKS` | `all`, `one`, `none` | `all` |
| `KAFKA_ASYNC` | Set `false` for "At-Least-Once" delivery | `false` |
| `KAFKA_WRITE_TIMEOUT` | Timeout for writing to the broker | `10s` |

### 5. Observability (OpenTelemetry)
| Variable | Description | Example |
| :--- | :--- | :--- |
| `OTEL_TRACES_EXPORTER` | Destination for trace data | `otlp`, `console`, `none` |
| `OTEL_METRICS_EXPORTER` | Destination for metric data | `otlp`, `prometheus`, `none` |
| `OTEL_EXPORTER_OTLP_ENDPOINT`| OTLP Collector address | `http://localhost:4317` |
| `OTEL_EXPORTER_OTLP_PROTOCOL`| Transport protocol | `grpc`, `http/protobuf` |
| `OTEL_BSP_SCHEDULE_DELAY` | Interval between span exports | `5000ms` |

### 6. Local Development & Testing
| Variable | Description | Default |
| :--- | :--- | :--- |
| `LOCAL_TEST_TOPIC` | Base name for publisher subjects/topics | `outbox.events.v1` |
| `LOCAL_NATS_STREAM` | JetStream stream name for NATS mode | `OUTBOX_EVENTS` |
| `LOCAL_OTEL_TEST_TRACE_COUNT`| Traces to simulate during `make test-otel` | `100` |
| `LOCAL_PRODUCER_BATCH_SIZE` | Records per batch inserted by test producer | `10000` |
| `LOCAL_PRODUCER_INTERVAL` | Interval between test batch insertions | `1s` |

## 📚 Documentation & Community

Stay connected and help us improve the OpenOutbox ecosystem:

* **[Contribution Guide](./CONTRIBUTING.md)**: Ready to help? Check out our guide on setting up your local environment, running tests with the `Makefile`, and our pull request process.
* **[Changelog](./CHANGELOG.md)**: View a detailed list of changes, improvements, and bug fixes for every release.
* **[OpenOutbox Specification](https://github.com/open-outbox/openoutbox-spec)**: Learn more about the standard this relay implements.
* **[Architecture Deep Dive](./docs/architecture.md)**: A technical look at the internal locking mechanisms and event lifecycle management.

---

## License

[MIT License](./LICENSE) — Created by the OpenOutbox Team.
