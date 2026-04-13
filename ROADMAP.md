# ROADMAP

## Phase 1: Core Service

[X] Dependency Injection (uber-go/dig): Refactor main.go to use a container. This allows us to "swap" a NATS publisher for a Kafka one just by changing the provider function.

[X] Structured Logging (uber-go/zap): Move away from log.Printf. We need JSON logging with fields (e.g., {"level":"info", "event_id":"...", "module":"engine"}).

[X] Dockerization: Create a multi-stage Dockerfile and a production-ready docker-compose.yml.

[X] Kafka: Main publisher.

[X] Nats: Second publisher

[X] OpenTelemetry (OTEL): Add tracing support so we can see an event move from the DB through the Relay into the Broker in a single trace.

## Phase 2: The Multi-Store / Multi-Publisher Ecosystem

[ ] Enterprise configurability: fully enable the remote providers for Consul and Etcd.

[ ] MySQL: Implement the relay.Storage interface for MySQL/MariaDB.

[ ] RabbitMQ: For traditional enterprise messaging.

[ ] Redis: Using Redis Streams or Pub/Sub.

[ ] HTTP/Webhooks: To call external APIs directly from the outbox.

## Phase 3: Message Ordering

[ ] Message ordering: Add message ordering support and partitioned relay
