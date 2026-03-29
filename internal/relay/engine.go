package relay

import (
	"context"
	"log"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	oteltrace "go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// Engine coordinates the movement of events from Storage to Publisher.
type Engine struct {
	storage   Storage
	publisher Publisher
	interval  time.Duration
	logger    *zap.Logger

	eventCounter metric.Int64Counter
	errorCounter metric.Int64Counter
	latencyGauge metric.Float64Histogram
	tracer       oteltrace.Tracer
}

// NewEngine creates a ready-to-run Relay Engine.
func NewEngine(s Storage, p Publisher, interval time.Duration, logger *zap.Logger, meter metric.Meter, tracer oteltrace.Tracer) *Engine {

	// 1. Initialize the Counter
	eventCounter, _ := meter.Int64Counter("outbox_events_published_total",
		metric.WithDescription("Total number of events successfully published"))

	// 2. Initialize the Error Counter
	errorCounter, _ := meter.Int64Counter("outbox_errors_total",
		metric.WithDescription("Total number of publication failures"))

	// 3. Initialize the Latency Histogram
	latencyGauge, _ := meter.Float64Histogram("outbox_batch_processing_duration_seconds",
		metric.WithDescription("Time taken to process a single batch"))

	return &Engine{
		storage:      s,
		publisher:    p,
		interval:     interval,
		logger:       logger.With(zap.String("module", "engine")),
		eventCounter: eventCounter,
		errorCounter: errorCounter,
		latencyGauge: latencyGauge,
		tracer:       tracer,
	}
}

// Run starts the polling loop. It stops when the context is cancelled.
func (e *Engine) Run(ctx context.Context) error {
	ticker := time.NewTicker(e.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := e.process(ctx); err != nil {
				log.Printf("Process error: %v", err)
			}
		}
	}
}

func (e *Engine) process(ctx context.Context) error {
	start := time.Now()
	ctx, span := e.tracer.Start(ctx, "Engine.ProcessBatch")
	defer span.End()
	// 1. Fetch a batch of events (we'll start with 10)
	events, err := e.storage.Fetch(ctx, 10)
	if err != nil {
		span.RecordError(err)
		e.logger.Error("failed to fetch events", zap.Error(err))
		return err // Could not connect to DB
	}

	for _, event := range events {

		_, childSpan := e.tracer.Start(ctx, "Publisher.Publish",
			oteltrace.WithAttributes(
				attribute.String("event_id", event.ID.String()),
				attribute.String("topic", event.Topic),
			))
		// 2. Publish the event
		if err := e.publisher.Publish(ctx, event); err != nil {
			childSpan.RecordError(err)
			childSpan.SetStatus(codes.Error, "publish failed")
			childSpan.End() // End child
			e.logger.Warn("publish failed",
				zap.String("event_id", event.ID.String()),
				zap.String("topic", event.Topic),
				zap.Error(err),
			)
			e.errorCounter.Add(ctx, 1)
			// Instead of just 'continue', we tell the DB it failed
			_ = e.storage.MarkFailed(ctx, event.ID.String(), err.Error())
			continue
		}

		e.eventCounter.Add(ctx, 1)
		e.logger.Info("event published",
			zap.String("event_id", event.ID.String()),
			zap.Duration("elapsed", time.Since(event.CreatedAt)),
		)

		// 3. Mark as successfully processed
		if err := e.storage.MarkDone(ctx, event.ID.String()); err != nil {
			e.logger.Warn("mark as done failed",
				zap.String("event_id", event.ID.String()),
				zap.String("topic", event.Topic),
				zap.Error(err),
			)
		}
		childSpan.SetStatus(codes.Ok, "success")
		childSpan.End() // End child
		e.latencyGauge.Record(ctx, time.Since(start).Seconds())
	}

	return nil
}
