// Package cli provides the command-line interface for managing the Open Outbox Relay.
//
// This package implements administrative and operational tools using the Cobra
// library. It leverages dependency injection via dig to provide access to core
// services like Storage and logging, enabling tasks such as database pruning,
// status monitoring, and manual event management.
package cli
