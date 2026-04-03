package telemetry

import (
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// Telemetry is the "Observability Bundle" for the entire system.
type Telemetry struct {
	Logger         *zap.Logger
	Metrics        *Metrics
	TracerProvider trace.TracerProvider
	MeterProvider  metric.MeterProvider
}

// Tracer returns a pre-configured tracer for a specific component.
func (t Telemetry) Tracer(name string) trace.Tracer {
	return t.TracerProvider.Tracer(name)
}

// Meter returns a pre-configured meter for a specific component.
func (t Telemetry) Meter(name string) metric.Meter {
	return t.MeterProvider.Meter(name)
}

// ScopedLogger returns a logger with the module name already baked in.
func (t Telemetry) ScopedLogger(moduleName string) *zap.Logger {
	return t.Logger.With(zap.String("module", moduleName))
}
