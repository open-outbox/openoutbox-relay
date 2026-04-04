package publishers

import (
	"context"
	"errors"
	"time"

	"github.com/open-outbox/relay/internal/relay"
	"github.com/open-outbox/relay/internal/telemetry"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type InstrumentedPublisher struct {
	publisher relay.Publisher
	logger    *zap.Logger
	metrics   *telemetry.Metrics
	tracer    trace.Tracer
	meter     metric.Meter
}

func NewInstrumentedPublisher(p relay.Publisher, tel telemetry.Telemetry) *InstrumentedPublisher {
	return &InstrumentedPublisher{
		publisher: p,
		logger:    tel.ScopedLogger("publisher"),
		metrics:   tel.Metrics,
		tracer:    tel.Tracer,
		meter:     tel.Meter,
	}
}

func (ip *InstrumentedPublisher) Publish(ctx context.Context, event relay.Event) error {
	ctx, span := ip.tracer.Start(ctx, "Publisher.Publish",
		trace.WithAttributes(
			attribute.String("event.id", event.ID.String()),
			attribute.String("event.type", event.Type),
			attribute.Int("event.attempt", event.Attempts),
		))
	defer span.End()

	start := time.Now()
	err := ip.publisher.Publish(ctx, event)

	// Move metric recording here so we can include the outcome
	status := "success"
	if err != nil {
		status = "failed"

		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		if !errors.Is(err, context.Canceled) {
			ip.logger.Warn("publish failed",
				zap.String("event_id", event.ID.String()),
				zap.String("type", event.Type),
				zap.Error(err),
			)
		}
	} else {
		span.SetStatus(codes.Ok, "success")
	}

	// Record latency with BOTH type and status
	ip.metrics.PublisherLatency.Record(ctx, time.Since(start).Seconds(),
		metric.WithAttributes(
			attribute.String("type", event.Type),
			attribute.String("status", status),
		))

	return err
}

func (ip *InstrumentedPublisher) Close() error {
	return ip.publisher.Close()
}
