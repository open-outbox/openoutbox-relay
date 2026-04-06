package telemetry

import (
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// Telemetry is a composite structure that bundles all observability tools
// used by the relay. It simplifies dependency injection by providing a
// single object containing logging, metrics, tracing, and metering capabilities.
type Telemetry struct {
	// Logger is the structured logger (zap) used for recording events.
	Logger *zap.Logger
	// Metrics provides access to the custom application metrics instruments.
	Metrics *Metrics
	// Tracer provides the ability to create trace spans for distributed tracing.
	Tracer trace.Tracer
	// Meter provides the ability to create metric instruments for observability.
	Meter metric.Meter
}

// ScopedLogger returns a new logger instance with a "module" field
// preset to the provided moduleName for consistent contextual logging.
func (t Telemetry) ScopedLogger(moduleName string) *zap.Logger {
	return t.Logger.With(zap.String("module", moduleName))
}
