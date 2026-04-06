package telemetry

import (
	"context"
	"errors"

	"go.opentelemetry.io/contrib/exporters/autoexport"
	"go.opentelemetry.io/otel"

	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0" // Ensure this is imported
)

// OTelProviders manages the lifecycle of OpenTelemetry SDK components.
// It holds the concrete Tracer and Meter providers and provides a unified
// Shutdown function to ensure all telemetry data is flushed before the
// application exits.
type OTelProviders struct {
	TraceProvider *trace.TracerProvider
	MeterProvider *metric.MeterProvider
	// Shutdown performs a graceful shutdown of both trace and meter providers.
	Shutdown func(context.Context) error
}

// serviceName defines the default name for this service in the OpenTelemetry ecosystem.
const serviceName = "openoutbox-relay"

// NewOTelProviders bootstraps the OpenTelemetry pipeline.
// It initializes the global propagator, creates a shared resource with service metadata,
// and configures auto-exporting providers for both traces and metrics.
func NewOTelProviders(ctx context.Context) (*OTelProviders, error) {
	var shutdownFuncs []func(context.Context) error
	var err error

	shutdown := func(ctx context.Context) error {
		var err error
		for _, fn := range shutdownFuncs {
			err = errors.Join(err, fn(ctx))
		}
		shutdownFuncs = nil
		return err
	}

	// Set up propagator.
	otel.SetTextMapPropagator(newPropagator())

	// Set up Shared Resource.
	res, err := newResource()
	if err != nil {
		return nil, errors.Join(err, shutdown(ctx))
	}

	// Set up trace provider.
	tp, err := newTracerProvider(res)
	if err != nil {
		return nil, errors.Join(err, shutdown(ctx))
	}
	shutdownFuncs = append(shutdownFuncs, tp.Shutdown)
	otel.SetTracerProvider(tp)

	// Set up meter provider.
	mp, err := newMeterProvider(res)
	if err != nil {
		return nil, errors.Join(err, shutdown(ctx))
	}
	shutdownFuncs = append(shutdownFuncs, mp.Shutdown)
	otel.SetMeterProvider(mp)

	return &OTelProviders{
		TraceProvider: tp,
		MeterProvider: mp,
		Shutdown:      shutdown,
	}, nil
}

// newPropagator creates a composite propagator that handles both 
// W3C Trace Context and Baggage.
func newPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}

// newResource returns a resource describing this application, merging 
// default attributes with the service identification metadata.
func newResource() (*resource.Resource, error) {

	return resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			"",
			semconv.ServiceName(InstrumentationName),
		))
}

// newTracerProvider initializes a TracerProvider with a batch span processor 
// and an automatically detected exporter (OTLP, Jaeger, etc.).
func newTracerProvider(res *resource.Resource) (*trace.TracerProvider, error) {
	ctx := context.Background()

	traceExporter, err := autoexport.NewSpanExporter(ctx)

	if err != nil {
		return nil, err
	}

	tp := trace.NewTracerProvider(
		trace.WithBatcher(traceExporter),
		trace.WithResource(res),
	)
	return tp, nil
}

// newMeterProvider initializes a MeterProvider with a metric reader 
// and custom views for histograms. It defines specific bucket boundaries 
// for latencies (logarithmic scale) and batch sizes to ensure meaningful 
// observability in dashboards.
func newMeterProvider(res *resource.Resource) (*metric.MeterProvider, error) {
	ctx := context.Background()
	batchBuckets := []float64{1, 5, 10, 25, 50, 75, 100, 250, 500, 1000}
	customBuckets := []float64{
		.0005, // 500µs (Micro-latencies)
		.001,  // 1ms
		.0025, // 2.5ms
		.005,  // 5ms
		.01,   // 10ms
		.025,  // 25ms
		.05,   // 50ms
		.1,    // 100ms
		.25,   // 250ms
		.5,    // 500ms
		1,     // 1s (The "Warning" threshold)
		2.5,   // 2.5s
		5,     // 5s
		10,    // 10s (The "Critical/Timeout" threshold)
	}
	latencyView := metric.NewView(
		metric.Instrument{Name: "openoutbox.*.latency"},
		metric.Stream{
			Aggregation: metric.AggregationExplicitBucketHistogram{
				Boundaries: customBuckets,
			},
		},
	)
	batchView := metric.NewView(
		metric.Instrument{Name: "openoutbox.events.batch_size"},
		metric.Stream{
			Aggregation: metric.AggregationExplicitBucketHistogram{
				Boundaries: batchBuckets,
			},
		},
	)
	metricReader, err := autoexport.NewMetricReader(ctx)
	if err != nil {
		return nil, err
	}

	mp := metric.NewMeterProvider(
		metric.WithResource(res),
		metric.WithReader(metricReader),
		metric.WithView(latencyView),
		metric.WithView(batchView),
	)

	return mp, nil
}
