package telemetry

import (
	"go.opentelemetry.io/otel/metric"
)

const (
	instrumentationName = "github.com/open-outbox/relay/engine"
)

type Metrics struct {
	// BatchSize tracks the number of events fetched in a single claim operation
	BatchSize metric.Int64Histogram
	// EventsTotal tracks throughput with labels: status (success/failed), type (event_type)
	EventsTotal metric.Int64Counter
	// EndToEndLatency tracks time from event.CreatedAt to successful delivery
	EndToEndLatency metric.Float64Histogram
	// StorageLatency tracks DB ops with label: op (claim, mark_delivered, mark_failed)
	StorageLatency metric.Float64Histogram
	// PublisherLatency tracks Broker ops with label: provider (kafka, nats, redis)
	PublisherLatency metric.Float64Histogram
	// Number of active pending events
	PendingGauge metric.Int64Gauge
	// The oldest event pending gauge
	OldestPendingSeconds metric.Int64Gauge
}

func NewMetrics(meterProvider metric.MeterProvider) (*Metrics, error) {
	meter := meterProvider.Meter(instrumentationName)
	m := &Metrics{}
	var err error

	m.BatchSize, err = meter.Int64Histogram(
		"openoutbox.events.batch_size",
		metric.WithDescription("Number of events claimed from the database in a single batch."),
		metric.WithUnit("{event}"),
	)
	if err != nil {
		return nil, err
	}

	m.EventsTotal, err = meter.Int64Counter(
		"openoutbox.events.total",
		metric.WithDescription("Total number of events processed by the relay."),
		metric.WithUnit("{event}"),
	)
	if err != nil {
		return nil, err
	}

	m.EndToEndLatency, err = meter.Float64Histogram(
		"openoutbox.events.e2e_latency",
		metric.WithDescription("Time from event creation in DB to successful publication (seconds)."),
		metric.WithUnit("s"),
	)
	if err != nil {
		return nil, err
	}

	m.StorageLatency, err = meter.Float64Histogram(
		"openoutbox.storage.latency",
		metric.WithDescription("Latency of database operations (seconds)."),
		metric.WithUnit("s"),
	)
	if err != nil {
		return nil, err
	}

	m.PublisherLatency, err = meter.Float64Histogram(
		"openoutbox.publisher.latency",
		metric.WithDescription("Latency of message broker publication (seconds)."),
		metric.WithUnit("s"),
	)
	if err != nil {
		return nil, err
	}

	m.PendingGauge, err = meter.Int64Gauge(
		"openoutbox.backlog.pending_count",
		metric.WithDescription("Current count of pending events in the outbox table."),
		metric.WithUnit("{event}"),
	)
	if err != nil {
		return nil, err
	}

	m.OldestPendingSeconds, err = meter.Int64Gauge(
		"openoutbox.backlog.oldest_age_seconds",
		metric.WithDescription("Age of the oldest pending event in the outbox table."),
		metric.WithUnit("s"),
	)
	if err != nil {
		return nil, err
	}

	return m, nil
}
