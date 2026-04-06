// Package relay contains the core domain logic and orchestration for the Outbox Relay.
//
// This package serves as the heart of the system, defining the fundamental
// interfaces and data models that allow the relay to remain agnostic of
// specific database or message broker implementations.
//
// The primary component is the Engine, which coordinates the lifecycle of an
// event: claiming it from Storage, attempting delivery via a Publisher,
// and handling failures through a RetryPolicy.
//
// Core Abstractions:
//   - Engine: Orchestrates the polling loop and background maintenance tasks.
//   - Storage: Defines how events are persisted, claimed, and updated in the database.
//   - Publisher: Defines how events are delivered to external systems.
//   - Event: The central data structure carrying the payload and delivery metadata.
//
// By adhering to the Transactional Outbox pattern, this package ensures
// reliable message delivery with at-least-once guarantees.
package relay
